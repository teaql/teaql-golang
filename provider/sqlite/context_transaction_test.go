package sqlite

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	tsql "github.com/teaql/teaql-golang/sql"
)

func saveContextCustomer(ctx *runtime.UserContext, id int64, comment string) error {
	intent, _ := core.NewMutationIntent(&comment)
	return ctx.ExecuteGraphSave(intent, func() error {
		request, err := ds.NewMutationRequest(nativeBatchLeaf(id, fmt.Sprint("purchase ", id), ""), comment)
		if err != nil {
			return err
		}
		_, err = ctx.GetResource("dataService").(ds.MutationExecutor).Mutate(ctx, request)
		return err
	})
}

func queryContextCustomer(ctx *runtime.UserContext, id int64) (int, error) {
	service, ok := ctx.GetResource("dataService").(ds.QueryExecutor)
	if !ok {
		return 0, errors.New("no data service for this context")
	}
	query := core.NewSelectQuery("Customer").WithFilter(core.ExprEq("id", core.ValI64(id))).Limit(1).
		Comment("read purchase").Purpose("verify transaction visibility")
	result, err := service.Query(ctx, &ds.QueryRequest{Query: query, Comment: query.CommentText, Purpose: query.PurposeText})
	if err != nil {
		return 0, err
	}
	return len(result.Rows), nil
}

func TestUserContextTransactionComposesIndependentSaves(t *testing.T) {
	for _, mode := range []string{"commit", "propagated", "caught", "admission"} {
		t.Run(mode, func(t *testing.T) {
			fixture := nativeBatchEnvironment(t)
			fixture.ctx.InsertResource("approvalPolicy", "purchasing")
			if err := saveContextCustomer(fixture.ctx, 1, "seed duplicate"); err != nil {
				t.Fatal(err)
			}
			fixture.audit.events = nil
			var escaped *runtime.UserContext
			var rejected error
			commit := mode == "commit"
			err := fixture.ctx.Transaction(func(tx *runtime.UserContext) error {
				escaped = tx
				if tx == fixture.ctx || tx.GetResource("approvalPolicy") != "purchasing" {
					return fmt.Errorf("transaction context did not preserve the configured policy")
				}
				if err := saveContextCustomer(tx, 20, "approve purchase order"); err != nil {
					return err
				}
				if count, err := queryContextCustomer(tx, 20); count != 1 || err != nil || len(fixture.audit.events) != 0 {
					return fmt.Errorf("staged state: rows=%d, audit=%d, err=%v", count, len(fixture.audit.events), err)
				}
				if commit {
					return saveContextCustomer(tx, 21, "reserve inventory")
				}
				if mode == "admission" {
					rejected = saveContextCustomer(tx, 21, "")
				} else {
					rejected = saveContextCustomer(tx, 1, "reserve duplicate inventory")
				}
				if rejected == nil {
					return errors.New("failed save unexpectedly succeeded")
				}
				if mode == "propagated" {
					return rejected
				}
				if err := saveContextCustomer(tx, 22, "reserve after rejected save"); !errors.Is(err, rejected) {
					return fmt.Errorf("next save did not observe rollback-only status: %v", err)
				}
				return nil
			})
			want := 0
			if commit {
				want = 1
				if err != nil {
					t.Fatal(err)
				}
			} else if rejected == nil || err != rejected {
				t.Fatalf("transaction returned %v instead of the first save error %v", err, rejected)
			}
			for _, id := range []int64{20, 21} {
				if count, err := queryContextCustomer(fixture.ctx, id); count != want || err != nil {
					t.Fatalf("independent read of %d: rows=%d, error=%v", id, count, err)
				}
			}
			if len(fixture.audit.events) != 2*want {
				t.Fatalf("committed audit=%d, expected %d", len(fixture.audit.events), 2*want)
			}
			if escaped == nil || escaped.GetResource("dataService") != nil {
				t.Fatal("escaped transaction context can still query")
			}
		})
	}
}

type transactionBeginProbe struct {
	*tsql.SqlDataServiceExecutor
	begun chan struct{}
}

func (probe transactionBeginProbe) Begin(ctx context.Context) (ds.Transaction, error) {
	probe.begun <- struct{}{}
	return probe.SqlDataServiceExecutor.Begin(ctx)
}

func TestUserContextTransactionUsesAnExclusiveSqliteConnection(t *testing.T) {
	fixture := nativeBatchEnvironment(t)
	begun := make(chan struct{}, 2)
	fixture.ctx.InsertResource("dataService", transactionBeginProbe{fixture.executor, begun})
	timeout, cancel := context.WithTimeout(fixture.ctx.Context, 10*time.Second)
	defer cancel()
	fixture.ctx.Context = timeout
	firstReady, finishFirst := make(chan struct{}), make(chan struct{})
	firstResult, secondResult := make(chan error, 1), make(chan error, 1)
	failure := errors.New("approval rejected")
	go func() {
		firstResult <- fixture.ctx.Transaction(func(tx *runtime.UserContext) error {
			if err := saveContextCustomer(tx, 20, "approve first purchase"); err != nil {
				return err
			}
			close(firstReady)
			select {
			case <-finishFirst:
				return failure
			case <-tx.Done():
				return tx.Err()
			}
		})
	}()
	select {
	case <-firstReady:
	case err := <-firstResult:
		t.Fatalf("first transaction exited without staging: %v", err)
	case <-timeout.Done():
		t.Fatal(timeout.Err())
	}
	<-begun
	go func() {
		secondResult <- fixture.ctx.Transaction(func(tx *runtime.UserContext) error {
			if count, err := queryContextCustomer(tx, 20); count != 0 || err != nil {
				return fmt.Errorf("other transaction's order got %d rows: %v", count, err)
			}
			return saveContextCustomer(tx, 21, "approve second purchase")
		})
	}()
	select {
	case <-begun:
	case <-timeout.Done():
		t.Fatal(timeout.Err())
	}
	close(finishFirst)
	if err := <-firstResult; err != failure {
		t.Fatalf("first transaction did not roll back: %v", err)
	}
	if err := <-secondResult; err != nil {
		t.Fatal(err)
	}
	if count, err := queryContextCustomer(fixture.ctx, 21); count != 1 || err != nil {
		t.Fatalf("second order did not commit: rows=%d, err=%v", count, err)
	}
	if len(fixture.audit.events) != 1 {
		t.Fatalf("expected one committed audit, got %d", len(fixture.audit.events))
	}
}

package contextfix

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	lib "context-fix-service-core-workspace/lib"
	"context-fix-service-core-workspace/lib/entry"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

const clockA int64 = 1988247845000
const clockB int64 = clockA + 1000
const clockC int64 = clockA + 2000

type providerProbe struct {
	data_service.TransactionExecutor
	begins, mutations int
}

func (p *providerProbe) Query(ctx context.Context, request *data_service.QueryRequest) (*data_service.QueryResult, error) {
	return p.TransactionExecutor.(data_service.QueryExecutor).Query(ctx, request)
}

func (p *providerProbe) Begin(ctx context.Context) (data_service.Transaction, error) {
	p.begins++
	tx, err := p.TransactionExecutor.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &transactionProbe{Transaction: tx, probe: p}, nil
}

type transactionProbe struct {
	data_service.Transaction
	probe *providerProbe
}

func (tx *transactionProbe) Mutate(ctx context.Context, request data_service.MutationRequest) (*data_service.MutationResult, error) {
	tx.probe.mutations++
	return tx.Transaction.Mutate(ctx, request)
}
func (tx *transactionProbe) GenerateId(entity string) (uint64, error) {
	return tx.Transaction.(interface{ GenerateId(string) (uint64, error) }).GenerateId(entity)
}
func (tx *transactionProbe) EnsureIdFloor(ctx context.Context, entity string, floor uint64) error {
	return tx.Transaction.(interface {
		EnsureIdFloor(context.Context, string, uint64) error
	}).EnsureIdFloor(ctx, entity, floor)
}

type auditProbe struct{ mutations int }

func (p *auditProbe) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	if event.Kind == runtime.RawAuditEventKindCreated || event.Kind == runtime.RawAuditEventKindUpdated {
		p.mutations++
	}
	return nil
}

// An application-owned SPI adds IP/actor semantics to ordinary string fields.
// Generated library files are never edited and still own their model checks.
type applicationChecker struct{ generated runtime.CheckerRegistry }

func (c applicationChecker) CheckAndFix(ctx *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	results := []runtime.CheckResult{}
	if input.Entity == "entry" {
		for _, field := range []string{"source_ip", "last_ip", "opened_by", "changed_by"} {
			if input.Operation == core.MutationUpdate && (field == "source_ip" || field == "opened_by") {
				continue
			}
			if failure := ctx.ApplyContextFix(input, field, "identity"); failure != nil {
				results = append(results, *failure)
			}
		}
	}
	return append(results, c.generated.CheckAndFix(ctx, input)...)
}

type applicationResolver struct {
	phase                string
	failToken, invalidIP bool
}

func (r *applicationResolver) ResolveContextFix(_ *runtime.UserContext, input *runtime.CheckAndFixInput, field, function string, _ []string) (core.Value, error) {
	if function == "now" {
		return core.ValTimestamp(input.Now.UnixMilli()), nil
	}
	if function == "uniStr" {
		if r.failToken {
			return core.Value{}, errors.New("PRIVATE_CONTEXT_PROVIDER_CANARY")
		}
		return core.ValText("fixture-token"), nil
	}
	if r.invalidIP && field == "last_ip" {
		return core.ValText(strings.Repeat("x", 101)), nil
	}
	return core.ValText(r.phase + ":" + field), nil
}

func requireLocation(t *testing.T, err error, path string) {
	t.Helper()
	var failure *runtime.RuntimeError
	if !errors.As(err, &failure) {
		t.Fatalf("expected structured Check, got %v", err)
	}
	for _, result := range failure.CheckResults {
		if result.InstancePath() == path && result.EntityType != "" {
			return
		}
	}
	t.Fatalf("missing ObjectLocation %s: %#v", path, failure.CheckResults)
}

func TestGeneratedContextFixSQLite(t *testing.T) {
	file := os.Getenv("TEAQL_CONTEXT_FIX_DB")
	if file == "" {
		t.Fatal("use verify-local.cjs for an isolated retained database")
	}
	t.Setenv("CONTEXT_FIX_SERVICE_CORE_DATABASE_URL", file)
	ctx, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	defer ctx.GetResource("db").(*sql.DB).Close()
	ctx.DisableSqlLog()
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.UnixMilli(clockA)))
	if err = lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	meta := ctx.Entity("entry")
	for _, field := range []string{"source_ip", "opened_by"} {
		meta.PropertyByName(field).OnCreateFromContext("identity")
	}
	for _, field := range []string{"last_ip", "changed_by"} {
		meta.PropertyByName(field).OnCreateFromContext("identity").OnUpdateFromContext("identity")
	}
	ctx.WithCheckerRegistry(applicationChecker{generated: lib.Module().CheckerRegistry})
	resolver := &applicationResolver{phase: "A"}
	ctx.InsertResource("contextFixResolver", resolver)
	probe := &providerProbe{TransactionExecutor: ctx.GetResource("dataService").(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	audits := &auditProbe{}
	ctx.WithAppAuditEventSink(audits)
	makeEntry := func(name string) *entry.Entry {
		value := lib.Q.Entries().Comment("Prepare a Context fixture entity").Purpose("Verify generated mutation and validation").NewEntity(ctx)
		return value.UpdateName(name).UpdateCatalogId(1)
	}
	load := func(id uint64) *entry.Entry {
		value, queryErr := lib.Q.Entries().WithIdIs(id).Limit(1).Comment("Reload stored Context fixture").Purpose("Observe provider values through typed Q and E").ExecuteForOne(ctx)
		if queryErr != nil || value == nil {
			t.Fatalf("reload: %v", queryErr)
		}
		return value
	}
	broken := makeEntry("failed token")
	resolver.failToken = true
	_, err = broken.AuditAs("Reject a failed automatic Context provider").Save(ctx)
	requireLocation(t, err, "/guardedToken")
	if strings.Contains(err.Error(), "PRIVATE_CONTEXT_PROVIDER_CANARY") {
		t.Fatal("provider secret leaked")
	}
	if probe.begins != 0 || probe.mutations != 0 || audits.mutations != 0 {
		t.Fatal("failed create reached provider")
	}
	if _, exists := broken.Base().GetDynamic("born_at"); exists {
		t.Fatal("failed create retained staged Context")
	}
	resolver.failToken = false
	first := makeEntry("first")
	if _, err = first.AuditAs("Create Context fixture A").Save(ctx); err != nil {
		t.Fatal(err)
	}
	second := makeEntry("second")
	if _, err = second.AuditAs("Create Context fixture sibling").Save(ctx); err != nil {
		t.Fatal(err)
	}
	first = load(first.Id())
	before := core.CloneRecord(first.IntoRecord())
	oldVersion := first.Version()
	resolver.phase = "B"
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.UnixMilli(clockB)))
	beforeBegins, beforeWrites, beforeAudits := probe.begins, probe.mutations, audits.mutations
	for round := 0; round < 2; round++ {
		if _, err = first.AuditAs("Save unchanged entity under Context B").Save(ctx); err != nil {
			t.Fatal(err)
		}
	}
	if probe.begins != beforeBegins || probe.mutations != beforeWrites || audits.mutations != beforeAudits {
		t.Fatal("no-op made provider calls or audit")
	}
	if !reflect.DeepEqual(load(first.Id()).IntoRecord(), before) {
		t.Fatal("no-op changed stored Context values")
	}
	first.UpdateName(strings.Repeat("n", 101))
	_, err = first.AuditAs("Reject invalid business value").Save(ctx)
	requireLocation(t, err, "/name")
	if first.LastIp() != "A:last_ip" || first.TouchedAt().UnixMilli() != clockA {
		t.Fatal("failed check leaked automatic Context", first.LastIp(), first.TouchedAt(), first.IntoRecord())
	}
	first.UpdateName("first")
	if _, err = first.AuditAs("Retry after reverting caller intent").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if probe.begins != beforeBegins || audits.mutations != beforeAudits {
		t.Fatal("revert after failed Fix was not a no-op")
	}
	first.UpdateName("renamed")
	resolver.invalidIP = true
	_, err = first.AuditAs("Reject invalid automatically supplied IP").Save(ctx)
	requireLocation(t, err, "/lastIp")
	if first.LastIp() != "A:last_ip" || first.Name() != "renamed" {
		t.Fatal("rollback lost caller input or leaked invalid IP")
	}
	resolver.invalidIP = false
	if _, err = first.AuditAs("Persist one real change under B").Save(ctx); err != nil {
		t.Fatal(err)
	}
	stored := load(first.Id())
	name, present := lib.E.Entry(stored).Name().Eval()
	if !present || name != "renamed" || stored.Version() != oldVersion+1 || stored.BornAt().UnixMilli() != clockA || stored.TouchedAt().UnixMilli() != clockB || stored.SourceIp() != "A:source_ip" || stored.LastIp() != "B:last_ip" || stored.ChangedBy() != "B:changed_by" {
		t.Fatal("incorrect persisted result", stored.IntoRecord())
	}
	if probe.mutations != beforeWrites+1 || audits.mutations != beforeAudits+1 {
		t.Fatal("real update should write and audit once")
	}

	// Earlier successful sibling Fix must unwind when a later child fails.
	parent, queryErr := lib.Q.Catalogs().WithIdIs(1).Limit(1).Comment("Load unchanged graph ancestor").Purpose("Verify per-entity Context isolation").ExecuteForOne(ctx)
	if queryErr != nil || parent == nil {
		t.Fatal(queryErr)
	}
	first, second = load(first.Id()), load(second.Id())
	parentVersion := parent.Version()
	parent.EntryList().Add(first)
	parent.EntryList().Add(second)
	first.UpdateName("graph edit")
	second.UpdateName(strings.Repeat("z", 101))
	resolver.phase = "C"
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.UnixMilli(clockC)))
	beforeBegins, beforeWrites, beforeAudits = probe.begins, probe.mutations, audits.mutations
	_, err = parent.AuditAs("Reject a later child after preparing its sibling").Save(ctx)
	requireLocation(t, err, "/entryList/1/name")
	if first.LastIp() != "B:last_ip" || first.TouchedAt().UnixMilli() != clockB || first.Name() != "graph edit" {
		t.Fatal("earlier sibling was not rolled back", first.IntoRecord())
	}
	if probe.begins != beforeBegins || probe.mutations != beforeWrites || audits.mutations != beforeAudits {
		t.Fatal("failed graph reached provider")
	}
	second.UpdateName("second")
	if _, err = parent.AuditAs("Retry with only one independently modified child").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if load(second.Id()).LastIp() != "A:last_ip" || parent.Version() != parentVersion || load(first.Id()).LastIp() != "C:last_ip" {
		t.Fatal("unchanged sibling or parent was refreshed")
	}
	if probe.mutations != beforeWrites+1 || audits.mutations != beforeAudits+1 {
		t.Fatal("retry should commit exactly one child")
	}
	fmt.Println("GO_CONTEXT_FIX_SQLITE=passed typedQ=true typedE=true noOpProviderCalls=0 failedCheckProviderCalls=0 updateAudits=1 objectLocation=true graphRollback=true")
}

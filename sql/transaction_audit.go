package sql

import (
	"context"
	"errors"

	ds "github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/internal/mutationaudit"
)

type transactionAudits struct{ deliveries []func() error }

func (a *transactionAudits) capture(ctx context.Context, request ds.MutationRequest, result *ds.MutationResult) error {
	delivery, err := mutationaudit.Prepare(ctx, request, result)
	if err != nil {
		return err
	}
	if delivery != nil {
		a.deliveries = append(a.deliveries, delivery)
	}
	return nil
}

func deliverCommittedAudit(delivery func() error) (err error) {
	finished := false
	defer func() {
		if !finished {
			_ = recover()
			// Never expose a panic value, which may contain sensitive payloads.
			err = errors.New("mutation audit consumer panicked")
		}
	}()
	err = delivery()
	finished = true
	return err
}

func (a *transactionAudits) flush() error {
	deliveries := a.deliveries
	a.deliveries = nil
	var first error
	for _, delivery := range deliveries {
		if err := deliverCommittedAudit(delivery); err != nil && first == nil {
			first = err
		}
	}
	if first != nil {
		return &ds.MutationCommittedError{Cause: first}
	}
	return nil
}

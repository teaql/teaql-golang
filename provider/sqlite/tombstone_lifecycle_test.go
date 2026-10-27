package sqlite

import (
	"fmt"
	"testing"

	"github.com/teaql/teaql-golang/core"
	ds "github.com/teaql/teaql-golang/data_service"
)

func TestTombstoneProviderReadbackCommitRollbackAndBlindDelete(t *testing.T) {
	for _, mode := range []string{"ordinary", "explicit-commit", "explicit-rollback-retry"} {
		t.Run(mode, func(t *testing.T) {
			e := nativeBatchEnvironment(t)
			seed, err := ds.NewMutationRequest(nativeBatchLeaf(1, "live synthetic customer", "synthetic credential"), "Prepare one native tombstone fixture")
			if err != nil {
				t.Fatal(err)
			}
			if _, err := e.executor.Mutate(e.ctx, seed); err != nil {
				t.Fatal(err)
			}
			if _, err := e.db.Exec(`CREATE TRIGGER native_ml13_tombstone AFTER UPDATE OF version ON customer_data WHEN NEW.version < 0 AND OLD.version > 0 BEGIN UPDATE customer_data SET version=-777, display_name='database tombstone' WHERE id=NEW.id; END`); err != nil {
				t.Fatal(err)
			}
			e.audit.events = nil
			request, err := ds.NewMutationRequest(&ds.DeleteMutation{Cmd: core.NewDeleteCommand("Customer", core.ValI64(1)).WithExpectedVersion(1)}, "Return the trigger's authoritative tombstone")
			if err != nil {
				t.Fatal(err)
			}
			var result *ds.MutationResult
			if mode == "ordinary" {
				result, err = e.executor.Mutate(e.ctx, request)
			} else {
				tx, beginErr := e.executor.Begin(e.ctx)
				if beginErr != nil {
					t.Fatal(beginErr)
				}
				result, err = tx.Mutate(e.ctx, request)
				if err != nil {
					_ = tx.Rollback(e.ctx)
					t.Fatal(err)
				}
				if len(e.audit.events) != 0 {
					_ = tx.Rollback(e.ctx)
					t.Fatal("explicit mutation published audit before Commit")
				}
				version, present := result.PersistedRecord["version"].TryI64()
				if !present || version != -777 {
					_ = tx.Rollback(e.ctx)
					t.Fatal("explicit mutation returned a computed tombstone", result.PersistedRecord)
				}
				if mode == "explicit-commit" {
					err = tx.Commit(e.ctx)
				} else {
					err = tx.Rollback(e.ctx)
					var restored int64
					if queryErr := e.db.QueryRow("SELECT version FROM customer_data WHERE id=1").Scan(&restored); queryErr != nil || restored != 1 || len(e.audit.events) != 0 {
						t.Fatal("explicit rollback left a tombstone or audit", queryErr, restored)
					}
					if err == nil {
						result, err = e.executor.Mutate(e.ctx, request)
					}
				}
			}
			if err != nil {
				t.Fatal(err)
			}
			version, present := result.PersistedRecord["version"].TryI64()
			name, namePresent := result.PersistedRecord["display_name"].TryText()
			if result.AffectedRows != 1 || !present || version != -777 || !namePresent || name != "database tombstone" || len(e.audit.events) != 1 {
				t.Fatal("provider lost authoritative trigger state or committed audit", result, len(e.audit.events))
			}
			before := len(e.logs.entries)
			invalid, _ := ds.NewMutationRequest(&ds.DeleteMutation{Cmd: core.NewDeleteCommand("Customer", core.ValI64(1)).WithExpectedVersion(-777)}, "Reject a deletion of the same tombstone")
			if _, err := e.executor.Mutate(e.ctx, invalid); err == nil || len(e.logs.entries) != before || len(e.audit.events) != 1 {
				t.Fatal("invalid lifecycle entered SQL or audit", err)
			}
			blind, _ := ds.NewMutationRequest(&ds.DeleteMutation{Cmd: core.NewDeleteCommand("Customer", core.ValI64(1))}, "Do not rewrite tombstones through a blind administrative delete")
			ignored, err := e.executor.Mutate(e.ctx, blind)
			if err != nil || ignored.AffectedRows != 0 || len(e.audit.events) != 1 {
				t.Fatal("blind delete rewrote or audited a tombstone", err, ignored)
			}
			var physical int64
			if err := e.db.QueryRow("SELECT version FROM customer_data WHERE id=1").Scan(&physical); err != nil || physical != -777 {
				t.Fatal("stored tombstone was rewritten", err, physical)
			}
			fmt.Printf("CORE_PROVIDER_TOMBSTONE mode=%s version=%d audits=%d\n", mode, physical, len(e.audit.events))
		})
	}
}

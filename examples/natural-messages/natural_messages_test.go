package naturalmessages

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	lib "natural-message-service-core-workspace/lib"
	"natural-message-service-core-workspace/lib/catalog"
	"natural-message-service-core-workspace/lib/entry"
)

var secret = "密钥🔑" + strings.Repeat("x", 98) // 101 scalars, 102 UTF-16 code units.

type provider struct {
	data_service.TransactionExecutor
	queries, begins, writes int
}

func (p *provider) Query(ctx context.Context, q *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.queries++
	return p.TransactionExecutor.(data_service.QueryExecutor).Query(ctx, q)
}
func (p *provider) Begin(ctx context.Context) (data_service.Transaction, error) {
	p.begins++
	tx, err := p.TransactionExecutor.Begin(ctx)
	if err != nil {
		return nil, err
	}
	return &transaction{tx, p}, nil
}

type transaction struct {
	data_service.Transaction
	probe *provider
}

func (tx *transaction) Mutate(ctx context.Context, m data_service.MutationRequest) (*data_service.MutationResult, error) {
	tx.probe.writes++
	return tx.Transaction.Mutate(ctx, m)
}
func (tx *transaction) GenerateId(entity string) (uint64, error) {
	return tx.Transaction.(interface{ GenerateId(string) (uint64, error) }).GenerateId(entity)
}
func (tx *transaction) EnsureIdFloor(ctx context.Context, entity string, floor uint64) error {
	return tx.Transaction.(interface {
		EnsureIdFloor(context.Context, string, uint64) error
	}).EnsureIdFloor(ctx, entity, floor)
}

type audit struct{ updates int }

func (a *audit) OnSafeEvent(_ *runtime.UserContext, event *runtime.SafeAuditEvent) error {
	if event.Kind == runtime.RawAuditEventKindUpdated {
		a.updates++
	}
	return nil
}

func TestGeneratedNaturalMessagesSQLite(t *testing.T) {
	file := os.Getenv("TEAQL_NATURAL_MESSAGES_DB")
	if file == "" {
		t.Fatal("use examples/natural-messages/verify-local.cjs")
	}
	t.Setenv("NATURAL_MESSAGE_SERVICE_CORE_DATABASE_URL", file)
	ctx, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := ctx.GetResource("db").(*sql.DB)
	defer db.Close()
	ctx.DisableSqlLog()
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.UnixMilli(1988247845000)))
	_ = ctx.InstallCheckerMessageWarning(func(runtime.CheckerMessageWarning) {})
	// Only generated module composition supplies metadata, sensitivity and Checkers.
	if lib.Module().CheckerMessagesModel == nil || lib.Module().CheckerMessagesModel.Revision() == "" {
		t.Fatal("generated module did not provide presentation metadata")
	}
	if fields := ctx.Entity("entry").AuditMaskFlds; len(fields) != 1 || fields[0] != "name" {
		t.Fatal("generated sensitivity missing")
	}
	if err = lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	if err = lib.EnsureSchema(ctx); err != nil {
		t.Fatal(err)
	}
	probe := &provider{TransactionExecutor: ctx.GetResource("dataService").(data_service.TransactionExecutor)}
	ctx.InsertResource("dataService", probe)
	audits := &audit{}
	ctx.WithAppAuditEventSink(audits)
	makeEntry := func(name string) *entry.Entry {
		return lib.Q.Entries().Comment("Prepare an example child").Purpose("Verify natural messages with generated audited mutations").NewEntity(ctx).UpdateName(name).UpdateCatalogId(1)
	}
	load := func(id uint64) *entry.Entry {
		v, e := lib.Q.Entries().WithIdIs(id).Limit(1).Comment("Reload complete child").Purpose("Verify stored values through typed E").ExecuteForOne(ctx)
		if e != nil || v == nil {
			t.Fatalf("reload: %v", e)
		}
		return v
	}
	first, second := makeEntry("first"), makeEntry("second")
	if _, err = first.AuditAs("Create the first example child").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err = second.AuditAs("Create the second example child").Save(ctx); err != nil {
		t.Fatal(err)
	}
	first, second = load(first.Id()), load(second.Id())
	version := second.Version()
	if name, present := lib.E.Entry(second).Name().Eval(); !present || name != "second" {
		t.Fatal("typed E mismatch")
	}
	parent, err := lib.Q.Catalogs().WithIdIs(1).Limit(1).Comment("Load the example graph owner").Purpose("Locate a rejected child within a graph").ExecuteForOne(ctx)
	if err != nil || parent == nil {
		t.Fatal(err)
	}
	parentVersion := parent.Version()
	beforeQueries, beforeBegins, beforeWrites, beforeAudits := probe.queries, probe.begins, probe.writes, audits.updates
	responses := 0
	reject := func(graph *catalog.Catalog, change func(), locale, expected string) {
		t.Helper()
		_ = ctx.SetLocaleCode(locale)
		if change != nil {
			change()
		}
		beforeQueries, beforeBegins, beforeWrites, beforeAudits = probe.queries, probe.begins, probe.writes, audits.updates
		_, failure := graph.AuditAs("Reject the second child before provider activity").Save(ctx)
		var check *runtime.RuntimeError
		if !errors.As(failure, &check) || check.Type != "Check" {
			t.Fatalf("expected structured Check: %v", failure)
		}
		wire := []runtime.WireCheckResult{}
		found := false
		for _, result := range check.CheckResults {
			v := result.ToWire(runtime.JsonFieldCamelCase)
			wire = append(wire, v)
			if v.InstancePath == "/entryList/1/name" && v.EntityType == "Entry" && v.Message == expected {
				found = true
			}
		}
		data, _ := json.Marshal(map[string]any{"violations": wire})
		if !found {
			t.Fatalf("missing natural indexed error: %v; safe violations=%s", failure, data)
		}
		if strings.Contains(string(data), secret) || strings.Contains(failure.Error(), secret) {
			t.Fatal("secret leaked")
		}
		if probe.queries != beforeQueries || probe.begins != beforeBegins || probe.writes != beforeWrites || audits.updates != beforeAudits {
			t.Fatal("failed validation reached provider or audit")
		}
		responses++
		fmt.Println("GO_NATURAL_RESPONSE=" + string(data))
	}
	// A required Go string cannot receive nil through its typed setter. Omit it
	// on a new child instead; do not replace the generated required validator.
	requiredParent, err := lib.Q.Catalogs().WithIdIs(1).Limit(1).Comment("Load owner for missing-input validation").Purpose("Verify generated required checking before IDs").ExecuteForOne(ctx)
	if err != nil || requiredParent == nil {
		t.Fatal(err)
	}
	missing := lib.Q.Entries().Comment("Prepare a child with missing required input").Purpose("Exercise the native generated required constraint").NewEntity(ctx).UpdateCatalogId(1)
	requiredFirst := load(first.Id()).UpdateName("changed")
	requiredParent.EntryList().Add(requiredFirst)
	requiredParent.EntryList().Add(missing)
	reject(requiredParent, nil, "en", "The name of the 2nd entry is required.")
	reject(requiredParent, nil, "zh-CN", "请填写第2项明细的名称。")
	if missing.Id() != 0 {
		t.Fatal("required failure reserved an identity")
	}
	parent.EntryList().Add(first)
	parent.EntryList().Add(second)
	first.UpdateName("changed")
	reject(parent, func() { second.UpdateName(secret) }, "en", "The name of the 2nd entry must not exceed 100 Unicode scalar values; its current length is 101.")
	rendererCalls := 0
	if err = ctx.InstallCheckerMessageRenderer(runtime.CheckerMessageRendererFunc(func(request runtime.CheckerMessageRequest) (runtime.CheckerMessageOutput, error) {
		rendererCalls++
		if request.Arguments["input"] != "密钥"+strings.Repeat("*", 97)+"xx" || request.Arguments["input_len"] != "101" {
			return runtime.CheckerMessageOutput{}, fmt.Errorf("customer renderer received incorrect projected arguments")
		}
		return runtime.CheckerMessageOutput{Message: "Private " + secret}, nil
	})); err != nil {
		t.Fatal(err)
	}
	reject(parent, nil, "en", "Private [REDACTED]")
	if rendererCalls != 1 {
		t.Fatal("graph relocation called the customer renderer again")
	}
	if second.Name() != secret || second.Version() != version {
		t.Fatal("failed validation lost caller input or changed version")
	}
	first.UpdateName("first")
	second.UpdateName("renamed")
	if _, err = parent.AuditAs("Persist only the corrected second child").Save(ctx); err != nil {
		t.Fatal(err)
	}
	if probe.writes != beforeWrites+1 || audits.updates != beforeAudits+1 {
		t.Fatal("expected exactly one child write and update audit")
	}
	stored := load(second.Id())
	name, present := lib.E.Entry(stored).Name().Eval()
	if !present || name != "renamed" || stored.Version() != version+1 {
		t.Fatal("persisted typed E mismatch")
	}
	if parent.Version() != parentVersion || rendererCalls != 1 {
		t.Fatal("unchanged parent or renderer state changed")
	}
	// Independent readback is the oracle, not an alternative application API.
	var storedName string
	var storedVersion int64
	if err = db.QueryRow("SELECT name,version FROM entry_data WHERE id=?", second.Id()).Scan(&storedName, &storedVersion); err != nil || storedName != "renamed" || storedVersion != int64(version+1) {
		t.Fatal("SQLite readback mismatch", err)
	}
	frame, _ := json.Marshal(map[string]any{"passed": true, "first": first.Id(), "child": second.Id(), "version": stored.Version(), "responses": responses, "providerCallsOnRejection": 0, "transactionCallsOnRejection": 0, "updatedAudits": 1, "typedQ": true, "typedE": true,
		"automaticMetadata": true, "originalGeneratedChecker": true, "unchangedParent": true, "requiredBeforeIds": true, "customerRendererCalls": rendererCalls})
	fmt.Println("GO_NATURAL_SQLITE_PASS=" + string(frame))
}

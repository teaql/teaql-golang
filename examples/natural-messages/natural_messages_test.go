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

	lib "context-fix-service-core-workspace/lib"
	"context-fix-service-core-workspace/lib/entry"
	"github.com/teaql/teaql-golang/core"
	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
)

const secret = "密钥🔑abcdef" // Nine Unicode scalar values, ten UTF-16 code units.

// This is application code around the unchanged generated Checker. It declares
// a local sensitivity policy and a deliberately small display-name limit.
type checker struct{ generated runtime.CheckerRegistry }

func (c checker) CheckAndFix(ctx *runtime.UserContext, input *runtime.CheckAndFixInput) []runtime.CheckResult {
	if input.Entity == "entry" && input.Operation == core.MutationInsert {
		for _, field := range []string{"source_ip", "last_ip", "opened_by", "changed_by"} {
			if _, exists := input.Values[field]; !exists {
				input.Values[field] = core.ValText("local")
			}
		}
	}
	results := c.generated.CheckAndFix(ctx, input)
	if input.Entity == "entry" {
		name, _ := input.Values["name"].TryText()
		if strings.TrimSpace(name) == "" {
			present := false
			for _, r := range results {
				if r.RuleID == "required" && r.ObjectLocation().ModelPath() == "name" {
					present = true
				}
			}
			if !present {
				results = append(results, runtime.CheckResult{RuleID: "required", CanonicalLocation: runtime.Location().Property("name")})
			}
		}
		if len([]rune(name)) > 8 {
			results = append(results, runtime.CheckResult{RuleID: "max_length", CanonicalLocation: runtime.Location().Property("name"), InputValue: name, SystemValue: 8})
		}
	}
	return results
}

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
	t.Setenv("CONTEXT_FIX_SERVICE_CORE_DATABASE_URL", file)
	ctx, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	db := ctx.GetResource("db").(*sql.DB)
	defer db.Close()
	ctx.DisableSqlLog()
	ctx.WithBusinessClock(runtime.NewFixedBusinessClock(time.UnixMilli(1988247845000)))
	_ = ctx.InstallCheckerMessageWarning(func(runtime.CheckerMessageWarning) {})
	presentation, err := runtime.ParseCheckerPresentationModel([]byte(`{
      "schema":"teaql.checker-model/v1","revision":"generated-natural-example-v1","entities":{
        "catalog":{"labels":{"en":"catalog","zh-CN":"目录"},"fields":[
          {"name":"name","labels":{"en":"name","zh-CN":"名称"}},
          {"name":"entry_list","aliases":["entryList"],"targetType":"entry","labels":{"en":"entry","zh-CN":"明细"},"classifiers":{"zh-CN":"条"}}]},
        "entry":{"labels":{"en":"entry","zh-CN":"明细"},"fields":[
          {"name":"name","labels":{"en":"name","zh-CN":"名称"}}]}}
    }`))
	if err != nil {
		t.Fatal(err)
	}
	if err = ctx.InstallCheckerPresentationModel(presentation); err != nil {
		t.Fatal(err)
	}
	ctx.WithCheckerRegistry(checker{generated: lib.Module().CheckerRegistry})
	ctx.Entity("entry").AuditMaskFlds = append(ctx.Entity("entry").AuditMaskFlds, "name")
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
	parent.EntryList().Add(first)
	parent.EntryList().Add(second)
	first.UpdateName("changed")
	beforeQueries, beforeBegins, beforeWrites, beforeAudits := probe.queries, probe.begins, probe.writes, audits.updates
	responses := 0
	reject := func(value, locale, expected string) {
		t.Helper()
		_ = ctx.SetLocaleCode(locale)
		second.UpdateName(value)
		_, failure := parent.AuditAs("Reject the second child before provider activity").Save(ctx)
		var check *runtime.RuntimeError
		if !errors.As(failure, &check) || check.Type != "Check" {
			t.Fatalf("expected structured Check: %v", failure)
		}
		wire := []runtime.WireCheckResult{}
		found := false
		for _, result := range check.CheckResults {
			v := result.ToWire(runtime.JsonFieldCamelCase)
			wire = append(wire, v)
			if v.InstancePath == "/entryList/1/name" && v.EntityType == "entry" && v.Message == expected {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing natural indexed error: %v", failure)
		}
		data, _ := json.Marshal(map[string]any{"violations": wire})
		if strings.Contains(string(data), secret) || strings.Contains(failure.Error(), secret) {
			t.Fatal("secret leaked")
		}
		if probe.queries != beforeQueries || probe.begins != beforeBegins || probe.writes != beforeWrites || audits.updates != beforeAudits {
			t.Fatal("failed validation reached provider or audit")
		}
		responses++
		fmt.Println("GO_NATURAL_RESPONSE=" + string(data))
	}
	reject("", "en", "The name of the 2nd entry is required.")
	reject("", "zh-CN", "请填写第2条明细的名称。")
	reject(secret, "en", "The name of the 2nd entry must not exceed 8 Unicode scalar values; its current length is 9.")
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
	// Independent readback is the oracle, not an alternative application API.
	var storedName string
	var storedVersion int64
	if err = db.QueryRow("SELECT name,version FROM entry_data WHERE id=?", second.Id()).Scan(&storedName, &storedVersion); err != nil || storedName != "renamed" || storedVersion != int64(version+1) {
		t.Fatal("SQLite readback mismatch", err)
	}
	frame, _ := json.Marshal(map[string]any{"passed": true, "first": first.Id(), "child": second.Id(), "version": stored.Version(), "responses": responses, "providerCallsOnRejection": 0, "transactionCallsOnRejection": 0, "updatedAudits": 1, "typedQ": true, "typedE": true})
	fmt.Println("GO_NATURAL_SQLITE_PASS=" + string(frame))
}

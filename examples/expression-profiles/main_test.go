package profiles

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/data_service"
	"github.com/teaql/teaql-golang/runtime"
	lib "scalar-type-conformance-service-core-workspace/lib"
)

type observedProvider struct {
	data_service.DataServiceExecutor
	queries, mutations, begins, ids int
}

func (p *observedProvider) Query(c context.Context, q *data_service.QueryRequest) (*data_service.QueryResult, error) {
	p.queries++
	return p.DataServiceExecutor.(data_service.QueryExecutor).Query(c, q)
}
func (p *observedProvider) Mutate(c context.Context, m data_service.MutationRequest) (*data_service.MutationResult, error) {
	p.mutations++
	return p.DataServiceExecutor.(data_service.MutationExecutor).Mutate(c, m)
}
func (p *observedProvider) Begin(c context.Context) (data_service.Transaction, error) {
	p.begins++
	return p.DataServiceExecutor.(data_service.TransactionExecutor).Begin(c)
}
func (p *observedProvider) NextId(c context.Context, entity string) (uint64, error) {
	p.ids++
	return p.DataServiceExecutor.(data_service.IdGeneratorExecutor).NextId(c, entity)
}
func (p *observedProvider) snapshot() [4]int { return [4]int{p.queries, p.mutations, p.begins, p.ids} }

func TestGeneratedExpressionModelProfile(t *testing.T) {
	profile := os.Getenv("TEAQL_EXPRESSION_PROFILE")
	pointer := map[string]string{"camelCase": "/scalarTypeProbeList", "snake_case": "/scalar_type_probe_list", "PascalCase": "/ScalarTypeProbeList"}[profile]
	suffix := map[string]string{"camelCase": "/0/requiredString", "snake_case": "/0/required_string", "PascalCase": "/0/RequiredString"}[profile]
	if pointer == "" {
		t.Fatal("unknown model profile")
	}
	t.Setenv("SCALAR_TYPE_CONFORMANCE_SERVICE_CORE_DATABASE_URL", os.Getenv("TEAQL_EXPRESSION_DB"))
	c, err := lib.ServiceRuntimeFromEnv()
	if err != nil {
		t.Fatal(err)
	}
	if db, ok := c.GetResource("db").(*sql.DB); ok {
		defer db.Close()
	}
	if err := lib.EnsureSchema(c); err != nil {
		t.Fatal(err)
	}
	observer := &observedProvider{DataServiceExecutor: c.GetResource("dataService").(data_service.DataServiceExecutor)}
	c.InsertResource("dataService", observer)
	rows, err := lib.Q.TypePlatforms().WithIdIs(1).Limit(1).
		Comment("Load the root without its reverse list").
		Purpose("Verify generated E diagnostics under the modeled JSON profile").ExecuteForList(c)
	if err != nil || len(rows.Data) != 1 {
		t.Fatal("seeded root missing", err)
	}
	row := rows.Data[0]
	name, _, err := lib.E.TypePlatform(row).Name().TryEval()
	if err != nil {
		t.Fatal(err)
	}
	before := observer.snapshot()
	if before[0] == 0 {
		t.Fatal("the observer must measure the actual generated Q query")
	}
	_, _, err = lib.E.TypePlatform(row).ScalarTypeProbeList().First().RequiredString().TryEval()
	var missing *runtime.NotLoadedError
	if !errors.As(err, &missing) {
		t.Fatalf("expected NotLoaded, got %T: %v", err, err)
	}
	wire := missing.ToWire()
	encoded, err := json.Marshal(wire)
	if err != nil {
		t.Fatal(err)
	}
	var document map[string]any
	if err := json.Unmarshal(encoded, &document); err != nil {
		t.Fatal(err)
	}
	location := []any{map[string]any{"kind": "property", "name": "scalar_type_probe_list"}}
	attempted := append(append([]any{}, location...), map[string]any{"kind": "index", "index": float64(0)}, map[string]any{"kind": "property", "name": "required_string"})
	if wire.Code != "NOT_LOADED" || !reflect.DeepEqual(document["location"], location) || !reflect.DeepEqual(document["attemptedLocation"], attempted) {
		t.Fatal("canonical location drift", document)
	}
	if wire.InstancePath != pointer || wire.AttemptedInstancePath != pointer+suffix || wire.BreakPoint != "scalar_type_probe_list" || wire.BreakPointIndex != 0 {
		t.Fatal(wire)
	}
	if wire.OriginalExpression != "E.TypePlatform(<root>).ScalarTypeProbeList().First().RequiredString().TryEval()" ||
		wire.ExpressionWithBrokenPoint != "E.TypePlatform(<root>).ScalarTypeProbeList()<broken>.First().RequiredString().TryEval()" {
		t.Fatal(wire)
	}
	if !reflect.DeepEqual(document["suggestedFix"], map[string]any{"action": "preload", "location": location}) {
		t.Fatal("preload identity drift")
	}
	human := missing.Error()
	if !strings.HasPrefix(human, "TeaQLNotLoadedError:") || !strings.Contains(human, "expression="+wire.ExpressionWithBrokenPoint) || strings.Contains(human, `"instancePath"`) {
		t.Fatal("human view drift", human)
	}
	afterName, _, err := lib.E.TypePlatform(row).Name().TryEval()
	if err != nil || afterName != name || observer.snapshot() != before {
		t.Fatal("E changed state or accessed provider", err)
	}
	round, err := strconv.Atoi(os.Getenv("TEAQL_EXPRESSION_ROUND"))
	if err != nil {
		t.Fatal(err)
	}
	result, err := json.Marshal(map[string]any{"language": "go", "profile": profile, "round": round, "passed": true,
		"wire": wire, "human": human, "providerEntries": 0, "queryCalibration": before[0], "selectedNameUnchanged": true})
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("GENERATED_E_MODEL_PROFILE=" + string(result))
}

package sql

import (
	"encoding/json"
	"reflect"
	"strings"
	"sync"
	"testing"

	"github.com/teaql/teaql-golang/core"
)

func TestTypedLikeCompilerRetainsExactFieldPolicyAndOwnsEachInvocation(t *testing.T) {
	public := core.NewPropertyDescriptor("public_name", core.TypeText)
	public.LogPolicy = "plain"
	entity := maskPolicyEntity().Property(public).
		Property(core.NewPropertyDescriptor("password_hash", core.TypeText)).
		Property(core.NewPropertyDescriptor("unknown", core.TypeText))
	dialect := &DefaultSqlDialect{Dialect: &TestDialect{}}
	var group sync.WaitGroup
	for i, field := range []string{"name", "public_name", "password_hash", "unknown"} {
		group.Add(1)
		go func(i int, field string) {
			defer group.Done()
			original := "%ORIGINAL_" + field + "\\%"
			query := core.NewSelectQuery("Customer").AndFilter(core.ExprBeginWith(field, original))
			compiled, err := dialect.CompileSelect(entity, query.Clone())
			if err != nil {
				t.Error(err)
				return
			}
			policy := []string{"masked", "plain", "credential", "unknown"}[i]
			if !reflect.DeepEqual(compiled.Params, []core.Value{core.ValText(original + "%")}) || !reflect.DeepEqual(compiled.ParameterLogPolicies, []string{policy}) {
				t.Error("bindings or policy changed")
			}
			if !reflect.DeepEqual(compiled.intentOperands, []intentOperand{{value: core.ValText(original), policy: policy}}) {
				t.Errorf("incorrect original provenance: %#v", compiled.intentOperands)
			}
			encoded, _ := json.Marshal(compiled)
			if strings.Contains(string(encoded), "intentOperands") {
				t.Error("private provenance serialized")
			}
		}(i, field)
	}
	group.Wait()
	if dialect.likeOperands != nil || dialect.logPolicies != nil {
		t.Fatal("reusable compiler retained request state")
	}
	raw, err := dialect.CompileSelect(entity, core.NewSelectQuery("Customer").AndFilter(core.ExprLike("name", "%RAW_%")))
	if err != nil || len(raw.intentOperands) != 0 {
		t.Fatal("raw LIKE inferred original operand", err)
	}
}

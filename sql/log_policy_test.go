package sql

import (
	"github.com/teaql/teaql-golang/core"
	"reflect"
	"sync"
	"testing"
)

func maskPolicyEntity() *core.EntityDescriptor {
	active := core.NewPropertyDescriptor("active", core.TypeBool)
	active.LogPolicy = "plain"
	return core.NewEntityDescriptor("Customer").Property(core.NewPropertyDescriptor("id", core.TypeI64).Id()).
		Property(core.NewPropertyDescriptor("version", core.TypeI64).Version()).
		Property(core.NewPropertyDescriptor("name", core.TypeText)).Property(active).AuditMaskFields([]string{"name"})
}

func TestLegacyMissingMaskMetadataFailsClosed(t *testing.T) {
	name := core.NewPropertyDescriptor("name", core.TypeText)
	name.LogPolicy = "plain" // Older generated descriptors could set this without declaring the entity policy.
	entity := core.NewEntityDescriptor("Customer").Property(name).
		Property(core.NewPropertyDescriptor("password", core.TypeText))
	if got := fieldLogPolicy(entity, "name"); got != "unknown" {
		t.Fatalf("missing entity policy must be unknown, got %q", got)
	}
	if got := fieldLogPolicy(entity, "password"); got != "credential" {
		t.Fatalf("credentials must remain masked, got %q", got)
	}
	d := &DefaultSqlDialect{Dialect: &TestDialect{}}
	compiled, err := d.CompileSelect(entity, core.NewSelectQuery("Customer").AndFilter(core.ExprEq("name", core.ValText("PRIVATE-CANARY"))))
	if err != nil || !reflect.DeepEqual(compiled.ParameterLogPolicies, []string{"unknown"}) {
		t.Fatalf("legacy binding policy: %v, %v", compiled, err)
	}
	entity.AuditMaskFields([]string{})
	if got := fieldLogPolicy(entity, "name"); got != "plain" {
		t.Fatalf("explicit empty policy should retain ordinary field, got %q", got)
	}
}

func TestCompilerMaskPoliciesBatchAndGuards(t *testing.T) {
	d := &DefaultSqlDialect{Dialect: &TestDialect{}}
	entity := maskPolicyEntity()
	insert := core.NewBatchInsertCommand("Customer")
	insert.BatchValues = []core.Record{{"name": core.ValText("Riverside"), "active": core.ValBool(true)}, {"name": core.ValText("O'Reilly"), "active": core.ValBool(false)}}
	compiled, err := d.CompileBatchInsert(entity, insert)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(compiled.ParameterLogPolicies, []string{"masked", "plain", "masked", "plain"}) {
		t.Fatal(compiled.ParameterLogPolicies)
	}
	update := core.NewBatchUpdateCommand("Customer", []string{"name"})
	update.BatchIds = []core.Value{core.ValI64(1), core.ValI64(2)}
	update.BatchValues = []core.Record{{"name": core.ValText("Riverside")}, {"name": core.ValText("O'Reilly")}}
	compiled, err = d.CompileBatchUpdate(entity, update)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(compiled.ParameterLogPolicies, []string{"unknown", "masked", "unknown", "masked", "unknown", "unknown"}) {
		t.Fatal(compiled.ParameterLogPolicies)
	}
	guarded := core.NewUpdateCommand("Customer", core.ValI64(1)).Value("active", core.ValBool(true)).Guard("name", core.ValText("Riverside"))
	compiled, err = d.CompileUpdate(entity, guarded)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(compiled.ParameterLogPolicies, []string{"plain", "unknown", "masked"}) {
		t.Fatal(compiled.ParameterLogPolicies)
	}
}

func TestCompilerMaskPoliciesArePerRequest(t *testing.T) {
	d := &DefaultSqlDialect{Dialect: &TestDialect{}}
	entity := maskPolicyEntity()
	var group sync.WaitGroup
	for i := 0; i < 64; i++ {
		group.Add(1)
		go func(i int) {
			defer group.Done()
			field, expected := "name", "masked"
			if i%2 == 0 {
				field, expected = "active", "plain"
			}
			query := core.NewSelectQuery("Customer").AndFilter(core.ExprEq(field, core.ValText("value")))
			compiled, err := d.CompileSelect(entity, query)
			if err != nil || !reflect.DeepEqual(compiled.ParameterLogPolicies, []string{expected}) {
				t.Errorf("policy cross-talk: %v %v", compiled, err)
			}
		}(i)
	}
	group.Wait()
	if d.logPolicies != nil {
		t.Fatal("shared compiler retains request state")
	}
	raw, err := d.CompileSelect(entity, core.NewSelectQuery("Customer").WithRawSql("SELECT 'INLINE-CANARY'"))
	if err != nil || raw.GeneratedSQL {
		t.Fatal("raw SQL marked trusted", err)
	}
}

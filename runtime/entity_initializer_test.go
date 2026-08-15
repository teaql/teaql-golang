package runtime

import "testing"

type testEntityInitializer struct {
	name string
}

func (i *testEntityInitializer) InitializeEntity(_ *UserContext, name string, entity any) any {
	i.name = name
	return entity
}

func TestInitializeEntityUsesTrustedContextHook(t *testing.T) {
	ctx := NewUserContext()
	initializer := &testEntityInitializer{}
	ctx.InsertResource("entityInitializer", initializer)
	entity := &struct{ Name string }{Name: "Ada"}
	if got := ctx.InitializeEntity("Person", entity); got != entity {
		t.Fatal("initializer changed the concrete entity")
	}
	if initializer.name != "Person" {
		t.Fatalf("initializer received %q", initializer.name)
	}
}

func TestInitializeEntityRejectsMissingContext(t *testing.T) {
	defer func() {
		if recover() == nil {
			t.Fatal("nil context did not fail")
		}
	}()
	var ctx *UserContext
	ctx.InitializeEntity("Person", &struct{}{})
}

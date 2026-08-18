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
	context := NewUserContext()
	initializer := &testEntityInitializer{}
	context.InsertResource("entityInitializer", initializer)
	entity := &struct{ Name string }{Name: "Ada"}
	if got := context.InitializeEntity("Person", entity); got != entity {
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
	var context *UserContext
	context.InitializeEntity("Person", &struct{}{})
}

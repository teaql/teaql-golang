package runtime

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestExpressionEarlyRepeatedRelationPreservesOneBreakpoint(t *testing.T) {
	parent := NewExpressionPath("E.Node(<root>)", JsonFieldCamelCase).Member("parent", "Parent").MarkMissing()
	leaf := parent.Member("parent", "Parent").Member("user_url", "UserUrl")
	err := NewNotLoadedError(parent, "parent").WithPath(leaf).Terminal("Eval()")
	wire := err.ToWire()
	if !strings.Contains(err.Error(), "<broken>") {
		t.Fatal("native errors must render a readable breakpoint")
	}
	if wire.Code != "NOT_LOADED" || wire.InstancePath != "/parent" || wire.AttemptedInstancePath != "/parent/parent/userUrl" {
		t.Fatalf("wrong structured paths: %s", err)
	}
	if wire.OriginalExpression != "E.Node(<root>).Parent().Parent().UserUrl().Eval()" {
		t.Fatal(wire)
	}
	if wire.ExpressionWithBrokenPoint != "E.Node(<root>).Parent()<broken>.Parent().UserUrl().Eval()" || strings.Count(wire.ExpressionWithBrokenPoint, "<broken>") != 1 {
		t.Fatal(wire)
	}
	if wire.BreakPointIndex != 0 || wire.SuggestedFix.Action != "preload" {
		t.Fatal(wire)
	}
}

func TestExpressionIndicesProfilesAndOperationsAreNotReparsed(t *testing.T) {
	for _, profile := range []JsonFieldNamingProfile{JsonFieldCamelCase, JsonFieldSnakeCase, JsonFieldPascalCase} {
		items := NewExpressionPath("E.Order(<root>)", profile).Member("order_items", "OrderItems").MarkMissing()
		err := NewNotLoadedError(items.At(12, false).Member("user_url", "UserUrl"), "order_items").Terminal("Eval()")
		wire := err.ToWire()
		if wire.AttemptedLocation[1].Kind != "index" || *wire.AttemptedLocation[1].Index != 12 {
			t.Fatal(wire)
		}
		if wire.Location[0].Name != "order_items" || !strings.Contains(err.Error(), ".Get(12).UserUrl().Eval()") {
			t.Fatal(wire)
		}
		first := NewNotLoadedError(items.At(0, true), "order_items").Terminal("Eval()")
		if !strings.Contains(first.Error(), ".First().Eval()") {
			t.Fatal(first)
		}
		size := NewNotLoadedError(items.Operation("Size()"), "order_items").Terminal("Eval()").ToWire()
		if len(size.AttemptedLocation) != 1 {
			t.Fatal(size)
		}
	}
}

func TestExpressionSiblingErrorsAndWireProjectionOmitInputValues(t *testing.T) {
	parent := NewExpressionPath("E.Node(<root>)", JsonFieldCamelCase).Member("a~/b.c", "SafeMethod").MarkMissing()
	first := NewNotLoadedError(parent.Member("name", "Name"), "a~/b.c").Terminal("OrElse(<fallback>)")
	second := NewNotLoadedError(parent.Member("code", "Code"), "a~/b.c").Terminal("Eval()")
	if first.ToWire().InstancePath != "/a~0~1b.c" || strings.Contains(first.Error(), ".Code()") {
		t.Fatal(first)
	}
	encoded, _ := json.Marshal(second.ToWire())
	var wire map[string]any
	if err := json.Unmarshal(encoded, &wire); err != nil {
		t.Fatal(err)
	}
	if _, ok := wire["root"]; ok {
		t.Fatal(wire)
	}
	if _, ok := wire["inputValue"]; ok {
		t.Fatal(wire)
	}
}

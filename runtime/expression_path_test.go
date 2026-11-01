package runtime

import (
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"
)

func TestExpressionLiteralPunctuationPreservesCanonicalIdentityInAllWireProfiles(t *testing.T) {
	for _, example := range []struct {
		profile        JsonFieldNamingProfile
		label, pointer string
	}{
		{JsonFieldCamelCase, "camelCase", "/schoolType/2/a~0~1b.c"},
		{JsonFieldSnakeCase, "snake_case", "/school_type/2/a~0~1b.c"},
		{JsonFieldPascalCase, "PascalCase", "/SchoolType/2/A~0~1b.c"},
	} {
		t.Run(example.label, func(t *testing.T) {
			path := NewExpressionPath("E.Node(<root>)", example.profile).Member("school_type", "SchoolType").At(2, false).Member("a~/b.c", "SafeField").MarkMissing()
			wire := NewNotLoadedError(path, "a~/b.c").Terminal("Eval()").ToWire()
			if wire.InstancePath != example.pointer || wire.AttemptedInstancePath != example.pointer || wire.BreakPoint != "a~/b.c" || wire.BreakPointIndex != 2 {
				t.Fatal(wire)
			}
			canonical, err := json.Marshal(wire.Location)
			if err != nil {
				t.Fatal(err)
			}
			var actual, expected any
			if err := json.Unmarshal(canonical, &actual); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(`[{"kind":"property","name":"school_type"},{"kind":"index","index":2},{"kind":"property","name":"a~/b.c"}]`), &expected); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actual, expected) {
				t.Fatal(string(canonical))
			}
			attempted, _ := json.Marshal(wire.AttemptedLocation)
			suggested, _ := json.Marshal(wire.SuggestedFix.Location)
			if string(attempted) != string(canonical) || string(suggested) != string(canonical) {
				t.Fatal(wire)
			}
			if wire.ExpressionWithBrokenPoint != "E.Node(<root>).SchoolType().Get(2).SafeField()<broken>.Eval()" {
				t.Fatal(wire)
			}
			record, err := json.Marshal(map[string]any{"language": "go", "profile": example.label, "wire": wire})
			if err != nil {
				t.Fatal(err)
			}
			t.Log("E_POINTER_PROFILE_CASE=" + string(record))
			if output := os.Getenv("TEAQL_POINTER_PROFILE_EVIDENCE"); output != "" {
				file, err := os.OpenFile(output, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0600)
				if err != nil {
					t.Fatal(err)
				}
				_, writeErr := file.Write(append(record, '\n'))
				closeErr := file.Close()
				if writeErr != nil {
					t.Fatal(writeErr)
				}
				if closeErr != nil {
					t.Fatal(closeErr)
				}
			}
		})
	}
}

func TestExpressionEarlyRepeatedRelationPreservesOneBreakpoint(t *testing.T) {
	parent := NewExpressionPath("E.Node(<root>)", JsonFieldCamelCase).Member("parent", "Parent").MarkMissing()
	leaf := parent.Member("parent", "Parent").Member("user_url", "UserUrl")
	err := NewNotLoadedError(parent, "parent").WithPath(leaf).Terminal("Eval()")
	wire := err.ToWire()
	if err.Error() != "TeaQLNotLoadedError: access_path=Parent break_point=parent — select the missing field or relation expression="+wire.ExpressionWithBrokenPoint {
		t.Fatal("default errors must be human-readable, not JSON", err.Error())
	}
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

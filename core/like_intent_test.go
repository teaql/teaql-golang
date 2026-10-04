package core

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/teaql/teaql-golang/internal/logprivacy"
)

func TestTypedLikeProvenanceClonesButDoesNotBecomeWireInput(t *testing.T) {
	for _, helper := range []func(string, string) *Expr{ExprContain, ExprNotContain, ExprBeginWith, ExprNotBeginWith, ExprEndWith, ExprNotEndWith} {
		const original = "%LITERAL_WILDCARD\\%"
		expr := helper("name", original)
		clone := expr.Clone()
		source, ok := logprivacy.ReadIntentSource(clone.DiagnosticLikeOperand()).(Value)
		if !ok || source.V != original {
			t.Fatal("clone lost exact original operand")
		}
		if strings.Contains(fmt.Sprintf("%+v %#v", clone.DiagnosticLikeOperand(), clone.DiagnosticLikeOperand()), original) {
			t.Fatal("opaque source formatting leaked")
		}
		encoded, err := json.Marshal(expr)
		if err != nil {
			t.Fatal(err)
		}
		raw := ExprBinaryNode(ExprColumnNode("name"), expr.Op, ExprValueNode(expr.Right.Value))
		rawJSON, _ := json.Marshal(raw)
		if string(encoded) != string(rawJSON) {
			t.Fatal("provenance changed wire shape")
		}
		var decoded Expr
		if err := json.Unmarshal(encoded, &decoded); err != nil {
			t.Fatal(err)
		}
		if logprivacy.ReadIntentSource(decoded.DiagnosticLikeOperand()) != nil {
			t.Fatal("wire input invented trusted original operand")
		}
		clone.Right.Value = ValText("changed%")
		if logprivacy.ReadIntentSource(clone.DiagnosticLikeOperand()) != nil {
			t.Fatal("rewritten binding retained stale original")
		}
		if logprivacy.ReadIntentSource(expr.DiagnosticLikeOperand()) == nil {
			t.Fatal("clone mutated original provenance")
		}
		clone = expr.Clone()
		clone.Op = OpEq
		if logprivacy.ReadIntentSource(clone.DiagnosticLikeOperand()) != nil {
			t.Fatal("rewritten operator retained stale original")
		}
	}
}

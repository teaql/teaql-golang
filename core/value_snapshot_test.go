package core

import (
	"reflect"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestCloneValueOwnsNestedSupportedContainers(t *testing.T) {
	bytes := []byte("original")
	object := map[string]any{"children": []any{Record{"secret": ValText("original")}, bytes}}
	values := []Value{ValJson(object)}
	source := Record{"payload": ValJson(values)}
	snapshot := CloneRecord(source)
	object["children"].([]any)[0].(Record)["secret"] = ValText("changed")
	bytes[0] = 'x'
	values = append(values, ValText("extra"))
	source["payload"] = ValJson(values)
	owned := snapshot["payload"].V.([]Value)[0].V.(map[string]any)["children"].([]any)
	if owned[0].(Record)["secret"].V != "original" || string(owned[1].([]byte)) != "original" {
		t.Fatal("snapshot retained a mutable source container")
	}
	owned[0].(Record)["secret"] = ValText("snapshot only")
	if object["children"].([]any)[0].(Record)["secret"].V != "changed" {
		t.Fatal("source was modified through the snapshot")
	}
}

func TestCloneValuePreservesScalarTypesAndNilContainers(t *testing.T) {
	values := []Value{ValNull(), ValTypedNull(TypeText), ValBool(true), ValU64(42), ValI64(-42),
		ValF64(3.5), ValDecimal(decimal.RequireFromString("12.34")), ValText("hello"),
		ValDate(time.Date(2026, 10, 3, 0, 0, 0, 0, time.UTC)), ValTimestamp(123456),
		ValJson(Record(nil)), ValJson(map[string]any(nil)), ValJson([]Value(nil)),
		ValJson([]any(nil)), ValJson([]byte(nil)), ValJson([]byte{})}
	for _, value := range values {
		if !reflect.DeepEqual(CloneValue(value), value) {
			t.Fatalf("clone changed scalar/container semantics: %#v", value)
		}
	}
	if CloneRecord(nil) != nil {
		t.Fatal("nil record became an empty map")
	}
}

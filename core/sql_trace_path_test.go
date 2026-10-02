package core

import (
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

//go:embed testdata/sql-trace-path-v1.json
var sqlTraceVectors []byte

func TestSharedSQLTracePathVectors(t *testing.T) {
	digest := sha256.Sum256(sqlTraceVectors)
	if hex.EncodeToString(digest[:]) != "7cb67eb1fd08a723611e9f8fad121e060a17546c3b01640052346498d5a6b0b7" {
		t.Fatal("frozen SQL trace fixture drifted")
	}
	type frame struct {
		Kind, Name, Detail string
		EntityID           *uint64 `json:"entityId"`
	}
	var fixture struct {
		Cases []struct {
			ID, Operation, Backend string
			Source, ExpectedPath   []frame
			ExpectedIntent         struct{ Comment, Purpose, AuditReason *string }
		}
	}
	if err := json.Unmarshal(sqlTraceVectors, &fixture); err != nil {
		t.Fatal(err)
	}
	if len(fixture.Cases) != 12 {
		t.Fatal("wrong SQL trace vector inventory")
	}
	for _, vector := range fixture.Cases {
		t.Run(vector.ID, func(t *testing.T) {
			convert := func(frames []frame) []*TraceNode {
				result := make([]*TraceNode, len(frames))
				for i, value := range frames {
					node := NewTypedTraceNode(strings.ToLower(value.Kind), value.Name, value.Detail)
					node.EntityId = value.EntityID
					result[i] = node
				}
				return result
			}
			source := convert(vector.Source)
			before := CloneTraceNodes(source)
			result := CanonicalSQLTracePath(source, vector.Backend, vector.Operation)
			if !reflect.DeepEqual(result.TraceChain, convert(vector.ExpectedPath)) {
				t.Fatalf("path mismatch: %v", result.TraceChain)
			}
			if !reflect.DeepEqual(result.Comment, vector.ExpectedIntent.Comment) || !reflect.DeepEqual(result.Purpose, vector.ExpectedIntent.Purpose) ||
				!reflect.DeepEqual(result.AuditReason, vector.ExpectedIntent.AuditReason) {
				t.Fatalf("intent mismatch: %+v", result)
			}
			again := CanonicalSQLTracePath(result.TraceChain, "ignored", vector.Operation)
			if !reflect.DeepEqual(again.TraceChain, result.TraceChain) {
				t.Fatal("canonical path is not idempotent")
			}
			result.TraceChain[0].Name = "changed consumer view"
			if !reflect.DeepEqual(source, before) {
				t.Fatal("canonicalization or consumer mutated source")
			}
		})
	}
}

func TestTraceCloneOwnsOptionalIdentity(t *testing.T) {
	id := uint64(100)
	source := []*TraceNode{NewTraceNode("Payment", &id, "authorize")}
	copy := CloneTraceNodes(source)
	*copy[0].EntityId = 200
	if *source[0].EntityId != 100 {
		t.Fatal("ID pointer aliases source")
	}
}

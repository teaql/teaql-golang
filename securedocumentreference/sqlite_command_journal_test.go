package securedocumentreference

import "testing"

func TestDocumentCommandSelectionDigestUsesSharedCanonicalKeys(t *testing.T) {
	command := SubmittedDocumentCommand{
		BusinessID: "ORD-10", DocumentToken: "tqd2.fixture", CommandID: "command-1",
		Body: map[string]any{"quantity": 3, "note": "ok"},
		BlobSelections: []DocumentBlobSelectionSubmission{{
			Token: "tqs1.fixture", TargetEntity: "OrderItem", CanonicalField: "attachment",
			Kind: "EXISTING_ROW", TargetKey: "item:0",
		}},
	}
	canonical, frozen, err := prepareDocumentCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	const expected = "6b001bdca6e167373b752cd47a4590574945dbcf9f12ed7c88737cf6b189bedb"
	if canonical != expected {
		t.Fatalf("selection digest differs from shared canonical JSON: %s", canonical)
	}
	legacy, err := legacyDocumentSelectionHash(frozen)
	if err != nil || legacy == canonical {
		t.Fatalf("historical PascalCase selection digest unavailable: %s %v", legacy, err)
	}
}

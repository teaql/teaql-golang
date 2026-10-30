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
	legacy, err := legacyDocumentCommandHash(frozen)
	if err != nil || legacy == canonical {
		t.Fatalf("historical PascalCase selection digest unavailable: %s %v", legacy, err)
	}
	command.Body["note"] = "合同 & 📄\u2028\u2029"
	unicodeHash, unicodeFrozen, err := prepareDocumentCommand(command)
	if err != nil || unicodeHash != "0072aa4ae7f3ec7d16e5e910d3e85bb523f11d272014c854ea0ba724129fba46" {
		t.Fatalf("Unicode body digest differs from shared UTF-8 JSON: %s %v", unicodeHash, err)
	}
	command.Body["note"] = "literal \\u2028 &"
	literalHash, _, err := prepareDocumentCommand(command)
	if err != nil || literalHash != "9262d35bc97288590e743bdfe6398bc7f7802181413b1a67ec09febbe6b6e008" {
		t.Fatalf("literal backslash-u text changed by separator normalization: %s %v", literalHash, err)
	}
	command.Body["note"] = "合同 & 📄\u2028\u2029"
	command.BlobSelections = nil
	withoutSelection, noSelectionFrozen, err := prepareDocumentCommand(command)
	if err != nil {
		t.Fatal(err)
	}
	oldUnicode, err := legacyDocumentCommandHash(noSelectionFrozen)
	if err != nil || oldUnicode == withoutSelection {
		t.Fatalf("historical escaped Unicode receipt unavailable: %s %v", oldUnicode, err)
	}
	oldSelected, err := legacyDocumentCommandHash(unicodeFrozen)
	if err != nil || oldSelected == unicodeHash {
		t.Fatalf("historical selection/Unicode receipt unavailable: %s %v", oldSelected, err)
	}
}

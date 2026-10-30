package core

import (
	"encoding/json"
	"os"
	"reflect"
	"sort"
	"testing"
)

type blobValueVectors struct {
	Valid       []string          `json:"valid"`
	Invalid     []string          `json:"invalid"`
	InvalidJSON []json.RawMessage `json:"invalidJson"`
	Formats     map[string]string `json:"formats"`
}

func TestBlobRefSharedVectorsAndSQLText(t *testing.T) {
	data, err := os.ReadFile("../test-vectors/blob-value-cases.json")
	if err != nil {
		t.Fatal(err)
	}
	var vectors blobValueVectors
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, key := range vectors.Valid {
		reference, err := ParseBlobRef(key)
		if err != nil {
			t.Fatalf("valid key rejected: %v", err)
		}
		if reference.String() != "BlobRef[masked]" {
			t.Fatal("string exposed Key")
		}
		if _, err := json.Marshal(reference); err == nil {
			t.Fatal("JSON exposed Key")
		}
		value, err := reference.Value()
		if err != nil || value != key {
			t.Fatal("SQL text mismatch")
		}
		var readback BlobRef
		if err := readback.Scan(key); err != nil || readback != reference {
			t.Fatal("SQL readback mismatch")
		}
		encoded, _ := json.Marshal(key)
		if err := readback.UnmarshalJSON(encoded); err != nil || readback != reference {
			t.Fatal("JSON parse mismatch")
		}
	}
	for _, key := range vectors.Invalid {
		if _, err := ParseBlobRef(key); err == nil {
			t.Fatal("invalid Key accepted")
		}
		var reference BlobRef
		if err := reference.Scan(key); err == nil {
			t.Fatal("invalid SQL Key accepted")
		}
	}
	for _, value := range vectors.InvalidJSON {
		var reference BlobRef
		if err := json.Unmarshal(value, &reference); err == nil {
			t.Fatal("invalid JSON accepted")
		}
	}
	var extensions []string
	var contentTypes []string
	for extension, mime := range vectors.Formats {
		extensions = append(extensions, extension)
		contentTypes = append(contentTypes, mime)
	}
	spec, err := NewBlobFieldSpec("Asset", "document", extensions, false, "default")
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(extensions)
	if !reflect.DeepEqual(spec.AllowedExtensions(), extensions) {
		t.Fatal("format set mismatch")
	}
	uniqueMIME := make(map[string]struct{})
	for _, mime := range contentTypes {
		uniqueMIME[mime] = struct{}{}
	}
	var expectedMIME []string
	for mime := range uniqueMIME {
		expectedMIME = append(expectedMIME, mime)
	}
	sort.Strings(expectedMIME)
	if !reflect.DeepEqual(spec.AllowedContentTypes(), expectedMIME) {
		t.Fatal("MIME set mismatch")
	}
	if _, err := NewBlobFieldSpec("Asset", "document", []string{"application.docx"}, false, "default"); err == nil {
		t.Fatal("MIME sample incorrectly accepted as extension")
	}
}

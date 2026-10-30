package blobreference

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func officeFixture(t *testing.T, extension string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sample."+extension))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func officeVerifier(t *testing.T) *OfficeContentVerifier {
	t.Helper()
	v, err := NewOfficeContentVerifier(10<<20, 100<<20, 1000, 200000, 1)
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func assertOfficeError(t *testing.T, verifier *OfficeContentVerifier, data []byte,
	extension string, maximum int64, code string) {
	t.Helper()
	_, err := verifier.Verify(bytes.NewReader(data), extension, maximum)
	if err == nil || err.Error() != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}

func rewriteOfficePart(t *testing.T, data []byte, name, replacement string) []byte {
	t.Helper()
	input, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	changed := false
	for _, file := range input.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			t.Fatalf("read fixture part: %v %v", readErr, closeErr)
		}
		if file.Name == name {
			content = []byte(replacement)
			changed = true
		}
		entry, err := writer.Create(file.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatal(err)
		}
	}
	if !changed {
		t.Fatalf("missing fixture part %s", name)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func duplicateOfficePart(t *testing.T, data []byte, name string) []byte {
	t.Helper()
	input, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	writer := zip.NewWriter(&output)
	var duplicate []byte
	for _, file := range input.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		content, err := io.ReadAll(reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		entry, err := writer.Create(file.Name)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := entry.Write(content); err != nil {
			t.Fatal(err)
		}
		if file.Name == name {
			duplicate = content
		}
	}
	if duplicate == nil {
		t.Fatalf("missing fixture part %s", name)
	}
	entry, err := writer.Create(name)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := entry.Write(duplicate); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	return output.Bytes()
}

func TestOfficeContentVerifierParsesRealPackagesAndRejectsMalformedInputs(t *testing.T) {
	v := officeVerifier(t)
	for _, entry := range []struct{ extension, mime string }{
		{"docx", "application/vnd.openxmlformats-officedocument.wordprocessingml.document"},
		{"xlsx", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
	} {
		data := officeFixture(t, entry.extension)
		got, err := v.Verify(bytes.NewReader(data), entry.extension, int64(len(data)))
		if err != nil {
			t.Fatalf("real %s: %v", entry.extension, err)
		}
		digest := sha256.Sum256(data)
		if got.SizeBytes != int64(len(data)) || got.SHA256 != hex.EncodeToString(digest[:]) ||
			got.ContentType != entry.mime {
			t.Fatalf("wrong %s verified facts: %+v", entry.extension, got)
		}
		wrong := "xlsx"
		if entry.extension == "xlsx" {
			wrong = "docx"
		}
		assertOfficeError(t, v, data, wrong, int64(len(data)), "BLOB_FORMAT_NOT_ALLOWED")
		assertOfficeError(t, v, data[:len(data)-12], entry.extension, int64(len(data)), "BLOB_FORMAT_NOT_ALLOWED")
		assertOfficeError(t, v, data, entry.extension, int64(len(data)-1), "BLOB_SIZE_EXCEEDED")
		assertOfficeError(t, v, append(append([]byte(nil), data...), []byte("junk")...),
			entry.extension, int64(len(data)+4), "BLOB_FORMAT_NOT_ALLOWED")
	}
	docx := officeFixture(t, "docx")
	xlsx := officeFixture(t, "xlsx")
	corrupt := append([]byte(nil), docx...)
	member, err := zip.NewReader(bytes.NewReader(corrupt), int64(len(corrupt)))
	if err != nil {
		t.Fatal(err)
	}
	for _, file := range member.File {
		if file.Name == "word/document.xml" {
			offset, err := file.DataOffset()
			if err != nil {
				t.Fatal(err)
			}
			corrupt[int(offset+int64(file.CompressedSize64/2))] ^= 1
			break
		}
	}
	assertOfficeError(t, v, corrupt, "docx", int64(len(corrupt)), "BLOB_FORMAT_NOT_ALLOWED")
	assertOfficeError(t, v, duplicateOfficePart(t, docx, "word/document.xml"),
		"docx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")
	assertOfficeError(t, v, []byte("PK\x03\x04fake"), "docx", 100, "BLOB_FORMAT_NOT_ALLOWED")
	assertOfficeError(t, v, rewriteOfficePart(t, docx, "word/document.xml",
		`<!DOCTYPE document [<!ENTITY x SYSTEM "file:///etc/passwd">]><document>&x;</document>`),
		"docx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")
	assertOfficeError(t, v, rewriteOfficePart(t, docx, "word/document.xml",
		`<document xmlns="http://schemas.openxmlformats.org/wordprocessingml/2006/main"/>`),
		"docx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")
	assertOfficeError(t, v, rewriteOfficePart(t, xlsx, "xl/worksheets/sheet1.xml",
		`<not-a-worksheet/>`), "xlsx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")
	assertOfficeError(t, v, rewriteOfficePart(t, xlsx, "xl/_rels/workbook.xml.rels",
		`<Relationships xmlns="http://schemas.openxmlformats.org/package/2006/relationships"/>`),
		"xlsx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")

	limited, err := NewOfficeContentVerifier(10<<20, 100, 1000, 200000, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertOfficeError(t, limited, docx, "docx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")
	limited, err = NewOfficeContentVerifier(10<<20, 100<<20, 1, 200000, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertOfficeError(t, limited, docx, "docx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")
	limited, err = NewOfficeContentVerifier(10<<20, 100<<20, 1000, 1, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertOfficeError(t, limited, docx, "docx", 10<<20, "BLOB_FORMAT_NOT_ALLOWED")
	if _, err := v.Verify(unreadableCompositeSource{}, "pdf", 1000); err == nil ||
		err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
		t.Fatal("unknown extension must reject before reading")
	}
}

func TestOfficeContentVerifierRejectsDuplicateAndTraversalParts(t *testing.T) {
	for _, name := range []string{"../word/document.xml", "word/../document.xml", "word/document.xml"} {
		var output bytes.Buffer
		writer := zip.NewWriter(&output)
		for _, part := range []string{"[Content_Types].xml", "_rels/.rels", name, name} {
			entry, err := writer.Create(part)
			if err != nil {
				t.Fatal(err)
			}
			if _, err := io.Copy(entry, strings.NewReader("<root/>")); err != nil {
				t.Fatal(err)
			}
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		assertOfficeError(t, officeVerifier(t), output.Bytes(), "docx", 10<<20,
			"BLOB_FORMAT_NOT_ALLOWED")
	}
}

func TestOfficeContentVerifierReleasesBusySlot(t *testing.T) {
	v := officeVerifier(t)
	data := officeFixture(t, "docx")
	reader := &pausedImageReader{input: bytes.NewReader(data), entered: make(chan struct{}),
		release: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := v.Verify(reader, "docx", int64(len(data))); done <- err }()
	<-reader.entered
	assertOfficeError(t, v, data, "docx", int64(len(data)), "BLOB_VERIFICATION_BUSY")
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := v.Verify(bytes.NewReader(data), "docx", int64(len(data))); err != nil {
		t.Fatal(err)
	}
}

func TestCompositeContentVerifierRoutesRealImageAndOffice(t *testing.T) {
	image, err := NewImageContentVerifier(10<<20, 100<<20, 1)
	if err != nil {
		t.Fatal(err)
	}
	office := officeVerifier(t)
	composite, err := NewCompositeContentVerifier(
		ContentVerifierRoute{Extensions: []string{"jpg", "png"}, Verifier: image},
		ContentVerifierRoute{Extensions: []string{"docx", "xlsx"}, Verifier: office})
	if err != nil {
		t.Fatal(err)
	}
	if err := composite.ValidateReady(); err != nil {
		t.Fatal(err)
	}
	for _, extension := range []string{"jpg", "docx", "xlsx"} {
		data, err := os.ReadFile(filepath.Join("testdata", "sample."+extension))
		if err != nil {
			t.Fatal(err)
		}
		got, err := composite.Verify(bytes.NewReader(data), extension, int64(len(data)))
		if err != nil || got.SizeBytes != int64(len(data)) {
			t.Fatalf("composite %s failed: %+v %v", extension, got, err)
		}
	}
	if _, err := composite.Verify(unreadableCompositeSource{}, "pdf", 1000); err == nil ||
		err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
		t.Fatal("unknown extension must reject before reading")
	}
}

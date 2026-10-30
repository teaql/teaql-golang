package blobreference

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

func pdfTools(t *testing.T) (string, string, string) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux prlimit PDF verifier")
	}
	info, err := exec.LookPath("pdfinfo")
	if err != nil {
		t.Skip("pdfinfo unavailable")
	}
	render, err := exec.LookPath("pdftoppm")
	if err != nil {
		t.Skip("pdftoppm unavailable")
	}
	limit, err := exec.LookPath("prlimit")
	if err != nil {
		t.Skip("prlimit unavailable")
	}
	return info, render, limit
}

func pdfVerifier(t *testing.T, pages int, rendered int64, concurrency int) *PDFContentVerifier {
	t.Helper()
	info, render, limit := pdfTools(t)
	verifier, err := NewPDFContentVerifier(info, render, limit,
		10<<20, pages, rendered, concurrency)
	if err != nil {
		t.Fatal(err)
	}
	return verifier
}

func pdfFixture(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func assertPDFError(t *testing.T, verifier *PDFContentVerifier, data []byte,
	extension string, maximum int64, code string) {
	t.Helper()
	_, err := verifier.Verify(bytes.NewReader(data), extension, maximum)
	if err == nil || err.Error() != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}

func TestPDFContentVerifierRendersRealPagesAndRejectsMalformedInputs(t *testing.T) {
	verifier := pdfVerifier(t, 100, 4<<20, 1)
	if err := verifier.ValidateReady(); err != nil {
		t.Fatal(err)
	}
	data := pdfFixture(t, "sample.pdf")
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) !=
		"4b6a645a3ed4736baaa092cffa6e06fd3176866f589edd2b748565ad6487731b" {
		t.Fatal("shared PDF fixture digest changed")
	}
	got, err := verifier.Verify(bytes.NewReader(data), "pdf", int64(len(data)))
	if err != nil || got.SizeBytes != int64(len(data)) ||
		got.SHA256 != hex.EncodeToString(digest[:]) || got.ContentType != "application/pdf" {
		t.Fatalf("real PDF facts: %+v %v", got, err)
	}
	for _, invalid := range [][]byte{data[:len(data)-2], append(append([]byte(nil), data...), 'x'),
		[]byte("%PDF-1.7\nstartxref\n0\n%%EOF"), nil} {
		maximum := int64(len(invalid))
		if maximum == 0 {
			maximum = 1
		}
		assertPDFError(t, verifier, invalid, "pdf", maximum, "BLOB_FORMAT_NOT_ALLOWED")
	}
	assertPDFError(t, verifier, data, "docx", int64(len(data)), "BLOB_FORMAT_NOT_ALLOWED")
	assertPDFError(t, verifier, data, "pdf", int64(len(data)-1), "BLOB_SIZE_EXCEEDED")
	limited := pdfVerifier(t, 100, 1, 1)
	assertPDFError(t, limited, data, "pdf", int64(len(data)),
		"BLOB_FORMAT_NOT_ALLOWED")
	two := pdfFixture(t, "two-page.pdf")
	if _, err := verifier.Verify(bytes.NewReader(two), "pdf", int64(len(two))); err != nil {
		t.Fatalf("valid two-page PDF: %v", err)
	}
	assertPDFError(t, pdfVerifier(t, 1, 4<<20, 1), two, "pdf", int64(len(two)),
		"BLOB_FORMAT_NOT_ALLOWED")
}

func TestParserOutputCapCannotBeBypassedByReaderFrom(t *testing.T) {
	out := &cappedOutput{maximum: 4}
	_, err := io.Copy(out, io.LimitReader(bytes.NewReader([]byte("123456")), 6))
	if err == nil || len(out.Bytes()) > 4 {
		t.Fatalf("parser output cap bypassed: %d bytes, error %v", len(out.Bytes()), err)
	}
}

func TestPDFContentVerifierReadinessBusySlotAndComposite(t *testing.T) {
	verifier := pdfVerifier(t, 100, 4<<20, 1)
	data := pdfFixture(t, "sample.pdf")
	reader := &pausedImageReader{input: bytes.NewReader(data), entered: make(chan struct{}),
		release: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := verifier.Verify(reader, "pdf", int64(len(data))); done <- err }()
	<-reader.entered
	assertPDFError(t, verifier, data, "pdf", int64(len(data)), "BLOB_VERIFICATION_BUSY")
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	info, _, limit := pdfTools(t)
	missing, err := NewPDFContentVerifier(info, filepath.Join(t.TempDir(), "missing"), limit,
		10<<20, 100, 4<<20, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := missing.ValidateReady(); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
		t.Fatalf("missing PDF renderer readiness: %v", err)
	}
	image, err := NewImageContentVerifier(10<<20, 100<<20, 1)
	if err != nil {
		t.Fatal(err)
	}
	composite, err := NewCompositeContentVerifier(
		ContentVerifierRoute{Extensions: []string{"jpg"}, Verifier: image},
		ContentVerifierRoute{Extensions: []string{"pdf"}, Verifier: verifier})
	if err != nil || composite.ValidateReady() != nil {
		t.Fatalf("composite readiness: %v", err)
	}
	if got, err := composite.Verify(bytes.NewReader(data), "pdf", int64(len(data))); err != nil ||
		got.ContentType != "application/pdf" {
		t.Fatalf("PDF composite route: %+v %v", got, err)
	}
}

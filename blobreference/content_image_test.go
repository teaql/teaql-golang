package blobreference

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"testing"
)

func realImage(t *testing.T, extension string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sample."+extension))
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func TestImageContentVerifierParsesRealPixelsAndRejectsMalformedInputs(t *testing.T) {
	v, err := NewImageContentVerifier(10000, 100<<20, 2)
	if err != nil {
		t.Fatal(err)
	}
	for _, extension := range []string{"jpg", "png"} {
		data := realImage(t, extension)
		got, err := v.Verify(bytes.NewReader(data), extension, 10000)
		if err != nil {
			t.Fatal(err)
		}
		digest := sha256.Sum256(data)
		mime := "image/jpeg"
		if extension == "png" {
			mime = "image/png"
		}
		if got.SizeBytes != int64(len(data)) || got.SHA256 != hex.EncodeToString(digest[:]) || got.ContentType != mime {
			t.Fatalf("incorrect verified facts for %s: %+v", extension, got)
		}
		wrong := "png"
		if extension == "png" {
			wrong = "jpg"
		}
		assertImageError(t, v, data, wrong, 10000, "BLOB_FORMAT_NOT_ALLOWED")
		assertImageError(t, v, data[:len(data)-12], extension, 10000, "BLOB_FORMAT_NOT_ALLOWED")
		assertImageError(t, v, data, extension, int64(len(data)-1), "BLOB_SIZE_EXCEEDED")
	}
	corrupt := append([]byte(nil), realImage(t, "png")...)
	corrupt[len(corrupt)-25] ^= 1
	assertImageError(t, v, corrupt, "png", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	assertImageError(t, v, []byte("not an image"), "jpg", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	assertImageError(t, v, nil, "png", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	assertImageError(t, v, realImage(t, "png"), "pdf", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	limited, err := NewImageContentVerifier(10000, 8, 1)
	if err != nil {
		t.Fatal(err)
	}
	assertImageError(t, limited, realImage(t, "png"), "png", 10000, "BLOB_FORMAT_NOT_ALLOWED")
}

func assertImageError(t *testing.T, v *ImageContentVerifier, data []byte, extension string,
	maximum int64, expected string) {
	t.Helper()
	_, err := v.Verify(bytes.NewReader(data), extension, maximum)
	if err == nil || err.Error() != expected {
		t.Fatalf("expected %s; got %v", expected, err)
	}
}

type pausedImageReader struct {
	input   io.Reader
	entered chan struct{}
	release chan struct{}
}

func (r *pausedImageReader) Read(p []byte) (int, error) {
	select {
	case <-r.entered:
	default:
		close(r.entered)
	}
	<-r.release
	return r.input.Read(p)
}

func TestImageContentVerifierBusySlotReleasesAfterDecode(t *testing.T) {
	v, err := NewImageContentVerifier(10000, 100<<20, 1)
	if err != nil {
		t.Fatal(err)
	}
	data := realImage(t, "jpg")
	reader := &pausedImageReader{input: bytes.NewReader(data), entered: make(chan struct{}), release: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := v.Verify(reader, "jpg", 10000); done <- err }()
	<-reader.entered
	assertImageError(t, v, data, "jpg", 10000, "BLOB_VERIFICATION_BUSY")
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := v.Verify(bytes.NewReader(data), "jpg", 10000); err != nil {
		t.Fatal(err)
	}
}

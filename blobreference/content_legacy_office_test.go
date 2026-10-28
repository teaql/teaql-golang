package blobreference

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func legacyOfficeTools(t *testing.T) (string, string, string) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux bubblewrap legacy Office verifier")
	}
	var paths [3]string
	for index, name := range []string{"soffice", "bwrap", "prlimit"} {
		paths[index] = filepath.Join("/usr/bin", name)
		if _, err := os.Stat(paths[index]); err != nil {
			t.Skip(name + " unavailable")
		}
	}
	return paths[0], paths[1], paths[2]
}

func legacyOfficeVerifier(t *testing.T, outputLimit int64,
	concurrency int) (*LegacyOfficeContentVerifier, string) {
	t.Helper()
	soffice, bwrap, prlimit := legacyOfficeTools(t)
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	verifier, err := NewLegacyOfficeContentVerifier(soffice, bwrap, prlimit,
		root, 100_000, outputLimit, concurrency)
	if err != nil {
		t.Fatal(err)
	}
	return verifier, root
}

func TestLegacyOfficeContentVerifierConvertsAndChecksRealCFB(t *testing.T) {
	verifier, root := legacyOfficeVerifier(t, 100_000, 1)
	if err := verifier.ValidateReady(); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct{ extension, mime, digest string }{
		{"doc", "application/msword", "9fca8d262d2aa905d58b1cc1c8914d1a8069c16c340f24cd3c4a77daff9c8b73"},
		{"xls", "application/vnd.ms-excel", "7f387858aaa3ed39d3c7c692aa865b36ac6620462714f3d0daa96388424cf75d"},
	} {
		t.Run(entry.extension, func(t *testing.T) {
			data, err := os.ReadFile(filepath.Join("testdata", "sample."+entry.extension))
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(data)
			if hex.EncodeToString(digest[:]) != entry.digest {
				t.Fatal("shared legacy Office fixture changed")
			}
			got, err := verifier.Verify(bytes.NewReader(data), entry.extension, int64(len(data)))
			if err != nil || got.SizeBytes != int64(len(data)) || got.SHA256 != entry.digest ||
				got.ContentType != entry.mime {
				t.Fatalf("legacy Office facts: %+v %v", got, err)
			}
			for _, bad := range [][]byte{data[:512], append([]byte(nil), data...)} {
				if len(bad) == len(data) {
					bad[0] ^= 1
				}
				_, err := verifier.Verify(bytes.NewReader(bad), entry.extension, int64(len(bad)))
				if err == nil || err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
					t.Fatalf("invalid legacy Office accepted: %v", err)
				}
			}
			_, err = verifier.Verify(bytes.NewReader(data), entry.extension, int64(len(data)-1))
			if err == nil || err.Error() != "BLOB_SIZE_EXCEEDED" {
				t.Fatalf("legacy Office input limit: %v", err)
			}
		})
	}
	entries, err := os.ReadDir(root)
	if err != nil || len(entries) != 0 {
		t.Fatalf("private conversion scratch not cleaned: %d entries %v", len(entries), err)
	}
}

func TestLegacyOfficeContentVerifierBoundsOutputBusyAndReadiness(t *testing.T) {
	data, err := os.ReadFile(filepath.Join("testdata", "sample.doc"))
	if err != nil {
		t.Fatal(err)
	}
	limited, _ := legacyOfficeVerifier(t, 1, 1)
	_, err = limited.Verify(bytes.NewReader(data), "doc", int64(len(data)))
	if err == nil || err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
		t.Fatalf("converted package output limit: %v", err)
	}
	verifier, root := legacyOfficeVerifier(t, 100_000, 1)
	reader := &pausedImageReader{input: bytes.NewReader(data), entered: make(chan struct{}),
		release: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := verifier.Verify(reader, "doc", int64(len(data))); done <- err }()
	<-reader.entered
	_, err = verifier.Verify(bytes.NewReader(data), "doc", int64(len(data)))
	if err == nil || err.Error() != "BLOB_VERIFICATION_BUSY" {
		t.Fatalf("busy legacy Office verifier admitted another upload: %v", err)
	}
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	soffice, bwrap, prlimit := legacyOfficeTools(t)
	missing, err := NewLegacyOfficeContentVerifier(filepath.Join(root, "missing"), bwrap,
		prlimit, root, 100_000, 100_000, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := missing.ValidateReady(); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
		t.Fatalf("missing converter readiness: %v", err)
	}
	unsafeRoot := filepath.Join(t.TempDir(), "unsafe")
	if err := os.Mkdir(unsafeRoot, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(unsafeRoot, 0755); err != nil {
		t.Fatal(err)
	}
	unsafe, err := NewLegacyOfficeContentVerifier(soffice, bwrap, prlimit,
		unsafeRoot, 100_000, 100_000, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := unsafe.ValidateReady(); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
		t.Fatalf("unsafe scratch root readiness: %v", err)
	}
	image, err := NewImageContentVerifier(10<<20, 100<<20, 1)
	if err != nil {
		t.Fatal(err)
	}
	composite, err := NewCompositeContentVerifier(
		ContentVerifierRoute{Extensions: []string{"jpg"}, Verifier: image},
		ContentVerifierRoute{Extensions: []string{"doc", "xls"}, Verifier: verifier})
	if err != nil || composite.ValidateReady() != nil {
		t.Fatalf("legacy Office composite readiness: %v", err)
	}
	if got, err := composite.Verify(bytes.NewReader(data), "doc", int64(len(data))); err != nil ||
		got.ContentType != "application/msword" {
		t.Fatalf("legacy Office composite route: %+v %v", got, err)
	}
}

package blobreference

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"testing"
)

type unreadableCompositeSource struct{}

func (unreadableCompositeSource) Read([]byte) (int, error) {
	return 0, errors.New("source must not be read")
}

func TestCompositeContentVerifierRealImageAndMedia(t *testing.T) {
	if runtime.GOOS != "linux" {
		t.Skip("media verifier uses Linux process limits")
	}
	for _, binary := range []string{"/usr/bin/ffprobe", "/usr/bin/ffmpeg", "/usr/bin/prlimit"} {
		if _, err := os.Stat(binary); err != nil {
			t.Skip("host media tools unavailable")
		}
	}
	image, err := NewImageContentVerifier(10<<20, 100<<20, 2)
	if err != nil {
		t.Fatal(err)
	}
	media, err := NewMediaContentVerifier("/usr/bin/ffprobe", "/usr/bin/ffmpeg",
		"/usr/bin/prlimit", 10<<20, 100000, 16000000, 2)
	if err != nil {
		t.Fatal(err)
	}
	verifier, err := NewCompositeContentVerifier(
		ContentVerifierRoute{Extensions: []string{"jpg", "jpeg", "png"}, Verifier: image},
		ContentVerifierRoute{Extensions: []string{"mp3", "mp4"}, Verifier: media})
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(verifier.SupportedExtensions(), []string{"jpeg", "jpg", "mp3", "mp4", "png"}) {
		t.Fatal("unexpected supported extensions")
	}
	if err := verifier.ValidateReady(); err != nil {
		t.Fatal(err)
	}
	for _, entry := range []struct{ extension, mime string }{
		{"jpg", "image/jpeg"}, {"mp3", "audio/mpeg"},
	} {
		data, err := os.ReadFile(filepath.Join("testdata", "sample."+entry.extension))
		if err != nil {
			t.Fatal(err)
		}
		actual, err := verifier.Verify(bytes.NewReader(data), entry.extension, int64(len(data)))
		digest := sha256.Sum256(data)
		if err != nil || actual.SizeBytes != int64(len(data)) ||
			actual.SHA256 != hex.EncodeToString(digest[:]) || actual.ContentType != entry.mime {
			t.Fatalf("real %s verifier result mismatch: %v", entry.extension, err)
		}
	}
	if _, err := verifier.Verify(unreadableCompositeSource{}, "docx", 1024); err == nil || err.Error() != "BLOB_FORMAT_NOT_ALLOWED" {
		t.Fatal("unregistered extension must reject before reading")
	}
	if _, err := NewCompositeContentVerifier(
		ContentVerifierRoute{Extensions: []string{"jpg"}, Verifier: image},
		ContentVerifierRoute{Extensions: []string{"jpg"}, Verifier: media}); err == nil {
		t.Fatal("overlapping routes must reject")
	}
	var missing *MediaContentVerifier
	if _, err := NewCompositeContentVerifier(ContentVerifierRoute{Extensions: []string{"mp3"}, Verifier: missing}); err == nil {
		t.Fatal("typed nil verifier must reject")
	}
}

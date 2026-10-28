package blobreference

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
	"time"
)

func mediaTools(t *testing.T) (string, string, string) {
	t.Helper()
	if runtime.GOOS != "linux" {
		t.Skip("Linux prlimit media verifier")
	}
	ffprobe, err := exec.LookPath("ffprobe")
	if err != nil {
		t.Skip("ffprobe not installed")
	}
	ffmpeg, err := exec.LookPath("ffmpeg")
	if err != nil {
		t.Skip("ffmpeg not installed")
	}
	prlimit, err := exec.LookPath("prlimit")
	if err != nil {
		t.Skip("prlimit not installed")
	}
	return ffprobe, ffmpeg, prlimit
}

func mediaFixture(t *testing.T, extension string) []byte {
	t.Helper()
	data, err := os.ReadFile(filepath.Join("testdata", "sample."+extension))
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{
		"mp3": "a4facfc9e36b342c3f7b2de8283ab9f4c9cf53af2dfb2f1959825aa930748d9b",
		"mp4": "5a3665235df157cb9d1cffe53501b86889849a38a5a643f0a64dbc957f4a58af",
	}[extension]
	digest := sha256.Sum256(data)
	if hex.EncodeToString(digest[:]) != want {
		t.Fatal("shared media fixture digest changed")
	}
	return data
}

func mediaVerifier(t *testing.T, maximumFrames int, maximumPixels int64,
	concurrency int) *MediaContentVerifier {
	t.Helper()
	ffprobe, ffmpeg, prlimit := mediaTools(t)
	verifier, err := NewMediaContentVerifier(ffprobe, ffmpeg, prlimit,
		10000, maximumFrames, maximumPixels, concurrency)
	if err != nil {
		t.Fatal(err)
	}
	return verifier
}

func assertMediaError(t *testing.T, verifier *MediaContentVerifier,
	data []byte, extension string, maximum int64, code string) {
	t.Helper()
	_, err := verifier.Verify(bytes.NewReader(data), extension, maximum)
	if err == nil || err.Error() != code {
		t.Fatalf("expected %s, got %v", code, err)
	}
}

func TestMediaContentVerifierRealDecodeAndNegativeControls(t *testing.T) {
	verifier := mediaVerifier(t, 100000, 16000000, 2)
	if err := verifier.ValidateReady(); err != nil {
		t.Fatal(err)
	}
	for _, extension := range []string{"mp3", "mp4"} {
		data := mediaFixture(t, extension)
		got, err := verifier.Verify(bytes.NewReader(data), extension, int64(len(data)))
		if err != nil {
			t.Fatal(err)
		}
		mime := "audio/mpeg"
		if extension == "mp4" {
			mime = "video/mp4"
		}
		digest := sha256.Sum256(data)
		if got.SizeBytes != int64(len(data)) || got.SHA256 != hex.EncodeToString(digest[:]) ||
			got.ContentType != mime {
			t.Fatalf("wrong verified media facts for %s: %+v", extension, got)
		}
		for _, invalid := range [][]byte{data[:len(data)-1], data[:len(data)-12],
			data[:len(data)/2], nil, []byte("not media")} {
			assertMediaError(t, verifier, invalid, extension, int64(len(data)), "BLOB_FORMAT_NOT_ALLOWED")
		}
		assertMediaError(t, verifier, data, extension, int64(len(data)-1), "BLOB_SIZE_EXCEEDED")
	}
	assertMediaError(t, verifier, mediaFixture(t, "mp3"), "mp4", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	assertMediaError(t, verifier, mediaFixture(t, "mp4"), "mp3", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	assertMediaError(t, verifier, mediaFixture(t, "mp3"), "pdf", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	assertMediaError(t, mediaVerifier(t, 1, 16000000, 1), mediaFixture(t, "mp3"),
		"mp3", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	assertMediaError(t, mediaVerifier(t, 100000, 1, 1), mediaFixture(t, "mp4"),
		"mp4", 10000, "BLOB_FORMAT_NOT_ALLOWED")
	_, ffmpeg, prlimit := mediaTools(t)
	missing, err := NewMediaContentVerifier(filepath.Join(t.TempDir(), "missing"), ffmpeg,
		prlimit, 10000, 100000, 16000000, 1)
	if err != nil {
		t.Fatal(err)
	}
	if err := missing.ValidateReady(); err == nil || err.Error() != "BLOB_CAPABILITY_MISSING" {
		t.Fatalf("missing FFprobe readiness not denied: %v", err)
	}
}

func TestMediaContentVerifierRejectsAudioOnlyMP4(t *testing.T) {
	_, ffmpeg, _ := mediaTools(t)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, ffmpeg, "-nostdin", "-v", "error", "-f", "mp3",
		"-i", "pipe:0", "-vn", "-c:a", "copy", "-movflags", "frag_keyframe+empty_moov",
		"-f", "mp4", "pipe:1")
	command.Stdin = bytes.NewReader(mediaFixture(t, "mp3"))
	media, err := command.Output()
	if err != nil {
		t.Fatal(err)
	}
	assertMediaError(t, mediaVerifier(t, 100000, 16000000, 1), media,
		"mp4", int64(len(media)), "BLOB_FORMAT_NOT_ALLOWED")
}

func TestMediaContentVerifierBusySlotReleased(t *testing.T) {
	verifier := mediaVerifier(t, 100000, 16000000, 1)
	data := mediaFixture(t, "mp3")
	reader := &pausedImageReader{input: bytes.NewReader(data), entered: make(chan struct{}),
		release: make(chan struct{})}
	done := make(chan error, 1)
	go func() { _, err := verifier.Verify(reader, "mp3", 10000); done <- err }()
	<-reader.entered
	assertMediaError(t, verifier, data, "mp3", 10000, "BLOB_VERIFICATION_BUSY")
	close(reader.release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if _, err := verifier.Verify(bytes.NewReader(data), "mp3", 10000); err != nil {
		t.Fatal(err)
	}
}

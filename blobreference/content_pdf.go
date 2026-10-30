package blobreference

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"image/png"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// PDFContentVerifier is a passive Linux Poppler adapter. The host supplies
// patched, trusted absolute pdfinfo, pdftoppm and prlimit paths. The PDF enters
// both processes through stdin; decoded page images return through a bounded
// pipe, so the verifier never writes document bytes or renderings to scratch.
type PDFContentVerifier struct {
	pdfinfoPath, pdftoppmPath, prlimitPath string
	maximumInputBytes                      int64
	maximumPages                           int
	maximumRenderedBytes                   int64
	permits                                chan struct{}
}

func NewPDFContentVerifier(pdfinfoPath, pdftoppmPath, prlimitPath string,
	maximumInputBytes int64, maximumPages int, maximumRenderedBytes int64,
	concurrentVerifications int) (*PDFContentVerifier, error) {
	if !filepath.IsAbs(pdfinfoPath) || !filepath.IsAbs(pdftoppmPath) ||
		!filepath.IsAbs(prlimitPath) || maximumInputBytes < 1 || maximumInputBytes > 100<<20 ||
		maximumPages < 1 || maximumPages > 200 ||
		maximumRenderedBytes < 1 || maximumRenderedBytes > 100<<20 ||
		concurrentVerifications < 1 || concurrentVerifications > 32 {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	return &PDFContentVerifier{pdfinfoPath: pdfinfoPath, pdftoppmPath: pdftoppmPath,
		prlimitPath: prlimitPath, maximumInputBytes: maximumInputBytes,
		maximumPages: maximumPages, maximumRenderedBytes: maximumRenderedBytes,
		permits: make(chan struct{}, concurrentVerifications)}, nil
}

func (v *PDFContentVerifier) ValidateReady() error {
	if v == nil || runtime.GOOS != "linux" {
		return uploadError("BLOB_CAPABILITY_MISSING")
	}
	for _, path := range []string{v.pdfinfoPath, v.pdftoppmPath, v.prlimitPath} {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			return uploadError("BLOB_CAPABILITY_MISSING")
		}
	}
	for _, entry := range []struct{ path, prefix string }{
		{v.pdfinfoPath, "pdfinfo version "}, {v.pdftoppmPath, "pdftoppm version "},
	} {
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		cmd := exec.CommandContext(ctx, v.prlimitPath,
			"--cpu=2:2", "--as=536870912:536870912", "--", entry.path, "-v")
		cmd.Env = []string{"LC_ALL=C"}
		output := &cappedOutput{maximum: 4096}
		cmd.Stdout, cmd.Stderr = output, output
		err := cmd.Run()
		cancel()
		if err != nil || !bytes.HasPrefix(output.Bytes(), []byte(entry.prefix)) {
			return uploadError("BLOB_CAPABILITY_MISSING")
		}
	}
	return nil
}

func completePDF(data []byte) bool {
	if len(data) < 32 || !bytes.HasPrefix(data, []byte("%PDF-")) ||
		(data[5] != '1' && data[5] != '2') || data[6] != '.' ||
		data[7] < '0' || data[7] > '9' {
		return false
	}
	trimmed := bytes.TrimRight(data, " \t\r\n\f")
	return bytes.HasSuffix(trimmed, []byte("%%EOF")) &&
		bytes.Contains(trimmed[:len(trimmed)-5], []byte("startxref"))
}

func (v *PDFContentVerifier) run(ctx context.Context, binary string,
	arguments []string, data []byte, maximumOutput int) ([]byte, error) {
	args := append([]string{"--cpu=10:10", "--as=536870912:536870912",
		"--fsize=1048576:1048576", "--", binary}, arguments...)
	cmd := exec.CommandContext(ctx, v.prlimitPath, args...)
	cmd.Env = []string{"LC_ALL=C"}
	cmd.Stdin = bytes.NewReader(data)
	output := &cappedOutput{maximum: maximumOutput}
	cmd.Stdout = output
	cmd.Stderr = io.Discard
	if err := cmd.Run(); err != nil || ctx.Err() != nil {
		return nil, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	return output.Bytes(), nil
}

func pdfPageCount(output []byte, maximum int) (int, bool) {
	pages, encrypted := 0, ""
	for _, line := range strings.Split(string(output), "\n") {
		if strings.HasPrefix(line, "Pages:") {
			if pages != 0 {
				return 0, false
			}
			value, err := strconv.Atoi(strings.TrimSpace(strings.TrimPrefix(line, "Pages:")))
			if err != nil || value < 1 || value > maximum {
				return 0, false
			}
			pages = value
		} else if strings.HasPrefix(line, "Encrypted:") {
			if encrypted != "" {
				return 0, false
			}
			fields := strings.Fields(strings.TrimPrefix(line, "Encrypted:"))
			if len(fields) == 0 {
				return 0, false
			}
			encrypted = fields[0]
		}
	}
	return pages, pages > 0 && encrypted == "no"
}

func pdfRenderedPages(output []byte, expected int) bool {
	signature := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	offset, pages := 0, 0
	for offset < len(output) {
		if pages >= expected || len(output)-offset < 33 ||
			!bytes.Equal(output[offset:offset+8], signature) ||
			binary.BigEndian.Uint32(output[offset+8:]) != 13 ||
			!bytes.Equal(output[offset+12:offset+16], []byte("IHDR")) {
			return false
		}
		width, height := binary.BigEndian.Uint32(output[offset+16:]),
			binary.BigEndian.Uint32(output[offset+20:])
		if width == 0 || height == 0 || width > 64 || height > 64 {
			return false
		}
		position, chunks := offset+8, 0
		for {
			if chunks >= 1000 || len(output)-position < 12 {
				return false
			}
			size := uint64(binary.BigEndian.Uint32(output[position:]))
			if size > uint64(len(output)-position-12) {
				return false
			}
			kind := output[position+4 : position+8]
			position += int(size) + 12
			chunks++
			if bytes.Equal(kind, []byte("IEND")) {
				if size != 0 {
					return false
				}
				decoded, err := png.Decode(bytes.NewReader(output[offset:position]))
				if err != nil || decoded.Bounds().Dx() != int(width) ||
					decoded.Bounds().Dy() != int(height) {
					return false
				}
				offset = position
				pages++
				break
			}
		}
	}
	return pages == expected && offset == len(output)
}

func (v *PDFContentVerifier) Verify(source io.Reader, extension string,
	maximumBytes int64) (VerifiedContent, error) {
	if extension != "pdf" {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if v == nil || source == nil || maximumBytes < 1 {
		return VerifiedContent{}, uploadError("BLOB_POLICY_INVALID")
	}
	if runtime.GOOS != "linux" {
		return VerifiedContent{}, uploadError("BLOB_CAPABILITY_MISSING")
	}
	select {
	case v.permits <- struct{}{}:
		defer func() { <-v.permits }()
	default:
		return VerifiedContent{}, uploadError("BLOB_VERIFICATION_BUSY")
	}
	bound := maximumBytes
	if bound > v.maximumInputBytes {
		bound = v.maximumInputBytes
	}
	data, err := io.ReadAll(io.LimitReader(source, bound+1))
	if err != nil {
		return VerifiedContent{}, uploadError("BLOB_STORE_FAILED")
	}
	if int64(len(data)) > bound {
		return VerifiedContent{}, uploadError("BLOB_SIZE_EXCEEDED")
	}
	if !completePDF(data) {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	info, err := v.run(ctx, v.pdfinfoPath, []string{"-"}, data, 65536)
	if err != nil {
		return VerifiedContent{}, err
	}
	pages, valid := pdfPageCount(info, v.maximumPages)
	if !valid {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	rendered, err := v.run(ctx, v.pdftoppmPath, []string{"-f", "1", "-l",
		strconv.Itoa(pages), "-scale-to", "64", "-png", "-"}, data,
		int(v.maximumRenderedBytes))
	if err != nil || !pdfRenderedPages(rendered, pages) {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	digest := sha256.Sum256(data)
	return VerifiedContent{SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(digest[:]),
		ContentType: "application/pdf"}, nil
}

package blobreference

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"hash/crc32"
	"image/jpeg"
	"image/png"
	"io"
)

const imageInputCeiling int64 = 100 << 20
const imageDecodedCeiling int64 = 512 << 20

// ImageContentVerifier is a passive, bounded JPEG/PNG parser for the trusted
// upload verifier SPI. It does not run a worker or touch storage at construction.
type ImageContentVerifier struct {
	maximumInputBytes   int64
	maximumDecodedBytes int64
	permits             chan struct{}
}

func NewImageContentVerifier(maximumInputBytes, maximumDecodedBytes int64,
	concurrentVerifications int) (*ImageContentVerifier, error) {
	if maximumInputBytes < 1 || maximumInputBytes > imageInputCeiling ||
		maximumDecodedBytes < 8 || maximumDecodedBytes > imageDecodedCeiling ||
		concurrentVerifications < 1 || concurrentVerifications > 32 {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	return &ImageContentVerifier{maximumInputBytes, maximumDecodedBytes,
		make(chan struct{}, concurrentVerifications)}, nil
}

func (v *ImageContentVerifier) Verify(source io.Reader, extension string,
	maximumBytes int64) (VerifiedContent, error) {
	var mime string
	switch extension {
	case "jpg", "jpeg":
		mime = "image/jpeg"
	case "png":
		mime = "image/png"
	default:
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if v == nil || source == nil || maximumBytes < 1 {
		return VerifiedContent{}, uploadError("BLOB_POLICY_INVALID")
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
	if len(data) == 0 {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if mime == "image/jpeg" {
		if len(data) < 4 || !bytes.HasPrefix(data, []byte{0xff, 0xd8}) ||
			!bytes.HasSuffix(data, []byte{0xff, 0xd9}) {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
	} else if !completePNG(data) {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	var width, height int
	if mime == "image/jpeg" {
		config, err := jpeg.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		width, height = config.Width, config.Height
	} else {
		config, err := png.DecodeConfig(bytes.NewReader(data))
		if err != nil {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		width, height = config.Width, config.Height
	}
	// The standard decoders may allocate 16-bit RGBA storage (eight bytes/pixel).
	pixels := v.maximumDecodedBytes / 8
	if width < 1 || height < 1 || int64(width) > pixels/int64(height) {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if mime == "image/jpeg" {
		decoded, err := jpeg.Decode(bytes.NewReader(data))
		if err != nil || decoded.Bounds().Dx() != width || decoded.Bounds().Dy() != height {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
	} else {
		decoded, err := png.Decode(bytes.NewReader(data))
		if err != nil || decoded.Bounds().Dx() != width || decoded.Bounds().Dy() != height {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
	}
	digest := sha256.Sum256(data)
	return VerifiedContent{SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(digest[:]),
		ContentType: mime}, nil
}

func completePNG(data []byte) bool {
	if len(data) < 45 || !bytes.Equal(data[:8], []byte{137, 80, 78, 71, 13, 10, 26, 10}) {
		return false
	}
	offset, count, hasData := 8, 0, false
	for offset <= len(data)-12 {
		length := uint64(binary.BigEndian.Uint32(data[offset:]))
		if length > uint64(len(data)-offset-12) {
			return false
		}
		end := offset + 8 + int(length)
		kind := data[offset+4 : offset+8]
		for _, letter := range kind {
			if letter < 'A' || letter > 'Z' && letter < 'a' || letter > 'z' {
				return false
			}
		}
		if count == 0 && (!bytes.Equal(kind, []byte("IHDR")) || length != 13) {
			return false
		}
		if crc32.ChecksumIEEE(data[offset+4:end]) != binary.BigEndian.Uint32(data[end:]) {
			return false
		}
		offset = end + 4
		count++
		if bytes.Equal(kind, []byte("IDAT")) {
			hasData = true
		}
		if bytes.Equal(kind, []byte("IEND")) {
			return length == 0 && hasData && offset == len(data)
		}
	}
	return false
}

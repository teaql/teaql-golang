package blobreference

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// MediaContentVerifier is a passive Linux MP3/MP4 adapter. The host provides
// trusted absolute FFmpeg and prlimit binaries; no binary is downloaded.
type MediaContentVerifier struct {
	ffprobePath, ffmpegPath, prlimitPath string
	maximumInputBytes                    int64
	maximumFrames                        int
	maximumVideoPixels                   int64
	permits                              chan struct{}
}

func NewMediaContentVerifier(ffprobePath, ffmpegPath, prlimitPath string,
	maximumInputBytes int64, maximumFrames int, maximumVideoPixels int64,
	concurrentVerifications int) (*MediaContentVerifier, error) {
	if !filepath.IsAbs(ffprobePath) || !filepath.IsAbs(ffmpegPath) || !filepath.IsAbs(prlimitPath) ||
		maximumInputBytes < 1 || maximumInputBytes > 100<<20 ||
		maximumFrames < 1 || maximumFrames > 1000000 ||
		maximumVideoPixels < 1 || maximumVideoPixels > 64000000 ||
		concurrentVerifications < 1 || concurrentVerifications > 32 {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	return &MediaContentVerifier{ffprobePath: ffprobePath, ffmpegPath: ffmpegPath,
		prlimitPath: prlimitPath, maximumInputBytes: maximumInputBytes,
		maximumFrames: maximumFrames, maximumVideoPixels: maximumVideoPixels,
		permits: make(chan struct{}, concurrentVerifications)}, nil
}

type cappedOutput struct {
	bytes.Buffer
	maximum int
}

func (out *cappedOutput) Write(data []byte) (int, error) {
	if len(data) > out.maximum-out.Len() {
		return 0, errors.New("media probe output limit")
	}
	return out.Buffer.Write(data)
}

func (v *MediaContentVerifier) execute(binary string, arguments []string,
	input []byte, timeout time.Duration) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	args := append([]string{"--cpu=10:10", "--as=536870912:536870912", "--", binary}, arguments...)
	command := exec.CommandContext(ctx, v.prlimitPath, args...)
	command.Env = []string{"LC_ALL=C"}
	command.Stdin = bytes.NewReader(input)
	output := &cappedOutput{maximum: 65536}
	command.Stdout = output
	command.Stderr = io.Discard
	if err := command.Run(); err != nil || ctx.Err() != nil {
		return nil, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	return output.Bytes(), nil
}

// ValidateReady is an explicit module-readiness operation; constructor and
// package import do not start FFmpeg or touch storage.
func (v *MediaContentVerifier) ValidateReady() error {
	if v == nil || runtime.GOOS != "linux" {
		return uploadError("BLOB_CAPABILITY_MISSING")
	}
	for _, binary := range []string{v.ffprobePath, v.ffmpegPath, v.prlimitPath} {
		info, err := os.Stat(binary)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			return uploadError("BLOB_CAPABILITY_MISSING")
		}
	}
	for _, check := range []struct{ binary, prefix string }{
		{v.ffprobePath, "ffprobe version "}, {v.ffmpegPath, "ffmpeg version "},
	} {
		output, err := v.execute(check.binary, []string{"-version"}, nil, 3*time.Second)
		if err != nil || !bytes.HasPrefix(output, []byte(check.prefix)) {
			return uploadError("BLOB_CAPABILITY_MISSING")
		}
	}
	return nil
}

func completeMP3(data []byte, maximumFrames int) bool {
	offset := 0
	if bytes.HasPrefix(data, []byte("ID3")) {
		if len(data) < 10 || data[3] < 2 || data[3] > 4 || data[4] == 255 {
			return false
		}
		for _, value := range data[6:10] {
			if value&0x80 != 0 {
				return false
			}
		}
		tagSize := int(data[6])<<21 | int(data[7])<<14 | int(data[8])<<7 | int(data[9])
		offset = 10 + tagSize
		if data[3] == 4 && data[5]&0x10 != 0 {
			offset += 10
		}
		if offset > len(data) {
			return false
		}
	}
	end := len(data)
	if end >= 128 && bytes.Equal(data[end-128:end-125], []byte("TAG")) {
		end -= 128
	}
	frames, firstRate := 0, 0
	for offset < end {
		if end-offset < 4 || frames >= maximumFrames {
			return false
		}
		header := binary.BigEndian.Uint32(data[offset:])
		version, layer := int(header>>19&3), int(header>>17&3)
		bitrateIndex, sampleIndex := int(header>>12&15), int(header>>10&3)
		if header>>21 != 0x7ff || version == 1 || layer != 1 ||
			bitrateIndex == 0 || bitrateIndex == 15 || sampleIndex == 3 {
			return false
		}
		rate := []int{44100, 48000, 32000}[sampleIndex]
		if version == 2 {
			rate /= 2
		} else if version == 0 {
			rate /= 4
		}
		if firstRate != 0 && firstRate != rate {
			return false
		}
		firstRate = rate
		bitrates := []int{0, 32, 40, 48, 56, 64, 80, 96, 112, 128,
			160, 192, 224, 256, 320, 0}
		lengthScale := 144000
		if version != 3 {
			bitrates = []int{0, 8, 16, 24, 32, 40, 48, 56, 64, 80,
				96, 112, 128, 144, 160, 0}
			lengthScale = 72000
		}
		length := lengthScale*bitrates[bitrateIndex]/rate + int(header>>9&1)
		if length < 4 || length > end-offset {
			return false
		}
		offset += length
		frames++
	}
	return frames > 0 && offset == end
}

func completeMP4(data []byte) bool {
	brands := map[string]bool{"isom": true, "iso2": true, "iso5": true,
		"iso6": true, "mp41": true, "mp42": true, "avc1": true,
		"M4V ": true, "M4V2": true, "dash": true}
	offset, count := 0, 0
	kinds := make(map[string]bool)
	for offset < len(data) {
		if len(data)-offset < 8 || count >= 1000 {
			return false
		}
		size := uint64(binary.BigEndian.Uint32(data[offset:]))
		kind := string(data[offset+4 : offset+8])
		header := uint64(8)
		if size == 1 {
			if len(data)-offset < 16 {
				return false
			}
			size = binary.BigEndian.Uint64(data[offset+8:])
			header = 16
		} else if size == 0 {
			if kind != "mdat" {
				return false
			}
			size = uint64(len(data) - offset)
		}
		if size < header || size > uint64(len(data)-offset) {
			return false
		}
		if count == 0 && (kind != "ftyp" || size < header+8 ||
			!brands[string(data[offset+int(header):offset+int(header)+4])]) {
			return false
		}
		if (kind == "ftyp" || kind == "moov") && kinds[kind] {
			return false
		}
		kinds[kind] = true
		offset += int(size)
		count++
	}
	return kinds["ftyp"] && kinds["moov"] && kinds["mdat"] && offset == len(data)
}

type mediaProbe struct {
	Streams []struct {
		CodecType   string `json:"codec_type"`
		CodecName   string `json:"codec_name"`
		Frames      string `json:"nb_read_frames"`
		SampleRate  string `json:"sample_rate"`
		Width       int    `json:"width"`
		Height      int    `json:"height"`
		Disposition struct {
			AttachedPic int `json:"attached_pic"`
		} `json:"disposition"`
	} `json:"streams"`
	Format struct {
		Name string `json:"format_name"`
	} `json:"format"`
}

func (v *MediaContentVerifier) Verify(source io.Reader, extension string,
	maximumBytes int64) (VerifiedContent, error) {
	if extension != "mp3" && extension != "mp4" {
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
	demuxer := "mp3"
	mime := "audio/mpeg"
	valid := completeMP3(data, v.maximumFrames)
	if extension == "mp4" {
		demuxer, mime = "mov", "video/mp4"
		valid = completeMP4(data)
	}
	if !valid {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if err := v.ValidateReady(); err != nil {
		return VerifiedContent{}, err
	}
	probeBytes, err := v.execute(v.ffprobePath, []string{"-v", "error", "-protocol_whitelist",
		"pipe", "-f", demuxer, "-count_frames", "-show_entries",
		"stream=codec_type,codec_name,nb_read_frames,width,height,sample_rate:stream_disposition=attached_pic:format=format_name",
		"-of", "json", "-i", "pipe:0"}, data, 15*time.Second)
	if err != nil {
		return VerifiedContent{}, err
	}
	var probe mediaProbe
	if json.Unmarshal(probeBytes, &probe) != nil || len(probe.Streams) < 1 || len(probe.Streams) > 8 {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	maps := []string{"-map", "0:a:0"}
	if extension == "mp3" {
		stream := probe.Streams[0]
		rate, rateErr := strconv.Atoi(stream.SampleRate)
		frames, frameErr := strconv.Atoi(stream.Frames)
		if probe.Format.Name != "mp3" || len(probe.Streams) != 1 ||
			stream.CodecType != "audio" || stream.CodecName != "mp3" ||
			rateErr != nil || rate < 8000 || rate > 192000 ||
			frameErr != nil || frames < 1 || frames > v.maximumFrames {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
	} else {
		videos := 0
		if !strings.Contains(probe.Format.Name, "mp4") {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		for _, stream := range probe.Streams {
			if stream.CodecType != "video" && stream.CodecType != "audio" {
				return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
			}
			if stream.CodecType == "video" {
				videos++
				frames, frameErr := strconv.Atoi(stream.Frames)
				pixels := int64(stream.Width) * int64(stream.Height)
				if stream.Disposition.AttachedPic != 0 || stream.Width < 1 || stream.Height < 1 ||
					pixels < 1 || pixels > v.maximumVideoPixels || frameErr != nil ||
					frames < 1 || frames > v.maximumFrames {
					return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
				}
			}
		}
		if videos < 1 || videos > 4 {
			return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
		}
		maps = []string{"-map", "0:v", "-map", "0:a?"}
	}
	arguments := append([]string{"-nostdin", "-v", "error", "-xerror", "-err_detect",
		"explode", "-threads", "1", "-protocol_whitelist", "pipe", "-f", demuxer,
		"-i", "pipe:0"}, maps...)
	arguments = append(arguments, "-f", "null", "-")
	if _, err := v.execute(v.ffmpegPath, arguments, data, 15*time.Second); err != nil {
		return VerifiedContent{}, err
	}
	digest := sha256.Sum256(data)
	return VerifiedContent{SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(digest[:]),
		ContentType: mime}, nil
}

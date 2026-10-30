//go:build linux

package blobreference

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"syscall"
	"time"
)

// LegacyOfficeContentVerifier is a passive Linux DOC/XLS adapter. A patched,
// host-managed LibreOffice process converts the original CFB file inside a
// network-isolated, read-only bubblewrap namespace. The converted package is
// fully checked by the bounded OOXML verifier before original-byte facts return.
type LegacyOfficeContentVerifier struct {
	sofficePath, bwrapPath, prlimitPath, scratchRoot string
	maximumInputBytes, maximumConvertedBytes         int64
	packageVerifier                                  *OfficeContentVerifier
	permits                                          chan struct{}
}

func validLegacyOfficeContainer(data []byte) bool {
	if len(data) < 1024 || !bytes.HasPrefix(data,
		[]byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}) ||
		binary.LittleEndian.Uint16(data[28:30]) != 0xfffe ||
		binary.LittleEndian.Uint16(data[32:34]) != 6 {
		return false
	}
	major := binary.LittleEndian.Uint16(data[26:28])
	shift := binary.LittleEndian.Uint16(data[30:32])
	if major != 3 && major != 4 || major == 3 && shift != 9 || major == 4 && shift != 12 {
		return false
	}
	sectorSize := 1 << shift
	if len(data)%sectorSize != 0 || len(data) < 2*sectorSize {
		return false
	}
	sectors := uint32(len(data)/sectorSize - 1)
	fatCount := binary.LittleEndian.Uint32(data[44:48])
	firstDirectory := binary.LittleEndian.Uint32(data[48:52])
	if fatCount == 0 || fatCount > sectors || firstDirectory >= sectors ||
		binary.LittleEndian.Uint32(data[56:60]) != 4096 {
		return false
	}
	miniFatCount := binary.LittleEndian.Uint32(data[64:68])
	firstMiniFat := binary.LittleEndian.Uint32(data[60:64])
	if miniFatCount > sectors || miniFatCount > 0 && firstMiniFat >= sectors {
		return false
	}
	difatCount := binary.LittleEndian.Uint32(data[72:76])
	firstDifat := binary.LittleEndian.Uint32(data[68:72])
	if difatCount > sectors || difatCount > 0 && firstDifat >= sectors {
		return false
	}
	return true
}

func NewLegacyOfficeContentVerifier(sofficePath, bwrapPath, prlimitPath,
	scratchRoot string, maximumInputBytes, maximumConvertedBytes int64,
	concurrentVerifications int) (*LegacyOfficeContentVerifier, error) {
	if !filepath.IsAbs(sofficePath) || !filepath.IsAbs(bwrapPath) ||
		!filepath.IsAbs(prlimitPath) || !filepath.IsAbs(scratchRoot) ||
		filepath.Clean(scratchRoot) != scratchRoot ||
		maximumInputBytes < 1 || maximumInputBytes > 100<<20 ||
		maximumConvertedBytes < 1 || maximumConvertedBytes > 10<<20 ||
		concurrentVerifications < 1 || concurrentVerifications > 32 {
		return nil, uploadError("BLOB_POLICY_INVALID")
	}
	packageVerifier, err := NewOfficeContentVerifier(maximumConvertedBytes,
		100<<20, 10000, 2000000, concurrentVerifications)
	if err != nil {
		return nil, err
	}
	return &LegacyOfficeContentVerifier{sofficePath, bwrapPath, prlimitPath,
		scratchRoot, maximumInputBytes, maximumConvertedBytes, packageVerifier,
		make(chan struct{}, concurrentVerifications)}, nil
}

func (v *LegacyOfficeContentVerifier) ValidateReady() error {
	if v == nil || runtime.GOOS != "linux" {
		return uploadError("BLOB_CAPABILITY_MISSING")
	}
	for _, path := range []string{v.sofficePath, v.bwrapPath, v.prlimitPath} {
		info, err := os.Stat(path)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			return uploadError("BLOB_CAPABILITY_MISSING")
		}
	}
	root, err := os.Lstat(v.scratchRoot)
	if err != nil || !root.IsDir() || root.Mode().Perm() != 0700 {
		return uploadError("BLOB_CAPABILITY_MISSING")
	}
	resolved, err := filepath.EvalSymlinks(v.scratchRoot)
	if err != nil || resolved != v.scratchRoot {
		return uploadError("BLOB_CAPABILITY_MISSING")
	}
	owner, ok := root.Sys().(*syscall.Stat_t)
	if !ok || owner.Uid != uint32(os.Geteuid()) {
		return uploadError("BLOB_CAPABILITY_MISSING")
	}
	return nil
}

func (v *LegacyOfficeContentVerifier) convert(data []byte, extension string) ([]byte, error) {
	job, err := os.MkdirTemp(v.scratchRoot, "legacy-office-")
	if err != nil {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	defer os.RemoveAll(job)
	if err := os.Chmod(job, 0700); err != nil {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	outputDir, profileDir := filepath.Join(job, "output"), filepath.Join(job, "profile")
	if os.Mkdir(outputDir, 0700) != nil || os.Mkdir(profileDir, 0700) != nil {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	inputPath := filepath.Join(job, "source."+extension)
	if err := os.WriteFile(inputPath, data, 0600); err != nil {
		return nil, uploadError("BLOB_STORE_FAILED")
	}
	convertedExtension := extension + "x"
	profileURL := (&url.URL{Scheme: "file", Path: profileDir}).String()
	args := []string{"--cpu=20:20", "--as=1073741824:1073741824",
		"--fsize=10485760:10485760", "--", v.bwrapPath,
		"--unshare-net", "--unshare-pid", "--die-with-parent",
		"--ro-bind", "/usr", "/usr", "--ro-bind", "/lib", "/lib",
		"--ro-bind", "/lib64", "/lib64", "--symlink", "usr/bin", "/bin",
		"--ro-bind", "/etc", "/etc", "--dev", "/dev", "--proc", "/proc",
		"--dir", "/tmp", "--bind", job, job, "--clearenv",
		"--setenv", "LC_ALL", "C", "--setenv", "PATH", "/usr/bin:/bin",
		"--setenv", "HOME", job, "--setenv", "TMPDIR", job,
		"--", v.sofficePath, "-env:UserInstallation=" + profileURL,
		"--headless", "--nodefault", "--norestore", "--nolockcheck",
		"--convert-to", convertedExtension, "--outdir", outputDir, inputPath}
	command := exec.Command(v.prlimitPath, args...)
	command.Env = []string{"LC_ALL=C", "PATH=/usr/bin:/bin"}
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	command.Stdout = &cappedOutput{maximum: 4096}
	command.Stderr = &cappedOutput{maximum: 4096}
	if err := command.Start(); err != nil {
		return nil, uploadError("BLOB_CAPABILITY_MISSING")
	}
	var fence sync.Mutex
	finished, timedOut := false, false
	timer := time.AfterFunc(30*time.Second, func() {
		fence.Lock()
		defer fence.Unlock()
		if !finished {
			timedOut = true
			_ = syscall.Kill(-command.Process.Pid, syscall.SIGKILL)
		}
	})
	err = command.Wait()
	fence.Lock()
	finished = true
	wasTimedOut := timedOut
	fence.Unlock()
	timer.Stop()
	// Bubblewrap's PID namespace and die-with-parent fence any descendants.
	if wasTimedOut {
		return nil, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if err != nil {
		return nil, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	entries, err := os.ReadDir(outputDir)
	if err != nil || len(entries) != 1 || entries[0].Name() != "source."+convertedExtension {
		return nil, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	converted := filepath.Join(outputDir, entries[0].Name())
	info, err := os.Lstat(converted)
	if err != nil || !info.Mode().IsRegular() || info.Size() < 1 ||
		info.Size() > v.maximumConvertedBytes {
		return nil, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	packageBytes, err := os.ReadFile(converted)
	if err != nil || int64(len(packageBytes)) != info.Size() {
		return nil, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	return packageBytes, nil
}

func (v *LegacyOfficeContentVerifier) Verify(source io.Reader, extension string,
	maximumBytes int64) (VerifiedContent, error) {
	if extension != "doc" && extension != "xls" {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	if v == nil || source == nil || maximumBytes < 1 {
		return VerifiedContent{}, uploadError("BLOB_POLICY_INVALID")
	}
	if err := v.ValidateReady(); err != nil {
		return VerifiedContent{}, err
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
	if !validLegacyOfficeContainer(data) {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	converted, err := v.convert(data, extension)
	if err != nil {
		return VerifiedContent{}, err
	}
	if _, err := v.packageVerifier.Verify(bytes.NewReader(converted), extension+"x",
		int64(len(converted))); err != nil {
		return VerifiedContent{}, uploadError("BLOB_FORMAT_NOT_ALLOWED")
	}
	digest := sha256.Sum256(data)
	contentType := "application/msword"
	if extension == "xls" {
		contentType = "application/vnd.ms-excel"
	}
	return VerifiedContent{SizeBytes: int64(len(data)), SHA256: hex.EncodeToString(digest[:]),
		ContentType: contentType}, nil
}

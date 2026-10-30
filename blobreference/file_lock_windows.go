//go:build windows

package blobreference

import (
	"errors"
	"os"

	"golang.org/x/sys/windows"
)

func tryManifestFileLock(file *os.File) (bool, error) {
	var overlap windows.Overlapped
	err := windows.LockFileEx(windows.Handle(file.Fd()),
		windows.LOCKFILE_EXCLUSIVE_LOCK|windows.LOCKFILE_FAIL_IMMEDIATELY,
		0, 1, 0, &overlap)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, windows.ERROR_LOCK_VIOLATION) {
		return false, nil
	}
	return false, err
}

func unlockManifestFile(file *os.File) {
	var overlap windows.Overlapped
	_ = windows.UnlockFileEx(windows.Handle(file.Fd()), 0, 1, 0, &overlap)
}

//go:build linux || darwin || freebsd || netbsd || openbsd || dragonfly || solaris || aix

package blobreference

import (
	"errors"
	"os"

	"golang.org/x/sys/unix"
)

func tryManifestFileLock(file *os.File) (bool, error) {
	err := unix.Flock(int(file.Fd()), unix.LOCK_EX|unix.LOCK_NB)
	if err == nil {
		return true, nil
	}
	if errors.Is(err, unix.EWOULDBLOCK) || errors.Is(err, unix.EAGAIN) {
		return false, nil
	}
	return false, err
}

func unlockManifestFile(file *os.File) { _ = unix.Flock(int(file.Fd()), unix.LOCK_UN) }

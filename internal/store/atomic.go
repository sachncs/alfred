package store

import (
	"errors"
	"os"
	"runtime"
	"time"
)

// maxRetries is the maximum number of retries for atomic writes on
// Windows (EPERM/EBUSY from antivirus or file locking).
const maxRetries = 3

// AtomicWrite writes data to path atomically using tmp + rename.
// On Windows, retries up to maxRetries times on EPERM/EBUSY/EACCES
// errors that commonly come from antivirus scanners or file locking.
func AtomicWrite(path string, data []byte) error {
	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		lastErr = atomicWriteOnce(path, data)
		if lastErr == nil {
			return nil
		}
		if !isRetryable(lastErr) || runtime.GOOS != "windows" {
			return lastErr
		}
		time.Sleep(time.Duration(attempt+1) * 10 * time.Millisecond)
	}
	return lastErr
}

func atomicWriteOnce(path string, data []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, data, 0o644); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

func isRetryable(err error) bool {
	if err == nil {
		return false
		// ponytail: only check syscall-level errors on Windows
	}
	var pathErr *os.PathError
	if errors.As(err, &pathErr) {
		// EPERM, EBUSY, EACCES
		return errors.Is(err, os.ErrPermission) || errors.Is(err, os.ErrExist)
	}
	return false
}

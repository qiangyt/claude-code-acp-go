//go:build !windows
// +build !windows

package comm

// a revised copy of github.com/allan-simon/go-singleinstance v0.0.0-20210120080615-d0997106ab37

import (
	"encoding/json"
	"os"
	"syscall"

	"github.com/spf13/afero"
)

// CreateLockFile tries to create a file with given name and acquire an
// exclusive lock on it. If the file already exists AND is still locked, it will
// fail.
func CreateLockFile(fs afero.Fs, filename string, data any) (afero.File, error) {
	f, err := fs.OpenFile(filename, os.O_WRONLY|os.O_CREATE, 0o600)
	if err != nil {
		return nil, err
	}

	if osFile, isOsFile := f.(*os.File); isOsFile {
		err = syscall.Flock(int(osFile.Fd()), syscall.LOCK_EX|syscall.LOCK_NB)
		if err != nil {
			f.Close()
			return nil, err
		}
	}

	// Write PID to lock file
	pid := os.Getpid()

	payload, err := json.Marshal(map[string]any{
		"pid":  pid,
		"data": data,
	})
	if err != nil {
		f.Close()
		return nil, err
	}

	if err := f.Truncate(0); err != nil {
		f.Close()
		return nil, err
	}
	if _, err := f.WriteString(string(payload)); err != nil {
		f.Close()
		return nil, err
	}

	return f, nil
}

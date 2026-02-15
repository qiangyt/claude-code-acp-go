package comm

// a revised copy of github.com/allan-simon/go-singleinstance v0.0.0-20210120080615-d0997106ab37

import (
	"encoding/json"

	"github.com/spf13/afero"
)

// If filename is a lock file, returns the PID of the process locking it
func ReadLockFile(fs afero.Fs, filename string) (int, any, error) {
	contents, err := afero.ReadFile(fs, filename)
	if err != nil {
		return 0, nil, err
	}

	payload := map[string]any{}
	if err = json.Unmarshal(contents, &payload); err != nil {
		return 0, nil, err
	}

	return payload["pid"].(int), payload["data"], nil
}

//go:build !windows

package install

import (
	"os"
	"path/filepath"
)

const exeName = "dtvlt"

func defaultDir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "bin"), nil
}

// On Linux ~/.local/bin is on the PATH of most distributions through
// ~/.profile. dtvlt does not edit shell startup files; the CLI prints a hint
// when the directory is missing.
func addToPath(dir string) (bool, error) { return false, nil }

func pathPersisted(dir string) bool { return false }

func removeFromPath(dir string) error { return nil }

// A running binary can be unlinked on Linux, and ~/.local/bin is shared, so
// only the file goes.
func removeBinary(dst, dir string) error {
	removeOld(dst)
	return os.Remove(dst)
}

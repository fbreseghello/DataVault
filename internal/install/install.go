// Package install copies dtvlt to a per-user location and puts it on the PATH,
// so a downloaded binary can set itself up without admin rights.
//
//	Windows  %LOCALAPPDATA%\Programs\DataVault\dtvlt.exe   (user PATH updated)
//	Linux    ~/.local/bin/dtvlt
//
// DTVLT_INSTALL_DIR overrides the directory. With an override the PATH is left
// alone, which also keeps tests from touching the real environment.
package install

import (
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

type Result struct {
	Path string
	// PathUpdated is true when the install dir was added to the user PATH.
	PathUpdated bool
	// OnPath is false when the user still has to add the dir to PATH by hand.
	OnPath bool
}

// Dir is where dtvlt gets installed.
func Dir() (string, error) {
	if d := os.Getenv("DTVLT_INSTALL_DIR"); d != "" {
		return filepath.Abs(d)
	}
	return defaultDir()
}

func customDir() bool { return os.Getenv("DTVLT_INSTALL_DIR") != "" }

// Target is the full path of the installed binary.
func Target() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, exeName), nil
}

// Installed reports whether an installed binary exists.
func Installed() bool {
	t, err := Target()
	if err != nil {
		return false
	}
	_, err = os.Stat(t)
	return err == nil
}

// RunningInstalled reports whether the current process is the installed binary.
func RunningInstalled() bool {
	t, err := Target()
	if err != nil {
		return false
	}
	self, err := self()
	if err != nil {
		return false
	}
	return samePath(self, t)
}

// InstalledVersion asks the installed binary for its version ("" if unknown).
func InstalledVersion() string {
	t, err := Target()
	if err != nil {
		return ""
	}
	out, err := exec.Command(t, "version").Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(out)), "dtvlt"))
}

// Install copies the running binary to Target and makes sure it is on the PATH.
func Install() (Result, error) {
	src, err := self()
	if err != nil {
		return Result{}, err
	}
	return InstallFile(src)
}

// InstallFile installs the binary at src (a downloaded update, for example).
func InstallFile(src string) (Result, error) {
	dir, err := Dir()
	if err != nil {
		return Result{}, err
	}
	dst := filepath.Join(dir, exeName)
	res := Result{Path: dst}

	if !samePath(src, dst) {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return res, err
		}
		if err := replace(src, dst); err != nil {
			return res, err
		}
	}

	if customDir() {
		res.OnPath = onPath(dir)
		return res, nil
	}
	res.PathUpdated, err = addToPath(dir)
	res.OnPath = err == nil && (res.PathUpdated || onPath(dir) || pathPersisted(dir))
	return res, err
}

// Uninstall removes the installed binary and the PATH entry. The vault and
// config are left untouched.
func Uninstall() error {
	dir, err := Dir()
	if err != nil {
		return err
	}
	if !customDir() {
		if err := removeFromPath(dir); err != nil {
			return err
		}
	}
	dst := filepath.Join(dir, exeName)
	if _, err := os.Stat(dst); errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return removeBinary(dst, dir)
}

// Cleanup removes the copies left behind when running binaries were replaced.
// Copies still in use stay until a later run.
func Cleanup() {
	t, err := Target()
	if err != nil {
		return
	}
	removeOld(t)
}

func removeOld(dst string) {
	olds, _ := filepath.Glob(dst + ".old*")
	for _, old := range olds {
		_ = os.Remove(old)
	}
}

// replace swaps dst for a copy of src. A running executable cannot be
// overwritten on Windows but it can be renamed, so the old one moves aside.
func replace(src, dst string) error {
	tmp := dst + ".new"
	if err := copyFile(src, tmp); err != nil {
		os.Remove(tmp)
		return err
	}
	var aside string
	if _, err := os.Stat(dst); err == nil {
		aside = asideName(dst)
		if err := os.Rename(dst, aside); err != nil {
			os.Remove(tmp)
			return err
		}
	}
	if err := os.Rename(tmp, dst); err != nil {
		if aside != "" {
			_ = os.Rename(aside, dst)
		}
		return err
	}
	if aside != "" {
		_ = os.Remove(aside) // fails harmlessly while the old binary still runs
	}
	return nil
}

// asideName picks where the current binary moves during an update. Usually
// dst.old, but that file may still be held by a session from an even older
// version, so a free dst.old-N is used instead.
func asideName(dst string) string {
	old := dst + ".old"
	if err := os.Remove(old); err == nil || errors.Is(err, os.ErrNotExist) {
		return old
	}
	for n := 1; ; n++ {
		name := fmt.Sprintf("%s-%d", old, n)
		if err := os.Remove(name); err == nil || errors.Is(err, os.ErrNotExist) {
			return name
		}
	}
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.OpenFile(dst, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o755)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func self() (string, error) {
	p, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(p)
}

func samePath(a, b string) bool {
	a, b = filepath.Clean(a), filepath.Clean(b)
	if runtime.GOOS == "windows" {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// onPath reports whether dir is on the PATH of this process.
func onPath(dir string) bool {
	for _, p := range filepath.SplitList(os.Getenv("PATH")) {
		if p != "" && samePath(os.ExpandEnv(p), dir) {
			return true
		}
	}
	return false
}

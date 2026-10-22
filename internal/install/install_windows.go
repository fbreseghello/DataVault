package install

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

const exeName = "dtvlt.exe"

func defaultDir() (string, error) {
	base := os.Getenv("LOCALAPPDATA")
	if base == "" {
		return "", errors.New("LOCALAPPDATA is not set")
	}
	return filepath.Join(base, "Programs", "DataVault"), nil
}

// The user PATH lives in HKCU\Environment. Changing it there needs no admin
// rights and applies to terminals opened afterwards.
const envKey = `Environment`

func readUserPath() (string, uint32, error) {
	k, err := registry.OpenKey(registry.CURRENT_USER, envKey, registry.QUERY_VALUE)
	if err != nil {
		return "", registry.EXPAND_SZ, err
	}
	defer k.Close()
	v, typ, err := k.GetStringValue("Path")
	if errors.Is(err, registry.ErrNotExist) {
		return "", registry.EXPAND_SZ, nil
	}
	return v, typ, err
}

func writeUserPath(v string, typ uint32) error {
	k, err := registry.OpenKey(registry.CURRENT_USER, envKey, registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer k.Close()
	if typ == registry.SZ {
		err = k.SetStringValue("Path", v)
	} else {
		err = k.SetExpandStringValue("Path", v)
	}
	if err == nil {
		broadcastEnvChange()
	}
	return err
}

func splitPath(v string) []string {
	var out []string
	for _, p := range strings.Split(v, ";") {
		if p = strings.TrimSpace(p); p != "" {
			out = append(out, p)
		}
	}
	return out
}

func containsDir(entries []string, dir string) bool {
	for _, p := range entries {
		expanded, err := registry.ExpandString(p)
		if err != nil {
			expanded = p
		}
		if samePath(strings.TrimRight(expanded, `\`), dir) {
			return true
		}
	}
	return false
}

func pathPersisted(dir string) bool {
	v, _, err := readUserPath()
	return err == nil && containsDir(splitPath(v), dir)
}

func addToPath(dir string) (bool, error) {
	v, typ, err := readUserPath()
	if err != nil {
		return false, err
	}
	entries := splitPath(v)
	if containsDir(entries, dir) {
		return false, nil
	}
	return true, writeUserPath(strings.Join(append(entries, dir), ";"), typ)
}

func removeFromPath(dir string) error {
	v, typ, err := readUserPath()
	if err != nil {
		return err
	}
	entries := splitPath(v)
	var kept []string
	for _, p := range entries {
		if !containsDir([]string{p}, dir) {
			kept = append(kept, p)
		}
	}
	if len(kept) == len(entries) {
		return nil
	}
	return writeUserPath(strings.Join(kept, ";"), typ)
}

// broadcastEnvChange tells Explorer (and new terminals it starts) that the
// environment changed, so no logoff is needed.
func broadcastEnvChange() {
	const (
		hwndBroadcast   = 0xffff
		wmSettingChange = 0x001a
		smtoAbortIfHung = 0x0002
	)
	env, _ := windows.UTF16PtrFromString("Environment")
	var result uintptr
	proc := windows.NewLazySystemDLL("user32.dll").NewProc("SendMessageTimeoutW")
	_, _, _ = proc.Call(hwndBroadcast, wmSettingChange, 0, uintptr(unsafe.Pointer(env)),
		smtoAbortIfHung, 5000, uintptr(unsafe.Pointer(&result)))
}

// removeBinary deletes dst. If it is the running process, Windows refuses, so a
// detached cmd deletes it (and the empty directory) a moment after we exit.
func removeBinary(dst, dir string) error {
	if err := os.Remove(dst); err == nil {
		removeOld(dst)
		_ = os.Remove(dir)
		return nil
	}
	cmd := exec.Command("cmd.exe")
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CmdLine:       fmt.Sprintf(`cmd.exe /c ping -n 3 127.0.0.1 >nul & del /f /q "%s" "%s.old*" & rmdir "%s"`, dst, dst, dir),
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
	return cmd.Start()
}

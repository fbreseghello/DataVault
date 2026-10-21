//go:build !windows

package cli

import (
	"os/exec"
	"syscall"
)

func enableTerminal() bool { return true }

// detachAttr puts the background session in its own session, away from this
// terminal's hangup signal.
func detachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{Setsid: true}
}

func openBrowser(url string) error {
	return exec.Command("xdg-open", url).Start()
}

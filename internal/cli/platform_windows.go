package cli

import (
	"os"
	"os/exec"
	"syscall"

	"golang.org/x/sys/windows"
)

// enableTerminal switches the console to UTF-8 and turns on ANSI escape
// handling. It reports whether escapes will render.
func enableTerminal() bool {
	const cpUTF8 = 65001
	_ = windows.SetConsoleOutputCP(cpUTF8)
	h := windows.Handle(os.Stdout.Fd())
	var mode uint32
	if windows.GetConsoleMode(h, &mode) != nil {
		return false
	}
	return windows.SetConsoleMode(h, mode|windows.ENABLE_VIRTUAL_TERMINAL_PROCESSING) == nil
}

// detachAttr starts the background session without a console, in its own
// process group, so closing this terminal does not kill it.
func detachAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		CreationFlags: windows.DETACHED_PROCESS | windows.CREATE_NEW_PROCESS_GROUP,
		HideWindow:    true,
	}
}

func openBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

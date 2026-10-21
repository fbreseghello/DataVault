package cli

import (
	"os"
	"os/exec"
	"syscall"
	"unsafe"

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

// ownsConsole reports whether this process is alone in its console, which is
// what happens when the .exe is opened from Explorer instead of a terminal.
func ownsConsole() bool {
	var pids [2]uint32
	proc := windows.NewLazySystemDLL("kernel32.dll").NewProc("GetConsoleProcessList")
	n, _, _ := proc.Call(uintptr(unsafe.Pointer(&pids[0])), uintptr(len(pids)))
	return n == 1
}

func openBrowser(url string) error {
	return exec.Command("rundll32", "url.dll,FileProtocolHandler", url).Start()
}

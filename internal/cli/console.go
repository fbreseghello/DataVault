package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"strings"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/install"
	"github.com/fbreseghello/DataVault/internal/server"
	"github.com/fbreseghello/DataVault/internal/session"
	"github.com/fbreseghello/DataVault/internal/version"
)

// openKey is what the user types in the console to open DataVault in the browser.
const openKey = ",,,"

// controller is the session the console drives: either served by this process
// or a background one reached through the control API.
type controller interface {
	LaunchURL() (string, error)
	Status() (server.Status, error)
	Lock() error
	Stop() error
}

// localSession is a session served by this process.
type localSession struct {
	srv  *server.Server
	sess *session.Session
	stop func()
}

func (l localSession) LaunchURL() (string, error) { return l.srv.LaunchURL(), nil }

func (l localSession) Status() (server.Status, error) {
	return server.Status{
		SessionID: l.sess.ID,
		Unlocked:  l.sess.Unlocked(),
		PID:       os.Getpid(),
		URL:       l.srv.URL(),
		Version:   version.Version,
	}, nil
}

func (l localSession) Lock() error { l.sess.Lock("console"); return nil }
func (l localSession) Stop() error { l.stop(); return nil }

const consoleHelp = `  ,,,       open DataVault in the browser
  status    show the session
  lock      lock the vault, keep the session
  update    download and install the latest release
  config    change preferences
  version   print the version
  clear     clear the screen
  stop      end the session
  exit      leave the console`

// console reads commands until exit. owns is true when this process serves
// the session: leaving the console then ends it. Otherwise the session keeps
// running in the background.
func console(u *ui, paths config.Paths, ctl controller, owns bool) error {
	prompt := "  " + u.dim("dtvlt") + " › "
	for {
		line, err := u.command(prompt)
		if err != nil {
			// End of input (ctrl+z on Windows, ctrl+d elsewhere) leaves like exit.
			u.println("")
			return leave(u, ctl, owns)
		}
		switch cmd := strings.ToLower(strings.TrimSpace(line)); cmd {
		case "":
		case openKey:
			url, err := ctl.LaunchURL()
			if err != nil {
				u.fail("%v", err)
				continue
			}
			if openOrPrint(u, url) {
				u.println("  %s", u.dim("opened in the browser"))
			}
		case "help", "?":
			u.println("%s", consoleHelp)
		case "status":
			st, err := ctl.Status()
			if err != nil {
				u.fail("%v", err)
				continue
			}
			printStatus(u, st)
		case "lock":
			if err := ctl.Lock(); err != nil {
				u.fail("%v", err)
				continue
			}
			u.println("  locked")
		case "update":
			updated, err := runUpdate(u, paths, true)
			if err != nil {
				u.fail("%v", err)
				continue
			}
			if updated {
				// This console and its session are still the old version: end
				// the session and hand the terminal to the new binary.
				target, err := install.Target()
				if err != nil {
					u.fail("%v", err)
					continue
				}
				u.println("")
				u.println("  %s", u.dim("restarting on the new version..."))
				relaunch = target
				return ctl.Stop()
			}
		case "config":
			if err := configure(u, paths); err != nil && !errors.Is(err, errAborted) {
				u.fail("%v", err)
			}
		case "version":
			u.println("  dtvlt %s", version.Version)
		case "clear", "cls":
			if u.color {
				u.mu.Lock()
				u.out.Write([]byte("\x1b[2J\x1b[H"))
				u.mu.Unlock()
			}
			printBanner(u)
		case "stop":
			if err := ctl.Stop(); err != nil {
				return err
			}
			u.println("  session ended · vault locked")
			return nil
		case "exit", "quit", "q":
			return leave(u, ctl, owns)
		default:
			u.fail("unknown command %q · type help", cmd)
		}
	}
}

func leave(u *ui, ctl controller, owns bool) error {
	if !owns {
		u.println("  %s", u.dim("console closed · the session keeps running in the background"))
		return nil
	}
	u.println("  session ended · vault locked")
	return ctl.Stop()
}

func printStatus(u *ui, st server.Status) {
	state := "locked"
	if st.Unlocked {
		state = "unlocked"
	}
	ver := st.Version
	if ver == "" {
		ver = "0.6.0 or older"
	}
	u.row("session", u.bold(st.SessionID))
	u.row("state", state)
	u.row("version", ver)
	u.row("url", st.URL)
}

// relaunch is set when the console updated dtvlt: once this process is done
// with its session, Run hands the terminal to the installed binary.
var relaunch string

// relaunchInto runs the binary at path in this terminal and returns its exit
// code. ctrl+c belongs to the new process now, so this one ignores it.
func relaunchInto(path string) int {
	signal.Ignore(os.Interrupt)
	cmd := exec.Command(path)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		fmt.Fprintf(os.Stderr, "  ! could not start %s: %v\n", path, err)
		return 1
	}
	return 0
}

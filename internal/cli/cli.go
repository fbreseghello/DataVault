// Package cli implements the dtvlt command.
package cli

import (
	_ "embed"
	"errors"
	"strconv"
	"strings"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/install"
	"github.com/fbreseghello/DataVault/internal/server"
	"github.com/fbreseghello/DataVault/internal/version"
)

// serveCommand is the hidden subcommand a detached session runs as.
const serveCommand = "__serve"

//go:embed banner.txt
var banner string

// printBanner prints the art with the running version under its right edge.
func printBanner(u *ui) {
	lines := strings.Split(strings.TrimRight(banner, "\n"), "\n")
	width := 0
	for _, line := range lines {
		width = max(width, len(line))
	}
	u.println("")
	for _, line := range lines {
		u.println("  %s", line)
	}
	ver := version.Version
	if !version.IsDev() {
		ver = "v" + ver
	}
	u.println("  %s%s", strings.Repeat(" ", max(0, width-len(ver))), u.dim(ver))
	u.println("")
}

const usage = `usage: dtvlt [command]

  (none)    open the console (,,, opens DataVault in the browser)
  status    show the running session
  lock      lock the vault, keep the session
  stop      end the session
  config    change preferences
  update    download and install the latest release
  install   install dtvlt for this user and add it to the PATH
  uninstall remove dtvlt (keeps your vault)
  version   print the version
`

// Run executes dtvlt with args (without the program name) and returns the exit code.
func Run(args []string) int {
	u := newUI()
	cmd := ""
	if len(args) > 0 {
		cmd = args[0]
	}
	code := run(u, cmd)
	// Opened by double-click: the console window closes with the process, so
	// keep it up when there is something to read.
	if cmd != serveCommand && ownsConsole() && (code != 0 || u.pause) {
		u.println("")
		u.prompt(u.dim("press enter to close"), "")
		_, _ = u.readLine()
	}
	return code
}

func run(u *ui, cmd string) int {
	install.Cleanup()
	paths, err := config.ResolvePaths()
	if err != nil {
		u.fail("%v", err)
		return 1
	}

	switch cmd {
	case "":
		err = start(u, paths)
	case "install":
		printBanner(u)
		err = runInstall(u)
	case "uninstall":
		err = runUninstall(u, paths)
	case "update":
		err = runUpdate(u, paths)
	case serveCommand:
		err = serveDetached(paths)
	case "status":
		err = status(u, paths)
	case "lock":
		err = withSession(u, paths, func(c *server.Client) error {
			if err := c.Lock(); err != nil {
				return err
			}
			u.println("  locked")
			return nil
		})
	case "stop":
		err = withSession(u, paths, func(c *server.Client) error {
			if err := c.Stop(); err != nil {
				return err
			}
			u.println("  stopped")
			return nil
		})
	case "config":
		err = configure(u, paths)
	case "version", "--version", "-v":
		u.println("dtvlt %s", version.Version)
	case "help", "--help", "-h":
		u.println("%s", usage)
	default:
		u.fail("unknown command %q", cmd)
		u.println("")
		u.println("%s", usage)
		return 2
	}

	if errors.Is(err, errAborted) {
		u.println("")
		return 130
	}
	if err != nil {
		u.fail("%v", err)
		return 1
	}
	return 0
}

func withSession(u *ui, paths config.Paths, fn func(*server.Client) error) error {
	c, err := server.Dial(paths.State)
	if errors.Is(err, server.ErrNotRunning) {
		u.println("  %s", u.dim("no session running"))
		return nil
	}
	if err != nil {
		return err
	}
	return fn(c)
}

func status(u *ui, paths config.Paths) error {
	return withSession(u, paths, func(c *server.Client) error {
		st, err := c.Status()
		if err != nil {
			return err
		}
		printStatus(u, st)
		u.row("pid", strconv.Itoa(st.PID))
		return nil
	})
}

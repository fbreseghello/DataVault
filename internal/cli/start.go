package cli

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/install"
	"github.com/fbreseghello/DataVault/internal/server"
	"github.com/fbreseghello/DataVault/internal/session"
	"github.com/fbreseghello/DataVault/internal/version"
	"github.com/fbreseghello/DataVault/web"
)

const maxPINTries = 3

// start is plain `dtvlt`: set up if needed, announce updates, check the PIN,
// then open the console on a session. The browser opens only from the console.
func start(u *ui, paths config.Paths) error {
	printBanner(u)
	if actsAsInstaller() {
		return installSelf(u, paths)
	}

	cfg, err := ensureSetup(u, paths)
	if err != nil {
		return err
	}
	if cfg.CheckUpdates {
		updateNotice(u, paths)
	}
	if cfg.PinEnabled {
		if err := checkPIN(u, cfg); err != nil {
			return err
		}
	}

	if c, err := server.Dial(paths.State); err == nil {
		restarted, err := replaceStale(u, c, paths)
		if err != nil {
			return err
		}
		if !restarted {
			return attach(u, paths, c, "already running")
		}
	}
	if cfg.KeepTerminal {
		return serveForeground(u, paths, cfg)
	}
	c, err := startDetached(paths)
	if err != nil {
		return err
	}
	return attach(u, paths, c, "running in the background")
}

// replaceStale stops a running session that belongs to another dtvlt version,
// typically the one that was running when dtvlt update replaced the binary,
// so this start serves the version the user now has. It reports whether the
// old session was stopped.
func replaceStale(u *ui, c *server.Client, paths config.Paths) (bool, error) {
	st, err := c.Status()
	if err != nil {
		return false, err
	}
	if st.Version == version.Version {
		return false, nil
	}
	old := st.Version
	if old == "" {
		old = "an older version"
	}
	u.row("session", st.SessionID+u.dim("  running on "+old+" · restarting on "+version.Version))
	u.println("")
	if err := c.Stop(); err != nil {
		return false, err
	}
	// Wait for the old process to let go of the state file.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if _, err := server.Dial(paths.State); errors.Is(err, server.ErrNotRunning) {
			install.Cleanup() // the old binary is free to delete now
			return true, nil
		}
		time.Sleep(100 * time.Millisecond)
	}
	return false, errors.New("the old session did not stop; run dtvlt stop and try again")
}

// attach opens the console on a session this process does not serve: one
// already running, or a background one just started. Leaving keeps it running.
func attach(u *ui, paths config.Paths, c *server.Client, note string) error {
	st, err := c.Status()
	if err != nil {
		return err
	}
	u.row("session", u.bold(st.SessionID)+u.dim("  "+note))
	u.println("")
	return console(u, paths, c, false)
}

func checkPIN(u *ui, cfg config.Config) error {
	for i := 0; i < maxPINTries; i++ {
		pin, err := u.secret("pin")
		if err != nil {
			return err
		}
		ok := config.VerifyPIN(cfg.PinHash, pin)
		clear(pin)
		if ok {
			return nil
		}
		u.fail("wrong pin")
		time.Sleep(time.Duration(i+1) * time.Second)
	}
	return errors.New("too many wrong pins")
}

// serve runs the session until it is stopped by a signal, `dtvlt stop` or the
// console. ready is called once the server is listening; stop ends the session.
func serve(paths config.Paths, cfg config.Config, log session.Logger, ready func(srv *server.Server, sess *session.Session, stop func())) error {
	sess := session.New(paths.Vault, time.Duration(cfg.AutoLockMinutes)*time.Minute, log)
	stop := make(chan struct{})
	var once sync.Once
	stopNow := func() { once.Do(func() { close(stop) }) }

	srv, err := server.New(server.Options{
		Session:    sess,
		Config:     cfg,
		ConfigPath: paths.Config,
		VaultPath:  paths.Vault,
		Assets:     web.Assets(),
		OnStop:     stopNow,
	})
	if err != nil {
		return err
	}
	if err := server.WriteState(paths.State, server.State{
		PID: os.Getpid(), Port: srv.Port(), SessionID: sess.ID, ControlToken: srv.ControlToken(),
	}); err != nil {
		return err
	}
	defer os.Remove(paths.State)

	errc := make(chan error, 1)
	go func() { errc <- srv.Serve() }()
	go sess.WatchIdle(stop)
	ready(srv, sess, stopNow)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	select {
	case <-sig:
	case <-stop:
	case err = <-errc:
	}
	stopNow()
	sess.Lock("")
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	sess.Log("stopped")
	return err
}

// serveForeground serves the session from this process, with the console in
// the same terminal. Leaving the console, ctrl+c or closing the terminal ends it.
func serveForeground(u *ui, paths config.Paths, cfg config.Config) error {
	return serve(paths, cfg, u.event, func(srv *server.Server, sess *session.Session, stop func()) {
		u.row("session", u.bold(sess.ID))
		u.row("url", srv.URL())
		u.println("")
		u.println("  %s", u.dim("closing this terminal ends the session and locks the vault"))
		u.event("started")
		u.println("")
		go func() {
			if err := console(u, paths, localSession{srv: srv, sess: sess, stop: stop}, true); err != nil {
				u.fail("%v", err)
				stop()
			}
		}()
	})
}

// handoff is what the detached child tells its parent once it is listening.
type handoff struct {
	SessionID string `json:"session_id"`
	LaunchURL string `json:"launch_url"`
}

// startDetached re-runs dtvlt as a background process, waits for it to be
// ready and returns a client for it.
func startDetached(paths config.Paths) (*server.Client, error) {
	exe, err := os.Executable()
	if err != nil {
		return nil, err
	}
	cmd := exec.Command(exe, serveCommand)
	cmd.SysProcAttr = detachAttr()
	out, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err := cmd.Start(); err != nil {
		return nil, err
	}

	got := make(chan handoff, 1)
	go func() {
		var h handoff
		if line, err := bufio.NewReader(out).ReadBytes('\n'); err == nil && json.Unmarshal(line, &h) == nil {
			got <- h
		}
		close(got)
	}()
	select {
	case _, ok := <-got:
		if !ok {
			return nil, errors.New("background session failed to start")
		}
		_ = cmd.Process.Release()
		return server.Dial(paths.State)
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		return nil, errors.New("background session did not start in time")
	}
}

// serveDetached is the hidden command run by startDetached.
func serveDetached(paths config.Paths) error {
	cfg, err := config.Load(paths.Config)
	if err != nil {
		return err
	}
	return serve(paths, cfg, nil, func(srv *server.Server, sess *session.Session, _ func()) {
		b, _ := json.Marshal(handoff{SessionID: sess.ID, LaunchURL: srv.LaunchURL()})
		fmt.Fprintf(os.Stdout, "%s\n", b)
		// The parent exits after reading this; nothing else may go to its pipe.
		os.Stdout.Close()
	})
}

// openOrPrint opens the launch link in the default browser and reports whether
// it did. DTVLT_NO_BROWSER=1 prints it instead (SSH sessions, other browsers,
// scripted runs).
func openOrPrint(u *ui, url string) bool {
	if os.Getenv("DTVLT_NO_BROWSER") == "1" {
		u.println("  %s", u.dim("open this link (single use, 60s):"))
		u.println("  %s", url)
		return false
	}
	if err := openBrowser(url); err != nil {
		u.println("  %s", u.dim("could not open a browser. open this link (single use, 60s):"))
		u.println("  %s", url)
		return false
	}
	return true
}

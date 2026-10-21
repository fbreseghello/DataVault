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
	"github.com/fbreseghello/DataVault/internal/server"
	"github.com/fbreseghello/DataVault/internal/session"
	"github.com/fbreseghello/DataVault/web"
)

const maxPINTries = 3

// start is plain `dtvlt`: set up if needed, check the PIN, then serve.
func start(u *ui, paths config.Paths) error {
	printBanner(u)

	if c, err := server.Dial(paths.State); err == nil {
		return reopen(u, c)
	}

	cfg, err := ensureSetup(u, paths)
	if err != nil {
		return err
	}
	if cfg.PinEnabled {
		if err := checkPIN(u, cfg); err != nil {
			return err
		}
	}
	if cfg.KeepTerminal {
		return serveForeground(u, paths, cfg)
	}
	return startDetached(u)
}

// reopen handles `dtvlt` while a session is already running: same session, new browser tab.
func reopen(u *ui, c *server.Client) error {
	st, err := c.Status()
	if err != nil {
		return err
	}
	url, err := c.LaunchURL()
	if err != nil {
		return err
	}
	u.row("session", st.SessionID+u.dim("  already running"))
	openOrPrint(u, url)
	return nil
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

// serve runs the session until it is stopped by a signal or `dtvlt stop`.
// ready is called once the server is listening.
func serve(paths config.Paths, cfg config.Config, log session.Logger, ready func(*server.Server, *session.Session)) error {
	sess := session.New(paths.Vault, time.Duration(cfg.AutoLockMinutes)*time.Minute, log)
	stop := make(chan struct{})
	var once sync.Once
	stopNow := func() { once.Do(func() { close(stop) }) }

	srv, err := server.New(server.Options{Session: sess, Config: cfg, Assets: web.Assets(), OnStop: stopNow})
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
	ready(srv, sess)

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

func serveForeground(u *ui, paths config.Paths, cfg config.Config) error {
	return serve(paths, cfg, u.event, func(srv *server.Server, sess *session.Session) {
		u.row("session", u.bold(sess.ID))
		u.row("url", srv.URL())
		u.println("")
		u.println("  %s", u.dim("ctrl+c stops the session · closing this terminal locks the vault"))
		u.println("")
		u.event("started")
		openOrPrint(u, srv.LaunchURL())
	})
}

// handoff is what the detached child tells its parent once it is listening.
type handoff struct {
	SessionID string `json:"session_id"`
	LaunchURL string `json:"launch_url"`
}

// startDetached re-runs dtvlt as a background process and waits for it to be ready.
func startDetached(u *ui) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, serveCommand)
	cmd.SysProcAttr = detachAttr()
	out, err := cmd.StdoutPipe()
	if err != nil {
		return err
	}
	if err := cmd.Start(); err != nil {
		return err
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
	case h, ok := <-got:
		if !ok {
			return errors.New("background session failed to start")
		}
		_ = cmd.Process.Release()
		u.row("session", u.bold(h.SessionID))
		u.println("")
		u.println("  %s", u.dim("running in the background · dtvlt stop ends it"))
		openOrPrint(u, h.LaunchURL)
		return nil
	case <-time.After(15 * time.Second):
		_ = cmd.Process.Kill()
		return errors.New("background session did not start in time")
	}
}

// serveDetached is the hidden command run by startDetached.
func serveDetached(paths config.Paths) error {
	cfg, err := config.Load(paths.Config)
	if err != nil {
		return err
	}
	return serve(paths, cfg, nil, func(srv *server.Server, sess *session.Session) {
		b, _ := json.Marshal(handoff{SessionID: sess.ID, LaunchURL: srv.LaunchURL()})
		fmt.Fprintf(os.Stdout, "%s\n", b)
		// The parent exits after reading this; nothing else may go to its pipe.
		os.Stdout.Close()
	})
}

// openOrPrint opens the launch link in the default browser. DTVLT_NO_BROWSER=1
// prints it instead (SSH sessions, other browsers, scripted runs).
func openOrPrint(u *ui, url string) {
	if os.Getenv("DTVLT_NO_BROWSER") == "1" {
		u.println("")
		u.println("  %s", u.dim("open this link (single use, 60s):"))
		u.println("  %s", url)
		return
	}
	if err := openBrowser(url); err != nil {
		u.println("")
		u.println("  %s", u.dim("could not open a browser. open this link (single use, 60s):"))
		u.println("  %s", url)
	}
}

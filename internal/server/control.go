package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/fbreseghello/DataVault/internal/version"
)

// The control API lets later `dtvlt` invocations (status, lock, stop, or a
// second `dtvlt`) talk to the running session. It needs the bearer token from
// the state file, which only the user can read.

func (s *Server) handleControlStatus(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, Status{
		SessionID: s.opts.Session.ID,
		Unlocked:  s.opts.Session.Unlocked(),
		PID:       os.Getpid(),
		URL:       s.origin,
		Version:   version.Version,
	})
}

func (s *Server) handleControlLock(w http.ResponseWriter, r *http.Request) {
	s.opts.Session.Lock("dtvlt lock")
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleControlStop(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
	if s.opts.OnStop != nil {
		go s.opts.OnStop()
	}
}

func (s *Server) handleControlLaunch(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]string{"url": s.LaunchURL()})
}

// State is written to disk while a session runs.
type State struct {
	PID          int    `json:"pid"`
	Port         int    `json:"port"`
	SessionID    string `json:"session_id"`
	ControlToken string `json:"control_token"`
}

type Status struct {
	SessionID string `json:"session_id"`
	Unlocked  bool   `json:"unlocked"`
	PID       int    `json:"pid"`
	URL       string `json:"url"`
	// Version is empty for sessions started by dtvlt 0.6.0 and older.
	Version string `json:"version,omitempty"`
}

var ErrNotRunning = errors.New("no dtvlt session is running")

func WriteState(path string, st State) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	b, err := json.Marshal(st)
	if err != nil {
		return err
	}
	return os.WriteFile(path, b, 0o600)
}

// Client talks to a running session through its control API.
type Client struct {
	st   State
	http *http.Client
}

// Dial finds the running session. A stale state file (process gone) is
// removed and reported as ErrNotRunning.
func Dial(statePath string) (*Client, error) {
	b, err := os.ReadFile(statePath)
	if errors.Is(err, os.ErrNotExist) {
		return nil, ErrNotRunning
	} else if err != nil {
		return nil, err
	}
	var st State
	if err := json.Unmarshal(b, &st); err != nil || st.Port == 0 {
		os.Remove(statePath)
		return nil, ErrNotRunning
	}
	c := &Client{st: st, http: &http.Client{Timeout: 3 * time.Second}}
	if _, err := c.Status(); err != nil {
		os.Remove(statePath)
		return nil, ErrNotRunning
	}
	return c, nil
}

func (c *Client) Status() (Status, error) {
	var st Status
	return st, c.do(http.MethodGet, "/control/status", &st)
}

func (c *Client) Lock() error { return c.do(http.MethodPost, "/control/lock", nil) }
func (c *Client) Stop() error { return c.do(http.MethodPost, "/control/stop", nil) }

func (c *Client) LaunchURL() (string, error) {
	var res struct {
		URL string `json:"url"`
	}
	return res.URL, c.do(http.MethodPost, "/control/launch", &res)
}

func (c *Client) do(method, path string, out any) error {
	req, err := http.NewRequest(method, fmt.Sprintf("http://127.0.0.1:%d%s", c.st.Port, path), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+c.st.ControlToken)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("control %s: %s", path, resp.Status)
	}
	if out != nil {
		return json.NewDecoder(resp.Body).Decode(out)
	}
	return nil
}

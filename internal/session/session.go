// Package session owns the unlocked vault for the lifetime of one dtvlt process.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"github.com/fbreseghello/DataVault/internal/vault"
)

// Logger receives one-line events for the terminal ("unlocked", "locked (idle 15m)").
// It must never be given secret values.
type Logger func(event string)

type Session struct {
	ID        string
	vaultPath string
	idle      time.Duration
	log       Logger

	unlockMu sync.Mutex // serializes unlock attempts so the backoff holds
	failures int

	mu       sync.Mutex
	v        *vault.Vault
	activity time.Time
	gen      uint64 // bumped on every unlock, so per-unlock grants die with the lock
}

func New(vaultPath string, idle time.Duration, log Logger) *Session {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	id := hex.EncodeToString(b)
	return &Session{ID: id[:4] + "-" + id[4:], vaultPath: vaultPath, idle: idle, log: log}
}

// SetLogger swaps the event sink, e.g. to silence output once detached.
func (s *Session) SetLogger(log Logger) {
	s.mu.Lock()
	s.log = log
	s.mu.Unlock()
}

func (s *Session) emit(event string) {
	s.mu.Lock()
	log := s.log
	s.mu.Unlock()
	if log != nil {
		log(event)
	}
}

// Log records an event coming from the frontend (copies and similar).
func (s *Session) Log(event string) { s.emit(event) }

// Backoff grows with consecutive failures: 0, 0.5s, 1s, 2s ... up to 8s.
func Backoff(failures int) time.Duration {
	if failures == 0 {
		return 0
	}
	d := 250 * time.Millisecond << min(failures, 6)
	return min(d, 8*time.Second)
}

// Unlock opens the vault. Wrong credentials return vault.ErrInvalidCredentials.
func (s *Session) Unlock(keyword, password []byte) error {
	s.unlockMu.Lock()
	defer s.unlockMu.Unlock()

	if s.Unlocked() {
		return nil
	}
	time.Sleep(Backoff(s.failures))
	v, err := vault.Open(s.vaultPath, keyword, password)
	if err != nil {
		if err == vault.ErrInvalidCredentials {
			s.failures++
			s.emit("unlock failed")
		}
		return err
	}
	s.failures = 0
	s.mu.Lock()
	s.v = v
	s.activity = time.Now()
	s.gen++
	s.mu.Unlock()
	s.emit("unlocked")
	return nil
}

func (s *Session) Unlocked() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.v != nil
}

// Lock wipes the key from memory. reason is shown in the terminal log.
func (s *Session) Lock(reason string) {
	s.mu.Lock()
	v := s.v
	s.v = nil
	s.mu.Unlock()
	if v == nil {
		return
	}
	v.Lock()
	if reason != "" {
		s.emit("locked  (" + reason + ")")
	} else {
		s.emit("locked")
	}
}

// Vault returns the unlocked vault and counts as activity for the idle timer.
func (s *Session) Vault() (*vault.Vault, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.v == nil {
		return nil, vault.ErrLocked
	}
	s.activity = time.Now()
	return s.v, nil
}

// Generation identifies the current unlock; it changes every time the vault
// is unlocked again. ok is false while locked.
func (s *Session) Generation() (gen uint64, ok bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.gen, s.v != nil
}

// SetIdle changes the auto-lock timeout for the running session.
func (s *Session) SetIdle(d time.Duration) {
	s.mu.Lock()
	s.idle = d
	s.mu.Unlock()
}

// Touch counts as activity without reading the vault.
func (s *Session) Touch() {
	s.mu.Lock()
	s.activity = time.Now()
	s.mu.Unlock()
}

// WatchIdle locks the vault after the configured idle time. It returns when stop is closed.
func (s *Session) WatchIdle(stop <-chan struct{}) {
	tick := time.NewTicker(10 * time.Second)
	defer tick.Stop()
	for {
		select {
		case <-stop:
			return
		case now := <-tick.C:
			s.mu.Lock()
			idle := s.idle
			expired := s.v != nil && now.Sub(s.activity) >= idle
			s.mu.Unlock()
			if expired {
				s.Lock(fmt.Sprintf("idle %dm", int(idle.Minutes())))
			}
		}
	}
}

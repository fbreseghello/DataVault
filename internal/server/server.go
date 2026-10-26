// Package server serves the DataVault frontend and API on 127.0.0.1.
//
// Every page and API call needs a session cookie, and the only way to get one
// is a single-use launch token that dtvlt hands to the browser it opens. The
// Host and Origin checks keep other sites (and DNS rebinding) out.
package server

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/session"
)

const launchTTL = 60 * time.Second

type Options struct {
	Session *session.Session
	Config  config.Config
	// ConfigPath is where the settings screen saves changes; VaultPath is shown there.
	ConfigPath string
	VaultPath  string
	// Assets is the built frontend (index.html at its root).
	Assets fs.FS
	// OnStop is called when `dtvlt stop` or the browser asks the session to end.
	OnStop func()
}

type Server struct {
	opts         Options
	ln           net.Listener
	port         int
	host         string // exact Host header we accept
	origin       string // exact Origin we accept
	cookieName   string
	controlToken string
	http         *http.Server

	mu      sync.Mutex
	launch  map[string]time.Time // sha256(token) -> expiry
	cookies map[string]bool      // sha256(cookie)

	cfgMu sync.Mutex
	cfg   config.Config // current settings; the settings screen can change them

	sealMu       sync.Mutex // serializes seal PIN attempts so the backoff holds
	sealFailures int
	grantsMu     sync.Mutex
	grants       map[string]grant // sha256(unseal token) -> item it opens
}

// New listens on a random port of 127.0.0.1. Call Serve to start handling requests.
func New(opts Options) (*Server, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, err
	}
	port := ln.Addr().(*net.TCPAddr).Port
	s := &Server{
		opts:         opts,
		ln:           ln,
		port:         port,
		host:         fmt.Sprintf("127.0.0.1:%d", port),
		origin:       fmt.Sprintf("http://127.0.0.1:%d", port),
		cookieName:   fmt.Sprintf("dtvlt_%d", port), // cookies ignore ports, so the name carries it
		controlToken: randomToken(),
		launch:       map[string]time.Time{},
		cookies:      map[string]bool{},
		cfg:          opts.Config,
		grants:       map[string]grant{},
	}
	s.http = &http.Server{
		Handler:           s.routes(),
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       30 * time.Second,
		WriteTimeout:      30 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	return s, nil
}

func (s *Server) Port() int            { return s.port }
func (s *Server) URL() string          { return s.origin }
func (s *Server) ControlToken() string { return s.controlToken }

func (s *Server) Serve() error {
	err := s.http.Serve(s.ln)
	if errors.Is(err, http.ErrServerClosed) {
		return nil
	}
	return err
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.http.Shutdown(ctx)
}

// LaunchURL returns a link that signs one browser into this session. It works
// once and expires after a minute.
func (s *Server) LaunchURL() string {
	token := randomToken()
	s.mu.Lock()
	now := time.Now()
	for k, exp := range s.launch {
		if now.After(exp) {
			delete(s.launch, k)
		}
	}
	s.launch[digest(token)] = now.Add(launchTTL)
	s.mu.Unlock()
	return s.origin + "/auth?t=" + token
}

func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /auth", s.handleAuth)

	api := http.NewServeMux()
	api.HandleFunc("GET /api/session", s.handleSession)
	api.HandleFunc("POST /api/unlock", s.handleUnlock)
	api.HandleFunc("POST /api/lock", s.handleLock)
	api.HandleFunc("POST /api/session/end", s.handleEndSession)
	api.HandleFunc("GET /api/items", s.handleListItems)
	api.HandleFunc("POST /api/items", s.handleCreateItem)
	api.HandleFunc("GET /api/items/{id}", s.handleGetItem)
	api.HandleFunc("PUT /api/items/{id}", s.handleUpdateItem)
	api.HandleFunc("DELETE /api/items/{id}", s.handleDeleteItem)
	api.HandleFunc("POST /api/items/{id}/unseal", s.handleUnseal)
	api.HandleFunc("POST /api/items/{id}/renewed", s.handleRenewed)
	api.HandleFunc("GET /api/notes", s.handleListNotes)
	api.HandleFunc("POST /api/notes", s.handleCreateNote)
	api.HandleFunc("GET /api/notes/{id}", s.handleGetNote)
	api.HandleFunc("PUT /api/notes/{id}", s.handleUpdateNote)
	api.HandleFunc("DELETE /api/notes/{id}", s.handleDeleteNote)
	api.HandleFunc("POST /api/notes/{id}/bury", s.handleBuryNote)
	api.HandleFunc("POST /api/notes/{id}/unseal", s.handleUnsealNote)
	api.HandleFunc("GET /api/tasks", s.handleListTasks)
	api.HandleFunc("POST /api/tasks", s.handleCreateTask)
	api.HandleFunc("POST /api/tasks/carry", s.handleCarryTasks)
	api.HandleFunc("PUT /api/tasks/{id}", s.handleUpdateTask)
	api.HandleFunc("POST /api/tasks/{id}/move", s.handleMoveTask)
	api.HandleFunc("DELETE /api/tasks/{id}", s.handleDeleteTask)
	api.HandleFunc("GET /api/files", s.handleListFiles)
	api.HandleFunc("POST /api/files", s.handleAddFile)
	api.HandleFunc("GET /api/files/{id}/content", s.handleFileContent)
	api.HandleFunc("PUT /api/files/{id}", s.handleRenameFile)
	api.HandleFunc("DELETE /api/files/{id}", s.handleDeleteFile)
	api.HandleFunc("GET /api/seal", s.handleSealStatus)
	api.HandleFunc("PUT /api/seal", s.handleSetSealPIN)
	api.HandleFunc("POST /api/events", s.handleEvent)
	api.HandleFunc("GET /api/settings", s.handleGetSettings)
	api.HandleFunc("PUT /api/settings", s.handlePutSettings)
	mux.Handle("/api/", s.requireCookie(s.requireSameOrigin(api)))

	control := http.NewServeMux()
	control.HandleFunc("GET /control/status", s.handleControlStatus)
	control.HandleFunc("POST /control/lock", s.handleControlLock)
	control.HandleFunc("POST /control/stop", s.handleControlStop)
	control.HandleFunc("POST /control/launch", s.handleControlLaunch)
	mux.Handle("/control/", s.requireControlToken(control))

	mux.Handle("/", s.requireCookie(s.static()))

	return s.securityHeaders(s.checkHost(mux))
}

func (s *Server) securityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h := w.Header()
		h.Set("Content-Security-Policy", "default-src 'self'; script-src 'self'; style-src 'self'; "+
			"img-src 'self' data:; font-src 'self'; connect-src 'self'; "+
			"frame-ancestors 'none'; base-uri 'none'; form-action 'self'; object-src 'none'")
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "no-referrer")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-origin")
		h.Set("Cache-Control", "no-store")
		next.ServeHTTP(w, r)
	})
}

// checkHost rejects any Host other than 127.0.0.1:<port>, which defeats DNS rebinding.
func (s *Server) checkHost(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Host != s.host {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) handleAuth(w http.ResponseWriter, r *http.Request) {
	key := digest(r.URL.Query().Get("t"))
	s.mu.Lock()
	exp, ok := s.launch[key]
	delete(s.launch, key)
	s.mu.Unlock()
	if !ok || time.Now().After(exp) {
		writePlain(w, http.StatusForbidden, "This link has expired or was already used.\nRun dtvlt to open DataVault again.")
		return
	}
	cookie := randomToken()
	s.mu.Lock()
	s.cookies[digest(cookie)] = true
	s.mu.Unlock()
	http.SetCookie(w, &http.Cookie{
		Name:     s.cookieName,
		Value:    cookie,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (s *Server) requireCookie(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		c, err := r.Cookie(s.cookieName)
		ok := err == nil
		if ok {
			s.mu.Lock()
			ok = s.cookies[digest(c.Value)]
			s.mu.Unlock()
		}
		if !ok {
			if strings.HasPrefix(r.URL.Path, "/api/") {
				writeError(w, http.StatusUnauthorized, "no session")
				return
			}
			writePlain(w, http.StatusUnauthorized, "No session.\nRun dtvlt to open DataVault.")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// requireSameOrigin guards state-changing requests: browsers will not send a
// custom header cross-origin without a CORS preflight, which we never answer.
func (s *Server) requireSameOrigin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			if r.Header.Get("X-Dtvlt") != "1" {
				writeError(w, http.StatusForbidden, "forbidden")
				return
			}
			if o := r.Header.Get("Origin"); o != "" && o != s.origin {
				writeError(w, http.StatusForbidden, "forbidden")
				return
			}
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) requireControlToken(next http.Handler) http.Handler {
	want := []byte("Bearer " + s.controlToken)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := []byte(r.Header.Get("Authorization"))
		if subtle.ConstantTimeCompare(got, want) != 1 {
			writeError(w, http.StatusUnauthorized, "unauthorized")
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (s *Server) static() http.Handler {
	if s.opts.Assets == nil {
		return notBuilt()
	}
	if _, err := fs.Stat(s.opts.Assets, "index.html"); err != nil {
		return notBuilt()
	}
	return http.FileServerFS(s.opts.Assets)
}

func notBuilt() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writePlain(w, http.StatusServiceUnavailable,
			"The frontend was not built into this binary.\nRun: npm --prefix web ci && npm --prefix web run build, then rebuild dtvlt.")
	})
}

func randomToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}

// digest is how tokens are stored and looked up, so a map lookup never
// compares raw secrets.
func digest(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func writePlain(w http.ResponseWriter, status int, msg string) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(status)
	fmt.Fprintln(w, msg)
}

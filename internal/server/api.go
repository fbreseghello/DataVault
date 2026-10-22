package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/fbreseghello/DataVault/internal/vault"
	"github.com/fbreseghello/DataVault/internal/version"
)

const maxBody = 1 << 20

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	// Polling this must not count as activity, or the vault would never auto-lock.
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                      s.opts.Session.ID,
		"version":                 version.Version,
		"unlocked":                s.opts.Session.Unlocked(),
		"auto_lock_minutes":       s.config().AutoLockMinutes,
		"clipboard_clear_seconds": s.config().ClipboardClearSeconds,
		"rest_minutes":            s.config().RestMinutes,
		"rest_style":              s.config().RestStyle,
	})
}

func (s *Server) handleUnlock(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Keyword  string `json:"keyword"`
		Password string `json:"password"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	keyword, password := []byte(req.Keyword), []byte(req.Password)
	err := s.opts.Session.Unlock(keyword, password)
	clear(keyword)
	clear(password)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleLock(w http.ResponseWriter, r *http.Request) {
	s.opts.Session.Lock("")
	s.dropGrants()
	w.WriteHeader(http.StatusNoContent)
}

// handleEndSession ends the whole session from the browser, like `dtvlt stop`:
// the vault locks and the process shuts down. Locking alone keeps it running.
func (s *Server) handleEndSession(w http.ResponseWriter, r *http.Request) {
	s.opts.Session.Lock("")
	s.dropGrants()
	s.opts.Session.Log("ended from the browser")
	w.WriteHeader(http.StatusNoContent)
	if s.opts.OnStop != nil {
		go s.opts.OnStop()
	}
}

func (s *Server) handleListItems(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	items, err := v.List(r.URL.Query().Get("q"))
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, items)
}

func (s *Server) handleGetItem(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	it, err := v.Get(r.PathValue("id"))
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if it.Sealed && !s.granted(r, it.ID) {
		writeError(w, http.StatusForbidden, "sealed")
		return
	}
	writeJSON(w, http.StatusOK, it)
}

func (s *Server) handleCreateItem(w http.ResponseWriter, r *http.Request) {
	var it vault.Item
	if !readJSON(w, r, &it) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if it.Sealed && !requireSealPIN(w, v) {
		return
	}
	created, err := v.Add(it)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("added   " + printable(created.Name))
	s.grantSaved(w, r, created)
	writeJSON(w, http.StatusCreated, created)
}

func (s *Server) handleUpdateItem(w http.ResponseWriter, r *http.Request) {
	var it vault.Item
	if !readJSON(w, r, &it) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	id := r.PathValue("id")
	current, err := v.Get(id)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	// Editing or unsealing a sealed item needs it opened with the PIN first.
	if current.Sealed && !s.granted(r, id) {
		writeError(w, http.StatusForbidden, "sealed")
		return
	}
	if it.Sealed && !current.Sealed && !requireSealPIN(w, v) {
		return
	}
	updated, err := v.Update(id, it)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("updated " + printable(updated.Name))
	s.grantSaved(w, r, updated)
	writeJSON(w, http.StatusOK, updated)
}

func (s *Server) handleDeleteItem(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	id := r.PathValue("id")
	it, err := v.Get(id)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if it.Sealed && !s.granted(r, id) {
		writeError(w, http.StatusForbidden, "sealed")
		return
	}
	if err := v.Delete(id); err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("deleted " + printable(it.Name))
	w.WriteHeader(http.StatusNoContent)
}

// handleEvent lets the frontend report things only it sees, like copies, to the terminal log.
func (s *Server) handleEvent(w http.ResponseWriter, r *http.Request) {
	var ev struct {
		Type  string `json:"type"`
		Item  string `json:"item"`
		Field string `json:"field"`
	}
	if !readJSON(w, r, &ev) {
		return
	}
	switch ev.Type {
	case "copy":
		s.opts.Session.Touch()
		s.opts.Session.Log(fmt.Sprintf("copied  %s/%s · clears in %ds",
			printable(ev.Item), printable(ev.Field), s.config().ClipboardClearSeconds))
	default:
		writeError(w, http.StatusBadRequest, "unknown event")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// printable strips control characters so item names cannot inject terminal
// escape sequences into the log, and caps the length.
func printable(s string) string {
	s = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return -1
		}
		return r
	}, s)
	if r := []rune(s); len(r) > 64 {
		s = string(r[:63]) + "…"
	}
	return s
}

func readJSON(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxBody)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func writeVaultError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, vault.ErrInvalidCredentials):
		writeError(w, http.StatusUnauthorized, "invalid credentials")
	case errors.Is(err, vault.ErrLocked):
		writeError(w, http.StatusLocked, "locked")
	case errors.Is(err, vault.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	case errors.Is(err, vault.ErrInvalidItem):
		writeError(w, http.StatusBadRequest, strings.TrimPrefix(err.Error(), vault.ErrInvalidItem.Error()+": "))
	default:
		writeError(w, http.StatusInternalServerError, "internal error")
	}
}

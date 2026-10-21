package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"unicode"

	"github.com/fbreseghello/DataVault/internal/vault"
)

const maxBody = 1 << 20

func (s *Server) handleSession(w http.ResponseWriter, r *http.Request) {
	// Polling this must not count as activity, or the vault would never auto-lock.
	writeJSON(w, http.StatusOK, map[string]any{
		"id":                      s.opts.Session.ID,
		"unlocked":                s.opts.Session.Unlocked(),
		"auto_lock_minutes":       s.opts.Config.AutoLockMinutes,
		"clipboard_clear_seconds": s.opts.Config.ClipboardClearSeconds,
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
	w.WriteHeader(http.StatusNoContent)
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
	created, err := v.Add(it)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("added   " + printable(created.Name))
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
	updated, err := v.Update(r.PathValue("id"), it)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("updated " + printable(updated.Name))
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
	if err == nil {
		err = v.Delete(id)
	}
	if err != nil {
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
			printable(ev.Item), printable(ev.Field), s.opts.Config.ClipboardClearSeconds))
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

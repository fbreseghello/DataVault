package server

import (
	"errors"
	"net/http"

	"github.com/fbreseghello/DataVault/internal/vault"
)

// Unspoken notes follow the same seal rules as grimoire entries: a sealed note
// needs the seal PIN, which grants a token for that note until the vault locks.
// The terminal log never names a note.

func (s *Server) handleListNotes(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	notes, err := v.Notes(r.URL.Query().Get("buried") == "1")
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, notes)
}

// openNote loads a note the request may read: sealed notes need a grant.
func (s *Server) openNote(w http.ResponseWriter, r *http.Request) (*vault.Vault, vault.Note, bool) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return nil, vault.Note{}, false
	}
	n, err := v.Note(r.PathValue("id"))
	if err != nil {
		writeVaultError(w, err)
		return nil, vault.Note{}, false
	}
	if n.Sealed && !s.granted(r, n.ID) {
		writeError(w, http.StatusForbidden, "sealed")
		return nil, vault.Note{}, false
	}
	return v, n, true
}

func (s *Server) handleGetNote(w http.ResponseWriter, r *http.Request) {
	if _, n, ok := s.openNote(w, r); ok {
		writeJSON(w, http.StatusOK, n)
	}
}

func (s *Server) handleCreateNote(w http.ResponseWriter, r *http.Request) {
	var in vault.NoteInput
	if !readJSON(w, r, &in) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if in.Sealed && !requireSealPIN(w, v) {
		return
	}
	n, err := v.AddNote(in)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.grantSavedNote(w, r, n)
	writeJSON(w, http.StatusCreated, n)
}

func (s *Server) handleUpdateNote(w http.ResponseWriter, r *http.Request) {
	var in vault.NoteInput
	if !readJSON(w, r, &in) {
		return
	}
	v, cur, ok := s.openNote(w, r)
	if !ok {
		return
	}
	if in.Sealed && !cur.Sealed && !requireSealPIN(w, v) {
		return
	}
	n, err := v.UpdateNote(cur.ID, in)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.grantSavedNote(w, r, n)
	writeJSON(w, http.StatusOK, n)
}

func (s *Server) handleBuryNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Buried bool `json:"buried"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	v, cur, ok := s.openNote(w, r)
	if !ok {
		return
	}
	n, err := v.BuryNote(cur.ID, req.Buried)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if req.Buried {
		s.opts.Session.Log("a note was buried")
	} else {
		s.opts.Session.Log("a note was unburied")
	}
	writeJSON(w, http.StatusOK, n)
}

func (s *Server) handleDeleteNote(w http.ResponseWriter, r *http.Request) {
	v, cur, ok := s.openNote(w, r)
	if !ok {
		return
	}
	if err := v.DeleteNote(cur.ID); err != nil {
		if errors.Is(err, vault.ErrNotBuried) {
			writeError(w, http.StatusConflict, "bury the note before deleting it")
			return
		}
		writeVaultError(w, err)
		return
	}
	if cur.Buried {
		s.opts.Session.Log("a note was deleted")
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUnsealNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Pin string `json:"pin"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	n, err := v.Note(r.PathValue("id"))
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if !n.Sealed {
		writeJSON(w, http.StatusOK, map[string]any{"note": n, "token": ""})
		return
	}
	hash, err := v.SealPIN()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if !s.checkSealPIN(hash, req.Pin) {
		s.opts.Session.Log("unseal failed  a note")
		writeError(w, http.StatusForbidden, "wrong pin")
		return
	}
	tok, ok := s.issueGrant(n.ID)
	if !ok {
		writeVaultError(w, vault.ErrLocked)
		return
	}
	s.opts.Session.Log("unsealed a note")
	writeJSON(w, http.StatusOK, map[string]any{"note": n, "token": tok})
}

// grantSavedNote keeps a sealed note open for whoever just wrote it.
func (s *Server) grantSavedNote(w http.ResponseWriter, r *http.Request, n vault.Note) {
	if !n.Sealed || s.granted(r, n.ID) {
		return
	}
	if tok, ok := s.issueGrant(n.ID); ok {
		w.Header().Set(unsealHeader, tok)
	}
}

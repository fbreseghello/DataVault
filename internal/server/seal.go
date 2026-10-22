package server

import (
	"net/http"
	"time"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/session"
	"github.com/fbreseghello/DataVault/internal/vault"
)

// Sealed items keep their fields behind a second PIN, stored as a hash inside
// the vault. Unsealing returns a token for that one item; the frontend sends it
// back while the item is open and forgets it when the item is closed, which
// seals it again. Tokens also die when the vault locks.

const unsealHeader = "X-Dtvlt-Unseal"

type grant struct {
	id  string
	gen uint64 // the unlock this grant belongs to
}

// granted reports whether r carries a live unseal token for item id.
func (s *Server) granted(r *http.Request, id string) bool {
	tok := r.Header.Get(unsealHeader)
	if tok == "" {
		return false
	}
	gen, ok := s.opts.Session.Generation()
	if !ok {
		return false
	}
	key := digest(tok)
	s.grantsMu.Lock()
	defer s.grantsMu.Unlock()
	g, found := s.grants[key]
	if found && g.gen != gen {
		delete(s.grants, key)
		return false
	}
	return found && g.id == id
}

// issueGrant opens item id for the current unlock and returns the token.
func (s *Server) issueGrant(id string) (string, bool) {
	gen, ok := s.opts.Session.Generation()
	if !ok {
		return "", false
	}
	tok := randomToken()
	s.grantsMu.Lock()
	s.grants[digest(tok)] = grant{id: id, gen: gen}
	s.grantsMu.Unlock()
	return tok, true
}

// grantSaved keeps a sealed item open for whoever just wrote it: they typed
// its contents, so there is nothing to hide from them.
func (s *Server) grantSaved(w http.ResponseWriter, r *http.Request, it vault.Item) {
	if !it.Sealed || s.granted(r, it.ID) {
		return
	}
	if tok, ok := s.issueGrant(it.ID); ok {
		w.Header().Set(unsealHeader, tok)
	}
}

func (s *Server) dropGrants() {
	s.grantsMu.Lock()
	clear(s.grants)
	s.grantsMu.Unlock()
}

// checkSealPIN verifies pin against hash, one attempt at a time with the same
// growing delay as the login, so the PIN cannot be guessed quickly.
func (s *Server) checkSealPIN(hash, pin string) bool {
	s.sealMu.Lock()
	defer s.sealMu.Unlock()
	time.Sleep(session.Backoff(s.sealFailures))
	b := []byte(pin)
	ok := hash != "" && config.VerifyPIN(hash, b)
	clear(b)
	if ok {
		s.sealFailures = 0
	} else {
		s.sealFailures++
	}
	return ok
}

func (s *Server) handleSealStatus(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	hash, err := v.SealPIN()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]bool{"has_pin": hash != ""})
}

// handleSetSealPIN sets the seal PIN, or changes it given the current one.
func (s *Server) handleSetSealPIN(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Current string `json:"current"`
		Pin     string `json:"pin"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	hash, err := v.SealPIN()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if hash != "" && !s.checkSealPIN(hash, req.Current) {
		writeError(w, http.StatusForbidden, "wrong pin")
		return
	}
	if len([]rune(req.Pin)) < config.MinPIN {
		writeError(w, http.StatusBadRequest, "the pin needs at least 4 characters")
		return
	}
	pin := []byte(req.Pin)
	next := config.HashPIN(pin)
	clear(pin)
	if err := v.SetSealPIN(next); err != nil {
		writeVaultError(w, err)
		return
	}
	// A new PIN seals everything that was open.
	s.dropGrants()
	if hash == "" {
		s.opts.Session.Log("seal pin set")
	} else {
		s.opts.Session.Log("seal pin changed")
	}
	w.WriteHeader(http.StatusNoContent)
}

func (s *Server) handleUnseal(w http.ResponseWriter, r *http.Request) {
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
	it, err := v.Get(r.PathValue("id"))
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if !it.Sealed {
		writeJSON(w, http.StatusOK, map[string]any{"item": it, "token": ""})
		return
	}
	hash, err := v.SealPIN()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if !s.checkSealPIN(hash, req.Pin) {
		s.opts.Session.Log("unseal failed  " + printable(it.Name))
		writeError(w, http.StatusForbidden, "wrong pin")
		return
	}
	tok, ok := s.issueGrant(it.ID)
	if !ok {
		writeVaultError(w, vault.ErrLocked)
		return
	}
	s.opts.Session.Log("unsealed " + printable(it.Name))
	writeJSON(w, http.StatusOK, map[string]any{"item": it, "token": tok})
}

// requireSealPIN rejects sealing an item while no seal PIN exists.
func requireSealPIN(w http.ResponseWriter, v *vault.Vault) bool {
	hash, err := v.SealPIN()
	if err != nil {
		writeVaultError(w, err)
		return false
	}
	if hash == "" {
		writeError(w, http.StatusBadRequest, "set a seal pin first")
		return false
	}
	return true
}

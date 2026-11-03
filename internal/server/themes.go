package server

import (
	"errors"
	"io"
	"io/fs"
	"net/http"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/theme"
)

// themes reads the built-in themes and the imported ones saved next to the config.
func (s *Server) themes() theme.Store {
	if s.opts.ConfigPath == "" {
		return theme.Store{}
	}
	return theme.Store{Dir: theme.DirFor(s.opts.ConfigPath)}
}

// activeTheme is the theme every screen draws, the login included, so it is
// served with the session before the vault is open. A theme holds no secrets.
// The login art chosen in the settings takes the place of the theme's own.
func (s *Server) activeTheme() theme.Theme {
	cfg := s.config()
	t := s.themes().Active(cfg.Theme)
	if art, ok := theme.LoginArt(cfg.LoginArt); ok {
		t.Login = art
	}
	return t
}

func (s *Server) handleListThemes(w http.ResponseWriter, r *http.Request) {
	if !s.requireUnlocked(w) {
		return
	}
	writeJSON(w, http.StatusOK, s.themes().List())
}

// handleGetTheme returns a theme with the gaps filled from the default, so an
// exported file shows every color, word and icon there is to change.
func (s *Server) handleGetTheme(w http.ResponseWriter, r *http.Request) {
	if !s.requireUnlocked(w) {
		return
	}
	t, err := s.themes().Get(r.PathValue("id"))
	if err != nil {
		writeError(w, http.StatusNotFound, "no such theme")
		return
	}
	t = t.Resolved()
	t.Builtin = false
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleImportTheme(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(http.MaxBytesReader(w, r.Body, maxBody))
	if err != nil {
		writeError(w, http.StatusBadRequest, "the theme file is too large")
		return
	}
	if !s.requireUnlocked(w) {
		return
	}
	if s.opts.ConfigPath == "" {
		writeError(w, http.StatusInternalServerError, "themes cannot be saved")
		return
	}
	t, err := s.themes().Import(data)
	if err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	s.opts.Session.Log("theme imported: " + t.ID)
	writeJSON(w, http.StatusOK, t.Summary())
}

// handleDeleteTheme removes an imported theme. When it was the one in use,
// DataVault goes back to the default theme.
func (s *Server) handleDeleteTheme(w http.ResponseWriter, r *http.Request) {
	if !s.requireUnlocked(w) {
		return
	}
	id := r.PathValue("id")
	switch err := s.themes().Remove(id); {
	case errors.Is(err, theme.ErrBuiltin):
		writeError(w, http.StatusBadRequest, err.Error())
		return
	case errors.Is(err, fs.ErrNotExist):
		writeError(w, http.StatusNotFound, "no such theme")
		return
	case err != nil:
		writeError(w, http.StatusInternalServerError, "could not remove the theme")
		return
	}

	s.cfgMu.Lock()
	if s.cfg.Theme == id {
		next := s.cfg
		next.Theme = theme.Default
		if err := config.Save(s.opts.ConfigPath, next); err == nil {
			s.cfg = next
		}
	}
	s.cfgMu.Unlock()
	s.opts.Session.Log("theme removed: " + id)
	w.WriteHeader(http.StatusNoContent)
}

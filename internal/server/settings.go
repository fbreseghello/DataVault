package server

import (
	"net/http"
	"time"

	"github.com/fbreseghello/DataVault/internal/config"
)

func (s *Server) config() config.Config {
	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	return s.cfg
}

// settingsView is what the settings screen shows. The PIN hash never leaves the server.
type settingsView struct {
	PinEnabled            bool   `json:"pin_enabled"`
	KeepTerminal          bool   `json:"keep_terminal"`
	AutoLockMinutes       int    `json:"auto_lock_minutes"`
	ClipboardClearSeconds int    `json:"clipboard_clear_seconds"`
	CheckUpdates          bool   `json:"check_updates"`
	RestMinutes           int    `json:"rest_minutes"`
	RestStyle             string `json:"rest_style"`
	ConfigPath            string `json:"config_path"`
	VaultPath             string `json:"vault_path"`
}

func (s *Server) viewOf(c config.Config) settingsView {
	return settingsView{
		PinEnabled:            c.PinEnabled,
		KeepTerminal:          c.KeepTerminal,
		AutoLockMinutes:       c.AutoLockMinutes,
		ClipboardClearSeconds: c.ClipboardClearSeconds,
		CheckUpdates:          c.CheckUpdates,
		RestMinutes:           c.RestMinutes,
		RestStyle:             c.RestStyle,
		ConfigPath:            s.opts.ConfigPath,
		VaultPath:             s.opts.VaultPath,
	}
}

// Settings are only reachable with the vault unlocked: whoever can read the
// secrets may also change how dtvlt starts.
func (s *Server) requireUnlocked(w http.ResponseWriter) bool {
	if _, err := s.opts.Session.Vault(); err != nil {
		writeVaultError(w, err)
		return false
	}
	return true
}

func (s *Server) handleGetSettings(w http.ResponseWriter, r *http.Request) {
	if !s.requireUnlocked(w) {
		return
	}
	writeJSON(w, http.StatusOK, s.viewOf(s.config()))
}

func (s *Server) handlePutSettings(w http.ResponseWriter, r *http.Request) {
	var req struct {
		PinEnabled   bool `json:"pin_enabled"`
		KeepTerminal bool `json:"keep_terminal"`
		// Pin is a new PIN; empty keeps the current one.
		Pin                   string `json:"pin"`
		AutoLockMinutes       int    `json:"auto_lock_minutes"`
		ClipboardClearSeconds int    `json:"clipboard_clear_seconds"`
		CheckUpdates          bool   `json:"check_updates"`
		RestMinutes           int    `json:"rest_minutes"`
		// RestStyle left out keeps the current style.
		RestStyle string `json:"rest_style"`
	}
	if !readJSON(w, r, &req) || !s.requireUnlocked(w) {
		return
	}
	if s.opts.ConfigPath == "" {
		writeError(w, http.StatusInternalServerError, "settings cannot be saved")
		return
	}

	s.cfgMu.Lock()
	defer s.cfgMu.Unlock()
	next := s.cfg
	next.PinEnabled = req.PinEnabled
	next.KeepTerminal = req.KeepTerminal
	next.AutoLockMinutes = req.AutoLockMinutes
	next.ClipboardClearSeconds = req.ClipboardClearSeconds
	next.CheckUpdates = req.CheckUpdates
	next.RestMinutes = req.RestMinutes
	if req.RestStyle != "" {
		next.RestStyle = req.RestStyle
	}
	switch {
	case !req.PinEnabled:
		next.PinHash = ""
	case req.Pin != "":
		if len([]rune(req.Pin)) < config.MinPIN {
			writeError(w, http.StatusBadRequest, "the pin needs at least 4 characters")
			return
		}
		pin := []byte(req.Pin)
		next.PinHash = config.HashPIN(pin)
		clear(pin)
	case next.PinHash == "":
		writeError(w, http.StatusBadRequest, "choose a pin to turn it on")
		return
	}
	if err := next.Validate(); err != nil {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	if err := config.Save(s.opts.ConfigPath, next); err != nil {
		writeError(w, http.StatusInternalServerError, "could not save settings")
		return
	}
	s.cfg = next
	s.opts.Session.SetIdle(time.Duration(next.AutoLockMinutes) * time.Minute)
	s.opts.Session.Log("settings updated")
	writeJSON(w, http.StatusOK, s.viewOf(next))
}

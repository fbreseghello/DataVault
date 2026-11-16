package server

import (
	"fmt"
	"net/http"
	"slices"
	"strings"

	"github.com/fbreseghello/DataVault/internal/vault"
)

// The hollow: the character that grows with the vault. Its deeds are earned in
// the vault as the siege changes; this side shows them and logs them.

func (s *Server) handleHollow(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	h, err := v.Hollow()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h)
}

// handleHollowSeen marks the level as seen once the page has shown it.
func (s *Server) handleHollowSeen(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	h, err := v.SeeLevel()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, h)
}

// tracked runs change and logs what it did to the hollow, in the session log's
// tone: "+10 souls  task conquered", "-10 souls  task conquered undone",
// "level 8". Only the kind of deed is logged, never the title of a task.
func (s *Server) tracked(v *vault.Vault, change func() error) error {
	before, err := v.Hollow()
	if err != nil {
		return change()
	}
	if err := change(); err != nil {
		return err
	}
	after, err := v.Hollow()
	if err != nil {
		return nil
	}
	// an imported theme names the XP; the log never prints its control characters
	xp := printable(s.activeTheme().Words["xp"])
	sources := func(ds []vault.Deed) []string {
		out := make([]string, len(ds))
		for i, d := range ds {
			out[i] = d.Source
		}
		return out
	}
	was, is := sources(before.Deeds), sources(after.Deeds)
	for _, d := range slices.Backward(after.Deeds) {
		if !slices.Contains(was, d.Source) {
			s.opts.Session.Log(fmt.Sprintf("+%d %s  %s", d.XP, xp, kindName(d.Kind)))
		}
	}
	for _, d := range before.Deeds {
		if !slices.Contains(is, d.Source) {
			s.opts.Session.Log(fmt.Sprintf("-%d %s  %s undone", d.XP, xp, kindName(d.Kind)))
		}
	}
	if after.Level > before.Level {
		s.opts.Session.Log(fmt.Sprintf("level %d", after.Level))
	}
	return nil
}

// kindName is how the log writes a deed's kind: "task conquered".
func kindName(k vault.DeedKind) string {
	return strings.ReplaceAll(string(k), "_", " ")
}

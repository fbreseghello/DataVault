package server

import (
	"net/http"

	"github.com/fbreseghello/DataVault/internal/vault"
)

// Quick writes from the terminal: `dtvlt note` and `dtvlt task` add to the
// open vault without going through the browser. Both need the vault unlocked.

// WriteNote adds an Unspoken note. It is never sealed: sealing asks for the
// seal PIN, which belongs in the browser.
func (s *Server) WriteNote(title, body string) (vault.Note, error) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		return vault.Note{}, err
	}
	n, err := v.AddNote(vault.NoteInput{Title: title, Body: body})
	if err == nil {
		s.opts.Session.Log("a note was written from the terminal")
	}
	return n, err
}

// WriteTask adds an order to a day's Siege board.
func (s *Server) WriteTask(title, note, day string) (vault.Task, error) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		return vault.Task{}, err
	}
	t, err := v.AddTask(vault.TaskInput{Title: title, Note: note, Day: day})
	if err == nil {
		s.opts.Session.Log("ordered " + printable(t.Title))
	}
	return t, err
}

func (s *Server) handleControlNote(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	n, err := s.WriteNote(req.Title, req.Body)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, n)
}

func (s *Server) handleControlTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
		Note  string `json:"note"`
		Day   string `json:"day"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	t, err := s.WriteTask(req.Title, req.Note, req.Day)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

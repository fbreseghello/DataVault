package server

import (
	"net/http"

	"github.com/fbreseghello/DataVault/internal/vault"
)

// Siege: one board of tasks per day. The day is the browser's local date.

func (s *Server) handleListTasks(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	day := r.URL.Query().Get("day")
	tasks, err := v.Tasks(day)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	pending, last, err := v.PendingBefore(day)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{
		"tasks": tasks,
		// unfinished work left on earlier days, offered to be brought here
		"pending_before": pending,
		"pending_day":    last,
	})
}

func (s *Server) handleCreateTask(w http.ResponseWriter, r *http.Request) {
	var in vault.TaskInput
	if !readJSON(w, r, &in) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	t, err := v.AddTask(in)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, t)
}

func (s *Server) handleUpdateTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
		Note  string `json:"note"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	t, err := v.UpdateTask(r.PathValue("id"), req.Title, req.Note)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleMoveTask(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Column vault.Column `json:"column"`
		Index  int          `json:"index"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	t, err := v.MoveTask(r.PathValue("id"), req.Column, req.Index)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, t)
}

func (s *Server) handleDeleteTask(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	if err := v.DeleteTask(r.PathValue("id")); err != nil {
		writeVaultError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// handleCarryTasks brings unfinished work from earlier days onto the given day.
func (s *Server) handleCarryTasks(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Day string `json:"day"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	n, err := v.CarryPending(req.Day)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]int{"moved": n})
}

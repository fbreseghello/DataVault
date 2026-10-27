package server

import (
	"net/http"

	"github.com/fbreseghello/DataVault/internal/vault"
)

// The ledger: subscriptions and fixed expenses. When each comes due is worked
// out by the page, on the browser's calendar.

func (s *Server) handleListBills(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	bills, err := v.Bills()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, bills)
}

func (s *Server) handleCreateBill(w http.ResponseWriter, r *http.Request) {
	var in vault.BillInput
	if !readJSON(w, r, &in) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	b, err := v.AddBill(in)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("inscribed " + printable(b.Name))
	writeJSON(w, http.StatusCreated, b)
}

func (s *Server) handleUpdateBill(w http.ResponseWriter, r *http.Request) {
	var in vault.BillInput
	if !readJSON(w, r, &in) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	b, err := v.UpdateBill(r.PathValue("id"), in)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, b)
}

func (s *Server) handleDeleteBill(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	b, err := v.DeleteBill(r.PathValue("id"))
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("struck  " + printable(b.Name))
	w.WriteHeader(http.StatusNoContent)
}

package server

import (
	"errors"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"

	"github.com/fbreseghello/DataVault/internal/vault"
)

// The crypt keeps files. An upload is the raw file as the request body, with
// its name and title in headers (URI-encoded); the contents never pass through
// JSON and may be up to vault.MaxFileSize.

func (s *Server) handleListFiles(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	files, err := v.Files()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, files)
}

func (s *Server) handleAddFile(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	r.Body = http.MaxBytesReader(w, r.Body, vault.MaxFileSize)
	data, err := io.ReadAll(r.Body)
	if err != nil {
		var tooBig *http.MaxBytesError
		if errors.As(err, &tooBig) {
			writeTooLarge(w)
			return
		}
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	defer clear(data)
	name, _ := url.PathUnescape(r.Header.Get("X-File-Name"))
	title, _ := url.PathUnescape(r.Header.Get("X-File-Title"))
	typ, _, _ := mime.ParseMediaType(r.Header.Get("Content-Type"))
	f, err := v.AddFile(title, name, typ, data)
	if errors.Is(err, vault.ErrFileTooLarge) {
		writeTooLarge(w)
		return
	}
	if err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("buried  " + printable(f.Title))
	writeJSON(w, http.StatusCreated, f)
}

func writeTooLarge(w http.ResponseWriter) {
	writeError(w, http.StatusRequestEntityTooLarge, fmt.Sprintf("files can be up to %d MB", vault.MaxFileSize>>20))
}

// handleFileContent hands a file back as a download. It is never shown inline,
// so nothing kept in the crypt can run inside the DataVault page.
func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	f, data, err := v.FileContent(r.PathValue("id"))
	if err != nil {
		writeVaultError(w, err)
		return
	}
	defer clear(data)
	h := w.Header()
	h.Set("Content-Type", "application/octet-stream")
	h.Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": f.Name}))
	h.Set("Cache-Control", "no-store")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
	s.opts.Session.Log("taken   " + printable(f.Title))
}

func (s *Server) handleRenameFile(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Title string `json:"title"`
	}
	if !readJSON(w, r, &req) {
		return
	}
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	f, err := v.RenameFile(r.PathValue("id"), req.Title)
	if err != nil {
		writeVaultError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, f)
}

func (s *Server) handleDeleteFile(w http.ResponseWriter, r *http.Request) {
	v, err := s.opts.Session.Vault()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	id := r.PathValue("id")
	files, err := v.Files()
	if err != nil {
		writeVaultError(w, err)
		return
	}
	title := ""
	for _, f := range files {
		if f.ID == id {
			title = f.Title
		}
	}
	if err := v.DeleteFile(id); err != nil {
		writeVaultError(w, err)
		return
	}
	s.opts.Session.Log("destroyed " + printable(title))
	w.WriteHeader(http.StatusNoContent)
}

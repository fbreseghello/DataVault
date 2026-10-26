package vault

import (
	"encoding/hex"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"
)

// MaxFileSize caps one file kept in the crypt.
const MaxFileSize = 25 << 20

// ErrFileTooLarge is returned for files over MaxFileSize.
var ErrFileTooLarge = fmt.Errorf("%w: files can be up to %d MB", ErrInvalidItem, MaxFileSize>>20)

// File is a document kept in the crypt. Its contents are not in the vault
// payload: each one is encrypted with the vault's key into its own file in a
// "crypt" folder next to the vault, so the vault itself stays small.
type File struct {
	ID string `json:"id"`
	// Title is what the crypt shows; Name is the file's own name, used when it
	// is taken out again.
	Title     string    `json:"title"`
	Name      string    `json:"name"`
	Type      string    `json:"type"`
	Size      int64     `json:"size"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (v *Vault) cryptDir() string {
	return filepath.Join(filepath.Dir(v.path), "crypt")
}

func (v *Vault) blobPath(id string) string {
	return filepath.Join(v.cryptDir(), id+".dtvf")
}

// the file's id is bound into its encryption, so one blob cannot be passed off
// as another
func blobAAD(id string) []byte {
	return []byte("dtvlt-file:" + id)
}

func cleanFileText(s string) string {
	s = strings.TrimSpace(s)
	return strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, s)
}

// Files lists the crypt, by title.
func (v *Vault) Files() ([]File, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return nil, ErrLocked
	}
	out := slices.Clone(v.data.Files)
	if out == nil {
		out = []File{}
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Title) < strings.ToLower(out[j].Title) })
	return out, nil
}

// AddFile encrypts data into the crypt and records it. An empty title takes
// the file's name without its extension.
func (v *Vault) AddFile(title, name, typ string, data []byte) (File, error) {
	name = cleanFileText(filepath.Base(strings.ReplaceAll(name, "\\", "/")))
	if name == "" || name == "." || name == "/" {
		name = "file"
	}
	title = cleanFileText(title)
	if title == "" {
		title = strings.TrimSuffix(name, filepath.Ext(name))
		if title == "" {
			title = name
		}
	}
	if len(data) > MaxFileSize {
		return File{}, ErrFileTooLarge
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return File{}, ErrLocked
	}
	now := time.Now().UTC()
	f := File{
		ID:        hex.EncodeToString(randomBytes(8)),
		Title:     title,
		Name:      name,
		Type:      cleanFileText(typ),
		Size:      int64(len(data)),
		CreatedAt: now,
		UpdatedAt: now,
	}
	box, err := seal(v.dek, data, blobAAD(f.ID))
	if err != nil {
		return File{}, err
	}
	if err := writeBlob(v.blobPath(f.ID), box); err != nil {
		return File{}, err
	}
	next := v.data
	next.Files = append(slices.Clone(v.data.Files), f)
	if err := v.save(next); err != nil {
		os.Remove(v.blobPath(f.ID))
		return File{}, err
	}
	v.data = next
	return f, nil
}

// FileContent decrypts a file from the crypt.
func (v *Vault) FileContent(id string) (File, []byte, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return File{}, nil, ErrLocked
	}
	i := v.fileIndex(id)
	if i < 0 {
		return File{}, nil, ErrNotFound
	}
	box, err := os.ReadFile(v.blobPath(id))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return File{}, nil, ErrCorrupt
		}
		return File{}, nil, err
	}
	data, err := open(v.dek, box, blobAAD(id))
	if err != nil {
		return File{}, nil, ErrCorrupt
	}
	return v.data.Files[i], data, nil
}

// RenameFile changes the title a file is shown under.
func (v *Vault) RenameFile(id, title string) (File, error) {
	title = cleanFileText(title)
	if title == "" {
		return File{}, fmt.Errorf("%w: a title is required", ErrInvalidItem)
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return File{}, ErrLocked
	}
	i := v.fileIndex(id)
	if i < 0 {
		return File{}, ErrNotFound
	}
	next := v.data
	next.Files = slices.Clone(v.data.Files)
	next.Files[i].Title = title
	next.Files[i].UpdatedAt = time.Now().UTC()
	if err := v.save(next); err != nil {
		return File{}, err
	}
	v.data = next
	return next.Files[i], nil
}

// DeleteFile drops a file from the crypt and removes its encrypted contents.
func (v *Vault) DeleteFile(id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return ErrLocked
	}
	i := v.fileIndex(id)
	if i < 0 {
		return ErrNotFound
	}
	next := v.data
	next.Files = slices.Delete(slices.Clone(v.data.Files), i, i+1)
	if err := v.save(next); err != nil {
		return err
	}
	v.data = next
	if err := os.Remove(v.blobPath(id)); err != nil && !errors.Is(err, fs.ErrNotExist) {
		return err
	}
	return nil
}

func (v *Vault) fileIndex(id string) int {
	return slices.IndexFunc(v.data.Files, func(f File) bool { return f.ID == id })
}

// writeBlob writes an encrypted file into place in one step.
func writeBlob(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(dir, ".file-*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), path)
}

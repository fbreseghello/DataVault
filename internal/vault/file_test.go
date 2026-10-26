package vault

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestCrypt(t *testing.T) {
	v, path := newVault(t)
	data := []byte("%PDF-1.7 a contract")
	f, err := v.AddFile("", `C:\scans\contract 2026.pdf`, "application/pdf", data)
	if err != nil {
		t.Fatal(err)
	}
	if f.Title != "contract 2026" || f.Name != "contract 2026.pdf" || f.Size != int64(len(data)) {
		t.Fatalf("file not recorded as expected: %+v", f)
	}
	// the contents are encrypted on disk, not in the vault payload
	blob, err := os.ReadFile(filepath.Join(filepath.Dir(path), "crypt", f.ID+".dtvf"))
	if err != nil {
		t.Fatal(err)
	}
	if bytes.Contains(blob, []byte("contract")) {
		t.Fatal("the file is stored in the clear")
	}
	if _, err := v.RenameFile(f.ID, "  rent contract "); err != nil {
		t.Fatal(err)
	}
	v.Lock()

	v2, err := Open(path, []byte("keyword"), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	files, _ := v2.Files()
	if len(files) != 1 || files[0].Title != "rent contract" {
		t.Fatalf("files after reopening: %+v", files)
	}
	got, content, err := v2.FileContent(f.ID)
	if err != nil || !bytes.Equal(content, data) || got.Name != "contract 2026.pdf" {
		t.Fatalf("content: %q %+v %v", content, got, err)
	}
	if err := v2.DeleteFile(f.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(filepath.Dir(path), "crypt", f.ID+".dtvf")); !os.IsNotExist(err) {
		t.Fatalf("the encrypted file should be gone: %v", err)
	}
	if _, _, err := v2.FileContent(f.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted file still readable: %v", err)
	}
}

func TestCryptLimitsAndTampering(t *testing.T) {
	v, path := newVault(t)
	if _, err := v.AddFile("big", "big.bin", "", make([]byte, MaxFileSize+1)); !errors.Is(err, ErrFileTooLarge) {
		t.Fatalf("an oversized file was accepted: %v", err)
	}
	a, _ := v.AddFile("a", "a.txt", "text/plain", []byte("first"))
	b, _ := v.AddFile("b", "b.txt", "text/plain", []byte("second"))
	// swapping two encrypted files on disk is caught, not served as the other
	dir := filepath.Join(filepath.Dir(path), "crypt")
	ba, _ := os.ReadFile(filepath.Join(dir, a.ID+".dtvf"))
	bb, _ := os.ReadFile(filepath.Join(dir, b.ID+".dtvf"))
	os.WriteFile(filepath.Join(dir, a.ID+".dtvf"), bb, 0o600)
	os.WriteFile(filepath.Join(dir, b.ID+".dtvf"), ba, 0o600)
	if _, _, err := v.FileContent(a.ID); !errors.Is(err, ErrCorrupt) {
		t.Fatalf("a swapped file was opened: %v", err)
	}
	if _, err := v.RenameFile(b.ID, "  "); !errors.Is(err, ErrInvalidItem) {
		t.Fatalf("an empty title was accepted: %v", err)
	}
}

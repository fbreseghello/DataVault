package config

import (
	"errors"
	"io/fs"
	"path/filepath"
	"testing"
)

func TestSaveLoad(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.toml")
	if _, err := Load(path); !errors.Is(err, fs.ErrNotExist) {
		t.Fatalf("missing file: got %v", err)
	}
	cfg := Default()
	cfg.PinEnabled = true
	cfg.PinHash = HashPIN([]byte("1234"))
	cfg.AutoLockMinutes = 5
	if err := Save(path, cfg); err != nil {
		t.Fatal(err)
	}
	got, err := Load(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != cfg {
		t.Fatalf("got %+v, want %+v", got, cfg)
	}
}

func TestPIN(t *testing.T) {
	h := HashPIN([]byte("1234"))
	if !VerifyPIN(h, []byte("1234")) {
		t.Fatal("correct pin rejected")
	}
	if VerifyPIN(h, []byte("1235")) {
		t.Fatal("wrong pin accepted")
	}
	if VerifyPIN("garbage", []byte("1234")) {
		t.Fatal("malformed hash accepted")
	}
}

func TestValidate(t *testing.T) {
	cfg := Default()
	cfg.PinEnabled = true
	if cfg.Validate() == nil {
		t.Fatal("pin enabled without hash should be invalid")
	}
}

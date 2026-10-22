package vault

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// fastKDF keeps tests quick; real vaults use DefaultKDF.
var fastKDF = KDFParams{MemoryKiB: 64, Iterations: 1, Threads: 1}

func newVault(t *testing.T) (*Vault, string) {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.dtv")
	v, err := Create(path, []byte("keyword"), []byte("password"), fastKDF)
	if err != nil {
		t.Fatal(err)
	}
	return v, path
}

func TestRoundTrip(t *testing.T) {
	v, path := newVault(t)
	added, err := v.Add(Item{
		Type:   TypeLogin,
		Name:   " github ",
		Fields: []Field{{Name: "user", Value: "me"}, {Name: "password", Value: "s3cret", Secret: true}, {}},
		Tags:   []string{"Work", "work", " "},
	})
	if err != nil {
		t.Fatal(err)
	}
	if added.Name != "github" || len(added.Fields) != 2 || len(added.Tags) != 1 || added.Tags[0] != "work" {
		t.Fatalf("item not normalized: %+v", added)
	}
	v.Lock()

	v2, err := Open(path, []byte("keyword"), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := v2.Get(added.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Fields[1].Value != "s3cret" || !got.Fields[1].Secret {
		t.Fatalf("secret field lost: %+v", got.Fields)
	}
}

func TestWrongCredentials(t *testing.T) {
	_, path := newVault(t)
	cases := []struct{ keyword, password string }{
		{"keyword", "wrong"},
		{"wrong", "password"},
		{"password", "keyword"},
		// Same concatenation as "keyword"+"password": the length prefix must tell them apart.
		{"keywordpass", "word"},
	}
	for _, c := range cases {
		if _, err := Open(path, []byte(c.keyword), []byte(c.password)); !errors.Is(err, ErrInvalidCredentials) {
			t.Errorf("Open(%q, %q) = %v, want ErrInvalidCredentials", c.keyword, c.password, err)
		}
	}
}

func TestTamperedFileIsRejected(t *testing.T) {
	_, path := newVault(t)
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, at := range []int{headerSize - 1, headerSize + 5, len(raw) - 1} {
		bad := append([]byte(nil), raw...)
		bad[at] ^= 0x01
		if err := os.WriteFile(path, bad, 0o600); err != nil {
			t.Fatal(err)
		}
		if _, err := Open(path, []byte("keyword"), []byte("password")); err == nil {
			t.Errorf("flipping byte %d was not detected", at)
		}
	}
}

func TestCreateRefusesToOverwrite(t *testing.T) {
	_, path := newVault(t)
	if _, err := Create(path, []byte("a"), []byte("b"), fastKDF); !errors.Is(err, ErrExists) {
		t.Fatalf("got %v, want ErrExists", err)
	}
}

func TestLockedVault(t *testing.T) {
	v, _ := newVault(t)
	v.Lock()
	if _, err := v.List(""); !errors.Is(err, ErrLocked) {
		t.Fatalf("got %v, want ErrLocked", err)
	}
}

func TestUpdateDeleteAndSearch(t *testing.T) {
	v, path := newVault(t)
	a, _ := v.Add(Item{Type: TypeAPIKey, Name: "openai", Tags: []string{"ai"}})
	b, _ := v.Add(Item{Type: TypeNote, Name: "Backup codes"})

	if _, err := v.Update(a.ID, Item{Type: TypeAPIKey, Name: "anthropic", Tags: []string{"ai"}}); err != nil {
		t.Fatal(err)
	}
	if res, _ := v.List("AI"); len(res) != 1 || res[0].Name != "anthropic" {
		t.Fatalf("search by tag: %+v", res)
	}
	if err := v.Delete(b.ID); err != nil {
		t.Fatal(err)
	}
	if err := v.Delete(b.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("second delete: %v", err)
	}
	if _, err := os.Stat(path + ".bak"); err != nil {
		t.Fatalf("backup not written: %v", err)
	}

	v.Lock()
	v2, _ := Open(path, []byte("keyword"), []byte("password"))
	if res, _ := v2.List(""); len(res) != 1 {
		t.Fatalf("after reopen: %+v", res)
	}
}

func TestInvalidItem(t *testing.T) {
	v, _ := newVault(t)
	for _, it := range []Item{{Type: TypeLogin}, {Type: "bogus", Name: "x"}, {Type: TypeNote, Name: "x", Fields: []Field{{Value: "orphan"}}}} {
		if _, err := v.Add(it); !errors.Is(err, ErrInvalidItem) {
			t.Errorf("Add(%+v) = %v, want ErrInvalidItem", it, err)
		}
	}
}

func TestSealPINSurvivesItemChanges(t *testing.T) {
	v, path := newVault(t)
	if err := v.SetSealPIN("hash"); err != nil {
		t.Fatal(err)
	}
	it, _ := v.Add(Item{Type: TypeNote, Name: "a", Sealed: true})
	v.Update(it.ID, Item{Type: TypeNote, Name: "b", Sealed: true})
	v.Add(Item{Type: TypeNote, Name: "c"})
	v.Delete(it.ID)
	v.Lock()

	v2, err := Open(path, []byte("keyword"), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	if h, _ := v2.SealPIN(); h != "hash" {
		t.Fatalf("seal pin lost: %q", h)
	}
}

func TestSummaryShowsSealed(t *testing.T) {
	v, _ := newVault(t)
	v.Add(Item{Type: TypeLogin, Name: "bank", Sealed: true, Fields: []Field{{Name: "password", Value: "x", Secret: true}}})
	list, _ := v.List("")
	if len(list) != 1 || !list[0].Sealed {
		t.Fatalf("summary: %+v", list)
	}
}

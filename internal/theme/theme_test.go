package theme

import (
	"strings"
	"testing"
)

func TestBuiltinsAreComplete(t *testing.T) {
	for id, th := range builtins {
		if err := th.complete(); err != nil {
			t.Errorf("%s: %v", id, err)
		}
	}
	for _, id := range []string{"souls", "arcane", "dungeon"} {
		if !IsBuiltin(id) {
			t.Errorf("%s should be built in", id)
		}
	}
}

func TestParseRejects(t *testing.T) {
	row := strings.Repeat(".", IconSize)
	icon := `["` + strings.Repeat(row+`","`, IconSize-1) + row + `"]`
	for name, body := range map[string]string{
		"bad id":        `{"id":"Bad Id","name":"x"}`,
		"no name":       `{"id":"x","name":" "}`,
		"unknown field": `{"id":"x","name":"x","font":"comic"}`,
		"unknown color": `{"id":"x","name":"x","colors":{"pink":"#fff"}}`,
		"unknown word":  `{"id":"x","name":"x","words":{"vault":"x"}}`,
		"long word":     `{"id":"x","name":"x","words":{"siege":"` + strings.Repeat("a", maxWord+1) + `"}}`,
		"short icon":    `{"id":"x","name":"x","icons":{"seal":["#"]}}`,
		"icon chars":    `{"id":"x","name":"x","icons":{"seal":` + strings.Replace(icon, ".", "x", 1) + `}}`,
		"large unknown": `{"id":"x","name":"x","large_icons":["moon"]}`,
		"built-in id":   `{"id":"souls","name":"souls"}`,
	} {
		if _, err := Parse([]byte(body)); err == nil {
			t.Errorf("%s: accepted", name)
		}
	}
	if _, err := Parse([]byte(`{"id":"x","name":"x","icons":{"seal":` + icon + `}}`)); err != nil {
		t.Errorf("a blank icon is still an icon: %v", err)
	}
}

func TestResolvedFillsFromDefault(t *testing.T) {
	th, err := Parse([]byte(`{"id":"x","name":"x","words":{"siege":"war"},"icons":{}}`))
	if err != nil {
		t.Fatal(err)
	}
	r := th.Resolved()
	if r.Words["siege"] != "war" || r.Words["crypt"] != "crypt" || r.Colors["bg"] == "" || len(r.Login) == 0 {
		t.Fatalf("resolved: %+v", r)
	}
	if err := r.complete(); err != nil {
		t.Fatal(err)
	}
	if len(r.LargeIcons) != 1 || r.LargeIcons[0] != "grimoire" {
		t.Fatalf("the default book should stay large while it is the default book: %v", r.LargeIcons)
	}
}

func TestStore(t *testing.T) {
	s := Store{Dir: t.TempDir()}
	if _, err := s.Import([]byte(`{"id":"moss","name":"moss","colors":{"fg":"#cfc"}}`)); err != nil {
		t.Fatal(err)
	}
	if got := s.Active("moss"); got.Colors["fg"] != "#cfc" || got.ID != "moss" {
		t.Fatalf("active: %+v", got)
	}
	if got := s.Active("gone"); got.ID != Default {
		t.Fatalf("a missing theme should fall back to the default, got %q", got.ID)
	}
	if err := s.Remove("souls"); err != ErrBuiltin {
		t.Fatalf("remove built-in: %v", err)
	}
	if err := s.Remove("moss"); err != nil {
		t.Fatal(err)
	}
	if got := s.Active("moss"); got.ID != Default {
		t.Fatalf("after removing: %q", got.ID)
	}
}

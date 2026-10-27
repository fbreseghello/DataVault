package forge

import (
	"strings"
	"testing"
)

func TestForge(t *testing.T) {
	seen := map[string]bool{}
	for range 300 {
		p, err := Forge(Defaults())
		if err != nil {
			t.Fatal(err)
		}
		if len(p) != DefaultLength {
			t.Fatalf("length %d, want %d", len(p), DefaultLength)
		}
		for _, set := range []string{upper, lower, digits, symbols} {
			if !strings.ContainsAny(p, set) {
				t.Fatalf("%q has nothing from %q", p, set)
			}
		}
		if seen[p] {
			t.Fatalf("%q forged twice", p)
		}
		seen[p] = true
	}
}

func TestForgePlain(t *testing.T) {
	o := Defaults()
	o.Plain = true
	o.Length = MaxLength
	for range 100 {
		p, _ := Forge(o)
		if strings.ContainsAny(p, lookalikes) {
			t.Fatalf("%q has a lookalike", p)
		}
	}
}

func TestForgeOptions(t *testing.T) {
	p, err := Forge(Options{Length: 3, Digits: true})
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != MinLength || strings.Trim(p, digits) != "" {
		t.Fatalf("got %q, want %d digits", p, MinLength)
	}
	if _, err := Forge(Options{Length: 20}); err != ErrNoKinds {
		t.Fatalf("no kinds: got %v", err)
	}
	if b := Bits(Options{Length: 10, Digits: true}); b != 33 {
		t.Fatalf("bits %d, want 33", b)
	}
}

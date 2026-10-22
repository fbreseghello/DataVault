package vault

import (
	"errors"
	"strings"
	"testing"
)

func TestNotesRoundTrip(t *testing.T) {
	v, path := newVault(t)
	n, err := v.AddNote(NoteInput{Title: " tuesday ", Body: "it hurt more than I said\nsecond line"})
	if err != nil {
		t.Fatal(err)
	}
	if n.Title != "tuesday" {
		t.Fatalf("title not trimmed: %q", n.Title)
	}
	if _, err := v.UpdateNote(n.ID, NoteInput{Title: "tuesday", Body: "rewritten", Sealed: true}); err != nil {
		t.Fatal(err)
	}
	v.Lock()

	v2, err := Open(path, []byte("keyword"), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := v2.Note(n.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Body != "rewritten" || !got.Sealed {
		t.Fatalf("note after reopen: %+v", got)
	}
}

func TestNoteSummaryHidesSealedText(t *testing.T) {
	v, _ := newVault(t)
	v.AddNote(NoteInput{Title: "open", Body: "first line\nsecond"})
	v.AddNote(NoteInput{Title: "secret title", Body: "secret body", Sealed: true})
	list, _ := v.Notes(false)
	if len(list) != 2 {
		t.Fatalf("list: %+v", list)
	}
	for _, s := range list {
		if s.Sealed && (s.Title != "" || s.Preview != "") {
			t.Fatalf("sealed note leaked in the list: %+v", s)
		}
		if !s.Sealed && s.Preview != "first line" {
			t.Fatalf("preview: %q", s.Preview)
		}
	}
	long, _ := v.AddNote(NoteInput{Body: strings.Repeat("x", 200)})
	list, _ = v.Notes(false)
	for _, s := range list {
		if s.ID == long.ID && len([]rune(s.Preview)) != previewLen {
			t.Fatalf("preview not cut: %d runes", len([]rune(s.Preview)))
		}
	}
}

func TestBuryAndDelete(t *testing.T) {
	v, _ := newVault(t)
	n, _ := v.AddNote(NoteInput{Body: "words"})
	if err := v.DeleteNote(n.ID); !errors.Is(err, ErrNotBuried) {
		t.Fatalf("deleting an unburied note: %v", err)
	}
	if _, err := v.BuryNote(n.ID, true); err != nil {
		t.Fatal(err)
	}
	if live, _ := v.Notes(false); len(live) != 0 {
		t.Fatalf("buried note still listed: %+v", live)
	}
	if dead, _ := v.Notes(true); len(dead) != 1 {
		t.Fatalf("dead words: %+v", dead)
	}
	if err := v.DeleteNote(n.ID); err != nil {
		t.Fatalf("deleting a buried note: %v", err)
	}

	empty, _ := v.AddNote(NoteInput{})
	if err := v.DeleteNote(empty.ID); err != nil {
		t.Fatalf("deleting an empty note: %v", err)
	}
}

func TestNotesKeepItemsAndSealPIN(t *testing.T) {
	v, path := newVault(t)
	v.SetSealPIN("hash")
	it, _ := v.Add(Item{Type: TypeNote, Name: "x"})
	v.AddNote(NoteInput{Body: "a"})
	v.Lock()
	v2, _ := Open(path, []byte("keyword"), []byte("password"))
	if _, err := v2.Get(it.ID); err != nil {
		t.Fatal("item lost after adding a note")
	}
	if h, _ := v2.SealPIN(); h != "hash" {
		t.Fatal("seal pin lost after adding a note")
	}
}

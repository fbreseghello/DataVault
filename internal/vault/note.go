package vault

import (
	"encoding/hex"
	"errors"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode/utf8"
)

// Note is an Unspoken entry: free text, optionally sealed, and possibly buried
// in Dead Words.
type Note struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Sealed    bool      `json:"sealed"`
	Buried    bool      `json:"buried"`
	BuriedAt  time.Time `json:"buried_at,omitzero"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NoteSummary is what the list shows. A sealed note shows neither its title
// nor its first line.
type NoteSummary struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Preview   string    `json:"preview"`
	Sealed    bool      `json:"sealed"`
	Buried    bool      `json:"buried"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NoteInput is what the editor sends.
type NoteInput struct {
	Title  string `json:"title"`
	Body   string `json:"body"`
	Sealed bool   `json:"sealed"`
}

var ErrNotBuried = errors.New("only buried or empty notes can be deleted")

const previewLen = 60

func (n Note) summary() NoteSummary {
	s := NoteSummary{ID: n.ID, Sealed: n.Sealed, Buried: n.Buried, CreatedAt: n.CreatedAt, UpdatedAt: n.UpdatedAt}
	if n.Sealed {
		return s
	}
	s.Title = n.Title
	first, _, _ := strings.Cut(strings.TrimSpace(n.Body), "\n")
	if utf8.RuneCountInString(first) > previewLen {
		first = string([]rune(first)[:previewLen-1]) + "…"
	}
	s.Preview = first
	return s
}

func (n Note) empty() bool {
	return strings.TrimSpace(n.Title) == "" && strings.TrimSpace(n.Body) == ""
}

// Notes lists the notes on one side of Dead Words, most recently written first.
func (v *Vault) Notes(buried bool) ([]NoteSummary, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return nil, ErrLocked
	}
	out := []NoteSummary{}
	for _, n := range v.data.Notes {
		if n.Buried == buried {
			out = append(out, n.summary())
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].UpdatedAt.After(out[j].UpdatedAt) })
	return out, nil
}

func (v *Vault) Note(id string) (Note, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Note{}, ErrLocked
	}
	i := v.noteIndex(id)
	if i < 0 {
		return Note{}, ErrNotFound
	}
	return v.data.Notes[i], nil
}

func (v *Vault) AddNote(in NoteInput) (Note, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Note{}, ErrLocked
	}
	now := time.Now().UTC()
	n := Note{
		ID:        hex.EncodeToString(randomBytes(8)),
		Title:     strings.TrimSpace(in.Title),
		Body:      in.Body,
		Sealed:    in.Sealed,
		CreatedAt: now,
		UpdatedAt: now,
	}
	next := v.data
	next.Notes = append(slices.Clone(v.data.Notes), n)
	if err := v.save(next); err != nil {
		return Note{}, err
	}
	v.data = next
	return n, nil
}

func (v *Vault) UpdateNote(id string, in NoteInput) (Note, error) {
	return v.changeNote(id, func(n *Note) error {
		n.Title = strings.TrimSpace(in.Title)
		n.Body = in.Body
		n.Sealed = in.Sealed
		n.UpdatedAt = time.Now().UTC()
		return nil
	})
}

// BuryNote moves a note into Dead Words, or back out of it.
func (v *Vault) BuryNote(id string, buried bool) (Note, error) {
	return v.changeNote(id, func(n *Note) error {
		n.Buried = buried
		n.BuriedAt = time.Time{}
		if buried {
			n.BuriedAt = time.Now().UTC()
		}
		return nil
	})
}

// DeleteNote removes a note for good. Only buried notes, or notes with nothing
// written in them, can go.
func (v *Vault) DeleteNote(id string) error {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return ErrLocked
	}
	i := v.noteIndex(id)
	if i < 0 {
		return ErrNotFound
	}
	if n := v.data.Notes[i]; !n.Buried && !n.empty() {
		return ErrNotBuried
	}
	next := v.data
	next.Notes = slices.Delete(slices.Clone(v.data.Notes), i, i+1)
	if err := v.save(next); err != nil {
		return err
	}
	v.data = next
	return nil
}

func (v *Vault) changeNote(id string, change func(*Note) error) (Note, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Note{}, ErrLocked
	}
	i := v.noteIndex(id)
	if i < 0 {
		return Note{}, ErrNotFound
	}
	next := v.data
	next.Notes = slices.Clone(v.data.Notes)
	if err := change(&next.Notes[i]); err != nil {
		return Note{}, err
	}
	if err := v.save(next); err != nil {
		return Note{}, err
	}
	v.data = next
	return next.Notes[i], nil
}

func (v *Vault) noteIndex(id string) int {
	return slices.IndexFunc(v.data.Notes, func(n Note) bool { return n.ID == id })
}

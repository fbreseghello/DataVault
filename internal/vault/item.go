package vault

import (
	"fmt"
	"strings"
	"time"
)

// ItemType is the kind of secret an item holds. The frontend uses it to pick
// a field template; the vault only checks that it is known.
type ItemType string

const (
	TypeLogin  ItemType = "login"
	TypeAPIKey ItemType = "api_key"
	TypeSSHKey ItemType = "ssh_key"
	TypeSeed   ItemType = "seed"
	TypeNote   ItemType = "note"
	// TypeCommand is a saved server command; dtvlt stores and copies it, never runs it.
	TypeCommand ItemType = "command"
)

func (t ItemType) valid() bool {
	switch t {
	case TypeLogin, TypeAPIKey, TypeSSHKey, TypeSeed, TypeNote, TypeCommand:
		return true
	}
	return false
}

// Field is one named value of an item. Secret fields are masked by the UI.
type Field struct {
	Name   string `json:"name"`
	Value  string `json:"value"`
	Secret bool   `json:"secret"`
}

type Item struct {
	ID     string   `json:"id"`
	Type   ItemType `json:"type"`
	Name   string   `json:"name"`
	Fields []Field  `json:"fields"`
	Tags   []string `json:"tags"`
	Notes  string   `json:"notes"`
	// Sealed items need the seal PIN before their fields are shown.
	Sealed bool `json:"sealed"`
	// Renewal, for keys that expire: every RenewEvery days counted from the
	// last renewal, or by a fixed date RenewBy (YYYY-MM-DD). See RenewDue.
	RenewEvery int        `json:"renew_every,omitempty"`
	RenewBy    string     `json:"renew_by,omitempty"`
	RenewedAt  *time.Time `json:"renewed_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// Summary is an item without its values, for listings.
type Summary struct {
	ID     string   `json:"id"`
	Type   ItemType `json:"type"`
	Name   string   `json:"name"`
	Tags   []string `json:"tags"`
	Sealed bool     `json:"sealed"`
	// RenewDue is when the item is due for renewal (YYYY-MM-DD), if ever.
	RenewDue  string    `json:"renew_due,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (it Item) summary() Summary {
	return Summary{ID: it.ID, Type: it.Type, Name: it.Name, Tags: it.Tags, Sealed: it.Sealed, RenewDue: it.RenewDue(), UpdatedAt: it.UpdatedAt}
}

// normalize trims user input and rejects items the vault should not store.
func (it *Item) normalize() error {
	it.Name = strings.TrimSpace(it.Name)
	if it.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidItem)
	}
	if !it.Type.valid() {
		return fmt.Errorf("%w: unknown type %q", ErrInvalidItem, it.Type)
	}
	fields := make([]Field, 0, len(it.Fields))
	for _, f := range it.Fields {
		f.Name = strings.TrimSpace(f.Name)
		if f.Name == "" && f.Value == "" {
			continue
		}
		if f.Name == "" {
			return fmt.Errorf("%w: field without a name", ErrInvalidItem)
		}
		fields = append(fields, f)
	}
	it.Fields = fields
	tags := make([]string, 0, len(it.Tags))
	seen := map[string]bool{}
	for _, t := range it.Tags {
		t = strings.ToLower(strings.TrimSpace(t))
		if t != "" && !seen[t] {
			seen[t] = true
			tags = append(tags, t)
		}
	}
	it.Tags = tags
	if it.RenewEvery < 0 || it.RenewEvery > MaxRenewDays {
		return fmt.Errorf("%w: renew every must be between 0 (off) and %d days", ErrInvalidItem, MaxRenewDays)
	}
	it.RenewBy = strings.TrimSpace(it.RenewBy)
	if it.RenewBy != "" {
		if _, err := time.Parse(time.DateOnly, it.RenewBy); err != nil {
			return fmt.Errorf("%w: renew by must be a date (YYYY-MM-DD)", ErrInvalidItem)
		}
	}
	return nil
}

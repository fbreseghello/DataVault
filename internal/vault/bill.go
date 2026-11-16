package vault

import (
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

// Cycle is how often a ledger bill comes due.
type Cycle string

const (
	Weekly  Cycle = "weekly"
	Monthly Cycle = "monthly"
	Yearly  Cycle = "yearly"
)

func (c Cycle) valid() bool {
	switch c {
	case Weekly, Monthly, Yearly:
		return true
	}
	return false
}

// MaxAmount caps a bill at a trillion, in cents.
const MaxAmount = 1_000_000_000_000_00

// Bill is a ledger entry: a subscription or fixed expense that comes due again
// and again. Amount is in cents of Currency. Since is the first charge, as the
// user's local date (YYYY-MM-DD); the next ones follow from it and the cycle.
// Until, when set, is the day it ends: no charge falls on it or after it.
// PauseFrom and PauseUntil skip the charges that fall between them (PauseUntil
// excluded); a pause with no PauseUntil holds until it is lifted.
type Bill struct {
	ID         string        `json:"id"`
	Name       string        `json:"name"`
	Amount     int64         `json:"amount"`
	Currency   string        `json:"currency"`
	Cycle      Cycle         `json:"cycle"`
	Since      string        `json:"since"`
	Until      string        `json:"until,omitempty"`
	PauseFrom  string        `json:"pause_from,omitempty"`
	PauseUntil string        `json:"pause_until,omitempty"`
	Method     string        `json:"method"`
	Note       string        `json:"note"`
	History    []PriceChange `json:"history,omitempty"`
	CreatedAt  time.Time     `json:"created_at"`
	UpdatedAt  time.Time     `json:"updated_at"`
}

// PriceChange is a price a bill used to have: Amount was charged on the days
// before Until, the local date it changed. The history is oldest first, is
// written by the vault when the amount changes, and starts over when the
// currency does.
type PriceChange struct {
	Amount int64  `json:"amount"`
	Until  string `json:"until"`
}

// BillInput is what the ledger sends when writing a bill.
type BillInput struct {
	Name       string `json:"name"`
	Amount     int64  `json:"amount"`
	Currency   string `json:"currency"`
	Cycle      Cycle  `json:"cycle"`
	Since      string `json:"since"`
	Until      string `json:"until,omitempty"`
	PauseFrom  string `json:"pause_from,omitempty"`
	PauseUntil string `json:"pause_until,omitempty"`
	Method     string `json:"method"`
	Note       string `json:"note"`
}

// optionalDate trims s and checks it is empty or a YYYY-MM-DD date.
func optionalDate(s *string, what string) error {
	*s = strings.TrimSpace(*s)
	if *s == "" {
		return nil
	}
	if _, err := time.Parse(time.DateOnly, *s); err != nil {
		return fmt.Errorf("%w: %s must be a date (YYYY-MM-DD)", ErrInvalidItem, what)
	}
	return nil
}

func (in *BillInput) normalize() error {
	in.Name = strings.TrimSpace(in.Name)
	in.Currency = strings.ToUpper(strings.TrimSpace(in.Currency))
	in.Method = strings.TrimSpace(in.Method)
	if in.Name == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidItem)
	}
	if in.Amount < 0 || in.Amount > MaxAmount {
		return fmt.Errorf("%w: amount must be zero or more", ErrInvalidItem)
	}
	if len(in.Currency) != 3 || strings.Trim(in.Currency, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") != "" {
		return fmt.Errorf("%w: currency must be a three-letter code, like BRL", ErrInvalidItem)
	}
	if in.Cycle == "" {
		in.Cycle = Monthly
	}
	if !in.Cycle.valid() {
		return fmt.Errorf("%w: unknown cycle %q", ErrInvalidItem, in.Cycle)
	}
	if _, err := time.Parse(time.DateOnly, in.Since); err != nil {
		return fmt.Errorf("%w: first charge must be a date (YYYY-MM-DD)", ErrInvalidItem)
	}
	if err := optionalDate(&in.Until, "end"); err != nil {
		return err
	}
	if in.Until != "" && in.Until <= in.Since {
		return fmt.Errorf("%w: end must come after the first charge", ErrInvalidItem)
	}
	if err := optionalDate(&in.PauseFrom, "pause start"); err != nil {
		return err
	}
	if err := optionalDate(&in.PauseUntil, "pause end"); err != nil {
		return err
	}
	if in.PauseUntil != "" && in.PauseFrom == "" {
		return fmt.Errorf("%w: a pause needs a start", ErrInvalidItem)
	}
	if in.PauseUntil != "" && in.PauseUntil <= in.PauseFrom {
		return fmt.Errorf("%w: a pause must end after it starts", ErrInvalidItem)
	}
	return nil
}

// Bills lists the ledger by name; the page orders it by what comes due first.
func (v *Vault) Bills() ([]Bill, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return nil, ErrLocked
	}
	out := slices.Clone(v.data.Bills)
	if out == nil {
		out = []Bill{}
	}
	sort.SliceStable(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}

func (v *Vault) AddBill(in BillInput) (Bill, error) {
	if err := in.normalize(); err != nil {
		return Bill{}, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Bill{}, ErrLocked
	}
	now := time.Now().UTC()
	b := Bill{ID: hex.EncodeToString(randomBytes(8)), CreatedAt: now}
	b.set(in, now)
	next := v.data
	next.Bills = append(slices.Clone(v.data.Bills), b)
	if err := v.save(next); err != nil {
		return Bill{}, err
	}
	v.data = next
	return b, nil
}

func (v *Vault) UpdateBill(id string, in BillInput) (Bill, error) {
	if err := in.normalize(); err != nil {
		return Bill{}, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Bill{}, ErrLocked
	}
	i := v.billIndex(id)
	if i < 0 {
		return Bill{}, ErrNotFound
	}
	next := v.data
	next.Bills = slices.Clone(v.data.Bills)
	now := time.Now()
	next.Bills[i].reprice(in, now.Format(time.DateOnly))
	next.Bills[i].set(in, now.UTC())
	if err := v.save(next); err != nil {
		return Bill{}, err
	}
	v.data = next
	return next.Bills[i], nil
}

// DeleteBill strikes a bill from the ledger and returns it.
func (v *Vault) DeleteBill(id string) (Bill, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Bill{}, ErrLocked
	}
	i := v.billIndex(id)
	if i < 0 {
		return Bill{}, ErrNotFound
	}
	b := v.data.Bills[i]
	next := v.data
	next.Bills = slices.Delete(slices.Clone(v.data.Bills), i, i+1)
	if err := v.save(next); err != nil {
		return Bill{}, err
	}
	v.data = next
	return b, nil
}

// reprice keeps the old price in the history when an edit changes it, as of
// today (the user's local date). A second change on the same day keeps the
// price the day started with, a change before the first charge keeps nothing,
// and a new currency starts the history over.
func (b *Bill) reprice(in BillInput, today string) {
	switch {
	case in.Currency != b.Currency:
		b.History = nil
	case in.Amount == b.Amount, today <= in.Since:
	case len(b.History) > 0 && b.History[len(b.History)-1].Until == today:
	default:
		b.History = append(slices.Clone(b.History), PriceChange{Amount: b.Amount, Until: today})
	}
	// the earliest prices fall away once there are many
	if n := len(b.History); n > maxPriceChanges {
		b.History = slices.Clone(b.History[n-maxPriceChanges:])
	}
}

// maxPriceChanges caps a bill's price history.
const maxPriceChanges = 50

func (b *Bill) set(in BillInput, now time.Time) {
	b.Name = in.Name
	b.Amount = in.Amount
	b.Currency = in.Currency
	b.Cycle = in.Cycle
	b.Since = in.Since
	b.Until = in.Until
	b.PauseFrom = in.PauseFrom
	b.PauseUntil = in.PauseUntil
	b.Method = in.Method
	b.Note = in.Note
	b.UpdatedAt = now
}

func (v *Vault) billIndex(id string) int {
	return slices.IndexFunc(v.data.Bills, func(b Bill) bool { return b.ID == id })
}

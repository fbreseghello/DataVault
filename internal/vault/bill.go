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
type Bill struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Amount    int64     `json:"amount"`
	Currency  string    `json:"currency"`
	Cycle     Cycle     `json:"cycle"`
	Since     string    `json:"since"`
	Method    string    `json:"method"`
	Note      string    `json:"note"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// BillInput is what the ledger sends when writing a bill.
type BillInput struct {
	Name     string `json:"name"`
	Amount   int64  `json:"amount"`
	Currency string `json:"currency"`
	Cycle    Cycle  `json:"cycle"`
	Since    string `json:"since"`
	Method   string `json:"method"`
	Note     string `json:"note"`
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
	next.Bills[i].set(in, time.Now().UTC())
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

func (b *Bill) set(in BillInput, now time.Time) {
	b.Name = in.Name
	b.Amount = in.Amount
	b.Currency = in.Currency
	b.Cycle = in.Cycle
	b.Since = in.Since
	b.Method = in.Method
	b.Note = in.Note
	b.UpdatedAt = now
}

func (v *Vault) billIndex(id string) int {
	return slices.IndexFunc(v.data.Bills, func(b Bill) bool { return b.ID == id })
}

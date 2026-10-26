package vault

import (
	"time"
)

// MaxRenewDays caps the renewal interval at about ten years.
const MaxRenewDays = 3650

// RenewDue is the date the item is due for renewal, as YYYY-MM-DD, or "" when
// it has none. A fixed date wins; otherwise the interval counts from the last
// renewal, or from when the item was created if it was never renewed.
func (it Item) RenewDue() string {
	if it.RenewBy != "" {
		return it.RenewBy
	}
	if it.RenewEvery <= 0 {
		return ""
	}
	from := it.CreatedAt
	if it.RenewedAt != nil {
		from = *it.RenewedAt
	}
	return from.UTC().AddDate(0, 0, it.RenewEvery).Format(time.DateOnly)
}

// Renewed records that the item was renewed now: the interval starts again
// from today, and a fixed date, now spent, is cleared.
func (v *Vault) Renewed(id string) (Item, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Item{}, ErrLocked
	}
	i := v.index(id)
	if i < 0 {
		return Item{}, ErrNotFound
	}
	now := time.Now().UTC()
	it := v.data.Items[i]
	it.RenewedAt = &now
	it.RenewBy = ""
	it.UpdatedAt = now

	next := v.data
	next.Items = append([]Item(nil), v.data.Items...)
	next.Items[i] = it
	if err := v.save(next); err != nil {
		return Item{}, err
	}
	v.data = next
	return it, nil
}

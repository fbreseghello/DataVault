package vault

import (
	"errors"
	"testing"
	"time"
)

func TestRenewDue(t *testing.T) {
	created := time.Date(2026, 1, 10, 15, 0, 0, 0, time.UTC)
	renewed := time.Date(2026, 3, 1, 9, 0, 0, 0, time.UTC)
	cases := []struct {
		name string
		it   Item
		want string
	}{
		{"none", Item{CreatedAt: created}, ""},
		{"every, never renewed", Item{RenewEvery: 30, CreatedAt: created}, "2026-02-09"},
		{"every, renewed", Item{RenewEvery: 30, RenewedAt: &renewed, CreatedAt: created}, "2026-03-31"},
		{"fixed date wins", Item{RenewEvery: 30, RenewBy: "2026-12-01", CreatedAt: created}, "2026-12-01"},
	}
	for _, c := range cases {
		if got := c.it.RenewDue(); got != c.want {
			t.Errorf("%s: due %q, want %q", c.name, got, c.want)
		}
	}
}

func TestRenewed(t *testing.T) {
	v, path := newVault(t)
	it, err := v.Add(Item{Type: TypeAPIKey, Name: "stripe", RenewEvery: 90, RenewBy: "2026-01-01"})
	if err != nil {
		t.Fatal(err)
	}
	if it.RenewDue() != "2026-01-01" {
		t.Fatalf("the fixed date should win, got %q", it.RenewDue())
	}
	got, err := v.Renewed(it.ID)
	if err != nil {
		t.Fatal(err)
	}
	want := time.Now().UTC().AddDate(0, 0, 90).Format(time.DateOnly)
	if got.RenewBy != "" || got.RenewedAt == nil || got.RenewDue() != want {
		t.Fatalf("after renewing: by %q, at %v, due %q (want %q)", got.RenewBy, got.RenewedAt, got.RenewDue(), want)
	}
	// editing the item keeps the last renewal, whatever the client sends
	edit := got
	edit.RenewedAt = nil
	edit.Name = "stripe live"
	if upd, err := v.Update(it.ID, edit); err != nil || upd.RenewedAt == nil {
		t.Fatalf("update dropped the renewal: %v %v", upd.RenewedAt, err)
	}
	v.Lock()
	v2, err := Open(path, []byte("keyword"), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	list, _ := v2.List("")
	if len(list) != 1 || list[0].RenewDue != want {
		t.Fatalf("summary due %+v, want %q", list, want)
	}
}

func TestRenewValidation(t *testing.T) {
	v, _ := newVault(t)
	for _, it := range []Item{
		{Type: TypeAPIKey, Name: "a", RenewEvery: -1},
		{Type: TypeAPIKey, Name: "b", RenewEvery: MaxRenewDays + 1},
		{Type: TypeAPIKey, Name: "c", RenewBy: "next week"},
	} {
		if _, err := v.Add(it); !errors.Is(err, ErrInvalidItem) {
			t.Errorf("%s accepted: %v", it.Name, err)
		}
	}
}

package vault

import (
	"errors"
	"testing"
	"time"
)

func TestBills(t *testing.T) {
	v, path := newVault(t)
	in := BillInput{Name: " stream ", Amount: 5590, Currency: "brl", Since: "2026-01-31", Method: "visa"}
	a, err := v.AddBill(in)
	if err != nil {
		t.Fatal(err)
	}
	if a.Name != "stream" || a.Currency != "BRL" || a.Cycle != Monthly {
		t.Fatalf("not normalized: %+v", a)
	}
	b, _ := v.AddBill(BillInput{Name: "Domain", Amount: 6000, Currency: "USD", Cycle: Yearly, Since: "2025-03-02"})

	in.Amount = 6490
	in.Until = "2027-01-31"
	in.PauseFrom = "2026-06-01"
	if _, err := v.UpdateBill(a.ID, in); err != nil {
		t.Fatal(err)
	}
	if _, err := v.DeleteBill(b.ID); err != nil {
		t.Fatal(err)
	}

	v.Lock()
	v, err = Open(path, []byte("keyword"), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	bills, err := v.Bills()
	if err != nil {
		t.Fatal(err)
	}
	if len(bills[0].History) != 1 || bills[0].History[0].Amount != 5590 {
		t.Fatalf("price history: %+v", bills[0].History)
	}
	if len(bills) != 1 || bills[0].Amount != 6490 || bills[0].Method != "visa" || bills[0].Until != "2027-01-31" || bills[0].PauseFrom != "2026-06-01" {
		t.Fatalf("after reopen: %+v", bills)
	}
	if _, err := v.DeleteBill(b.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("delete twice: %v", err)
	}
}

func TestBillValidation(t *testing.T) {
	v, _ := newVault(t)
	for _, in := range []BillInput{
		{Amount: 1, Currency: "BRL", Since: "2026-01-01"},
		{Name: "x", Amount: -1, Currency: "BRL", Since: "2026-01-01"},
		{Name: "x", Currency: "R$", Since: "2026-01-01"},
		{Name: "x", Currency: "BRL", Cycle: "daily", Since: "2026-01-01"},
		{Name: "x", Currency: "BRL", Since: "01/02/2026"},
		{Name: "x", Currency: "BRL", Since: "2026-01-01", Until: "2026-13-01"},
		{Name: "x", Currency: "BRL", Since: "2026-01-01", Until: "2026-01-01"},
		{Name: "x", Currency: "BRL", Since: "2026-01-01", PauseUntil: "2026-03-01"},
		{Name: "x", Currency: "BRL", Since: "2026-01-01", PauseFrom: "2026-03-01", PauseUntil: "2026-03-01"},
		{Name: "x", Currency: "BRL", Since: "2026-01-01", PauseFrom: "march"},
	} {
		if _, err := v.AddBill(in); !errors.Is(err, ErrInvalidItem) {
			t.Fatalf("%+v: got %v", in, err)
		}
	}
}

func TestBillReprice(t *testing.T) {
	b := Bill{Amount: 4990, Currency: "BRL", Since: "2025-01-10"}
	in := BillInput{Amount: 5590, Currency: "BRL", Since: "2025-01-10"}
	b.reprice(in, "2026-03-01")
	b.set(in, time.Time{})
	if len(b.History) != 1 || b.History[0] != (PriceChange{Amount: 4990, Until: "2026-03-01"}) {
		t.Fatalf("first change: %+v", b.History)
	}

	// a second change the same day keeps the price the day started with
	in.Amount = 5990
	b.reprice(in, "2026-03-01")
	b.set(in, time.Time{})
	if len(b.History) != 1 || b.History[0].Amount != 4990 {
		t.Fatalf("same day: %+v", b.History)
	}

	// the same amount keeps the history as it is
	b.reprice(in, "2026-04-01")
	if len(b.History) != 1 {
		t.Fatalf("same amount: %+v", b.History)
	}

	// a change before the first charge is not a price the bill ever had
	c := Bill{Amount: 100, Currency: "BRL", Since: "2026-12-01"}
	c.reprice(BillInput{Amount: 200, Currency: "BRL", Since: "2026-12-01"}, "2026-11-16")
	if len(c.History) != 0 {
		t.Fatalf("before the first charge: %+v", c.History)
	}

	// a new currency starts over
	in.Currency = "USD"
	b.reprice(in, "2026-05-01")
	if b.History != nil {
		t.Fatalf("new currency: %+v", b.History)
	}
}

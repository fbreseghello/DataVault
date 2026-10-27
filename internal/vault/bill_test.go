package vault

import (
	"errors"
	"testing"
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
	if len(bills) != 1 || bills[0].Amount != 6490 || bills[0].Method != "visa" {
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
	} {
		if _, err := v.AddBill(in); !errors.Is(err, ErrInvalidItem) {
			t.Fatalf("%+v: got %v", in, err)
		}
	}
}

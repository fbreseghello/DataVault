package vault

import (
	"errors"
	"testing"
)

func titles(t *testing.T, v *Vault, day string, c Column) []string {
	t.Helper()
	tasks, err := v.Tasks(day)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, x := range tasks {
		if x.Column == c {
			out = append(out, x.Title)
		}
	}
	return out
}

func same(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTasksAddMoveDelete(t *testing.T) {
	v, path := newVault(t)
	day := "2026-10-22"
	a, _ := v.AddTask(TaskInput{Title: "a", Day: day})
	b, _ := v.AddTask(TaskInput{Title: "b", Day: day})
	v.AddTask(TaskInput{Title: "c", Day: day})
	if got := titles(t, v, day, Orders); !same(got, []string{"a", "b", "c"}) {
		t.Fatalf("orders: %v", got)
	}

	// Reorder inside the column, then send one into battle.
	v.MoveTask(a.ID, Orders, 2)
	if got := titles(t, v, day, Orders); !same(got, []string{"b", "c", "a"}) {
		t.Fatalf("after reorder: %v", got)
	}
	v.MoveTask(b.ID, InBattle, 0)
	if got := titles(t, v, day, Orders); !same(got, []string{"c", "a"}) {
		t.Fatalf("orders after move: %v", got)
	}
	if got := titles(t, v, day, InBattle); !same(got, []string{"b"}) {
		t.Fatalf("in battle: %v", got)
	}
	v.MoveTask(b.ID, Conquered, 99)
	v.DeleteTask(a.ID)
	v.Lock()

	v2, _ := Open(path, []byte("keyword"), []byte("password"))
	if got := titles(t, v2, day, Orders); !same(got, []string{"c"}) {
		t.Fatalf("orders after reopen: %v", got)
	}
	if got := titles(t, v2, day, Conquered); !same(got, []string{"b"}) {
		t.Fatalf("conquered after reopen: %v", got)
	}
}

func TestTaskValidation(t *testing.T) {
	v, _ := newVault(t)
	for _, in := range []TaskInput{
		{Title: " ", Day: "2026-10-22"},
		{Title: "x", Day: "22/10/2026"},
		{Title: "x", Day: "2026-10-22", Column: "retreat"},
	} {
		if _, err := v.AddTask(in); !errors.Is(err, ErrInvalidItem) {
			t.Errorf("AddTask(%+v) = %v", in, err)
		}
	}
}

func TestCarryPending(t *testing.T) {
	v, _ := newVault(t)
	v.AddTask(TaskInput{Title: "old order", Day: "2026-10-20"})
	fight, _ := v.AddTask(TaskInput{Title: "still fighting", Day: "2026-10-21"})
	v.MoveTask(fight.ID, InBattle, 0)
	won, _ := v.AddTask(TaskInput{Title: "won", Day: "2026-10-21"})
	v.MoveTask(won.ID, Conquered, 0)
	v.AddTask(TaskInput{Title: "today", Day: "2026-10-22"})

	n, last, _ := v.PendingBefore("2026-10-22")
	if n != 2 || last != "2026-10-21" {
		t.Fatalf("pending: %d, last %s", n, last)
	}
	moved, err := v.CarryPending("2026-10-22")
	if err != nil || moved != 2 {
		t.Fatalf("carry: %d %v", moved, err)
	}
	if got := titles(t, v, "2026-10-22", Orders); !same(got, []string{"today", "old order"}) {
		t.Fatalf("orders today: %v", got)
	}
	if got := titles(t, v, "2026-10-22", InBattle); !same(got, []string{"still fighting"}) {
		t.Fatalf("in battle today: %v", got)
	}
	// Finished work stays on its day.
	if got := titles(t, v, "2026-10-21", Conquered); !same(got, []string{"won"}) {
		t.Fatalf("history: %v", got)
	}
	if n, _, _ := v.PendingBefore("2026-10-22"); n != 0 {
		t.Fatalf("pending after carry: %d", n)
	}
}

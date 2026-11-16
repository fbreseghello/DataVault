package vault

import (
	"testing"
	"time"
)

// at sets the hollow's clock to noon of day for the rest of the test.
func at(t *testing.T, day string) {
	t.Helper()
	d, err := time.ParseInLocation(time.DateOnly, day, time.Local)
	if err != nil {
		t.Fatal(err)
	}
	old := clock
	clock = func() time.Time { return d.Add(12 * time.Hour) }
	t.Cleanup(func() { clock = old })
}

func xp(t *testing.T, v *Vault) HollowState {
	t.Helper()
	s, err := v.Hollow()
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestLevels(t *testing.T) {
	for n, want := range map[int]int{1: 100, 2: 260, 3: 470, 4: 700, 5: 950, 10: 2510} {
		if got := LevelNeed(n); got != want {
			t.Errorf("LevelNeed(%d) = %d, want %d", n, got, want)
		}
	}
	for _, c := range []struct{ xp, level, into, need int }{
		{0, 1, 0, 100},
		{99, 1, 99, 100},
		{100, 2, 0, 260},
		{1530, 5, 0, 950},
		{9240, 10, 0, 2510},
	} {
		level, into, need := LevelFor(c.xp)
		if level != c.level || into != c.into || need != c.need {
			t.Errorf("LevelFor(%d) = %d %d %d, want %d %d %d", c.xp, level, into, need, c.level, c.into, c.need)
		}
	}
}

func TestConquerAndUndo(t *testing.T) {
	at(t, "2026-11-16")
	v, _ := newVault(t)
	a, _ := v.AddTask(TaskInput{Title: "a", Day: "2026-11-16"})
	b, _ := v.AddTask(TaskInput{Title: "b", Day: "2026-11-16"})

	if _, err := v.MoveTask(a.ID, Conquered, 0); err != nil {
		t.Fatal(err)
	}
	if s := xp(t, v); s.XP != 10 || s.Attrs[Vigor] != 10 {
		t.Fatalf("one conquered: %+v", s)
	}

	// back and forth earns it once
	_, _ = v.MoveTask(a.ID, Orders, 0)
	if s := xp(t, v); s.XP != 0 {
		t.Fatalf("moved out: %d", s.XP)
	}
	_, _ = v.MoveTask(a.ID, Conquered, 0)
	_, _ = v.MoveTask(a.ID, Conquered, 0)
	if s := xp(t, v); s.XP != 10 {
		t.Fatalf("back again: %d", s.XP)
	}

	// the last task conquered clears the day
	_, _ = v.MoveTask(b.ID, Conquered, 1)
	if s := xp(t, v); s.XP != 10+10+25 {
		t.Fatalf("day cleared: %d", s.XP)
	}

	// a new order on the day takes the cleared day back
	c, _ := v.AddTask(TaskInput{Title: "c", Day: "2026-11-16"})
	if s := xp(t, v); s.XP != 20 {
		t.Fatalf("new order: %d", s.XP)
	}

	// fallen neither gives nor takes, and clears the day again only on a conquest
	_, _ = v.MoveTask(c.ID, Fallen, 0)
	if s := xp(t, v); s.XP != 20 {
		t.Fatalf("fallen: %d", s.XP)
	}

	// deleting a conquered task takes its deed
	if err := v.DeleteTask(a.ID); err != nil {
		t.Fatal(err)
	}
	if s := xp(t, v); s.XP != 10 || len(s.Deeds) != 1 {
		t.Fatalf("deleted: %+v", s)
	}
}

func TestDailyLimit(t *testing.T) {
	at(t, "2026-11-16")
	v, _ := newVault(t)
	for i := range tasksPerDay + 3 {
		task, _ := v.AddTask(TaskInput{Title: "t", Day: "2026-11-16", Column: Fallen})
		if _, err := v.MoveTask(task.ID, Conquered, i); err != nil {
			t.Fatal(err)
		}
	}
	// 20 tasks and the cleared day
	if s := xp(t, v); s.XP != tasksPerDay*taskXP+dayClearedXP {
		t.Fatalf("limit: %d", s.XP)
	}
	at(t, "2026-11-17")
	task, _ := v.AddTask(TaskInput{Title: "t", Day: "2026-11-17"})
	_, _ = v.MoveTask(task.ID, Conquered, 0)
	if s := xp(t, v); s.XP != tasksPerDay*taskXP+2*dayClearedXP+taskXP {
		t.Fatalf("next day: %d", s.XP)
	}
}

func TestCarryNeverClears(t *testing.T) {
	at(t, "2026-11-16")
	v, path := newVault(t)
	a, _ := v.AddTask(TaskInput{Title: "a", Day: "2026-11-15"})
	_, _ = v.AddTask(TaskInput{Title: "b", Day: "2026-11-15"})
	_, _ = v.MoveTask(a.ID, Conquered, 0)
	if _, err := v.CarryPending("2026-11-16"); err != nil {
		t.Fatal(err)
	}
	// yesterday has nothing pending now, but carrying is not clearing
	if s := xp(t, v); s.XP != 10 {
		t.Fatalf("after carry: %d", s.XP)
	}

	// a level seen is kept with the vault
	if _, err := v.SeeLevel(); err != nil {
		t.Fatal(err)
	}
	v.Lock()
	v, err := Open(path, []byte("keyword"), []byte("password"))
	if err != nil {
		t.Fatal(err)
	}
	if s := xp(t, v); s.XP != 10 || s.SeenLevel != 1 || s.Deeds[0].Kind != TaskConquered {
		t.Fatalf("after reopen: %+v", s)
	}
}

func TestClass(t *testing.T) {
	at(t, "2026-11-16")
	v, _ := newVault(t)
	if s := xp(t, v); s.Class != "" {
		t.Fatalf("class before choosing: %q", s.Class)
	}
	if _, err := v.SetClass("bard"); err == nil {
		t.Fatal("unknown class accepted")
	}
	for range 3 {
		task, _ := v.AddTask(TaskInput{Title: "t", Day: "2026-11-16", Column: Fallen})
		_, _ = v.MoveTask(task.ID, Conquered, 0)
	}
	// 3 tasks and the cleared day: 55 vigor, and a knight leans to vigor
	s, err := v.SetClass("knight")
	if err != nil {
		t.Fatal(err)
	}
	if s.Class != "knight" || s.XP != 55 || s.Attrs[Vigor] != 60 {
		t.Fatalf("knight: %+v", s)
	}
	// changing class is free and moves the lean
	if s, _ := v.SetClass("scribe"); s.XP != 55 || s.Attrs[Vigor] != 55 {
		t.Fatalf("scribe: %+v", s)
	}
}

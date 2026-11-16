package vault

import (
	"fmt"
	"math"
	"slices"
	"time"
)

// The hollow: a character that grows with the vault (docs/PERSONAGEM.md). It
// is kept as a list of deeds, each tied to the thing that earned it, so the
// experience, the level and the attributes are always worked out from them and
// undoing a deed is only taking it off the list.

// The attributes, one per page; a deed adds its XP to the one of its page.
const (
	Vigor  = "vigor"  // siege
	Memory = "memory" // unspoken
	Ward   = "ward"   // grimoire
	Hoard  = "hoard"  // crypt
	Thrift = "thrift" // ledger
)

// Attributes in the order the page lists them.
var Attributes = []string{Vigor, Memory, Ward, Hoard, Thrift}

// Companions are what the hollow can be, each in its category
// (docs/PERSONAGEM.md, "O companheiro"). Any theme can choose any of them.
var Companions = map[string]string{
	"hermit": "wanderers", "mercenary": "wanderers", "sorcerer": "wanderers", "witch": "wanderers", "necromancer": "wanderers",
	"dog": "animals", "cat": "animals", "raven": "animals", "bat": "animals", "owl": "animals", "wolf": "animals", "serpent": "animals",
	"reaper": "shinigami", "gaunt": "shinigami", "masked": "shinigami",
	"imp": "demons", "brute": "demons", "fiend": "demons", "hellhound": "demons",
	"samurai": "bushi", "ronin": "bushi", "shinobi": "bushi", "komuso": "bushi", "sohei": "bushi",
	"pixie": "fae", "wisp": "fae", "dryad": "fae", "moth": "fae", "changeling": "fae",
	"goblin": "denizens", "skeleton": "denizens", "slime": "denizens", "mimic": "denizens", "golem": "denizens",
}

// DeedKind is what was done to earn a deed. The page names it; the deed never
// keeps the title of what it came from.
type DeedKind string

const (
	// TaskConquered is a siege task moved into Conquered.
	TaskConquered DeedKind = "task_conquered"
	// DayCleared is a day's board left with nothing in Orders or In Battle.
	DayCleared DeedKind = "day_cleared"
)

// What each deed is worth, and how many of a kind count in one day.
const (
	taskXP       = 10
	tasksPerDay  = 20
	dayClearedXP = 25
)

// Deed is one gain of experience. Source is the id of what earned it (a task,
// or "day:YYYY-MM-DD" for a cleared day); no source earns twice at once. Day is
// the user's local date it was earned, for the history and the daily limits.
type Deed struct {
	XP     int       `json:"xp"`
	Attr   string    `json:"attr"`
	Kind   DeedKind  `json:"kind"`
	Source string    `json:"source"`
	Day    string    `json:"day"`
	At     time.Time `json:"at"`
}

// Hollow is the character's part of the vault.
type Hollow struct {
	// Companion is what the user chose; empty until the page first asks. (The
	// pixel art classes of 0.22 were kept as "class", which is no longer read.)
	Companion string `json:"companion,omitempty"`
	Deeds     []Deed `json:"deeds,omitempty"`
	// SeenLevel is the last level the user has seen; a higher one is news.
	SeenLevel int `json:"seen_level,omitempty"`
}

// HollowState is the character as the page draws it.
type HollowState struct {
	Companion string `json:"companion"`
	XP        int    `json:"xp"`
	Level     int    `json:"level"`
	// Into is the XP gained inside the current level; Need is what the level takes.
	Into      int            `json:"into"`
	Need      int            `json:"need"`
	SeenLevel int            `json:"seen_level"`
	Attrs     map[string]int `json:"attrs"`
	// Deeds are the latest ones, newest first.
	Deeds []Deed `json:"deeds"`
}

// maxDeedsShown caps the history the page gets; the totals count them all.
const maxDeedsShown = 300

// clock is the time the hollow reads, so tests can set the day.
var clock = time.Now

// LevelNeed is the XP it takes to go from level n to n+1: 100 × n^1.4, to the
// nearest ten.
func LevelNeed(n int) int {
	return int(math.Round(100*math.Pow(float64(n), 1.4)/10)) * 10
}

// LevelFor works out the level that xp reaches, the XP gained inside it and
// what the level takes in all.
func LevelFor(xp int) (level, into, need int) {
	level, into = 1, xp
	for into >= LevelNeed(level) {
		into -= LevelNeed(level)
		level++
	}
	return level, into, LevelNeed(level)
}

func (h Hollow) state() HollowState {
	s := HollowState{Companion: h.Companion, SeenLevel: h.SeenLevel, Attrs: map[string]int{}}
	for _, a := range Attributes {
		s.Attrs[a] = 0
	}
	for _, d := range h.Deeds {
		s.XP += d.XP
		s.Attrs[d.Attr] += d.XP
	}
	s.Level, s.Into, s.Need = LevelFor(s.XP)
	n := min(len(h.Deeds), maxDeedsShown)
	s.Deeds = make([]Deed, 0, n)
	for i := len(h.Deeds) - 1; i >= len(h.Deeds)-n; i-- {
		s.Deeds = append(s.Deeds, h.Deeds[i])
	}
	return s
}

// Hollow returns the character as it stands.
func (v *Vault) Hollow() (HollowState, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return HollowState{}, ErrLocked
	}
	return v.data.Hollow.state(), nil
}

// SetCompanion gives the hollow its companion. Changing it later is free: it
// costs no XP and no level.
func (v *Vault) SetCompanion(companion string) (HollowState, error) {
	if _, ok := Companions[companion]; !ok {
		return HollowState{}, fmt.Errorf("%w: unknown companion %q", ErrInvalidItem, companion)
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return HollowState{}, ErrLocked
	}
	next := v.data
	next.Hollow.Companion = companion
	if err := v.save(next); err != nil {
		return HollowState{}, err
	}
	v.data = next
	return next.Hollow.state(), nil
}

// SeeLevel marks the current level as seen, once the page shows it.
func (v *Vault) SeeLevel() (HollowState, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return HollowState{}, ErrLocked
	}
	s := v.data.Hollow.state()
	if s.Level == v.data.Hollow.SeenLevel {
		return s, nil
	}
	next := v.data
	next.Hollow.SeenLevel = s.Level
	if err := v.save(next); err != nil {
		return HollowState{}, err
	}
	v.data = next
	s.SeenLevel = s.Level
	return s, nil
}

// siegeDeeds brings the deeds in line with the siege after a change to the
// boards of days. A task that is no longer in Conquered (moved out or deleted)
// loses its deed; a task in conquered just now earns one, up to the day's
// limit. A day loses its deed once it has work pending again or nothing
// conquered, and earns one only when a task is conquered on it, so carrying
// unfinished work away never clears a day by itself. The deeds are copied, never
// changed in place.
func siegeDeeds(deeds []Deed, tasks []Task, days []string, conquered string) []Deed {
	now := clock()
	today := now.Format(time.DateOnly)
	byID := make(map[string]Task, len(tasks))
	for _, t := range tasks {
		byID[t.ID] = t
	}

	out := slices.DeleteFunc(slices.Clone(deeds), func(d Deed) bool {
		if d.Kind == TaskConquered {
			t, ok := byID[d.Source]
			return !ok || t.Column != Conquered
		}
		if d.Kind == DayCleared && slices.Contains(days, d.Source[len("day:"):]) {
			return !cleared(tasks, d.Source[len("day:"):])
		}
		return false
	})

	if conquered == "" {
		return out
	}
	t, ok := byID[conquered]
	if !ok || t.Column != Conquered {
		return out
	}
	has := func(source string) bool {
		return slices.ContainsFunc(out, func(d Deed) bool { return d.Source == source })
	}
	if !has(t.ID) {
		n := 0
		for _, d := range out {
			if d.Kind == TaskConquered && d.Day == today {
				n++
			}
		}
		if n < tasksPerDay {
			out = append(out, Deed{XP: taskXP, Attr: Vigor, Kind: TaskConquered, Source: t.ID, Day: today, At: now.UTC()})
		}
	}
	// one cleared day a day: clearing old boards one after another earns once
	clearedToday := slices.ContainsFunc(out, func(d Deed) bool { return d.Kind == DayCleared && d.Day == today })
	if day := "day:" + t.Day; t.Day <= today && !clearedToday && !has(day) && cleared(tasks, t.Day) {
		out = append(out, Deed{XP: dayClearedXP, Attr: Vigor, Kind: DayCleared, Source: day, Day: today, At: now.UTC()})
	}
	return out
}

// cleared reports whether a day's board has something conquered and nothing
// left in Orders or In Battle.
func cleared(tasks []Task, day string) bool {
	won := false
	for _, t := range tasks {
		if t.Day != day {
			continue
		}
		if t.Column.pending() {
			return false
		}
		won = won || t.Column == Conquered
	}
	return won
}

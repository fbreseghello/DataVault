package vault

import (
	"encoding/hex"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"
)

// Column is where a Siege task stands on its day's board.
type Column string

const (
	Orders    Column = "orders"
	InBattle  Column = "in_battle"
	Conquered Column = "conquered"
	Fallen    Column = "fallen"
)

func (c Column) valid() bool {
	switch c {
	case Orders, InBattle, Conquered, Fallen:
		return true
	}
	return false
}

// pending reports whether the task still has to be fought.
func (c Column) pending() bool { return c == Orders || c == InBattle }

// Task is one card on a Siege board. Day is the user's local date (YYYY-MM-DD).
type Task struct {
	ID        string    `json:"id"`
	Title     string    `json:"title"`
	Note      string    `json:"note"`
	Column    Column    `json:"column"`
	Day       string    `json:"day"`
	Position  int       `json:"position"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// TaskInput is what the board sends when creating or editing a task.
type TaskInput struct {
	Title  string `json:"title"`
	Note   string `json:"note"`
	Column Column `json:"column"`
	Day    string `json:"day"`
}

func validDay(day string) error {
	if _, err := time.Parse("2006-01-02", day); err != nil {
		return fmt.Errorf("%w: day must be YYYY-MM-DD", ErrInvalidItem)
	}
	return nil
}

func (in *TaskInput) normalize() error {
	in.Title = strings.TrimSpace(in.Title)
	if in.Title == "" {
		return fmt.Errorf("%w: title is required", ErrInvalidItem)
	}
	if in.Column == "" {
		in.Column = Orders
	}
	if !in.Column.valid() {
		return fmt.Errorf("%w: unknown column %q", ErrInvalidItem, in.Column)
	}
	return validDay(in.Day)
}

// Tasks returns a day's board, column by column, in position order.
func (v *Vault) Tasks(day string) ([]Task, error) {
	if err := validDay(day); err != nil {
		return nil, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return nil, ErrLocked
	}
	out := []Task{}
	for _, t := range v.data.Tasks {
		if t.Day == day {
			out = append(out, t)
		}
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Position < out[j].Position })
	return out, nil
}

// AddTask puts a new task at the bottom of its column.
func (v *Vault) AddTask(in TaskInput) (Task, error) {
	if err := in.normalize(); err != nil {
		return Task{}, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Task{}, ErrLocked
	}
	now := time.Now().UTC()
	t := Task{
		ID:        hex.EncodeToString(randomBytes(8)),
		Title:     in.Title,
		Note:      in.Note,
		Column:    in.Column,
		Day:       in.Day,
		Position:  v.columnLen(in.Day, in.Column),
		CreatedAt: now,
		UpdatedAt: now,
	}
	next := v.data
	next.Tasks = append(slices.Clone(v.data.Tasks), t)
	if err := v.save(next); err != nil {
		return Task{}, err
	}
	v.data = next
	return t, nil
}

// UpdateTask edits the title and note; moving goes through MoveTask.
func (v *Vault) UpdateTask(id, title, note string) (Task, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return Task{}, fmt.Errorf("%w: title is required", ErrInvalidItem)
	}
	return v.changeTasks(id, func(tasks []Task, i int) []Task {
		tasks[i].Title = title
		tasks[i].Note = note
		tasks[i].UpdatedAt = time.Now().UTC()
		return tasks
	})
}

// MoveTask places a task in column at index (0 is the top), shifting the rest.
func (v *Vault) MoveTask(id string, column Column, index int) (Task, error) {
	if !column.valid() {
		return Task{}, fmt.Errorf("%w: unknown column %q", ErrInvalidItem, column)
	}
	return v.changeTasks(id, func(tasks []Task, i int) []Task {
		moving := tasks[i]
		// The column as it will be, without the moving task, in order.
		var col []int
		for j, t := range tasks {
			if j != i && t.Day == moving.Day && t.Column == column {
				col = append(col, j)
			}
		}
		sort.SliceStable(col, func(a, b int) bool { return tasks[col[a]].Position < tasks[col[b]].Position })
		index = max(0, min(index, len(col)))
		col = slices.Insert(col, index, i)

		from := moving.Column
		tasks[i].Column = column
		tasks[i].UpdatedAt = time.Now().UTC()
		for pos, j := range col {
			tasks[j].Position = pos
		}
		if from != column {
			renumber(tasks, moving.Day, from)
		}
		return tasks
	})
}

func (v *Vault) DeleteTask(id string) error {
	_, err := v.changeTasks(id, func(tasks []Task, i int) []Task {
		gone := tasks[i]
		tasks = slices.Delete(tasks, i, i+1)
		renumber(tasks, gone.Day, gone.Column)
		return tasks
	})
	return err
}

// PendingBefore reports how many unfinished tasks (orders and in battle) are
// left on days before day, and the latest such day.
func (v *Vault) PendingBefore(day string) (count int, last string, err error) {
	if err := validDay(day); err != nil {
		return 0, "", err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return 0, "", ErrLocked
	}
	for _, t := range v.data.Tasks {
		if t.Day < day && t.Column.pending() {
			count++
			if t.Day > last {
				last = t.Day
			}
		}
	}
	return count, last, nil
}

// CarryPending moves every unfinished task from earlier days onto day, keeping
// its column and appending below what is already there.
func (v *Vault) CarryPending(day string) (int, error) {
	if err := validDay(day); err != nil {
		return 0, err
	}
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return 0, ErrLocked
	}
	next := v.data
	next.Tasks = slices.Clone(v.data.Tasks)
	var moved []int
	for i, t := range next.Tasks {
		if t.Day < day && t.Column.pending() {
			moved = append(moved, i)
		}
	}
	if len(moved) == 0 {
		return 0, nil
	}
	// Older days first, then their order on the board.
	sort.SliceStable(moved, func(a, b int) bool {
		ta, tb := next.Tasks[moved[a]], next.Tasks[moved[b]]
		if ta.Day != tb.Day {
			return ta.Day < tb.Day
		}
		return ta.Position < tb.Position
	})
	now := time.Now().UTC()
	from := map[[2]string]bool{}
	for _, i := range moved {
		t := &next.Tasks[i]
		from[[2]string{t.Day, string(t.Column)}] = true
		t.Position = columnLen(next.Tasks, day, t.Column)
		t.Day = day
		t.UpdatedAt = now
	}
	for k := range from {
		renumber(next.Tasks, k[0], Column(k[1]))
	}
	if err := v.save(next); err != nil {
		return 0, err
	}
	v.data = next
	return len(moved), nil
}

func (v *Vault) changeTasks(id string, change func(tasks []Task, i int) []Task) (Task, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.dek == nil {
		return Task{}, ErrLocked
	}
	i := slices.IndexFunc(v.data.Tasks, func(t Task) bool { return t.ID == id })
	if i < 0 {
		return Task{}, ErrNotFound
	}
	next := v.data
	next.Tasks = change(slices.Clone(v.data.Tasks), i)
	if err := v.save(next); err != nil {
		return Task{}, err
	}
	v.data = next
	if j := slices.IndexFunc(next.Tasks, func(t Task) bool { return t.ID == id }); j >= 0 {
		return next.Tasks[j], nil
	}
	return Task{}, nil
}

func (v *Vault) columnLen(day string, c Column) int { return columnLen(v.data.Tasks, day, c) }

func columnLen(tasks []Task, day string, c Column) int {
	n := 0
	for _, t := range tasks {
		if t.Day == day && t.Column == c {
			n++
		}
	}
	return n
}

// renumber closes the gaps in a column's positions, keeping its order.
func renumber(tasks []Task, day string, c Column) {
	var col []int
	for j, t := range tasks {
		if t.Day == day && t.Column == c {
			col = append(col, j)
		}
	}
	sort.SliceStable(col, func(a, b int) bool { return tasks[col[a]].Position < tasks[col[b]].Position })
	for pos, j := range col {
		tasks[j].Position = pos
	}
}

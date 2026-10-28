// Package theme holds the looks DataVault can wear: the colors, the names of
// the pages and their features, the 1-bit icons and the login art. A few
// themes are built in; more can be imported as JSON files and live next to
// the config, in the themes directory.
package theme

import (
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

// Default is the theme DataVault starts with, and the one every other theme
// falls back to for whatever it leaves out.
const Default = "souls"

// Colors are the CSS variables a theme sets, in the order the style guide lists them.
var Colors = []string{"bg", "raise", "line", "dim", "muted", "fg", "accent"}

// Words are the names a theme gives to the pages and their features.
var Words = []string{
	// the pages
	"grimoire", "unspoken", "siege", "crypt", "ledger",
	// grimoire: the password generator, and sealing an entry behind the seal PIN
	"forge", "seal", "sealed",
	// unspoken: where buried notes go, and moving a note there and back
	"dead_words", "bury", "unbury",
	// siege: the four columns
	"orders", "in_battle", "conquered", "fallen",
	// crypt: putting a file in, and the label for when it went in
	"keep", "kept",
	// ledger: deleting a bill
	"strike",
	// the sidebar button that ends the session
	"end",
}

// Icons are the 16x16 pixel art slots: one per page, the end button and the seal.
var Icons = []string{"grimoire", "unspoken", "siege", "crypt", "ledger", "end", "seal"}

// Limits for imported themes.
const (
	IconSize      = 16
	maxWord       = 24
	maxName       = 32
	maxDesc       = 160
	maxLoginLines = 24
	maxLoginCols  = 160
)

type Theme struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	// Builtin is set by DataVault; an imported file cannot claim it.
	Builtin bool              `json:"builtin"`
	Colors  map[string]string `json:"colors"`
	Words   map[string]string `json:"words"`
	// Icons maps a slot to 16 rows of 16 characters: '#' lit, '.' dark.
	Icons map[string][]string `json:"icons"`
	// LargeIcons are drawn a step larger in the sidebar, for art that leaves
	// more of its grid empty than the others.
	LargeIcons []string `json:"large_icons,omitempty"`
	// Login is the art above the login prompt, one string per line.
	Login []string `json:"login"`
}

// Summary is what the theme picker lists.
type Summary struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Builtin     bool   `json:"builtin"`
}

func (t Theme) Summary() Summary {
	return Summary{ID: t.ID, Name: t.Name, Description: t.Description, Builtin: t.Builtin}
}

//go:embed builtin/*.json
var builtinFS embed.FS

var builtins = loadBuiltins()

func loadBuiltins() map[string]Theme {
	out := map[string]Theme{}
	files, _ := fs.Glob(builtinFS, "builtin/*.json")
	for _, f := range files {
		data, err := builtinFS.ReadFile(f)
		if err != nil {
			panic(err)
		}
		t, err := parse(data)
		if err != nil {
			panic(fmt.Sprintf("built-in theme %s: %v", f, err))
		}
		t.Builtin = true
		out[t.ID] = t
	}
	if _, ok := out[Default]; !ok {
		panic("the default theme is not built in")
	}
	// Every built-in theme is complete, so imported ones can lean on any of them.
	for id, t := range out {
		if err := t.complete(); err != nil {
			panic(fmt.Sprintf("built-in theme %s: %v", id, err))
		}
	}
	return out
}

// IsBuiltin reports whether id names a theme that ships with DataVault.
func IsBuiltin(id string) bool {
	_, ok := builtins[id]
	return ok
}

var idPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,31}$`)

// ValidID reports whether id can name a theme (and its file).
func ValidID(id string) bool {
	return idPattern.MatchString(id)
}

var colorPattern = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)

// Parse reads an imported theme. Anything it leaves out comes from the default
// theme, so a theme can change only the colors, or only a few names.
func Parse(data []byte) (Theme, error) {
	t, err := parse(data)
	if err != nil {
		return Theme{}, err
	}
	if IsBuiltin(t.ID) {
		return Theme{}, fmt.Errorf("%q is a built-in theme; choose another id", t.ID)
	}
	t.Builtin = false
	return t, nil
}

func parse(data []byte) (Theme, error) {
	var t Theme
	dec := json.NewDecoder(strings.NewReader(string(data)))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&t); err != nil {
		return Theme{}, fmt.Errorf("not a theme file: %w", err)
	}
	if err := t.validate(); err != nil {
		return Theme{}, err
	}
	return t, nil
}

func (t Theme) validate() error {
	if !ValidID(t.ID) {
		return errors.New("id must be 1 to 32 lowercase letters, digits or dashes")
	}
	if name := strings.TrimSpace(t.Name); name == "" || utf8.RuneCountInString(name) > maxName {
		return fmt.Errorf("name must be 1 to %d characters", maxName)
	}
	if utf8.RuneCountInString(t.Description) > maxDesc {
		return fmt.Errorf("description must be at most %d characters", maxDesc)
	}
	for k, v := range t.Colors {
		if !slices.Contains(Colors, k) {
			return fmt.Errorf("unknown color %q (known: %s)", k, strings.Join(Colors, ", "))
		}
		if !colorPattern.MatchString(v) {
			return fmt.Errorf("color %s must look like #1a1a1a, got %q", k, v)
		}
	}
	for k, v := range t.Words {
		if !slices.Contains(Words, k) {
			return fmt.Errorf("unknown word %q (known: %s)", k, strings.Join(Words, ", "))
		}
		if w := strings.TrimSpace(v); w == "" || utf8.RuneCountInString(w) > maxWord || strings.ContainsAny(w, "\r\n\t") {
			return fmt.Errorf("word %s must be one line of 1 to %d characters", k, maxWord)
		}
	}
	for k, rows := range t.Icons {
		if !slices.Contains(Icons, k) {
			return fmt.Errorf("unknown icon %q (known: %s)", k, strings.Join(Icons, ", "))
		}
		if err := validIcon(rows); err != nil {
			return fmt.Errorf("icon %s: %w", k, err)
		}
	}
	for _, k := range t.LargeIcons {
		if !slices.Contains(Icons, k) {
			return fmt.Errorf("large_icons: unknown icon %q", k)
		}
	}
	if len(t.Login) > maxLoginLines {
		return fmt.Errorf("login art must have at most %d lines", maxLoginLines)
	}
	for i, line := range t.Login {
		if utf8.RuneCountInString(line) > maxLoginCols || strings.ContainsAny(line, "\r\n\t") {
			return fmt.Errorf("login art line %d must be one line of at most %d characters", i+1, maxLoginCols)
		}
	}
	return nil
}

func validIcon(rows []string) error {
	if len(rows) != IconSize {
		return fmt.Errorf("needs %d rows, has %d", IconSize, len(rows))
	}
	for i, row := range rows {
		if len(row) != IconSize || strings.Trim(row, "#.") != "" {
			return fmt.Errorf("row %d must be %d characters of '#' and '.'", i+1, IconSize)
		}
	}
	return nil
}

// complete reports what a theme still lacks to be drawn on its own.
func (t Theme) complete() error {
	for _, set := range []struct {
		what string
		keys []string
		has  func(string) bool
	}{
		{"color", Colors, func(k string) bool { _, ok := t.Colors[k]; return ok }},
		{"word", Words, func(k string) bool { _, ok := t.Words[k]; return ok }},
		{"icon", Icons, func(k string) bool { _, ok := t.Icons[k]; return ok }},
	} {
		for _, k := range set.keys {
			if !set.has(k) {
				return fmt.Errorf("missing %s %q", set.what, k)
			}
		}
	}
	if len(t.Login) == 0 {
		return errors.New("missing login art")
	}
	return nil
}

// Resolved fills in whatever t leaves out from the default theme. The large
// icons and the login art only come along when the icon or art does too.
func (t Theme) Resolved() Theme {
	base := builtins[Default]
	out := t
	out.Colors = maps.Clone(base.Colors)
	maps.Copy(out.Colors, t.Colors)
	out.Words = maps.Clone(base.Words)
	maps.Copy(out.Words, t.Words)
	out.Icons = maps.Clone(base.Icons)
	maps.Copy(out.Icons, t.Icons)
	out.LargeIcons = slices.Clone(t.LargeIcons)
	for _, k := range base.LargeIcons {
		if _, own := t.Icons[k]; !own && !slices.Contains(out.LargeIcons, k) {
			out.LargeIcons = append(out.LargeIcons, k)
		}
	}
	if len(t.Login) == 0 {
		out.Login = base.Login
	}
	return out
}

// Store reads built-in themes from the binary and imported ones from Dir.
type Store struct {
	Dir string
}

// DirFor is the themes directory that sits next to a config file.
func DirFor(configPath string) string {
	return filepath.Join(filepath.Dir(configPath), "themes")
}

func (s Store) path(id string) string {
	return filepath.Join(s.Dir, id+".json")
}

// Get finds a theme by id, built in or imported, as it was written.
func (s Store) Get(id string) (Theme, error) {
	if t, ok := builtins[id]; ok {
		return t, nil
	}
	if !ValidID(id) || s.Dir == "" {
		return Theme{}, fs.ErrNotExist
	}
	data, err := os.ReadFile(s.path(id))
	if err != nil {
		return Theme{}, err
	}
	t, err := Parse(data)
	if err != nil {
		return Theme{}, err
	}
	if t.ID != id {
		return Theme{}, fmt.Errorf("%s holds theme %q", s.path(id), t.ID)
	}
	return t, nil
}

// Active is the theme to draw: id resolved against the default, or the
// default itself when id is gone or broken.
func (s Store) Active(id string) Theme {
	t, err := s.Get(id)
	if err != nil {
		t = builtins[Default]
	}
	return t.Resolved()
}

// List returns the built-in themes, the default first, then the imported ones
// by name. Files that do not parse are left out.
func (s Store) List() []Summary {
	var built, imported []Summary
	for _, t := range builtins {
		built = append(built, t.Summary())
	}
	sort.Slice(built, func(i, j int) bool {
		if (built[i].ID == Default) != (built[j].ID == Default) {
			return built[i].ID == Default
		}
		return built[i].Name < built[j].Name
	})
	if s.Dir != "" {
		files, _ := filepath.Glob(filepath.Join(s.Dir, "*.json"))
		for _, f := range files {
			if t, err := s.Get(strings.TrimSuffix(filepath.Base(f), ".json")); err == nil {
				imported = append(imported, t.Summary())
			}
		}
		sort.Slice(imported, func(i, j int) bool { return imported[i].Name < imported[j].Name })
	}
	return append(built, imported...)
}

// Import validates data and saves it as an imported theme, replacing an
// imported theme with the same id.
func (s Store) Import(data []byte) (Theme, error) {
	t, err := Parse(data)
	if err != nil {
		return Theme{}, err
	}
	if s.Dir == "" {
		return Theme{}, errors.New("themes cannot be saved")
	}
	out, err := json.MarshalIndent(t, "", "  ")
	if err != nil {
		return Theme{}, err
	}
	if err := os.MkdirAll(s.Dir, 0o700); err != nil {
		return Theme{}, err
	}
	tmp := s.path(t.ID) + ".tmp"
	if err := os.WriteFile(tmp, append(out, '\n'), 0o600); err != nil {
		return Theme{}, err
	}
	if err := os.Rename(tmp, s.path(t.ID)); err != nil {
		return Theme{}, err
	}
	return t, nil
}

// ErrBuiltin is returned when removing a theme that ships with DataVault.
var ErrBuiltin = errors.New("built-in themes cannot be removed")

// Remove deletes an imported theme.
func (s Store) Remove(id string) error {
	if IsBuiltin(id) {
		return ErrBuiltin
	}
	if !ValidID(id) || s.Dir == "" {
		return fs.ErrNotExist
	}
	return os.Remove(s.path(id))
}

package theme

import (
	"embed"
	"slices"
	"strings"
)

// The login art can follow the theme or be one of these, whatever the theme.
const LoginFromTheme = "theme"

// LoginArts lists the login art choices, in the order the settings offer them.
var LoginArts = []string{LoginFromTheme, "poison", "gothic"}

//go:embed login/*.txt
var loginFS embed.FS

// LoginArt returns the lines of a named login art; ok is false for "theme"
// and for names that do not exist.
func LoginArt(name string) (lines []string, ok bool) {
	if name == LoginFromTheme || !slices.Contains(LoginArts, name) {
		return nil, false
	}
	data, err := loginFS.ReadFile("login/" + name + ".txt")
	if err != nil {
		return nil, false
	}
	return strings.Split(strings.TrimRight(string(data), "\n"), "\n"), true
}

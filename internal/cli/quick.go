package cli

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/server"
	"github.com/fbreseghello/DataVault/internal/vault"
)

// Quick writes: a note for Unspoken or an order for today's Siege board, from
// the terminal, into the running session's vault.

const noteUsage = `usage: dtvlt note "title" ["content"]
       dtvlt note "title(content)"`

const taskUsage = `usage: dtvlt task "title" ["detail"]
       dtvlt task "title(detail)"`

// writer is a session that takes quick writes: served here or reached
// through the control API.
type writer interface {
	WriteNote(title, body string) (vault.Note, error)
	WriteTask(title, note, day string) (vault.Task, error)
}

var parens = regexp.MustCompile(`(?s)^([^()]*?)\s*\((.*)\)\s*$`)

// titleAndText reads "title(text)" or "title" "text" into its two parts.
func titleAndText(args []string) (string, string) {
	if m := parens.FindStringSubmatch(strings.Join(args, " ")); m != nil && strings.TrimSpace(m[1]) != "" {
		return strings.TrimSpace(m[1]), strings.TrimSpace(m[2])
	}
	if len(args) == 0 {
		return "", ""
	}
	return strings.TrimSpace(args[0]), strings.TrimSpace(strings.Join(args[1:], " "))
}

func wantsHelp(args []string) bool {
	return len(args) == 1 && (args[0] == "-h" || args[0] == "--help" || args[0] == "help")
}

func quickNote(u *ui, w writer, args []string) error {
	title, body := titleAndText(args)
	if title == "" || wantsHelp(args) {
		u.println("%s", noteUsage)
		return nil
	}
	if _, err := w.WriteNote(title, body); err != nil {
		return lockedHint(err)
	}
	u.println("  written in unspoken  %s", u.dim(title))
	return nil
}

func quickTask(u *ui, w writer, args []string) error {
	title, detail := titleAndText(args)
	if title == "" || wantsHelp(args) {
		u.println("%s", taskUsage)
		return nil
	}
	if _, err := w.WriteTask(title, detail, time.Now().Format(time.DateOnly)); err != nil {
		return lockedHint(err)
	}
	u.println("  ordered on today's siege  %s", u.dim(title))
	return nil
}

func lockedHint(err error) error {
	if errors.Is(err, vault.ErrLocked) {
		return server.ErrVaultLocked
	}
	return err
}

// withWriter runs fn against the running session, if there is one.
func withWriter(u *ui, paths config.Paths, fn func(writer) error) error {
	c, err := server.Dial(paths.State)
	if errors.Is(err, server.ErrNotRunning) {
		return fmt.Errorf("no session running · start one with dtvlt and open the vault")
	}
	if err != nil {
		return err
	}
	return fn(c)
}

// splitArgs splits a console line into words, keeping "quoted text" together.
func splitArgs(line string) []string {
	var out []string
	var cur strings.Builder
	quote, in := rune(0), false
	for _, r := range line {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote == 0 && (r == '"' || r == '\''):
			quote, in = r, true
		case quote == 0 && (r == ' ' || r == '\t'):
			if in {
				out = append(out, cur.String())
				cur.Reset()
				in = false
			}
		default:
			cur.WriteRune(r)
			in = true
		}
	}
	if in {
		out = append(out, cur.String())
	}
	return out
}

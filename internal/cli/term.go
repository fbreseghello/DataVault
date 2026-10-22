package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/term"
)

// ui is the small terminal toolkit dtvlt uses: monochrome, two weights, no boxes.
type ui struct {
	in    *bufio.Reader
	out   io.Writer
	color bool
	// pause keeps a double-clicked console window open after the command ends.
	pause bool

	// mu serializes output: session events arrive from other goroutines while
	// the console waits for input. waiting holds the console prompt being shown,
	// so an event can clear that line and draw it again below.
	mu      sync.Mutex
	waiting string
}

func newUI() *ui {
	color := term.IsTerminal(int(os.Stdout.Fd())) && os.Getenv("NO_COLOR") == "" && enableTerminal()
	return &ui{in: bufio.NewReader(os.Stdin), out: os.Stdout, color: color}
}

func (u *ui) style(code, s string) string {
	if !u.color {
		return s
	}
	return "\x1b[" + code + "m" + s + "\x1b[0m"
}

func (u *ui) dim(s string) string  { return u.style("2", s) }
func (u *ui) bold(s string) string { return u.style("1", s) }

func (u *ui) println(format string, args ...any) {
	u.mu.Lock()
	defer u.mu.Unlock()
	fmt.Fprintf(u.out, format+"\n", args...)
}

// row prints an aligned "  label  value" line.
func (u *ui) row(label, value string) {
	u.println("  %s  %s", u.dim(fmt.Sprintf("%-8s", label)), value)
}

func (u *ui) fail(format string, args ...any) {
	u.println("  "+u.bold("!")+" "+format, args...)
}

// event prints a timestamped line for the session log, above the console
// prompt if one is waiting.
func (u *ui) event(e string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if u.waiting != "" {
		if u.color {
			fmt.Fprint(u.out, "\r\x1b[2K")
		} else {
			fmt.Fprintln(u.out)
		}
	}
	fmt.Fprintf(u.out, "%s  %s\n", u.dim(time.Now().Format("15:04")), e)
	if u.waiting != "" {
		fmt.Fprint(u.out, u.waiting)
	}
}

// command shows the console prompt and reads one line.
func (u *ui) command(prompt string) (string, error) {
	u.mu.Lock()
	u.waiting = prompt
	fmt.Fprint(u.out, prompt)
	u.mu.Unlock()
	line, err := u.readLine()
	u.mu.Lock()
	u.waiting = ""
	u.mu.Unlock()
	return line, err
}

func (u *ui) prompt(label, hint string) {
	u.mu.Lock()
	defer u.mu.Unlock()
	if hint != "" {
		fmt.Fprintf(u.out, "  %-22s %s ", label, u.dim(hint))
	} else {
		fmt.Fprintf(u.out, "  %-22s ", label)
	}
}

var errAborted = errors.New("aborted")

func (u *ui) readLine() (string, error) {
	line, err := u.in.ReadString('\n')
	if err != nil && (line == "" || !errors.Is(err, io.EOF)) {
		return "", errAborted
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func (u *ui) confirm(label string, def bool) (bool, error) {
	hint := "[y/N]"
	if def {
		hint = "[Y/n]"
	}
	for {
		u.prompt(label, hint)
		s, err := u.readLine()
		if err != nil {
			return false, err
		}
		switch strings.ToLower(strings.TrimSpace(s)) {
		case "":
			return def, nil
		case "y", "yes", "s", "sim":
			return true, nil
		case "n", "no", "nao", "não":
			return false, nil
		}
	}
}

// choose asks for one of options; an empty answer keeps def.
func (u *ui) choose(label, def string, options []string) (string, error) {
	for {
		u.prompt(label, "["+def+"] "+strings.Join(options, "/"))
		s, err := u.readLine()
		if err != nil {
			return "", err
		}
		s = strings.ToLower(strings.TrimSpace(s))
		if s == "" {
			return def, nil
		}
		if slices.Contains(options, s) {
			return s, nil
		}
		u.fail("enter one of %s", strings.Join(options, ", "))
	}
}

func (u *ui) number(label string, def, lo, hi int) (int, error) {
	for {
		u.prompt(label, fmt.Sprintf("[%d]", def))
		s, err := u.readLine()
		if err != nil {
			return 0, err
		}
		s = strings.TrimSpace(s)
		if s == "" {
			return def, nil
		}
		if n, err := strconv.Atoi(s); err == nil && n >= lo && n <= hi {
			return n, nil
		}
		u.fail("enter a number from %d to %d", lo, hi)
	}
}

// secret reads without echo. When stdin is not a terminal (scripts, tests) it
// falls back to a plain line.
func (u *ui) secret(label string) ([]byte, error) {
	u.prompt(label, "")
	fd := int(os.Stdin.Fd())
	if !term.IsTerminal(fd) {
		s, err := u.readLine()
		u.println("")
		return []byte(s), err
	}
	b, err := term.ReadPassword(fd)
	u.println("")
	if err != nil {
		return nil, errAborted
	}
	return b, nil
}

// newSecret asks twice and enforces a minimum length.
func (u *ui) newSecret(label string, minLen int) ([]byte, error) {
	for {
		a, err := u.secret(label)
		if err != nil {
			return nil, err
		}
		if len([]rune(string(a))) < minLen {
			u.fail("use at least %d characters", minLen)
			clear(a)
			continue
		}
		b, err := u.secret("confirm " + label)
		if err != nil {
			clear(a)
			return nil, err
		}
		same := string(a) == string(b)
		clear(b)
		if same {
			return a, nil
		}
		clear(a)
		u.fail("they don't match, try again")
	}
}

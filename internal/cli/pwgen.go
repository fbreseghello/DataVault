package cli

import (
	"fmt"
	"os"
	"strconv"

	"golang.org/x/term"

	"github.com/fbreseghello/DataVault/internal/forge"
)

const pwgenUsage = `usage: dtvlt pwgen [length] [options]

  length        8 to 128 (default 24)
  --plain       leave out lookalikes: 0 O o 1 l I |
  --no-upper    no capital letters
  --no-lower    no small letters
  --no-digits   no digits
  --no-symbols  no symbols`

// parsePwgen reads the arguments of pwgen into forge options.
func parsePwgen(args []string) (forge.Options, error) {
	o := forge.Defaults()
	for _, a := range args {
		switch a {
		case "--plain":
			o.Plain = true
		case "--no-upper":
			o.Upper = false
		case "--no-lower":
			o.Lower = false
		case "--no-digits":
			o.Digits = false
		case "--no-symbols":
			o.Symbols = false
		default:
			n, err := strconv.Atoi(a)
			if err != nil {
				return o, fmt.Errorf("unknown option %q", a)
			}
			if n < forge.MinLength || n > forge.MaxLength {
				return o, fmt.Errorf("length must be %d to %d", forge.MinLength, forge.MaxLength)
			}
			o.Length = n
		}
	}
	return o, nil
}

// pwgen forges a password. It needs no session and saves nothing. Piped, it
// prints the password alone so it can feed another command.
func pwgen(u *ui, args []string) error {
	for _, a := range args {
		if a == "-h" || a == "--help" || a == "help" {
			u.println("%s", pwgenUsage)
			return nil
		}
	}
	o, err := parsePwgen(args)
	if err != nil {
		return fmt.Errorf("%v · dtvlt pwgen --help", err)
	}
	p, err := forge.Forge(o)
	if err != nil {
		return err
	}
	if !term.IsTerminal(int(os.Stdout.Fd())) {
		u.println("%s", p)
		return nil
	}
	u.println("")
	u.println("  %s", u.bold(p))
	u.println("  %s", u.dim(fmt.Sprintf("%d bits · never saved", forge.Bits(o))))
	u.println("")
	return nil
}

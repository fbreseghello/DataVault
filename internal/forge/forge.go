// Package forge makes passwords from the system's cryptographic randomness. It
// draws every character without bias, and each kind asked for appears at least
// once. It mirrors web/src/lib/forge.ts, the forge of the grimoire.
package forge

import (
	"crypto/rand"
	"errors"
	"math"
	"math/big"
	"strings"
)

const (
	MinLength     = 8
	MaxLength     = 128
	DefaultLength = 24
)

// Options says what a password is made of.
type Options struct {
	Length  int
	Upper   bool
	Lower   bool
	Digits  bool
	Symbols bool
	// Plain leaves out characters that are easy to mistake for one another.
	Plain bool
}

// Defaults are the options of the grimoire's forge.
func Defaults() Options {
	return Options{Length: DefaultLength, Upper: true, Lower: true, Digits: true, Symbols: true}
}

const (
	upper      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	lower      = "abcdefghijklmnopqrstuvwxyz"
	digits     = "0123456789"
	symbols    = "!#$%&*+-=?@^_~.:;,/|()[]{}<>"
	lookalikes = "0Oo1lI|"
)

var ErrNoKinds = errors.New("choose at least one kind of character")

func (o Options) sets() []string {
	var out []string
	for _, k := range []struct {
		on  bool
		set string
	}{{o.Upper, upper}, {o.Lower, lower}, {o.Digits, digits}, {o.Symbols, symbols}} {
		if !k.on {
			continue
		}
		set := k.set
		if o.Plain {
			set = strings.Map(func(r rune) rune {
				if strings.ContainsRune(lookalikes, r) {
					return -1
				}
				return r
			}, set)
		}
		if set != "" {
			out = append(out, set)
		}
	}
	return out
}

func (o Options) length() int {
	return min(MaxLength, max(MinLength, o.Length))
}

// below is a uniform integer in [0, n).
func below(n int) int {
	v, err := rand.Int(rand.Reader, big.NewInt(int64(n)))
	if err != nil {
		panic("forge: no randomness: " + err.Error())
	}
	return int(v.Int64())
}

// Forge makes a password for the options.
func Forge(o Options) (string, error) {
	sets := o.sets()
	if len(sets) == 0 {
		return "", ErrNoKinds
	}
	pool := strings.Join(sets, "")
	// one of each kind first, the rest from the whole pool, then shuffled
	chars := make([]byte, 0, o.length())
	for _, s := range sets {
		chars = append(chars, s[below(len(s))])
	}
	for len(chars) < o.length() {
		chars = append(chars, pool[below(len(pool))])
	}
	for i := len(chars) - 1; i > 0; i-- {
		j := below(i + 1)
		chars[i], chars[j] = chars[j], chars[i]
	}
	return string(chars), nil
}

// Bits is how hard the password is to guess: length × log2(pool size).
func Bits(o Options) int {
	pool := len(strings.Join(o.sets(), ""))
	if pool == 0 {
		return 0
	}
	return int(float64(o.length()) * math.Log2(float64(pool)))
}

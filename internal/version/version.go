// Package version holds the dtvlt version, which follows Semantic Versioning.
//
// Release builds set it from the git tag (v1.2.3 -> "1.2.3"):
//
//	go build -ldflags "-X github.com/fbreseghello/DataVault/internal/version.Version=1.2.3"
//
// `go run ./scripts/build` does this automatically.
package version

import (
	"strconv"
	"strings"
)

// Version is "dev" for plain `go build` and "x.y.z" (optionally "-pre") for releases.
var Version = "dev"

// IsDev reports whether this is a local build rather than a tagged release.
func IsDev() bool {
	return Version == "dev" || strings.Contains(Version, "-dev")
}

// Compare orders two semantic versions: -1 if a < b, 0 if equal, 1 if a > b.
// A pre-release sorts before its release (1.0.0-rc.1 < 1.0.0). Build metadata
// is ignored. Unparseable parts compare as zero.
func Compare(a, b string) int {
	ca, pa := split(a)
	cb, pb := split(b)
	for i := range 3 {
		if ca[i] != cb[i] {
			if ca[i] < cb[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case pa == pb:
		return 0
	case pa == "":
		return 1
	case pb == "":
		return -1
	}
	return comparePre(strings.Split(pa, "."), strings.Split(pb, "."))
}

func split(v string) ([3]int, string) {
	v = strings.TrimPrefix(v, "v")
	v, _, _ = strings.Cut(v, "+")
	core, pre, _ := strings.Cut(v, "-")
	var n [3]int
	for i, p := range strings.SplitN(core, ".", 3) {
		n[i], _ = strconv.Atoi(p)
	}
	return n, pre
}

// comparePre follows semver: numeric identifiers compare as numbers and sort
// before alphanumeric ones; a shorter list of equal identifiers sorts first.
func comparePre(a, b []string) int {
	for i := 0; i < len(a) && i < len(b); i++ {
		na, errA := strconv.Atoi(a[i])
		nb, errB := strconv.Atoi(b[i])
		switch {
		case errA == nil && errB == nil:
			if na != nb {
				if na < nb {
					return -1
				}
				return 1
			}
		case errA == nil:
			return -1
		case errB == nil:
			return 1
		case a[i] != b[i]:
			if a[i] < b[i] {
				return -1
			}
			return 1
		}
	}
	switch {
	case len(a) < len(b):
		return -1
	case len(a) > len(b):
		return 1
	}
	return 0
}

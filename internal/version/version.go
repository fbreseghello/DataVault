// Package version holds the dtvlt version, which follows Semantic Versioning.
//
// Release builds set it from the git tag (v1.2.3 -> "1.2.3"):
//
//	go build -ldflags "-X github.com/fbreseghello/DataVault/internal/version.Version=1.2.3"
//
// `go run ./scripts/build` does this automatically.
package version

import "strings"

// Version is "dev" for plain `go build` and "x.y.z" (optionally "-pre") for releases.
var Version = "dev"

// IsDev reports whether this is a local build rather than a tagged release.
func IsDev() bool {
	return Version == "dev" || strings.Contains(Version, "-dev")
}

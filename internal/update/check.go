package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/fbreseghello/DataVault/internal/version"
)

// CheckEvery is how often the startup check may ask GitHub.
const CheckEvery = 24 * time.Hour

type checkCache struct {
	CheckedAt time.Time `json:"checked_at"`
	// Version is the latest release seen, empty when the check failed. A failed
	// check is cached too, so an offline machine does not wait on every start.
	Version string `json:"version,omitempty"`
}

// Available returns the newer release version, or "" when this build is up
// to date, is a dev build, or nothing could be learned in time. GitHub is
// asked at most once per CheckEvery; in between the cached answer is used.
func (c *Client) Available(ctx context.Context, cachePath string, now time.Time) string {
	if version.IsDev() {
		return ""
	}
	var cache checkCache
	if b, err := os.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(b, &cache)
	}
	if cache.CheckedAt.IsZero() || now.Sub(cache.CheckedAt) >= CheckEvery || now.Before(cache.CheckedAt) {
		cache = checkCache{CheckedAt: now}
		if rel, err := c.Latest(ctx); err == nil {
			cache.Version = rel.Version
		}
		if b, err := json.Marshal(cache); err == nil {
			_ = os.MkdirAll(filepath.Dir(cachePath), 0o700)
			_ = os.WriteFile(cachePath, b, 0o600)
		}
	}
	if cache.Version != "" && version.Compare(cache.Version, version.Version) > 0 {
		return cache.Version
	}
	return ""
}

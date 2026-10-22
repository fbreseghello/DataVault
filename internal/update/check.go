package update

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"time"

	"github.com/fbreseghello/DataVault/internal/version"
)

type checkCache struct {
	CheckedAt time.Time `json:"checked_at"`
	Version   string    `json:"version,omitempty"`
}

// Available returns the newer release version, or "" when this build is up to
// date, is a dev build, or nothing is known. It asks GitHub every time; the
// last answer is kept so an offline start still announces a release seen before.
func (c *Client) Available(ctx context.Context, cachePath string, now time.Time) string {
	if version.IsDev() {
		return ""
	}
	var cache checkCache
	if rel, err := c.Latest(ctx); err == nil {
		cache = checkCache{CheckedAt: now, Version: rel.Version}
		if b, err := json.Marshal(cache); err == nil {
			_ = os.MkdirAll(filepath.Dir(cachePath), 0o700)
			_ = os.WriteFile(cachePath, b, 0o600)
		}
	} else if b, err := os.ReadFile(cachePath); err == nil {
		_ = json.Unmarshal(b, &cache)
	}
	if cache.Version != "" && version.Compare(cache.Version, version.Version) > 0 {
		return cache.Version
	}
	return ""
}

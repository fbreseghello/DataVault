// Package config holds user preferences and the on-disk locations of DataVault files.
package config

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	PinEnabled            bool   `toml:"pin_enabled"`
	PinHash               string `toml:"pin_hash,omitempty"`
	KeepTerminal          bool   `toml:"keep_terminal"`
	AutoLockMinutes       int    `toml:"auto_lock_minutes"`
	ClipboardClearSeconds int    `toml:"clipboard_clear_seconds"`
	// CheckUpdates lets dtvlt ask GitHub for a newer release, each time it starts.
	CheckUpdates bool `toml:"check_updates"`
	// RestMinutes is the idle time before the login shows the rest screen; 0 turns it off.
	RestMinutes int `toml:"rest_minutes"`
	// RestStyle picks what the rest screen shows: one of RestStyles.
	RestStyle string `toml:"rest_style"`
}

// Rest screen styles: one of the still images (the broken pillar or one of
// two skulls), or one of them drawn at random each time the screen rests.
const (
	RestPillar  = "pillar"
	RestSkullI  = "skull-i"
	RestSkullII = "skull-ii"
	RestRandom  = "random"
)

// RestStyles lists every rest style, in the order the screens offer them.
var RestStyles = []string{RestPillar, RestSkullI, RestSkullII, RestRandom}

// MaxRestMinutes caps the rest screen delay.
const MaxRestMinutes = 240

func Default() Config {
	return Config{KeepTerminal: true, AutoLockMinutes: 15, ClipboardClearSeconds: 20, CheckUpdates: true, RestMinutes: 15, RestStyle: RestPillar}
}

func (c Config) Validate() error {
	if c.AutoLockMinutes < 1 || c.AutoLockMinutes > 24*60 {
		return errors.New("auto_lock_minutes must be between 1 and 1440")
	}
	if c.ClipboardClearSeconds < 5 || c.ClipboardClearSeconds > 600 {
		return errors.New("clipboard_clear_seconds must be between 5 and 600")
	}
	if c.RestMinutes < 0 || c.RestMinutes > MaxRestMinutes {
		return fmt.Errorf("rest_minutes must be between 0 (off) and %d", MaxRestMinutes)
	}
	if !slices.Contains(RestStyles, c.RestStyle) {
		return fmt.Errorf("rest_style must be one of %s", strings.Join(RestStyles, ", "))
	}
	if c.PinEnabled && c.PinHash == "" {
		return errors.New("pin_enabled is set but no pin is configured")
	}
	return nil
}

// Paths are the files DataVault reads and writes.
type Paths struct {
	Config string
	Vault  string
	// State describes the running session so other dtvlt invocations can reach it.
	State string
	// UpdateCache remembers the last update check.
	UpdateCache string
}

// ResolvePaths follows the platform conventions. DTVLT_HOME puts every file in
// one directory, which is handy for tests and portable installs.
func ResolvePaths() (Paths, error) {
	if home := os.Getenv("DTVLT_HOME"); home != "" {
		return pathsIn(home, home), nil
	}
	switch runtime.GOOS {
	case "windows":
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return Paths{}, errors.New("APPDATA is not set")
		}
		dir := filepath.Join(appData, "DataVault")
		return pathsIn(dir, dir), nil
	default:
		home, err := os.UserHomeDir()
		if err != nil {
			return Paths{}, err
		}
		cfg := envOr("XDG_CONFIG_HOME", filepath.Join(home, ".config"))
		data := envOr("XDG_DATA_HOME", filepath.Join(home, ".local", "share"))
		return pathsIn(filepath.Join(cfg, "datavault"), filepath.Join(data, "datavault")), nil
	}
}

func pathsIn(configDir, dataDir string) Paths {
	return Paths{
		Config:      filepath.Join(configDir, "config.toml"),
		Vault:       filepath.Join(dataDir, "vault.dtv"),
		State:       filepath.Join(dataDir, "session.json"),
		UpdateCache: filepath.Join(dataDir, "update-check.json"),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

// Load reads the config file. A missing file returns an error matching fs.ErrNotExist.
func Load(path string) (Config, error) {
	cfg := Default()
	if _, err := toml.DecodeFile(path, &cfg); err != nil {
		return Config{}, err
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, fmt.Errorf("%s: %w", path, err)
	}
	return cfg, nil
}

func Save(path string, cfg Config) error {
	if err := cfg.Validate(); err != nil {
		return err
	}
	var buf bytes.Buffer
	buf.WriteString("# DataVault preferences. Change them in the settings screen or with `dtvlt config`.\n\n")
	if err := toml.NewEncoder(&buf).Encode(cfg); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, buf.Bytes(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, path)
}

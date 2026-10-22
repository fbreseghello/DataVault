package cli

import (
	"bytes"
	"errors"
	"io/fs"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/vault"
)

const (
	minKeyword  = 4
	minPassword = 10
)

// ensureSetup runs the first-run wizard for whatever is missing (preferences,
// vault, or both) and returns the config to use.
func ensureSetup(u *ui, paths config.Paths) (config.Config, error) {
	cfg, err := config.Load(paths.Config)
	needConfig := errors.Is(err, fs.ErrNotExist)
	if err != nil && !needConfig {
		return config.Config{}, err
	}
	needVault := !vault.Exists(paths.Vault)
	if !needConfig && !needVault {
		return cfg, nil
	}

	u.println("  %s", u.bold("first run"))
	u.println("")
	if needConfig {
		if cfg, err = askPreferences(u, config.Default()); err != nil {
			return config.Config{}, err
		}
	}
	if needVault {
		if err := createVault(u, paths.Vault); err != nil {
			return config.Config{}, err
		}
	}
	if needConfig {
		if err := config.Save(paths.Config, cfg); err != nil {
			return config.Config{}, err
		}
		u.row("config", paths.Config)
	}
	u.println("")
	return cfg, nil
}

// askPreferences walks through the settings, using cur as defaults.
func askPreferences(u *ui, cur config.Config) (config.Config, error) {
	cfg := cur
	var err error
	if cfg.PinEnabled, err = u.confirm("pin on start", cur.PinEnabled); err != nil {
		return cfg, err
	}
	if cfg.PinEnabled {
		change := true
		if cur.PinHash != "" {
			if change, err = u.confirm("change pin", false); err != nil {
				return cfg, err
			}
		}
		if change {
			u.println("  %s", u.dim("the pin only gates starting dtvlt. it does not encrypt anything."))
			pin, err := u.newSecret("pin", config.MinPIN)
			if err != nil {
				return cfg, err
			}
			cfg.PinHash = config.HashPIN(pin)
			clear(pin)
		}
	} else {
		cfg.PinHash = ""
	}
	if cfg.KeepTerminal, err = u.confirm("keep terminal open", cur.KeepTerminal); err != nil {
		return cfg, err
	}
	if cfg.AutoLockMinutes, err = u.number("auto-lock (minutes)", cur.AutoLockMinutes, 1, 24*60); err != nil {
		return cfg, err
	}
	if cfg.ClipboardClearSeconds, err = u.number("clear clipboard (s)", cur.ClipboardClearSeconds, 5, 600); err != nil {
		return cfg, err
	}
	u.println("  %s", u.dim("once a day dtvlt can ask github.com for a newer version. nothing else is sent."))
	if cfg.CheckUpdates, err = u.confirm("check for updates", cur.CheckUpdates); err != nil {
		return cfg, err
	}
	return cfg, nil
}

func createVault(u *ui, path string) error {
	u.println("")
	u.println("  %s", u.dim("the vault is encrypted with a keyword and a password."))
	u.println("  %s", u.dim("both are needed to open it. there is no recovery."))
	u.println("")

	keyword, err := u.newSecret("keyword", minKeyword)
	if err != nil {
		return err
	}
	defer clear(keyword)
	var password []byte
	for {
		if password, err = u.newSecret("password", minPassword); err != nil {
			return err
		}
		if !bytes.Equal(password, keyword) {
			break
		}
		clear(password)
		u.fail("the password must differ from the keyword")
	}
	defer clear(password)

	u.println("")
	u.println("  %s", u.dim("deriving key (argon2id, 64 MiB)..."))
	v, err := vault.Create(path, keyword, password, vault.DefaultKDF)
	if err != nil {
		return err
	}
	v.Lock()
	u.row("vault", path)
	return nil
}

// configure is `dtvlt config`: edit preferences without touching the vault.
func configure(u *ui, paths config.Paths) error {
	cur, err := config.Load(paths.Config)
	if errors.Is(err, fs.ErrNotExist) {
		cur = config.Default()
	} else if err != nil {
		return err
	}
	cfg, err := askPreferences(u, cur)
	if err != nil {
		return err
	}
	if err := config.Save(paths.Config, cfg); err != nil {
		return err
	}
	u.println("")
	u.row("saved", paths.Config)
	u.println("  %s", u.dim("changes apply the next time dtvlt starts."))
	return nil
}

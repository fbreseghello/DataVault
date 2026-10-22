package cli

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/install"
	"github.com/fbreseghello/DataVault/internal/server"
	"github.com/fbreseghello/DataVault/internal/update"
	"github.com/fbreseghello/DataVault/internal/version"
)

// actsAsInstaller reports whether plain `dtvlt` should install instead of
// starting a session: a release binary running from anywhere but the install
// location is a download. Local dev builds always run normally.
func actsAsInstaller() bool {
	return !version.IsDev() && !install.RunningInstalled()
}

// installSelf is what a downloaded binary does when opened: install or update
// the per-user copy, then tell the user to type dtvlt. It never starts a session.
func installSelf(u *ui, paths config.Paths) error {
	u.pause = true
	if install.Installed() {
		current := install.InstalledVersion()
		switch {
		case current == version.Version:
			u.row("ready", "dtvlt "+version.Version+" is already installed")
			u.println("")
			u.println("  %s", u.dim("open a terminal and type dtvlt to start."))
			return nil
		case current != "" && version.Compare(version.Version, current) < 0:
			ok, err := u.confirm("replace the newer "+current+" with "+version.Version, false)
			if err != nil || !ok {
				return err
			}
		case current != "":
			u.row("update", current+" → "+version.Version)
		}
	}
	if err := runInstall(u); err != nil {
		return err
	}
	if _, err := server.Dial(paths.State); err == nil {
		u.println("  %s", u.dim("the session still runs the old version. stop it, then run dtvlt again."))
	}
	return nil
}

func runInstall(u *ui) error {
	u.pause = true
	res, err := install.Install()
	if err != nil {
		return err
	}
	printInstalled(u, res)
	return nil
}

func printInstalled(u *ui, res install.Result) {
	u.row("installed", res.Path)
	u.println("")
	switch {
	case res.PathUpdated:
		u.println("  %s", u.dim("open a new terminal and type dtvlt to start."))
	case !res.OnPath:
		dir, _ := install.Dir()
		if runtime.GOOS == "windows" {
			u.println("  %s", u.dim("add "+dir+" to your PATH, then type dtvlt in a terminal."))
		} else {
			u.println("  %s", u.dim("add this to your shell profile, then type dtvlt:"))
			u.println("  export PATH=\"%s:$PATH\"", dir)
		}
	default:
		u.println("  %s", u.dim("type dtvlt in a terminal to start."))
	}
}

// updateCheckTimeout caps how long startup waits on GitHub. A slow or missing
// network skips the notice (and is cached for a day) instead of delaying the PIN.
const updateCheckTimeout = 1500 * time.Millisecond

// updateNotice tells the user, before the PIN, that a newer release exists.
func updateNotice(u *ui, paths config.Paths) {
	ctx, cancel := context.WithTimeout(context.Background(), updateCheckTimeout)
	defer cancel()
	c := &update.Client{HTTP: &http.Client{Timeout: updateCheckTimeout}}
	if latest := c.Available(ctx, paths.UpdateCache, time.Now()); latest != "" {
		u.row("update", u.bold(latest)+" available"+u.dim("  (you have "+version.Version+") · type update"))
		u.println("")
	}
}

// runUpdate is `dtvlt update`: install the newest GitHub release, verified
// against its SHA256SUMS. It reports whether a new version was installed.
// From the console the caller restarts on the new version itself, so the
// "run dtvlt again" hints are left out.
func runUpdate(u *ui, paths config.Paths, fromConsole bool) (bool, error) {
	u.pause = true
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	c := &update.Client{}

	u.row("checking", "github.com/"+update.Repo)
	rel, err := c.Latest(ctx)
	if errors.Is(err, update.ErrNoRelease) {
		return false, err
	}
	if err != nil {
		return false, fmt.Errorf("could not get the latest release: %w", err)
	}

	// What matters is the copy on the PATH, not necessarily this process.
	current := version.Version
	if install.Installed() {
		current = install.InstalledVersion()
	} else if version.IsDev() {
		current = ""
	}
	have := current
	if have == "" {
		have = "not installed"
	}
	u.row("latest", rel.Version+u.dim("  (you have "+have+")"))
	if current != "" && version.Compare(rel.Version, current) <= 0 {
		u.println("")
		u.println("  already up to date")
		return false, nil
	}

	dir, err := install.Dir()
	if err != nil {
		return false, err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false, err
	}
	d, err := c.Fetch(ctx, rel, dir)
	if err != nil {
		return false, err
	}
	defer d.Remove()
	u.row("download", fmt.Sprintf("%s  %.1f MiB", d.Asset, float64(d.Size)/(1<<20)))
	u.row("verified", "sha256 "+d.SHA256[:16]+"…")

	res, err := install.InstallFile(d.Path)
	if err != nil {
		return false, err
	}
	u.row("notes", rel.URL)
	if fromConsole {
		u.row("installed", res.Path)
		return true, nil
	}
	printInstalled(u, res)
	if _, err := server.Dial(paths.State); err == nil {
		u.println("  %s", u.dim("the session still runs the old version. stop it, then run dtvlt again."))
	}
	return true, nil
}

func runUninstall(u *ui, paths config.Paths) error {
	u.pause = true
	if !install.Installed() {
		u.println("  %s", u.dim("dtvlt is not installed"))
		return nil
	}
	ok, err := u.confirm("uninstall dtvlt", false)
	if err != nil || !ok {
		return err
	}
	if c, err := server.Dial(paths.State); err == nil {
		_ = c.Stop()
	} else if !errors.Is(err, server.ErrNotRunning) {
		return err
	}
	if err := install.Uninstall(); err != nil {
		return err
	}
	u.println("  uninstalled")
	u.println("  %s", u.dim("your vault and settings were kept:"))
	u.row("vault", paths.Vault)
	u.row("config", paths.Config)
	return nil
}

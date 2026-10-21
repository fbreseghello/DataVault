package cli

import (
	"context"
	"errors"
	"fmt"
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
		u.println("  %s", u.dim("a session is running on the old version. dtvlt stop, then dtvlt, to switch."))
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

// runUpdate is `dtvlt update`: install the newest GitHub release, verified
// against its SHA256SUMS. It only touches the network when asked.
func runUpdate(u *ui, paths config.Paths) error {
	u.pause = true
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	c := &update.Client{}

	u.row("checking", "github.com/"+update.Repo)
	rel, err := c.Latest(ctx)
	if errors.Is(err, update.ErrNoRelease) {
		return err
	}
	if err != nil {
		return fmt.Errorf("could not get the latest release: %w", err)
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
		return nil
	}

	dir, err := install.Dir()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	d, err := c.Fetch(ctx, rel, dir)
	if err != nil {
		return err
	}
	defer d.Remove()
	u.row("download", fmt.Sprintf("%s  %.1f MiB", d.Asset, float64(d.Size)/(1<<20)))
	u.row("verified", "sha256 "+d.SHA256[:16]+"…")

	res, err := install.InstallFile(d.Path)
	if err != nil {
		return err
	}
	printInstalled(u, res)
	if _, err := server.Dial(paths.State); err == nil {
		u.println("  %s", u.dim("a session is running on the old version. dtvlt stop, then dtvlt, to switch."))
	}
	u.row("notes", rel.URL)
	return nil
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

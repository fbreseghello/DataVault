package cli

import (
	"errors"
	"runtime"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/install"
	"github.com/fbreseghello/DataVault/internal/server"
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

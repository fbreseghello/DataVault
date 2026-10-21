package cli

import (
	"errors"
	"runtime"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/install"
	"github.com/fbreseghello/DataVault/internal/server"
	"github.com/fbreseghello/DataVault/internal/version"
)

// offerInstall runs when a release binary starts from somewhere other than
// the install location (a fresh download, typically): it offers to install or
// to update the installed copy. Local dev builds never ask.
func offerInstall(u *ui) error {
	if version.IsDev() || install.RunningInstalled() {
		return nil
	}
	question := "install dtvlt for this user"
	if install.Installed() {
		current := install.InstalledVersion()
		if current == version.Version {
			return nil
		}
		if current == "" {
			current = "?"
		}
		question = "update dtvlt " + current + " → " + version.Version
	}
	ok, err := u.confirm(question, true)
	if err != nil || !ok {
		return err
	}
	return runInstall(u)
}

func runInstall(u *ui) error {
	res, err := install.Install()
	if err != nil {
		return err
	}
	u.row("installed", res.Path)
	switch {
	case res.PathUpdated:
		u.println("  %s", u.dim("added to your PATH. open a new terminal and type dtvlt."))
	case !res.OnPath:
		dir, _ := install.Dir()
		if runtime.GOOS == "windows" {
			u.println("  %s", u.dim("add "+dir+" to your PATH to run dtvlt from any terminal."))
		} else {
			u.println("  %s", u.dim("add this to your shell profile to run dtvlt from any terminal:"))
			u.println("  export PATH=\"%s:$PATH\"", dir)
		}
	default:
		u.println("  %s", u.dim("type dtvlt in any terminal to start."))
	}
	u.println("")
	return nil
}

func runUninstall(u *ui, paths config.Paths) error {
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

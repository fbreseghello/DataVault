package install

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReplaceKeepsContentAndCleansUp(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "new")
	dst := filepath.Join(dir, "installed", exeName)
	os.MkdirAll(filepath.Dir(dst), 0o755)
	os.WriteFile(src, []byte("v2"), 0o755)
	os.WriteFile(dst, []byte("v1"), 0o755)

	if err := replace(src, dst); err != nil {
		t.Fatal(err)
	}
	got, _ := os.ReadFile(dst)
	if string(got) != "v2" {
		t.Fatalf("installed content = %q", got)
	}
	for _, leftover := range []string{dst + ".new", dst + ".old"} {
		if _, err := os.Stat(leftover); err == nil {
			t.Errorf("%s left behind", leftover)
		}
	}
}

func TestCustomDirInstallAndUninstall(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "bin")
	t.Setenv("DTVLT_INSTALL_DIR", dir)

	if Installed() {
		t.Fatal("installed before install")
	}
	res, err := Install()
	if err != nil {
		t.Fatal(err)
	}
	if res.PathUpdated {
		t.Fatal("a custom dir must not touch the PATH")
	}
	if !Installed() {
		t.Fatal("not installed after install")
	}
	if err := Uninstall(); err != nil {
		t.Fatal(err)
	}
	if Installed() {
		t.Fatal("still installed after uninstall")
	}
}

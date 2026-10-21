// Command build compiles dtvlt with the frontend embedded. Run it from the
// repository root:
//
//	go run ./scripts/build                   bin/dtvlt(.exe) for this machine
//	go run ./scripts/build -release          every release target into release/
//	go run ./scripts/build -version 1.2.3    force a version
//
// Without -version the version comes from the git tag on HEAD (v1.2.3 -> 1.2.3),
// or "dev" when HEAD is not tagged or the tree has uncommitted changes. Release
// builds require a real version.
package main

import (
	"archive/tar"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
)

const versionVar = "github.com/fbreseghello/DataVault/internal/version.Version"

// semverRe is the official Semantic Versioning 2.0.0 pattern.
var semverRe = regexp.MustCompile(`^(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)` +
	`(?:-((?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*)(?:\.(?:0|[1-9]\d*|\d*[a-zA-Z-][0-9a-zA-Z-]*))*))?` +
	`(?:\+([0-9a-zA-Z-]+(?:\.[0-9a-zA-Z-]+)*))?$`)

type target struct{ goos, goarch string }

var releaseTargets = []target{
	{"windows", "amd64"},
	{"windows", "arm64"},
	{"linux", "amd64"},
	{"linux", "arm64"},
}

func main() {
	release := flag.Bool("release", false, "build every release target into release/")
	ver := flag.String("version", "", "version to embed (default: from the git tag)")
	skipWeb := flag.Bool("skip-web", false, "reuse the existing web/dist")
	flag.Parse()

	if err := run(*release, *ver, *skipWeb); err != nil {
		fmt.Fprintln(os.Stderr, "build:", err)
		os.Exit(1)
	}
}

func run(release bool, ver string, skipWeb bool) error {
	if _, err := os.Stat("go.mod"); err != nil {
		return errors.New("run this from the repository root")
	}
	ver = strings.TrimPrefix(ver, "v")
	if ver == "" {
		ver = versionFromGit()
	}
	if ver != "dev" && !semverRe.MatchString(ver) {
		return fmt.Errorf("%q is not a semantic version (expected x.y.z)", ver)
	}
	if release && ver == "dev" {
		return errors.New("release builds need a version: tag HEAD (git tag v1.2.3) or pass -version")
	}
	fmt.Println("version", ver)

	if !skipWeb {
		if err := buildWeb(); err != nil {
			return err
		}
	}

	if !release {
		out := filepath.Join("bin", "dtvlt"+exeSuffix(runtime.GOOS))
		if err := goBuild(target{runtime.GOOS, runtime.GOARCH}, ver, out); err != nil {
			return err
		}
		fmt.Println("built", out)
		return nil
	}

	if err := os.RemoveAll("release"); err != nil {
		return err
	}
	var artifacts []string
	for _, t := range releaseTargets {
		name := fmt.Sprintf("dtvlt-%s-%s-%s", ver, t.goos, t.goarch)
		bin := filepath.Join("release", "tmp", name, "dtvlt"+exeSuffix(t.goos))
		if err := goBuild(t, ver, bin); err != nil {
			return err
		}
		var artifact string
		if t.goos == "windows" {
			// A bare .exe: download, double-click, done.
			artifact = filepath.Join("release", name+".exe")
			if err := os.Rename(bin, artifact); err != nil {
				return err
			}
		} else {
			// A tarball keeps the executable bit.
			artifact = filepath.Join("release", name+".tar.gz")
			if err := tarGz(artifact, map[string]string{"dtvlt": bin, "LICENSE": "LICENSE", "README.md": "README.md"}); err != nil {
				return err
			}
		}
		artifacts = append(artifacts, artifact)
		fmt.Println("built", artifact)
	}
	if err := os.RemoveAll(filepath.Join("release", "tmp")); err != nil {
		return err
	}
	return writeChecksums(filepath.Join("release", "SHA256SUMS"), artifacts)
}

// versionFromGit returns the tag on HEAD, but only for a clean tree: a build
// with uncommitted changes is not that release, so it is "dev".
func versionFromGit() string {
	out, err := exec.Command("git", "describe", "--tags", "--exact-match", "--match", "v[0-9]*").Output()
	if err != nil {
		return "dev"
	}
	if dirty, err := exec.Command("git", "status", "--porcelain").Output(); err != nil || len(dirty) > 0 {
		return "dev"
	}
	return strings.TrimPrefix(strings.TrimSpace(string(out)), "v")
}

func buildWeb() error {
	if _, err := os.Stat(filepath.Join("web", "node_modules")); err != nil {
		if err := runCmd(nil, "npm", "--prefix", "web", "ci"); err != nil {
			return err
		}
	}
	return runCmd(nil, "npm", "--prefix", "web", "run", "build")
}

func goBuild(t target, ver, out string) error {
	env := append(os.Environ(), "CGO_ENABLED=0", "GOOS="+t.goos, "GOARCH="+t.goarch)
	return runCmd(env, "go", "build", "-trimpath",
		"-ldflags", fmt.Sprintf("-s -w -X %s=%s", versionVar, ver),
		"-o", out, "./cmd/dtvlt")
}

func runCmd(env []string, name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("%s %s: %w", name, strings.Join(args, " "), err)
	}
	return nil
}

func exeSuffix(goos string) string {
	if goos == "windows" {
		return ".exe"
	}
	return ""
}

// tarGz writes files (archive name -> source path) into a gzipped tarball.
func tarGz(dst string, files map[string]string) error {
	f, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer f.Close()
	gz := gzip.NewWriter(f)
	tw := tar.NewWriter(gz)

	names := make([]string, 0, len(files))
	for n := range files {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, name := range names {
		src, err := os.Open(files[name])
		if err != nil {
			return err
		}
		info, err := src.Stat()
		if err != nil {
			src.Close()
			return err
		}
		mode := int64(0o644)
		if name == "dtvlt" {
			mode = 0o755
		}
		hdr := &tar.Header{Name: name, Mode: mode, Size: info.Size(), ModTime: info.ModTime()}
		if err := tw.WriteHeader(hdr); err != nil {
			src.Close()
			return err
		}
		_, err = io.Copy(tw, src)
		src.Close()
		if err != nil {
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}
	return gz.Close()
}

func writeChecksums(dst string, files []string) error {
	var b strings.Builder
	for _, path := range files {
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(data)
		fmt.Fprintf(&b, "%s  %s\n", hex.EncodeToString(sum[:]), filepath.Base(path))
	}
	return os.WriteFile(dst, []byte(b.String()), 0o644)
}

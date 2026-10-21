// Package update fetches the newest dtvlt release from GitHub and verifies it
// against the release's SHA256SUMS before anything is installed.
//
// The checksum comes from the same release as the binary, so it catches
// corrupted or truncated downloads and tampering in transit, not a compromised
// release. Signed releases would close that gap later.
package update

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"runtime"
	"strings"
	"time"

	"github.com/fbreseghello/DataVault/internal/version"
)

const (
	// Repo is where releases are published.
	Repo = "fbreseghello/DataVault"

	maxBinary = 64 << 20 // no dtvlt build comes close; anything bigger is refused
	maxSums   = 64 << 10
)

// semverTag accepts plain release tags only (v1.2.3), never pre-releases.
var semverTag = regexp.MustCompile(`^v(0|[1-9]\d*)\.(0|[1-9]\d*)\.(0|[1-9]\d*)$`)

var (
	ErrNoAsset   = errors.New("this release has no build for your system")
	ErrNoRelease = errors.New("no public release found on github.com/" + Repo + " (none published yet, or the repository is private)")
	errNotFound  = errors.New("not found")
)

// Client talks to GitHub. The zero value uses the real endpoints.
type Client struct {
	// API is the GitHub API root; Download is the root for release files.
	API      string
	Download string
	HTTP     *http.Client
}

func (c *Client) api() string {
	if c.API != "" {
		return c.API
	}
	return "https://api.github.com"
}

func (c *Client) download() string {
	if c.Download != "" {
		return c.Download
	}
	return "https://github.com"
}

func (c *Client) http() *http.Client {
	if c.HTTP != nil {
		return c.HTTP
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

// Release is the newest published version.
type Release struct {
	Version string // without the leading v
	Tag     string
	URL     string // the release page
}

// Latest asks GitHub for the newest non-prerelease release.
func (c *Client) Latest(ctx context.Context) (Release, error) {
	url := fmt.Sprintf("%s/repos/%s/releases/latest", c.api(), Repo)
	resp, err := c.get(ctx, url, "application/vnd.github+json")
	if errors.Is(err, errNotFound) {
		// GitHub answers 404 both when nothing is published and when the repo is private.
		return Release{}, ErrNoRelease
	}
	if err != nil {
		return Release{}, err
	}
	defer resp.Body.Close()
	var body struct {
		TagName string `json:"tag_name"`
		HTMLURL string `json:"html_url"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&body); err != nil {
		return Release{}, fmt.Errorf("reading the release: %w", err)
	}
	if !semverTag.MatchString(body.TagName) {
		return Release{}, fmt.Errorf("unexpected release tag %q", body.TagName)
	}
	return Release{
		Version: strings.TrimPrefix(body.TagName, "v"),
		Tag:     body.TagName,
		URL:     fmt.Sprintf("https://github.com/%s/releases/tag/%s", Repo, body.TagName),
	}, nil
}

// AssetName is the release file for this OS and architecture, matching scripts/build.
func AssetName(ver string) string {
	name := fmt.Sprintf("dtvlt-%s-%s-%s", ver, runtime.GOOS, runtime.GOARCH)
	if runtime.GOOS == "windows" {
		return name + ".exe"
	}
	return name + ".tar.gz"
}

// Downloaded is a verified binary waiting to be installed.
type Downloaded struct {
	Path   string // a temporary file; Remove it when done
	Asset  string
	SHA256 string
	Size   int64
}

func (d Downloaded) Remove() { _ = os.Remove(d.Path) }

// Fetch downloads the release's binary for this system into dir, checks it
// against SHA256SUMS and, for Linux tarballs, unpacks the dtvlt executable.
// Nothing is left behind on failure.
func (c *Client) Fetch(ctx context.Context, rel Release, dir string) (Downloaded, error) {
	asset := AssetName(rel.Version)
	base := fmt.Sprintf("%s/%s/releases/download/%s/", c.download(), Repo, rel.Tag)

	want, err := c.checksum(ctx, base+"SHA256SUMS", asset)
	if err != nil {
		return Downloaded{}, err
	}

	resp, err := c.get(ctx, base+asset, "application/octet-stream")
	if err != nil {
		return Downloaded{}, err
	}
	defer resp.Body.Close()

	tmp, err := os.CreateTemp(dir, ".dtvlt-download-*")
	if err != nil {
		return Downloaded{}, err
	}
	fail := func(err error) (Downloaded, error) {
		tmp.Close()
		os.Remove(tmp.Name())
		return Downloaded{}, err
	}
	h := sha256.New()
	n, err := io.Copy(io.MultiWriter(tmp, h), io.LimitReader(resp.Body, maxBinary+1))
	if err != nil {
		return fail(fmt.Errorf("downloading %s: %w", asset, err))
	}
	if n > maxBinary {
		return fail(fmt.Errorf("%s is larger than expected", asset))
	}
	got := h.Sum(nil)
	if subtle.ConstantTimeCompare(got, want) != 1 {
		return fail(fmt.Errorf("checksum mismatch for %s: the download is corrupt or was altered", asset))
	}
	if err := tmp.Close(); err != nil {
		return fail(err)
	}

	d := Downloaded{Path: tmp.Name(), Asset: asset, SHA256: hex.EncodeToString(got), Size: n}
	if strings.HasSuffix(asset, ".tar.gz") {
		bin, err := extract(d.Path, dir)
		os.Remove(d.Path)
		if err != nil {
			return Downloaded{}, err
		}
		d.Path = bin
	}
	return d, nil
}

// checksum finds asset in a SHA256SUMS file ("<hex>  <name>" per line).
func (c *Client) checksum(ctx context.Context, url, asset string) ([]byte, error) {
	resp, err := c.get(ctx, url, "text/plain")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	sc := bufio.NewScanner(io.LimitReader(resp.Body, maxSums))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) == 2 && strings.TrimPrefix(fields[1], "*") == asset {
			sum, err := hex.DecodeString(fields[0])
			if err != nil || len(sum) != sha256.Size {
				return nil, fmt.Errorf("malformed checksum for %s", asset)
			}
			return sum, nil
		}
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return nil, ErrNoAsset
}

// extract copies the dtvlt entry of a release tarball into a new file in dir.
func extract(tarball, dir string) (string, error) {
	f, err := os.Open(tarball)
	if err != nil {
		return "", err
	}
	defer f.Close()
	gz, err := gzip.NewReader(f)
	if err != nil {
		return "", err
	}
	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if errors.Is(err, io.EOF) {
			return "", errors.New("the release archive has no dtvlt binary")
		}
		if err != nil {
			return "", err
		}
		if hdr.Name != "dtvlt" || hdr.Typeflag != tar.TypeReg {
			continue
		}
		out, err := os.CreateTemp(dir, ".dtvlt-bin-*")
		if err != nil {
			return "", err
		}
		if _, err := io.Copy(out, io.LimitReader(tr, maxBinary)); err != nil {
			out.Close()
			os.Remove(out.Name())
			return "", err
		}
		if err := out.Close(); err != nil {
			os.Remove(out.Name())
			return "", err
		}
		if err := os.Chmod(out.Name(), 0o755); err != nil {
			os.Remove(out.Name())
			return "", err
		}
		return out.Name(), nil
	}
}

func (c *Client) get(ctx context.Context, url, accept string) (*http.Response, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", accept)
	req.Header.Set("User-Agent", "dtvlt/"+version.Version)
	resp, err := c.http().Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		if resp.StatusCode == http.StatusNotFound {
			return nil, fmt.Errorf("%w: %s", errNotFound, url)
		}
		return nil, fmt.Errorf("%s: %s", url, resp.Status)
	}
	return resp, nil
}

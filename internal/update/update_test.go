package update

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

const fakeBinary = "#!dtvlt test binary"

// asset returns what the release would host for this OS: a bare binary on
// Windows, a tarball elsewhere.
func asset(t *testing.T, name string) []byte {
	if !strings.HasSuffix(name, ".tar.gz") {
		return []byte(fakeBinary)
	}
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, f := range []struct{ name, body string }{{"LICENSE", "MIT"}, {"dtvlt", fakeBinary}} {
		tw.WriteHeader(&tar.Header{Name: f.name, Mode: 0o755, Size: int64(len(f.body)), Typeflag: tar.TypeReg})
		tw.Write([]byte(f.body))
	}
	tw.Close()
	gz.Close()
	return buf.Bytes()
}

// fakeGitHub serves a latest release with the given tag; sums overrides SHA256SUMS.
func fakeGitHub(t *testing.T, tag string, sums func(name string, body []byte) string) *Client {
	ver := strings.TrimPrefix(tag, "v")
	name := AssetName(ver)
	body := asset(t, name)
	if sums == nil {
		sums = func(name string, body []byte) string {
			s := sha256.Sum256(body)
			return fmt.Sprintf("%s  other-file.exe\n%s  %s\n", strings.Repeat("0", 64), hex.EncodeToString(s[:]), name)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/repos/"+Repo+"/releases/latest", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, `{"tag_name":%q,"html_url":"ignored"}`, tag)
	})
	dl := "/" + Repo + "/releases/download/" + tag + "/"
	mux.HandleFunc(dl+"SHA256SUMS", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, sums(name, body))
	})
	mux.HandleFunc(dl+name, func(w http.ResponseWriter, r *http.Request) { w.Write(body) })
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return &Client{API: srv.URL, Download: srv.URL, HTTP: srv.Client()}
}

func TestFetchVerifiesAndUnpacks(t *testing.T) {
	c := fakeGitHub(t, "v9.8.7", nil)
	rel, err := c.Latest(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if rel.Version != "9.8.7" {
		t.Fatalf("version = %q", rel.Version)
	}
	d, err := c.Fetch(context.Background(), rel, t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	defer d.Remove()
	got, _ := os.ReadFile(d.Path)
	if string(got) != fakeBinary {
		t.Fatalf("installed content = %q", got)
	}
}

func TestChecksumMismatchIsRejected(t *testing.T) {
	c := fakeGitHub(t, "v1.0.0", func(name string, body []byte) string {
		return strings.Repeat("ab", 32) + "  " + name + "\n"
	})
	dir := t.TempDir()
	rel, _ := c.Latest(context.Background())
	if _, err := c.Fetch(context.Background(), rel, dir); err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("got %v, want checksum mismatch", err)
	}
	if left, _ := os.ReadDir(dir); len(left) != 0 {
		t.Fatalf("download left behind: %v", left)
	}
}

func TestMissingAsset(t *testing.T) {
	c := fakeGitHub(t, "v1.0.0", func(string, []byte) string { return "" })
	rel, _ := c.Latest(context.Background())
	if _, err := c.Fetch(context.Background(), rel, t.TempDir()); err != ErrNoAsset {
		t.Fatalf("got %v, want ErrNoAsset", err)
	}
}

func TestOddTagsAreRefused(t *testing.T) {
	for _, tag := range []string{"v1.0.0-rc.1", "1.0.0", "v1.0", "../../evil"} {
		c := fakeGitHub(t, tag, nil)
		if _, err := c.Latest(context.Background()); err == nil {
			t.Errorf("tag %q accepted", tag)
		}
	}
}

func TestNoPublicRelease(t *testing.T) {
	srv := httptest.NewServer(http.NotFoundHandler())
	defer srv.Close()
	c := &Client{API: srv.URL, Download: srv.URL, HTTP: srv.Client()}
	if _, err := c.Latest(context.Background()); err != ErrNoRelease {
		t.Fatalf("got %v, want ErrNoRelease", err)
	}
}

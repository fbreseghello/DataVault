package server

import (
	"encoding/json"
	"net/http"
	"net/http/cookiejar"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/session"
	"github.com/fbreseghello/DataVault/internal/vault"
)

func startServer(t *testing.T) *Server {
	t.Helper()
	path := filepath.Join(t.TempDir(), "vault.dtv")
	v, err := vault.Create(path, []byte("kw"), []byte("pw"), vault.KDFParams{MemoryKiB: 64, Iterations: 1, Threads: 1})
	if err != nil {
		t.Fatal(err)
	}
	v.Lock()
	sess := session.New(path, time.Hour, nil)
	s, err := New(Options{
		Session: sess,
		Config:  config.Default(),
		Assets:  fstest.MapFS{"index.html": {Data: []byte("<!doctype html>ok")}},
	})
	if err != nil {
		t.Fatal(err)
	}
	go s.Serve()
	t.Cleanup(func() { s.ln.Close() })
	return s
}

// browser returns a client with its own cookie jar, like a separate browser profile.
func browser(t *testing.T) *http.Client {
	jar, _ := cookiejar.New(nil)
	return &http.Client{Jar: jar, Timeout: 5 * time.Second}
}

func signIn(t *testing.T, s *Server, c *http.Client) {
	t.Helper()
	resp, err := c.Get(s.LaunchURL())
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("launch: %s", resp.Status)
	}
}

func post(t *testing.T, s *Server, c *http.Client, path, body string, headers map[string]string) *http.Response {
	t.Helper()
	req, _ := http.NewRequest(http.MethodPost, s.URL()+path, strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	return resp
}

var sameOrigin = map[string]string{"X-Dtvlt": "1"}

func TestNoCookieNoAccess(t *testing.T) {
	s := startServer(t)
	for _, p := range []string{"/", "/api/session"} {
		resp, err := http.Get(s.URL() + p)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusUnauthorized {
			t.Errorf("%s without cookie: %s", p, resp.Status)
		}
	}
}

func TestLaunchTokenIsSingleUse(t *testing.T) {
	s := startServer(t)
	url := s.LaunchURL()
	c := browser(t)
	resp, _ := c.Get(url)
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("first use: %s", resp.Status)
	}
	resp, _ = browser(t).Get(url)
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("second use: %s", resp.Status)
	}
}

func TestHostHeaderIsChecked(t *testing.T) {
	s := startServer(t)
	c := browser(t)
	signIn(t, s, c)
	req, _ := http.NewRequest(http.MethodGet, s.URL()+"/api/session", nil)
	req.Host = "evil.example:" + strings.Split(s.host, ":")[1]
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("rebinding host: %s", resp.Status)
	}
}

func TestMutationsNeedSameOrigin(t *testing.T) {
	s := startServer(t)
	c := browser(t)
	signIn(t, s, c)
	if r := post(t, s, c, "/api/lock", "", nil); r.StatusCode != http.StatusForbidden {
		t.Errorf("no custom header: %s", r.Status)
	}
	if r := post(t, s, c, "/api/lock", "", map[string]string{"X-Dtvlt": "1", "Origin": "http://evil.example"}); r.StatusCode != http.StatusForbidden {
		t.Errorf("foreign origin: %s", r.Status)
	}
	if r := post(t, s, c, "/api/lock", "", map[string]string{"X-Dtvlt": "1", "Origin": s.URL()}); r.StatusCode != http.StatusNoContent {
		t.Errorf("same origin: %s", r.Status)
	}
}

func TestUnlockAndItems(t *testing.T) {
	s := startServer(t)
	c := browser(t)
	signIn(t, s, c)

	resp, _ := c.Get(s.URL() + "/api/items")
	resp.Body.Close()
	if resp.StatusCode != http.StatusLocked {
		t.Fatalf("items while locked: %s", resp.Status)
	}
	if r := post(t, s, c, "/api/unlock", `{"keyword":"kw","password":"nope"}`, sameOrigin); r.StatusCode != http.StatusUnauthorized {
		t.Fatalf("wrong password: %s", r.Status)
	}
	if r := post(t, s, c, "/api/unlock", `{"keyword":"kw","password":"pw"}`, sameOrigin); r.StatusCode != http.StatusNoContent {
		t.Fatalf("unlock: %s", r.Status)
	}
	if r := post(t, s, c, "/api/items", `{"type":"login","name":"github","fields":[{"name":"password","value":"x","secret":true}]}`, sameOrigin); r.StatusCode != http.StatusCreated {
		t.Fatalf("create: %s", r.Status)
	}
	resp, _ = c.Get(s.URL() + "/api/items")
	var list []vault.Summary
	json.NewDecoder(resp.Body).Decode(&list)
	resp.Body.Close()
	if len(list) != 1 || list[0].Name != "github" {
		t.Fatalf("list: %+v", list)
	}
}

func TestControlNeedsToken(t *testing.T) {
	s := startServer(t)
	req, _ := http.NewRequest(http.MethodGet, s.URL()+"/control/status", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusUnauthorized {
		t.Fatalf("no token: %s", resp.Status)
	}
	c := &Client{st: State{Port: s.Port(), ControlToken: s.ControlToken()}, http: http.DefaultClient}
	st, err := c.Status()
	if err != nil || st.SessionID == "" {
		t.Fatalf("status: %+v %v", st, err)
	}
}

func TestPrintableStripsEscapes(t *testing.T) {
	if got := printable("evil\x1b]0;pwned\x07name"); got != "evil]0;pwnedname" {
		t.Fatalf("got %q", got)
	}
}

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

func TestSettings(t *testing.T) {
	s := startServer(t)
	s.opts.ConfigPath = filepath.Join(t.TempDir(), "config.toml")
	c := browser(t)
	signIn(t, s, c)

	put := func(body string) *http.Response {
		req, _ := http.NewRequest(http.MethodPut, s.URL()+"/api/settings", strings.NewReader(body))
		req.Header.Set("X-Dtvlt", "1")
		resp, err := c.Do(req)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		return resp
	}
	valid := `{"pin_enabled":false,"keep_terminal":false,"auto_lock_minutes":5,"clipboard_clear_seconds":30}`

	if r := put(valid); r.StatusCode != http.StatusLocked {
		t.Fatalf("while locked: %s", r.Status)
	}
	post(t, s, c, "/api/unlock", `{"keyword":"kw","password":"pw"}`, sameOrigin)

	if r := put(`{"pin_enabled":true,"keep_terminal":true,"auto_lock_minutes":5,"clipboard_clear_seconds":30}`); r.StatusCode != http.StatusBadRequest {
		t.Errorf("pin on without a pin: %s", r.Status)
	}
	if r := put(`{"pin_enabled":true,"pin":"12","keep_terminal":true,"auto_lock_minutes":5,"clipboard_clear_seconds":30}`); r.StatusCode != http.StatusBadRequest {
		t.Errorf("short pin: %s", r.Status)
	}
	if r := put(`{"pin_enabled":false,"keep_terminal":true,"auto_lock_minutes":0,"clipboard_clear_seconds":30}`); r.StatusCode != http.StatusBadRequest {
		t.Errorf("invalid auto-lock: %s", r.Status)
	}
	if r := put(valid); r.StatusCode != http.StatusOK {
		t.Fatalf("valid settings: %s", r.Status)
	}
	saved, err := config.Load(s.opts.ConfigPath)
	if err != nil {
		t.Fatal(err)
	}
	if saved.AutoLockMinutes != 5 || saved.ClipboardClearSeconds != 30 || saved.KeepTerminal {
		t.Fatalf("saved: %+v", saved)
	}
	if s.config().ClipboardClearSeconds != 30 {
		t.Fatal("settings not applied to the running server")
	}

	if r := put(`{"pin_enabled":true,"pin":"4321","keep_terminal":false,"auto_lock_minutes":5,"clipboard_clear_seconds":30}`); r.StatusCode != http.StatusOK {
		t.Fatalf("set pin: %s", r.Status)
	}
	saved, _ = config.Load(s.opts.ConfigPath)
	if !config.VerifyPIN(saved.PinHash, []byte("4321")) {
		t.Fatal("pin not saved")
	}
}

// call sends a same-origin request and decodes a JSON reply into out (if given).
func call(t *testing.T, s *Server, c *http.Client, method, path, body string, headers map[string]string, out any) int {
	t.Helper()
	req, _ := http.NewRequest(method, s.URL()+path, strings.NewReader(body))
	req.Header.Set("X-Dtvlt", "1")
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if out != nil {
		json.NewDecoder(resp.Body).Decode(out)
	}
	return resp.StatusCode
}

func TestSealedItems(t *testing.T) {
	s := startServer(t)
	c := browser(t)
	signIn(t, s, c)
	call(t, s, c, "POST", "/api/unlock", `{"keyword":"kw","password":"pw"}`, nil, nil)

	sealed := `{"type":"login","name":"bank","sealed":true,"fields":[{"name":"password","value":"hunter2","secret":true}]}`
	if code := call(t, s, c, "POST", "/api/items", sealed, nil, nil); code != http.StatusBadRequest {
		t.Fatalf("sealing without a seal pin: %d", code)
	}
	if code := call(t, s, c, "PUT", "/api/seal", `{"pin":"1234"}`, nil, nil); code != http.StatusNoContent {
		t.Fatalf("set seal pin: %d", code)
	}
	var item vault.Item
	if code := call(t, s, c, "POST", "/api/items", sealed, nil, &item); code != http.StatusCreated {
		t.Fatalf("create sealed: %d", code)
	}
	path := "/api/items/" + item.ID

	var list []vault.Summary
	call(t, s, c, "GET", "/api/items", "", nil, &list)
	if len(list) != 1 || !list[0].Sealed {
		t.Fatalf("list: %+v", list)
	}

	var got map[string]any
	if code := call(t, s, c, "GET", path, "", nil, &got); code != http.StatusForbidden || got["fields"] != nil {
		t.Fatalf("sealed item leaked without pin: %d %v", code, got)
	}
	if code := call(t, s, c, "POST", path+"/unseal", `{"pin":"0000"}`, nil, nil); code != http.StatusForbidden {
		t.Fatalf("wrong pin: %d", code)
	}
	var opened struct {
		Item  vault.Item `json:"item"`
		Token string     `json:"token"`
	}
	if code := call(t, s, c, "POST", path+"/unseal", `{"pin":"1234"}`, nil, &opened); code != http.StatusOK || opened.Token == "" {
		t.Fatalf("unseal: %d %+v", code, opened)
	}
	if opened.Item.Fields[0].Value != "hunter2" {
		t.Fatalf("unsealed item: %+v", opened.Item)
	}
	grant := map[string]string{unsealHeader: opened.Token}

	if code := call(t, s, c, "GET", path, "", grant, nil); code != http.StatusOK {
		t.Fatalf("get with token: %d", code)
	}
	edit := `{"type":"login","name":"bank","sealed":false,"fields":[]}`
	if code := call(t, s, c, "PUT", path, edit, nil, nil); code != http.StatusForbidden {
		t.Fatalf("unsealing by edit without token: %d", code)
	}
	if code := call(t, s, c, "DELETE", path, "", nil, nil); code != http.StatusForbidden {
		t.Fatalf("delete without token: %d", code)
	}
	if code := call(t, s, c, "PUT", path, sealed, grant, nil); code != http.StatusOK {
		t.Fatalf("edit with token: %d", code)
	}

	// The token dies with the lock.
	call(t, s, c, "POST", "/api/lock", "", nil, nil)
	call(t, s, c, "POST", "/api/unlock", `{"keyword":"kw","password":"pw"}`, nil, nil)
	if code := call(t, s, c, "GET", path, "", grant, nil); code != http.StatusForbidden {
		t.Fatalf("token survived a lock: %d", code)
	}

	if code := call(t, s, c, "PUT", "/api/seal", `{"current":"9999","pin":"5678"}`, nil, nil); code != http.StatusForbidden {
		t.Fatalf("changing the seal pin without the current one: %d", code)
	}
	if code := call(t, s, c, "PUT", "/api/seal", `{"current":"1234","pin":"5678"}`, nil, nil); code != http.StatusNoContent {
		t.Fatalf("change seal pin: %d", code)
	}
}

func TestNotes(t *testing.T) {
	s := startServer(t)
	c := browser(t)
	signIn(t, s, c)
	call(t, s, c, "POST", "/api/unlock", `{"keyword":"kw","password":"pw"}`, nil, nil)

	var open vault.Note
	if code := call(t, s, c, "POST", "/api/notes", `{"title":"t","body":"plain words"}`, nil, &open); code != http.StatusCreated {
		t.Fatalf("create: %d", code)
	}
	if code := call(t, s, c, "POST", "/api/notes", `{"body":"x","sealed":true}`, nil, nil); code != http.StatusBadRequest {
		t.Fatalf("sealed note without a seal pin: %d", code)
	}
	call(t, s, c, "PUT", "/api/seal", `{"pin":"1234"}`, nil, nil)
	var sealed vault.Note
	call(t, s, c, "POST", "/api/notes", `{"title":"hidden","body":"what hurts","sealed":true}`, nil, &sealed)

	var list []vault.NoteSummary
	call(t, s, c, "GET", "/api/notes", "", nil, &list)
	if len(list) != 2 {
		t.Fatalf("list: %+v", list)
	}
	for _, n := range list {
		if n.Sealed && (n.Title != "" || n.Preview != "") {
			t.Fatalf("sealed note leaked in the list: %+v", n)
		}
	}

	path := "/api/notes/" + sealed.ID
	if code := call(t, s, c, "GET", path, "", nil, nil); code != http.StatusForbidden {
		t.Fatalf("sealed note without pin: %d", code)
	}
	if code := call(t, s, c, "PUT", path, `{"body":"overwritten"}`, nil, nil); code != http.StatusForbidden {
		t.Fatalf("editing a sealed note without pin: %d", code)
	}
	var opened struct {
		Note  vault.Note `json:"note"`
		Token string     `json:"token"`
	}
	if code := call(t, s, c, "POST", path+"/unseal", `{"pin":"1234"}`, nil, &opened); code != http.StatusOK || opened.Note.Body != "what hurts" {
		t.Fatalf("unseal: %d %+v", code, opened)
	}
	grant := map[string]string{unsealHeader: opened.Token}
	if code := call(t, s, c, "GET", path, "", grant, nil); code != http.StatusOK {
		t.Fatalf("get with token: %d", code)
	}

	// Only buried notes can be deleted.
	if code := call(t, s, c, "DELETE", "/api/notes/"+open.ID, "", nil, nil); code != http.StatusConflict {
		t.Fatalf("deleting an unburied note: %d", code)
	}
	if code := call(t, s, c, "POST", "/api/notes/"+open.ID+"/bury", `{"buried":true}`, nil, nil); code != http.StatusOK {
		t.Fatalf("bury: %d", code)
	}
	var dead []vault.NoteSummary
	call(t, s, c, "GET", "/api/notes?buried=1", "", nil, &dead)
	if len(dead) != 1 || dead[0].ID != open.ID {
		t.Fatalf("dead words: %+v", dead)
	}
	if code := call(t, s, c, "DELETE", "/api/notes/"+open.ID, "", nil, nil); code != http.StatusNoContent {
		t.Fatalf("delete buried: %d", code)
	}
}

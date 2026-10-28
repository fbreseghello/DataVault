package server

import (
	"fmt"
	"net/http"
	"path/filepath"
	"testing"

	"github.com/fbreseghello/DataVault/internal/config"
	"github.com/fbreseghello/DataVault/internal/theme"
)

func TestThemes(t *testing.T) {
	s := startServer(t)
	s.opts.ConfigPath = filepath.Join(t.TempDir(), "config.toml")
	c := browser(t)
	signIn(t, s, c)

	var session struct {
		Theme theme.Theme `json:"theme"`
	}
	if code := call(t, s, c, http.MethodGet, "/api/session", "", nil, &session); code != http.StatusOK || session.Theme.ID != theme.Default {
		t.Fatalf("the login should draw the default theme: %d %q", code, session.Theme.ID)
	}
	if code := call(t, s, c, http.MethodGet, "/api/themes", "", nil, nil); code != http.StatusLocked {
		t.Fatalf("listing while locked: %d", code)
	}
	post(t, s, c, "/api/unlock", `{"keyword":"kw","password":"pw"}`, sameOrigin)

	mine := `{"id":"ember","name":"ember","colors":{"accent":"#ff6600"},"words":{"grimoire":"keys"}}`
	if code := call(t, s, c, http.MethodPost, "/api/themes", `{"id":"souls","name":"x"}`, nil, nil); code != http.StatusBadRequest {
		t.Errorf("importing over a built-in theme: %d", code)
	}
	if code := call(t, s, c, http.MethodPost, "/api/themes", `{"id":"ember","name":"ember","colors":{"accent":"orange"}}`, nil, nil); code != http.StatusBadRequest {
		t.Errorf("a color that is not hex: %d", code)
	}
	if code := call(t, s, c, http.MethodPost, "/api/themes", mine, nil, nil); code != http.StatusOK {
		t.Fatalf("import: %d", code)
	}

	var list []theme.Summary
	call(t, s, c, http.MethodGet, "/api/themes", "", nil, &list)
	if len(list) < 2 || list[0].ID != theme.Default || list[len(list)-1].ID != "ember" || list[len(list)-1].Builtin {
		t.Fatalf("list: %+v", list)
	}

	settings := `{"pin_enabled":false,"keep_terminal":true,"auto_lock_minutes":5,"clipboard_clear_seconds":30,"theme":"%s"}`
	if code := call(t, s, c, http.MethodPut, "/api/settings", fmt.Sprintf(settings, "nope"), nil, nil); code != http.StatusBadRequest {
		t.Errorf("choosing a theme that does not exist: %d", code)
	}
	if code := call(t, s, c, http.MethodPut, "/api/settings", fmt.Sprintf(settings, "ember"), nil, nil); code != http.StatusOK {
		t.Fatalf("choosing the imported theme: %d", code)
	}
	call(t, s, c, http.MethodGet, "/api/session", "", nil, &session)
	if got := session.Theme; got.ID != "ember" || got.Colors["accent"] != "#ff6600" || got.Words["grimoire"] != "keys" || got.Words["siege"] != "siege" || len(got.Icons) != len(theme.Icons) {
		t.Fatalf("the session should draw ember over souls: %+v", got)
	}

	var exported theme.Theme
	call(t, s, c, http.MethodGet, "/api/themes/ember", "", nil, &exported)
	if len(exported.Words) != len(theme.Words) || len(exported.Login) == 0 {
		t.Fatalf("an export should be complete: %+v", exported)
	}

	if code := call(t, s, c, http.MethodDelete, "/api/themes/souls", "", nil, nil); code != http.StatusBadRequest {
		t.Errorf("removing a built-in theme: %d", code)
	}
	if code := call(t, s, c, http.MethodDelete, "/api/themes/ember", "", nil, nil); code != http.StatusNoContent {
		t.Fatalf("remove: %d", code)
	}
	if saved, _ := config.Load(s.opts.ConfigPath); saved.Theme != theme.Default || s.config().Theme != theme.Default {
		t.Fatalf("removing the theme in use should go back to the default, got %q", saved.Theme)
	}
}

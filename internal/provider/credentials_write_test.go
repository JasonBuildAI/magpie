package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/yetone/magpie/internal/steady"
)

// codexSignInSoon is a Codex sign-in whose access token expires in four
// minutes: the next use of it refreshes, as codexToken does within five.
func codexSignInSoon(t *testing.T, home, refresh string) string {
	t.Helper()
	path := filepath.Join(home, ".codex", "auth.json")
	writeFile(t, path, map[string]any{
		"auth_mode": "chatgpt",
		"tokens": map[string]any{
			"id_token":      fakeJWT(map[string]any{"email": "work@example.com", "https://api.openai.com/auth": map[string]any{"chatgpt_plan_type": "plus"}}),
			"access_token":  fakeJWT(map[string]any{"exp": float64(time.Now().Add(4 * time.Minute).Unix())}),
			"refresh_token": refresh, "account_id": "acct-work@example.com",
		},
	})
	return path
}

// codexRefreshAnswers points codexTokenURL at an endpoint that rotates the
// tokens as OpenAI's does, naming the one it was given.
func codexRefreshAnswers(t *testing.T) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body map[string]string
		json.NewDecoder(r.Body).Decode(&body)
		json.NewEncoder(w).Encode(map[string]any{
			"access_token":  fakeJWT(map[string]any{"exp": float64(time.Now().Add(time.Hour).Unix())}),
			"refresh_token": body["refresh_token"] + "-2",
		})
	}))
	t.Cleanup(srv.Close)
	old := codexTokenURL
	codexTokenURL = srv.URL
	t.Cleanup(func() { codexTokenURL = old })
}

// A refresh writes Codex's own auth.json the way every other write of a
// sign-in goes: into a temp file flushed to the disk before it is renamed
// over the old one (steady.WriteFile, #1505). The old file is never opened
// for writing, so it is never truncated: a machine that goes down mid-write
// leaves the user's Codex sign-in whole, for Codex CLI, the Codex app and
// every IDE extension that reads the same file — and a refresh token that
// OpenAI has already rotated past is never dropped in favour of nothing.
func TestCodexRefreshFlushesAuthJSONBeforeReplacingIt(t *testing.T) {
	home := signIn(t)
	auth := codexSignInSoon(t, home, "r-work")
	before, err := os.ReadFile(auth)
	if err != nil {
		t.Fatal(err)
	}
	codexRefreshAnswers(t)

	flushed := false
	oldSync := steady.Sync
	steady.Sync = func(f *os.File) error {
		flushed = true
		if b, _ := os.ReadFile(auth); !bytes.Equal(b, before) {
			t.Errorf("auth.json was replaced before the new sign-in reached the disk: %s", b)
		}
		return errors.New("the disk went away")
	}
	t.Cleanup(func() { steady.Sync = oldSync })

	if _, _, err := codexToken(context.Background(), auth); err != nil {
		t.Fatalf("a refresh whose write failed: %v", err)
	}
	if !flushed {
		t.Error("auth.json was written without the new sign-in being flushed to the disk first")
	}
	if after, err := os.ReadFile(auth); err != nil || !bytes.Equal(after, before) {
		t.Errorf("a flush that failed left auth.json %q (%v), want the sign-in as it was", after, err)
	}
}

// The same for Claude Code's own credential file, which magpie writes when
// it switches the sign-in Claude Code is on (saveClaudeCredential): the
// file is flushed before it is renamed over, so an interrupted write leaves
// Claude Code signed in to what it was signed in to.
func TestClaudeCredentialFlushesBeforeReplacing(t *testing.T) {
	home := signIn(t)
	path := filepath.Join(home, ".claude", ".credentials.json")
	before := []byte("{\n  \"claudeAiOauth\": {\n    \"accessToken\": \"old-access\",\n    \"refreshToken\": \"old-refresh\",\n    \"expiresAt\": 1\n  }\n}\n")
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, before, 0o600); err != nil {
		t.Fatal(err)
	}
	c, ok := parseClaudeCredentials(before)
	if !ok {
		t.Fatal("the fixture is not a Claude sign-in")
	}
	c.OAuth.AccessToken, c.OAuth.RefreshToken = "new-access", "new-refresh"

	flushed := false
	oldSync := steady.Sync
	steady.Sync = func(f *os.File) error {
		flushed = true
		if b, _ := os.ReadFile(path); !bytes.Equal(b, before) {
			t.Errorf("the credentials file was replaced before the new sign-in reached the disk: %s", b)
		}
		return errors.New("the disk went away")
	}
	t.Cleanup(func() { steady.Sync = oldSync })

	if err := writeClaudeCredential(claudeCredentialLocation{path: path}, c); err == nil {
		t.Error("a write whose flush failed reported as saved")
	}
	if !flushed {
		t.Error("the credentials file was written without the new sign-in being flushed to the disk first")
	}
	if after, err := os.ReadFile(path); err != nil || !bytes.Equal(after, before) {
		t.Errorf("a flush that failed left the credentials file %q (%v), want it as it was", after, err)
	}
}

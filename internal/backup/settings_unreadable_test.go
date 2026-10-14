package backup

import (
	"os"
	"testing"

	"github.com/yetone/magpie/internal/settings"
)

// A settings.json that can't be read is unknown, never the defaults: a
// backup or a sync carrying the defaults would bring them in on every
// machine the bundle is put back on, taking the user's theme, language, text
// size and allow list away. Collect stops on it, as it does on a
// providers.json that can't be read.
func TestCollectStopsOnUnreadableSettings(t *testing.T) {
	home(t)
	if err := os.MkdirAll(settings.Dir(), 0o755); err != nil {
		t.Fatal(err)
	}
	whole := `{"theme":"dark","lang":"zh","port":3999,"lanKey":"a-real-key"}`
	if err := os.WriteFile(settings.Path(), []byte(whole[:len(whole)/2]), 0o600); err != nil {
		t.Fatal(err)
	}
	if s := settings.Load(); s.LANKey == "a-real-key" {
		t.Fatal("misread the broken settings.json as a good one")
	}
	b, err := Collect(false, "test")
	if err == nil {
		t.Fatalf("Collect read a settings.json that can't be read as settings: %+v", b.Settings)
	}
	if b.Settings != nil {
		t.Fatalf("a backup carrying settings from a file that couldn't be read: %+v", b.Settings)
	}
}

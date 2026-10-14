package sessions

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"testing"
)

// linkFolder makes path a link to the folder target: on Windows a junction,
// which is what the Codex CLI itself makes of ~/.codex/sessions when .codex
// is kept on another drive, and a symlink elsewhere.
func linkFolder(t *testing.T, target, path string) {
	t.Helper()
	if runtime.GOOS == "windows" {
		if out, err := exec.Command("cmd", "/c", "mklink", "/J", path, target).CombinedOutput(); err != nil {
			t.Skipf("mklink /J: %v %s", err, out)
		}
		return
	}
	if err := os.Symlink(target, path); err != nil {
		t.Skipf("symlink: %v", err)
	}
}

// A Codex folder whose sessions folder is a link is read all the same: the
// Codex CLI makes ~/.codex/sessions a junction to the drive .codex is kept
// on, and the sessions behind it were not listed at all (#2233).
func TestCodexRolloutsBehindALink(t *testing.T) {
	_, codex := setup(t)
	rollouts := codexFiles()
	if len(rollouts) == 0 {
		t.Fatal("the fixture has no rollouts")
	}
	sessions := len(ListAgent("codex"))
	if sessions == 0 {
		t.Fatal("the fixture has no Codex session")
	}

	live := filepath.Join(codex, "sessions")
	elsewhere := filepath.Join(t.TempDir(), "sessions")
	if err := os.Rename(live, elsewhere); err != nil {
		t.Fatal(err)
	}
	linkFolder(t, elsewhere, live)

	if got := codexFiles(); len(got) != len(rollouts) {
		t.Fatalf("rollouts behind a link: %d, want %d", len(got), len(rollouts))
	}
	if got := len(ListAgent("codex")); got != sessions {
		t.Fatalf("Codex sessions behind a link: %d, want %d", got, sessions)
	}
}

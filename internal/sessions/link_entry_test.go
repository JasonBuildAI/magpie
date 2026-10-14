package sessions

import (
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
)

// ccFilePaths is every session file ccFiles found, by path: one that stops
// being listed, or is listed from somewhere else, stands out.
func ccFilePaths(fs []file) []string {
	out := make([]string, 0, len(fs))
	for _, f := range fs {
		out = append(out, f.path)
	}
	sort.Strings(out)
	return out
}

// A Claude Code project folder inside projects/ that is a link is read all
// the same. It is the shape of a heavy project's sessions kept on another
// drive with a junction left where they were, the layout this machine's
// ~/.codex already has 22 of. Go reports a junction as irregular, not as a
// symlink (its lstat is not a directory either), so ccFiles skipped the
// folder whole: Sessions listed none of the project's sessions and Usage
// none of its calls (#2233's shape inside a store, the case #2250 left out).
func TestClaudeProjectFolderBehindALink(t *testing.T) {
	claude, _ := setup(t)
	before := ccFiles("claude", claude)
	if len(before) == 0 {
		t.Fatal("the fixture has no Claude session files")
	}
	listed := len(ListAgent("claude"))
	live := filepath.Join(claude, "projects", "-work-app")
	elsewhere := filepath.Join(t.TempDir(), "-work-app")
	if err := os.Rename(live, elsewhere); err != nil {
		t.Fatal(err)
	}
	linkFolder(t, elsewhere, live)

	after := ccFiles("claude", claude)
	if !reflect.DeepEqual(ccFilePaths(after), ccFilePaths(before)) {
		t.Fatalf("Claude session files behind a link: %v, want %v", ccFilePaths(after), ccFilePaths(before))
	}
	if got := len(ListAgent("claude")); got != listed {
		t.Fatalf("Claude sessions behind a link: %d, want %d", got, listed)
	}
}

// The session's own folder — where its subagents and their workflows live —
// being a link is read too.
func TestClaudeSessionFolderBehindALink(t *testing.T) {
	claude, _ := setup(t)
	before := ccFiles("claude", claude)
	if len(before) == 0 {
		t.Fatal("the fixture has no Claude session files")
	}
	live := filepath.Join(claude, "projects", "-work-app", "11111111-2222-3333-4444-555555555555")
	elsewhere := filepath.Join(t.TempDir(), "11111111-2222-3333-4444-555555555555")
	if err := os.Rename(live, elsewhere); err != nil {
		t.Fatal(err)
	}
	linkFolder(t, elsewhere, live)

	after := ccFiles("claude", claude)
	if !reflect.DeepEqual(ccFilePaths(after), ccFilePaths(before)) {
		t.Fatalf("a Claude session folder behind a link: %v, want %v", ccFilePaths(after), ccFilePaths(before))
	}
}

// A Hermes profile folder that is a link is read too: HermesDirs skipped it
// the same way and the profile's sessions were gone from Sessions and Usage
// (#2250 named this check beside ccFiles as the uncovered pair).
func TestHermesProfileBehindALink(t *testing.T) {
	setup(t)
	root := filepath.Join(t.TempDir(), "hermes")
	t.Setenv("HERMES_HOME", root)
	hermesFixture(t, root, "", "link-default", true)
	hermesFixture(t, root, "work", "link-work", true)
	if got := hermesSessions(); got != 2 {
		t.Fatalf("Hermes sessions: %d, want 2", got)
	}
	live := filepath.Join(root, "profiles", "work")
	elsewhere := filepath.Join(t.TempDir(), "work")
	if err := os.Rename(live, elsewhere); err != nil {
		t.Fatal(err)
	}
	linkFolder(t, elsewhere, live)
	if got := hermesSessions(); got != 2 {
		t.Fatalf("Hermes sessions behind a link: %d, want 2", got)
	}
}

func hermesSessions() int {
	n := 0
	for _, s := range List(0) {
		if s.Agent == "hermes" {
			n++
		}
	}
	return n
}

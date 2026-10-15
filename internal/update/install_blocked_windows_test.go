package update

import (
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// The other install tests put text in for the new version, which Windows
// won't start: they are about moving files, so the try is left out.
func init() { canStart = func(string) error { return nil } }

// #894: a download Windows won't start isn't put in; the running exe
// stays where it is and the download goes.
func TestInstallBinaryRefusedStaysOut(t *testing.T) {
	was := canStart
	canStart = func(string) error { return &os.PathError{Op: "fork/exec", Err: errIntegrityPolicy} }
	t.Cleanup(func() { canStart = was })
	exe, staged := rollbackFiles(t)
	old, err := InstallBinaryAside(staged, exe)
	var be *BlockedError
	if !errors.As(err, &be) || !be.Policy() || old != "" {
		t.Fatalf("InstallBinaryAside = %q, %v", old, err)
	}
	wantInstallFile(t, exe, "old version")
	for _, p := range []string{staged, exe + ".old"} {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s: %v", filepath.Base(p), err)
		}
	}
}

// trialStart starts a real program, under the .new name a download has,
// without running it, and fails on what isn't one.
func TestTrialStart(t *testing.T) {
	ping := filepath.Join(os.Getenv("SystemRoot"), "System32", "PING.EXE")
	b, err := os.ReadFile(ping)
	if err != nil {
		t.Skip("no PING.EXE:", err)
	}
	dir := t.TempDir()
	good := filepath.Join(dir, "magpie.exe.new")
	os.WriteFile(good, b, 0o755)
	if err := trialStart(good); err != nil {
		t.Fatalf("a program: %v", err)
	}
	bad := filepath.Join(dir, "text.exe.new")
	os.WriteFile(bad, []byte("not a program"), 0o755)
	if err := trialStart(bad); err == nil {
		t.Fatal("text started")
	}
	// ended, not left running: its file can go
	if err := os.Remove(good); err != nil {
		t.Fatalf("still held: %v", err)
	}
}

// The restart for real: a new version that can't be started, one that
// quits at once and one that runs on without saying it is up are each
// ended, their file let go of, and the running exe put back.
func TestStartCheckedPutsBackForReal(t *testing.T) {
	ping := filepath.Join(os.Getenv("SystemRoot"), "System32", "PING.EXE")
	b, err := os.ReadFile(ping)
	if err != nil {
		t.Skip("no PING.EXE:", err)
	}
	was := startWait
	startWait = 2 * time.Second
	t.Cleanup(func() { startWait = was })
	for _, c := range []struct {
		name string
		prog []byte
		args []string
	}{
		{"not a program", []byte("not a program"), nil},
		{"quits at once", b, []string{"-n", "1", "127.0.0.1"}},
		{"never says it is up", b, []string{"-n", "60", "127.0.0.1"}},
	} {
		t.Run(c.name, func(t *testing.T) {
			dir := t.TempDir()
			exe, old := filepath.Join(dir, "magpie.exe"), filepath.Join(dir, "magpie.exe.old")
			os.WriteFile(exe, c.prog, 0o755)
			os.WriteFile(old, []byte("running version"), 0o755)
			err := startChecked(exe, old, c.args, os.Environ())
			var be *BlockedError
			if !errors.As(err, &be) {
				t.Fatalf("err = %v", err)
			}
			t.Log(err)
			wantInstallFile(t, exe, "running version")
			if _, err := os.Stat(old); !os.IsNotExist(err) {
				t.Errorf("old left: %v", err)
			}
		})
	}
}

// The file trialStart started stays held a moment afterwards — by
// whatever scans a program as it starts — and the install moves it in as
// soon as the trial ends: trialStart waits for it to be let go first, so
// the move doesn't fail with a sharing violation the trial itself caused
// (#2272). The file here is held like a scanner's until 150ms in, and the
// rename right after trialStart fails if it returned while still held.
func TestTrialStartWaitsForTheFile(t *testing.T) {
	ping := filepath.Join(os.Getenv("SystemRoot"), "System32", "PING.EXE")
	b, err := os.ReadFile(ping)
	if err != nil {
		t.Skip("no PING.EXE:", err)
	}
	good := filepath.Join(t.TempDir(), "magpie.exe.new")
	os.WriteFile(good, b, 0o755)
	letGo := holdForRollbackTest(t, good)
	released := make(chan struct{})
	go func() {
		time.Sleep(150 * time.Millisecond)
		letGo()
		close(released)
	}()
	if err := trialStart(good); err != nil {
		t.Fatalf("a program: %v", err)
	}
	if err := os.Rename(good, good+".moved"); err != nil {
		t.Fatalf("trialStart returned while the file was still held: %v", err)
	}
	<-released // the hold is let go before the test's folder goes
}

// The install right after the trial, the order InstallBinaryAside makes:
// a machine that holds the download a moment must not make the update
// fail (#2272).
func TestInstallRightAfterTheTrial(t *testing.T) {
	ping := filepath.Join(os.Getenv("SystemRoot"), "System32", "PING.EXE")
	b, err := os.ReadFile(ping)
	if err != nil {
		t.Skip("no PING.EXE:", err)
	}
	was := canStart
	canStart = trialStart
	t.Cleanup(func() { canStart = was })
	exe, staged := rollbackFiles(t)
	if err := os.WriteFile(staged, b, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := InstallBinary(staged, exe); err != nil {
		t.Fatalf("install after the trial: %v", err)
	}
	wantInstallFile(t, exe, string(b))
}

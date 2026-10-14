package proc

import (
	"path/filepath"
	"testing"
)

// Windows' PowerShell is named by its path in the system folder: by its
// bare name a PATHEXT without .EXE (WSL passed one on as ".CPL") found
// nothing, though it is there, and every PowerShell magpie starts failed
// with it (#2043, xiaobaifly7).
func TestPowerShellNamesItsPath(t *testing.T) {
	want := filepath.Join(`C:\WINDOWS`, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
	if got := powershellPath("windows", `C:\WINDOWS`); got != want {
		t.Fatalf("PowerShell %q, want %q", got, want)
	}
	// no SystemRoot to go by: the bare name, as before
	if got := powershellPath("windows", ""); got != "powershell" {
		t.Fatalf("without SystemRoot: %q", got)
	}
	for _, goos := range []string{"darwin", "linux", "android"} {
		if got := powershellPath(goos, `C:\WINDOWS`); got != "powershell" {
			t.Errorf("%s: %q", goos, got)
		}
	}
}

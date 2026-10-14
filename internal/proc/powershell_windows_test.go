package proc

import (
	"os/exec"
	"strings"
	"testing"
)

// A magpie that inherited a PATHEXT without .EXE (one WSL passed on as
// ".CPL", #2043) can't have the bare name found at all; the path
// PowerShell() gives runs where it doesn't.
func TestPowerShellRunsWhereABareNameWouldNot(t *testing.T) {
	t.Setenv("PATHEXT", ".CPL")
	err := exec.Command("powershell", "-NoProfile", "-Command", "exit 0").Run()
	if err == nil {
		t.Fatal("the bare name was found under PATHEXT=.CPL, so this proves nothing")
	}
	if !strings.Contains(err.Error(), "executable file not found") {
		t.Fatalf("the bare name failed with %v, not a lookup failure", err)
	}
	// PowerShell writes its module analysis cache under the current
	// directory when it has no LocalAppData to put it in, which a test's
	// sandbox is: keep it out of the source tree.
	cmd := Command(PowerShell(), "-NoProfile", "-NonInteractive", "-Command", "Write-Output magpie")
	cmd.Dir = t.TempDir()
	out, err := cmd.Output()
	if err != nil || strings.TrimSpace(string(out)) != "magpie" {
		t.Fatalf("PowerShell() by its path: %q, %v", out, err)
	}
}

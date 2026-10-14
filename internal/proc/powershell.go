package proc

import (
	"os"
	"path/filepath"
	"runtime"
)

// PowerShell is the program magpie hands its PowerShell one-liners to,
// named by its path in the system folder. By its bare name Go looks it up
// with PATHEXT's extensions, and a PATHEXT without .EXE (one WSL passed on
// as ".CPL", #2043) finds none, though it is there: resuming a session,
// installing a vendor's CLI and the WMI and Get-NetTCPConnection probes
// each failed that way on Windows. browserCommand names rundll32 so, for
// the same reason.
func PowerShell() string { return powershellPath(runtime.GOOS, os.Getenv("SystemRoot")) }

// powershellPath is PowerShell on goos: in the system folder when there is
// one to go by, the bare name (as before) otherwise, which only Windows'
// callers reach.
func powershellPath(goos, systemRoot string) string {
	if goos != "windows" || systemRoot == "" {
		return "powershell"
	}
	return filepath.Join(systemRoot, "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
}

package proc

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"strconv"
	"strings"

	"golang.org/x/sys/windows"
)

// listeningOn asks Get-NetTCPConnection, which names each listener's owner.
// The cmdlet exits 1 for none found, as lsof does, and that is no error: the
// gateway's take-over waits on this read for an older magpie to let the port
// go. Only an exit that wrote nothing on either stream, within the caller's
// deadline, is none found: a read that failed says why on stderr, and one cut
// short by the deadline writes nothing at all.
func listeningOn(ctx context.Context, port int) ([]int, error) {
	out, err := CommandContext(ctx, "powershell", "-NoProfile", "-NonInteractive", "-Command",
		`Get-NetTCPConnection -State Listen -LocalPort `+strconv.Itoa(port)+` -ErrorAction SilentlyContinue | ForEach-Object { $_.OwningProcess }`).Output()
	var ee *exec.ExitError
	if err != nil && !(errors.As(err, &ee) && len(bytes.TrimSpace(out)) == 0 && len(bytes.TrimSpace(ee.Stderr)) == 0 && ctx.Err() == nil) {
		return nil, err
	}
	var ps []int
	seen := map[int]bool{}
	for _, l := range strings.Fields(string(out)) {
		if n, err := strconv.Atoi(l); err == nil && n > 0 && !seen[n] {
			seen[n] = true
			ps = append(ps, n)
		}
	}
	return ps, nil
}

func executable(pid int) (string, error) {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return "", err
	}
	defer windows.CloseHandle(h)
	buf := make([]uint16, windows.MAX_LONG_PATH)
	n := uint32(len(buf))
	if err := windows.QueryFullProcessImageName(h, 0, &buf[0], &n); err != nil {
		return "", err
	}
	return windows.UTF16ToString(buf[:n]), nil
}

// terminate ends it at once: a tray app has no console for a Ctrl+Break,
// and no window a close message would reach.
func terminate(pid int) error { return kill(pid) }

func kill(pid int) error {
	h, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, uint32(pid))
	if err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return ErrDenied
		}
		if errors.Is(err, windows.ERROR_INVALID_PARAMETER) {
			return nil // gone already
		}
		return err
	}
	defer windows.CloseHandle(h)
	if err := windows.TerminateProcess(h, 1); err != nil {
		if errors.Is(err, windows.ERROR_ACCESS_DENIED) {
			return ErrDenied
		}
		return err
	}
	return nil
}

package update

import (
	"errors"
	"os"
	"os/exec"
	"syscall"
	"time"

	"github.com/yetone/magpie/internal/proc"
)

const (
	synchronize  = 0x00100000
	waitTimeout  = 0x00000102
	detachedProc = 0x00000008
	newProcGroup = 0x00000200
	suspended    = 0x00000004
	noWindow     = 0x08000000
	deleteAccess = 0x00010000 // DELETE: to open a file so it may be renamed
)

// alive reports whether pid is still running.
func alive(pid int) bool {
	h, err := syscall.OpenProcess(synchronize, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(h)
	ev, _ := syscall.WaitForSingleObject(h, 0)
	return ev == waitTimeout
}

// detach starts a relaunched magpie outside this one's console, so it
// outlives it. It gets no console at all, so Windows ignores the
// CREATE_NO_WINDOW proc may have set.
func detach(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= detachedProc | newProcGroup
}

// Reexec starts exe with args and env in this console and returns, for
// the caller to exit: Windows can't run another program in a process's
// place, so the new version goes on printing where this one did.
func Reexec(exe string, args, env []string) error {
	cmd := proc.Command(exe, args...)
	cmd.Env = env
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Start(); err != nil {
		return err
	}
	return cmd.Process.Release()
}

// trialStart starts path suspended and ends it at once: Windows checks a
// program when it creates its process, so one Smart App Control or
// another application control policy won't run fails here, and none of
// one it will run runs. A program's file stays held a moment after the
// process ends — whatever scans a program as it starts — and the caller
// moves it in right after (InstallBinaryAside): wait for it to be let go
// first, as moveAside waits out the same hold, so the trial doesn't make
// that move fail with a sharing violation it caused itself (#2272).
func trialStart(path string) error {
	cmd := proc.Command(path)
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.CreationFlags |= suspended | noWindow
	if err := cmd.Start(); err != nil {
		return err
	}
	cmd.Process.Kill()
	cmd.Wait()
	awaitMovable(path)
	return nil
}

// movable reports whether path may be renamed or removed now: a handle
// another program holds without sharing delete — a scanner reading a
// program that just started — makes opening it for delete fail with a
// sharing violation. Any other failure is not about the hold, and the
// move that fails on it says what it is.
func movable(path string) bool {
	p, err := syscall.UTF16PtrFromString(path)
	if err != nil {
		return true
	}
	h, err := syscall.CreateFile(p, deleteAccess, syscall.FILE_SHARE_READ|syscall.FILE_SHARE_WRITE, nil, syscall.OPEN_EXISTING, 0, 0)
	if err == nil {
		syscall.CloseHandle(h)
		return true
	}
	return !errors.Is(err, errSharingViolation) && !errors.Is(err, errLockViolation)
}

// awaitMovable waits for the file at path to be let go, over the pauses
// moveAside uses: the hold the trial's own start caused is a moment at
// most, and a file still held after them is left to the move that fails
// on it, whose error says why.
func awaitMovable(path string) {
	for i := 0; ; i++ {
		if movable(path) || i == len(asideWaits) {
			return
		}
		time.Sleep(asideWaits[i])
	}
}

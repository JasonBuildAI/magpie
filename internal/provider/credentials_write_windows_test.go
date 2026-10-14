package provider

import (
	"bytes"
	"context"
	"os"
	"syscall"
	"testing"
	"time"
)

// Windows: a file another program holds open without sharing it for a
// moment refuses to be renamed over, and refuses to be opened for writing
// too ("Access is denied", the steady package). The write that opened
// auth.json for writing itself failed on that moment, and the error was
// dropped: the rotated tokens never reached the disk while the refresh
// token on it had been spent — the sign-in OpenAI refuses on the next
// refresh, magpie's or Codex CLI's. The flush-then-rename write waits the
// holder out and lands.
func TestCodexRefreshOutlastsAHeldAuthJSON(t *testing.T) {
	home := signIn(t)
	auth := codexSignInSoon(t, home, "r-work")
	codexRefreshAnswers(t)

	name, err := syscall.UTF16PtrFromString(auth)
	if err != nil {
		t.Fatal(err)
	}
	h, err := syscall.CreateFile(name, syscall.GENERIC_READ, syscall.FILE_SHARE_READ, nil, syscall.OPEN_EXISTING, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	go func() {
		time.Sleep(100 * time.Millisecond)
		syscall.CloseHandle(h)
	}()

	if _, _, err := codexToken(context.Background(), auth); err != nil {
		t.Fatalf("a refresh while the file was held: %v", err)
	}
	b, err := os.ReadFile(auth)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Contains(b, []byte("r-work-2")) {
		t.Errorf("the rotated tokens are not on the disk, only the spent one's: %s", b)
	}
}

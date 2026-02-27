package nix

import (
	"errors"
	"os/exec"
	"testing"

	dogeboxd "github.com/Dogebox-WG/dogeboxd/pkg"
)

type testSubLogger struct {
	logs []string
	errs []string
	cmds [][]string
}

func (l *testSubLogger) Log(msg string)                    { l.logs = append(l.logs, msg) }
func (l *testSubLogger) Logf(msg string, a ...any)         { l.logs = append(l.logs, msg) }
func (l *testSubLogger) Err(msg string)                    { l.errs = append(l.errs, msg) }
func (l *testSubLogger) Errf(msg string, a ...any)         { l.errs = append(l.errs, msg) }
func (l *testSubLogger) Progress(p int) dogeboxd.SubLogger { return l }
func (l *testSubLogger) LogCmd(cmd *exec.Cmd)              { l.cmds = append(l.cmds, cmd.Args) }

func TestRebuildCacheFirstSuccess(t *testing.T) {
	t.Setenv("DOGEBOXD_NIX_CACHE_FIRST", "")
	logger := &testSubLogger{}
	nm := nixManager{}

	prev := runNixCommand
	t.Cleanup(func() { runNixCommand = prev })
	runNixCommand = func(cmd *exec.Cmd) error { return nil }

	if err := nm.Rebuild(logger); err != nil {
		t.Fatalf("Rebuild returned error: %v", err)
	}

	if len(logger.cmds) != 1 {
		t.Fatalf("expected 1 rebuild command, got %d", len(logger.cmds))
	}

	if got := logger.cmds[0]; len(got) < 6 || got[4] != "--max-jobs" || got[5] != "0" {
		t.Fatalf("expected cache-first args with --max-jobs 0, got %v", got)
	}
}

func TestRebuildCacheFirstFallback(t *testing.T) {
	t.Setenv("DOGEBOXD_NIX_CACHE_FIRST", "true")
	logger := &testSubLogger{}
	nm := nixManager{}

	prev := runNixCommand
	t.Cleanup(func() { runNixCommand = prev })
	callCount := 0
	runNixCommand = func(cmd *exec.Cmd) error {
		callCount++
		if callCount == 1 {
			return errors.New("cache miss")
		}
		return nil
	}

	if err := nm.Rebuild(logger); err != nil {
		t.Fatalf("Rebuild returned error: %v", err)
	}

	if len(logger.cmds) != 2 {
		t.Fatalf("expected 2 rebuild commands, got %d", len(logger.cmds))
	}

	if got := logger.cmds[0]; len(got) < 6 || got[4] != "--max-jobs" || got[5] != "0" {
		t.Fatalf("expected first command to be cache-only, got %v", got)
	}

	if got := logger.cmds[1]; len(got) != 4 {
		t.Fatalf("expected fallback command without cache-only args, got %v", got)
	}
}

func TestRebuildCacheFirstDisabled(t *testing.T) {
	t.Setenv("DOGEBOXD_NIX_CACHE_FIRST", "false")
	logger := &testSubLogger{}
	nm := nixManager{}

	prev := runNixCommand
	t.Cleanup(func() { runNixCommand = prev })
	runNixCommand = func(cmd *exec.Cmd) error { return nil }

	if err := nm.Rebuild(logger); err != nil {
		t.Fatalf("Rebuild returned error: %v", err)
	}

	if len(logger.cmds) != 1 {
		t.Fatalf("expected 1 rebuild command, got %d", len(logger.cmds))
	}

	if got := logger.cmds[0]; len(got) != 4 {
		t.Fatalf("expected standard rebuild command when cache-first is disabled, got %v", got)
	}
}

func TestRebuildCacheFirstAndFallbackFail(t *testing.T) {
	t.Setenv("DOGEBOXD_NIX_CACHE_FIRST", "true")
	logger := &testSubLogger{}
	nm := nixManager{}

	prev := runNixCommand
	t.Cleanup(func() { runNixCommand = prev })
	runNixCommand = func(cmd *exec.Cmd) error { return errors.New("build failed") }

	if err := nm.Rebuild(logger); err == nil {
		t.Fatal("expected rebuild to fail when both attempts fail")
	}

	if len(logger.cmds) != 2 {
		t.Fatalf("expected 2 rebuild commands, got %d", len(logger.cmds))
	}
}

func TestIsNixCacheFirstEnabled(t *testing.T) {
	t.Setenv("DOGEBOXD_NIX_CACHE_FIRST", "")
	if !isNixCacheFirstEnabled() {
		t.Fatal("expected cache-first enabled by default")
	}

	t.Setenv("DOGEBOXD_NIX_CACHE_FIRST", "off")
	if isNixCacheFirstEnabled() {
		t.Fatal("expected cache-first disabled for 'off'")
	}

	t.Setenv("DOGEBOXD_NIX_CACHE_FIRST", "1")
	if !isNixCacheFirstEnabled() {
		t.Fatal("expected cache-first enabled for '1'")
	}
}

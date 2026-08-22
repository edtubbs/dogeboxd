package nix

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSwapFileSizeForMissingPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")

	if got := swapFileSizeFor(missing); got != 0 {
		t.Errorf("expected 0 for an unstattable path, got %d", got)
	}
}

func TestSwapFileSizeForIsBounded(t *testing.T) {
	dir := t.TempDir()

	got := swapFileSizeFor(dir)

	if got != 0 && got < minSwapFileSizeMB {
		t.Errorf("expected either no swapfile or at least %d MiB, got %d", minSwapFileSizeMB, got)
	}

	if got > desiredSwapFileSizeMB {
		t.Errorf("expected at most %d MiB, got %d", desiredSwapFileSizeMB, got)
	}
}

func TestDetectSwapWithoutSwapDevice(t *testing.T) {
	if _, err := os.Stat(swapDevicePath); err == nil {
		t.Skipf("%s exists on this host", swapDevicePath)
	}

	device, size := detectSwap()

	if device != "" {
		t.Errorf("expected no swap device, got %q", device)
	}

	if size < 0 {
		t.Errorf("expected a non-negative swapfile size, got %d", size)
	}
}

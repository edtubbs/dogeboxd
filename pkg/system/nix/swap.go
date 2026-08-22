package nix

import (
	"os"
	"syscall"
)

const (
	// swapDevicePath is the by-label symlink created by `mkswap -L swap`,
	// which the installer runs on the dedicated swap partition.
	swapDevicePath = "/dev/disk/by-label/swap"

	// desiredSwapFileSizeMB is the swapfile size we aim for on hosts that
	// have no swap partition.
	desiredSwapFileSizeMB = 32 * 1024

	// minSwapFileSizeMB is the smallest swapfile worth creating. If we can't
	// fit at least this much, we create no swapfile at all rather than
	// leaving a swap unit that fails to activate.
	minSwapFileSizeMB = 4 * 1024

	// swapFileHeadroomMB is the amount of free space we always leave on the
	// filesystem holding the swapfile.
	swapFileHeadroomMB = 10 * 1024

	// swapFileDir is the directory the swapfile lives in. NixOS creates the
	// file itself, so we only need its filesystem to have room.
	swapFileDir = "/var/lib"
)

// detectSwap works out how the host should get swap. It is deliberately done
// here, at nix-write time on the box itself, rather than with
// builtins.pathExists in the nix template: nix evaluation may happen in a pure
// or sandboxed context that cannot see /dev at all, which would silently pick
// the wrong branch.
//
// If the host has a swap partition (created and labelled by the installer) we
// use it. Otherwise we fall back to a swapfile, sized to what the filesystem
// can actually spare so the swap unit doesn't fail at boot.
func detectSwap() (device string, swapFileSizeMB int) {
	if _, err := os.Stat(swapDevicePath); err == nil {
		return swapDevicePath, 0
	}

	return "", swapFileSizeFor(swapFileDir)
}

// swapFileSizeFor returns the swapfile size, in MiB, that fits in the
// filesystem containing dir while leaving headroom free. Returns 0 if a
// usefully sized swapfile won't fit.
func swapFileSizeFor(dir string) int {
	var stat syscall.Statfs_t
	if err := syscall.Statfs(dir, &stat); err != nil {
		return 0
	}

	availableMB := int(uint64(stat.Bavail) * uint64(stat.Bsize) / (1024 * 1024))

	usableMB := availableMB - swapFileHeadroomMB
	if usableMB < minSwapFileSizeMB {
		return 0
	}

	if usableMB > desiredSwapFileSizeMB {
		return desiredSwapFileSizeMB
	}

	return usableMB
}

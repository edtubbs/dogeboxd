package cmd

import "testing"

func TestClearSwapDevicesDeclGeneratedList(t *testing.T) {
	input := `{ config, lib, pkgs, modulesPath, ... }:

{
  fileSystems."/" = { device = "/dev/disk/by-uuid/abc"; fsType = "ext4"; };

  swapDevices =
    [ { device = "/dev/disk/by-uuid/deadbeef"; }
    ];

  networking.useDHCP = lib.mkDefault true;
}
`

	got, err := clearSwapDevicesDecl(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	want := `{ config, lib, pkgs, modulesPath, ... }:

{
  fileSystems."/" = { device = "/dev/disk/by-uuid/abc"; fsType = "ext4"; };

  swapDevices = [ ];

  networking.useDHCP = lib.mkDefault true;
}
`

	if got != want {
		t.Errorf("got:\n%s\nwant:\n%s", got, want)
	}
}

func TestClearSwapDevicesDeclNoSwap(t *testing.T) {
	input := "{ }\n"

	got, err := clearSwapDevicesDecl(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if got != input {
		t.Errorf("expected input to be unchanged, got: %s", got)
	}
}

func TestClearSwapDevicesDeclMalformed(t *testing.T) {
	if _, err := clearSwapDevicesDecl("swapDevices = [ { device = \"/dev/sda2\"; }"); err == nil {
		t.Error("expected an error for a declaration with no terminating semicolon")
	}
}

{ config, pkgs, lib, ... }:

{
  services.create_ap.enable = lib.mkForce false; # Disable create_ap in case it was enabled by the T6 Installer.

  networking = {
    {{if .USE_ETHERNET}}
    interfaces = {
      {{.INTERFACE}} = {
        useDHCP = true;
      };
    };
    {{else if .USE_WIRELESS}}
    wireless = {
      iwd = {
        # Override the OS image default (set to false in the OS fork's
        # nanopc-t6/base.nix) so that an explicit user-selected wireless
        # network actually brings up iwd.
        enable = lib.mkForce true;
      };
      interfaces = [ "{{.INTERFACE}}" ];
      networks = {
        "{{.WIFI_SSID}}" = {
          psk = "{{.WIFI_PASSWORD}}";
        };
      };
    };
    {{end}}
  };
}

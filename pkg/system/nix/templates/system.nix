{ config, pkgs, lib, ... }:

{
  networking.hostName = lib.mkForce "{{ .SYSTEM_HOSTNAME }}";
  networking.networkmanager.enable = lib.mkDefault false;

  console.keyMap = lib.mkForce "{{ .KEYMAP }}";

  time.timeZone = lib.mkForce "{{ .TIMEZONE }}";

  # Swap is configured here, on the dogebox host, and never inside a pup's
  # nspawn container: containers share the host kernel and its memory, so
  # they cannot bring their own swap.
  # Installs that already carry a dedicated swap partition (created by
  # `dbx install-to-disk`) keep using that partition; anything else (eg. the
  # ARM images, which have no swap partition) gets a host swapfile instead.
  swapDevices = lib.mkIf (!builtins.pathExists "/dev/disk/by-label/swap") [
    {
      device = "/var/lib/swapfile";
      size = 32 * 1024;
    }
  ];

  services.openssh.settings = {
    AllowUsers = [ "shibe" ];
  };

  services.openssh.banner = ''
+===================================================+
|                                                   |
|      ____   ___   ____ _____ ____   _____  __     |
|     |  _ \ / _ \ / ___| ____| __ ) / _ \ \/ /     |
|     | | | | | | | |  _|  _| |  _ \| | | \  /      |
|     | |_| | |_| | |_| | |___| |_) | |_| /  \      |
|     |____/ \___/ \____|_____|____/ \___/_/\_\     |
|                                                   |
+===================================================+
'';

  services.openssh.enable = lib.mkForce {{ .SSH_ENABLED }};

  users.users.shibe = {
    isNormalUser = true;
    group = "wheel";
    openssh = {
      authorizedKeys = {
        keys = [
          {{ range .SSH_KEYS }}"{{.Key}} # {{.ID}}"{{ end }}
        ];
      };
    };
  };

  {{ if gt (len .BINARY_CACHE_SUBS) 0 }}
  nix.settings.substituters = [
    {{ range .BINARY_CACHE_SUBS }}"{{.}}"{{ end }}
  ];
  {{ end }}

  {{ if gt (len .BINARY_CACHE_KEYS) 0 }}
  nix.settings.trusted-public-keys = [
    {{ range .BINARY_CACHE_KEYS }}"{{.}}"{{ end }}
  ];
  {{ end }}
}

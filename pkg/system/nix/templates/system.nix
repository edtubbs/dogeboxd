{ config, pkgs, lib, ... }:

{
  networking.hostName = lib.mkForce "{{ .SYSTEM_HOSTNAME }}";
  networking.networkmanager.enable = lib.mkDefault false;

  console.keyMap = lib.mkForce "{{ .KEYMAP }}";

  time.timeZone = lib.mkForce "{{ .TIMEZONE }}";

  # Swap is configured here, on the dogebox host, and never inside a pup's
  # nspawn container: containers share the host kernel and its memory, so
  # they cannot bring their own swap.
  #
  # Which swap we get is decided by dogeboxd when it writes this file, on the
  # box itself, not by evaluating builtins.pathExists here: nix evaluation can
  # happen in a pure/sandboxed context with no view of /dev, which would
  # silently pick the wrong branch.
  {{ if .SWAP_DEVICE }}
  # This host has a dedicated swap partition, labelled by `mkswap -L swap`
  # during install.
  swapDevices = [
    { device = "{{ .SWAP_DEVICE }}"; }
  ];
  {{ else if gt .SWAP_FILE_SIZE_MB 0 }}
  # No swap partition on this host, so use a swapfile instead. The size has
  # already been checked against the free space on / so the swap unit can
  # actually activate.
  swapDevices = [
    {
      device = "/var/lib/swapfile";
      size = {{ .SWAP_FILE_SIZE_MB }};
    }
  ];
  {{ end }}

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

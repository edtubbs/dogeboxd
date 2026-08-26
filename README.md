![Dogebox Logo](/docs/dogebox-logo.png)

# Dogeboxd

Dogeboxd is the core system manager service for [Dogebox OS](https://dogebox.org), a Linux distribution designed for the Dogecoin community. It provides a secure, user-friendly platform for running decentralized applications (called "Pups") with a focus on Dogecoin services.

## Overview

Dogeboxd acts as the central orchestrator for the Dogebox system, managing:

- **Pup Lifecycle Management**: Install, configure, start, stop, and remove containerized applications
- **System Configuration**: Network setup, storage management, SSH access, and system updates
- **Security**: Authentication through the Doge Key Manager (DKM) with cryptographic key management
- **API Services**: REST API, WebSocket connections, and internal routing for Pup communication
- **Resource Monitoring**: CPU, memory, and disk usage tracking for all running Pups

## Architecture

Dogeboxd follows a job-based architecture where actions are processed through a central dispatcher:

```
 REST API  ─────┐                    ┌──────────────┐
                │                    │  Dogeboxd{}  │
 WebSocket ─────┼─── Actions ───────►│  Job Queue   │────► Changes ───► WebSocket
                │                    │              │
 System     ────┘                    │ SystemUpdater│
 Events                              └──────────────┘
                                            │
                                            ▼
                                     Nix CLI / SystemD
```

## Key Components

### Pups (Dogebox Apps)

Pups are containerized applications that run in the Dogebox Runtime Environment (DRE). Each Pup:

- Has a manifest file (`pup.json`) defining its requirements and capabilities
- Runs in an isolated NixOS container with controlled access to system resources
- Can expose web interfaces, APIs, and interact with other Pups through defined interfaces
- Supports configuration management and automatic dependency resolution

### System Services

- **PupManager**: Manages Pup lifecycle, state persistence, and health monitoring
- **SystemUpdater**: Handles system-level changes, Nix configurations, and updates
- **NetworkManager**: Manages network interfaces (Ethernet/WiFi) and connectivity
- **DKM (Doge Key Manager)**: Provides cryptographic key management and authentication
- **SourceManager**: Manages Pup sources and package repositories

## Security Considerations

- All Pups run in isolated containers with restricted capabilities
- Authentication is required for all management operations
- Cryptographic operations are handled by the separate [DKM](https://github.com/dogebox-wg/dkm) service
- Network isolation ensures Pups can only communicate through defined interfaces
- File system access is limited to designated storage directories

### OP-TEE Secure Boot PTA Client

For Rockchip secure-boot provisioning via the OP-TEE secure boot PTA (`rk_secure_boot`), Linux userspace is optional as a client (this can be done pre-OS in bootloader stages).

If Dogebox-WG needs an OS-side client in Linux userspace, this repository (`dogeboxd`) is the right ownership boundary for it (for example, via `dogeboxd` or its privileged helper) because it already orchestrates host-level, privileged system operations.

This repository now includes secure-boot client commands:

- `dbx secure-boot info`
- `dbx secure-boot enable --hash <64-hex-char-sha256> --key-size <2048|4096> --yes`

**Please note:** Although pups are isolated, we provide no guarantees that a malicious pup cannot attack your host or other pups. Therefore, we recommend only installing pups from known-good sources.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Links

- [Dogebox Documentation](https://dogebox.org)
- [DogeDev Discord Community](https://discord.com/invite/VEUMWpThg9)

## Acknowledgments

Dogeboxd is developed by the Dogecoin Foundation and the Dogebox community. Special thanks to all contributors who have helped make this project possible.

---

**Note**: This project is under active development. Features and APIs may change. Please refer to the official documentation for the most up-to-date information.

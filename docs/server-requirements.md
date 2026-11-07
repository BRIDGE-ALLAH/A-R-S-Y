# Managed Server Requirements & Security Hardening

To ensure secure communication and deterministic application deployments, target Linux hosts managed by A-R-S-Y must adhere to these prerequisites:

---

## 1. Operating System & Architecture
- **Supported Distributions**: Debian 11+, Ubuntu 20.04+, RHEL/Rocky Linux 8+, Alpine 3.16+
- **Architectures**: x86_64 (amd64), aarch64 (arm64)
- **Init System**: systemd or OpenRC

---

## 2. Docker & Compose Runtime
- **Docker Engine**: Version 20.10 or newer.
- **Docker Compose**: Docker Compose v2 plugin (`docker compose version` must return exit code 0). Legacy standalone `docker-compose` (v1) is deprecated and unsupported.
- The SSH deployment user must have permission to interact with Docker without `sudo` prompts (member of `docker` group) or have passwordless sudo for docker commands.

---

## 3. SSH Configuration & Hardening
- OpenSSH server daemon running.
- Key authentication enabled (`PubkeyAuthentication yes`).
- Password authentication disabled in production (`PasswordAuthentication no`).
- Root login disabled or restricted to key-only (`PermitRootLogin prohibit-password`).
- Target release directory path: `/opt/arsy/apps/` must be writable by the deployment user.

---

## 4. Port Rules & Firewall
- Inbound SSH (default port `22` or custom port) allowed from A-R-S-Y control-plane IP.
- Inbound application ports (e.g., `80`, `443`, `8080`) mapped in your Compose manifest exposed to the public or private network as needed.
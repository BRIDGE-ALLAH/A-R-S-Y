# A-R-S-Y Installation & Quickstart Guide

A-R-S-Y is a self-hosted application deployment and infrastructure control plane (a lightweight, modern Coolify alternative) built natively in Go and backed by SQLite.

---

## 1. Prerequisites

- **Control Server**: Linux (amd64 / arm64), macOS, or Windows host with network connectivity.
- **Managed Servers (Targets)**:
  - Linux with systemd
  - OpenSSH server running with key authentication enabled
  - Docker Engine & Docker Compose plugin installed (`docker compose version` supported)
  - Unprivileged deployment user with sudo rights or access to Docker group

---

## 2. Standalone Binary Startup

Download the pre-compiled binary for your architecture or build from source:

```bash
# Build standalone server
mkdir -p .builds
go build -o .builds/arsy ./examples/base

# Run the server
./.builds/arsy serve --http="0.0.0.0:8090"
```

The server binds to port `8090` by default:
- **A-R-S-Y Console UI**: `http://localhost:8090/arsy/`
- **PocketBase Admin UI**: `http://localhost:8090/_/`
- **API Base**: `http://localhost:8090/api/arsy/v1/`

---

## 3. First-Run Setup & Superuser Creation

A-R-S-Y never exposes an unauthenticated superuser creation endpoint over the web. Create the primary superuser from the CLI:

```bash
./.builds/arsy superuser create admin@example.com "YourSecurePassword123"
```

---

## 4. Secret Encryption Key Setup

A-R-S-Y protects private SSH keys and application environment variables with AES-GCM encryption at rest:

```bash
# Generate a new 32-byte hex secret key
SECRET_KEY=$(./.builds/arsy secret-key generate)

# Export the key in your production environment
export ARSY_SECRET_KEY="$SECRET_KEY"

# Start A-R-S-Y with the key configured
./.builds/arsy serve
```

Verify your setup with the diagnostic doctor command:
```bash
./.builds/arsy doctor
```

---

## 5. First Deployment Walkthrough

1. **Sign in**: Open `http://localhost:8090/arsy/` and sign in with your superuser credentials.
2. **Register Server**:
   - Go to **Servers** tab.
   - Enter your target host IP/FQDN, SSH port, deployment user, and unencrypted private key.
   - Click **Test SSH / host key**. Verify and confirm the probed fingerprint.
   - Run **Check readiness** to verify Docker & Compose.
3. **Create Project**:
   - Go to **Projects** tab. Enter a name (e.g., `Production API`).
4. **Create Application**:
   - Select your project and ready server.
   - Provide application name and valid Docker Compose manifest (using pre-built images).
5. **Review Deployments**:
   - Open **Deployments** to inspect execution state, logs, and perform rollback or stop actions.
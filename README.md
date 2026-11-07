# A-R-S-Y

**A-R-S-Y** is an independently maintained, Go-native application deployment control plane. Its goal is to provide a self-hosted way to manage Linux hosts and deploy applications with Docker Compose.

A-R-S-Y is derived from [PocketBase](https://github.com/pocketbase/pocketbase) and retains PocketBase as its embedded application foundation (SQLite, authentication, migrations, and HTTP runtime). It is not an unchanged PocketBase distribution; A-R-S-Y-specific functionality is developed in this repository.

## Project status

A-R-S-Y is under active early-stage development. The MVP-1 plan defines the intended first release: server registration over SSH, explicit SSH host-key pinning, Docker/Compose readiness checks, projects and applications, Compose deployments, deployment history, logs, and safe lifecycle operations.

**Status note:** planned features are not necessarily implemented yet. Consult the [MVP-1 task checklist](.plan/task-mvp-1.md) and the [Actions results](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions) for the current implementation and verification status. Do not use this project to manage production infrastructure until the relevant security and integration checks have passed.

## Requirements

- Go version specified in [go.mod](go.mod) (currently Go 1.27).
- Node.js 24 and npm are required only when building the console frontend.
- Docker is required only for Docker integration tests and for target hosts that will run deployed applications.

## Build

Build the A-R-S-Y console assets and standalone Go server:

```sh
cd dashboard
npm install
npm run build
cd ..
mkdir -p .builds
CGO_ENABLED=0 go build -o .builds/arsy ./examples/base
```

Run the server:

```sh
.builds/arsy serve
```

The dashboard build is embedded into the Go executable. Node.js is not required at runtime. The separate `ui/` directory remains the PocketBase Superuser UI and is not replaced by the A-R-S-Y console.

The server retains PocketBase's existing administrative interface and runtime capabilities. A-R-S-Y-specific console routes and deployment features are being introduced incrementally.

## Test

Run the Go test suite:

```sh
go test ./...
```

Run the A-R-S-Y dashboard build:

```sh
cd dashboard
npm install
npm run build
```

To verify the existing PocketBase Superuser UI separately, run `npm ci && npm run build` in `ui/`.

Continuous integration runs these checks on pushes and pull requests. See [GitHub Actions](https://github.com/BRIDGE-ALLAH/A-R-S-Y/actions).

## Configuration

A-R-S-Y-specific settings use these environment variables:

| Variable | Default | Purpose |
| --- | --- | --- |
| `ARSY_DATA_DIR` | `pb_data/arsy` | Private A-R-S-Y runtime data directory |
| `ARSY_API_PREFIX` | `/api/arsy/v1` | Versioned A-R-S-Y API prefix |
| `ARSY_CONSOLE_PATH` | `/arsy` | A-R-S-Y web console route |
| `ARSY_SECRET_KEY_ENV` | `ARSY_SECRET_KEY` | Name of the environment variable that supplies the secret key |

The secret key must decode from base64 to exactly 32 bytes for AES-256-GCM. Generate one explicitly with `arsy secret-key generate`; the command prints it once to stdout and does not save it. Store it in a secret manager or a file readable only by the service account (mode `0600`), then provide it through the environment. Diagnostics never print the key value. Do not commit it to source control.

## Architecture and scope

The intended MVP is a single Go control-plane process backed by PocketBase/SQLite, with a separately built web console. Remote host operations are performed over SSH. The first deployment target is Docker Compose on Linux.

MVP-1 intentionally excludes source-code builds, Git-based deployment, automatic domain/proxy/TLS provisioning, managed databases, Kubernetes, billing, and multi-tenant RBAC.

## Security

- SSH host keys must be explicitly pinned before a host is trusted.
- Stored SSH credentials and application secrets must be encrypted at rest.
- Remote commands and deployment paths must be validated and bounded.
- Production deployments must be initiated by an authenticated administrator.
- Never commit private keys, passwords, API tokens, encryption keys, or production environment files.

Report security issues privately as described in [SECURITY.md](SECURITY.md).

## License and attribution

A-R-S-Y is independently maintained and derived from PocketBase. Preserve all upstream copyright notices and license terms in [LICENSE.md](LICENSE.md), including applicable third-party dependency licenses. Changes specific to A-R-S-Y are maintained in this repository.

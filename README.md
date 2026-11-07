# A-R-S-Y

A-R-S-Y is an independently maintained, Go-native control plane for self-hosted application deployment and infrastructure management. It is derived from PocketBase and currently retains PocketBase's embedded SQLite storage, authentication, API, migration system, and Superuser dashboard as its foundation.

## Project status

A-R-S-Y is in early MVP development. The repository contains the inherited PocketBase foundation, a separate A-R-S-Y console shell, and initial control-plane endpoints. Remote server management over SSH, Docker Compose deployment, and application operations are still planned and are not available yet.

The intended MVP-1 workflow is to register a Linux server with pinned SSH host keys, verify Docker readiness, and deploy existing container images with Docker Compose. The control plane runs as a standalone Go executable. The A-R-S-Y console uses React, Vite, and shadcn/ui; Node.js is needed only to build its static assets.

## Requirements

- Go 1.27 or newer
- A supported operating system and architecture for the bundled SQLite driver

## Build and run

Build the console assets, then the standalone executable from the repository root:

```sh
npm --prefix dashboard ci
npm --prefix dashboard run build
go build -o arsy ./examples/base
```

Start the server:

```sh
./arsy serve
```

On Windows, use `go build -o arsy.exe ./examples/base` and run `./arsy.exe serve` after the same dashboard build steps.

The inherited PocketBase data directory defaults to `pb_data` beside the executable (or in the current directory when run with `go run`). Use the existing `--help` output to inspect available server flags. The embedded A-R-S-Y console is served at `/arsy/`. The PocketBase Superuser dashboard at `/_/` remains temporarily available while its administrator workflows are migrated into A-R-S-Y; it will be disabled after parity is complete. SSH server management and the deployment API are not implemented yet.

### A-R-S-Y environment settings

The control-plane configuration recognizes these variables and validates them before startup:

| Variable | Default | Purpose |
| --- | --- | --- |
| `ARSY_DATA_DIR` | PocketBase default (`pb_data` beside the executable, or in the current directory with `go run`) | A-R-S-Y data directory. |
| `ARSY_API_PREFIX` | `/api/arsy/v1` | URL path reserved for the A-R-S-Y API. |
| `ARSY_CONSOLE_ROUTE` | `/arsy` | URL path reserved for the A-R-S-Y console. |
| `ARSY_SECRET_KEY_ENV` | `ARSY_SECRET_KEY` | Name of the environment variable reserved as the secret-key source. The key value is not copied into configuration or logs. |

Generate a 256-bit secret key only when needed and store it outside the database, preferably in a secret manager or a file readable only by the service account:

```sh
./arsy generate-secret-key
```

Set the resulting 64-character hexadecimal value in the environment variable named by `ARSY_SECRET_KEY_ENV` (default `ARSY_SECRET_KEY`). Treat it like a password: do not commit it or include it in logs. Encryption support uses this key for A-R-S-Y-managed credentials.

Run the Go test suite with:

```sh
go test ./...
```

## Relationship to PocketBase

A-R-S-Y is independently maintained and is not an official PocketBase release or product. It builds on PocketBase source and preserves its upstream license and required attribution. See [LICENSE.md](LICENSE.md) for the MIT license and copyright notice.

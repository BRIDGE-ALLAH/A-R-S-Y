# Completion Report: Stop a running application

## Status: COMPLETE

## Task results

- [x] T1 `arsy/httpapi/stop_application.go` + route in `arsy/httpapi/routes.go`.
- [x] T2 `arsy/deployments/stop.go` — guards, remote `compose down`, lifecycle +
  audit.
- [x] T3 Frontend `dashboard/src/main.js` stop button + handler; dist rebuilt.
- [x] T4 `arsy/httpapi/stop_application_test.go` — 6 tests (5 guards + happy
  path).
- [x] T5 report.

## Why this was real work, not decoration

`servers.Executor.StopComposeManifest` (remote `docker compose down`) already
existed and was tested, but **nothing in the product could reach it**: there was
no endpoint, no `lifecycle_state` transition, and no UI. That is the exact
"backend without a usable frontend" gap this workstream closes. No new remote
capability was invented and no success is fabricated — the handler returns 502
when the remote command fails.

## Verification log

| # | Command | Result |
|---|---------|--------|
| 1 | `go build ./...` | PASS (exit 0) |
| 2 | `go vet ./arsy/httpapi/...` | PASS (exit 0) |
| 3 | `go test ./arsy/httpapi/... -run TestStopApplication -count=1 -v` | PASS — 6/6 |
| 4 | `cd dashboard; npm run build` | PASS (dist `index-B70DCC0p.js`) |
| 5 | `go test ./arsy/... -count=1` | PASS (all arsy packages) |

The happy-path test uses `servers.NewTestSSHServer` — a real in-process SSH
server — so the test exercises an actual SSH handshake, pinned host-key
verification, public-key auth, and an actual exec of the generated
`docker compose ... down` command, then asserts:

- HTTP 200 with `lifecycle_state=stopped`;
- the application record is persisted as `stopped`;
- one `application_stopped` audit-log row exists.

Guard tests assert 409 for: non-running application, revoked host key, missing
pinned fingerprint, no succeeded deployment, and an active deployment.

## Deviations

- Stop is synchronous (like the existing readiness/test-server endpoints) rather
  than queued. This matches the existing SSH-operation UX and avoids inventing a
  second job queue; the operation is bounded by a 60s remote timeout.

## Remaining risks / follow-ups

- Stopping leaves the release files on the host (only `compose down` runs). This
  matches Compose semantics and keeps rollback possible by redeploying.
- If a queue-based stop is later required (very slow hosts), the
  `StopApplication` function is already isolated and reusable by a worker.

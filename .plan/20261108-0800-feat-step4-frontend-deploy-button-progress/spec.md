# Spec — Step 4 Frontend (#64)

## Requirements

- REQ-064-01: Each application card in Projects page shows a "Deploy" button.
- REQ-064-02: Clicking Deploy POSTs to `/api/arsy/v1/deployments` with `application_id`.
- REQ-064-03: After queuing, client polls `GET /api/arsy/v1/deployments/{id}/progress` every 3 seconds.
- REQ-064-04: Polling stops when deployment reaches terminal state (succeeded, failed, cancelled).
- REQ-064-05: Progress bar shows percent and step label, sanitized via escapeHTML.
- REQ-064-06: Deployments page shows animated progress bar for active deployments.
- REQ-064-07: Dashboard builds (`npm run build`) and is embedded into Go binary.

## Acceptance Criteria

- `npm run build` exits 0, produces assets in `arsy/console/dist/`.
- `go build ./...` exits 0.
- Existing tests remain green.

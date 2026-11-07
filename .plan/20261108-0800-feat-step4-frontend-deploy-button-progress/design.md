# Design — Step 4 Frontend (#64)

## Changes

### `dashboard/src/main.js`
- Added Deploy button to application cards in `refreshProjects()` rendering.
- Added `pollDeploymentProgress(deploymentId, progressEl)` — bounded 3s polling loop, max 200 iterations.
- Added `ACTIVE_STATES` set for terminal-state detection.
- Added click handler on project list for `deploy-app` action.
- Updated `renderDeploymentManager()` to show animated progress bar for active deployments.

### `dashboard/src/style.css`
- Added `.progress-bar-wrap`, `.progress-bar`, `.progress-bar-active` (CSS animation).
- Added `.progress-step`, `.progress-error`, `.deploy-btn` styles.

## Data Flow
1. User clicks Deploy on app card → POST `/api/arsy/v1/deployments` → gets `deployment_id`.
2. Client starts `pollDeploymentProgress()` — GET `/api/arsy/v1/deployments/{id}/progress` every 3s.
3. Progress bar updates with `percent` and `step` from response.
4. Polling stops when `state` not in ACTIVE_STATES.

## Security
- All dynamic content escaped via `escapeHTML()`.
- Bounded polling prevents runaway requests.
- Button disabled during active operation.

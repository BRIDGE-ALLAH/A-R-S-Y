# Task — Step 4: Frontend Deploy Button + Progress Bar (#64)

Status: COMPLETE
Scope: feature
Issue: #64

## Checklist

- [x] [S] Add Deploy button to each application card in Projects page (data-action="deploy-app" + data-app-id).
- [x] [S] Add deploy progress container below each app card (deploy-progress-{appId} div, initially hidden).
- [x] [M] Add click handler for Deploy button: POST /api/arsy/v1/deployments, show progress container.
- [x] [M] Add pollDeploymentProgress(deploymentId, progressEl): bounded 3s polling, stops at terminal state.
- [x] [S] Add ACTIVE_STATES set for state classification.
- [x] [S] Add progress bar animation in Deployments page for active deployments.
- [x] [S] Add CSS: .progress-bar-wrap, .progress-bar, .progress-bar-active (animated), .progress-step, .progress-error, .deploy-btn.
- [x] [S] Sanitized output: all user data via escapeHTML() (already enforced throughout).
- [x] [S] Run npm run build — SUCCESS (vite build, assets generated in arsy/console/dist/).
- [x] [S] Run go build ./... — SUCCESS (embedded dist included).
- [x] [S] Commit and push.

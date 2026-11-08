# Rules

- No fake success: deletes/updates must reflect real persisted state.
- Never return `manifest_ciphertext`, `compose_manifest` plaintext, SSH credentials, or secrets in DTOs.
- Keep server-side validation/guardrails (slug/name format, active deployments/app references before destructive delete).
- Use superuser-only access consistent with existing A-R-S-Y control-plane routes.
- No unrelated refactors; keep diffs focused.
- All claims require command evidence.

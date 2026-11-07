# Rules — Step 1 CI/deployments regression fix

1. Do not weaken #149 guarantees: no two non-queued (in-progress) deployments for the same application simultaneously; API must still reject a new deployment when the application already has an active/pending deployment (409), except idempotency returning the original result.
2. Fix root cause, not symptoms: claim logic must remain correct for both single-app and multi-app workloads.
3. Preserve existing state machine transitions and unrelated behavior.
4. Add/adjust tests only where necessary to reflect correct semantics; never delete tests to make the suite pass.
5. Conventional Commits; only files required by this step; no production writes, no force push.
6. All artifacts in English.

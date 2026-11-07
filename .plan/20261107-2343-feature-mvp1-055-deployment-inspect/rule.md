# Rules — MVP1-055 Deployment Inspection

1. Only the existing closed set of read-only SSH operations (`arsy/servers`) may run remotely; never execute caller-supplied commands.
2. Never log or return decrypted SSH keys; clear key material after use.
3. Treat `deploymentID` and all record fields as untrusted input; validate the deployment ID pattern before any lookup.
4. Do not change deployment state transitions, the processor flow, or the record schema — this feature is read-only.
5. Do not add dependencies; use the standard library plus existing module deps only.
6. Preserve existing exported behavior of `AllComposeContainersRunning` for previously tested inputs.
7. All artifacts, code, and commit messages in English; Conventional Commits only.
8. No production writes, no pushing to other branches, no force pushes.

1. Never weaken tests, lint, security checks or CI to make a task pass
2. Fix root causes; keep diffs focused and backward-compatible unless change is explicitly intended
3. Add meaningful tests for behavior changes and bugs; test according to impact
4. Preserve existing user modifications, untracked files and local data
5. Never bypass branch protection, required reviews or required CI
6. No destructive production data/schema/storage operation without explicit approval
7. Use isolated environments and rollback plans for high-risk changes
8. A missing Git tag is not a release blocker; verify canonical repository and permissions
9. Evidence before completion claims: run relevant commands, inspect output and exit status
10. Distinguish PASS, FAIL, NOT RUN and BLOCKED in reporting

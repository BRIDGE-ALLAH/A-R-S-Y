 # Hard Rules for Implementation

 1. Never add dependencies the active task does not list.
 2. Do not disable linters, type-checks, or CI gates to make a task pass.
 3. Fix root causes rather than applying surface-level patches.
 4. Preserve existing tests; add regression tests for every P0/P1 fix.
 5. Update documentation as necessary for actual behavior changes.
 6. Keep changes consistent with the style of the existing codebase.
 7. Use repo-relative paths in all documentation.
 8. Never weaken tests, lint, security checks or CI to get green.
 9. Do not push, deploy, or run production writes unless the task explicitly authorizes it.
 10. Do not do drive-by refactors — only change what the task requires.

 # Security Rules
 1. Secrets must never be in source, commits, or logs.
 2. Authentication and authorization must be implemented for all sensitive operations.
 3. Tenant isolation must be preserved where applicable.
 4. Input validation must be enforced on all API endpoints.
 5. Database constraints must prevent invalid data.

 # Quality Rules
 1. Every P0/P1 fix must add a regression test.
 2. Code must pass existing lint and type-checks.
 3. Performance considerations must be addressed for critical paths.
 4. Error handling must be comprehensive and user-friendly.
 5. Documentation must be synchronized with actual code behavior.

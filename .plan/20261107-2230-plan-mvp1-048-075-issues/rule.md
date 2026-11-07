# Rules: MVP1 Remaining Issues

## Implementation Rules

1. **Follow existing conventions**: All new code must follow the existing codebase style, naming, and patterns.

2. **No breaking changes**: All implementations must maintain backward compatibility with existing APIs and data contracts.

3. **Test coverage**: Every P0/P1 fix must add a regression test. New features should include appropriate tests.

4. **Preserve existing tests**: Do not modify or delete existing tests unless absolutely required for the fix.

5. **Use repo-relative paths**: All documentation and code should use repo-relative paths (e.g., rsy/deployments/log.go).

6. **Conventional Commits**: All commits must follow Conventional Commits format:
   - eat: for new features
   - ix: for bug fixes
   - docs: for documentation changes
   - 
efactor: for code restructuring
   - 	est: for test additions
   - chore: for maintenance tasks

7. **Security**: 
   - Never hard-code secrets or encryption keys in source
   - Use environment variables or PocketBase for key storage
   - Follow the existing pattern in rs/security/encryption.go

8. **State machine integrity**: For Issue #149, all deployment state transitions must go through the Transition function in state.go. Do not bypass state checks.

9. **Atomic operations**: For Issue #142, all SFTP operations must maintain the atomic temp-file + rename pattern already established in sftp.go.

10. **Opt-in design**: For Issue #70, new configuration must be opt-in and not affect existing functionality when disabled.

## Forbidden Actions

- Do not disable linters, type-checks, or CI gates to make a task pass
- Do not add dependencies the active task does not list
- Do not push, deploy, or run production writes unless the task explicitly authorizes it
- Do not do drive-by refactors — only change what the task requires
- Do not add copyright or license headers unless specifically requested

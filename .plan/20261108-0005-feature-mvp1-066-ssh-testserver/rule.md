# Rules — MVP1-066 SSH Test Server

1. No real SSH infrastructure; use embedded test server or mock in tests.
2. Test keys must be deterministic fixtures (embed in test code or use fixed seed).
3. Server must clean up gracefully (close listeners, goroutines) after each test.
4. No dependency on external SSH daemons or VPS access.
5. Tests must not bind to privileged ports (<1024) or interfere with system SSH.
6. Mock executor for remote commands; capture output and exit codes.

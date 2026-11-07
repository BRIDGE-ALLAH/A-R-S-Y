# Rules — MVP1-067 Deterministic Tests

1. All tests use local fixtures; no external network access.
2. Tests are deterministic: same input → same result every time.
3. No timeouts or flakiness from system load or clock skew.
4. Existing test suite is respected; new tests extend coverage.

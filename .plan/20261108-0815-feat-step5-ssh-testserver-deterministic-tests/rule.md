# Rules — Step 5

1. All tests must be network-independent (127.0.0.1 loopback only, or in-process fakes).
2. SSH test server must use deterministic keys and canned responses (no external dependencies).
3. Worker tests must use fake Process executors — no real SSH/Compose/Docker.
4. Do not add external test server dependencies (e.g., gliderlabs/ssh).
5. Existing test behavior must not regress.

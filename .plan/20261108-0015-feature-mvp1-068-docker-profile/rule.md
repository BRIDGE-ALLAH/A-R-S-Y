# Rules — MVP1-068 Docker Profile

1. Docker integration is **opt-in** via `ARSY_ENABLE_DOCKER_TESTS` env var.
2. All containers use disposable project/temp volumes; no host mounts.
3. Cleanup removes only created containers/volumes; never `docker system prune`.
4. Tests skip gracefully if Docker not available or opt-in flag not set.
5. No production Docker integration; local testing only.

# Rules — Step 4 Frontend

1. All dynamic content rendered via escapeHTML() — no raw innerHTML with user data.
2. Bounded polling: 3s interval, stops at terminal state (succeeded/failed/cancelled). Max 200 polls as safety.
3. Progress bar must be visible and informative (percent + step label).
4. Deploy button must be disabled during active deploy to prevent double-submission.
5. Dashboard must build with `npm run build` and backend with `go build ./...`.
6. Do not break existing server/project/application/deployment list functionality.

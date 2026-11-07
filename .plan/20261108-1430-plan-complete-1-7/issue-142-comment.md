## Verification evidence (main)
- Manifest encryption at rest (AES-256-GCM) + ciphertext storage/redaction implemented: 90c2ef5, 5bfa7dc.
- API/DTO responses redact manifest ciphertext/plaintext; decrypt only for deployment validation flow.
- Verified: `go test ./arsy/httpapi/... ./arsy/deployments/...` pass; no-plaintext-leakage acceptance covered by tests.
Issue closed per MVP1 acceptance reconciliation.
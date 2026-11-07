# Rules — MVP1-073 Manifest Encryption

1. Manifest encryption uses the existing `arsy/security.EncryptWithEnv` helper only.
2. Plaintext manifests must not appear in API DTOs, logs, or audit records.
3. Decrypt only in the processor execution path, immediately before upload.
4. Backward compatibility: if no ciphertext exists (legacy record), fall back to plaintext `compose_manifest`.
5. Validation (ValidateCompose) still runs on plaintext before encryption.
6. Encrypted manifests are stored in `manifest_ciphertext`; `compose_manifest` stores structural-only (non-secret) view.

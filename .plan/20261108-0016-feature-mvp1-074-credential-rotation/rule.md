# Rules — MVP1-074 Credential Rotation

1. New credential is encrypted before write; old is deleted only after successful encryption.
2. Rotation is atomic: success or fail, never partial update.
3. All cached SSH clients for this server are invalidated immediately.
4. Rotation requires superuser auth.
5. Old credential is cleared from memory (wipe secrets).

**Note**: Full audit trail (who/when/log entries) deferred to follow-up.

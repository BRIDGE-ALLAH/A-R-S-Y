# AENG-014: Application Module Schema Migration Plan Format

## Versioned format
```yaml
module: <module-id>
version: <semver-or-monotonic>
migrations:
  - id: <stable-unique-id>
    checksum: <sha256-of-canonical-operation-payload>
    description: <human>
    operations:
      - op: create_model|add_field|add_index|add_relation|deprecate_field|deprecate_model
        ... # typed payload
    destructive: false # required true for drop/delete/data-loss ops
```

## Rules
- **Ordered & deterministic**: operations apply in declared order; IDs are globally unique per module.
- **Identity & checksum**: executor records `(module, id, checksum, status, applied_at, error?)`; checksum drift MUST be rejected before apply.
- **Destructive operations**: require explicit opt-in (`destructive: true` + operator approval flag in executor config). Default deny.
- **Failure/retry**: record safe failure state; retry only idempotent/recoverable operations per documented policy; never mark success on partial failure.
- **Compatibility**: no edits to historical PocketBase migrations (`migrations/`, `arsy/migrations` timestamps immutable).

## Examples
- Initial: `create_model company` + fields + indexes.
- Upgrade: `add_field company.tax_id`, `add_index company.tax_id_unique` (non-destructive).
- Deprecation: `deprecate_field company.legacy_code` (no data drop).

## Non-goals
- Automatic destructive data loss.
- Executing migrations in this documentation issue (executor is AENG-032/#133).

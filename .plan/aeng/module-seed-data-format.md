# AENG-018: Idempotent Module Seed Data Format

## Data classes (distinct)
- **System data**: required by engine/core; installed once; protected.
- **Module master data**: required/reference data owned by module (currencies, sequences templates) with stable keys.
- **Demo data**: opt-in only; never mixed with required system/module master data.

## Record contract
```yaml
module: <id>
kind: system|master|demo
locale: <optional BCP-47>
effective_from: <optional date>
records:
  - key: <stable-module-scoped-key>   # idempotency/provenance key
    model: <registered-model>
    policy: insert_only|upsert_unless_user_edited|never_overwrite
    values: { ... }
```

## Rules
- **Idempotent**: re-running seed MUST NOT duplicate records; match by `(module, model, key)`.
- **Provenance**: store source module/key and policy; user edits mark record as user-owned/customized.
- **Upgrades**: default `upsert_unless_user_edited`; never overwrite user-edited business data by default.
- **Ordering/dependencies**: resolve module dependencies first; seed order deterministic (system -> master -> demo).
- **Locale/effective date**: locale-specific variants keyed separately; effective-date records must not retroactively corrupt posted business data.

## Examples
- Currency: `key: currency.USD`, `model: core.currency`, `policy: upsert_unless_user_edited`.
- Country: stable ISO key.
- Sequence: template record with `policy: insert_only` for defaults.

## Non-goals
- Bulk national datasets, accounting/tax rules, or ISP-specific records in generic core.
- Seed execution in this issue (runner is AENG-034/#135).

# AENG-009: Model Relation Metadata Contracts

## Scope
Engine-neutral relation metadata for application models: many-to-one, one-to-many, many-to-many, inverse fields, delete behavior, target validation.

## Contract (normative)

- **Cardinality**: explicit (`many_to_one`, `one_to_many`, `many_to_many`).
- **Target**: `many_to_one` MUST reference a valid registered target model.
- **Inverse**: `one_to_many` MUST reference a valid `many_to_one` inverse field on the target model; inverse cardinality/target must agree.
- **Many-to-many**: MUST declare join-table ownership and deterministic naming (module + models + relation key); no implicit ad-hoc table names.
- **Delete behavior**: explicit per relation (`restrict`, `set_null` where nullable, `detach_join` for M:N). **Cascading deletion is never implicit.**
- **Validation**: unknown targets/fields, cardinality mismatches, missing inverse, unsafe delete modes MUST fail validation with path-addressable errors.

## Implementation anchors
- `arsy/engine/relations/types.go`, `validate.go`, `validate_test.go` (table-driven malformed definition coverage).
- No query preloading implementation in this contract issue.

## Non-goals
- Runtime cascade deletes without explicit policy.
- Cross-tenant relation resolution.

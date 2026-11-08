# AENG-019: Module Security Metadata Contract

## Principles
- **Deny by default**.
- **Server-side only**: frontend visibility is never authorization.
- **Tenant isolation ≠ company scope**: tenant boundary is enforced by BaaS/tenancy; multi-company/business-unit scope is an application-layer record filter within an authorized tenant.

## Metadata
```yaml
roles:
  - id: <stable>
permissions:
  - model: <registered-model>
    operations: [create, read, update, delete, list, count]
    roles: [...]
    fields:
      readable: [*|fields]
      writable: [fields]
    record_rule:
      ast: <validated query AST node> # never raw SQL
    company_scope: none|same_company|explicit_rule
```

## Rules
- Model-level permissions and record-level rules are distinct; both must pass.
- Record rules use **validated query AST** (same constraints as `arsy/engine/query`), never raw SQL strings.
- Unknown models/fields/operations in metadata MUST fail validation.
- Company scope is evaluated only after tenant isolation and membership authorization.
- Unauthorized fields are omitted/unwritable in server paths.

## Examples
- Positive: `sales.order read` for `sales_user` with `record_rule.ast` limiting to `company_id = :auth.company_id`.
- Negative: rule referencing unknown field `companyId` (typo) -> validation error; raw SQL fragment -> rejected.

## Non-goals
- Full authorization engine implementation in this issue (runtime enforcement is AENG-035/#136).
- Granting tenant access; tenant membership remains authoritative.

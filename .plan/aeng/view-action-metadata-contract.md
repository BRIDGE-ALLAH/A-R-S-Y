# AENG-020: Metadata-Driven View and Action Manifest

## Scope
Declarative, framework-neutral UI metadata for menus, actions, forms, lists, widget hints, and access references.

## Contract
```yaml
module: <id>
views:
  - id: <stable-id>
    kind: form|list
    model: <registered-model>
    fields:
      - name: <registered-field>
        widget: <hint>
        label_key: <i18n-key>
    actions: [<action-id>]
menus:
  - id: <stable-id>
    label_key: <i18n-key>
    action: <action-id>
actions:
  - id: <stable-id>
    kind: open_view|server_action|url
    target: <view-id|endpoint-id>
    access: <permission-ref> # references server-side permission; narrows only
```

## Rules
- View/action references MUST resolve to registered models, fields, and declared actions; unknown references fail validation with path-aware errors.
- **UI metadata cannot grant server-side permissions**; `access` may only narrow existing server authorization.
- Stable IDs required for module upgrades and deterministic serialization.
- Framework-neutral: no React/Vue component names or executable client code.

## Example (company)
- Form view `company.form`: model `base.company`, fields `name`, `slug`, `is_active`.
- List view `company.list`: columns `name`, `slug`; action `open_form`.

## Non-goals
- Frontend renderer implementation.
- Arbitrary client-side code execution.
- Duplicate of this spec: issue #118 mirrors #117; same contract applies.

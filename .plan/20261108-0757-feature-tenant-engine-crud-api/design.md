# Design

- Add `httpapi` engine handlers that: require superuser; resolve tenant (catalog+provider/router); open tenant datastore; wrap `dbx`/sql handle for `persist.Store`; use app/registry dependency injected at registration/bootstrap.
- Reuse `persist.Store` CRUD/query (keeps validation in engine/query + explicit allowlist) — no raw SQL from HTTP.
- Map `persist` errors: `ErrUnknownModel`/`ErrNotFound`->404, `ErrUnknownField`/validation->400, `ErrConstraintViolation`->409.
- Keep response envelope `{status, data|items|id}` and avoid leaking internals.

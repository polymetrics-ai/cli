# Overview

Younium is a subscription billing platform. This bundle reads Younium accounts, subscriptions,
and invoices through the Younium REST API (`GET {base_url}/Accounts|Subscriptions|Invoices`). It
migrates `internal/connectors/younium` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip. Read-only: `capabilities.write` is `false`
and this bundle ships no `writes.json`.

## Auth setup

Provide `username` (config) and `password` (secret) for HTTP Basic auth
(`Authorization: Basic base64(username:password)`), matching legacy's `connsdk.Basic(username,
password)`. An optional `legal_entity` config value is sent as the `X-Younium-Legal-Entity` request
header when set; when unset, the header is omitted entirely (not sent empty) — `legal_entity` is
declared in `spec.json` but not in `required[]`, so the engine's conditional-header omission
semantics apply (`docs/migration/conventions.md` §3).

## Streams notes

All 3 streams (`accounts`, `subscriptions`, `invoices`) share the same shape: `GET` against the
Younium list endpoint (`/Accounts`, `/Subscriptions`, `/Invoices`), records at `data`, primary key
`["id"]`, cursor field `updated_at`. No pagination is declared — legacy issues a single unpaginated
request per stream and emits every record in the response's `data` array, so this bundle's
`streams.json` omits any `pagination` block (defaulting to `none`), matching legacy exactly.

`computed_fields` rename each raw field to the schema's snake_case name: `updated` -> `updated_at`
(all 3 streams), `invoiceNumber` -> `name` (invoices only, matching legacy's
`nameKeys: {"invoiceNumber", "number", "name"}` primary preference). `account_id` is derived from
the raw `accountId` field.

## Write actions & risks

None. Younium is modeled read-only in legacy (`capabilities.Write: false`); this bundle matches
that exactly and ships no `writes.json`.

## Known limits

- **Multi-key fallback chains are approximated by the primary key only.** Legacy's `mapRecord`
  tries several candidate raw field names in preference order for `id` (accounts:
  `{"id","accountId"}`, invoices: `{"id","invoiceId"}`), `name` (accounts:
  `{"name","accountName"}`, invoices: `{"invoiceNumber","number","name"}`), and `updated_at`
  (`{"updated","updatedAt","updated_at"}` on every stream) — only when the first-choice key is
  absent does it fall through to the next. The engine's `computed_fields` dialect has no
  coalesce/fallback filter (a single template resolves a single dotted path, hard-erroring or
  silently skipping on absence, never trying a second path), so this bundle wires only each field's
  first-preference legacy key (`id`, `name`/`invoiceNumber`, `updated`) — the response shape
  Younium's REST API actually returns in practice. A response using ONLY a fallback-only key
  (e.g. `accountId` present but `id` absent) would silently drop that field here where legacy would
  have recovered it. Documented, not silently worked around; revisit if `ENGINE_GAP` recurrence
  (a coalesce/first-non-null filter) crosses the §6 threshold.
- Full Younium API surface (orders, products, plans, usage) is out of scope for wave2; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries.

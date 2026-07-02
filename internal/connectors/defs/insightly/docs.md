# Overview

Insightly is a read-only declarative-HTTP migration (wave2 fan-out) of
`internal/connectors/insightly` (legacy Go package `insightly`). It reads Insightly CRM contacts,
organisations, opportunities, leads, projects, and tasks through the Insightly REST API v3.1
(`https://api.<pod>.insightly.com/v3.1`). This bundle targets capability parity with the legacy
connector; the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide an Insightly API token via the `token` secret; it is sent as the HTTP Basic auth username
with a blank password (`auth: [{"mode":"basic","username":"{{ secrets.token }}","password":""}]`),
matching legacy's `connsdk.Basic(token, "")` exactly. It is never logged.

`base_url` is required-with-default (`https://api.na1.insightly.com/v3.1`, matching legacy's
default `na1` pod) rather than derived from a separate `pod` config value the way legacy computes
it (`fmt.Sprintf("https://api.%s.insightly.com/v3.1", pod)`): the engine's `spec.json` `"default"`
materialization mechanism only fills in a FIXED literal value for an absent key, it cannot derive
one config value from another at read/check time (`docs/migration/conventions.md` §3's "DERIVED
default" note — sentry's hostname-based URL and chargebee's site-based URL hit the identical
limitation). An account on a pod other than `na1` must set `base_url` explicitly to
`https://api.<their-pod>.insightly.com/v3.1` instead of a bare `pod` value; this is a documented
config-surface narrowing, not a behavior change for any account already on the default pod.

## Streams notes

All 6 streams (`contacts`, `organisations`, `opportunities`, `leads`, `projects`, `tasks`) share
the same shape: `GET` against the Insightly PascalCase resource path, a bare top-level JSON array
response (`records.path: ""`), and `top`/`skip` offset-limit pagination
(`pagination.type: offset_limit`, `limit_param: top`, `offset_param: skip`, `page_size: 100`,
matching legacy's `insightlyDefaultPageSize`) — the next page advances `skip` by `page_size` and
stops when a page returns fewer than `page_size` records, identical to legacy's `harvest` loop.

Every raw Insightly object exposes its primary key as a resource-specific SCREAMING_SNAKE_CASE
field (`CONTACT_ID`, `ORGANISATION_ID`, etc.); `computed_fields` maps each into BOTH a normalized
`id` and the resource-specific snake_case name (`contact_id`, `organisation_id`, etc.), matching
legacy's record mappers exactly (legacy's own `insightlyContactRecord` etc. set both `"id"` and
`"contact_id"` from the same raw `CONTACT_ID` field). Every other field is likewise renamed from
its raw SCREAMING_SNAKE_CASE key to the schema's snake_case name via a bare single-reference
`computed_fields` entry, which the engine's typed-extraction rule preserves the RAW JSON type for
(numeric ids stay integers, `OPPORTUNITY_VALUE` stays a number, `CONVERTED`/`COMPLETED` stay
booleans) — no stringify-widening needed.

Primary key is `id` on every stream; `x-cursor-field: date_updated_utc` matches legacy's
`CursorFields` on every stream. Legacy never actually sends a server-side incremental filter
param for any Insightly stream (the API's `top`/`skip` pagination has no date-filter query
parameter used here), so no `incremental` block is declared on any stream, matching legacy's own
full-refresh-only behavior.

## Write actions & risks

None. Insightly is read-only in this bundle (`capabilities.write: false`); legacy also rejects
every write with `connectors.ErrUnsupportedOperation`. No `writes.json` is shipped.

## Known limits

- Full Insightly API surface (emails, events, notes, pipelines, custom fields, writes) is out of
  scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass
  B capability expansion"}` entries. Only the 6 legacy-parity streams are implemented.
- **`pod`-derived `base_url` is not modeled** (see Auth setup above): only an explicit `base_url`
  override is supported; the convenience `pod` shorthand legacy computes from is dropped. Every
  account on the default `na1` pod is unaffected (the bundle's `base_url` default already resolves
  to the identical URL); an account on another pod must supply the full `base_url` instead of just
  a `pod` value.
- `page_size`/`max_pages` are not exposed as `spec.json` config: `pagination.page_size` is a
  static JSON integer with no `{{ }}` template support (unlike a `stream.Query` entry), so a
  runtime `config.page_size` value could never actually resize the `top` page — declaring it would
  be dead config a bundle author cannot wire to any real behavior (F6, `docs/migration/
  conventions.md`). The static `page_size: 100` matches legacy's own default exactly.
  `PaginationSpec.MaxPages` has the identical static-int limitation, so legacy's `max_pages`
  (0/`all`/`unlimited` = no cap) config is reproduced as the engine's default unbounded behavior
  (omitted `MaxPages`) rather than a runtime override.

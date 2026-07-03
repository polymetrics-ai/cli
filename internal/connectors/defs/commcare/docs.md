# Overview

CommCare is a mobile data-collection platform (Dimagi). This bundle reads forms and cases from a
CommCare HQ project space through the CommCare HQ API v0.5 (`GET
{base_url}/a/{project_space}/api/v0.5/<form|case>/`). It migrates
`internal/connectors/commcare` (the hand-written connector) at capability parity; the legacy
package stays registered and unchanged until wave6's registry flip.

The catalog entry for this connector (`source-commcare`) is correctly labeled a source with no
write capability, matching the legacy implementation exactly (`Capabilities.Write: false`,
`Write()` returns `ErrUnsupportedOperation`).

## Auth setup

Provide a CommCare HQ API key via the `api_key` secret; it is sent as an `ApiKey`-scheme
Authorization header (`Authorization: ApiKey <api_key>`, matching legacy's
`connsdk.APIKeyHeader("Authorization", key, "ApiKey ")`) and is never logged. Also provide
`project_space`, the CommCare HQ project space slug every stream path is scoped under
(`/a/{{ config.project_space }}/api/v0.5/...`) — legacy defaults an unset `project_space` to the
literal string `"project"` (`commcare.go:163-166`), but this bundle declares it `required` instead
of reproducing that fallback: a silent `"project"` default is very unlikely to resolve to a real
project space and would more likely surface as a confusing 404 than a useful default. `app_id` is
optional; when set, it scopes both streams via an `app_id` query parameter (legacy:
`commcare.go:106`).

## Streams notes

Both streams (`forms`, `cases`) hit `GET /a/{project_space}/api/v0.5/<form|case>/`, extracting
records from the top-level `objects` array (`records.path: "objects"`), matching legacy's
`connsdk.RecordsAt(resp.Body, "objects")`. Legacy's `Read` emits every record's raw fields verbatim
(`emit(connectors.Record(rec))`, no field-built mapping) — both streams declare
`"projection": "passthrough"` per `docs/migration/conventions.md` §8 rule 1, so every raw field
CommCare returns survives unfiltered, not just the primary-key/cursor fields the schema documents
for typing purposes.

Primary key is `id` for both streams; `forms`' cursor field is `received_on`, `cases`' is
`server_modified_on`, matching legacy's `Catalog()` `CursorFields` declarations exactly. Neither
stream declares an `incremental.request_param`: legacy's `Catalog()` publishes these cursor fields
for manifest-surface parity, but `Read()` itself never sends a server-side filter parameter derived
from them (no `modified_since`-style query key anywhere in `commcare.go`) — a full stream read
always happens, on both a fresh sync and a resumed one, exactly like `searxng`'s equivalent
declared-but-unfiltered cursor field.

## Pagination

Legacy paginates by following the response body's `meta.next` field, which CommCare's API returns
as a **relative path with an embedded query string** (e.g.
`/a/demo/api/v0.5/form/?offset=2&limit=2&app_id=app_1`), stopping when `meta.next` is absent or
empty (`commcare.go:120-125`). This bundle instead declares `pagination.type: offset_limit`
(`limit_param: limit`, `offset_param: offset`, `page_size: 100`), stopping on a short/empty final
page — a **documented parity deviation** (`docs/migration/conventions.md` §5): the engine's
`next_url` pagination type is the literal dialect match for a body-embedded next-page reference, but
its SSRF guard (`checkOrigin`) hard-rejects any next value that parses with an empty host, which is
exactly what a relative `meta.next` value is; `allow_cross_host: true` bypasses the guard entirely
rather than fixing the relative-URL case narrowly, which is a strictly worse trade than the
equivalent-in-practice `offset_limit` short-page stop. CommCare's HQ API pages exhaustively
(`offset`/`limit` are the same query params `meta.next` itself encodes) and always terminates with a
short/empty final page, so `offset_limit`'s stop condition and legacy's `meta.next`-absent stop
condition agree on every input legacy itself would accept — this never changes emitted record DATA
or cardinality for any real CommCare project, only the mechanical stop signal the engine inspects
(the same class of deviation as the jamf-pro `totalCount` ledger entry, `docs/migration/
conventions.md` §5 item 13). `limit=100` matches legacy's `defaultPageSize`.

## Write actions & risks

None. CommCare is read-only here (`capabilities.write: false`), matching legacy
(`Capabilities.Write: false`) exactly — `Write` returns `ErrUnsupportedOperation` on the legacy side
and this bundle ships no `writes.json` at all.

## Known limits

- **`page_size`/`max_pages` config-driven overrides are not modeled.** Legacy accepts optional
  `page_size` (default 100) and `max_pages` (default 100) config values, each validated as a
  positive integer (`intConfig`, `commcare.go:172-180`). The engine's `offset_limit` paginator's
  `PaginationSpec.PageSize` is a static JSON literal with no templating support (unlike
  `stream.Query`, which does support templated, optionally-absent values), so it cannot be wired to
  a runtime `config.*` value; there is likewise no declarative `MaxPages` override mechanism tied to
  a config key. `limit=100` (the fixed pagination page size) matches legacy's *default* exactly. This
  is a documented, accepted config-surface narrowing (`docs/migration/conventions.md` §5's
  meta-rule: it never changes emitted record DATA for any input legacy itself would accept at its
  own default) — declaring dead `page_size`/`max_pages` spec properties that no template consumes
  would itself violate the "declared config must be consumed" rule (F6).
- **`project_space`'s legacy default (`"project"`) is not modeled**; see Auth setup above —
  `project_space` is `required` here instead.
- Full CommCare HQ API surface (applications, users, form submission/receiver endpoints) is out of
  scope for this pass; see `api_surface.json`'s `excluded: {category: out_of_scope, reason: "Pass B
  capability expansion"}` entries.

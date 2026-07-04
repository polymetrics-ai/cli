# Overview

Spotler CRM is a wave2 fan-out migration from `internal/connectors/spotlercrm` (the legacy
hand-written connector this bundle replaces at capability parity). It reads Spotler CRM contacts,
accounts, opportunities, and tasks through the Spotler CRM API. Read-only; the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Two independent credential shapes, gated by `when` (first match wins), matching the
zendesk-support dual-auth golden pattern (`docs/migration/conventions.md` §3):

1. `access_token` (preferred when present) — an OAuth2 access token for the REAL, live CRM API v4
   (`https://apiv4.reallysimplesystems.com`), sent as `Authorization: Bearer <access_token>`.
   Required by every stream this Pass B pass adds (`activities`/`campaigns`/`cases`).
2. `api_key` (fallback) — sent as the `X-API-Key` header value with no prefix (`mode:
   api_key_header`, empty `prefix`), matching legacy's `connsdk.APIKeyHeader("X-API-Key", key,
   "")`. Used only by the 4 legacy-parity streams (`contacts`/`accounts`/`opportunities`/`tasks`),
   which target a DIFFERENT host (`config.base_url`, default `https://api.spotlercrm.com/api/v1`)
   than the real CRM API v4 — see Known limits for why this host/auth pairing does not actually
   reach a live server.

Both are optional at the `spec.json` level (neither is in `required`); `selectAuth` errors clearly
if BOTH are absent at read/check time, matching the zendesk-support precedent's own behavior when
no credential is configured at all.

## Streams notes

All 4 legacy-parity streams (`contacts`, `accounts`, `opportunities`, `tasks`) share the identical
shape: `GET`, records at `data`, `page_number` pagination (`page_param: page`, `size_param: limit`,
`start_page: 1`, `page_size: 100`) — matches legacy's `connsdk.PageNumberPaginator{PageParam:
"page", SizeParam: "limit", StartPage: 1, PageSize: pageSize}` with `defaultPageSize = 100` (note:
legacy names its size query parameter `limit` even though the pagination style is
page-number-based, not offset/limit — this bundle reproduces that exact query-param name). Primary
key is `id` for every stream, matching legacy's `PrimaryKey: []string{"id"}`.

Legacy performs no incremental/state-cursor filtering during `Read` — no stream declares an
`incremental` block.

### Pass B additions

Three new read streams, added against the REAL, live CRM API v4 (`https://
apiv4.reallysimplesystems.com`) — a completely different host, auth mechanism, and response
envelope than every legacy-parity stream above (see Known limits for the full discovery). All
three:

- Use an **absolute `stream.path`** (bypassing `config.base_url` entirely — the same sanctioned
  pattern `defillama`'s `stablecoins` stream and `squarespace`'s `contacts` stream use for a
  differently-hosted resource), since `config.base_url`'s existing default targets the WRONG host
  for these new streams and changing it would alter the 4 already-migrated streams' resolved
  request shape outside this task's scope.
- Carry a stream-level `conformance.skip_dynamic` marker, since an absolute-URL `stream.path` is
  never reachable by conformance's fixture-replay server (`connsdk.Requester.resolveURL` treats an
  `http(s)://`-prefixed path as already-absolute and never substitutes the replay server's origin
  for it).
- Extract records via `records.path: "list"` (the real API's own top-level array key) plus
  `computed_fields` reaching two levels deep — `{{ record.record.<field> }}` — since each array
  element is itself `{"metadata": {...}, "record": {...}}` (the real API's own per-item envelope,
  confirmed from its documented `GET /accounts` worked example), and only the nested `record`
  object holds the actual CRM fields. Every `computed_fields` entry here is a bare single
  reference, so the engine's typed-extraction rule applies: `id`/`ownerid` copy through as their
  real numeric JSON type, not stringified.
- Use `page_number` pagination (`page_param: page`, `size_param: limit`, `start_page: 1`) — the
  real API's own documented convention (`GET /accounts?limit=50&page=2`), with `metadata.has_more`
  as the real stop signal (this bundle relies on the engine's default short-page stop rather than a
  declared `stop_path`, since the real API's own docs describe `has_more` as accurate — unlike
  Zendesk's documented caveat that motivated `stop_path`'s addition — see
  `docs/migration/conventions.md` §3).

`activities`, `campaigns`, and `cases` are 3 of the real API's 11 documented object types
(`accounts`/`activities`/`contacts`/`campaigns`/`campaigndetails`/`campaignstages`/`cases`/
`documents`/`opportunities`/`opportunityhistories`/`opportunity_lines`, per the API's own overview
page). Their schemas declare only `id`/`ownerid`/`createddate`/`modifieddate` (`campaigns` also
`name`) — the common metadata fields confirmed from the API's own fully-enumerated `GET /accounts`
worked example — rather than a fuller field list, since the API's public documentation names these
objects but does not enumerate their own per-field shapes anywhere reachable in this pass's
research (the `/datadictionary/{object_name}` endpoint would return this authoritatively, but
requires a live, authenticated call this pass could not make). A minimal, confidently-correct
schema was judged better than a fuller but guessed one.

## Write actions & risks

None. The real CRM API v4 fully supports `POST`/`PATCH`/`DELETE` per its own documentation, but no
write action was added in this pass: every new stream above uses an absolute `stream.path` (the
only way to reach the real API host given `config.base_url`'s existing default targets a
different, incorrect host), and write actions have no `conformance.skip_dynamic` equivalent —
`write_request_shape`'s capture-server replay always points `b.HTTP.URL` at the test double, which
an absolute-URL `action.path` bypasses entirely, so a write action here could never be proven
correct by this repo's conformance harness. Shipping an untestable write action was judged worse
than not shipping it, matching `squarespace`'s own documented reasoning for the identical
host-mismatch problem (`create_contact`/`delete_contact`, evaluated and excluded there for the
same reason). `capabilities.write` stays `false`.

## Known limits

- **`page_size`/`max_pages` are not exposed as config properties (legacy-parity streams only).**
  Legacy accepts `config.page_size` (bounded 1-500, default 100) and `config.max_pages` (default
  unbounded) at read time. The engine's `PaginationSpec.PageSize`/`MaxPages` fields are static
  values baked into `streams.json`'s pagination block, with no `{{ }}` templating support from
  `config.*` (matching the split-io/spotify-ads/searxng precedent, F6 REVIEW.md). This bundle
  hard-codes `page_size: 100` (legacy's own default) and declares no `max_pages` (unbounded,
  matching legacy's own default). A caller that previously overrode either value per-run loses
  that capability; every default-config caller sees byte-identical behavior.
- **The wave2 legacy-parity streams' host, auth mechanism, and response envelope do not match any
  real, live Spotler CRM API, and this bundle does not correct them.** This is a pre-existing
  wave2 discrepancy discovered during this Pass B research pass, not introduced by it — the 4
  legacy-parity streams are left untouched at their already-migrated shape per the meta-rule
  against altering already-migrated behavior:
  - `contacts`/`accounts`/`opportunities`/`tasks` request `{{ config.base_url }}` (default
    `https://api.spotlercrm.com/api/v1`) with an `X-API-Key` header and expect a flat `{"data":
    [...]}` response envelope. "Spotler CRM" is a rebrand of Really Simple Systems; its real,
    documented API is CRM API v4, served from `https://apiv4.reallysimplesystems.com`, authenticated
    with an OAuth2 Bearer access token, and returning a nested `{"metadata": {...}, "list":
    [{"metadata": {...}, "record": {...}}]}` envelope (confirmed directly from the API's own
    published documentation at `https://support.reallysimplesystems.com/api-v4/`) — none of which
    match the wave2 bundle's assumptions. `config.base_url`'s default host does not correspond to
    any documented Spotler/Really Simple Systems API endpoint this pass could find.
  - The real API also has **no `tasks` object at all** — its documented catalog is
    `accounts`/`activities`/`contacts`/`campaigns`/`campaigndetails`/`campaignstages`/`cases`/
    `documents`/`opportunities`/`opportunityhistories`/`opportunity_lines`. The wave2 `tasks`
    stream's target object does not exist on the real API under any name this pass could confirm.
  - Legacy's own `spotlercrm.go` invented these assumptions wholesale (there is no evidence in the
    repository that it was ever validated against a live Spotler CRM account); this bundle
    faithfully reproduces legacy's assumptions rather than silently correcting them outside this
    task's scope.
  - The NEW `activities`/`campaigns`/`cases` streams added in this pass were authored directly
    against the live, documented CRM API v4 and do not inherit any of the above discrepancies —
    they use the correct host, the correct OAuth2 Bearer auth (via the new `access_token` secret),
    and the correct nested response envelope.
- **5 of the real API's 11 documented objects are not covered** (`campaigndetails`/
  `campaignstages`/`documents`/`opportunityhistories`) for lack of a confidently-sourced per-field
  schema, and `opportunity_lines` because it is reachable inline via the `opportunities` object's
  own documented `lines=true` request parameter rather than needing a separate stream — see
  `api_surface.json` for the full per-endpoint reasoning.

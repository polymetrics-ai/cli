# Overview

Zoho Books is a wave2 fan-out migration from `internal/connectors/zoho-books` (the hand-written
Go connector it replaces). It reads Zoho Books contacts, invoices, and items through the Zoho
Books REST API v3. Read-only, matching legacy's capabilities exactly (`Write` returns
`ErrUnsupportedOperation`). The legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Zoho OAuth access token via the `access_token` secret. It is sent as the `Authorization`
header with legacy's exact non-standard prefix (`Zoho-oauthtoken <access_token>`, NOT the standard
`Bearer <token>` shape) via `streams.json` base `auth`'s `api_key_header` mode
(`header: Authorization`, `prefix: "Zoho-oauthtoken "`) — never logged. An optional
`organization_id` config value is sent as the `organization_id` query parameter on every stream
request when set; when unset (the config key genuinely absent), it is omitted entirely (not sent
empty), matching legacy's `baseQuery` behavior. A caller that explicitly sets `organization_id` to
an empty string (rather than omitting the key) is a narrow edge case where this bundle diverges
from legacy: legacy's `configValue` treats an explicit empty string identically to an absent key
(both omit the param), while the engine's `omit_when_absent` only fires on a genuinely-missing
config key — an explicit `""` would resolve and be sent as a literal empty `organization_id=`
query value. No realistic operator config form produces an explicit-empty-string override for an
optional field (it either omits the key or a non-empty value), so this is accepted as a
theoretical, not practical, divergence.

## Streams notes

All 3 streams (`contacts`, `invoices`, `items`) share the same shape: `GET` against the Zoho Books
list endpoint, records extracted at the stream's own top-level array key (`contacts`/`invoices`/
`items`), `page_number` pagination (`page`/`per_page` query params, page size 200, stop on a short
page — identical to legacy's `harvest` loop's default behavior).

Every stream uses `projection: "passthrough"` (every raw API field survives, matching legacy's
`mapRecord`, which copies every input field verbatim) plus a `computed_fields` alias for each
stream's authoritative primary-key/name/cursor field to the parity fields `id`/`name`/`updated_at`
legacy also synthesizes (`contact_id`->`id`, `contact_name`->`name`, `last_modified_time`->
`updated_at` for `contacts`; equivalent aliases for `invoices`/`items`). `computed_fields`' bare
`{{ record.<path> }}` shape gets typed extraction (native JSON type preserved) and is silently
skipped when the source path is absent on a given record, matching legacy's own
`out["id"] == nil` fallback-only-if-absent semantics for the common case where the field is
already present.

## Write actions & risks

None. Legacy `zoho-books` is read-only (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

- `spec.json` declares `page_size`/`max_pages` (matching legacy's own tunable config keys and
  defaults) for documentation continuity with the legacy connector, but — exactly like the
  stripe golden's identical `page_size`/`max_pages` properties (`docs/migration/conventions.md`
  ledger item 3) — neither is wired to any template in this bundle: `PaginationSpec.PageSize`/
  `MaxPages` are static ints declared directly in `streams.json`'s `base.pagination` block (no
  `config.*` reference), and the engine's pagination spec has no per-request template support for
  either field. A caller setting `config.page_size`/`config.max_pages` has no runtime effect,
  identical to stripe's own documented "informational, not enforced" precedent. This bundle uses
  legacy's real defaults (200 records/page, unbounded pages) as the static values, so behavior is
  unchanged for every caller that relied on legacy's defaults; a caller that previously overrode
  `page_size`/`max_pages` away from the default loses that override capability. Documented,
  accepted scope narrowing — not a silent behavior change for the common (default-config) case.
- Legacy's `firstValue` fallback tried MULTIPLE candidate keys per derived field in priority order
  (e.g. `contacts.id`: `contact_id`, then bare `id`) to guard against alternate/legacy API response
  shapes. This bundle's `computed_fields` aliases only the FIRST (authoritative, real-wire-shape)
  candidate key per field — the engine's computed_fields dialect has no multi-key fallback-chain
  primitive (each entry is a single template, not an ordered candidate list). The secondary
  fallback keys are legacy defensive coding for a shape the real, documented Zoho Books v3 API
  never actually emits (confirmed against `docs.md`'s `docs_url` and the legacy connector's own
  tests, which only ever exercise the first key). Documented scope narrowing, not a silent
  behavior change for any real Zoho Books response. See
  `docs/migration/conventions.md`'s parity-deviation ledger meta-rule.
- Full Zoho Books API surface (estimates, credit notes, customer payments, bills, expenses, chart
  of accounts) is out of scope for this wave; see `api_surface.json`'s `excluded: {category:
  out_of_scope}` entries. Only the 3 legacy-parity streams are implemented.
- `fixtures/streams/contacts/{page_1,page_2}.json` is the required 2-page pagination proof (200
  full-size records on page 1 to trigger a genuine second request under the real `page_size`
  default, 1 record on page 2 to stop) — `invoices`/`items` ship single-page fixtures, matching the
  stripe golden's "only the first declared stream proves 2-page termination" pattern.

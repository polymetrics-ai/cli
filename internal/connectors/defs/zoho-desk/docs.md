# Overview

Zoho Desk is a wave2 fan-out migration from `internal/connectors/zoho-desk` (the hand-written Go
connector it replaces). It reads Zoho Desk tickets, contacts, and accounts through the Zoho Desk
REST API v1. Read-only, matching legacy's capabilities exactly (`Write` returns
`ErrUnsupportedOperation`). The legacy package stays registered and unchanged until wave6's
registry flip.

## Auth setup

Provide a Zoho OAuth access token via the `access_token` secret. It is sent as the `Authorization`
header with legacy's exact non-standard prefix (`Zoho-oauthtoken <access_token>`, NOT the standard
`Bearer <token>` shape) via `streams.json` base `auth`'s `api_key_header` mode
(`header: Authorization`, `prefix: "Zoho-oauthtoken "`) — never logged. An optional `org_id`
config value is sent as the `orgId` header on every request when set (matching legacy's
`DefaultHeaders["orgId"]`); when unset (the config key genuinely absent), the header is omitted
entirely (not sent empty) — the same optional-header shape as stripe's `Stripe-Account`. As with
`zoho-books`/`zoho-expense`'s `organization_id` query param, a caller explicitly setting `org_id`
to an empty string (rather than omitting the key) is a narrow, impractical edge case where this
diverges from legacy's identical-treatment-of-absent-and-empty `configValue` helper; see those
bundles' docs for the full explanation.

## Streams notes

All 3 streams (`tickets`, `contacts`, `accounts`) share the same shape: `GET` against the Zoho
Desk list endpoint, records extracted at the shared envelope key `data`. Pagination is
`offset_limit` (`from`/`limit` query params, page size 100, stop on a short page) — this is Zoho
Desk's real documented offset-based pagination convention and drives the same page-advance
arithmetic as legacy's `pageQuery` (`from = (page-1)*size`).

Every stream uses `projection: "passthrough"` (every raw API field survives, matching legacy's
`mapRecord`, which copies every input field verbatim) plus a `computed_fields` alias for each
stream's authoritative name/cursor field to the parity fields `name`/`updated_at` legacy also
synthesizes (`subject`->`name`, `modifiedTime`->`updated_at` for `tickets`; equivalent aliases for
`contacts`/`accounts`). Unlike zoho-books/zoho-campaign/zoho-expense, Zoho Desk's real API already
emits a literal `id` field on every record, so the `id` computed_fields entry is a same-value
alias (documents intent/parity explicitly rather than relying on passthrough alone).
`computed_fields`' bare `{{ record.<path> }}` shape gets typed extraction (native JSON type
preserved) and is silently skipped when the source path is absent on a given record, matching
legacy's own `out["id"] == nil` fallback-only-if-absent semantics.

## Write actions & risks

None. Legacy `zoho-desk` is read-only (`Capabilities.Write: false`); this bundle ships no
`writes.json`.

## Known limits

- Legacy's `pageQuery` sent FOUR query params on every page request: `from`/`limit` (the real
  Zoho Desk offset-pagination convention) AND `page`/`per_page` (a redundant, defensive pair the
  documented Zoho Desk API does not require for correct pagination). The engine's `offset_limit`
  paginator type sends exactly one param pair (`limit_param`/`offset_param`); there is no
  dialect mechanism to send two independently-shaped paired params (one page-number-based, one
  offset-based) simultaneously from a single pagination block. This bundle sends only `from`/
  `limit` — the pair that actually drives real Zoho Desk pagination — and drops the inert `page`/
  `per_page` duplicate params. This never changes emitted record DATA (identical records, identical
  page-advance/stop behavior) for any real Zoho Desk server, matching the "informational vs.
  enforced" pattern already accepted for stripe's dead `page_size`/`max_pages` pagination fields
  (`docs/migration/conventions.md` ledger item 3). ACCEPTABLE per the parity-deviation meta-rule.
- `spec.json` declares `page_size`/`max_pages` (matching legacy's own tunable config keys and
  defaults) for documentation continuity with the legacy connector, but — exactly like the
  stripe golden's identical `page_size`/`max_pages` properties (`docs/migration/conventions.md`
  ledger item 3) — neither is wired to any template in this bundle: `PaginationSpec.PageSize`/
  `MaxPages` are static ints declared directly in `streams.json`'s `base.pagination` block (no
  `config.*` reference), and the engine's pagination spec has no per-request template support for
  either field. A caller setting `config.page_size`/`config.max_pages` has no runtime effect,
  identical to stripe's own documented "informational, not enforced" precedent. This bundle uses
  legacy's real defaults (100 records/page, unbounded pages) as the static values, so behavior is
  unchanged for every caller that relied on legacy's defaults; a caller that previously overrode
  `page_size`/`max_pages` away from the default loses that override capability. Documented,
  accepted scope narrowing — not a silent behavior change for the common (default-config) case.
- Legacy's `firstValue` fallback tried MULTIPLE candidate keys per derived field in priority order
  (e.g. `tickets.name`: `subject`, then `ticketNumber`, then bare `name`) to guard against
  alternate/legacy API response shapes. This bundle's `computed_fields` aliases only the FIRST
  (authoritative, real-wire-shape) candidate key per field — the engine's computed_fields dialect
  has no multi-key fallback-chain primitive. The secondary fallback keys are legacy defensive
  coding for a shape the real, documented Zoho Desk API never actually emits. Documented scope
  narrowing, not a silent behavior change for any real Zoho Desk response.
- Full Zoho Desk API surface (agents, departments, ticket threads/comments, ticket writes) is out
  of scope for this wave; see `api_surface.json`'s `excluded: {category: out_of_scope}` entries.
  Only the 3 legacy-parity streams are implemented.
- `fixtures/streams/tickets/{page_1,page_2}.json` is the required 2-page pagination proof (100
  full-size records on page 1 to trigger a genuine second request under the real `page_size`
  default, 1 record on page 2 to stop) — `contacts`/`accounts` ship single-page fixtures, matching
  the stripe golden's "only the first declared stream proves 2-page termination" pattern.

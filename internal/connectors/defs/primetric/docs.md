# Overview

Primetric is a wave2 fan-out declarative-HTTP migration. It reads Primetric employees, projects,
clients, and roles through the Primetric REST API v1 (`GET https://api.primetric.com/api/v1/...`)
using OAuth2 client-credentials authentication. This bundle targets capability parity with
`internal/connectors/primetric` (the hand-written connector it migrates); the legacy package stays
registered and unchanged until wave6's registry flip.

## Auth setup

Provide `client_id`/`client_secret` secrets. The engine's `oauth2_client_credentials` auth mode
fetches and caches a bearer token from `token_url` (default
`https://api.primetric.com/oauth/token`, matching legacy's `defaultTokenURL` constant exactly —
legacy derives `token_url` independently of `base_url`, so this default is a fixed literal, not a
`base_url`-derived one, per `conventions.md`'s "derived default" guidance) using the
`client_credentials` grant, matching legacy's `connsdk.OAuth2ClientCredentials` (`primetric.go:151-158`)
field-for-field, including the `scopes: "read"` value legacy hard-codes (`Scopes: []string{"read"}`).
`base_url` defaults to `https://api.primetric.com/api/v1`.

## Streams notes

All 4 streams (`employees`, `projects`, `clients`, `roles`) read from the identical shape: records
live at the `data` key, and pagination follows Primetric's own `page` query parameter with no
size parameter ever sent (matching legacy's `url.Values{"page": [...]}}` exactly — legacy's
`defaultPageSize` constant is dead code, unused in the actual request per its own `var _ =
defaultPageSize` blank-assignment).

**Pagination stop condition is an approximation, documented per `conventions.md`'s freshsales
precedent.** Legacy's real stop condition reads an explicit `meta.total_pages` field
(`page >= totalPages`), not a short-page heuristic (`primetric.go:119-125`). The engine's
`page_number` pagination type has no equivalent "stop on an explicit total-count field" signal —
its only stop signal is a short/empty page (`recordCount < PageSize`). This bundle declares
`page_size: 50` (legacy's own `defaultPageSize` constant, its documented assumption of Primetric's
real fixed page size) as the short-page stop threshold. For any full page Primetric's real API
actually returns, this is data-identical to legacy's `total_pages`-driven stop: a page shorter than
the true server-side page size genuinely IS the last page under both schemes, and the engine never
requests fewer pages than legacy would (see Known limits for the corner case where this could
diverge).

`employees`' `name` field models legacy's 3-tier fallback (`item["name"]`, then
`item["full_name"]`, then `TrimSpace(first_name + " " + last_name)`, `primetric.go:172-188`) using
only its 3rd branch: `computed_fields`' dialect has no multi-key coalesce filter, and legacy's own
test fixture (`primetric_test.go`) only ever exercises `first_name`/`last_name` on real employee
records (no `name`/`full_name` key), so the 3rd branch is the only one with concrete ground truth.
`projects`, `clients`, and `roles` are modeled with a direct `name` schema-projection (no
computed_fields needed) — Primetric's public docs site is a JS-rendered SPA this migration could
not fetch structured field definitions from (a `DOCS_UNREACHABLE`-adjacent limitation for
field-shape specifics only, not a full blocker), so this bundle relies on legacy's own object
model (projects/clients/roles are not person-shaped entities, unlike employees) to choose the most
conservative, least-surprising mapping — legacy's shared generic `mapRecord` does not actually
disambiguate between these streams' real shapes either.

**Every dynamic (fixture-replay) conformance check is skipped at the bundle level** (see
`metadata.json`'s `conformance.skip_dynamic`): `oauth2_client_credentials`'s `token_url` is a
config value independent of `base_url` (unlike dwolla's `{{ config.base_url }}/token` shape, which
naturally follows the replay server), so conformance's synthetic
`"synthetic-conformance-value"` config fill-in is not a resolvable URL — the token exchange fails
before any declarative request is issued. This mirrors `sendpulse`'s identical, already-accepted
shape exactly.

## Write actions & risks

None. Primetric has no reverse-ETL writes modeled by legacy (`Write` returns
`connectors.ErrUnsupportedOperation`); `capabilities.write` is `false` and this bundle ships no
`writes.json`.

## Known limits

- **Pagination stop condition is a short-page approximation, not an exact `total_pages` port.**
  See Streams notes above. The one corner case where this could genuinely diverge from legacy: if
  Primetric's real page size is NOT exactly 50 (this bundle's `page_size` threshold, taken from
  legacy's own `defaultPageSize` constant — the only documented assumption available, since
  Primetric's real fixed page size could not be confirmed from the (JS-rendered, unreachable)
  public docs site), a full page shorter than 50 would incorrectly be treated as the last page
  even if `total_pages` says otherwise. This is a data-completeness risk (missing trailing pages),
  not a duplicate/corrupt-data risk. A future wave with live Primetric API access should confirm
  the true fixed page size and correct this threshold if it differs from 50.
- **`employees`' `name` field does not model legacy's `name`/`full_name` raw-key precedence.**
  Only the `first_name`+`last_name` concatenation branch is modeled (see Streams notes); if a real
  employee record ever carries a raw `name` or `full_name` key, legacy would prefer it but this
  bundle would still emit the concatenated form. Not exercised by legacy's own test suite either.
- **`projects`/`clients`/`roles`' `name` field mapping is a best-effort default**, not confirmed
  against live Primetric API responses (docs site unreachable for structured field definitions).
  If Primetric's real objects use a different field name for these entities' display name, this
  bundle's schema projection would silently emit `null` for `name` rather than the real value — a
  data-completeness gap, not a corruption. A future wave with live API access or a captured real
  response should confirm and correct if needed.
- **All dynamic (fixture-replay) conformance checks are skipped** (bundle-level marker, see Streams
  notes). Fixtures are still fully authored (2-page pagination shape per stream) to document the
  intended real wire shape and satisfy static presence checks; parity is asserted here by
  structural comparison against legacy source, not by a live replay test.
- **Legacy's `raw` escape-hatch field is not modeled.** `mapRecord` stamps a full copy of the
  source item onto every record under the key `raw` (`primetric.go:181`). The engine's
  `computed_fields` dialect has no whole-record reference primitive, so this cannot be expressed.
- **Legacy's fixture-mode-only synthetic records are not modeled.** Legacy's `readFixture` path
  (only reached when `config.mode == "fixture"`) emits synthetic records with a different id type
  (integer literal `i`, not derived from any real API field) than the live path; this bundle
  targets the live path only, matching every other wave1/wave2 bundle's documented precedent.

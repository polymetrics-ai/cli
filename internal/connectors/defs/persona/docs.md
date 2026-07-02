# Overview

Persona is a wave2 fan-out declarative-HTTP migration. It reads Persona inquiries, accounts,
reports, transactions, and cases through core JSON:API list endpoints
(`GET https://api.withpersona.com/api/v1/...`). This bundle is engine-vs-legacy parity-tested
against `internal/connectors/persona` (the hand-written connector it migrates); the legacy package
stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Persona API key via the `api_key` secret; it is sent as a Bearer token (`Authorization:
Bearer <api_key>`) and is never logged, matching legacy's `connsdk.Bearer(key)`
(`persona.go:151`). `base_url` defaults to `https://api.withpersona.com/api/v1` and may be
overridden for tests/proxies.

## Streams notes

All 5 streams (`inquiries`, `accounts`, `reports`, `transactions`, `cases`) share the same shape:
`GET` against the Persona JSON:API list endpoint, records at the top-level `data` key, primary key
`["id"]`, and JSON:API's own `id`/`type`/`attributes`/`relationships` object properties — matching
legacy's `streams()` field set exactly (`persona.go:127`). Each stream sends `page[size]=50` on the
first request (legacy's own default `pageSize`, `defaultPageSize = 50`). Pagination follows
Persona's `links.next` absolute-URL convention (`pagination.type: next_url`, `next_url_path:
"links.next"`), matching legacy's own manual loop that follows `resp.Body`'s `links.next` field
verbatim until it is empty (`persona.go:104-111`).

Legacy declares `CursorFields: []string{"attributes.updated-at"}` on every stream's `Catalog`
metadata (`persona.go:127`), but never actually implements a cursor-filtered/incremental read path
— there is no request parameter or client-side filter applied against that field anywhere in
`persona.go`'s `Read`; every read is a full-refresh traversal of `links.next`. This bundle
therefore declares no `incremental` block for any stream (matching legacy's REAL read behavior,
not its unused catalog metadata) and no `x-cursor-field` on any schema (per conventions.md §2:
`x-cursor-field` is declared "when the stream is incremental" — none of these streams are).

## Write actions & risks

None. Legacy `persona.Write` always returns `connectors.ErrUnsupportedOperation`;
`capabilities.write` is `false` and this bundle ships no `writes.json`.

## Known limits

- **`max_pages` is not runtime-configurable.** Legacy exposes a `max_pages` config-driven override
  (`persona.go:159-161`, `optionalInt`) that caps the manual pagination loop. The engine's
  `next_url` paginator has no analogous config-driven page-count knob (it never reads
  `PaginationSpec.MaxPages`), so this bundle does not declare `max_pages` in `spec.json` at all
  (F6, REVIEW.md: a declared-but-unwireable config key is worse than an absent one) — matching
  bitly's identical, already-accepted limitation (`docs/migration/conventions.md`, bitly's
  `docs.md`). Pagination is bounded only by the short/empty `links.next` stop signal, matching
  Persona's own real termination behavior.
- **Fixtures are single-page** for every stream, per `docs/migration/conventions.md` §4's
  sanctioned `next_url` exception: the next-page URL is the replay server's own runtime address and
  cannot be embedded in a static fixture file. Every fixture's `links.next` is `null`, so
  `pagination_terminates` (which runs against the bundle's first eligible stream, `inquiries`)
  correctly observes exactly one request for one fixture page and terminates. Real 2-page
  `links.next` correctness is proven by legacy's own `persona_test.go`'s
  `TestReadInquiriesPaginatesAndAuthenticates` (a live `httptest.Server` asserting the second page
  is requested via the exact absolute `links.next` URL); this bundle's declarative engine path uses
  the identical `next_url` mechanism bitly's `bitlinks` stream already exercises in production.
- Full Persona API surface (inquiry creation/resume, verifications) is out of scope; see
  `api_surface.json`'s `excluded: {category: out_of_scope}` entries — legacy itself never
  implemented these.

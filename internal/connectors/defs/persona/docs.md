# Overview

Persona is a wave2 fan-out declarative-HTTP migration, expanded in Pass B beyond legacy parity. It
reads Persona inquiries, accounts, reports, transactions, and cases through core JSON:API list
endpoints (`GET https://api.withpersona.com/api/v1/...`), and now also performs 13 lifecycle write
actions. The 5 read streams are engine-vs-legacy parity-tested against
`internal/connectors/persona` (the hand-written connector this bundle migrates); the legacy package
stays registered and unchanged until wave6's registry flip. All 13 write actions have no legacy
counterpart — legacy `Write` always returned `ErrUnsupportedOperation` — so they carry no parity
constraint. This bundle's write/surface research is grounded directly in Persona's live-published
OpenAPI 3.1 spec (`https://docs.withpersona.com/openapi/api-reference.json`, fetched during this
review; 169 distinct paths / 204 method+path operations), not secondary sources.

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

Legacy's `Read` decodes each page's `data` array with `connsdk.RecordsAt` and emits every record
verbatim (`emit(connectors.Record(rec))`, `persona.go:99-103`) — there is no `mapRecord`-style
field-building or filtering anywhere in the read path. Every stream therefore declares
`"projection": "passthrough"` (conventions.md §8 rule 1) so the engine's default schema-mode
projection does not silently drop any JSON:API attribute/relationship field legacy would have
passed through unfiltered. The `id`/`type`/`attributes`/`relationships` properties in each
`schemas/*.json` remain a documentation surface of the well-known JSON:API envelope shape, not an
allow-list.

Legacy declares `CursorFields: []string{"attributes.updated-at"}` on every stream's `Catalog`
metadata (`persona.go:127`), but never actually implements a cursor-filtered/incremental read path
— there is no request parameter or client-side filter applied against that field anywhere in
`persona.go`'s `Read`; every read is a full-refresh traversal of `links.next`. This bundle
therefore declares no `incremental` block for any stream (matching legacy's REAL read behavior,
not its unused catalog metadata) and no `x-cursor-field` on any schema (per conventions.md §2:
`x-cursor-field` is declared "when the stream is incremental" — none of these streams are).

## Write actions & risks

**Pass B addition** (no legacy counterpart; legacy `persona.Write` always returned
`connectors.ErrUnsupportedOperation`). Persona's live OpenAPI spec (fetched directly during this
Pass B review) is large (169 paths), but almost every create/update endpoint nests its request body
under a JSON:API `data` and/or `meta` envelope key that the engine's declarative `WriteAction`
dialect cannot express (see Known limits' `ENGINE_GAP` writeup). The 13 actions below are exactly
the mutations confirmed, per-endpoint from the live requestBody schema, to have no required
top-level body field — an absent/empty JSON body is a genuinely valid request — so they ARE fully
dialect-expressible:

- `redact_inquiry` / `redact_account` / `redact_case` / `redact_report` / `redact_transaction`
  (`DELETE /{resource}/{{ record.id }}`, `body_type: none`, `delete.missing_ok_status: [404]`):
  Persona's own docs state redaction "permanently and irreversibly" deletes all PII on the target
  object ("This action cannot be undone"). Destructive; `confirm: destructive`; approval required
  before executing.
- `approve_inquiry` / `decline_inquiry` / `expire_inquiry` / `resume_inquiry`
  (`POST /inquiries/{{ record.id }}/{action}`, `body_type: none`): Inquiry lifecycle-state
  transitions. `approve`/`decline` finalize a verification decision and trigger any workflows/
  webhooks tied to that transition (their live schema has an optional `meta` property, but nothing
  `required`, so an empty body is valid); `expire` ends an in-progress flow; `resume` re-opens a
  previously expired/paused one (neither declares any requestBody at all). Approval required before
  executing (`metadata.json.risk.approval`).
- `rerun_report` (`POST /reports/{{ record.id }}/run`, `body_type: none`): re-runs a continuously
  monitored Report immediately, outside its normal recurrence schedule. A metered, billed external
  side-effecting action per Persona's own docs; approval required.
- `pause_report_monitoring` / `resume_report_monitoring` (`POST /reports/{{ record.id
  }}/pause|resume`, `body_type: none`): pauses/resumes continuous monitoring on a Report (Persona's
  own docs: "Requires additional permissions"). Approval required.
- `redact_transaction_biometrics` (`POST /transactions/{{ record.id }}/redact-biometrics`,
  `body_type: none`): permanently and irreversibly deletes ONLY the biometric data for a Transaction
  and its associated objects (narrower than `redact_transaction`, which redacts the whole record).
  Destructive; `confirm: destructive`; approval required.

Every OTHER Persona mutation (create/update an Inquiry/Account/Case/Report/Transaction, tag
management, case assignment/status/objects, account relations/consolidation, transaction
relations/labels, and more) is **not implemented** — see Known limits (`ENGINE_GAP`).

## Known limits

- **`ENGINE_GAP`: every Persona create/update endpoint, and most sub-resource action endpoints,
  nest their request body under a JSON:API envelope the `WriteAction` dialect cannot express.**
  Confirmed directly from Persona's live OpenAPI schema for each endpoint (not a secondary source):
  creating or updating an Inquiry/Account/Case/Report/Transaction requires a top-level `data`
  property (with a further `meta` property on most create endpoints); tag management (add/remove/
  set-tags on every resource), case assign/set-status/mark-for-review, and Persona-object add/
  remove all require a nested `meta` or `data` object. The engine's `WriteAction` dialect always
  sends either the record's own fields directly (default JSON body) or an explicit `body_fields`
  allow-list of top-level record keys — there is no mechanism to wrap the outgoing body under a
  named nested envelope key (no `StreamSpec.Body`-style wrapper exists for writes). Expressing any
  of these correctly would need either a new hook interface (forbidden this pass — Tier-2 is capped
  at the 5 existing interfaces and this bundle adds none) or an engine dialect addition (a
  body-envelope/wrapper spec) not present today. Every affected endpoint is recorded in
  `api_surface.json` as `out_of_scope` with this exact `ENGINE_GAP` reasoning rather than guessed or
  silently flattened (flattening the envelope away would send a request Persona's real API rejects,
  not a parity-safe approximation).
- Several other endpoints (verifications/documents detail-by-id, devices, org-wide events/audit
  logs, webhooks, API keys, lists, importers, inquiry-sessions, workflows, Graph, Connect, Relay,
  OAuth) are excluded as `out_of_scope`/`requires_elevated_scope`/`non_data_endpoint`/`duplicate_of`
  — see `api_surface.json` for the full per-endpoint breakdown and reasoning; none of these are
  modeled by legacy and most are their own distinct multi-endpoint subsystems or admin-scoped
  operations outside this bundle's core verification/identity-data focus.
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

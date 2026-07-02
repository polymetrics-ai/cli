# Overview

Klaus API is a wave2 fan-out declarative-HTTP migration. It reads Klaus (Zendesk QA) users and rating
categories through the Klaus public REST API v2 (default `https://kibbles.klausapp.com/api/v2`). This
bundle migrates `internal/connectors/klaus-api` (the hand-written connector); the legacy package stays
registered and unchanged until wave6's registry flip. **The `reviews` stream is NOT migrated in this
wave** — see "Known limits" below for the specific engine gap blocking it. Klaus is read-only here —
legacy has no reverse-ETL write set — so `capabilities.write` is `false` and no `writes.json` is
shipped.

## Auth setup

Provide a Klaus API key via the `api_key` secret, sent as a Bearer token
(`auth: [{"mode": "bearer", "token": "{{ secrets.api_key }}"}]`), matching legacy's
`connsdk.Bearer(secret)` (`klausapi.go:276`) exactly — never logged. `account` (a Klaus account id,
required) is templated into every stream's path (`/account/{{ config.account }}/...`), matching
legacy's `klausAccount`/account-scoped path construction. `workspace` (required only for the
`categories` stream, which is workspace-scoped: `/account/{account}/workspace/{workspace}/categories`)
is templated the same way; the `users` stream never references it. `base_url` defaults to
`https://kibbles.klausapp.com/api/v2` and may be overridden for tests/proxies.

## Streams notes

Two streams are migrated, matching 2 of legacy's 3 `klausStreamDefs` entries:

- `users` — `GET /account/{account}/users`, records at `users`, unpaginated (matches legacy's
  `harvestSingle`). Flat field-for-field passthrough (`id`/`name`/`email`); no `computed_fields`
  needed.
- `categories` — `GET /account/{account}/workspace/{workspace}/categories`, records at `categories`,
  unpaginated. Schema and projection match legacy's `klausCategoryRecord` mapper exactly, including the
  two array-valued fields (`rootCauses`, `scorecards`) legacy's mapper emits but its own `Field` catalog
  (`klausCategoryFields()`) omits — schema-as-projection is derived from what the mapper actually
  emits, per `docs/migration/conventions.md` §2, not from the (incomplete) legacy catalog.

Neither migrated stream is incremental — Klaus's users/categories endpoints support full-refresh sync
only, matching legacy (no `CursorFields` declared on either legacy stream).

**`reviews` is NOT implemented — see Known limits.**

## Write actions & risks

None. Klaus is `capabilities.write: false`; no `writes.json` is shipped, matching legacy's `Write`
always returning `connectors.ErrUnsupportedOperation`.

## Known limits

- **Blocked: `reviews` stream (`ENGINE_GAP`).** Legacy's `harvestWindowed` (`klausapi.go:172-216`)
  reads Klaus's `reviews` endpoint (`/account/{account}/workspace/{workspace}/reviews`) by walking
  FORWARD through fixed-size (`P1W`, one calendar week) `fromDate`/`toDate` date windows from a lower
  bound (the incremental state cursor, else `start_date` config, else one window before now) to an
  upper bound (`end_date` config, else now), issuing one sequential HTTP request PER WINDOW (bounded
  defensively at 520 windows / ~10 years) — this is fundamentally a windowed-loop read pattern, not a
  page/cursor/offset/link-based one. None of the engine's 6 `PaginationSpec` types (`none`,
  `link_header`, `page_number`, `offset_limit`, `cursor`, `next_url` — `docs/migration/conventions.md`
  §3's pagination table) express a fixed-step sliding date-window request loop: every type either reads
  a single response's own continuation signal (a header, a body token, a last-record field) or counts
  pages/offsets, none re-issue sequential requests driven by a client-side date-window step
  independent of the previous response's content. The `incremental` block is similarly insufficient:
  `IncrementalSpec.RequestParam` sends exactly ONE lower-bound param per read (see klaviyo/leadfeeder's
  `docs.md` for the analogous "one request_param, not a paired window" gap), never a PAIRED
  `fromDate`+`toDate` window that advances and re-requests across a sync — there is no engine
  primitive for "issue N sequential requests, each with a computed window, until reaching an upper
  bound." This is squarely a `StreamHook`-shaped whole-stream read override (`docs/migration/
  conventions.md` §1's Tier-2 table: "whole-stream override... sub-resource fan-out" is the closest
  named legitimate trigger, and a windowed date-range fan-out is the same shape), which this wave's
  hard rules forbid writing (JSON + docs.md only; no Go/hooks packages). Confirmed by reading
  `internal/connectors/engine/{bundle,paginate,read}.go` directly (no grep hit for any date-window/
  time-step pagination concept) and by `klausapi_test.go`'s own
  `TestReadReviewsPaginatesByDateWindow`, which asserts exactly this multi-request date-window
  behavior against a live `httptest.Server`. This is an `ENGINE_GAP` blocker for a follow-up engine
  dialect increment (a 7th `PaginationSpec` type, e.g. `date_window` with `window_step`/`from_param`/
  `to_param`/`end_config_key` fields) or a Tier-2 `StreamHook` written in a follow-up wave once Go
  authoring is back in scope — not a per-connector patch attempted here. Once closed, `reviews` should
  follow legacy's exact shape: workspace-scoped path, `records.path: "conversations"`,
  `x-cursor-field: "lastUpdatedISO"`, and the same field-for-field mapping already declared for
  `users`/`categories` (schema can be authored now from `klausReviewRecord`/`klausReviewFields` even
  though the stream itself cannot be wired until the windowing primitive exists).
- Scorecards, root-cause management, and comment/dispute endpoints are out of scope for this wave; see
  `api_surface.json`'s `excluded` entries (`out_of_scope` for the blocked `reviews` endpoint and Pass B
  deferrals).

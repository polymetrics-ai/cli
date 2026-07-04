# Overview

Beamer is a feedback/changelog connector, expanded to full API-surface coverage (reads and
writes) in Pass B. It reads NPS survey responses, announcement posts, feature requests, comments,
post reactions, feature-request votes, and end users through the Beamer REST API
(`https://api.getbeamer.com/v0`), and writes posts, feature requests, their comments, reactions,
and votes. This bundle migrates `internal/connectors/beamer` (the hand-written legacy connector);
the legacy package stays registered and unchanged until wave6's registry flip.

## Auth setup

Provide a Beamer API key via the `api_key` secret; it is sent as a Bearer token
(`Authorization: Bearer <api_key>`) and is never logged, matching legacy's
`connsdk.Bearer(secret)` wiring exactly.

## Streams notes

All 4 streams (`nps`, `posts`, `feature_requests`, `comments`) share the same shape: `GET` against
a Beamer list endpoint that returns a bare JSON array at the response root (`records.path: "."`),
primary key `["id"]`. Pagination is genuinely 0-indexed (`pagination.type: page_number`,
`start_page: 0`, `page_param: page`, `size_param: maxResults`, `page_size: 100`) — the first
request sends `page=0`, matching legacy's `beamer_test.go:27-35` assertion and `harvest`'s
`for page := 0; ...` loop exactly; pagination stops on a short/empty page (fewer than `page_size`
records), identical to legacy's `len(records) < pageSize` check. Only `nps` is incremental: its
`date` field is filtered server-side via the `dateFrom` query param
(`incremental.request_param: dateFrom`, `param_format: rfc3339`, `start_config_key: start_date`),
matching legacy's `incrementalLowerBound` (persisted cursor, else `start_date` config) — Beamer's
API expects RFC3339 timestamps verbatim, so no reformatting is applied (legacy's own comment notes
the same: "no reformatting is required"). `posts`, `feature_requests`, and `comments` declare no
`incremental` block (full-refresh only) even though legacy exposes `date` as a `CursorFields`
candidate on every stream — only `nps` actually wires a request-time filter (`cursorParam`) in
legacy's per-stream routing table, so per §8 rule 2 the other three streams keep `x-cursor-field`
in their schemas only, with no `incremental` block.

`check` issues a single bounded `GET /nps?page=0&maxResults=1`, mirroring legacy's `Check`
implementation exactly (a 1-record probe of the `nps` endpoint confirms auth and connectivity
without mutating anything).

Pass B adds 3 new streams, all sharing the same base pagination/auth/records shape:
`post_reactions` (`fan_out` over `posts`: lists every post id via `GET /posts`, then paginates
`GET /posts/{post_id}/reactions` once per post, stamping `post_id` onto every emitted record),
`feature_request_votes` (identical `fan_out` shape over `feature_requests` —
`GET /feature-requests/{feature_request_id}/votes`, stamping `feature_request_id`), and `users`
(top-level `GET /users`, Beamer's own end-user identity store — Scale-plan-only per Beamer's docs,
but the read endpoint itself has no plan-gating expressed in the API contract, so it is included
like any other stream; a non-Scale account simply receives an auth/plan error from Beamer, handled
by the existing `error_map`).

**Path/auth discrepancy note**: Beamer's official docs page renders no static, fetchable endpoint
reference. This bundle's Pass B research cross-checked a third-party reverse-engineered API
reference (`getbeamer-api.pages.dev`) against this bundle's own tested legacy code
(`internal/connectors/beamer/*.go`), which is the authoritative parity source whenever the two
disagree — two disagreements were found: (1) the third-party reference names the feature-request
path `/requests`; legacy's tested, parity-verified path is `/feature-requests`, used consistently
here for `feature_request_votes` and every feature-request-scoped write action; (2) the third-party
reference documents a `Beamer-Api-Key` header; legacy's tested implementation (and its own passing
unit test, `beamer_test.go:55`) sends `Authorization: Bearer <api_key>` — this bundle's existing,
already-correct Bearer auth is unchanged (a parity-locked read path is not re-litigated by a
third-party doc during a surface-expansion pass).

## Write actions & risks

Beamer's legacy connector was read-only, but Pass B's full-surface research found Beamer's
create/update/delete mutation surface for posts, feature requests, and their comments/reactions/
votes is plain-JSON-bodied and fully dialect-expressible — `capabilities.write` now flips to `true`
and this bundle ships a full `writes.json`. Every action requires operator approval per its own
`risk` string:

- `create_post` / `update_post` / `delete_post` — announcement post CRUD. **A published post
  (`publish: true`) is immediately visible in Beamer's customer-facing widget** — the highest-risk
  action in this bundle.
- `create_post_comment` / `delete_post_comment` — comment CRUD on a post, on behalf of a named
  user (`userId`/`userEmail`/`userFirstname`/`userLastname`).
- `create_feature_request` / `update_feature_request` / `delete_feature_request` — feature request
  CRUD; a `visible: true` request is end-user-visible, and a `status` change is commonly
  user-facing (e.g. "planned"/"shipped" notifications).
- `create_feature_request_comment` / `delete_feature_request_comment` — comment CRUD on a feature
  request.
- `create_post_reaction` / `delete_post_reaction` — records/removes a reaction (`positive`/
  `neutral`/`negative`) on a post on behalf of a user.
- `create_feature_request_vote` / `delete_feature_request_vote` — records/removes a vote on a
  feature request on behalf of a user.

Every action's `title`/`content`/`language`/`linkUrl`/`linkText` fields are Beamer's own
translation arrays (parallel string arrays, one entry per language) — passed straight through as
JSON arrays via the default `body_type: json`; no special encoding is applied.

## Known limits

- **Every excluded endpoint is documented individually in `api_surface.json`.** Count-only
  aggregates and single-record detail lookups are `duplicate_of` their sibling list stream/action;
  `/unread`(`/count`) is `duplicate_of` `posts` (a stateful, GET-time-mutating filtered view a sync
  connector must never trigger implicitly); NPS survey dispatch, bulk end-user write/delete, team
  member management, and GDPR-style privacy erasure are `requires_elevated_scope`/
  `destructive_admin` (account-administration or irreversible-erasure actions, not routine
  reverse-ETL data mutations).
- `posts`/`feature_requests`/`comments`/`post_reactions`/`feature_request_votes` are full-refresh
  only (no server-side incremental filter wired in legacy for any of them), even though `date` is a
  schema-declared cursor candidate on most — matches legacy's own per-stream `cursorParam` routing
  table, where only `nps` sets a non-empty value.
- `users` has no `x-primary-key`-safe incremental cursor (`firstSeen`/`lastSeen` are informational
  fields, not declared as `x-cursor-field`, since Beamer's own docs give no incremental-filter query
  parameter for `/users`) — full-refresh only, matching every other stream in this bundle.

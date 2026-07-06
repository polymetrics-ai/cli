# SPEC — wave1-pilot

Phase of milestone connector-architecture-v2 (`.planning/ROADMAP.md`). Plan of record:
`~/.claude/plans/please-check-all-the-serialized-storm.md`; program PRD:
`docs/plans/universal-programming-loop-prd.md`; orchestration:
`docs/migration/orchestration-plan.md` (wave 1 row). Prerequisite phase wave0-engine-harness is
COMPLETE with reviewer verdict GO (`.planning/phases/wave0-engine-harness/REVIEW.md` re-review
section, HEAD 7fb4eb6).

## 1. Scope

Migrate **10 pilot connectors** to declarative defs bundles per
`docs/migration/conventions.md` (THE recipe — deviations are defects), one Sonnet
`gsd-loop-backend` agent each, at **capability PARITY with legacy** (every stream and write action
legacy implements today; full documented API surface is Pass B / wave5 — see DECISIONS.md #4
minimal-honest `api_surface.json` rule, unchanged for the pilot). Each connector gets an
engine-vs-legacy **parity test** following the wave0 golden pattern
(`internal/connectors/engine/parity_stripe_test.go`, `parity_searxng_test.go`): both connectors
driven live against the same `httptest` servers, RAW record equality, real wire-shape fixtures.

Also in scope, sequenced FIRST (§4): two carried engine follow-ups from wave0 (N1, N4). Also in
scope, sequenced LAST: Fable line-by-line review of all 10 diffs, conventions.md + executor-prompt
patch from pilot learnings, and `docs/migration/pilot-costs.json` (feeds the user's Pass B budget
decision).

Out of scope: registry flip / legacy deletion (wave6), Pass B surface expansion (wave5), any
non-pilot connector, `go.mod` changes (NEEDS_NEW_DEP is a human gate), history rewrite of the
11 MB blob (optional pre-push item, coordinator's call).

## 2. Pilot roster (inventory rows from `docs/migration/inventory.json`)

| # | name | loc | bucket | runtime_kind | documentation_url | tier expectation |
|---|---|---|---|---|---|---|
| 1 | xkcd | 186 | S | declarative_http_go | https://xkcd.com/json.html | 1 |
| 2 | vitally | 188 | S | declarative_http_go | https://docs.vitally.io/pushing-data-to-vitally/rest-api | 1 |
| 3 | bitly | 544 | M | declarative_http_go | https://dev.bitly.com/api-reference/ | 1 |
| 4 | calendly | 673 | M | declarative_http_go | https://developer.calendly.com/api-docs | 1 |
| 5 | sentry | 661 | M | declarative_http_go | https://docs.sentry.io/api/ | 1 (link_header risk, §5.3) |
| 6 | chargebee | 719 | L | declarative_http_go | https://apidocs.chargebee.com/docs/api/versioning | 1 (envelope unwrap, §5.4) |
| 7 | zendesk-support | 673 | M | declarative_http_go | https://developer.zendesk.com/api-reference/ticketing/introduction/ | 1 (dual auth) |
| 8 | monday | 758 | L | declarative_http_go | https://developer.monday.com/api-reference/docs | **2 (StreamHook — GraphQL, §5.5)** |
| 9 | github | 3664 | XL | native_go | https://docs.github.com/en/rest | **2 (AuthHook + WriteHook, §5.6)** |
| 10 | gmail | 718 | L | declarative_http_go | https://developers.google.com/gmail/api/reference/rest | **2 (AuthHook — OAuth refresh, §5.7)** |

All catalog slugs are `source-<name>`; all stream_count fields in inventory.json read 0 for the
declarative_http_go pilots (inventorygen counts a legacy manifest shape these packages don't use)
except github (19) — the parity source of truth is the legacy package's stream/write specs, not
the inventory count.

## 3. Acceptance criteria (ROADMAP wave1-pilot)

1. All 10 connectors reach status `migrated` (or `partial`/`blocked` with a **typed blocker** per
   `docs/migration/result.schema.json` — never a silent approximation) with agent self-check green
   (conventions.md §7 command block).
2. Per-connector parity test green: `go test ./internal/connectors/paritytest/...` (§6 location
   decision) plus wave0's existing `go test ./internal/connectors/engine -run TestParity`.
3. Wave gate green: path guard, `connectorgen gen` (hookset regen), `go build ./...`,
   `go test ./internal/connectors/... ./cmd/...`, `TestConformance/<name>` for all 10,
   `golangci-lint run` / `make verify`.
4. Fable line-by-line review of **100% of the 10 diffs** (orchestration-plan layer 3, pilot
   override of the 20% sample) with verdicts per `docs/migration/review.schema.json`; zero
   unresolved blocker findings.
5. `docs/migration/conventions.md` and the executor template in
   `docs/prompts/universal-programming-loop-prompts.md` patched with pilot learnings.
6. `docs/migration/pilot-costs.json` written (per-connector tokens/duration/status/deviations).
7. Pass B budget decision made **with the user** (human gate — coordinator presents the cost
   report; the decision itself is not automatable).

## 4. Pre-pilot engine follow-ups (MUST land before any pilot dispatch)

Carried from wave0 (`SUMMARY.md` "Carried into wave1-pilot"; REVIEW.md re-review "New findings"):

- **N1 — `formatCursorForAssertion` `github_date_range` alignment**
  (`internal/connectors/conformance/dynamic.go`): the `github_date_range` branch returns
  `">=" + value` VERBATIM while the engine's `formatParam`
  (`internal/connectors/engine/read.go`) normalizes digits → RFC3339 and offsets/fractional
  seconds → UTC second precision. A bundle combining `github_date_range` with a numeric or
  non-UTC cursor would falsely FAIL `cursor_advances`. The github bundle (P-9) is exactly the
  shape that could trip this. Fix = mirror the engine's normalization on the assertion side;
  red-first via a new conformance self-test bundle
  (`internal/connectors/conformance/testdata/good/` — follow the existing
  `acme-numeric-cursor` pattern from the B2 fix).
- **N4 — stale doc comment** on `incrementalLowerBoundValue`
  (`internal/connectors/engine/read.go`): still says the lower bound is "always RFC3339 when
  present" — false since the B1 digits-passthrough fix. Docs-only; batch into the same task.

The other carried flags are **noted, not blocking**: N2 (digit-shaped non-unix start values —
watch during pilot, promote to a validate-time guard only if a pilot actually hits it), N3
(relative next-page URLs fail closed — bitly/calendly return ABSOLUTE next URLs per their legacy
comments, so no engine work expected; any pilot that hits a relative next URL files ENGINE_GAP),
N5 (`..%5C` residuals — no pilot interpolates untrusted values into paths beyond config/record
ids; stays a wave-level hardening note).

## 5. Per-connector specification

Common to all 10: bundle layout per conventions.md §1 Tier 1; naming §2; fixtures §4
(recorded-real-shape, sanitized, 2-page fixture REQUIRED wherever pagination is declared, real
wire types — numeric cursors stay JSON numbers); parity-deviation ledger discipline §5; escape
hatches §6; self-verify §7. Legacy packages are read-only reference (FORBIDDEN to edit). Every
agent inlines its inventory row and reads its legacy package in full before writing.

### 5.1 xkcd (P-1, Tier 1)

Legacy `internal/connectors/xkcd/xkcd.go` (186 loc): no auth; streams `latest`
(`info.0.json`) and `comic` (`<comic_number>/info.0.json` — a **templated path**,
`{{ config.comic_number }}` — exercises wave0's F1 stream-path-interpolation fix); each returns a
**single JSON object**, so `records.single_object: true`; no pagination; no incremental; legacy
stamps a `stream` marker field (static-literal `computed_fields`, searxng pattern) and validates
`comic_number` as a path segment (engine's `InterpolatePath` urlencode + `..` guard covers this —
assert in parity that a hostile `comic_number` fails closed on both sides). `Write` returns
`ErrUnsupportedOperation` → `capabilities.write: false`, no `writes.json`. Note: legacy has a
fixture mode emitting a `fixture: true` field — fixture mode is a legacy affordance, NOT part of
the bundle; parity is asserted against legacy's LIVE read path via httptest.

### 5.2 vitally (P-2, Tier 1) · bitly (P-3, Tier 1) · calendly (P-4, Tier 1)

- **vitally** (`internal/connectors/vitally/vitally.go`, 188 loc): auth via
  `connsdk.APIKeyHeader("Authorization", auth, "")` (vitally.go:104) — read the legacy value
  construction and express as engine `basic` mode or `api_key_header` + `base64` filter,
  whichever reproduces the exact header byte-for-byte (assert in parity). `accounts`-family
  streams; read legacy for pagination/cursor specifics.
- **bitly** (`internal/connectors/bitly/bitly.go` + `streams.go`, 544 loc): Bearer auth
  (bitly.go:239); paginated endpoints (bitlinks) follow a `pagination.next` **absolute** URL →
  engine `next_url` paginator (N3 relative-URL fail-closed does not bite; verify the fixture uses
  absolute URLs like legacy's own tests do); core list endpoints are full-refresh (no cursor
  field — bitly.go:36 comment). Legacy fixture-mode-only fields (`connector`, `fixture`,
  `previous_cursor` — bitly.go:213-217) are NOT part of the live record shape; parity targets the
  live path.
- **calendly** (`internal/connectors/calendly/calendly.go` + `streams.go`, 673 loc): Bearer auth
  (calendly.go:317); records under `collection`; pagination via `pagination.next_page` absolute
  URL → `next_url` paginator; incremental raised by `start_date` config (calendly.go:97 comment) —
  map to `incremental` + `start_config_key` per engine dialect; scoping config
  (organization/user URIs) per legacy.

### 5.3 sentry (P-5, Tier 1 with a pre-identified risk)

Legacy `internal/connectors/sentry/sentry.go` (661 loc): Bearer auth (sentry.go:243). Pagination
is RFC 5988 Link header **with Sentry's twist**: a `rel="next"` link is ALWAYS present and the
`results="true|false"` attribute is the real more-pages signal (sentry.go:7-9, 144-152) — legacy
hand-rolls this precisely because `connsdk.LinkHeaderPaginator` follows `rel="next"`
unconditionally, and the engine's `link_header` type wraps connsdk semantics
(`internal/connectors/engine/paginate.go:50-51`). Resolution ladder (in order, stop at first that
holds — decided by evidence in the parity/fixture run, never assumed):
1. If the engine's empty-page/loop-guard behavior yields **identical emitted records** with at most
   one extra trailing request, ship Tier 1 `link_header` + document the extra-request deviation
   (conventions §5 meta-rule: request-count delta with identical record DATA is ACCEPTABLE,
   documented).
2. Else Tier-2 `StreamHook` porting legacy's Link/`results` handling.
3. Else typed `ENGINE_GAP` (candidate future engine feature: `link_header` stop-attribute).

### 5.4 chargebee (P-6, Tier 1, envelope unwrap)

Legacy `internal/connectors/chargebee/chargebee.go` + `streams.go` (719 loc): HTTP Basic with the
site API key as username, empty password (chargebee.go:262-264) → engine `basic` mode. Pagination:
`offset` query param carrying the `next_offset` token from the body → engine `cursor` paginator
with `cursor_param: offset`, `token_path: next_offset`. Records under top-level `list`, each item
wrapped in a resource envelope (`{"customer": {...}}`) → schema projection alone would drop
everything; use per-field `computed_fields` (`"id": "{{ record.customer.id }}"`, …) to flatten,
matching legacy's mapRecord field-for-field (conventions §2 schema-as-projection). Incremental
cursor `updated_at` (streams.go:40,47) with Chargebee's Unix-seconds wire values — `param_format:
unix_seconds` and **numeric fixtures** (the exact B1/B2 shape wave0 fixed; do not regress it).
If per-field computed_fields cannot reproduce legacy's record shape exactly (e.g. absent-field
semantics), fall back to Tier-2 `RecordHook` with justification — never a silently different
shape.

### 5.5 monday (P-8, Tier 2 — StreamHook)

Legacy `internal/connectors/monday/monday.go` + `streams.go` (758 loc): raw-token `Authorization`
header + `API-Version` header (monday.go:396); ALL reads are **GraphQL POSTs** whose pagination
lives INSIDE the query text — page-number `(limit: N, page: M)` for boards/users/teams/tags
(monday.go:151-160) and a `next_items_page` cursor for `items` (monday.go:131-146). The engine
cannot express this: `StreamSpec.Body` exists (`engine/bundle.go:160`) but the read path never
sends it (`engine/read.go:142` passes `nil` body — **pre-identified latent gap**; record as an
authoring warning in conventions during P-12, candidate engine feature if wave2+ needs
POST-body reads ≥3 times), and no paginator mutates a request body. Specification: Tier-2
`hooks/monday/hooks.go` implementing `StreamHook` (all streams handled; declarative fallback never
taken) + `CheckHook` if the legacy check is also a GraphQL POST — that is 2 hook interfaces, at
the `connectorgen validate` cap. Hard cap ~300 lines (conventions §1 Tier 2); if the port cannot
fit honestly, escalate Tier 3 (native component split) rather than compressing — flag to
coordinator first (scope change). Bundle still declares every stream/schema/fixture; `streams.json`
stream entries carry the declarative metadata (names, schemas, incremental cursor `updated_at`)
even though reads dispatch through the hook.

### 5.6 github (P-9, Tier 2 — AuthHook + WriteHook; XL, highest risk)

Legacy `internal/connectors/github/` (3664 loc, 19 streams — streams.go:12-30; **the only pilot
with real writes**: 16 actions, github.go:1759+ `githubWriteActionSpecs`, executed by
`github.go:236 Write`). Specification:
- **Streams (Tier-1 JSON)**: page_number pagination (`per_page`/`page`, short-page stop —
  github.go:254-289), templated paths `/repos/{{ config.owner }}/{{ config.repo }}/...` (F1 fix
  landed in wave0 — first real consumer), `since` config forwarded on the streams legacy forwards
  it for (github.go:91-92,152-162), heavy `computed_fields` flattening (legacy emits flattened
  fields like `author_login`, `commit_committer_date`, and stamps `repository` on every record —
  streams.go fields). `repository` stream is `single_object`. The issues stream filters out pull
  requests (streams.go:13) → `records.filter` (`FilterSpec`, engine/bundle.go:176-180) if
  expressible (`field_absent: pull_request`), else RecordHook — decide from legacy's exact filter.
- **Auth (Tier-2 AuthHook)**: declarative candidates for token-family modes (bearer via
  `{{ secrets.token }}` with `when` gating) + `mode: custom, hook: github` for
  `auth_type=github_app` — the JWT→installation-token exchange (github/auth.go:117-155
  `githubAppInstallationToken`, `githubAppJWT`) ports into `hooks/github/hooks.go`
  `Authenticator(ctx, ...)` (ctx is honored — wave0 F8 fix). Private key arrives via secrets
  (`githubAppPrivateKey` / base64 variant — auth.go:211-218); never logged, never in fixtures
  (THREAT-MODEL §2).
- **Writes**: simple actions (create_issue, update_issue, comment_issue, create_label, ...) as
  declarative `writes.json` actions with `path_fields`; **compound** actions
  (create_pull_request's labels/assignees/milestone/reviewers follow-up requests, close_issue's
  optional comment, close_pull_request's optional comment) via `WriteHook`
  (`hooks/github/hooks.go` `ExecuteWrite`, falling back `handled=false` for the simple actions).
  AuthHook + WriteHook = 2 interfaces (at cap); the ~300-line budget is tight — the agent
  reports `partial` with typed blockers for any write actions that don't fit honestly rather than
  exceeding the cap or faking parity. Parity for `partial` is acceptable per §3.1 as long as
  every ported piece passes and every unported piece is a typed, ledgered blocker.
- **Parity floor (must-have even if partial)**: all 19 streams; write actions at minimum
  create_issue, update_issue, comment_issue, create_pull_request (compound), merge_pull_request,
  delete_label (delete semantics / missing_ok).

### 5.7 gmail (P-10, Tier 2 — AuthHook) and the gmail decision

**Decision: KEEP gmail in the roster; an OAuth2 refresh-token AuthHook suffices.** Evidence read
from legacy (`internal/connectors/gmail/auth.go`, 127 loc — the whole file): gmail is neither a
stub nor api-key-based. It implements the **OAuth 2.0 refresh-token grant** only: `oauthRefreshAuth`
POSTs `grant_type=refresh_token` + `refresh_token` + `client_id` [+ `client_secret`, `scope`] to
the Google token endpoint (config-overridable `token_url`, gmail.go:339), caches the access token
until 60s before expiry (auth.go:64-127), and sets `Authorization: Bearer` per request. The
3-legged **consent/acquisition** dance is NOT in legacy — the refresh token arrives as a secret
(`client_refresh_token`, generic `refresh_token` alias — gmail.go:321-327), i.e. the credentials
layer already owns acquisition/storage, which was the open question in REVIEW.md carried item 7.
Therefore:
- `hooks/gmail/hooks.go` implements `AuthHook` with a refresh-token authenticator mirroring
  `oauthRefreshAuth` (~130 loc ports comfortably under the 300-line cap; injectable `now` for
  tests carries over). Bundle auth: `[{"mode":"custom","hook":"gmail", ...}]`; the existing
  `AuthSpec` fields `token_url`/`client_id`/`client_secret`/`scopes`
  (engine/bundle.go:103-106) carry the templated config into the hook.
- Engine has `oauth2_client_credentials` but NOT a refresh-token grant (engine/bundle.go:92 mode
  list) — do NOT add an engine mode in this phase; if wave2+ hits refresh-token grants ≥3 times,
  that becomes a mini wave-0 engine increment per the ENGINE_GAP recurrence rule (record this in
  P-12 conventions notes).
- Streams (Tier-1 JSON): `messages`, `threads`, `drafts`, `labels` (streams.go:35-62); templated
  path with the `%s` userId slot → `{{ config.user_id }}` (default `me` per legacy); cursor
  pagination `pageToken` param / `nextPageToken` token_path; **no incremental** (streams.go:31-34:
  list endpoints are newest-first, no cursor field published) — full-refresh only, matching
  legacy. Read-only (`Write` returns `ErrUnsupportedOperation`, gmail.go:191-192).
- No roster swap needed; the fallback (swap for an M-bucket connector from inventory.json) is
  moot because legacy gmail is a self-contained, portable refresh-grant connector.

## 6. Parity-test location decision

**Decision: new per-connector packages `internal/connectors/paritytest/<name>/`** (e.g.
`paritytest/xkcd/parity_test.go`, package `xkcdparity_test`-style external test package + a
one-line `doc.go` so `go build ./...` sees a real package; `zendesk-support` →
dir `paritytest/zendesk-support/`, package `zendesksupportparity`). Rationale:
- Wave0's parity tests live in `internal/connectors/engine/` as `package engine_test`
  (parity_stripe_test.go:1). Ten agents adding files there would (a) violate the
  orchestration-plan path-guard rule that assigned dirs are DISJOINT and (b) share one Go test
  package namespace — two agents independently defining the same helper (`mustJSON`,
  `newServer`, ...) is a compile-time collision that serializes the whole wave.
- Per-connector directories give clean 10-way parallelism, a per-agent path guard
  (`defs/<name>/**`, `paritytest/<name>/**`, `hooks/<name>/**`), and per-connector
  `go test ./internal/connectors/paritytest/<name>` self-verify.
- Bundles load exactly as wave0 does: `engine.LoadAll(defs.FS)` / `engine.Load(defs.FS, name)`
  (parity_stripe_test.go:31-50; `defs.go`'s `//go:embed all:*` means NO shared-file edit when
  adding `defs/<name>/` — verified).
- **stripe/searxng parity tests stay in `internal/connectors/engine/`** (wave0 artifacts; moving
  them buys nothing and churns reviewed code). postgres parity stays in
  `internal/connectors/native/postgres/`.
- Tier-2 pilots blank-import their own `hooks/<name>` package from the parity test to trigger
  `engine.RegisterHooks` init — no dependency on the shared generated
  `hooks/hookset/hookset_gen.go`, which is regenerated ONLY by the orchestrator at wave close via
  `go run ./cmd/connectorgen gen` (cmd/connectorgen/gen.go:14-24).
- conventions.md §7's parity command becomes
  `go test ./internal/connectors/paritytest/<name> -v` for wave1+ connectors (patched in P-12).

## 7. Interfaces & dependencies

- Engine surface used: `engine.Load`/`LoadAll`, `engine.NewConnector`(assembly),
  `engine.RegisterHooks` + the 5 hook interfaces (`engine/hooks.go`), `AuthSpec` mode `custom`.
  No engine behavior changes in this phase except P-0 (conformance assertion mirror + doc
  comment).
- No new Go module dependencies (github JWT signing uses stdlib `crypto/rsa` exactly as legacy
  auth.go does; gmail hook uses stdlib HTTP). Any agent that believes it needs one files
  NEEDS_NEW_DEP and stops — human gate.
- Shared/generated files remain orchestrator-only: `hooks/hookset/hookset_gen.go`,
  `registryset/registry_gen.go`, `catalog_data.json`, `go.mod` (conventions §7 FORBIDDEN list).
  No pilot registers a connector factory — registration flip is wave6 (postgres grep-guard
  pattern applies to hooks packages too: hooks register HOOKS, never connector factories).

## 8. Human gates in this phase

- **Pass B budget decision** (acceptance #7): coordinator presents `pilot-costs.json` to the user;
  the phase does not close the decision autonomously.
- **NEEDS_NEW_DEP** if any agent files it (none expected).
- **Quality-gate reductions**: none planned; any parity test weakening found in review is a
  blocker, not a negotiation.
- github/gmail hooks touch auth/secret handling — not a stop-gate (no production auth
  infrastructure changes; hooks are unregistered until wave6), but P-11's review covers them
  line-by-line at 100% and THREAT-MODEL.md §2 governs.

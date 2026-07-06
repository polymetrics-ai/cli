# PLAN — wave1-pilot

Task list with dispatch waves. Tags: `[behavior]` (has a paired red-first test task inline) ·
`[docs-only]` · `[process]` (orchestrator/coordinator bookkeeping, no code). Models per
`docs/prompts/universal-programming-loop-prompts.md`: executors = sonnet (gsd-loop-backend),
review = fable (gsd-loop-reviewer). Every executor prompt is rendered from the migration-executor
template (prompts doc §"Template: migration executor") with the per-task block below inlined.

## Dispatch waves

| Wave | Tasks | Parallelism | Gate to advance |
|---|---|---|---|
| DW-0 | P-0 | serial (touches shared engine/conformance packages) | P-0 self-verify green + red-first evidence in TDD-LEDGER.md |
| DW-1 | P-1 … P-10 | **fully parallel (10 agents)** — disjoint dirs per SPEC §6 | per-agent self-verify + structured result JSON |
| DW-2 | P-11 (+ repair dispatches it spawns) | review parallel per connector; repairs serialized per connector | zero unresolved blocker findings |
| DW-3 | P-12, P-13 | parallel (different files) | artifacts exist + accurate |
| DW-4 | P-14 | serial (single-writer wave close) | full gate green; phase artifacts refreshed |

DW-1 parallelism is legal because each agent's writable set is exactly
`internal/connectors/defs/<name>/**` + `internal/connectors/paritytest/<name>/**`
(+ `internal/connectors/hooks/<name>/**` for monday/github/gmail) + nothing else; `defs.go`'s
`//go:embed all:*` and the orchestrator-only `hooks/hookset/hookset_gen.go` mean no shared file is
ever touched (SPEC §6, §7). Path guard enforces this at P-14 (and spot-checks after DW-1).

## Tasks

### P-0 [behavior] Pre-pilot engine follow-ups N1 + N4 (DW-0, serial, TDD)

Files: `internal/connectors/conformance/dynamic.go`,
`internal/connectors/conformance/testdata/good/<new self-test bundle>/**`,
`internal/connectors/engine/read.go` (comment only).
- **Test first (RED)**: add a conformance self-test bundle
  `testdata/good/acme-github-range-cursor` — `github_date_range` `param_format` with (a) a numeric
  Unix-seconds cursor fixture and (b) a non-UTC-offset RFC3339 cursor fixture — modeled on the
  existing `acme-numeric-cursor` bundle (added for wave0's B2). Assert `cursor_advances` PASSES;
  it currently FAILS because `formatCursorForAssertion`'s `github_date_range` branch returns
  `">=" + value` verbatim while the engine normalizes (REVIEW.md N1). Record RED output in
  TDD-LEDGER.md.
- **Fix (GREEN)**: route the `github_date_range` branch through the same
  digits→RFC3339/UTC-second-precision normalization the engine's `formatParam` uses
  (mirror on the independent assertion side, matching how B2 added
  `parseLowerBoundTimeForAssertion`).
- **N4 [docs-only, batched]**: correct `incrementalLowerBoundValue`'s doc comment in
  `engine/read.go` ("always RFC3339 when present" → digits-or-RFC3339 per B1). No test pair
  (comment-only).
- Verify: `go test ./internal/connectors/conformance ./internal/connectors/engine` +
  `make verify`. Reuse: `parseLowerBoundTime` (engine/read.go), `cursorValueString` /
  `parseLowerBoundTimeForAssertion` (conformance/dynamic.go) — extend, don't duplicate.

### P-1 … P-10 [behavior] one pilot connector each (DW-1, parallel)

Shared task shape (differences per connector below; full behavioral spec in SPEC §5):

- **Writable dirs (exclusive)**: `internal/connectors/defs/<name>/**`,
  `internal/connectors/paritytest/<name>/**`, and for monday/github/gmail
  `internal/connectors/hooks/<name>/**`. FORBIDDEN: everything in conventions.md §7's list, all
  other dirs, `git commit`.
- **Red-first protocol (the paired test IS the parity suite)**: (1) write
  `paritytest/<name>/parity_test.go` FIRST — it loads the bundle via
  `engine.Load(defs.FS, "<name>")` and fails RED because the bundle doesn't exist; capture the
  failure line in the agent trace; (2) author the bundle (+hooks) until parity is GREEN; (3) never
  weaken an assertion to get green — a shape the engine can't produce is a typed blocker or
  ledgered deviation (conventions §5 meta-rule).
- **Parity suite minimum** (pattern: `internal/connectors/engine/parity_stripe_test.go` /
  `parity_searxng_test.go` — drive BOTH connectors live against the same httptest server, RAW
  `reflect.DeepEqual` record equality, legacy-side sanity assertions): per-stream record parity;
  pagination parity (2 pages, request-shape assertions); incremental parity incl. the
  app-persisted cursor round-trip shape where the connector is incremental (feed both sides the
  digit-string cursor `internal/app/sync_modes.go` actually persists — the B1 lesson); auth
  header parity; error-path parity (non-2xx mapping); write parity for github (dry-run request
  shape + live httptest execution per action, incl. delete missing_ok and fail-fast accounting).
- **Self-verify** (conventions §7): `go run ./cmd/connectorgen validate
  internal/connectors/defs/<name>` · `go build ./internal/connectors/... && go vet
  ./internal/connectors/...` · `go test ./internal/connectors/conformance -run
  'TestConformance/<name>'` · `go test ./internal/connectors/paritytest/<name> -v`.
- **Report**: JSON per `docs/migration/result.schema.json`, honest status
  (migrated|partial|blocked), deviations → conventions §5 ledger candidates, blockers typed.

| Task | Connector | Inventory row (inline in prompt) | Extra dirs | Task-specific requirements (SPEC ref) |
|---|---|---|---|---|
| P-1 | xkcd | `{"name":"xkcd","path":"internal/connectors/xkcd","loc":186,"bucket":"S","runtime_kind":"declarative_http_go","catalog_slugs":["source-xkcd"],"documentation_url":"https://xkcd.com/json.html","stream_count":0}` | — | single_object streams; templated comic path; static-literal `stream` marker; hostile-path fail-closed parity (§5.1) |
| P-2 | vitally | `{"name":"vitally","path":"internal/connectors/vitally","loc":188,"bucket":"S","runtime_kind":"declarative_http_go","catalog_slugs":["source-vitally"],"documentation_url":"https://docs.vitally.io/pushing-data-to-vitally/rest-api","stream_count":0}` | — | byte-exact Authorization header parity (§5.2) |
| P-3 | bitly | `{"name":"bitly","path":"internal/connectors/bitly","loc":544,"bucket":"M","runtime_kind":"declarative_http_go","catalog_slugs":["source-bitly"],"documentation_url":"https://dev.bitly.com/api-reference/","stream_count":0}` | — | `next_url` paginator, absolute URLs; ignore legacy fixture-mode fields (§5.2) |
| P-4 | calendly | `{"name":"calendly","path":"internal/connectors/calendly","loc":673,"bucket":"M","runtime_kind":"declarative_http_go","catalog_slugs":["source-calendly"],"documentation_url":"https://developer.calendly.com/api-docs","stream_count":0}` | — | `next_url` via pagination.next_page; records path `collection`; start_date incremental (§5.2) |
| P-5 | sentry | `{"name":"sentry","path":"internal/connectors/sentry","loc":661,"bucket":"M","runtime_kind":"declarative_http_go","catalog_slugs":["source-sentry"],"documentation_url":"https://docs.sentry.io/api/","stream_count":0}` | — | link_header `results=` twist — resolution ladder §5.3; termination MUST be proven by 2-page fixture parity |
| P-6 | chargebee | `{"name":"chargebee","path":"internal/connectors/chargebee","loc":719,"bucket":"L","runtime_kind":"declarative_http_go","catalog_slugs":["source-chargebee"],"documentation_url":"https://apidocs.chargebee.com/docs/api/versioning","stream_count":0}` | — | cursor token_path=next_offset; per-field computed_fields envelope unwrap; unix_seconds numeric fixtures (§5.4) |
| P-7 | zendesk-support | `{"name":"zendesk-support","path":"internal/connectors/zendesk-support","loc":673,"bucket":"M","runtime_kind":"declarative_http_go","catalog_slugs":["source-zendesk-support"],"documentation_url":"https://developer.zendesk.com/api-reference/ticketing/introduction/","stream_count":0}` | — | dual auth candidates (`basic` `{{ config.email }}/token` + bearer, `when`-gated); cursor page[after]/meta.after_cursor (§5.2 table) |
| P-8 | monday | `{"name":"monday","path":"internal/connectors/monday","loc":758,"bucket":"L","runtime_kind":"declarative_http_go","catalog_slugs":["source-monday"],"documentation_url":"https://developer.monday.com/api-reference/docs","stream_count":0}` | `hooks/monday/` | Tier-2 StreamHook (GraphQL POST reads, in-body pagination); ≤300 loc or escalate; hook test file required (§5.5) |
| P-9 | github | `{"name":"github","path":"internal/connectors/github","loc":3664,"bucket":"XL","runtime_kind":"native_go","catalog_slugs":["source-github"],"documentation_url":"https://docs.github.com/en/rest","stream_count":19}` | `hooks/github/` | Tier-2 AuthHook (github_app JWT exchange) + WriteHook (compound writes); 19 streams; 16 write actions, parity floor per §5.6; templated repo paths |
| P-10 | gmail | `{"name":"gmail","path":"internal/connectors/gmail","loc":718,"bucket":"L","runtime_kind":"declarative_http_go","catalog_slugs":["source-gmail"],"documentation_url":"https://developers.google.com/gmail/api/reference/rest","stream_count":0}` | `hooks/gmail/` | Tier-2 AuthHook (OAuth2 refresh-token grant, ports `gmail/auth.go` oauthRefreshAuth incl. expiry caching + injectable clock); pageToken/nextPageToken cursor; no incremental (§5.7) |

Sequencing note inside DW-1: all 10 dispatch simultaneously; P-9 (github) is the long pole —
if staggering is needed for operator attention, start P-9 and P-8 first.

### P-11 [process + behavior-adjacent] Fable line-by-line review, 100% coverage (DW-2)

One review per connector (10 verdicts), model=fable, READ-ONLY, prompt = adversarial-reviewer
template + pilot override "100% of diffs, line-by-line" (orchestration-plan layer 3; ROADMAP
pilot acceptance). Checklist: schema fidelity vs fetched documentation_url (3 streams
spot-checked + EVERY github write action's method/path/required fields), fixture realism (real
wire shapes — numeric cursors as numbers), escape-hatch justification (monday/github/gmail hooks
vs conventions §6 triggers), secret redaction, parity-test integrity (no weakened assertions, RAW
equality, legacy driven live), conventions adherence. Output: verdict JSON per
`docs/migration/review.schema.json` per connector, saved to
`.planning/phases/wave1-pilot/traces/review-<name>.json`. Blocker findings → repair dispatch
(repair-agent template; 1 retry, then quarantine per orchestration-plan). Review notes feed P-12.

### P-12 [docs-only] Patch conventions.md + executor prompt template (DW-3)

Files: `docs/migration/conventions.md`, `docs/prompts/universal-programming-loop-prompts.md`.
From P-1..P-11 evidence: new worked examples/rules (expected candidates, confirm from actual
results — parity-test location §6 rule + `paritytest/<name>` self-verify command; GraphQL/POST-body
reads are hook territory (`StreamSpec.Body` unwired — engine/read.go:142); OAuth refresh-token
grant = AuthHook pattern w/ gmail as worked example; envelope-unwrap via computed_fields w/
chargebee example; link_header stop-attribute note from sentry outcome; deviation-ledger
additions). Also record prompt-eval notes: which executor-prompt ambiguities caused deviations or
retries, and the exact template wording changes made (input to EVAL-PLAN prompt metrics). No code.

### P-13 [process] Pilot cost report (DW-3)

File: `docs/migration/pilot-costs.json`. Written by the coordinator from data IT records at each
dispatch (agents cannot see their own usage): per connector `{name, model, input_tokens,
output_tokens, total_tokens, wall_clock_minutes, dispatches (initial+repairs), status,
deviations_count, blockers[], reviewer_verdict}` + wave totals + projection columns for the Pass B
decision (per-bucket average × remaining inventory histogram S137/M388/L31/XL1 from wave0
SUMMARY.md). Schema documented inline in the file header comment field. See OBSERVABILITY.md for
the capture method.

### P-14 [process] Wave gate + phase close (DW-4, serial, single-writer)

1. Path guard: `git status --porcelain` — every changed path under
   `internal/connectors/{defs,paritytest,hooks}/<pilot-names>/**`,
   `internal/connectors/conformance/**` (P-0), `internal/connectors/engine/read.go` (P-0 comment),
   `docs/migration/**`, `docs/prompts/**`, `.planning/phases/wave1-pilot/**`. Violations →
   revert + repair dispatch (this is the gate wave0's B3 proved must actually RUN — record the
   command output in VERIFICATION.md).
2. `go run ./cmd/connectorgen gen` (hookset + registryset regen; assert registryset byte-identical
   — no registration in this phase) then full: `go build ./... && go test ./... && make lint`
   (or `make verify`), `go test ./internal/connectors/conformance -run TestConformance` (all),
   `go test ./internal/connectors/paritytest/...`.
3. Refresh phase bookkeeping at close: SUMMARY.md, VERIFICATION.md (with recorded HEAD),
   TDD-GATE.json (task rows for P-0 and P-1..P-10 red-first evidence), RUN-STATE.json,
   TDD-LEDGER.md → traces. Commit once (orchestrator; wave-close commit per orchestration-plan).
4. Present pilot-costs.json to the user for the Pass B decision (HUMAN GATE — phase acceptance
   #7; do not decide autonomously).

## Test-pairing summary (TDD gate input)

| Behavior task | Paired red-first test |
|---|---|
| P-0 fix | new conformance self-test bundle failing `cursor_advances` before the fix |
| P-1..P-10 bundle (+hooks) | `paritytest/<name>/parity_test.go` written first, RED on missing bundle; hooks additionally get `hooks/<name>/hooks_test.go` (unit: dispatch, token caching/expiry for gmail, JWT exchange request shape for github, GraphQL pagination loops for monday) |
| P-11/P-12/P-13/P-14 | no production behavior — process/docs (exempt, recorded as such in TDD-GATE.json) |

## Stop conditions (all tasks)

Same failure twice with no new evidence → stop, report. ENGINE_GAP that blocks a connector →
typed blocker + quarantine path, never invented Go in declarative fields. Any need to touch a
FORBIDDEN file → stop. NEEDS_NEW_DEP → stop (human gate). Reviewer finds a weakened gate → phase
halt + coordinator escalation.

# S2 zendesk-support repair — TDD ledger

Agent: gsd-loop-backend (Sonnet), gap-loop cycle-1 Step 2 (GAP-LOOP-PLAN.md),
repairing REVIEW-B.md's zendesk-support FAIL verdict (1 blocker, 2 majors, 2
minors). Files: `internal/connectors/defs/zendesk-support/**`,
`internal/connectors/paritytest/zendesk-support/**`,
`.planning/phases/wave1-pilot/TEST-PLAN.md` (row amendment per fix 1).

## Baseline (before this cycle)

`go test ./internal/connectors/paritytest/zendesk-support/... -v` — all 8
top-level tests PASS on the pre-repair bundle (green harness masking the
finding-1/finding-2 divergences per REVIEW-B.md's "green harness is necessary
but NOT sufficient" framing).

## Red-first evidence (this cycle)

Edited `parity_test.go` FIRST, before touching any `defs/zendesk-support/`
source file, adding/inverting exactly the two tests the review's fixes
require, then ran the suite to capture RED:

```
go test ./internal/connectors/paritytest/zendesk-support/... -run TestParityZendesk -v

=== RUN   TestParityZendesk_HasMoreFalseWithNonNullCursorStopsPagination
    parity_test.go:331: Read(tickets): zendesk-support stream=tickets: cursor(token_path): loop detected — token "STALE_CURSOR_AFTER_LAST_PAGE" requested twice
--- FAIL: TestParityZendesk_HasMoreFalseWithNonNullCursorStopsPagination (0.00s)

=== RUN   TestParityZendesk_StartDateConfigNeverSendsServerFilter
    parity_test.go:422: updated_at[gte] = "2026-01-01T00:00:00Z", want empty (legacy sends no server-side incremental filter; "updated_at[gte]" is not real Zendesk API surface)
--- FAIL: TestParityZendesk_StartDateConfigNeverSendsServerFilter (0.00s)
FAIL
```

Both RED for the expected reasons:
- `HasMoreFalseWithNonNullCursorStopsPagination` (new test, REVIEW-B finding
  2): the pre-fix `streams.json` pagination block had no `stop_path`, so the
  token-path cursor paginator kept following a non-null
  `meta.after_cursor` even though `meta.has_more:false` — on THIS server
  (single page, unchanging cursor) that manifests as the engine's own
  loop-guard tripping (`tokenPathCursor` request the same token twice),
  concretely proving the pre-fix bundle would issue at least one extra
  request past the real last page against any live account that populates a
  stale cursor on its final page (exactly REVIEW-B's "worst case: a
  pagination loop with no guard" scenario — the loop-guard converts an
  infinite loop into a bounded failure, but a failure is still wrong).
- `StartDateConfigNeverSendsServerFilter` (inversion of
  `StartDateConfigRaisesLowerBound`, REVIEW-B finding 1 / blocker): the
  pre-fix bundle declared `incremental.request_param: "updated_at[gte]"` on
  every stream, so a `start_date` config value WAS forwarded as a real query
  parameter — the inverted assertion (expect NO such param) fails against
  that behavior, proving the invented incremental filter was live.

## Fixes applied (green transitions)

1. **Blocker — invented `updated_at[gte]` incremental (REVIEW-B finding 1).**
   Removed the entire `incremental` block from all 5 streams in
   `streams.json` (`tickets`, `users`, `organizations`, `groups`,
   `satisfaction_ratings`) — legacy's `harvest()` (zendesk_support.go:152-195)
   sends no server-side incremental filter of any kind; `start_date` is
   doc-comment-only in legacy, never wired to a query key. Each stream's
   schema (`schemas/*.json`) already declares `x-cursor-field: updated_at`
   from the original authoring pass and needed NO change — this is exactly
   the calendly-3-streams pattern (event_types/organization_memberships/
   groups/users there: schema-only `x-cursor-field`, no `incremental` block).
   `TestParityZendesk_StartDateConfigNeverSendsServerFilter` (the inverted
   test) went GREEN once the param was removed:
   `go test -run TestParityZendesk_StartDateConfigNeverSendsServerFilter -v`
   → PASS.
   - Deleted `TestParityZendesk_StartDateConfigRaisesLowerBound` (replaced in
     place by the inverted `..._NeverSendsServerFilter` test — same function
     slot, opposite assertion, per the plan's "invert/delete" instruction).
   - Fixed the self-contradicting comment above
     `TestParityZendesk_IncrementalConfigAcceptedWithoutServerFilter`
     (formerly parity_test.go:297-306): the comment already correctly said
     "this bundle therefore intentionally does NOT declare an
     incremental.request_param" while the shipped `streams.json` declared it
     on all 5 streams and the very next test asserted the param WAS sent —
     now both the code and the comment agree (no `incremental` block
     anywhere, no request_param ever sent, sibling test asserts exactly
     that).
   - `docs.md` "Known limits": replaced the "documented parity deviation"
     paragraph (which rationalized the invented filter as an "acceptable
     capability addition") with an honest "No server-side incremental filter
     (matches legacy exactly)" paragraph explaining what was wrong, what
     changed, and pointing at the real Zendesk incremental export endpoints
     (`/api/v2/incremental/*`) as the Pass B path for true incremental sync.
   - `TEST-PLAN.md` row amended: "Incremental parity" column for
     zendesk-support changed from `✓ start_date-raised` to `N/A — legacy
     sends no incremental filter (gap-loop cycle-1 fix; original
     "start_date-raised" row was a planning error — no such wire behavior
     exists to parity-test)`.
   - **Second-order dead-key consequence (not explicitly named in the plan
     bullet, but required by the same F6 rule already being enforced):**
     with no `incremental` block on any stream, `spec.json`'s `start_date`
     property became wired to nothing at all (confirmed by reading
     `engine/read.go`'s `incrementalLowerBoundValue`: it returns `""`
     unconditionally when `stream.Incremental == nil`, so `start_config_key`/
     `start_date` are never consulted). Dropped `start_date` from
     `spec.json` alongside `page_size`/`max_pages`/`mode` rather than leave
     a newly-dead key behind.

2. **Major — pagination stop-signal (REVIEW-B finding 2).** Added
   `"stop_path": "meta.has_more"` to `streams.json`'s `base.pagination`
   block (`type: cursor`, `token_path: meta.after_cursor`). This uses the
   `stop_path` support the Step 1 engine mini-wave already added to
   `engine/paginate.go`'s `tokenPathCursor` (gap-loop cycle-1 item 5 —
   `stopPath` falsy-stops via `connsdk.StringAt`, any value other than the
   literal string `"true"` stops unconditionally, matching legacy's
   `hasMore != "true" || nextCursor == ""` rule exactly). No engine code
   changed by this task (connsdk/paginate.go and engine/paginate.go are
   Step-1/mini-wave deliverables, outside this task's editable file set);
   only the bundle's declarative `streams.json` was touched.
   Regression-tested by the new
   `TestParityZendesk_HasMoreFalseWithNonNullCursorStopsPagination`
   (has_more:false + non-null after_cursor page): legacy and engine both
   issue exactly 1 request and emit exactly 1 record — went GREEN once
   `stop_path` was declared.
   - `docs.md` "Streams notes" and "Known limits" rewritten to describe the
     real stop rule (`stop_path: meta.has_more`, falsy-stops regardless of
     cursor value) instead of the prior unverified "live-API behavior always
     emit after_cursor: null on the final page" claim, which cited only
     legacy's own test fixture as evidence and was never exercised against
     the divergent case.

3. **Minor — dead spec keys (REVIEW-B finding 3 + the start_date
   consequence above).** Deleted `page_size`, `max_pages`, `mode`, and
   `start_date` from `spec.json` — none is consumed by any template or
   engine mechanism (`page[size]=100` is a static query literal per the
   stripe/bitly pattern; `mode`/`max_pages` were never wired to fixture-mode
   or a page cap in the engine dialect at all).

4. **Minor — `email` config-key/dotted-secret-key documentation gap
   (REVIEW-B finding 4).** Added a "Documented config-surface deviation
   (email)" paragraph to `docs.md`'s Auth setup section: legacy additionally
   accepts dotted `credentials.api_token`/`credentials.email` secret keys
   AND a plain non-secret `config.email` fallback
   (zendesk_support.go:271-287,366-378); this bundle canonicalizes to bare
   `email`/`api_token` secrets only, narrowing the non-secret `config.email`
   path specifically. No request/data change for any input using the
   canonical secret keys.

5. **docs.md honesty (dispatch item 4).** Beyond the `after_cursor` claim
   fixed under item 2 above, also corrected the incremental-deviation
   paragraph (item 1) and added the email config-surface note (item 4) so
   every remaining docs.md claim traces to actual wired behavior.

## Final self-verify (green)

```
go build ./internal/connectors/...                                    # exit 0
go vet ./internal/connectors/paritytest/zendesk-support/... \
       ./internal/connectors/defs/...                                 # exit 0
gofmt -l internal/connectors/paritytest/zendesk-support/parity_test.go # empty (formatted)

go run ./cmd/connectorgen validate internal/connectors/defs --json \
  | jq '.findings | map(select(.connector=="zendesk-support"))'
# -> [] (0 findings)
go run ./cmd/connectorgen validate internal/connectors/defs --json | jq '{findings: (.findings|length)}'
# -> {"findings": 0}   (0 findings repo-wide)

go test ./internal/connectors/conformance -run 'TestConformance/zendesk-support' -v
# --- PASS: TestConformance (0.01s)
#     --- PASS: TestConformance/zendesk-support (0.01s)

go test ./internal/connectors/paritytest/zendesk-support/... -v
# PASS — all 9 top-level tests (5 stream subtests under StreamRecords):
#   TestParityZendesk_StreamRecords (+5 subtests: tickets/users/organizations/groups/satisfaction_ratings)
#   TestParityZendesk_TicketsTwoPagePagination
#   TestParityZendesk_HasMoreFalseWithNonNullCursorStopsPagination   (NEW)
#   TestParityZendesk_IncrementalConfigAcceptedWithoutServerFilter
#   TestParityZendesk_StartDateConfigNeverSendsServerFilter          (INVERTED, was StartDateConfigRaisesLowerBound)
#   TestParityZendesk_APITokenBasicAuthParity
#   TestParityZendesk_OAuthBearerAuthParity
#   TestParityZendesk_ErrorPathParity
#   TestParityZendesk_BundleLoadsAndValidates

make lint
# golangci-lint run ./internal/connectors/engine/... ./internal/connectors/defs/... \
#   ./internal/connectors/hooks/... ./internal/connectors/native/... \
#   ./internal/connectors/conformance/... ./internal/connectors/certify/... \
#   ./cmd/connectorgen/... ./cmd/inventorygen/...
# 0 issues.

go build ./...                                                         # exit 0
go test ./internal/connectors/... ./cmd/...
# ok  	polymetrics.ai/internal/connectors/paritytest/zendesk-support
# FAIL polymetrics.ai/internal/connectors/paritytest/gmail       (TestParityGmail_ComputedFieldsStringifyLabelCountFields — pre-existing/other-agent scope, gmail dir has NO working-tree diff from this task)
# FAIL polymetrics.ai/internal/connectors/conformance            (TestConformance/github fails; TestConformance/zendesk-support PASSES — github dir has NO working-tree diff from this task)
```

Both whole-tree failures above are OUTSIDE this task's disjoint file set
(`defs/zendesk-support/**`, `paritytest/zendesk-support/**`): `git status
--short` shows no modification to `internal/connectors/defs/gmail`,
`internal/connectors/paritytest/gmail`, `internal/connectors/defs/github`, or
`internal/connectors/paritytest/github` from this session — those bullets
("github (majors)", "gmail (majors)") belong to sibling backend-repair
agents per GAP-LOOP-PLAN.md Step 2's "parallel, disjoint dirs" instruction.
Isolated re-runs confirm zendesk-support and every shared package this
bundle depends on (`engine`, `connsdk`, `conformance/TestConformance/
zendesk-support`, `connectorgen`) are green.

## Parity-deviation ledger (conventions.md §5 format) — updated

| connector | description | verdict |
|---|---|---|
| zendesk-support | `spec.json` declares `base_url` as the sole required config (no `subdomain` property) whereas legacy accepts EITHER a bare `subdomain` (deriving `https://<subdomain>.zendesk.com/api/v2` itself) OR a `base_url` override. Engine's declarative `url` template cannot branch on two config keys. Every legacy-accepted `subdomain`-only configuration remains reachable via the equivalent `base_url`. | ACCEPTABLE (config-surface narrowing, operator workaround always available) |
| zendesk-support | `streams.json`'s `base.auth` lists OAuth Bearer BEFORE Basic API-token to reproduce legacy's exact access_token-first precedence when both secrets are present. | ACCEPTABLE (mitigated — no deviation once correctly ordered) |
| zendesk-support | `email` is secrets-only; legacy also accepts dotted `credentials.email` and a plain non-secret `config.email` fallback. Canonicalization narrows the non-secret fallback path specifically. | ACCEPTABLE (documented, no request/data change for canonical-key inputs) |
| ~~zendesk-support~~ | ~~Bundle declares `incremental.request_param: "updated_at[gte]"` ... "capability addition" framing~~ | **SUPERSEDED THIS CYCLE — REJECTED, not acceptable.** REVIEW-B.md's adjudication: narrows sync SCOPE against inputs legacy accepts and ignores, which IS changed emitted data under the §5 meta-rule, and the param is not real Zendesk API surface (silent no-op in production). Removed entirely; see fix 1 above. |

## Notes on scope discipline

- Did not touch `connsdk/paginate.go` or `engine/paginate.go` (Step-1
  engine-mini-wave deliverables, explicitly out of this task's editable file
  set per the ledger note in `engine/paginate.go`'s `tokenPathCursor` doc
  comment) — only consumed the already-shipped `stop_path` field via
  `streams.json`.
- Did not touch legacy `internal/connectors/zendesk-support/` (read-only
  ground truth per dispatch).
- No dependency additions, no schema/migration changes, no auth/security
  changes, no destructive data actions, no secret access — no human gate
  triggered.

# Verification checklist — transport source eligibility club

## Scope correction — 2026-08-16

PR 4175 owns #3976. This branch retains PostgreSQL `polling_watermark` as `planned` because no
shipped-binary production preflight can yet bind its dynamic source/object/destination contract.
The CLI inspection guard remains unchanged and is re-run as the green proof.

## Behavioral acceptance

- [x] GitHub `commits` is explicitly listed and accepted through production composition.
- [x] GitHub eligibility contains no wildcard and matches the executable declarative stream set.
- [x] A stream absent from the declaration returns `SourceStreamIneligibleError` before I/O.
- [x] Typed refusal records zero source requests, warehouse pages, target sends/rows, and checkpoint
      movement.
- [x] Case-equivalent stream identifiers remain absent from the exact allowlist and are refused.
- [x] Omitted `max_pages` reads one provider page; a positive integer caps; `0`, `all`, and
      `unlimited` exhaust pagination.
- [x] Declarative provider callbacks remain bounded transport pages and no 1,000-record global cap
      discards a large collection.
- [x] PostgreSQL polling is not advertised as executable: inspection reports `planned` and a
      blocking reason until a production preflight can bind it.
- [x] Polling remains distinct from CDC; snapshot and explicit bootstrap/CDC behavior remain
      unchanged.
- [x] Cross-family API→database conformance is proven from `app.Open`; PostgreSQL polling is
      intentionally deferred to PR 4175 rather than hand-built or claimed here.

## Required edge matrix

| Edge | Expected outcome and effect assertion | Evidence |
| --- | --- | --- |
| cancellation mid-run | context cancellation; interrupted candidate not committed; bounded resources close | green locally — `TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply`; binary proof integration-gated |
| process death partway | no persisted candidate for unacknowledged page; restart replays it | green locally — GitHub composed source replay |
| empty input | zero target sends/rows; existing checkpoint unchanged | green locally — `TestOrchestratorAdmitsEmptyResultOnlyFromExplicitSourceMarker`; binary proof integration-gated |
| single row | exactly one delivered row and one durable acknowledged candidate | covered by existing transport/orchestrator evidence; no PostgreSQL polling claim in this PR |
| large input | all independently counted rows delivered in bounded pages | green locally for 103 committed fixtures; 99,345 live certification pending |
| duplicate delivery | keyed destination remains one logical row or typed refusal before effects | covered at the transport/orchestrator boundary; PostgreSQL polling deferred to PR 4175 |
| out-of-order delivery | typed traversal/order error; zero stage/send/checkpoint for refused page | PostgreSQL polling deferred to PR 4175 |
| schema drift | typed rebootstrap outcome; zero fetch/apply/checkpoint under mismatched fingerprint | PostgreSQL polling deferred to PR 4175 |
| auth refusal | typed credential/auth admission error; zero provider query/stage/send/checkpoint | GitHub live credential rotation pending; PostgreSQL polling deferred to PR 4175 |
| concurrent same-target runs | lease/CAS fencing prevents double commit; loser has zero state advance | green locally — `internal/app/transport_dispatch_test.go` state-conflict coverage |
| resume after interruption | strict tuple resumes after last acknowledged row with no skip | green locally; duplicate cursor delivery integration-gated |
| replay acknowledged item | durable target idempotency prevents an added row; checkpoint is monotonic | green locally at orchestrator boundary; binary proof integration-gated |
| undeclared stream | `SourceStreamIneligibleError`; zero source/stage/send/row/checkpoint | green locally — includes case-equivalent `ISSUES` refusal |

## Live evidence

- [ ] Authenticated `rails/rails` `commits` run uses `max_pages=unlimited` — pending: certification
      credential rotation is incomplete.
- [ ] Independent warehouse extracted count recorded — pending with the live run.
- [ ] Independent PostgreSQL delivered count recorded against the 99,345-row reference — pending
      with the live run; no count is claimed.
- [x] No credential value was read, printed, stored, or placed in command text, traces, fixtures,
      artifacts, git diff, or the PR body while the rotated live credential is pending.

## Focused local gates

- [x] `gofmt -w` on changed Go packages.
- [x] Focused tests with `-count=1` for `internal/synctransport`, `internal/app`, and
      `internal/cli`; the planned-polling CLI guard is rerun after deconfliction.
- [x] Focused race tests for concurrency-bearing changed packages.
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] `make tidy-check`.
- [x] `make lint`.
- [x] `make docs-check-no-build`.
- [x] `make smoke-no-build`.
- [x] `make agent-contract-check`.
- [x] `make connectorgen-validate`.
- [x] `make connectorgen-surface-sync` after one-pass derived regeneration.
- [x] `make connector-boundary`.
- [x] `make release-workflow-check`.
- [x] `go run ./cmd/agentcontractgen check`.
- [x] `scripts/verify-gsd-workflow` against the final diff.
- [x] `go test -timeout 20m -count=1 ./internal/connectors/certify`, including both declaration
      registration guards after the exact source-reference update.
- [x] Inline `scripts/gsd prompt verify-work issue-4171-3976-3862-transport-eligibility-r1` record.
- [x] Inline `scripts/gsd prompt code-review issue-4171-3976-3862-transport-eligibility-r1` record and
      all actionable findings dispositioned.
- [x] `fm-ensure-agents-md.sh .` run; no unrelated project memory appended.
- [x] Generated drift checks pass; the worktree is clean after the final evidence commit.

## Delivery gate

- [ ] Conventional PR title and body link `Closes #4171, #3862` and records #3976's deconfliction
      to PR 4175.
- [ ] PR body contains the GitHub production call chain, live counts/unavailable statement, edge table,
      required skills, focused gates, and deliberate parity/non-goal notes.
- [ ] Branch pushed only to `fm/cli-transport-stream-eligibility-club-r1`.
- [ ] PR base API read-back equals `integration/4015-mvp-flat-r1`.
- [ ] Final sparse status line is `done: PR <url>`.

## Manual verify-work and review record

The official `verify-work` and `code-review` prompts were resolved with `scripts/gsd prompt`.
This non-numbered, direct-PR issue club is executed inline because the task requires one autonomous
worker and the canonical contract forbids lifecycle role spawning. Verification used the focused
commands above rather than the disallowed monolithic suite. Manual source review covered the
registry admission order, production dispatch, checkpoint sequencing, keyset tuple encoding, and
generated declarations. No critical or warning finding remained; see `REVIEW.md`.

### Gap G1 — certification registration guard

Clean comparison before the fix: `ef3c71caf` passed both
`TestCertificationDeclaredTransportPair*` tests; `73280ed81` failed both because the test still
named the superseded `issue_label_source` reference. The guard remains exact: it now requires the
new declaration-owned `declarative_stream_source` reference on both the success and intentionally
unregistered paths. The focused certificate package, app, and transport packages pass.

### Gap G2 — planned polling truthfulness

At `f1873c340`, the unchanged real CLI inspection test fails because the definition reports
`polling_watermark.status=implemented`. Static call-chain review shows `app.Open` only composes the
outer transport; the attempted `engine.PollingPreflight` bind occurs within `ReadTransport` after
authentication admission and typed-catalog I/O. The green repair restores the existing planned
declaration and removes the overlapping adapter work, which PR 4175 owns.

Green: `go test -v -count=1 -timeout 20m ./internal/cli -run
'^TestInspectPostgresKeepsPollingWatermarkPlannedUntilPreflightCanBindIt$'` passes without changing
the guard. Generator-derived skills, connector docs, and `connectors_inspect_github_json` golden
also pass after their real generator targets ran.

# Verification checklist — transport source eligibility club

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
- [x] PostgreSQL's implemented polling declaration selects exact native source/apply references and
      immutable evidence.
- [x] Production composition invokes the shared `engine.PollingSourceExecutor`; no capability or
      connector-name fallback exists.
- [x] Native PostgreSQL polling preserves a lossless cursor plus the complete composite primary-key
      tie tuple and uses a strict lexicographic resume predicate.
- [x] Polling is distinct from CDC; snapshot and explicit bootstrap/CDC behavior remains green.
- [x] Cross-family API→database and database→database conformance is proven from `app.Open` or built
      `pm`, not a hand-built component.

## Required edge matrix

| Edge | Expected outcome and effect assertion | Evidence |
| --- | --- | --- |
| cancellation mid-run | context cancellation; interrupted candidate not committed; bounded resources close | green locally — `TestOrchestratorStopsOnCancellationBetweenWarehouseStageAndApply`; binary proof integration-gated |
| process death partway | no persisted candidate for unacknowledged page; restart replays it | green locally — GitHub composed source replay; PostgreSQL live test integration-gated |
| empty input | zero target sends/rows; existing checkpoint unchanged | green locally — `TestOrchestratorAdmitsEmptyResultOnlyFromExplicitSourceMarker`; binary proof integration-gated |
| single row | exactly one delivered row and one durable acknowledged candidate | integration-gated native/binary coverage; not run because shared runtime is unavailable |
| large input | all independently counted rows delivered in bounded pages | green locally for 103 committed fixtures; 99,345 live certification pending |
| duplicate delivery | keyed destination remains one logical row or typed refusal before effects | integration-gated binary replay coverage; not run because shared runtime is unavailable |
| out-of-order delivery | typed traversal/order error; zero stage/send/checkpoint for refused page | green locally — native strict traversal test |
| schema drift | typed rebootstrap outcome; zero fetch/apply/checkpoint under mismatched fingerprint | integration-gated binary/native coverage; not run because shared runtime is unavailable |
| auth refusal | typed credential/auth admission error; zero provider query/stage/send/checkpoint | integration-gated binary coverage; credential rotation pending |
| concurrent same-target runs | lease/CAS fencing prevents double commit; loser has zero state advance | green locally — `internal/app/transport_dispatch_test.go` state-conflict coverage |
| resume after interruption | strict tuple resumes after last acknowledged row with no skip | green locally; duplicate cursor delivery integration-gated |
| replay acknowledged item | durable target idempotency prevents an added row; checkpoint is monotonic | green locally at orchestrator boundary; binary proof integration-gated |
| undeclared stream | `SourceStreamIneligibleError`; zero source/stage/send/row/checkpoint | green locally — includes case-equivalent `ISSUES` refusal |

## Live evidence

- [ ] Real PostgreSQL container starts through the repository harness and passes the polling
      integration route — pending: the task explicitly prohibits retrying the unavailable shared
      container runtime.
- [ ] Authenticated `rails/rails` `commits` run uses `max_pages=unlimited` — pending: certification
      credential rotation is incomplete.
- [ ] Independent warehouse extracted count recorded — pending with the live run.
- [ ] Independent PostgreSQL delivered count recorded against the 99,345-row reference — pending
      with the live run; no count is claimed.
- [x] The unavailable result is recorded without a count claim: databaseintegration tests were
      invoked but skipped by their opt-in guard; no container or credential retry was attempted.
- [x] No credential value was read, printed, stored, or placed in command text, traces, fixtures,
      artifacts, git diff, or the PR body while the rotated live credential is pending.

## Focused local gates

- [x] `gofmt -w` on changed Go packages.
- [x] Focused tests with `-count=1` for `internal/synctransport`,
      `internal/connectors/engine`, `internal/connectors/native/postgres`, `internal/app`, and
      `internal/cli`.
- [x] Focused race tests for concurrency-bearing changed packages.
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] PostgreSQL `databaseintegration` focused package test invoked; skipped by its required
      opt-in because the container is unavailable.
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

- [ ] Conventional PR title and body link `Closes #4171`, `Closes #3976`, and `Closes #3862`.
- [ ] PR body contains both production call chains, live counts/unavailable statement, edge table,
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

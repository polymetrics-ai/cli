# Verification checklist — transport source eligibility club

## Behavioral acceptance

- [ ] GitHub `commits` is explicitly listed and accepted through production composition.
- [ ] GitHub eligibility contains no wildcard and matches the executable declarative stream set.
- [ ] A stream absent from the declaration returns `SourceStreamIneligibleError` before I/O.
- [ ] Typed refusal records zero source requests, warehouse pages, target sends/rows, and checkpoint
      movement.
- [ ] Omitted `max_pages` reads one provider page; a positive integer caps; `0`, `all`, and
      `unlimited` exhaust pagination.
- [ ] Declarative provider callbacks remain bounded transport pages and no 1,000-record global cap
      discards a large collection.
- [ ] PostgreSQL's implemented polling declaration selects exact native source/apply references and
      immutable evidence.
- [ ] Production composition invokes the shared `engine.PollingSourceExecutor`; no capability or
      connector-name fallback exists.
- [ ] Native PostgreSQL polling preserves a lossless cursor plus the complete composite primary-key
      tie tuple and uses a strict lexicographic resume predicate.
- [ ] Polling is distinct from CDC; snapshot and explicit bootstrap/CDC behavior remains green.
- [ ] Cross-family API→database and database→database conformance is proven from `app.Open` or built
      `pm`, not a hand-built component.

## Required edge matrix

| Edge | Expected outcome and effect assertion | Evidence |
| --- | --- | --- |
| cancellation mid-run | context cancellation; interrupted candidate not committed; bounded resources close | pending |
| process death partway | no persisted candidate for unacknowledged page; restart replays it | pending |
| empty input | zero target sends/rows; existing checkpoint unchanged | pending |
| single row | exactly one delivered row and one durable acknowledged candidate | pending |
| large input | all independently counted rows delivered in bounded pages | pending |
| duplicate delivery | keyed destination remains one logical row or typed refusal before effects | pending |
| out-of-order delivery | typed traversal/order error; zero stage/send/checkpoint for refused page | pending |
| schema drift | typed rebootstrap outcome; zero fetch/apply/checkpoint under mismatched fingerprint | pending |
| auth refusal | typed credential/auth admission error; zero provider query/stage/send/checkpoint | pending |
| concurrent same-target runs | lease/CAS fencing prevents double commit; loser has zero state advance | pending |
| resume after interruption | strict tuple resumes after last acknowledged row with no skip | pending |
| replay acknowledged item | durable target idempotency prevents an added row; checkpoint is monotonic | pending |
| undeclared stream | `SourceStreamIneligibleError`; zero source/stage/send/row/checkpoint | pending |

## Live evidence

- [ ] Real PostgreSQL container starts through the repository harness and passes the polling
      integration route.
- [ ] Authenticated `rails/rails` `commits` run uses `max_pages=unlimited`.
- [ ] Independent warehouse extracted count recorded.
- [ ] Independent PostgreSQL delivered count recorded against the 99,345-row reference.
- [ ] If the target/container cannot start, the exact unavailable result is recorded and no count
      is claimed.
- [ ] Credential value is read directly into the process environment and never appears in command
      text output, traces, fixtures, artifacts, git diff, or PR body.

## Focused local gates

- [ ] `gofmt -w` on changed Go packages.
- [ ] Focused tests with `-count=1` for `internal/synctransport`,
      `internal/connectors/engine`, `internal/connectors/native/postgres`, `internal/app`, and
      `internal/cli`.
- [ ] Focused race tests for concurrency-bearing changed packages.
- [ ] `go vet ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] PostgreSQL `databaseintegration` focused package test when the container is available.
- [ ] `make tidy-check`.
- [ ] `make lint`.
- [ ] `make docs-check`.
- [ ] `make smoke-no-build`.
- [ ] `make agent-contract-check`.
- [ ] `make connectorgen-validate`.
- [ ] `make connectorgen-surface-sync` after one-pass derived regeneration.
- [ ] `make connector-boundary`.
- [ ] `make release-workflow-check`.
- [ ] `go run ./cmd/agentcontractgen check`.
- [ ] `scripts/verify-gsd-workflow` against the final diff if its interface supports local use.
- [ ] Inline `scripts/gsd prompt verify-work issue-4171-3976-3862-transport-eligibility-r1` record.
- [ ] Inline `scripts/gsd prompt code-review issue-4171-3976-3862-transport-eligibility-r1` record and
      all actionable findings dispositioned.
- [ ] `fm-ensure-agents-md.sh .` run; no unrelated project memory appended.
- [ ] Generated drift checks pass and `git status` is clean after the final commit.

## Delivery gate

- [ ] Conventional PR title and body link `Closes #4171`, `Closes #3976`, and `Closes #3862`.
- [ ] PR body contains both production call chains, live counts/unavailable statement, edge table,
      required skills, focused gates, and deliberate parity/non-goal notes.
- [ ] Branch pushed only to `fm/cli-transport-stream-eligibility-club-r1`.
- [ ] PR base API read-back equals `integration/4015-mvp-flat-r1`.
- [ ] Final sparse status line is `done: PR <url>`.


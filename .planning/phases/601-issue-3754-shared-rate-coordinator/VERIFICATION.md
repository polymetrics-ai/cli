# Phase 601 verification checklist — #3754

**Status:** `verify-work`, inline deep code review, and bounded supervisor
evidence passed; terminal PR gates remain pending.

## Required behavior

- [x] Existing `RateLimitRegistry`/`RateLimiter` behavior remains a
      dependency-free `process_local` backend.
- [x] Matching policies are reserved atomically with one in-flight lease; a
      non-grant consumes neither policy capacity nor lease occupancy.
- [x] Policy identity contains only a deterministic typed-contract fingerprint
      and #3863 opaque rate scope; no raw credential, revision, subject, URL,
      request body/header/variables, endpoint, or epoch reaches evidence.
- [x] UDS owner directory/socket modes are 0700/0600; protocol is versioned,
      closed, length-bounded, context/deadline-aware, and has no generic RPC.
- [x] `Finish` is opaque-lease-idempotent, releases concurrency, retains
      indeterminate consumption, and applies only stricter observations.
- [x] `require_shared` has no local fallback; missing/dead/old/incompatible
      owners refuse before a transport request.
- [x] Deadline-too-short refusal and owner crash have zero unapproved sends.
- [x] Eight actual helpers produce shared 3 grants/5 blocks and local 8
      grants; normal close proves zero socket/run-directory residue.

## Local commands to record after GREEN

- [x] Focused mandatory suite and focused requester/engine coverage.
- [x] `go test -race -count=1 -timeout 20m ./internal/coordination`.
- [x] `gofmt` check, targeted `go vet`, and `go build ./cmd/pm`.
- [x] Individual non-monolithic gates required by AGENTS.md: tidy-check,
      lint, docs-check, smoke-no-build, agent-contract-check,
      connectorgen-validate, connectorgen-surface-sync, connector-boundary,
      and release-workflow-check.
- [x] `scripts/verify-gsd-workflow` against the parent base after production
      evidence lands.
- [x] Generated inline `verify-work` with coverage-aware UAT evidence.
- [x] Generated inline `code-review` with one resolved warning.
- [x] #3995-compatible bounded supervisor evidence.
- [ ] Terminal child no-mistakes gate.

## Executed verification

All commands below exited 0 on 2026-08-11:

```text
go test -count=1 -timeout 20m ./internal/coordination ./internal/connectors/engine -run '^(TestRateBudgetReserveBatchAllOrNothing|TestUnixRateBudgetCoordinatorMultiProcessTinyBudget|TestRequireSharedRefusesWithoutCoordinatorBeforeSend|TestSharedRateBudgetScopesRemainIndependent|TestSharedRateBudgetDeadlineTooShortDoesNotSend|TestSharedRateBudgetOwnerCrashFailsClosed)$'
go test -count=1 -timeout 20m ./internal/coordination
go test -count=1 -timeout 20m ./internal/connectors/connsdk
go test -count=1 -timeout 20m ./internal/connectors/engine
go test -race -count=1 -timeout 20m ./internal/coordination
go vet ./internal/coordination ./internal/connectors/connsdk ./internal/connectors/engine ./internal/connectors
go build ./cmd/pm
golangci-lint run --new-from-rev=origin/docs/4015-connector-release-certification ./internal/coordination/... ./internal/connectors/connsdk/... ./internal/connectors/engine/...
make tidy-check
make lint
make docs-check
make smoke-no-build
make agent-contract-check
make connectorgen-validate
make connectorgen-surface-sync
make connector-boundary
make connector-runtime-preflight
make connector-canon-check
make release-workflow-check
scripts/verify-gsd-workflow origin/docs/4015-connector-release-certification
```

`601-UAT.md` records the coverage-aware inline `verify-work` outcome. The
multi-process test itself asserts the 3/5 shared and 8/0 local outcomes, the
0700/0600 permission modes, and absence of the owned socket/run directory
after close; no endpoint or epoch is recorded in this evidence.

## Explicit non-applicability

No CLI command/flag/help/manual/website output changes, no provider/GraphQL
code, no connector bundle/generated surface, no external service, and no live
credentialed check is in this phase. The PR records that scope fence rather
than claiming parity work that did not apply.

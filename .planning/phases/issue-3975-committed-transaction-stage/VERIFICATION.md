# Verification checklist — Issue #3975: committed-transaction staging and durable receipts

**Status:** Local verification complete; no-mistakes and remote delivery hold pending.

## Scope and topology proof

- [x] #3975 is implemented only in `internal/connectors/database/**` plus
      issue-local GSD evidence. No PostgreSQL decoder/CDC, LSN acknowledgement,
      poller, target/DML, generic SQL, CLI, warehouse, docs, or capability
      promotion changed.
- [x] #3974 is an ancestor of this child through
      `be561871e6bb7d1a5b54d7687743ef8396a2cafe`.
- [x] The immutable audit checkpoint is
      `2afa128e5a9844863c7b33d4ca52dacb867398ed`; it was not rewritten or
      force-pushed. The user directed no subsequent push or PR action before
      final local gates.

## Required behaviour proof

- [x] `TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt` has a
      retained behavioural RED trace and a GREEN/reopen proof: chunks remain
      private until commit, stream in append order once, and only a persisted
      receipt exposes `synccontract.DownstreamAcknowledgement` eligibility.
- [x] Abort deletes active chunks and supports clean identity reuse. Quota
      refusals name `TransactionStageLimitExceeded` for byte, record, age, and
      retained-root byte limits, publish nothing, and leave no final artifact.
- [x] Fixed 32 KiB copying is exercised with a larger controlled reader; the
      stage never accumulates the full payload in memory.
- [x] Cancellation during append and after receiver delivery leaves no final
      chunk, receipt, acknowledgement, or temporary residue.
- [x] Restart removes active/incomplete/orphan state, retains only sealed
      receipt-less work for retry, preserves recovered byte accounting, and
      removes post-receipt residue on recovery.
- [x] ENOSPC, write, file-sync, rename, parent-directory-sync, receiver,
      receipt, and cleanup transitions use injected faults. Each failure
      asserts durable artifacts/receipts and reopens the root at the boundary.
- [x] A receiver must stream every complete chunk; a false receipt return
      without full consumption is refused. A caller cannot forge receipt-based
      acknowledgement eligibility.
- [x] Opaque transaction identities are hashed before filesystem use; traversal
      and control-character values remain root-contained and distinct.

## Executed focused/static checks

- [x] `go test -timeout 20m -count=1 ./internal/connectors/database -run '^TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt$'`
- [x] `go test -timeout 20m -count=1 ./internal/connectors/database -run 'Cancellation|Fault|RestartRecovery|RecoveryCleans'`
- [x] `go test -timeout 20m -count=1 ./internal/connectors/database`
- [x] `go test -timeout 20m -race -count=1 ./internal/connectors/database`
- [x] `go test -timeout 20m -count=1 ./internal/synccontract`
- [x] `go test -timeout 20m -count=1 ./internal/app`
- [x] `gofmt -w` changed Go files, `git diff --check`,
      `golangci-lint run ./internal/connectors/database/...`, and `go vet ./...`
- [x] `go build ./cmd/pm`

## Executed repository gates

- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make connector-boundary`
- [x] `make release-workflow-check`

`make verify` is intentionally not invoked as one command in the per-command
timeout harness. Its documented component gates are listed above; CI remains
the full-suite authority.

## GSD/manual delivery gates

- [x] Adapter prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
      `verify-work`, and `code-review` were resolved. #3975 is not a numbered
      roadmap phase, and the canonical single-worker contract prohibits GSD
      role spawning, so this is an explicit manual inline fallback.
- [x] Inline UAT is recorded in `UAT.md`; the code-review disposition is in
      `REVIEW.md`; no human-judgment UI/provider mutation is in scope.
- [x] Bounded #3995-compatible Shepherd evidence is in
      `SHEPHERD-COMPATIBILITY.{md,json}` and does not claim an automatic
      approval.
- [ ] Canonical child no-mistakes run is pending on the local implementation
      commit, without `--yes` and with push/PR/CI explicitly skipped by the
      child contract.
- [ ] No push, PR creation, remote CI observation, automated-review request,
      or merge is permitted until the user-directed hold is lifted after the
      no-mistakes result.

## Deliberate non-applicability

No public CLI command/flag/help/manual/website or connector surface changed.
No credentialed, live-database, Podman, PostgreSQL wire, source checkpoint, or
destination interaction is applicable to this private source-agnostic stage.

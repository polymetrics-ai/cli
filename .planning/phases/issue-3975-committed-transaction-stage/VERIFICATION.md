# Verification checklist — Issue #3975: committed-transaction staging and durable receipts

**Status:** Planned — evidence pending execution

## Scope and dependency proof

- [x] #3975 issue is open, uniquely mapped to this child branch, and has no
      existing child PR/branch to duplicate.
- [x] #3974 is present at `be561871e` in
      `origin/feat/3972-postgres-parity`; child branch starts at that ref.
- [x] Parent PR #4017 exists as a draft and child base is exactly
      `feat/3972-postgres-parity`.
- [ ] Diff stays within `internal/connectors/database/**` and this phase's
      planning evidence; no PostgreSQL/CDC/app/warehouse/CLI behavior change.

## Required behaviour proof

- [ ] `TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt` has a
      captured behavioural RED run and a subsequent GREEN run.
- [ ] Chunk invisibility before commit, atomic in-order whole-transaction
      publication, and one immutable receipt are covered by assertions.
- [ ] Abort, missing/duplicate boundaries, independent transaction isolation,
      and zero-residue active cleanup are covered.
- [ ] Byte, record, and time limits return named
      `TransactionStageLimitExceeded`, publish nothing, and expose no
      acknowledgement eligibility.
- [ ] Bounded-memory streaming is proven with a controlled large reader rather
      than a whole-payload in-memory test fixture.
- [ ] Receipt-before-acknowledgement is proven through a fake source/consumer
      and a persisted receipt artifact, not a return code.
- [ ] Disk-full, write/fsync/rename/parent-sync, receiver, receipt, and
      cancellation failures leave no deceptive final receipt/eligibility.
- [ ] Every crash boundary restarts into only safe states and cleans incomplete
      stage/orphan temporary residue.
- [ ] Recovered accounting prevents a restart from bypassing stage limits.

## Focused checks

- [ ] `go test -timeout 20m -count=1 ./internal/connectors/database -run '^TestCommittedTransactionStagePublishesOnlyAfterDurableReceipt$'`
- [ ] `go test -timeout 20m -count=1 ./internal/connectors/database`
- [ ] `go test -timeout 20m -race -count=1 ./internal/connectors/database`
- [ ] `go test -timeout 20m -count=1 ./internal/synccontract ./internal/app`
- [ ] `go vet ./...`
- [ ] `go build ./cmd/pm`
- [ ] direct changed-package `golangci-lint` or repository `make lint`

## Repository gates

- [ ] `make tidy-check`
- [ ] `make lint`
- [ ] `make docs-check`
- [ ] `make smoke-no-build`
- [ ] `make agent-contract-check`
- [ ] `make connectorgen-validate`
- [ ] `make connectorgen-surface-sync`
- [ ] `make connector-boundary`
- [ ] `make release-workflow-check`
- [ ] `make verify` is attempted only if the harness can give the full suite
      its required time; otherwise individual gates above plus CI are recorded.

## Delivery and review gates

- [ ] Generated `execute-phase`, `verify-work`, and `code-review` prompts are
      recorded with the manual inline fallback.
- [ ] Manual GSD verification/UAT, gaps loop if needed, and code-review
      dispositions are completed.
- [ ] Bounded #3995-compatible Shepherd evidence is recorded; no automatic
      Shepherd verdict is claimed while #3995 is outside the parent ancestry.
- [ ] `no-mistakes axi run` executes the canonical child command without
      `--yes`; every gate is individually dispositioned.
- [ ] Child branch is pushed and one draft PR targets
      `feat/3972-postgres-parity` with Refs #3975/#3972/#2986/#2988.
- [ ] CI and the stacked automated-review route are recorded; no child merge is
      performed.

## Deliberate non-applicability

No public CLI command, flag, output, connector manifest, help/manual, website
page, credentialed check, or live database interaction changes. CLI/docs/site
parity and Podman/database integration are not applicable; PostgreSQL remains
fail-closed and is covered by the final scope scan.

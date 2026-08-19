# Verification: Gong release-0.3.0 live parity reconciliation

## Credential-free verification

- [x] Isolated worktree and preserved remote branch identity verified.
- [x] `no-mistakes doctor` passed; no daemon action was taken.
- [x] `scripts/gsd doctor` passed and command sources were resolved.
- [x] Current official OpenAPI refetch returned 69 operations with the current GET/POST/PUT/PATCH/DELETE distribution.
- [x] Exact method/path/operation-ID/deprecation comparison against Batch 2/3 Gong source lock passed.
- [ ] Current-main, typed-destination, and Batch 2/3 foundation reconciliation completed.
- [ ] `go run ./cmd/connectorgen validate --json` passes after reconciliation.
- [ ] `go run ./cmd/connectorgen surface-sync --check` passes after reconciliation.
- [ ] `go run ./cmd/connectorgen surface-reconcile --check --json` reports no Gong runtime-surface drift.
- [ ] Focused Gong `connectorgen`, engine, commandrunner, conformance, application, and CLI tests pass with `-timeout 20m`.
- [ ] Built `pm` credential-free direct-read sweep records every result classification and proves no `unknown command` or exact-endpoint preflight block.
- [ ] `pm help gong`, `pm gong`, and affected command help/docs/website generated-artifact checks pass.
- [ ] `go vet ./...`, `go build ./cmd/pm`, individual `make verify` static gates, and detached `make connector-boundary` pass.
- [ ] `go test -timeout 20m ./...` and `make verify` are attempted with a non-cutoff runner or truthfully left to CI.
- [ ] Inline code review is recorded in `REVIEW.md`; automated-review route/dispositions are recorded in PR #3552.

## Live certification hold

No approved disposable Gong credential reference was provided or discovered. The live stage is
therefore intentionally not run. Required eventual evidence is: persisted App-path credential
use; reads, writes, application commands, pagination, required-input behavior, ETL,
plan/preview/approval/apply/readback reverse ETL, binary routes if declared, representative CRUD
with cleanup, and bounded non-secret request/result fingerprints. No browser session may replace
connector authentication.

# Plan: Issue #3866 shared transport family conformance

See [CONTEXT.md](CONTEXT.md) for the completed Task Delivery Header, evidence table, decisions, and required skills.

## Goal

Add one deterministic shared-transport conformance matrix that proves the four family half-paths and their declared/durable coordination behavior without claiming connector-route or live certification coverage.

## TDD plan

1. **Red — family matrix contract.** Add a named table-driven conformance test at the shared transport seam using existing #3810-style semantic records and only fakes. Initially reference a missing matrix fixture/helper so the test fails to compile or reports the missing behavior. Record the exact focused command and red output in `TDD-LEDGER.md`.
2. **Green — happy family paths and modes.** Build the smallest test-only fixtures/helper that invokes the closed dispatch, asserts exact produced records/receipts/acknowledgements/checkpoints, and accounts for each canonical sync mode with a named reason for each excluded half-path.
3. **Green — invariant and refusal cases.** Assert verified-auth fencing, durable rate parking/resume without replay, acknowledgement-before-checkpoint, cancellation, CAS conflict, and restart from durable checkpoint. Every bad case must inspect the concrete typed error and prove source/destination counters remain zero before I/O.
4. **Red/Green — sensitivity demonstration.** Make one schema-valid fixture binding wrong after schema compilation; run only its named test and record the failure. Restore the binding and record the passing named test.
5. **Refactor/review.** Keep fakes private to tests and narrow to the shared seam. Run the required verification and an inline deep diff review; resolve every actionable finding before opening the PR.

## Scope guards

- Test and planning files only unless the existing test seam has a demonstrable missing test-only adapter. No production source/destination registration or executor may be added.
- Do not touch `internal/connectors/certify/**`, connector definitions, certification profiles/matrices/capability flags, PostgreSQL adapter/profile, GitHub direct-read coverage, CLI help/dispatch, or live change capture.
- CLI help/docs/website parity is not applicable: the CLI surface and generated documentation are unchanged. Generated-doc stability is still a required repository gate.

## Verification checklist

- [ ] Focused matrix commands prove each named happy/bad/edge case.
- [ ] Scratch fault proves the named case fails after schema compilation and restoration makes it pass.
- [ ] `go test -timeout 20m ./internal/synctransport ./internal/synccontract ./internal/coordination`
- [ ] `go test -timeout 20m ./internal/app`
- [ ] `go test -timeout 20m ./internal/cli`
- [ ] `go test -timeout 20m ./cmd/connectorgen`
- [ ] `go vet` over changed packages and their direct application consumer.
- [ ] `go build ./cmd/pm`
- [ ] Repository non-test verification gates, `pnpm --dir website run gen:docs` twice, `connectorgen boundary`, and a clean generated diff.
- [ ] Inline/manual `verify-work` evidence and code-review are recorded in `VERIFICATION.md` and `REVIEW.md`.

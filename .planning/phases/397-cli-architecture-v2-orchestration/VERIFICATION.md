# Issue #397 Parent Verification Checklist

Status: not yet run at final HEAD
`verificationPassed`: false

## Existing child reconciliation

- [x] Parent checkout and remote parent head matched at `56a7ecb08f755184af7b55318c3285582d5adfb7`.
- [x] PR #460 local/remote/PR head continuity confirmed at `8d696cd4c27fad6840e905917e7658e785fa5436`; remote checks green.
- [x] PR #461 local/remote/PR head continuity confirmed at `c6138292cfcc7205f7968a54b57a65f933a3c1fa`; final verify check was still pending at initial inspection.
- [ ] #424 independent review finding corrected and exact-head re-review clean.
- [ ] #415 independent review findings corrected and exact-head re-review clean.

## Per-unit gate

For every remaining unit:

- [ ] Plan, TDD ledger, verification, summary, and run-state updated before production edits.
- [ ] Sol/high worker session, starting HEAD, ending HEAD, branch, and worktree recorded.
- [ ] Focused RED captured before production behavior edit.
- [ ] Focused GREEN and issue safety/parity checks pass.
- [ ] Coherent green commit created.
- [ ] Independent Sol/xhigh exact-head review is clean.
- [ ] Reviewed commit promoted with head continuity confirmed.

## Final exact-head campaign

- [ ] `gofmt -w cmd internal`
- [ ] `git diff --exit-code -- cmd internal`
- [ ] `git diff --check`
- [ ] `go vet ./...`
- [ ] `go test ./...`
- [ ] `go test -race ./...`
- [ ] `go build ./cmd/pm`
- [ ] `make verify`
- [ ] `make verify-duckdb` when CGO is available, otherwise explicit not-applicable evidence
- [ ] module/import boundaries and `go mod tidy -diff` / `go mod verify`
- [ ] dependency delta matches accepted ADR 0002-0004 budgets
- [ ] generated docs/manual/goldens/website data are clean
- [ ] runtime help, bare namespaces, command help, invalid-action errors, and JSON/stdout parity
- [ ] security and secret-pattern review without reading credential values
- [ ] repository hygiene (clean tree, no unrelated files, no tracked generated binaries)
- [ ] runtime-backed integration explicitly marked not run unless requested

## Final review

- [ ] Correctness review at exact final HEAD: Sol/xhigh, clean.
- [ ] Security review at exact final HEAD: Sol/xhigh, clean.
- [ ] Architecture/issue-coverage/evidence review at exact final HEAD: Sol/xhigh, clean.
- [ ] Every correction reviewed at its new exact HEAD.
- [ ] PR #438 final CI green at the same pushed HEAD.

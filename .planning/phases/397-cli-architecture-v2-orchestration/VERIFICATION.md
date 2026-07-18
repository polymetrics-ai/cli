# Issue #397 Parent Verification Checklist

Status: not yet run at final HEAD
`verificationPassed`: false

## Existing child reconciliation

- [x] Parent checkout and remote parent head matched at `56a7ecb08f755184af7b55318c3285582d5adfb7`.
- [x] PR #460 local/remote/PR head continuity confirmed at `8d696cd4c27fad6840e905917e7658e785fa5436`; remote checks green.
- [x] PR #461 local/remote/PR head continuity confirmed at `c6138292cfcc7205f7968a54b57a65f933a3c1fa`; final verify check was still pending at initial inspection.
- [x] #424 corrected at `323d4a91b465cdee5fdb94ea338f4272b76de781`; exact-head Sol/xhigh re-review `05a92a52-3893-4eb9-855e-1a5b001ab64e` clean; CI green; ancestry preserved in parent `1f5bd80f`.
- [x] #415 corrected at `6cf5c48f1b2cf218ed35c15ba77096db89969575`; exact-head Sol/xhigh re-review `933b6246-2377-4c5d-8d9d-9e9af2ce159d` clean; CI green; combined conflict-resolution review `4ec8f305-9f7f-40c4-97c2-68c2e01c0d36` clean; ancestry preserved in parent `1f5bd80f`.
- [x] #425 implemented/corrected at `784153c7ed7cbb94360601b84c40c821eec21823`; exact-head Sol/xhigh re-review `905f2b84-2325-451d-bda9-ec4b08983307` clean; promoted in parent `0c57ec39`.
- [x] #426 implemented at `fe2a937b5809ee53518549d6148d41879b6a8c2d`; exact-head Sol/xhigh review `60adcff2-041f-40fa-9ec9-2d6ae6837a3e` clean; promoted in parent `bb12f265`.
- [x] #427 implemented/corrected at `aacb15361b9f42a381442f79b9ca50e56b482205`; exact-head Sol/xhigh re-review `ef326655-bbd1-4c08-92c6-d160ef91b536` clean; promoted in parent `e68ccdf7`.
- [x] #428 implemented/security-corrected at `924ebfe016143504502ffeebcee7002f6d520c6f`; final exact-head Sol/xhigh review `e18a80dd-5de2-4d3d-8458-1f99f0f98397` clean; promoted in parent `569536d1`.

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

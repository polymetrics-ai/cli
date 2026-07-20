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
- [x] #429 implemented/corrected at `9e966e85868aedb0ddfd79ca0de8556ed78345c5`; final exact-head Sol/xhigh review `1b30b51b-73a3-4c2d-90da-bf5161d36a8f` clean; bounded integration race/non-race suites passed after one over-broad timeout; promoted in parent `a490eeba`.
- [x] #430 implemented/corrected at `ad0f23bbe6b9fc71713d651d0b25ff6c42d43a06`; exact-head Sol/xhigh re-review `6049b205-ccea-4d8d-ba25-3046a865c19c` clean; integration race passed; promoted in parent `4a9fa0fb`.
- [x] #431 implemented/corrected at `d628fce2916c390f51c8e7e519d481c2cc9f51fe`; exact-head Sol/xhigh re-review `6ea302fa-ce00-42f0-a7e5-ed4b2282bce5` clean; actual CLI/app/safety integration race gates passed; promoted in parent `573d6222`.
- [x] #432 implemented at `b8377b7b200e50ccb5ec164670fed4f78a5c486a`; exact-head Sol/xhigh review `9dbfa83b-6355-4e9c-9a68-5ae39a7aabe9` clean; integration race passed; promoted in parent `ad6e4331`.
- [x] #433 implemented at `701569ee985f7c87f011d8a1cfab39afcc3cc8c2`; exact-head Sol/xhigh review `1a49a57e-51dc-4abc-ae4b-fda7152a416d` clean; fake-only integration race passed; promoted in parent `990b8f60`.
- [x] #434 implemented/corrected at `8177e342ad03b5fbf3750f2c0ecf9aa11f695f92`; exact-head Sol/xhigh re-review `65f48296-b24a-498a-8b25-6c4a3143d9c9` clean; fake-only integration race passed; promoted in parent `96680756`.
- [x] #435 implemented/evidence-corrected at `f712e696e075792492397ab1d556d1dfceadba04`; exact-head Sol/xhigh re-review `d549e7cf-50bc-4d9d-94ff-04734f048d3b` clean; fake-only integration race passed; promoted in parent `afd765e9`.
- [x] #436 implemented/corrected at `7da245b8da2b8590766d99ca9e967d366e50cfcc`; exact-head Sol/xhigh re-review `19fd2452-9b9b-4479-bbba-2bf9f13a1bbb` clean; temp-only integration race passed; promoted in parent `b28e67fb`.
- [x] #462 / PR #465 CI green and integrated at `a5474bcb9efdbaddcd6d2c83a96a29be03b20bfa`; automated review did not complete because Claude is disabled and Copilot quota is exhausted.
- [x] #462 first accepted correction batch integrated through PR #467 at parent `93a117100c6421955262aa32794a91a158d267e1`.
- [x] #462 follow-up PR #468 opened at current head `5092e115d4aa35ab4595a9b9537f64d3f63e6406`; local docs checks and exact-file closure review pass.
- [x] #462 GitHub blocked-by metadata directly updated for #408/#409/#411/#412/#414/#416/#418/#463.
- [x] #462 PR #468 current-head CI green, including `verify` in 18m19s.
- [ ] #462 human review coverage; local `pm-reviewer` is not substitute coverage.
- [x] #437 / PR #466 phase planning/TDD/verification refreshed before every correction; current PR head `26f98a72419010b961b5b8378ef4a695b0c0a06f`.
- [x] #437 local focused, repeated, race, full package, `go test ./...`, vet, build, `make verify`, help/docs/website parity, fixture-only sample certify, and connectorgen 547/0 gates passed.
- [x] #437 remote CI timing failure reproduced as exact evidence and corrected with deterministic concurrency proof; all current GitHub checks pass.
- [x] #437 local exact-path code re-review has no remaining actionable runtime finding.
- [ ] #437 external automated or human/parent fallback review coverage.
- [ ] #437 promotion into parent branch.

## Current live blockers

- [x] #462 accepted design/safety findings corrected in isolated stacked PRs #467/#468.
- [ ] #462 PR #468 human review coverage recorded; local `pm-reviewer` is not substitute coverage.
- [ ] #408 TUI worker launch unblocked only after #462 correction and review gate.
- [ ] #419 explicit optional dependency decision.

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

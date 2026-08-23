# Refs #4303 — Canonical repair R2 verification checklist

## Lifecycle evidence

- [x] Review authority read in full; final verdict `BLOCKED`; reviewed branch/SHA match `846ff9e5d9f56cef3f9835d35b57b1b4d468b379`.
- [x] Fresh source-isolated local attachment verified at the reviewed SHA.
- [x] GSD doctor/sources/prompts and `agentcontractgen check` passed.
- [x] No roadmap phase `4303`; documented inline/manual lifecycle fallback created.
- [ ] `discuss-phase --auto` decisions, TDD execution, verify-work coverage, and deep code review complete.

## Behavioral proof

- [x] AB-B01 through AB-B09 red/green evidence is linked in `TDD-LEDGER.md` (B01–B02 commit `3a72b6ab4c0f3811b31b32da1b15423a68780e36`; B03–B09 pending this atomic delivery commit).
- [x] AB-W01 through AB-W05 red/green evidence is linked in `TDD-LEDGER.md` (W01–W04 commit `4e8b2506d`; W05 pending this atomic delivery commit).
- [x] Provider tests are synthetic and assert provider-call, checkpoint, receipt, output, and key behavior; no credentialed call occurred.

## Local gates

- [x] `gofmt -w` changed Go files and `git diff --check` pass.
- [x] Focused `go test -timeout 20m` App/engine/connectors/synccontract/CLI cohorts pass.
- [x] Relevant focused `go test -race -timeout 20m` App receipt/output and engine idempotency cohorts pass.
- [ ] `go vet` passed for changed packages; `go build -o ./pm ./cmd/pm` passed.
- [x] Generated CLI docs and `make docs-check-no-build` pass; remaining generator/surface/contract/boundary/release gates are final-pipeline work.
- [ ] Deep code review has no unresolved actionable finding.
- [ ] Complete no-mistakes pipeline has a successful terminal/checks-passed result without opening a PR.

## Delivery

- [x] W01–W04 declaration/schema group committed and non-force pushed (`4e8b2506d`).
- [x] B01–B02 approval/action-owned read-back group committed and non-force pushed (`3a72b6ab4c0f3811b31b32da1b15423a68780e36`).
- [ ] B03–B09 and W05 receipt/output/idempotency/operator-parity group is focused-green and awaiting its atomic commit/push.
- [ ] Branch and remote SHA match after final push.
- [ ] Worktree is clean.
- [ ] External fix report crosswalks all AB IDs to code, commits, and proof.

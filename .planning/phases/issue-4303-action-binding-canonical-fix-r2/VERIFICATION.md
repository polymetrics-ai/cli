# Refs #4303 — Canonical repair R2 verification checklist

## Lifecycle evidence

- [x] Review authority read in full; final verdict `BLOCKED`; reviewed branch/SHA match `846ff9e5d9f56cef3f9835d35b57b1b4d468b379`.
- [x] Fresh source-isolated local attachment verified at the reviewed SHA.
- [x] GSD doctor/sources/prompts and `agentcontractgen check` passed.
- [x] No roadmap phase `4303`; documented inline/manual lifecycle fallback created.
- [ ] `discuss-phase --auto` decisions, TDD execution, verify-work coverage, and deep code review complete.

## Behavioral proof

- [ ] AB-B01 through AB-B09 resolved with red/green test output linked in `TDD-LEDGER.md`.
- [ ] AB-W01 through AB-W05 resolved with red/green test output linked in `TDD-LEDGER.md`.
- [ ] Provider tests are synthetic and assert provider-call, checkpoint, receipt, output, and key behavior; no credentialed call occurred.

## Local gates

- [ ] `gofmt -w` changed Go files and `git diff --check` pass.
- [ ] Focused `go test -timeout 20m` App/engine/connectors/synctransport/CLI cohorts pass.
- [ ] Relevant focused `go test -race -timeout 20m` cohorts pass.
- [ ] `go vet` and `go build ./cmd/pm` pass.
- [ ] Generator/docs/surface/contract/boundary/release gates pass.
- [ ] Deep code review has no unresolved actionable finding.
- [ ] Complete no-mistakes pipeline has a successful terminal/checks-passed result without opening a PR.

## Delivery

- [ ] Every green dependency group is committed and non-force pushed.
- [ ] Branch and remote SHA match after final push.
- [ ] Worktree is clean.
- [ ] External fix report crosswalks all AB IDs to code, commits, and proof.

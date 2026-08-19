# VERIFICATION — connector-guardrail-remediation-r1

## Required full parent gates

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

## Focused guardrail gates

```bash
go test ./internal/connectors/boundary ./cmd/connectorgen
go run ./cmd/connectorgen ownership . --base origin/main --scope-file <fixture>
go run ./cmd/connectorgen boundary . --json
make connector-boundary
```

## GitHub/read-back gates

```bash
gh-axi pr checks <parent-pr>
gh-axi api /repos/polymetrics-ai/cli/rulesets
gh-axi api /repos/polymetrics-ai/cli/branches/main/protection
```

If GitHub permissions deny required ruleset/branch-protection configuration, record the exact `gh-axi` failure and stop claiming the required remote merge boundary is satisfied.

## no-mistakes gates

- Run `no-mistakes doctor` before validation; do not restart daemon on error.
- After the integrated implementation is committed and pushed, Firstmate will instruct this orchestrator to drive `/no-mistakes` on the existing parent PR.
- Ask-user findings are escalated to Firstmate; respond through `no-mistakes axi respond`, never `--yes`.

## Current evidence

- `scripts/gsd doctor`: pass
- `scripts/gsd list`: pass
- `no-mistakes doctor`: pass; daemon running
- Parent issue: #3579 created
- Parent branch: `fix/3579-connector-path-ownership-guardrails` created from `origin/main`
- #3583 / #3588 no-mistakes run `01KZ0SEAKBB9TG7N3SMG97XKJS`: passed at head `0c321595d7ae4852550a5012a895c3e11f7e8298` with review/test/document/lint/push/pr/ci complete.
- #3588 checks at merge time: required/current checks green per captain authority, including `connector-boundary`; sub-PR open, cleanly mergeable into parent before merge.
- #3588 parent integration: squash-merged into `fix/3579-connector-path-ownership-guardrails` as `86b91fc40f46b8653538531fc40c183913676f05`; parent PR #3580 remains draft and human-gated.
- Parent roadmap restore/update: this ledger commit restores parent planning/state artifacts after the no-mistakes-generated documentation prune and records #3583 provisional integration plus #3590 as next critical-path child.
- #3595 icon registry foundation: issue #3595 attached to parent #3579; draft PR #3596 opened against parent branch with planning scaffold `b814e85a6`; implementation and comprehensive native-Codex `gpt-5.6-sol` validation pending.
- PR #3590 R5/R6 gate remains unanswered. R5's second-registry approach is rejected/superseded by #3595; R6 waits until the foundation lands and #3590 is reconciled.

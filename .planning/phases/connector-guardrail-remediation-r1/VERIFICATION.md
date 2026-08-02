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

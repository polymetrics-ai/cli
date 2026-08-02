# VERIFICATION — issue 3595 icon registry single-source foundation

## Required focused gates

```bash
gofmt -w cmd internal
go test ./internal/connectors ./internal/connectors/boundary ./cmd/iconregistrygen ./cmd/connectorgen
go test ./cmd/pm ./internal/cli
node --check website/scripts/gen-connector-bundles.mjs
node --check website/scripts/fetch-simple-icons.mjs
```

Add or adjust package-specific commands if implementation places tests elsewhere. Use existing package managers/tools only; do not add dependencies without approval.

## Repository gates before integration

```bash
go vet ./...
go test ./...
go build ./cmd/pm
make connector-boundary
make verify
```

If a gate is not applicable or blocked by environment, record the exact reason and do not claim it passed.

## GitHub / no-mistakes gates

- PR targets `fix/3579-connector-path-ownership-guardrails` and uses `Refs #3595` and `Refs #3579`.
- Required/current GitHub checks green before parent integration.
- Comprehensive native-Codex `gpt-5.6-sol` no-mistakes validation at `xhigh`, including full-diff comprehensive review/rereview of all material substantiated issues.
- Do not integrate PR #3590 from the prior no-mistakes run; #3590 needs fresh 5.6 SOL validation after this foundation lands and is reconciled.

## Current evidence

- `scripts/gsd doctor`: pass in `/Users/karthiksivadas/.treehouse/cli-83d592/5/worker-3595-icon-registry`.
- `scripts/gsd prompt programming-loop init --phase connector-guardrail-remediation-r1/workers/issue-3595 --dry-run`: failed with `unknown GSD command: programming-loop`; manual GSD fallback uses `.pi/prompts/pm-gsd-loop.md` and must be recorded in PR evidence.
- Pre-edit audit/proof commands completed without credentialed checks; production red tests pending next step.

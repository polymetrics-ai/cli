# Verification: Intercom CLI Parity Parent Orchestration

## Adapter Preflight

```bash
scripts/gsd doctor
scripts/gsd verify-pi
scripts/gsd list --json
```

Result: passed on 2026-07-09; 69 GSD commands registered. `programming-loop` is not one of the registered commands, so manual GSD fallback is recorded in `PLAN.md`.

## Parent PR / Branch Checks

```bash
gh pr list --head feat/164-intercom-cli-parity --base main --json number,title,state,isDraft,url,headRefName,baseRefName,mergeable,reviewDecision,statusCheckRollup
```

Result: `[]`; parent PR not open yet.

## Parent Final Gates

Run before final handoff:

```bash
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Runtime / Credentials

- No credentialed connector checks are planned or allowed without explicit human request.
- Runtime-backed checks are not required for this connector metadata phase.
- Reverse ETL execution is forbidden outside plan → preview → approval → execute.

## CLI Help / Docs / Website Parity

Parent applies. Each CLI-visible slice must record runtime help, bare namespace behavior, command help, docs/website/generated artifact status, and tests or explicit exemptions.

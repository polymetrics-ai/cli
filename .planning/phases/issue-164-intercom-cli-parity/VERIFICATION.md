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

Result at planning start: `[]`; parent PR was not open. Draft parent PR #220 was opened after the plan seed commit: https://github.com/polymetrics-ai/cli/pull/220.

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

2026-07-10 combined #166-#171 local verification:

- `gofmt -w cmd internal` passed.
- `go vet ./...` passed.
- `go test ./... -timeout=20m` passed.
- `go build ./cmd/pm` passed.
- `make verify` passed, including `go run ./cmd/connectorgen validate internal/connectors/defs` with 547 connectors and 0 findings.

## Runtime / Credentials

- No credentialed connector checks are planned or allowed without explicit human request.
- Runtime-backed checks are not required for this connector metadata phase.
- Reverse ETL execution is forbidden outside plan → preview → approval → execute.

## CLI Help / Docs / Website Parity

Parent applies. Each CLI-visible slice must record runtime help, bare namespace behavior, command help, docs/website/generated artifact status, and tests or explicit exemptions.

2026-07-10 combined #166-#171 parity checks passed:

- `./pm help intercom`
- `./pm intercom`
- `./pm intercom contact list --help`
- `./pm intercom contact view --help`
- `./pm intercom contact create --help`
- temp-root `./pm intercom contact create --credential intercom-local --email test@example.com --preview --json` returned a write plan without a live Intercom call.
- Added `docs/cli/intercom.md` and `website/content/docs/intercom-cli-surface.mdx`.

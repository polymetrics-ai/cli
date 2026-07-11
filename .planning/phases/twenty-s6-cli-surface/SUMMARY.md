# Twenty S6 CLI surface + help/manual/website parity summary (#283)

Status: REVIEW_FIX_F1_ACCEPTED_METADATA_APPROVAL_CORRECTED; local re-verification pending.

- Branch: `feat/283-twenty-cli-surface` from `62b8b46cde85ae620591194467d3474fda2a1e8a`.
- Manual GSD fallback: `scripts/gsd prompt programming-loop ...` unavailable (`unknown GSD command: programming-loop`).
- Red evidence captured before production edits: `internal/connectors/defs/twenty/cli_surface.json` absent; `go run ./cmd/pm help twenty`, `go run ./cmd/pm twenty`, and `go run ./cmd/pm twenty --help` failed; `go run ./cmd/pm connectors` already rendered bare help.
- Implemented: 168-command Twenty `cli_surface.json`; dynamic connector help/manual fallback for `pm help twenty`, `pm twenty`, and `pm twenty --help`; CLI tests; embedded surface count test; docs/manual/skill and website generated data.
- Review fix F1: corrected Twenty metadata/manual/skill/catalog approval wording so every reverse ETL write, including creates, requires plan/preview/approval/execute; delete actions additionally require typed `--confirm destructive`.
- Gates passed: `jq`, Python count gate, `connectorgen validate`, Twenty conformance, focused packages, `go vet ./...`, `go build ./cmd/pm`, `gofmt -l cmd internal`, `./pm` help commands, docs grep, docs validate, website generation/idempotence, `go test -timeout 20m ./...`.
- Safety: no credentials, no live connector checks, no reverse ETL execution, no destructive actions, no new dependencies, no generic/raw JSON write flags.
- `make verify` skipped by safety because Makefile `smoke` executes `pm reverse run`.
- GSD evidence gate: `scripts/verify-gsd-workflow 62b8b46c` passed after commit.
- Next: run focused re-verification, amend/push, and re-review PR #320.

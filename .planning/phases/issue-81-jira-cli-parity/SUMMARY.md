# Summary: Jira CLI Parity Parent

Status: draft parent PR #129 opened; #104-#110 verified locally with full gates.

## Completed

- Read required issue-first, parent-orchestrator, stacked PR, GSD, CodeRabbit, automated-review, connector architecture, CLI parity, and skill-routing references.
- Ran GSD adapter health checks.
- Confirmed parent PR for `feat/81-jira-cli-parity` was missing at kickoff.
- Confirmed Jira baseline through metadata-only `pm connectors inspect jira --json` after reading connector help.
- Created parent planning artifacts and orchestration state.
- Committed and pushed parent seed commit `982fa4c1`.
- Opened draft parent PR #129: https://github.com/polymetrics-ai/cli/pull/129.
- Added #104 red test and Jira CLI-surface metadata; targeted engine/connectorgen/conformance checks passed.
- Ran full #104 local gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and `go run ./cmd/connectorgen validate internal/connectors/defs`.
- Added #105 red tests and generic connector command-surface help rendering for `pm jira --help`, `pm jira help`, bare `pm jira`, `pm help jira`, and JSON help.
- Regenerated Jira connector manual/skill plus website connector data for Jira `cliSurface`; website connector-data tests and `pnpm build` passed.
- Ran full #105 local gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`, and `cd website && pnpm build`.
- Routed automated review for #105: CodeRabbit reported a 51 minute review-limit window, Copilot backup review produced four comments across two passes, and CodeRabbit later completed manual review after the retry window with one trivial nitpick. All comments were fixed and re-verified with full gates.
- Added #106 Jira stream-runner flags and tests for `pm jira issue list --jql`, `pm jira project list --query`, and `pm jira user list --query` against local `httptest` Jira endpoints.
- Ran full #106 local gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`, and `cd website && pnpm build`.
- Routed CodeRabbit review for #106; fixed both findings (Jira issue-create risk text and handler-goroutine test assertions) and re-ran full gates.
- Completed #107-#110 full-surface pass from the official Jira Cloud OpenAPI: 620 operations inventoried, 333 write actions modeled, 268 generated direct reads added, 3 existing streams retained, and 16 binary/file-upload/rest-query executor gaps explicitly blocked with operation-ledger evidence.
- Added generic `json_redacted` direct-read output policy and `rest_query` operation kind for REST-only body-variable read-query applicability (#109).
- Added bundle-registry load caching after expanded metadata exposed repeated CLI registry parse cost in `internal/connectors/certify`; full `go test ./...` passed after the optimization.
- Ran full #107-#110 local gates: `gofmt -w cmd internal`, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, `go run ./cmd/connectorgen validate internal/connectors/defs`, and `cd website && pnpm build`.

## Current Blockers

- `scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`); manual GSD fallback is active and recorded.
- No Pi `subagent` tool is exposed in this harness; mutating workers are not spawned. #104-#110 ran locally as `local_critical_path`.

## Next

1. Commit/push the #107-#110 full-surface slice and update parent PR #129.
2. Route automated review coverage for the new full-surface commits, using CodeRabbit primary and Copilot fallback only if CodeRabbit is rate-limited/skipped.
3. Keep parent PR #129 draft/human-gated for merge to `main`.

# TDD Ledger: Issue #132 HubSpot CLI Feature Parity Parent

Date: 2026-07-10

## Parent planning gate

No production code changes before these parent artifacts were created:

- `PLAN.md`
- `TDD-LEDGER.md`
- `VERIFICATION.md`
- `RUN-STATE.json`
- `ORCHESTRATION-STATE.json`

## GSD/TDD mode

- Desired command: `scripts/gsd prompt programming-loop init --phase issue-132-hubspot-cli-parity --dry-run`
- Result: blocked because the pinned command registry does not contain `programming-loop`.
- Active fallback: manual universal programming loop using `scripts/gsd prompt plan-phase issue-132-hubspot-cli-parity --skip-research` and `scripts/gsd prompt execute-phase issue-132-hubspot-cli-parity --dry-run` prompts as the GSD adapter path.

## Red-test plan for first implementation lane (#134)

Before production edits, add failing tests for:

1. `cli_surface.json` accepts safe `binary` intent only when backed by typed operation metadata.
2. Implemented binary commands without typed operations are rejected.
3. HubSpot CLI surface metadata exists and has no `raw_api` or `direct_write` commands.
4. HubSpot API inventory metrics match the official baseline once the ledger is introduced: 3,060 unique operations; method counts GET 1,038, POST 1,314, PUT 169, PATCH 232, DELETE 307.

## Red evidence

Issue #134 added failing tests before production edits:

- Binary `cli_surface.json` intent rejected by schema before adding the enum.
- Implemented binary command without a typed operation produced the expected safety finding after validator work.
- HubSpot bundle metadata test failed before the HubSpot definition scaffold existed.

## Green evidence

Issue #134 passed targeted and broad gates:

- `go test ./cmd/connectorgen -run 'CLISurface|HubSpot' -count=1`
- `go test ./cmd/connectorgen ./internal/connectors/engine`
- `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- `gofmt -w cmd internal && go vet ./... && go test ./... && go build ./cmd/pm && make verify && go run ./cmd/connectorgen validate internal/connectors/defs`

## Refactor evidence

- Added `binary` CLI intent validation with typed operation enforcement.
- Added HubSpot metadata-only bundle scaffold and docs/catalog entries.
- Updated connector counts for the 548th declarative bundle.
- Cached declarative bundle loading inside `bundleregistry.New()` so cold full-suite tests stay under the default package timeout while callers still receive fresh registries.

## Safety/TDD notes

- Do not use credentials.
- Do not run live connector checks.
- Do not expose generic raw HTTP write or direct-write execution.
- Keep writes as reverse ETL plan → preview → approval → execute.
- If a safe engine shape is missing, add a typed test and implementation or record an issue-linked blocker with exact evidence.

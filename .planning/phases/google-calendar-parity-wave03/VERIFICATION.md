# Google Calendar parity resume — verification checklist

## Completed focused gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 550 connectors, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync` and `--check` — 550 connectors, 0 updates/drift.
- [x] `go test ./internal/connectors/conformance/...` — pass after restoring the recovered fixture-auth sentinel.
- [x] `go test ./internal/connectors/commandrunner -run TestEveryImplementedCommandPassesRuntimePreflight` — pass.
- [x] `go test ./internal/connectors/hooks/google-calendar ./internal/connectors/bundleregistry ./internal/connectors/native/nativeset` — pass.
- [x] `go test ./internal/cli/...` — pass, including regenerated golden transcripts.
- [x] `go vet` for changed packages and `go build ./cmd/pm` — pass.
- [x] `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build`, `make connector-boundary`, and `make release-workflow-check` — pass.
- [x] `cd website && pnpm run gen:website-data && pnpm run typecheck` — pass; generated data committed.
- [x] `pm help google-calendar`, bare `pm google-calendar`, and `pm google-calendar freebusy query --help` — pass; the required flags render as required. A no-flag FreeBusy request and a removed mutation command both fail without any credential/provider access.
- [x] `git diff --check` — pass.

## Ledger evidence

- Official Discovery total: 38 operations (11 GET, 15 POST, 4 PATCH, 4 PUT, 4 DELETE).
- Executable now: 12 (11 streams plus `freeBusy.query`).
- Blocked: 26, all with the identical named foundation blocker: `rest_write` has no command-runner dispatch.
- Planned: 0.
- Provider citations: 25/25 declared request-field uses and 38/38 operation-level sources, recorded in `REQUEST-FIELD-RESEARCH.md` and `google-calendar-official-operations.json`.

## Constraints

Run no credentialed provider checks and no write execution. Do not claim full `make verify` or `go test ./...`; the shared parity contract requires focused gates because the full suite exceeds bounded command windows.

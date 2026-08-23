# Verification — issue #4288

## Planned local gates

- `go test -timeout 20m ./cmd/connectorgen`
- `go run ./cmd/connectorgen certification-matrix --all`
- `go run ./cmd/connectorgen certification-matrix --check`
- `go run ./cmd/connectorgen certification-sweep --connector jira --check` (and Asana, Notion)
- `go run ./cmd/connectorgen certification-candidates --connector jira --check` (and Asana, Notion)
- `node --check scripts/certify-connector-live.mjs`
- `node scripts/certify-connector-live.mjs <connector> --definition-check`
- The normal secret-safe live harness, one connector/cell at a time, with an immediate matrix check
  after every accepted record
- `go run ./cmd/connectorgen validate`
- `go run ./cmd/connectorgen surface-sync --check`
- `go build ./cmd/pm`
- `make tidy-check`, `make lint`, `make docs-check`, `make smoke-no-build`,
  `go run ./cmd/agentcontractgen check`, `make connectorgen-validate`,
  `make connectorgen-surface-sync`, and `make release-workflow-check`
- `make connector-boundary` run detached to a log and polled to completion
- `git diff --check`
- `bash scripts/verify-gsd-workflow origin/main` after planning evidence is committed

## Executed

- `scripts/gsd doctor` — passed.
- `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review` — each resolved through the pinned registry.
- `go run ./cmd/agentcontractgen check` — passed.
- Initial red scope checks for Jira, Asana, and Notion — expected refusal; recorded in `TDD-LEDGER.md`.

Remaining commands and their results will be appended as the serial live run proceeds. A command that
cannot run locally will be recorded with its exact name and reason before completion.

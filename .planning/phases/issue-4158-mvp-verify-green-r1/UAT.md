# Verify Work — fixture migration

`verify-work --auto` was executed inline because the task contract forbids
lifecycle-role spawning. The relevant acceptance contract is a production
entry-point test, not an in-process replacement.

| Check | Result |
| --- | --- |
| `TestFreshBinaryDeclarativeGitHubWarehouseFlowRoundTrip` | Pass in 37.53s. Fresh 152,132,226-byte binary: 1 sync record, 1 action record, 2 provider comments, 1 warehouse row, durable checkpoint and receipt. |
| Missing approved job | Pass. Typed `validation/flow_job_reference_refused` / malformed, no stored flow or target event. |
| Revoked approved job | Pass. Typed unapproved job-reference refusal wrapping `AuthorizationRevokedError`, zero target events. |
| Stale approved job | Pass. Typed missing job-reference refusal, zero target events. |
| `go test -timeout 20m ./internal/cli` | Pass in 382.569s. |
| `go test -timeout 20m ./internal/flow ./internal/app` | Pass: flow 0.454s; app 222.533s. |
| `go vet ./...`, `go build ./cmd/pm` | Pass. |
| `make tidy-check lint docs-check-no-build smoke-no-build agent-contract-check connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connector-boundary connector-canon-check release-workflow-check` | Each invoked separately and passed. |
| `./pm help flow`, `./pm flow`, `./pm flow --help` | Each exits successfully; `docs/cli/flow.md` already states the job-backed contract. |

The databaseintegration PostgreSQL test remains separately owned by #4158 and
correctly skipped without its explicit runtime opt-in; no result from it is
claimed as part of this fixture migration.

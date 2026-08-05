# Verification checklist — GitLab provider-inventory parity G0 + G1

## Inventory correctness

- [x] The checked-in ledger is the report's exact 1,745-operation JSON payload.
- [x] Provider artifact URL, version, retrieval date, and SHA-256 appear in `api_surface.json`.
- [x] Each row has exactly one executable, blocked, or justified-excluded disposition.
- [x] Counts match the report: executable 4; blocked pending GitLab declarations 1,618; provider restricted 64; blocked on multipart executor 45; justified excluded 14.
- [x] Each executable G1 command carries a provider `source_url` in `cli_surface.json`; its matching `covered_by` endpoint binds that citation to the stream.

## CLI/help/docs/website parity

- [x] `pm gitlab` exits successfully and renders contextual command help.
- [x] `pm help gitlab` resolves.
- [x] Each of `projects list`, `groups list`, `users list`, and `issues list` is individually reachable and resolves through runtime preflight.
- [x] `docs.md`, generated connector manual/skill, and website catalog state four executable reads and no executable writes.
- [x] No planned/blocked provider operation is presented as executable.

## Local gates

- [x] `go run ./cmd/connectorgen validate internal/connectors/defs/gitlab`
- [x] `go run ./cmd/connectorgen surface-sync --check internal/connectors/defs/gitlab`
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/gitlab' -count=1`
- [x] `go test ./internal/connectors/commandrunner -run 'TestEveryImplementedCommandPassesRuntimePreflight' -count=1`
- [x] `go test ./internal/cli -count=1`
- [x] `go build ./cmd/pm`
- [x] `go run ./cmd/pm docs validate --connectors-dir docs/connectors`
- [x] `make connector-boundary`
- [x] `git diff --check`

## Supplemental scoped gates

- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `make agent-contract-check`
- [x] `make connectorgen-validate`
- [x] `make connectorgen-surface-sync`
- [x] `make release-workflow-check`

## Delivery gates

- [ ] GSD inline fallback, skills, TDD, verification, and automated-review evidence are recorded in the PR body.
- [ ] No-mistakes pipeline completes with a green PR; do not merge.

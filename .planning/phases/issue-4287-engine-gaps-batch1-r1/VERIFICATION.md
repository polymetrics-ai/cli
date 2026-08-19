# Verification: Issue #4287

## Required Checks

- [x] Targeted engine, command-runner, keychain/credential, and CLI tests with `-timeout 20m`.
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs`.
- [x] `go run ./cmd/connectorgen surface-sync --check`.
- [x] `make connector-boundary` via the required detached task-local log and exit-code file.
- [x] `make lint` and `make docs-check`.
- [x] `make verify` after individual gates complete.
- [x] `git diff --check`.
- [x] Synthetic fixture-bundle proof for each declared capability; Docker Hub reclassification is a post-merge sweep-lane integration check and is intentionally out of this worktree's scope.
- [x] Deep code review with all actionable findings resolved or dispositioned.

## CLI Parity Assessment

The task changes generated connector command contracts and output safety, but introduces no hand-authored public command, flag, help topic, or documentation page. Before PR, verify existing connector help/manual generation and `docs-check`; record any resulting generated/docs delta. The final PR will state that namespace/help commands are not directly changed.

## Sync-transport foundation alignment

Checked `origin/main` at `acb85dc03` (PR #4286) against
`docs/sync-transport-definition.md`. This issue adds no `sync_transport.json`
or source/destination executor, and its typed operation capabilities compose
beside the declaration-owned ETL/reverse-ETL transport contract without a
connector-name branch. The former sync-transport foundation gap is closed on
main and is not tracked in this phase.

## Results

- PASS — `go test -timeout 20m ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/connectors/connsdk ./cmd/connectorgen` after the final rebase. Synthetic fixtures prove operation-owned page/page-size pagination and source disagreement refusal; HEAD success plus non-HEAD/over-cap refusal; closed SCIM JSON plus unknown-type refusal; bounded CSV success, no-cap refusal, and no artifact on cap overflow; and encrypted-store secret persistence with complete response/error retention.
- PASS — `go vet ./...`, `go build ./cmd/pm`, `go run ./cmd/connectorgen validate internal/connectors/defs`, and `go run ./cmd/connectorgen surface-sync --check` after the final rebase.
- PASS — `make lint && make docs-check && make tidy-check smoke-no-build agent-contract-check release-workflow-check`.
- PASS — detached tracked `make connector-boundary` wrote task-local stdout/stderr and an exit-code file under `.task-tmp/issue-4287/`; exit code was `0`.
- PASS — `make verify` after the final rebase, including `go test -timeout 20m ./...`, generated snapshots, docs, smoke, lint, generator, boundary, canon, and release parity gates. `internal/cli` completed in 883.051 seconds under the repository-required timeout.
- PASS — final source review fixed the generator/runtime mirror omissions for `status_check`, `text_export`, `secret_stored`, and `application/scim+json`; no connector-name branch or open review finding remains.

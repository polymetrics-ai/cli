# Verification: Issue #4287

## Required Checks

- [ ] Targeted engine, command-runner, keychain/credential, and CLI tests with `-timeout 20m`.
- [ ] `go vet ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs`.
- [ ] `go run ./cmd/connectorgen surface-sync --check`.
- [ ] `make connector-boundary` via the required detached task-local log and exit-code file.
- [ ] `make lint` and `make docs-check`.
- [ ] `make verify` after individual gates complete.
- [ ] `git diff --check`.
- [ ] Synthetic fixture-bundle proof for each declared capability; Docker Hub reclassification is a post-merge sweep-lane integration check and is intentionally out of this worktree's scope.
- [ ] Deep code review with all actionable findings resolved or dispositioned.

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

Pending implementation.

# Verification checklist — issue #4299 definition embed slim

Status: planning in progress.

## Required behavior

- [ ] Real `defs.FS` inventory is sorted, attributed, and rejects `api_surface.json`, `fixtures/**`, and every source lock except the explicit GitHub exception.
- [ ] GitHub source-lock literal bytes and SHA-256 match the committed raw file.
- [ ] `go test -timeout 20m ./internal/connectors/certify` proves GitHub GraphQL schema certification remains available offline.
- [ ] Installed/archive proof runs from a directory without the source checkout and retains the GitHub full-certification boundary.
- [ ] Current rebased release-like before/after measurements report identical commands, build metadata, byte sizes, archive sizes, and deltas.

## Repository gates

- [ ] `go test -timeout 20m` for changed packages and `internal/cli`.
- [ ] `go vet ./...`.
- [ ] `go build ./cmd/pm`.
- [ ] `go run ./cmd/connectorgen validate`.
- [ ] `go run ./cmd/connectorgen surface-sync --check`.
- [ ] certification matrix/candidates/sweep checks.
- [ ] `make connector-boundary`, `make lint`, `make docs-check-no-build`, `make tidy-check`, `make agent-contract-check`, and `make release-workflow-check`.
- [ ] `go test -timeout 20m ./...` and `make verify` are run with the repository's long timeout or recorded as CI-owned only if the execution environment terminates them.
- [ ] `scripts/verify-gsd-workflow origin/main`.

## Delivery

- [ ] Commit messages and PR body include `Refs #4299`.
- [ ] Rebase immediately before push; never force-push or merge.
- [ ] Open/update a Conventional Commit PR to `main` with measurement, inventory, TDD/GSD/skill, exception, and exclusion evidence.
- [ ] Read the PR base back through the GitHub API and record `main`.

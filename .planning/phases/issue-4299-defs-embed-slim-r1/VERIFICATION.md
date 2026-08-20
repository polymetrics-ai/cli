# Verification checklist — issue #4299 definition embed slim

Status: local verification complete; PR review pending automatic Claude coverage.

## Required behavior

- [x] Real `defs.FS` inventory is sorted, attributed, and rejects `api_surface.json`, `fixtures/**`, and every source lock except the explicit GitHub exception.
- [x] GitHub source-lock literal bytes and SHA-256 match the committed raw file.
- [x] `go test -timeout 20m ./internal/connectors/certify` proves GitHub GraphQL schema certification remains available offline.
- [x] Installed/archive proof runs from a directory without the source checkout and retains the GitHub full-certification boundary.
- [x] Current rebased release-like before/after measurements report identical commands, build metadata, byte sizes, archive sizes, and deltas in [MEASUREMENTS.md](MEASUREMENTS.md).

Focused proof:

```text
go test -timeout 20m ./internal/connectors/defs ./internal/connectors/certify -count=1  PASS
scripts/tests/release-size-budget.sh                                                   PASS
```

`TestGithubFullGraphQLInventoryUsesEmbeddedSourceLockOutsideCheckout` changes
to a temporary non-checkout directory, then compiles GitHub's full GraphQL
inventory from `defs.FS` (29 schema-conformant, 2 live-required, 274
fixture-bound commands). It is deliberately offline: full live certification
requires maintainer credentials and is not a substitute for the source-lock
runtime boundary proved here.

## Repository gates

- [x] `go test -timeout 20m ./internal/connectors/defs ./internal/connectors/certify ./internal/cli` (the complete suite later covered these again).
- [x] `go vet ./...`.
- [x] `go build ./cmd/pm`.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — 552 connectors, 0 findings.
- [x] `go run ./cmd/connectorgen surface-sync --check` — 552 connectors, 0 changes.
- [x] `go run ./cmd/connectorgen certification-matrix --check`, `certification-candidates --connector github --check`, and `certification-sweep --connector github --check`.
- [x] `make connector-boundary` — detached and polled; clean, 552 connectors, 0 findings.
- [x] `make lint`, `make docs-check`, `make smoke-no-build`, `make tidy-check`, `make agent-contract-check`, and `make release-workflow-check`.
- [x] `make connector-runtime-preflight`, `make connector-canon-check`, and `make github-parity-artifacts-check`.
- [x] `go test -timeout 20m ./...` — exit 0.
- [x] `scripts/verify-gsd-workflow origin/main`.

The repository's `make verify` target is the composition of the checked
commands above. Its long test and boundary members were run with tracked
20-minute sessions and the remaining targets were run independently, as the
repository agent guidance requires under per-command time limits.

## Manual code review

Reviewed commit `a2a3e4a51` against `origin/main`: the exception is a literal
path (not a new glob), the guard validates both compressed archive and streamed
installed-binary bytes, `--print-subjects` stays machine-readable via `--quiet`,
and no connector definition, source lock, compression, or checkout fallback is
in the diff. No local finding.

Automated review route: `claude_auto` after the non-draft main-targeted PR is
opened by this trusted repository worker. No manual Claude or Copilot request
is appropriate before the automatic trigger is observed.

## Delivery

- [x] Commit messages include `Refs #4299`.
- [x] PR body includes `Refs #4299`.
- [x] Rebased against `origin/main` immediately before the branch push; no force-push or merge.
- [x] Opened [PR #4309](https://github.com/polymetrics-ai/cli/pull/4309), Conventional Commit title, targeting `main`, with measurement, inventory, TDD/GSD/skill, exception, and exclusion evidence.
- [x] Verified the GitHub API-reported base: `gh-axi pr list --state open --head fm/cli-defs-embed-slim-r1 --base main` returned exactly PR #4309. Head at PR open: `d5721f66fe6b67dd64624c076379042f255dbc69`.

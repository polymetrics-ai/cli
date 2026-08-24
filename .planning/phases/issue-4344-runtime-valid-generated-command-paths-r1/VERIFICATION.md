# Verification checklist — #4344

- [x] `GOFLAGS='-p=3' go test -count=1 -timeout 20m ./cmd/connectorgen` — PASS (final rerun: 248.518s).
- [x] Targeted behavioral test — PASS:
  `GOFLAGS='-p=3' go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestSourceProjectionGeneratedParameterizedCommandIsRuntimeValidAndStable$'`.
- [x] Deliberate raw-source-ID reintroduction — FAIL as required with
  `generated command path retained a raw source parameter`; restored encoder,
  then re-ran targeted test successfully.
- [x] `GOFLAGS='-p=3' make connectorgen-validate` — PASS: 552 connectors, 0 findings.
- [x] `GOFLAGS='-p=3' make connectorgen-surface-sync` — PASS: 552 connectors, 0 drift.
- [x] `GOFLAGS='-p=3' make connectorgen-operation-evidence` — PASS: 1,525 rows; fixed-100 passed.
- [x] `GOFLAGS='-p=3' make connector-runtime-preflight`, `make connector-canon-check`, and `make connector-boundary` — PASS; boundary scanned 316 files/552 connectors clean.
- [x] `GOFLAGS='-p=3' go build ./cmd/pm`; isolated credential-free sweep of the three Bitbucket commands checked into this base — all returned `error: missing --credential`.
- [x] `./pm help connectors`, `./pm bitbucket`, `./pm bitbucket repositories create --help`, and `./pm docs validate --connectors-dir docs/connectors` — PASS. No current reachable generated path changes, so no generated manual or website edit applies.
- [x] `gofmt`, `GOFLAGS='-p=3' go vet ./...`, `make tidy-check`, `make docs-check`, `make smoke-no-build`, `make lint`, `make agent-contract-check`, and `make release-workflow-check` — PASS.
- [!] Full `go test -timeout 20m ./...` was attempted by the repository commit hook before production edits and failed at 601.236s in baseline `internal/cli` `TestBahmniDeclaredCommandMatrixIsRecognizedOrExplicitlyBlocked` while logging a local Redis connection refusal. Per constrained-machine guidance, the complete suite was not restarted; all changed-package and repository-owned targeted gates above passed.
- [!] `GOFLAGS='-p=3' go run ./cmd/connectorgen source-import bitbucket --check` cannot run because `internal/connectors/defs/bitbucket/sources` is absent on this base. Therefore the requested reviewed-artifact 50-command/28-path Bitbucket sweep is recorded as an integration dependency, not falsely reported as passed.
- [x] Inline verify-work/code-review evidence — PASS; see `REVIEW.md`.
- [x] PR API base read-back — `GET /repos/polymetrics-ai/cli/pulls/4346`
  confirmed base `main`, head `fm/cli-runtime-valid-generated-command-paths-r1`,
  and open PR https://github.com/polymetrics-ai/cli/pull/4346.

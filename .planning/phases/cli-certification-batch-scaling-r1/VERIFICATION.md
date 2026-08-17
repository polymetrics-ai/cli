# Verification — certification batch scaling

## Local results

| Command or check | Result |
| --- | --- |
| `scripts/gsd doctor`; resolved generated prompts for discuss/plan/execute/verify/review | pass; inline fallback recorded because the contract forbids role spawning |
| `go run ./cmd/agentcontractgen check` | pass |
| Pinned PR #4214 source: `go test -timeout 20m ./cmd/connectorgen` | pass |
| Pinned PR #4214 source: `go test -timeout 20m ./internal/connectors/certify` | pass |
| Pinned PR #4214 source: `go run ./cmd/connectorgen certification-candidates --connector github --check` | pass |
| Live batches | pass: 10, then 100, then two additional fresh 100-operation repeats; safe summaries and taxonomy in `LIVE-RESULTS.md` |
| `go vet ./...` | pass |
| `go test -timeout 20m ./cmd/connectorgen` | pass |
| `go test -timeout 20m ./internal/connectors/certify` | pass |
| `go test -timeout 20m ./internal/cli` | pass |
| `go build ./cmd/pm` | pass |
| `make fmt`; `gofmt -l cmd internal` | pass; no source formatting drift |
| `make tidy-check docs-check smoke-no-build lint agent-contract-check` | pass |
| `make connectorgen-validate connectorgen-surface-sync github-parity-artifacts-check connectorgen-certification-matrix connectorgen-certification-sweep connector-boundary connector-canon-check release-workflow-check` | pass |
| `bash -n run-live-batch.sh`; `bash -n classify-failed-direct-reads.sh`; `jq empty RUN-STATE.json STAGED-LIVE-REPORT.json` | pass |
| `pnpm run gen:docs`; `pnpm run gen:website-data`; `pm skills generate --dir docs/skills --json`; `connectorgen certification-matrix --all`; `connectorgen certification-sweep --connector github` | run twice after rebase; byte-stable (`git diff --exit-code`) |
| `scripts/verify-gsd-workflow` | pending final committed verification artifact |

## Intentionally not run as one command

`make verify` and its full `make test` component were not run as a single command. Repository guidance says this worktree's command window routinely cuts off the 550+ connector full suite and makes a cutoff indistinguishable from a hang. Its non-test gates were run individually above, and the changed package, certification consumer, and `internal/cli` tests passed with the required 20-minute timeout. CI remains responsible for the full suite.

`security/snyk` is not a lane regression: the task reports it fails identically on the base branch. It is excluded as directed.

## Safety verification

- No mutation operation was invoked.
- The credential was passed only as an environment-variable name; no secret is in tracked files or command output.
- Temporary source/project copies were removed before repository-wide contract validation, so no nested agent inventory or credential storage remains.
- No accepted certification evidence was hand-authored or published.


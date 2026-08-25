# Verification — issue 4325 declaration-admission foundation

## Acceptance evidence

- The new `connectorgen declaration-admission [dir] [--json]` check loads only
  `sources/<connector>-declaration-admission.json` sidecars. Its focused tests
  cover one runnable read; deferred reverse-ETL write/delete and binary
  upload/download rows; a retained importer/descriptor gap; missing,
  duplicate, citation-free, malformed, stale, base-path-mismatched,
  lane-changing, destructive-metadata-free, and falsely implemented rows; and
  a complete zero-runnable connector.
- The on-disk fixture has only source URL, exact citation, raw operation ID,
  endpoint, lane, command, and state. It contains neither an artifact nor a
  hash, proving those are not admission inputs.
- Deferred command metadata is projected into the command surface and rejected
  by `commandrunner` with the typed `system/missing_foundation` classification
  before an executor can perform provider I/O.
- GitHub's implemented `label delete` action is the destructive green control:
  its source-cited declaration is admitted and the actual commandrunner
  preflight succeeds. Deferred state is therefore endpoint-specific rather
  than a generic delete/destructive classification.
- No `internal/connectors/defs/<connector>` file changed. Existing
  source-lock, surface-sync, runtime-preflight, certification, and live-proof
  gates remain independent and were exercised below where hermetic.

## Commands and results

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestDeclarationAdmission'` | pass |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestDeclarationAdmissionAdmitsGitHubImplementedDeleteControl$'` | pass |
| `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestPreflightDeferredCommandReturnsNamedFoundationBeforeExecutor$'` | pass |
| `go test -timeout 20m ./internal/connectors/engine -run '^TestCommandSurfaceProjectsDeferredFoundationGap$'` | pass |
| `go test -timeout 20m ./cmd/connectorgen` | pass (153.707s including the GitHub implemented-delete control) |
| Fresh local project, no credential: `pm github label delete --json` | exit 1, `missing --credential`; command dispatches without provider I/O |
| Fresh local project, no credential: `pm stripe accounts delete --json` | exit 2, `unknown command "accounts delete"`; no generic account-delete projection or provider I/O |
| `go test -timeout 20m ./internal/connectors/commandrunner` | pass |
| `go test -timeout 20m ./internal/connectors` | pass |
| `go test -timeout 20m ./internal/connectors/engine` | pass (13.447s) |
| `go vet ./...` | pass |
| `go build ./cmd/pm` | pass |
| `make tidy-check`, `make lint`, `make docs-check-no-build`, `make smoke-no-build` | pass |
| `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-surface-sync`, `make connectorgen-declaration-admission` | pass |
| `make connectorgen-operation-evidence`, `make github-parity-artifacts-check`, `make connector-boundary`, `make connector-canon-check` | pass |
| `make connectorgen-certification-subject`, `make connectorgen-certification-matrix`, `make connectorgen-certification-candidates`, `make connectorgen-certification-sweep` | pass |
| `make release-workflow-check` | pass |
| `go run ./cmd/connectorgen declaration-admission --json` | pass; zero sidecars and zero findings on current definitions |
| `go run ./cmd/connectorgen` | expected usage error; confirms the command is listed in the internal generator help |
| `git diff --check` | pass |

The aggregate `go test -timeout 20m ./...` and serial `make verify` are not
run as one process: the repository's `AGENTS.md` explicitly says a per-command
timeout routinely cuts off the full suite and directs agents to run changed
packages plus Make gates individually, leaving the aggregate suite to CI.

## CLI/documentation parity

This changes the internal `connectorgen` generator command, not the shipped
`pm` command tree. Its usage string and certification/canon documentation are
updated. `docs/cli/**`, website docs, generated `pm` manual/help, bare `pm`
namespace behavior, and shell completion are not applicable; no files in those
surfaces changed. `make docs-check-no-build` passed.

# Issue #4193 verification checklist and binary inventory

## Built-binary discovery baseline

Build command: `go build -o .tmp/pm-4193/pm ./cmd/pm` from base
`ff6a87101`.

Root output enumerated 17 documented core namespaces, 36 dynamic connector
namespaces, 556 connector catalog entries, and 1,571 catalog entries marked as
provider commands. Traversing the loaded command surfaces found 8,900 declared
dynamic command leaves; this was substantially larger than the original sample
and is now asserted programmatically (17,800 `--help`/`-h` variants) instead of
being silently treated as a two-command exception.

| Namespace | Leaf paths read from the compiled binary manual |
| --- | --- |
| init | `init` |
| connectors | `list`, `catalog`, `inspect`, `help`, `certify` |
| credentials | `add`, `link`, `list`, `inspect`, `test`, `remove` |
| connections | `create`, `list` |
| catalog | `refresh`, `show` |
| etl | `check`, `catalog`, `read`, `run`, `status`, `transport github-issue-label plan`, `transport github-issue-label preview`, `transport postgres-managed-target plan`, `transport postgres-managed-target preview`, `transport github-issue-label cleanup plan`, `transport github-issue-label cleanup run` |
| query | `run` |
| reverse | `list`, `plan`, `preview`, `run`, `status` |
| flow | `create`, `plan`, `preview`, `run`, `status`, `list` |
| rlm | `run` |
| schedule | `create`, `list`, `inspect`, `status`, `install`, `remove`, `fire` |
| agent | `plan` |
| runtime | `doctor` |
| perf | `compare`, `sync-modes` |
| docs | `generate`, `validate` |
| skills | `generate` |
| version | `version` |

The pre-fix direct-binary `--help` sweep showed nonzero exits for most
project-bound leaves and, for selected non-project commands, handler execution
instead of a `NAME` manual. The post-fix sweep reports exit 0, a `NAME` line,
and no stderr for every derived path under both `--help` and `-h`, outside and
inside a freshly initialized project.

## Red / green evidence

| Stage | Command | Result |
| --- | --- | --- |
| Red | `go test -timeout 20m ./internal/cli -run 'TestEveryLegacyLeafHelpIsExecutable|TestLeafHelpDoesNotMaskInvalidCommandsOrApprovalCarrierSyntax|TestLeafHelpWithOtherFlagsRendersBeforeRequiredFlags' -count=1` | Failed as expected: five wrapper manuals were missing; unknown help used exit 1; required-flag help failed before rendering. |
| Green | `go test -timeout 20m ./internal/cli -run 'TestEveryLegacyLeafHelpIsExecutable|TestLeafHelpDoesNotMaskInvalidCommandsOrApprovalCarrierSyntax|TestLeafHelpWithOtherFlagsRendersBeforeRequiredFlags|TestCobraRouterShellRejectsUnknownHelpWithUsage' -count=1` | Passed. |
| Binary | `go build -o .tmp/pm-4193/pm ./cmd/pm` plus the 63-path legacy sweep below | Passed: 252/252 help requests exited 0, contained `NAME`, and wrote no stderr; each path ran under `--help` and `-h` in empty and initialized directories. |
| Dynamic leaves | `go test -timeout 20m ./internal/cli -run TestEveryDynamicConnectorLeafHelpRendersWithoutDispatch -count=1 -v` | Passed: 17,800 declaration-derived help variants (8,900 leaves × `--help`/`-h`) rendered a `NAME` manual before dispatch. |
| Dynamic roots | Built `pm` root traversal of every dynamically exposed connector in both project states | Passed: 36 roots × `--help`/`-h` × empty/initialized = 144 successful binary renders. |

Binary sweep method (run from the repository-managed `.tmp/pm-4193` fixture):

```text
for state in outside inside; do
  for help in --help -h; do
  for each of the 63 leaf paths above, including help/man/extract/worker; do
      cd to the state directory; pm <path> <help>
      assert exit == 0 and output contains NAME
    done
  done
done
# binary_help_sweep paths=63 runs=252 failures=0 invalid_leaf_exit=2 version_exit=0
```

The binary root enumeration supplied all dynamic connector names. Their
declaration-owned leaf paths are rendered by the pre-dispatch resolver in
`runMaybeConnectorCommand`; the direct `CommandSurface.Commands` sweep above
checks every generated leaf without starting thousands of duplicate binaries.

## Required gates

- [x] `gofmt -w cmd internal`
- [x] `go test -timeout 20m ./internal/cli -count=1`
- [x] `go test -timeout 20m ./cmd/connectorgen`
- [x] `go vet ./...`
- [x] `go build ./cmd/pm`
- [x] `make tidy-check`
- [x] `make lint`
- [x] `make docs-check`
- [x] `make smoke-no-build`
- [x] `go run ./cmd/agentcontractgen check`
- [x] `go run ./cmd/connectorgen validate`
- [x] `go run ./cmd/connectorgen surface-sync --check`
- [x] `go run ./cmd/connectorgen boundary`
- [x] `scripts/tests/release-target-parity.sh`
- [x] `pnpm --dir website run gen:docs` twice with byte-stability confirmed
- [x] `git diff --check`

The repository rule prohibits a monolithic `go test ./...`/`make verify` under a
per-command timeout because it is routinely cut off. Run changed-package and CLI
tests separately plus each non-test `make verify` gate above; CI owns the full
suite. Record all results below before opening the PR.

## Results

| Command | Result |
| --- | --- |
| `go test -timeout 20m ./internal/cli -count=1` | PASS — `ok polymetrics.ai/internal/cli` in 513.833s. |
| `go test -timeout 20m ./cmd/connectorgen` | PASS — `ok polymetrics.ai/cmd/connectorgen` in 168.031s. This consumer suite is required because the change updates CLI documentation source and generated artifacts. |
| Focused red test | Expected failure before implementation: missing `init`, `help`, `man`, `extract`, and `worker` manuals; unknown help exited 1; required-flag help failed before its manual. |
| Focused green test | PASS — `TestEveryLegacyLeafHelpIsExecutable`, `TestEveryDynamicConnectorLeafHelpRendersWithoutDispatch`, unknown-command, approval-carrier, required-flag/positional, and ETL-transport contextual-manual cases. |
| `go test -timeout 20m ./internal/cli -run TestGoldenTranscripts -count=1` | PASS after regenerating the transcript fixture. |
| `go run ./cmd/pm docs generate --dir docs/cli` and `go test -timeout 20m ./internal/cli -run TestGoldenDocsGenerateMatchesTrackedCLIManuals -count=1` | PASS — five newly tracked CLI manuals are current. |
| `go vet ./...`; `go build ./cmd/pm`; `make tidy-check`; `make lint`; `make docs-check`; `make smoke-no-build` | PASS. |
| `go run ./cmd/agentcontractgen check`; `go run ./cmd/connectorgen validate`; `go run ./cmd/connectorgen surface-sync --check`; `make connector-boundary`; `scripts/tests/release-target-parity.sh` | PASS — 552 connector definitions validate; surface-sync corrected 0 fields; boundary reported clean. |
| `pnpm --dir website run gen:docs` twice and `git diff --quiet -- website` after each run | PASS — 12 pages generated both times with no website diff. |
| `git diff --check` | PASS. |

The repository explicitly prohibits a monolithic `go test ./...` / `make verify`
under the per-command execution limit because those runs are routinely truncated.
The changed package (`internal/cli`) and the documentation/artifact consumer
(`cmd/connectorgen`) were run in full; the non-test `make verify` gates were run
individually above. CI owns the monolithic aggregate suite.

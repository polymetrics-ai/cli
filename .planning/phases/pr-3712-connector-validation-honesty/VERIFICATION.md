# Verification — PR 3712 Connector Validation Honesty

## Required commands

| Command | Result | Notes |
|---|---|---|
| `scripts/gsd doctor` | Pass | Adapter healthy; node v24.13.1, 69 commands registered. |
| `scripts/gsd prompt programming-loop init --phase pr-3712-connector-validation-honesty --dry-run` | Fallback | Registry returned `unknown GSD command: programming-loop`; explicit manual-GSD fallback recorded in `PLAN.md`. |
| `scripts/gsd prompt plan-phase pr-3712-connector-validation-honesty --skip-research` | Pass | 142-line planning prompt generated and applied inline. |
| `gofmt -l cmd internal` | Pass | No output. |
| `go vet ./internal/connectors/conformance/ ./cmd/connectorgen/ ./internal/connectors/commandrunner/` | Pass | Exit 0. |
| `go test ./internal/connectors/conformance/` | Pass | `ok polymetrics.ai/internal/connectors/conformance 10.639s`. |
| `go test ./cmd/connectorgen/ ./internal/connectors/commandrunner/ ./internal/connectors/engine/` | Pass | `ok cmd/connectorgen 6.064s`, `ok commandrunner 4.069s`, `ok engine 11.084s`. |
| `go test -timeout 25m ./internal/cli/` | Pass | `ok polymetrics.ai/internal/cli 387.356s`. Run scoped per `AGENTS.md`; CI's own `Verify` job also reported `ok polymetrics.ai/internal/cli 703.654s` on this branch. |
| `go test ./internal/connectors/...` | Pass | Every connector package green; no non-`ok` lines. |
| `go build ./cmd/pm` | Pass | Exit 0. |

`go test ./...` was deliberately not run as one command: per `AGENTS.md` the full suite spans 550+
connectors and routinely exceeds a per-command timeout, which is indistinguishable from a hang. Local
runs are scoped to the changed packages plus `internal/cli`; CI carries the full suite.

## Root cause fixed

CI's `Verify` job is serial (`verify: fmt tidy-check vet test build docs-check smoke lint …`), so it
stopped at `test` and never reached the later gates.

| Check | Result | Notes |
|---|---|---|
| `Verify` failure isolated | Root cause found | Only `TestConformance/github` failed, on the `surface_complete` static check: `endpoint 7 (GET /repos/{owner}/{repo}/actions/artifacts/{artifact_id}/{archive_format}) covered_by.direct_read "artifact download" is not an implemented direct_read command`. Reproduced locally at `ca0776836`. |
| Cause | Stale second copy of the coverage rule | `checkSurfaceComplete` in `internal/connectors/conformance/static.go` admitted only `intent: direct_read` into its `directReads` coverage set. `cmd/connectorgen/validate.go:576-587` had already been widened to `direct_read \|\| binary_download` in this PR; conformance's copy had not, so `github`'s `artifact download` stopped covering its own endpoint once it became a `binary_download`. |
| Fix | One condition, matching the sibling copy | `directReads` now admits implemented `direct_read` **or** `binary_download` commands, with the same comment `cmd/connectorgen` carries. `github`'s `api_surface.json` was not touched; the checker was stale, not the bundle. |
| `gsd-workflow-evidence` failure | Addressed | `cmd/**` and `internal/**` changed with no `.planning/**` evidence in the diff. This phase directory supplies `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, and `RUN-STATE.json` with an explicit manual-GSD fallback plus `Red:`/`Green:` evidence. |

## Gate-pinning evidence

| Check | Result | Notes |
|---|---|---|
| Pre-fix red, real corpus | Red as expected | `go test ./internal/connectors/conformance/ -run TestConformance/github` fails on `surface_complete` with the exact CI error string. |
| Pre-fix red, focused test | Red as expected | New `TestCheckSurfaceComplete_BinaryDownloadSatisfiesDirectReadCoverage` fails with the same message shape on a 1-endpoint bundle. |
| Availability-guard mutant | Red as expected | Removing `cmd.Availability == "implemented"` from the widened condition fails the test's `planned` half: `a planned binary_download command must not satisfy covered_by.direct_reads`. Mutation reverted immediately; the branch never carried a weakened guard. |

## `make verify` gates run individually

Each gate CI never reached was run on its own, since `verify` aborts at the first failure.

| Gate | Result | Notes |
|---|---|---|
| `make tidy-check` | Pass | `go mod tidy` + `git diff --exit-code` clean. |
| `make build` | Pass | Exit 0. |
| `make docs-check-no-build` | Pass | `Validated connector docs in docs/connectors`. |
| `make smoke-no-build` | Pass | `smoke ok: /var/folders/.../tmp.ysNoZGMfXk`. |
| `make lint` | Pass | `golangci-lint run …`: `0 issues.` |
| `make connectorgen-validate` | Pass | `550 connector(s) checked, 0 findings`. |
| `make connectorgen-surface-sync` | Pass | `550 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)`. |
| `make connector-boundary` | Pass | JSON report emitted, exit 0. |
| `make release-workflow-check` | Pass | `homebrew release notification assertions passed`. |

## Executable evidence (built and run, not read)

Per the phase constraint that `implemented` is proven by execution rather than by inspecting files:

| Command | Result | Notes |
|---|---|---|
| `./pm github artifact download --help` | Pass | Renders `INTENT binary_download`, `AVAILABILITY implemented`, the operation id, both path flags with `maps_to`, and the `DOWNLOAD FLAGS` block (`--dest-root` required, `--file-name`, `--max-bytes`). |
| `./pm github artifact download --artifact-id 1 --archive-format zip --dest-root <tmp>/dl --root <tmp>` | Pass | Fails with `error: missing --credential`, i.e. it reaches credential resolution. It is **not** blocked by `commandrunner.Preflight`, which is exactly the claim `availability: implemented` makes. |
| `./pm connectors --help`, `./pm github` | Pass | Bare namespace renders contextual help and exits 0. |

## CLI/help/docs/website parity evidence

| Check | Result | Notes |
|---|---|---|
| `pm` help, manual, `docs/cli/**`, `website/**` | Not applicable to this CI slice | The only production change here is a coverage set inside `internal/connectors/conformance`, a validation package with no `pm` command, flag, output, help topic, or connector surface. The PR's own parity work (github surface docs, `docs/architecture/connector-operation-kernel.md`, `docs/migration/conventions.md`, `website/**` generated data and counts) landed in commits `51a1de177`, `ce76276e6`, and `ca0776836` and is unchanged by this slice. |

## Open items

- The `cli_surface`/`streams.json` Tier-3 defect is reported, not fixed: out of scope for this PR by
  explicit instruction.
- `cmd/connectorgen/validate.go` and `internal/connectors/conformance/static.go` still carry two
  hand-maintained copies of `checkSurfaceComplete` over separate corpora. This slice restored parity
  between them rather than merging them; unification is a refactor for a separate issue. The standing
  runtime-truth guard remains `TestEveryImplementedCommandPassesRuntimePreflight`, which asserts
  against the real `Preflight` instead of a description of it.

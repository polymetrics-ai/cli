# TDD LEDGER — issue #3718 canonical delivery contract

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
|---|---|---|---|---|
| R1 | Every canonical GSD command resolves | `go test ./internal/agentcontract` failed: `GSD command "programming-loop" does not resolve: scripts/gsd: unknown GSD command or prompt: programming-loop` | `TestReferencedGSDCommandsResolve` loads the canonical list and passed by executing `scripts/gsd sources` for every listed command, including excluded-but-resolved `ship` | Focused package test and `make agent-contract-check` passed |
| R2 | Diverged projection fails | Same RED run failed: `CheckProjection accepted a diverged projection` against the pre-enforcement stub | `TestProjectionDriftCheckAndSync` writes a matching wrapper, mutates `Receive one assigned job` to `Receive many jobs`, observes `CheckProjections` fail, runs bounded sync, and observes the check pass | Focused package test passed; sync preserves harness-owned prefix/suffix |
| R3 | Stable deterministic rendering | N/A (added after R2 green) | `TestRenderIsStableAndConnectorInheritsBase` renders twice and locks the base output SHA-256 | Both base and connector renderings contain every base state; connector rendering contains every overlay state |
| R4 | Required fields, single-worker/no-delegation, overlay inheritance | N/A (source validator coverage) | Table-driven mutation test rejects multiple workers, delegation, broken inheritance, allowed `--yes`, GSD ship, and missing authority pause state | Focused package test passed |
| R5 | CI drift entry point | N/A | `make agent-contract-check` executed successfully and is included in serial and parallel `verify` targets | `go vet ./internal/agentcontract ./cmd/agentcontractgen ./internal/cli` passed |

## Refactor gate evidence

The first broad `make lint` run failed on seven unchecked diagnostic writes/close handling and one
`bytes.Contains` idiom. The first whole-tree connector-boundary run also rejected provider-specific
`GitHub` identifiers in the new shared Go package. Commit `d8bead33c` checked all command-output
writes, handled close explicitly, used the required bytes idiom, and renamed the Go model/rendering
to provider-neutral tracker terminology while preserving GitHub-specific prose in the JSON source.
Executed reruns: `make lint` reported `0 issues`; `make connector-boundary` reported `outcome:
"clean"` with zero findings and warnings.

Tests record executed output, not inferred behavior. Prose decisions are documented; only executable
enforcement claims receive red/green evidence.

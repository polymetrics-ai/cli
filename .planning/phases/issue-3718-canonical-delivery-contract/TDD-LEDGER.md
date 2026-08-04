# TDD LEDGER — issue #3718 canonical delivery contract

| ID | Enforcement | RED evidence | GREEN evidence | Refactor/verification |
|---|---|---|---|---|
| R1 | Every canonical GSD command resolves | `go test ./internal/agentcontract` failed: `GSD command "programming-loop" does not resolve: scripts/gsd: unknown GSD command or prompt: programming-loop` | Pending | Pending |
| R2 | Diverged projection fails | Same executed run failed: `CheckProjection accepted a diverged projection` against the pre-enforcement stub | Pending | Pending |
| R3 | Stable deterministic rendering | N/A (added after R2 green) | Pending | Pending |
| R4 | Required fields, single-worker/no-delegation, overlay inheritance | N/A (source validator coverage) | Pending | Pending |
| R5 | CI drift entry point | N/A | Pending: `make agent-contract-check` | Pending |

Tests record executed output, not inferred behavior. Prose decisions are documented; only executable
enforcement claims receive red/green evidence.

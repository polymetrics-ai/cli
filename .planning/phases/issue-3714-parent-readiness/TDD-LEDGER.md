# TDD ledger — issue #3714 parent readiness

| ID | Contract | RED evidence | GREEN evidence | Refactor / verification |
|---|---|---|---|---|
| R1 | Parent PR head contains current `origin/main` | Pending: `git merge-base --is-ancestor origin/main HEAD` must exit nonzero before integration | Pending: the same command exits zero after merging `origin/main` into the existing parent branch | `git log` records the three mainline commits beneath the parent head |
| R2 | Every required Codex, Claude, and Pi projection matches the canonical source | Pending only if the post-merge drift check detects divergence; no projection will be edited by hand | `go run ./cmd/agentcontractgen sync` then `go run ./cmd/agentcontractgen check` succeeds | Focused generator tests preserve complete three-harness coverage |
| R3 | Mainline destructive-write confirmation behavior survives the integration | Pending conflict/path inspection; no deliberate regression is introduced | Focused tests from the #3730 merge path pass on the integrated head | Broader parent/CI verification remains green |

Only executed results will replace `Pending`; no pass result is inferred from the pre-integration
parent CI run.


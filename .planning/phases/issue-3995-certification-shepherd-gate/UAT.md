# UAT — issue #3995 shared connector-certification Shepherd gate

## Operator-visible checks

1. Run the structural projection check:

   ```sh
   go run ./cmd/agentcontractgen check
   ```

   Expected: the canonical contract and all eight registered Claude/Codex/Pi/OpenCode projections
   are current. This remains green even when no connector is certified.

2. Evaluate GitHub `integrate_sub_pr` through the canonical protected-transition command:

   ```sh
   go run ./cmd/agentcontractgen certification-gate \
     --root "$(pwd -P)" \
     --connector github \
     --transition integrate_sub_pr
   ```

   Expected: exit status `1` and deterministic JSON `RETRY` containing
   `capability/github/capability:check/live_evidence`. A file, a reachable route, or
   `implemented: true` does not substitute for accepted live evidence.

3. Evaluate a complete generated fixture for the same transition.

   Expected: `PROCEED`. Removing one binding criterion returns that criterion's exact failure ID;
   malformed/version-unknown input returns `HALT` with its exact cell/evidence coordinate.

4. Run the same command with each protected transition (`integrate_sub_pr`, `accepted`,
   `ready_parent`, and `human_ready`) while the generated baseline remains non-certified.

   Expected: the gate blocks `integrate_sub_pr`, `accepted`, `ready_parent`, and `human_ready`
   while preserving the original deterministic failures for Shepherd retry/halt handling.

## Safety acceptance

The evaluator reads only declared generated JSON through a root-bound reader; symlink ancestors and
non-regular evidence records halt rather than escaping the supplied root. It creates no evidence,
has no provider action dependency, and does not access a credential or mutate production state. A
future #3989 proof schema revision must be integrated explicitly; unsupported versions halt closed.

The command requires a canonical absolute root with no symlinked component or contract; omitted,
relative, traversing, and symlinked roots halt for every protected transition. `--help` is a blocked
request and emits a JSON `HALT`, so only an encoded `PROCEED` exits zero.

`go run ./cmd/connectorgen certification-matrix --check` confirms the unchanged producer remains
bound to `flow-matrix.json`; `go run ./cmd/agentcontractgen sync` and `check` generate and verify
the consumer catalog from that same matrix. A flow override may change only bound live evidence:
immutable fact promotion, safe-missing/mismatched/wrong-coordinate base records, and distinct large
proof numbers all halt before a transition can proceed.

Each flow pair is also bound to its canonical flow kind’s source and destination roles: changing
applicability, declared/implemented conjunctions, or the exact derived not-applicable code/reason
halts at that pair coordinate before evidence or status derivation. The catalog’s stable accessor
fails closed if generated data is absent, empty, or invalid; running `agentcontractgen sync` restores
an absent data-only `flow_gen.go` from the producer matrix without changing the producer path.

This internal `agentcontractgen` command is not part of the `pm` CLI, so `pm` help/manual/website
documentation has no new parity surface.

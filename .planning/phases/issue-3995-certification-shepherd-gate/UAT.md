# UAT — issue #3995 shared connector-certification Shepherd gate

## Operator-visible checks

1. Run the structural projection check:

   ```sh
   go run ./cmd/agentcontractgen check
   ```

   Expected: the canonical contract and all eight registered Claude/Codex/Pi/OpenCode projections
   are current. This remains green even when no connector is certified.

2. Evaluate GitHub `integrate_sub_pr` using the generated current inputs through the public
   `internal/agentcontract` API.

   Expected: deterministic `RETRY` containing
   `capability/github/capability:check/live_evidence`. A file, a reachable route, or
   `implemented: true` does not substitute for accepted live evidence.

3. Evaluate a complete generated fixture for the same transition.

   Expected: `PROCEED`. Removing one binding criterion returns that criterion's exact failure ID;
   malformed/version-unknown input returns `HALT` with its exact cell/evidence coordinate.

4. Attempt each protected state transition with a non-`PROCEED` verdict.

   Expected: the gate blocks `integrate_sub_pr`, `accepted`, `ready_parent`, and `human_ready`
   while preserving the original deterministic failures for Shepherd retry/halt handling.

## Safety acceptance

The evaluator reads only the declared generated JSON beneath its supplied root. It creates no
evidence, has no provider action dependency, and does not access a credential or mutate production
state. A future #3989 proof schema revision must be integrated explicitly; unsupported versions
halt closed.

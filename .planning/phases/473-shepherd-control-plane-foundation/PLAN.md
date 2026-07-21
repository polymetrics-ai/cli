# Plan: #473 Autonomous Shepherd Control-Plane Foundation

Issue: https://github.com/polymetrics-ai/cli/issues/473
Parent: #471
Parent PR: #472
Branch: `feat/473-shepherd-control-plane-foundation`
Base: `feat/471-pi-agent-session-shepherd` at `c8d86291`

## Objective

Finish and adversarially harden the existing in-process control-plane foundation so the complete
autonomous replacement in #474-#481 can depend on a correct lifecycle, SDK boundary,
target-evidence collector, durable state store, and exclusive run lease. The temporary diagnostic
lanes in this foundation are read-only; that is not the replacement product boundary. This issue
does not yet implement autonomous scheduling or GitHub mutation.

## Workflow and skills

- Issue-first delivery and the parent orchestrator contract.
- `gsd-programming-loop`; the repository adapter lacks the command on this base, so the documented
  manual RED/GREEN/refactor fallback is active.
- `javascript-testing-patterns` and `architecture-patterns`.
- Repository security/agent routing from
  `.agents/agentic-delivery/references/required-skills-routing.md`.

## Tasks

1. Linearize first-wins stop/shutdown cancellation and join terminal persistence.
2. Pin state-root path/device/inode, keep root-security failures outside best-effort cleanup, and
   document macOS as trusted same-user local state rather than a hostile same-UID boundary.
3. Make lease resolution revalidate the highest exact anchor before accepting a missing successor.
4. Confirm every linked owner is the authoritative head before return; regenerate tokens after an
   orphan publication; verify epoch authority before and after cleanup.
5. Bound epoch names to 12 digits, reject malformed reserved epoch names, and never let a delayed
   lower epoch delete a newer epoch.
6. Reject impossible assessed state: empty assessments, failed+halted mixtures, missing interrupted
   lanes, non-finite aggregate scores, and lane gates absent from the aggregate.
7. Retain the inherited exact-target, SDK cleanup/deadline, extension reservation, and state
   projection hardening as regression-tested foundation behavior.
8. Bind PR evidence to the exact local repository rather than ambient `GH_REPO` or a matching fork.
9. Make crash reconciliation interrupt every unfinished lane, including a checkpoint saved before
   its first lane dispatch.
10. Do not acknowledge `stop` during initialization until a durable cancellation checkpoint can be
    represented.
11. Quarantine timed-out AgentSessions and setup tasks until genuine settlement, retain mutator
    ownership, and let `close()` boundedly rejoin them.
12. Persist shutdown interruption when initialization is still capturing the target, and poison the
    runner after any rejected settlement so failed cleanup can never admit another mutator.
13. Compare canonical timestamps by instant, reject mixed failed/halted aggregates and pending work
    in interrupted checkpoints, and normalize controller outcomes to the persisted invariants.
14. Verify terminal provider/model routing after execution as well as before prompt dispatch.

## Verification

```bash
node --test .pi/extensions/shepherd/*.test.ts
pi --list-extensions
git diff --check
go vet ./...
go test ./...
go build ./cmd/pm
make verify
```

The sub-PR targets `feat/471-pi-agent-session-shepherd` and uses `Refs #473` plus `Refs #471`.

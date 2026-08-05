# VERIFICATION — issue #3714 parent readiness

Status: planned; execution results are recorded as the reconciliation proceeds.

## Checklist

- [ ] Existing parent branch contains current `origin/main`.
- [ ] Conflicts (if any) preserve #3730 destructive-write confirmation behavior.
- [ ] Canonical source retains complete Codex, Claude, and Pi policy/projection metadata.
- [ ] Generated projections were regenerated through `agentcontractgen sync`, not hand-edited.
- [ ] `agentcontractgen check` confirms no projection drift.
- [ ] Focused generator and integrated-main package tests pass.
- [ ] Reconciliation commit is on the existing parent branch and pushed.
- [ ] PR #3723 checks are green and no actionable review finding remains.
- [ ] PR #3723 is ready for captain review; it is not merged.

## CLI parity

Not applicable: this phase changes parent integration evidence and generated worker projections,
not a `pm` command, flag, help topic, connector surface, manual, or website documentation.


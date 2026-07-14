# Automated Review Routing Loop

Use this workflow to choose local automated review coverage after implementation and local
verification. The default route is local, runtime-native review by an independent reviewer/verifier
agent. Remote PR-bot review is not required by default.

## Policy

Automated review coverage is satisfied by a recorded local review pass that is independent from the
implementing worker and bound to the exact candidate head or diff range. The review output is input,
not authority: every actionable finding needs a disposition before handoff.

## Review order

1. Run focused local verification for the changed scope.
2. Run an independent local `reviewer` pass for general implementation quality and scope.
3. Add a local `security-auditor` pass when the change touches auth, secrets, external effects,
   filesystem boundaries, untrusted input, or privileged operations.
4. Add a local `verifier` pass when command evidence, reproducibility, or gate interpretation is the
   main risk.
5. Escalate to human review when local reviewers report a human-gated concern or when repository
   branch protection requires human approval.

## Decision table

| Condition | Route | Notes |
| --- | --- | --- |
| Normal implementation change | `local_reviewer` | Independent read-only local review pass. |
| Security-sensitive or external-effect change | `local_security` | Add security-auditor coverage and disposition findings. |
| Verification or reproducibility risk | `local_verifier` | Add verifier coverage for exact commands and head. |
| Local review unavailable or human-gated finding | `human` | Record `needs_human`; do not weaken the gate. |

## Review coverage record

Record review routes in the PR body, parent issue state ledger, phase artifact, or worker handoff:

- route: `local_reviewer`, `local_security`, `local_verifier`, `human`, or `blocked`;
- exact head SHA or diff range;
- reviewer role/runtime;
- reviewed files/scope;
- findings and dispositions;
- verification rerun after accepted fixes;
- blocker and human-gate status, if applicable.

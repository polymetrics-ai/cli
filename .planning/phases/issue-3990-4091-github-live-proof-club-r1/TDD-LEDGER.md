# Issues #3990 and #4091 — TDD ledger

| Slice | Red | Green | Refactor / evidence note |
| --- | --- | --- | --- |
| Lifecycle/manual fallback | `gsd-sdk query init.phase-op issue-3990-4091-github-live-proof-club-r1` returned `phase_found: false`; the official roadmap-phase workflow cannot own this combined captain task. | Manual issue-backed CONTEXT/PLAN/TDD/VERIFICATION artifacts created before production edits; all required command prompts are generated and executed inline. | Reason is compatibility, not a lifecycle exemption. |
| #3990 live proof | RED pending: existing verification proves local/multi-process admission but contains no real-binary whole-surface provider proof. | Pending. | Must assert touched resources, sends/not-sent, waits, terminal results, and cleanup. |
| #4091 live proof | RED is committed in `.planning/phases/issue-4091-github-destination-modes-r1/VERIFICATION.md`: live credentialed proof is explicitly missing. | Pending. | Must run real plan/preview/approval/execute plus read-back and refusal paths. |
| Proof driver changes | Pending only if current harness cannot encode the required sanitized evidence. | Pending. | No speculative harness rewrite. |

## Red/green rule

Any defect discovered live receives a production-path failing regression before source changes. A
refusal regression must assert its typed error and negative side effects; a passing exit code or
absence of panic is never GREEN.

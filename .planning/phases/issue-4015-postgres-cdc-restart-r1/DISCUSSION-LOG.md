# PostgreSQL CDC Restart Recovery — Discussion Log

## 2026-08-18 — autonomous brief intake

The supervisor brief supplies the product decision boundary and acceptance evidence, so no human clarification is required before investigation.

### Resolved context

- Delivery is a direct PR from `fm/cli-cdc-resume-fix-r1` into `integration/4015-mvp-flat-r1`, with `Refs #4015` and a title naming 0.2.1.
- The task is a behavior repair, not certification-only work: reproduce red, add the failing regression, implement the smallest correct fix, prove exact target state live, verify, review, commit, push, and open the PR.
- The 0.2.0 release branch and PR #4250 are out of scope and must remain untouched.
- PostgreSQL CDC's accepted design is logical replication (`pgoutput` v2), not timestamp/cursor polling. A result that merely accepts the current checkpoint without proving its position is explicitly disallowed.

### Gray areas to resolve from evidence

1. Does the failing restart checkpoint actually identify itself as a polling checkpoint, or is the error text masking a decode/version/identity mismatch?
2. Which layer writes that checkpoint: the logical-replication CDC machine, bootstrap coordinator, app pipeline, or certification harness?
3. At what interruption boundary is state durable: receiver receipt, checkpoint commit, or standby acknowledgement?
4. Does the current capability artifact claim more than the live implementation proves after restart?

### Decision rule

Trace the checkpoint from write through restart validation and correlate it with the live target. Prefer repairing the producer/consumer contract around the canonical logical-replication checkpoint. Reject any fix that permits polling fallback, loses source identity, advances beyond durable receipt, or cannot prove post-interruption row multiplicity exactly one.

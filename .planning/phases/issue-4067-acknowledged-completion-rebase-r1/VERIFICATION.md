---
phase: issue-4067-acknowledged-completion-rebase-r1
status: pending
---

# #4067 verification checklist

## Pre-production gates

- [x] Controlling Sol correction directive read completely.
- [x] #3862 child tree, #3864, #4046, #4059 checks/comments/reviews, and the three Sol/R9 reports read before edits.
- [x] New child #4067 created, linked under #3864, and read back before production/generated changes.
- [x] Existing branch/PR custody preserved; rejected `883a86cf0040d559edcd4777413d1c2de20cd94a` is an immutable baseline.
- [x] CodeGraph absence recorded; required skills and GSD/agent-contract checks completed.
- [x] Named inline/manual GSD problem, context, discussion, TDD plan, execution record, verifier evidence, review, summary, and run-state artifacts exist before code.
- [x] Behavioral RED exits non-zero for the durable completion leak before production mutation: all seven canonical modes retained acknowledged/unrelated state first, then observed zero returned run plus durable reopened `running` run.

## Required focused matrix

- [ ] Exact post-checkpoint/pre-completion two-App interleaving.
- [ ] Returned/durable identity and reopened terminal truth.
- [ ] Target running/exact-checkpoint eligibility and fail-closed changed/missing/terminal targets.
- [ ] Winner/acknowledged checkpoint and unrelated state preservation.
- [ ] Cancellation after acknowledgement.
- [ ] All seven canonical modes.
- [ ] Focused race detector.
- [ ] #4046 typed-conflict and R7/R8 regression suite.

## Subsequent gates

- [ ] Canonical website generator/check refreshes `website/lib/docs.generated.ts` only.
- [ ] Canonical certification generator/check refreshes `internal/connectors/certifications/flow-matrix.json` only.
- [ ] Affected tests, lint, vet/build, and individual repository gates pass after the heavy-validation window notification.
- [ ] Manual `verify-work` and `code-review` records contain evidence and every finding disposition.
- [ ] Fresh #4067 no-mistakes run starts at 0/5 without `--yes`; old run is not queried for control or modified.
- [ ] Existing draft #4059 is updated normally, stays unmerged, and exact-head CI is green before requesting an independent Sol audit.

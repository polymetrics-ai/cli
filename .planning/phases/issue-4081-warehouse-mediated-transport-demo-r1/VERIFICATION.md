# #4081 — verification checklist

## Phase and topology

- [x] Live parent/sub-issue/PR reconciliation completed through `gh-axi`.
- [x] No existing open or closed issue/PR owns the production Transport
  construction walking slice; #4081 is the single direct #4015 child.
- [x] #4015 visibly links #4081 and records its prerequisite/base guard.
- [x] GSD adapter doctor, all five command sources, and canonical contract check passed.
- [x] Required planning skills are loaded and recorded.
- [x] Required combined branch is authoritative head
  `e7d2b2963fc1dd164f63b31fccb8a3bab8084bec`; the squashed #4019 content is
  proven by exact #4077 Transport blob identity with `aaf288d...`, not ancestry.
- [x] Final `feat/4081-warehouse-mediated-transport-demo` branch exists from
  that remote combined head only.

## Before completion

- [ ] Planning artifacts committed on final child branch before production edits.
- [ ] RED commit precedes every implementation commit and fails for the expected
  construction/stage/reopen/order failures.
- [ ] GREEN tests prove independent durable reopen, exact closed GitHub legs,
  receipt-before-CAS, safe replay, and #4079 isolation.
- [ ] Fresh exact `pm` binary digest and size are recorded; bounded local
  faithful-server run asserts returned record counts—not only exit code.
- [ ] Stage/Parquet/DuckDB identity, independent provider read-back, typed
  inverse cleanup, second cleanup, and zero residue are demonstrated.
- [ ] Live provider result is recorded only if the approved PM credential boundary
  resolves normally; otherwise a precise safe blocker is recorded.
- [ ] Targeted tests, split local gates, GSD verify-work/review, no-mistakes,
  exact-head CI, and automated-review disposition are terminal green/allowed.
- [ ] Draft PR is opened only to `docs/4015-connector-release-certification`,
  with `Refs #4081`, `Refs #4015`, `Refs #3862`, and `Refs #4077`; never merged.

## Resolved base-admission evidence

```text
remote combined head: e7d2b2963fc1dd164f63b31fccb8a3bab8084bec
accepted #4077 source head: aaf288d069adc1b67a09500afcca4be4a6d1bab3
types.go blob (combined/source): 7b5eafd34c78bea690dcb23eee716ea840ea813e
orchestrator.go blob (combined/source): 442ea03826f6385da3d6a04fd65af3936dc35188
registry.go blob (combined/source): c9ac2eeb6045546390c04f30769601028f8e5571
transport_test.go blob (combined/source): b191b8e2905afa397e9c184f18ee26952dda00ac
combined tree contains .planning/phases/issue-4077-transport-record-isolation-r1/
```

The non-ancestry result is expected for the #4019 squash merge. Content identity
is the required proof. No correction-loop budget is consumed.

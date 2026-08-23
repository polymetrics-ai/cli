# Verification checklist — issue 4336

## Before push

- [ ] `e338cd301` confirmed an ancestor of the branch start and `origin/main`.
- [ ] Red provider-document importer tests and exact failure evidence recorded.
- [ ] Each of the seven named cases passes through the real importer path with
  its deliberate support/retention/bound result.
- [ ] Existing importable connector projection is byte-identical.
- [ ] A pathological deep document remains refused by the finite new bound.
- [ ] `git diff --check` passes and `internal/connectors/defs/github/rate_limits.json` is unchanged.
- [ ] Changed-package tests use `-timeout 20m` and pass.
- [ ] Generator, snapshot, static, and required repository verification entry points pass or are recorded with an exact blocker.
- [ ] Inline GSD review records no unresolved actionable finding.

## PR read-back

- [ ] API reports PR base exactly `main`.
- [ ] PR body includes `Refs #4336`, per-case reasoning, GSD/TDD evidence,
  skills, verification, safety, delivery, and review routing records.
- [ ] No merge is performed.

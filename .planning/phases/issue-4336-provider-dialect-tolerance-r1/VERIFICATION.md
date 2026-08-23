# Verification checklist — issue 4336

## Before push

- [x] `e338cd301` confirmed an ancestor of the branch start and `origin/main`.
- [x] Red provider-document importer tests and exact failure evidence recorded.
- [x] Each of the seven named cases passes through the real importer path with
  its deliberate support/retention/bound result, using source-derived reduced
  operations. This is not represented as full-artifact proof.
- [x] Existing importable connector projection is byte-identical:
  `TestSourceImportPreservesFrozenGitHubArtifacts` verifies the exact
  checked-in GitHub descriptor bytes and SHA-256; it passed in the full changed
  package suite. Full Batch-1 provider verification remains blocked by the
  independently measured request-contract inventory.
- [x] A pathological deep document remains refused by the finite new bound.
- [x] `git diff --check` passes and `internal/connectors/defs/github/rate_limits.json` is unchanged.
- [x] Changed-package tests use `-timeout 20m` and pass.
- [x] Generator, snapshot, static, and required repository verification entry points pass; exact commands are in `RUN-STATE.md`.
- [x] Inline/manual GSD review records no unresolved actionable finding in `INLINE-CODE-REVIEW.md`.

## PR read-back

- [x] API reports PR base exactly `main` for #4339.
- [ ] PR body includes `Refs #4336`, per-case reasoning, GSD/TDD evidence,
  skills, verification, safety, delivery, and review routing records.
- [x] No merge is performed.

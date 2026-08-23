# Verification checklist — issue 4323

## Before push

- [ ] Focused red test run recorded before importer production edit.
- [ ] Focused green test run passes with all recursive and non-recursive cases.
- [ ] Real affected connector `source-import --check` passes without credentials.
- [ ] GitHub source lock measurement is `3420025` bytes and SHA-256
  `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`.
- [ ] GitHub descriptor measurement is `43354021` bytes and SHA-256
  `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.
- [ ] `git diff --check` passes; `internal/connectors/defs/github/rate_limits.json` is unchanged.
- [ ] Repository generated-file/snapshot checks pass through `make verify`.
- [ ] Full `make verify` passes locally; no test timeout exceeds the shared 20m budget.
- [ ] Inline GSD code review records no unresolved actionable findings.

## PR read-back

- [ ] GitHub API reports PR base exactly `main`.
- [ ] PR body contains `Refs #4323`, GSD/TDD evidence, skills, verification,
  safety, and review-route disposition.
- [ ] No merge is performed.

# Verification checklist — issue 4323

## Before push

- [x] Focused red test run recorded before importer production edit.
- [x] Focused green test run passes with all recursive and non-recursive cases.
- [x] Real affected Grafana import succeeds without credentials: 314 operations
  retained and 52 recursive-schema gaps emitted from its pinned public artifact.
- [x] GitHub source lock measurement is `3420025` bytes and SHA-256
  `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`.
- [x] GitHub descriptor measurement is `43354021` bytes and SHA-256
  `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.
- [x] `git diff --check` passes; `internal/connectors/defs/github/rate_limits.json` is unchanged.
- [x] Repository generated-file/snapshot checks pass through `make verify`.
- [x] Full `make verify` passes locally; no test timeout exceeds the shared 20m budget.
- [x] Final low-load `make verify` exit is 0 after the #4326 commits: full
  `cmd/connectorgen` passed in 189.173s, `internal/cli` in 528.045s, and
  `internal/connectors/boundary` in 406.747s; lint, generated/snapshot,
  operation-evidence, certification, boundary, and release checks all passed.
- [x] Inline GSD code review records no unresolved actionable findings.
- [x] #4326 red test records current OpenAPI 3.0 descriptive-sibling rejection.
- [x] #4326 focused green tests prove `description` and `summary` import,
  schema `readOnly` and `type` handling are bounded, an unequal `type` is an
  explicit source-bound gap, a response `content` sibling remains rejected,
  and existing `x-` allowlisted behavior holds.
- [x] Real Asana pinned source imports all 249 operations and verifies with
  `--check`, without connector-local source mutation.
- [ ] GitLab live source import: blocked before resolver preflight by the
  provider's `master` artifact drifting from its historical strict lock.
- [ ] Docker Hub pinned source import: advances beyond the addressed 401
  description sibling, then rejects the provider's unrelated dangling schema
  reference `#/components/responses/team_repo`; no source mutation or
  suppression was applied.
- [x] `operation-evidence --check` passed during `make verify`; the diff of
  both operation-evidence generated files against `origin/main` is empty.

## PR read-back

- [ ] GitHub API reports PR base exactly `main` after the final push.
- [ ] PR body contains `Refs #4323` and `Refs #4326`, GSD/TDD evidence, skills, verification,
  safety, and review-route disposition.
- [ ] No merge is performed.

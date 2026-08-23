# Run state — issue 4323

State: verification-green; final evidence commit, push, and PR read-back pending

- Issue opened: #4323.
- Base ancestry confirmed: `cf493b834` is an ancestor of the branch start.
- GSD adapter and canonical contract checks passed.
- Red/green source-import contract recorded for direct, mutual, deeply nested,
  finite, and unused-schema behavior.
- Real Grafana source import emitted 52 explicit recursive-schema gaps while
  retaining 314 operations; its temporary v2 lock and output have been removed.
- Focused recursive-schema behavioral tests, `go vet ./...`, `go build ./cmd/pm`,
  `git diff --check`, frozen GitHub artifact measurements, and full `make verify`
  are green. The shared 20-minute test timeout remains unchanged.
- Inline GSD code review found no unresolved actionable findings.
- The widened #4326 test recorded red failures for response `description` and
  `summary`, schema `readOnly`, and a non-equivalent schema `type` sibling.
- The shared resolver now retains the bounded allowed fields; a non-equivalent
  schema `type` is exact descriptor data plus a pointer-named source-bound
  runtime gap, while a response `content` sibling remains rejected.
- Asana's pinned public source imports and verifies all 249 operations. GitLab
  has provider source drift from its strict historical lock; Docker Hub's
  pinned artifact has an unrelated dangling schema reference. Neither is
  suppressed, refreshed, or repaired by this PR.
- Final low-load `make verify` exited 0 with the shared 20-minute budget
  unchanged. The full suite passed, as did lint, generated/snapshot checks,
  operation-evidence, certification, boundary, and release checks.
- Frozen GitHub measurements remain source lock `3420025` bytes /
  `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6` and
  descriptor `43354021` bytes /
  `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.
- Next gate: commit the updated evidence, push, update PR #4327 with both
  issue references, confirm API base `main`, and stop without merging.

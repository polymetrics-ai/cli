# Run state — #4344

- State: planning complete; production edits have not started.
- GSD: prompts resolved inline on 2026-08-24; no role spawning.
- Baseline: `sourceProjectionGeneratedCommandPath` transforms raw `SourceID`
  only by replacing `/` and `_`, while commandrunner rejects `{`/`}`.
- Dependency boundary: current `main` does not carry Bitbucket/GitLab
  operation descriptors. Their regeneration and 50-command sweep must use the
  reviewed batch-1 input artifact without importing unrelated connector changes.

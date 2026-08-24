# Run state — #4344

- State: pull request open; awaiting captain-held review and integration.
- GSD: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`,
  and `code-review` prompts resolved inline on 2026-08-24; no role spawning.
- Red: restoring the raw `SourceID` command identity made
  `TestSourceProjectionGeneratedParameterizedCommandIsRuntimeValidAndStable`
  fail with `generated command path retained a raw source parameter`.
- Green: new commands encode the normalized `METHOD path` endpoint in hex;
  only legacy generated paths rejected by runtime identifier validation migrate.
  Valid legacy generated paths remain unchanged.
- Generated artifacts: no checked-in surface changed. `surface-sync --check`
  scanned 552 connectors and found zero drift. The current Bitbucket definition
  has only three implemented commands and no `sources/` directory.
- Dependency boundary: `GOFLAGS='-p=3' go run ./cmd/connectorgen source-import
  bitbucket --check` cannot run here because
  `internal/connectors/defs/bitbucket/sources` does not exist. The reviewed
  batch-1 source descriptor is required for the requested 50-command/28-path
  binary sweep; this foundation branch must not copy unrelated connector work.
- PR: https://github.com/polymetrics-ai/cli/pull/4346. Read-only GitHub API
  read-back confirmed base `main` and head
  `fm/cli-runtime-valid-generated-command-paths-r1`.

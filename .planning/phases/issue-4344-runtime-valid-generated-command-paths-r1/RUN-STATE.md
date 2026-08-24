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
- Generated artifacts: no connector-local artifact is checked into this
  foundation PR. `surface-sync --check` scanned 552 connectors and found zero
  drift. The current Bitbucket definition has three implemented commands and
  no `sources/` directory.
- Exact batch proof: copied the independently measured Bitbucket batch fixture
  into an isolated temporary directory read-only, then ran this branch's
  `source-import`. It imported 297 operations and updated exactly 28 CLI paths.
  The resulting fixture retained 50 implemented commands, zero paths with
  `{`/`}`, and 28 `api op-<hex>` commands. A `go build -overlay` binary using
  those regenerated `writes.json`, `cli_surface.json`, and `api_surface.json`
  artifacts ran all 50 implemented commands from an initialized no-credential
  project: 50 reached `error: missing --credential`, 0 were unknown, 0 were
  invalid paths, and 0 had another outcome.
- The batch fixture is intentionally not committed here: the fix changes the
  shared generator, while the batch owns its 2,775-line connector-local surface
  reconciliation. This proof executes the PR's generator and binary without
  absorbing that unrelated artifact diff.
- PR: https://github.com/polymetrics-ai/cli/pull/4346. Read-only GitHub API
  read-back confirmed base `main` and head
  `fm/cli-runtime-valid-generated-command-paths-r1`.

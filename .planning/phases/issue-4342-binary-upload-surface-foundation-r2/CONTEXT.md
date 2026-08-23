# Context — issue #4342 binary upload surface foundation

## Task Delivery Header

- Issue: Refs #4342 — feat(connectors): add binary upload CLI and certification surface
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, with committed implementation, targeted behavioral tests, generated-file checks, and applicable local verification recorded.
- Working branch: fm/cli-binary-upload-surface-foundation-r2
- Task: Expose only declaration-bound binary upload write actions through a new `binary_upload` connector-command intent, preserving the existing plan → preview → approval → execute pipeline; add a separate truthful upload certification projection and evidence stage.
- Verification: Targeted engine, commandrunner, CLI, certification, and connectorgen tests; `connectorgen validate`, `surface-sync --check`, operation-evidence checks, generated docs checks, and applicable package-level Go checks with `GOFLAGS=-p=3`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A binary-upload command is admitted only for a declared binary/base64/multipart file write action | live | Commandrunner preflight accepts the declared GitHub binary action and rejects ordinary JSON, non-file multipart, undeclared, and mismatched source-field actions before any request runner can execute. |
| The command follows the established approval-bound write path | live | CLI/app test creates a connector-command plan, previews it, supplies a valid approval, and proves the exact fixture bytes and approved SHA-256 reach the test provider; changing the file invalidates the approved plan before I/O. |
| Upload certification cannot call a refusal a pass | live | Certification stage test makes a declared candidate return a refusal and asserts capability/stage status is `blocked` or `not_live`, `Passed=false`; a fixture transfer asserts byte count, digest, response, and cleanup before `pass`. |
| Upload is independently represented in generated certification/evidence surfaces | live | Sweep and operation-evidence tests assert `binary_upload` has its own class and rejects `file_upload` as executable. |
| Help, manuals, and website projections describe the same closed surface | live | Generator tests and drift checks show `binary_upload` command flags, plan/approval requirement, and no generic path/body/URL channel. |

## Decisions recorded by discuss-phase (inline fallback)

- Captain ruling fixes `binary_upload` as a release-required, eighth first-class capability; no product question remains open.
- A `binary_upload` command is a write-action binding, never an operation executor or raw upload primitive. It must carry `write`, not arbitrary body/path/URL input.
- The command reuses `BuildWriteCommand` and the existing connector-command plan. It is therefore never executable directly by `commandrunner.Run`.
- `file_upload` remains declarable but unexecutable; it must not be mapped to `binary_upload` or receive an executor.
- A successful safety refusal is still useful evidence, but reports `blocked` or `not_live` and the owning stage is not passing. Only an evidenced transfer plus cleanup can set upload `pass`.
- Existing binary upload/base64/multipart write paths retain source confinement, digest binding, byte caps, and declared media policy. Any missing declaration metadata is a preflight error, not a caller override.

## GSD fallback

`scripts/gsd prompt discuss-phase issue-4342-binary-upload-surface-foundation-r2` and `scripts/gsd prompt plan-phase issue-4342-binary-upload-surface-foundation-r2 --tdd` were resolved. The compatible isolated Pi/GSD runtime is unavailable in this Codex worktree and the repo's single-worker contract forbids role spawning, so the generated workflow is being performed inline. This records the same decisions, TDD slices, verification, and later review evidence without weakening any gate.

## Required skills loaded

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-safety`, `golang-security`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `vercel-react-best-practices`, and `vercel-composition-patterns`.

The website skills are loaded for the generated documentation projection; no website component architecture or dependency change is intended.

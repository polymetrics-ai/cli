# Plan — issue 4261 certification foundation port

## Task Delivery Header

- Issue: Refs #4261 — certification foundation parity projection port to main
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, local gates green, live GitHub
  evidence regenerated on this branch, and required review routing recorded.
- Working branch: fm/cli-foundation-to-main-r1
- Task: Port only the certification parity projection and safe evidence
  publication foundation from `ac2944115`, regenerate the GitHub artifacts,
  re-run the authorized live proof, and preserve `main` release and planning
  history.
- Verification: targeted connectorgen tests; generated-artifact checks; GitHub
  live proof; CLI/docs inspection as applicable; `make verify`; diff-base and
  release-metadata assertions; review routing and API base read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Foundation commands project parity and publish only safe proof fields | live | New focused Go tests fail before the port and pass after it; command output and checked generated artifacts contain the expected projection. |
| GitHub sweep is derived from the `main` bundle | live | The repository generator recreates `internal/connectors/defs/github/certification-sweep.json`, and its check accepts the result. |
| Live evidence matches this branch without credential disclosure | live | The approved script produces a redacted evidence record whose verification accepts the real read result and contains no credential-shaped value. |
| The PR cannot regress main release metadata or planning history | live | `git diff origin/main...HEAD -- CHANGELOG.md .release-please-manifest.json` and deletion checks are empty. |
| Repository-wide contract remains executable | live | `make verify` passes before push and the installed pre-push hook invokes the same command. |

## TDD execution slices

1. **Red:** introduce the upstream focused certification tests without their
   implementation and run the target package; record the missing-command or
   missing-symbol failure.
2. **Green:** port the source implementation files, run the focused tests, and
   regenerate the GitHub sweep from `main` definitions.
3. **Refactor/proof:** verify the generated diff is source-consistent but
   main-derived; run authorized live proof, review the redacted record, and
   run all local gates.

## CLI docs parity

The new `connectorgen` maintenance subcommands are internal generator surfaces,
not `pm` end-user commands. `docs/architecture/connector-certification-design.md`
is the applicable documentation surface. `pm help`, bare namespace behavior,
`docs/cli/**`, website documentation, completion, and manual regeneration are
not applicable because no `pm` command or connector command contract changes.

## Skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-documentation`, `golang-design-patterns`, and
`golang-structs-interfaces`.

Resolved command path: `scripts/gsd doctor`, `scripts/gsd sources` for
`discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and
`code-review`, then their generated prompts. Inline execution is required by
the canonical single-worker contract.

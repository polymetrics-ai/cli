# Issue #4193 Plan — Scalable command help

## Task Delivery Header

- Issue: Refs #4193 — leaf `--help` must render before project, credential, or required-flag resolution.
- Base branch: `integration/4015-mvp-flat-r1` (`ff6a87101` before production edits).
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: PR open against `integration/4015-mvp-flat-r1`, with local required gates recorded and review coverage initiated.
- Working branch: `fm/cli-command-usability-scalable-r1`.
- Task: Make all legacy CLI leaf help render as documentation before handler effects, remove the per-command leaf-help mechanism, and prove every documented legacy leaf works both outside and inside a project.
- Verification: Build the binary; enumerate the binary surface; run `--help` and `-h` for every derived legacy leaf in both project states; run focused and package tests, generated-doc stability, lint, connector boundary, and the repository gates listed in `VERIFICATION.md`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every documented legacy leaf renders help outside a project | live | The actual `cli.Run` public entry point returns 0 and writes the leaf's manual token for every path parsed from registered wrapper manuals; pre-fix commands either opened a project or executed required work. |
| Short help and project-present help remain correct | live | Each derived leaf runs with `-h` in a temporary directory and with `--help` after `pm init`; output contains the owning manual's `NAME` and no stderr. |
| The switch mechanism cannot silently return | live | The test derives wrapper names from `cobraLegacyCommands` and leaf paths from their manual synopsis/usage. A new wrapper without a manual or a documented leaf without working help fails the test. |
| Declaration-backed connector leaves remain pre-dispatch | live | The test derives all paths from `CommandSurface.Commands` and verifies both help spellings resolve to a `NAME` manual before dispatch for every generated connector command. |
| Bad commands stay errors and approval syntax stays guarded | live | Tests assert unknown commands retain usage exit code 2 and malformed approval carriers fail rather than rendering help. |
| Help never executes work | live | Representative required-flag and effectful leaf paths render their documented manual, not a runtime result, when help is present. |

## Required skills and references loaded

- `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
  `golang-security`, `golang-safety`, `golang-documentation`, and
  `golang-spf13-cobra`.
- `.agents/agentic-delivery/references/required-skills-routing.md`.
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`.

## GSD lifecycle

Validated adapter with `scripts/gsd doctor`, resolved `discuss-phase`,
`plan-phase`, `execute-phase`, `verify-work`, and `code-review` through
`scripts/gsd sources`, and ran `go run ./cmd/agentcontractgen check`.
Generated prompts are executed inline because this environment lacks the required
Pi runtime; the canonical single-worker contract also forbids role spawning.

## TDD slices

1. **RED — executable inventory test.** Replace the two-row
   `TestChangedCLICommandHelpIsExecutable` with a programmatic test that obtains
   registered legacy wrappers and parses each wrapper manual's synopsis/usage to
   enumerate every leaf. Assert `--help`, `-h`, combined flags, and initialized
   project behavior. Add bad unknown-command and approval-carrier assertions.
2. **GREEN — generic help resolution.** Remove `legacyLeafManualTopic`; add a
   single generic wrapper pre-handler resolution that renders the wrapper manual
   for a leaf help request. Ensure every registered wrapper has a manual and add
   missing stable manuals where the current help function has no content.
3. **REFACTOR — error and docs parity.** Keep unknown commands in the established
   usage taxonomy; make manual/topic handling consistent for aliases and hidden
   wrappers. Regenerate only affected tracked CLI manuals; no website prose change
   is expected because command semantics and documented flags do not change.
4. **VERIFY — built binary.** Rebuild and run all 63 legacy leaves for both help
   spellings outside and inside a project, then the same root-level proof for
   every dynamically exposed connector. Record exact counts and results in
   `VERIFICATION.md`; the declaration-driven leaf resolver is swept in-process
   from all generated `CommandSurface.Commands` to avoid thousands of redundant
   process startups.

## Commit checkpoints

1. Planning evidence before source edits.
2. Red test evidence, then implementation and generated-document update as one green slice.
3. Review fixes, if any, as a separate green checkpoint.

## CLI help/manual/website parity

- Runtime help: required and covered by exhaustive entry-point tests plus built-binary inventory.
- Bare namespaces: existing manual behavior is retained and rerun for every registered wrapper.
- `docs/cli/**`: regenerate if new manual topics change the generated set; verify byte stability.
- Website: no command spelling, flag, or user workflow changes are planned. Run the docs generator
  twice for byte stability and verify references; record the no-source-change rationale in the PR.
- JSON: help uses the existing `CommandManual` envelope and is covered by regression tests.

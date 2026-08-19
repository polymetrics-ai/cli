# Plan — issue 4264 operation-backed mutation action kinds

## Task Delivery Header

- Issue: Refs #4264 — derive operation-backed mutation action kinds
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, exact GitHub API base read-back
  equals `main`, all local gates are green, and required automated-review
  routing is recorded. A captain retains merge authority.
- Working branch: fm/cli-delete-action-kind-fix-r1
- Task: Extend the generated certification parity classifier so command-linked
  operations deterministically project `create`, `update`, `upsert`, `delete`,
  or `custom` when `writes.json` is absent; preserve `writes.json` behavior;
  reject indeterminate mutations; regenerate GitHub and Zoom sweep artifacts.
- Verification: focused red/green `cmd/connectorgen` tests; artifact
  regeneration and `--check` for GitHub and Zoom; byte-drift/snapshot checks;
  `make connector-boundary`; all repository local gates including full
  `make verify` before push; PR check and review inspection.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| An operation-backed delete receives a generated delete kind | live | Classifier/sweep tests assert the operation-derived `write_action_kind` is `delete`; the pre-change implementation leaves it empty. |
| A `writes.json` delete remains correctly derived | live | Classifier test asserts a declared write action projects `delete` without consulting or regressing the operation path. |
| Non-delete mutations classify correctly | live | Table-driven classifier tests assert create/update/upsert/custom operation kinds. |
| An indeterminate operation-backed mutation fails loudly | live | A test asserts a concrete generator error rather than an empty or `custom` action kind. |
| Generated delete cells are selectable for real bundles | live | Regenerated GitHub and Zoom sweeps contain populated delete rows and deterministic selector checks identify them independently. |

## TDD execution slices

1. **Red:** add focused classifier and real-bundle sweep assertions covering
   operation-backed delete, `writes.json` delete, non-delete mutation, and an
   indeterminate operation; record the expected failure before code edits.
2. **Green:** extend the shared classifier only, run focused tests, regenerate
   GitHub and Zoom generated sweeps, then assert artifact byte-drift checks.
3. **Refactor/proof:** simplify derivation around existing normalized action
   kinds, inspect the diff for connector-specific logic, run the full local
   gates and review workflow.

## CLI/docs parity

`connectorgen` is internal generator tooling and no `pm` or connector runtime
command, help text, flag, manual, website page, completion, or public API
changes. CLI/docs/website parity is therefore not applicable.

## Skills and lifecycle

Loaded: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, and `golang-testing`.

Resolved lifecycle: `scripts/gsd doctor`; `scripts/gsd sources` and generated
prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check`.
Inline/manual execution is the compatible fallback because this task's
single-worker environment cannot run Pi GSD role workers.

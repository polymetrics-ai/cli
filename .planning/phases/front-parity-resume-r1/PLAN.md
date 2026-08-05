# PLAN — Front documented-operation parity resume

## Goal

Recover the preserved Front bundle on current `origin/main`, then make every operation in Front's
provider-owned API reference genuinely reachable through the supported declarative mechanisms.
Do not describe the connector as complete while any operation remains `planned`.

## Scope and ownership

- Own: `internal/connectors/defs/front/**`, Front-generated docs/catalog/website data, focused
  CLI generated fixtures, and this phase's artifacts.
- Do not edit shared engine/schema/validator files, another connector, or a generic write executor.
- Use `writes.json` for record-shaped mutations; never mark `rest_write` implemented.
- Keep the duplicate-document row and its recorded rationale. Promote the five bounded binary
  downloads now that current main dispatches `binary_download`.

## Required evidence

1. Establish the actual operation total from Front-owned documentation, not a historic count.
2. Maintain a provider-field research matrix: every path/query/body/form/file request field gets
   a provider URL/section, evidence tier, confidence, and requiredness rationale. The matrix is
   research evidence only until the shared citation convention lands; do not invent a competing
   bundle schema.
3. Prove each `availability: implemented` command passes the real commandrunner preflight and
   execute representative built-`pm` commands without credentials.
4. Regenerate only Front-derived documentation/website artifacts after the bundle stabilizes.

## Execution slices

1. Capture current-main baseline and runtime reachability failures (red evidence).
2. Reconcile Front's current official reference, revise the operation ledger/count, and fill the
   citation matrix.
3. Convert supported operations to streams, writes, direct reads, or bounded binary downloads;
   retain only an exact accepted reason for any remainder.
4. Run surface sync, focused conformance/commandrunner/CLI/build gates, documentation generation,
   and live help/command smoke checks.
5. Update phase evidence, commit coherent green slices, then hand off to no-mistakes.

## Review remediation dependency

Provider-valid fixed command variants and fixtures are owned by this Front lane. Exact conditional
rules are preserved in `internal/connectors/defs/front/schemas/request_contracts.json` as draft-07
composition. Current main cannot evaluate that composition, so final record-driven enforcement and
the completion claim remain paused until the shared normalized-schema composition lane lands. This
lane must not add a private evaluator or modify shared validation/runtime files.

The current review correction keeps that ownership boundary: align template-array cardinality and
call-field nullability in both connector contracts, prohibit Twilio-only secrets in the custom
channel branch, exercise the empty-array template form in the write fixture, and synchronize CLI
help plus cited generated surfaces without changing shared code.

## GSD / skills

Loaded: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-cli`,
`golang-documentation`, `golang-lint`, `gsd-programming-loop`, and `no-mistakes`.

`scripts/gsd doctor` passed, but `scripts/gsd prompt programming-loop ...` is absent from the
adapter and `scripts/programming-loop.mjs` does not exist. This phase follows the manual GSD
fallback recorded in `PROMPTS.md` and retains strict TDD evidence.

## Orchestration decision

`local_critical_path`: this is one connector-owned bundle in an isolated worktree. The active
runtime instruction forbids spawning subagents; no mutable delegation is safe or necessary.

# Plan — issue 4364 typed deferred visibility bridge

## Task Delivery Header

- Issue: Refs #4364 — Batch R1 typed deferred visibility bridge.
- Base branch: `fm/cli-top100-declaration-batch-r1` at
  `687eb1ded6b42cc456f8cc3c1e97f0a84fd042a8`.
- Merges into: `fm/cli-top100-declaration-batch-r1` → `main` only after
  independent review and captain approval; this task does not integrate.
- Delivery: one candidate commit/push from
  `codex/4364-deferred-visibility-r1`, with exact-SHA review evidence.
- Task: add a deterministic, check-only source-evidence report for each
  applicable `mapped_unproven` or `missing_foundation` seven-lane cell in the
  frozen Batch R1 cohort.
- Verification: red → green → refactor, actual-cohort inspection, focused
  non-race Go tests, JSON parsing, source-cohort validation, agent contract,
  diff check, and an inline code review.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion |
| --- | --- | --- |
| Every applicable deferred cell is discoverable | ten real Batch R1 matrices plus source locks | deterministic report has one entry per normalized M-U/M-F cell with exact source ID and lane; no implemented/N/A entry appears |
| Facts and citations are source-bound | source-lock parser plus matrix facts | report method/path/location agree with the cited connector-owned lock; primary and declared supplemental locks are both checked |
| Stable reasons and named capabilities exist | matrices, connector gap ledgers, Foundation Atlas | M-U gets the stable generic authoring prerequisite; M-F resolves a named foundation and known Atlas capability, or validation fails |
| Mutations/deletes/removes remain visible | source semantic facts in real matrices | an explicit source mutation semantic cannot leave direct-write or reverse-ETL hidden as N/A |
| No executable claim is manufactured | report schema/CLI behavior | report has zero executable declarations and no command/operation/stream/write/transport/credential/runtime artifact field; checker makes no network/runtime call |
| Certification cannot erase membership | source-lock cohort and matrix reconciliation | missing/duplicate/unknown IDs or lanes fail before any visibility report succeeds |

## TDD execution slices

1. **Red — command and real cohort:** create the check-only command test that
   requires a deterministic report for all ten matrices, all seven lanes,
   source-bound fact/citation records, and zero executable declarations. It
   fails because the command/projection does not exist.
2. **Green — strict normalized reader:** add a single generic normalizer for
   `source_operations/lanes` and `operations/cells`, using the existing source
   lock parser and source cohort check. Validate matrix membership, exact lane
   coverage, duplicate IDs, citations/facts, and declared supplemental locks.
3. **Green — typed reason/capability resolution:** emit stable bridge reasons;
   resolve M-U to `source.projection-admission.v1`; resolve M-F from the
   connector-local gap ledger or matrix `foundation_atlas` evidence, and reject
   unknown capability names or unstable/incomplete reason facts.
4. **Green — safety edges:** prove exact lock identity, an explicitly declared
   provider identity with exact route, a missing opaque identity, a route
   mismatch, and an ambiguous provider identity are handled deterministically.
   Ensure all current Batch R1 rows validate without changing locks/matrices.
5. **Refactor and prove:** establish deterministic ordering and byte-stable
   JSON output, update the migration convention for this authoring-only
   diagnostic, complete the verification ledger, and run scoped gates.

## Exact intended file ownership

| Path | Responsibility |
| --- | --- |
| `cmd/connectorgen/main.go` | register/help only for the check-only authoring command |
| `cmd/connectorgen/deferredvisibility.go` | generic source/matrix/Atlas evidence normalizer and report; no runtime imports |
| `cmd/connectorgen/deferredvisibility_test.go` | red/green/edge/actual-cohort evidence |
| `docs/connector-canon/OPERATION-EVIDENCE.md` | authoring-only command and no-execution boundary |
| `.planning/phases/issue-4364-deferred-visibility-r1/**` | GSD/TDD/verification evidence |

No source lock, source-lane matrix, connector runtime definition, operations,
writes, streams, CLI surface, transport, descriptor, certification, engine, or
Atlas edit is planned. Atlas was consulted; current M-F records already name
their actual demand. If an unrecorded actual runtime gap appears, stop and
report it before implementing anything.

## CLI help/docs/website parity

`connectorgen deferred-visibility` is a repository authoring/validation tool,
not a `pm` runtime command or connector surface. The applicable surface is
`connectorgen --help` and command `--help`, with tests. `pm` help, manuals,
website docs, generated user-facing help, completions, credentials, and
approval/reverse-ETL documentation are not applicable because no user runtime
behavior or action is added. The migration convention documents the
authoring-only output and no-execution restriction.

## Skills and lifecycle

Loaded: `connector-lane-build-order`, `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, and `go-engineering`.

GSD command sources were resolved with `scripts/gsd doctor` and the official
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and
`code-review` prompts. The documented inline/manual fallback is recorded in
`CONTEXT.md`.

# Schema-v3 source projection/importer foundation — Outreach vertical proof

## Task Delivery Header

- Issue: Refs #4354 — feat(connectors): make Outreach full-surface pilot auditable
- Base branch: main
- Merges into: main
- Delivery: Ordinary pull request open against `main`, with the final code SHA
  independently audited, normal non-force publication completed, and local
  scoped gates recorded.
- Working branch: feat/4354-schema-v3-source-projection
- Task: Make the canonical source importer/projector admit the immutable
  retained schema-v3 Outreach provider document and project every source
  operation into source-backed six-lane evidence. Preserve identity and
  citations, leave unavailable executor shapes as `missing_foundation`, and
  prove the change with a second schema-v3 source case. No connector-local
  generic executor or false `implemented` operation is allowed.
- Verification: focused red/green/refactor Go tests; `connectorgen`
  source-import/check, validate, surface-sync, and operation-evidence checks
  for the scoped proof; six-lane evidence inspection; `go vet`, build,
  formatter, lint/docs and generator gates as applicable; `git diff --check`;
  fresh final-SHA audit; GitHub API read-back of the PR base.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Retained schema-v3 source is admitted without changing its identity | live | A source-import test and the scoped Outreach command assert the exact retained bytes/digest and an operation count; the previous importer rejection is reproduced first. |
| Import/projection works for a non-Outreach schema-v3 input | live | A minimal independent schema-v3 fixture must project a distinct source operation with its origin/citation; without generic reader support it fails before a projector result exists. |
| Unsupported shapes remain visible and truthful in every lane | live | Projected operation evidence asserts all six classifications and `missing_foundation` rows rather than omission or `implemented` promotion. |
| Source provenance is not a credential/certification gate | live | Tests reject malformed/unsupported source kinds and identity mismatches while asserting a valid retained hash alone never selects an executor or certification state. |
| No generic endpoint/command escape is introduced | live | Existing source-import/validation tests and final changed-path audit show execution remains constrained to canonical capability mappings. |
| A materialized command reaches credential preflight | fake | Expected usable-surface delta is `0`: this foundation only admits/projections evidence and must not materialize a new command. Final inspection must either replace this row with binary preflight evidence or retain this precise no-command reason. |

## GSD and required skills

- Resolved/reviewed: `scripts/gsd doctor`, `scripts/gsd sources` and generated
  prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` is
  green before implementation.
- Inline/manual fallback is required by the canonical single-worker contract;
  see `DISCUSSION-LOG.md`.
- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-documentation`, and `golang-lint`.
- CLI help/manual/website parity: `connectorgen` source-import is a developer
  command. Its help and migration docs are in scope only if their behavior or
  output changes; the user-facing `pm` command surface, `docs/cli/**`, and
  `website/**` are otherwise not applicable. Record the final inspection.

## TDD plan

1. **Red — schema-v3 admission:** add a focused importer test based on the
   retained Outreach lock that asserts the current reader rejects or discards
   the source before projection, then record the exact failure. Add an
   independent minimal schema-v3 fixture that exercises the same failure.
2. **Green — source reader/importer:** change the smallest common reader and
   importer contract needed to recognize the source form and preserve raw-byte
   identity, source URLs, descriptor origin, and operation inventory. Reject
   malformed or unsupported kinds explicitly and retain all existing limits.
3. **Red/green — projection/evidence:** add an observable source-to-descriptor
   to six-lane mapping test. The test must demonstrate an executor-shaped gap
   becomes a cited `missing_foundation` row rather than being dropped or marked
   implemented. Implement the minimal common projection/evidence update.
4. **Refactor:** sort and defensively copy output where needed; retain
   error-context and trust-boundary checks. Regenerate only canonical derived
   artifacts that report the newly admitted proof.
5. **Verification/review:** run scoped importer, projection, validation,
   surface-sync, operation-evidence, build, vet, documentation and generator
   checks. Audit the exact final SHA independently, run code review inline,
   commit/push coherent checkpoints, open the PR, and read its base from the
   GitHub API.

## Foundation gaps to preserve

The final evidence must name each operation shape identified by the source but
not supported by an executor as `missing_foundation`. It must distinguish that
state from malformed source, unavailable source, provenance-digest mismatch,
credential absence, and certification; none of those are a substitute reason.

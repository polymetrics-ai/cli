# Schema-v3 source projection/importer foundation — Outreach vertical proof

## Task Delivery Header

- Issue: Refs #4354 — feat(connectors): make Outreach full-surface pilot auditable
- Base branch: main (`origin/main` currently `b9b2478b3b2451d632d28b9aa138a170ad835110`)
- Merges into: main
- Delivery: Ordinary pull request open against `main`, with the final code SHA
  independently audited, normal non-force publication completed, and local
  scoped gates recorded.
- Working branch: feat/4354-schema-v3-source-projection
- Task: Add an explicit declaration-only source-reference path to the canonical
  importer/projector. It admits the retained Outreach operation lock and its
  canonical provider document URLs without pretending that unavailable raw
  bytes were imported. Preserve every cited source/operation identity, leave
  unavailable contract detail as `source_contract_unavailable` and executor
  shapes as `missing_foundation`, and prove the shared path with a second
  schema-v3 source case. Byte-backed importing remains strict; no
  connector-local generic executor or false `implemented` operation is
  allowed.
- Verification: focused red/green/refactor Go tests; `connectorgen`
  source-import/check, validate, surface-sync, and operation-evidence checks
  for the scoped proof; six-lane evidence inspection; `go vet`, build,
  formatter, lint/docs and generator gates as applicable; `git diff --check`;
  fresh final-SHA audit; GitHub API read-back of the PR base.
- PR API read-back: [#4358](https://github.com/polymetrics-ai/cli/pull/4358)
  is an ordinary non-draft PR with `base=main`
  (`b33983927d863032dac8220949990506e812937d`) and
  `head=feat/4354-schema-v3-source-projection`
  (`cdaf4849eee5da74998dc097eb60d6ba7d81b7cd`) at creation.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Byte-backed source importing preserves identity | live | Existing byte-backed import tests retain exact byte/digest checks. The new declaration-only path explicitly refuses to claim raw-byte import, credential validity, or certification. |
| Source-referenced projection works for Outreach and a non-Outreach input | live | Two distinct source-reference fixtures project cited operations through common code, preserving canonical source URLs and operation identities without network access. |
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

1. **Red — source-reference admission:** add a focused test based on the
   retained Outreach lock that demonstrates the current byte-backed reader
   rejects the cited-only source before projection, then record the failure.
   Add an independent minimal schema-v3 fixture that exercises the same
   declaration-only failure.
2. **Green — separate declaration reader/importer:** add the smallest common,
   explicit source-reference contract needed to preserve source URLs,
   descriptor origin, and operation inventory without raw bytes. Keep the
   byte-backed importer unchanged and strict; reject malformed/unsupported
   reference kinds and retain all existing limits.
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
state from malformed source, absent source-contract detail
(`source_contract_unavailable`), byte-backed provenance-digest mismatch,
credential absence, and certification; none of those are a substitute reason.

## Frozen independent-audit gap repair (2026-08-27)

The independent re-audit rejected exact head
`21480fcd9ce5701164bafb82666ffe5bbc3934c4`. This is one shared
connectorgen repair wave, not an Outreach connector change. The canonical
single-worker GSD fallback executes `discuss-phase` → `plan-phase --tdd` →
`execute-phase` → `verify-work` → `code-review` inline; the adapter prompts
were freshly resolved before this plan update.

1. **F1 — source-import to validate parity.** Add an end-to-end red test for
   both the retained legacy-v2 reference form and a true schema-v3
   `source_reference` document. Factor one expected-provenance constructor
   shared with the importer’s reference-form rules. A valid imported descriptor
   must validate as schema v3; byte-backed legacy descriptors remain schema v2.
2. **F2 — fail-closed operation-evidence reader.** Add a mutation matrix
   against ordinary byte-backed v2 locks for every reference-only field,
   including `source_kind: null`. Detect field *presence*, not merely decoded
   non-empty values, then reject before the tolerant legacy projection. Preserve
   the valid byte-backed v2 evidence reader.
3. **F3 — exact source location binding.** Bind every legacy reference
   operation to the primary document's declared location or its exact
   supplement's declared location. A two-operation primary/supplemental URL
   swap must fail even when all per-source counts still match.
4. **Green and integration.** Run the end-to-end and adversarial tests; then
   merge `origin/main` at `b9b2478b3b2451d632d28b9aa138a170ad835110` normally
   without reset, stash, force, or discarded work. Re-run the scoped tests on
   the merged tree before broader generator/engine/runner gates.
5. **Truth boundaries.** Preserve all 259 Outreach rows as
   `source_contract_unavailable` in all six lanes. Do not write an Outreach
   lock/descriptor, fabricate a command/request contract, claim credentials or
   certification, or increase usable surface (delta remains `0`).

## Independent-audit repair plan (2026-08-27)

The independent audit of `2738c6a9ff7172c74bedbcede092a77f16a05ba2` found
four correctness gaps. Treat them as one red/green wave:

1. Filter `source_contract_unavailable` descriptors before **all** source
   projection transforms, including blocked/reachable direct-read sets and
   path-flag restoration. A GET citation must leave existing CLI and API
   bytes untouched in write and check modes.
2. Replace the two-operation Outreach approximation with an immutable,
   byte-identical test fixture of the exact 259-row candidate lock. Exercise
   its real operation IDs, 253/6 primary/supplemental identity split, and the
   current-main Outreach API surface while producing no declaration changes.
3. Restore closed v1/v2 byte-backed wire decoding. The legacy cited-only
   discriminator receives its own wire type; every reference-only root, REST,
   and operation field is rejected on ordinary legacy locks.
4. Extract one closed reference-operation validator and apply it to both the
   legacy adapter and v3 `source_reference` documents: REST protocol, uppercase
   allow-listed HTTP method, normalized source ID/location, valid path, unique
   route/identity, and prohibited citation-field mixtures.

The repair remains declaration-only: no provider calls, no raw-byte rewrite,
no source-lock admission into production Outreach definitions, no generic
executor, and no executable command materialization.

# PLAN — #3857 declarative polling-watermark preflight

## Task Delivery Header

- Issue: Refs #3857 — declarative polling-watermark preflight (parent #3855)
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1` with its checks green; API read-back must confirm that exact base.
- Working branch: `fm/cli-3857-polling-watermark-preflight-r1`
- Task: Add the closed, definition-owned native polling watermark and target-apply declarations, a no-I/O runtime `PollingPreflight`, an eligibility projection through that same gate, and the #3810 legacy-name adapter without creating a CDC, REST, engine-specific, source-executor, or target-DML surface.
- Verification: exact happy/sad/edge fake assertions; engine/connectors/database/synccontract and app tests; `go vet`; build; repository gates; CLI parity probes; review; opened-PR base API read-back and green checks.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Correct implementation admits only registered source/apply executors and then sync can proceed | fake | #3857 owns no driver or source/apply executor and forbids live database calls. A guarded fake increments `reads`, `prepared`, and `emitted` exactly once only after the real preflight result. |
| Every declaration/runtime misconfiguration is refused before source I/O | fake | The no-live-database rule makes an in-process guarded fake necessary. Every subcase asserts the full refusal text and leaves both `reads` and `prepared` at zero. |
| Null, empty, precision-boundary, and corpus-admitted types retain/refuse their exact policies | fake | #3856’s immutable corpus is this issue’s executable contract while #3858/#3859 own source/apply execution. The fake exposes a separate empty-page counter, zero emitted records, and exact retained cursor values. |
| Bundle loading and public definition projection preserve the new declaration without an API/CDC claim | fake | No engine-specific declaration is in scope. A real loader over an in-memory bundle observes the declaration and a defensive-copy mutation; it never performs provider I/O. |
| The five legacy names use #3810’s compatibility mapping | fake | No public input boundary changes in this issue. A table-driven test reads `synccontract.PublicModes()` and proves each exact existing mapping is reused with no local aliases. |

## Scope

Add the shared native-database `polling_watermark.json` declaration, target-apply declaration,
registered executor registry, and no-I/O `PollingPreflight`. The runtime must
fail closed before source reads when the declaration, runtime executor, mode,
or immutable corpus registration is wrong.

Owned paths are limited to the shared database-definition and engine preflight
seams, their tests, and this issue's delivery evidence. No connector-specific
directory, native driver, commandrunner REST preflight, changefeed/CDC
capability derivation, query taxonomy, source transport, target DML, or public
CLI surface is in scope. The authoring documentation names the new separate file;
no existing engine declares or embeds it in this shared foundation slice.

## GSD lifecycle and skills

Passed: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase`,
`plan-phase`, `execute-phase`, `verify-work`, and `code-review`; and
`go run ./cmd/agentcontractgen check`. The generated `discuss-phase` and
`plan-phase --tdd` workflows are executed inline (documented in
`DISCUSSION-LOG.md`) because a compatible isolated Pi runtime is unavailable
and the canonical single-worker contract forbids role spawning.

Loaded: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-cli`, and `golang-documentation`.
The CLI help/docs/website parity reference was read. No command/flag/help/output
changes are planned; this remains explicitly not applicable unless the runtime
investigation proves otherwise.

## Declaration contract

The source declaration is closed and definition-owned. It names a concrete
native polling executor; a catalog object selector; a closed read/paging and
keyset-predicate dialect; maximum page and request bounds; snapshot/barrier;
lossless cursor codec/type/precision; the full ordering tuple; mutation,
commit-order, and bounded-overlap policy; source identity; schema compatibility;
delete visibility; and only #3810 canonical modes. The exact registered
source/apply executors carry the immutable-corpus evidence that the runtime checks.

The paired target declaration names a concrete native apply executor plus its
bounded batch, staging/replace, stable-key mapping, conditional ordering fence,
transaction/partial-result, and validity-window policies. A declaration that
would make a polling scan look like `change_capture`, advertise hard deletes,
or claim a mode unsupported by either side is refused. Descriptor modes remain
canonical `synccontract.Mode` values; the legacy adapter reads #3810's existing
five-name compatibility table rather than minting a second vocabulary.

## TDD slices

1. **Red — real runtime admission and happy path.** Add tests for a complete
   declaration and registered source/apply fakes. Show a resolved preflight
   result is necessary before a fake sync can increment its source-read
   counter; then make that counter and the emitted record observable.
2. **Red — exact pre-I/O refusals.** Parameterize absent source/apply executor,
   missing corpus registration, non-lossless codec, incomplete/non-unique
   ordering, unstable paging/keyset, unsafe mutation/commit/overlap, unsafe
   delete visibility, unsupported mode, incompatible target strategy, and
   history without transaction plus retry-safe close-and-insert. Every case
   asserts the specific error and a zero source-read counter.
3. **Red — corpus-derived edge domain.** Drive the declaration through the
   immutable v1 fixture classes: null watermark, empty page, timestamp
   precision boundary, large numeric tie-breaker, and every admitted cursor
   type. Assert the accepted values are retained exactly, rejected values name
   their policy, and none starts source I/O until admission succeeds.
4. **Green — minimal shared implementation.** Add schema/loading and defensive
   descriptor accessors, a concurrency-safe exact registry, and `PollingPreflight`.
   It returns copies only after source/apply/runtime/corpus/mode compatibility
   is proved. It does not execute I/O.
5. **Green — eligibility and sweep.** Provide the shared mode-eligibility
   projection from `PollingPreflight` and extend the existing implemented
   runtime sweep through that one runtime gate. Do not reproduce the decision
   tree in a generator or make an API-surface entry.
6. **Refactor/review.** Preserve defensive copies, nil safety, contextual
   errors, bounded values, and check that no owned exclusion was modified.

## Commit checkpoints

1. Plan/TDD/verification evidence.
2. Failing preflight tests (recorded red output).
3. Green implementation and focused test matrix.
4. Review/final verification evidence.

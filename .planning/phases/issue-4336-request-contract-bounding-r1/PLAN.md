# PLAN — request-contract execution envelopes

## Delivery contract

The task delivery header, decided policy, compatibility limits, and inline GSD
fallback are in `CONTEXT.md`. This plan implements the report's sections 4 and 5
as provider-neutral shared behavior and does not edit connector definitions.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Official Gong-shaped operation imports | fake | A hermetic artifact uses the exact blocked method/path/required `workspaceId` declaration and the real OpenAPI importer/resolver/descriptor path. Only network retrieval is replaced; source schema, operation identity, and envelope are asserted. |
| Provider schema is not rewritten | fake | The serialized descriptor's parameter schema remains exactly `{"type":"string"}` and has no injected `maxLength`; a separate execution envelope records PM policy/version/unit/default/hard/effective values. |
| Generation cannot omit the finite bound | fake | Descriptor and projection validation reject/sabotage an executable common input whose execution envelope is removed or made non-positive. |
| Runtime rejects cap+1 before I/O | real local HTTP | The actual operation executor observes zero test-server hits for an over-cap encoded query and one hit at the exact cap. The local server is the provider transport boundary, not a mocked validator. |
| Runtime reports PM provenance | real local HTTP | Error text identifies a PM execution limit and exact encoded-byte unit, never a provider validation or truncation. |
| Numeric lexemes stay exact | real local HTTP | Integers beyond `int64` and high-precision finite decimals pass type validation, retain the caller's exact wire bytes, and remain subject to the encoded byte cap. NaN/Inf syntax rejects before I/O. |
| Structural ambiguity stays blocked | fake | Dynamic objects, unsupported composition, and untyped arbitrary schemas still produce source-traced operation-local gaps; optional semantic maxima alone do not. |
| Existing projections do not drift | fake | A known bounded source document and currently generated bundle behavior remain byte-equivalent apart from the additive descriptor envelope. Existing projection defaults are frozen. |

## Required skills and lifecycle evidence

- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-naming`,
  `golang-documentation`, `github-issue-first-delivery`, and
  `gsd-ns-workflow`. Load `golang-lint` before code review.
- Passed before planning: `scripts/gsd doctor`; `scripts/gsd sources` for all
  five lifecycle commands; `go run ./cmd/agentcontractgen check`.
- Generated and executed inline: `discuss-phase` and `plan-phase --tdd`.
- No `no-mistakes` command or daemon interaction is authorized for this task.

## Implementation slices

1. **Typed disposition and descriptor envelope (TDD).** Add a typed schema
   disposition instead of matching error strings. For supported path/query
   common inputs, attach a versioned PM encoded-byte envelope while retaining
   the exact source schema. Leave dynamic/composed/untyped serialization as a
   source gap. Generation/projection checks must fail closed when an executable
   input lacks a positive envelope.
2. **Gong behavioral path (TDD).** Reproduce the exact
   `GET /v2/all-permission-profiles` `workspaceId` declaration through the real
   importer, assert no per-unit gap, assert the source schema is unchanged, and
   assert the PM envelope is present. Cover declared tighter bounds and the
   64 KiB hard ceiling without inventing source constraints.
3. **Runtime convergence (TDD).** Centralize exact finite scalar parsing for
   operation parameters using `big.Int`/`big.Rat`, preserve the caller lexeme,
   and update encoded path/query cap errors to identify PM execution policy and
   unit. Prove cap/cap+1, UTF-8 percent expansion, dangerous characters, enums,
   requiredness, and numeric exactness before I/O.
4. **Projection/surface provenance (TDD).** Ensure source-projected and
   params-imported executable parameter flags expose the effective PM byte cap
   in synthesized help/inspection. Preserve existing action schema and flag
   effective limits in this slice; record policy provenance in immutable source
   descriptors and generated error text without a fleet-wide JSON rewrite.
5. **Docs and compatibility.** Document provider constraints versus PM
   execution limits in migration conventions and generated surface semantics.
   Record why the header default is deferred. Run focused generation checks and
   CLI help/manual/website parity checks applicable to changed text/output.
6. **Verify and review.** Run the official execute/verify/code-review prompts
   inline, deliberate sabotage the envelope/enforcement path and prove the new
   tests fail, restore it, then run required bounded-concurrency local gates.

## TDD checkpoints

- Red: real importer refuses the Gong-shaped valid source at missing
  `maxLength`; new envelope assertions fail before production edits.
- Green: the source imports unchanged, the descriptor carries a positive PM
  envelope, and cap/cap+1 runtime tests prove pre-I/O enforcement.
- Red sabotage: after green, temporarily break the envelope or runtime cap and
  capture a focused test failure; restore immediately and rerun green.

## Commit and push checkpoints

- Planning/TDD contract.
- Red behavioral tests and trace.
- Green descriptor/classifier implementation.
- Green runtime/surface implementation.
- Documentation, verification, and review fixes.

## CLI help/manual/website parity

The effective `max_bytes` already appears in synthesized connector help and
inspection. Tests will freeze that behavior and PM-labelled errors; no new flag
or command is added. If generated text changes, run the repository generator and
commit all docs/manual/skill/website projections together. Required checks:
`pm help <connector>`, bare connector namespace, affected command `--help`,
docs/website grep, `make docs-check`, and surface-sync check.

## Out of scope / named dependency

PR #4339 owns operation-local quarantine for the four malformed paths and two
finite retained-media quota cases plus unrelated provider dialect prerequisites.
This direct PR neither stacks on nor copies that open work. After both land, the
real Batch-1 replay must measure document and executable-operation counts; the
inventory cannot prove an exact command count in advance.

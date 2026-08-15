# Plan — Issue 4095: PostgreSQL CDC delete binding

## Goal

Bind explicit PostgreSQL `pgoutput` delete events to the existing sealed
keyed-apply path. The binding must preserve source keys until
`MappingContractV1` maps them, then use the existing native history close
implementation. No missing record may be interpreted as deletion.

## GSD and skills

Manual inline GSD execution is recorded because this direct issue phase is
non-numeric and the single-worker contract forbids compatible role spawning.
Resolved command path: `discuss-phase` → `plan-phase --tdd` → `execute-phase`
→ `verify-work` → `code-review` via `scripts/gsd prompt`.

Loaded skills: `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, and
`golang-database`. CLI help/manual/website parity is not applicable: no public
CLI, flag, bundle capability, or documentation surface changes.

## TDD slices

1. **Red — source delete envelope.** Add a test for the absent PostgreSQL CDC
   delete conversion: it must derive a deterministic explicit tombstone from
   the emitted delete record and reject all non-delete/missing-key inputs.
2. **Green — shared key mapping.** Give `MappingContractV1` one sealed
   tombstone-key projection and route existing immutable-workset tombstones
   through it. It must retain only the configured key subset and rename every
   source key to the exact declared target key.
3. **Red — live delete behavior.** Replace synthetic manual delete envelopes
   in the existing tagged PostgreSQL tests with the CDC-derived source
   tombstone. Assert source absence leaves a physical target row intact, and
   assert the same explicit tombstone closes (not deletes) the history row.
4. **Green — narrow binding.** Implement only the native PostgreSQL CDC
   conversion and shared mapping seam required by the tests. Reuse existing
   plan/preview/approval/receipt and native write behavior; do not add a
   generic change-capture destination or alter #4154's boundary work.
5. **Verify and review.** Run the prescribed unit and tagged live test
   commands, relevant no-build gates, an inline focused review, then rebase,
   push, open the explicit-base PR, and read its base through the API.

## Guardrails

- One connector scope: PostgreSQL. Production changes are limited to its CDC
  adapter and the existing generic mapping contract consumed by it.
- No credentials, arbitrary SQL/HTTP/shell writes, dependencies, capabilities,
  generated surface rewrites, or changes to #4125, #4136, #4090, or #4154.
- A failure converting or mapping a delete happens before the write executor;
  tests must observe no target mutation for every refusal.
- Live evidence reads target rows and validity columns; an error-free method
  call is not evidence.

## Checkpoints

1. Commit this plan and red-test evidence.
2. Commit the green conversion/mapping implementation and focused tests.
3. Commit live-proof and review evidence, rebase, push, open PR, and verify
   `integration/4015-mvp-flat-r1` with the GitHub API.

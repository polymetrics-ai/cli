# Plan — destructive-write confirmation foundation

## GSD and skills

- GSD path: `scripts/gsd prompt plan-phase delete-confirmation-foundation-r1 --skip-research --tdd`.
- Programming-loop adapter note: `scripts/gsd prompt programming-loop ...` is absent from the
  healthy registry, so the approved manual programming-loop helper is used and recorded.
- Execution mode: `local_critical_path`; this is one tightly coupled safety state machine and the
  user assigned a single autonomous crewmate rather than parallel workers.
- Loaded skills: `gsd-programming-loop`, `golang-how-to`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-testing`, `golang-concurrency`, `golang-context`, `golang-lint`, `golang-cli`,
  `golang-documentation`, and `no-mistakes`.

## Tasks

- [x] task:tdd-plan-schema type:tdd — RED then GREEN the closed confirmation schema and normalized destructive target policy for writes and future operations.
- [x] task:tdd-preview type:tdd — RED then GREEN mandatory persisted preview digest before a destructive plan becomes approvable/executable.
- [x] task:tdd-approval type:tdd — RED then GREEN typed approval evidence and closed confirmation parsing at the engine/app boundary.
- [x] task:tdd-execute-seam type:tdd — RED then GREEN the generic pre-dispatch wrapper and demonstrate a future `rest_write` executor closure can use it unchanged.
- [x] task:tdd-bypass type:tdd — RED then GREEN direct-engine, connector-command, generic reverse-ETL, state-tamper, digest-mismatch, and bulk bypass resistance while preserving `batchable:false`.
- [x] task:tdd-authenticated-grant-hardening type:tdd — RED then GREEN an external-key authenticated, target-bound, expiring, atomically consumed execution grant; canonical prepared-request digests; native SQS gate adoption; shared previewability; and real app approval fixtures.
- [x] task:tdd-trusted-input-hardening type:tdd — RED then GREEN App-owned vault authority, authenticated plan lifetime, configuration/batchability binding, CAS-protected whole-state writes, and monotonic project consumption; complete the trusted-input sweep.
- [x] task:tdd-preview-execution-identity-review type:tdd — RED then GREEN App-private production authority, monotonic fixture grants, redirect refusal, hook-identical native preparation, and unsigned-plan expiry.
- [x] task:docs-verification type:docs — update lifecycle help/docs only where observable behavior changes, complete phase evidence, run separated local gates, and record 623/171/452 coverage.

## Path guard

Production changes stay in the engine write path, shared write request types, approval flow,
schemas, CLI lifecycle plumbing required by the changed flow, and tests. Captain decision
`defs-delete-fixtures` approved test-only edits to the Asana and Zendesk Support reverse-ETL
execution fixtures so their existing DELETE cases supply real preview-bound evidence. No bundle
JSON or connector runtime definition is modified; documentation housekeeping may correct stale
connector prose.

## Refactor gate

Only after targeted tests pass: consolidate confirmation normalization, preview hashing, and gate
errors; keep the executor-facing API small and provider-neutral; then run broader verification.

## CLI parity

- Runtime `pm reverse` and canonical provider-command help must state that destructive approval is
  issued only after preview.
- Bare namespaces and invalid-action behavior do not change.
- Update generated manual/skills/website outputs only if the existing generators detect drift.
- Verify with `pm help reverse`, `pm reverse`, provider `--help`, and docs/website greps.

# DISCUSSION LOG — issue #3995 shared connector-certification Shepherd gate

## Fixed inputs accepted from dispatch

- The child owns the shared, machine-readable gate only; it must not absorb #3989 proof-schema,
  #3991 live-provider, connector, or transport work.
- Gate enforcement is required before `integrate_sub_pr`, `accepted`, `ready_parent`, and
  `human_ready` transitions. The gate is a policy boundary even where a named transition is owned
  by a parent workflow rather than the existing local state machine.
- Inputs are the #3984 generated capability, flow, status, and evidence artifacts. The evaluator
  is pure/read-only and has no provider-action surface.
- The first RED uses current or temporary generated inputs for connector `github` and expects a
  deterministic `RETRY` containing
  `capability/github/capability:check/live_evidence`; an all-green fixture expects `PROCEED`.
- The delivery must use generated Claude, Codex, Pi, and OpenCode projections from a canonical
  registry. No harness gets adapter-local fields that could omit or weaken the gate.

## Decisions made during discussion

- Gate input and verdict schemas are explicitly versioned in the canonical delivery contract.
  The evaluator uses strict JSON decoding (`DisallowUnknownFields` plus trailing-data rejection)
  for the generated artifacts and evidence records.
- `RETRY` means a structurally valid generated artifact proves that a binding is not yet certified;
  `HALT` means the contract/input/projection is unsafe to trust (unknown schema, malformed data,
  missing required gate field, duplicate identity, or an evidence pointer that cannot be verified).
- Failure identifiers are stable, machine-readable coordinates. Criterion failures name the exact
  capability/workflow/sync/flow cell and criterion; evidence-record faults name the evidence record
  and preserve the referring cell in the verdict.
- Evidence sidecars are optional only when no cell claims live evidence. A claimed pointer without
  a strict, matching sidecar halts; an absent pointer is a retryable `live_evidence` criterion.
- Existing generic `agentcontractgen check` validates contract/projection drift only. It must not
  evaluate every connector, since the baseline deliberately contains zero certifications.

## Required skills used

- `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and
  `gsd-code-review` — issue-first lifecycle and recorded manual fallback.
- `github-issue-first-delivery` and `no-mistakes` — child/parent topology, delivery record, and
  final child-local validation without `--yes`.
- `cc-skills-golang:golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`,
  `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-lint` —
  strict Go API, deterministic tests, safe file handling, and static checks.

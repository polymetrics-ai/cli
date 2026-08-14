# Context — Issue 4001: shared connector failure classification

## Decision record

- This is a shared, dependency-free foundation for native database connectors and API destinations.
- The contract lives outside `internal/connectors/engine`, `commandrunner`, and certification so none
  of those consumers owns the vocabulary or introduces an import cycle.
- `internal/synccontract.RecoveryOutcome` remains CDC-only: it names durable checkpoint recovery
  outcomes, whereas this contract classifies configuration, system, and transient failures across
  connector execution and certification.
- The database hardening implementation remains owned by #3998. Main does not yet contain its
  typed database configuration boundary, so this issue adds a generic configuration-boundary
  compatibility test instead of editing a PostgreSQL driver or pre-empting #3998.

## Contract decisions

1. `Domain` is a closed JSON code: `configuration`, `system`, or `transient`.
2. Only `transient` is retryable. In particular, a configuration classification is non-retryable
   by its type method, not by caller convention.
3. Dispatch failures use an optional closed `DispatchKind`, limited to
   `direct_stub`, `helper_delegated_refusal`, `wrapped_typed_unsupported`,
   `declared_but_unroutable_command`, and `unresolved_dynamic_target`. A dispatch kind is valid
   only for a system classification.
4. A classification carries a stable reason code, user-facing message, optional JSON-Pointer field
   path, and a bounded typed reference list. The internal Go cause remains available through
   `Unwrap`/`Cause` but never serializes.
5. The safe reference list is identifiers only (source, connector, command, operation, capability),
   never credential values, request values, or provider response bodies.

## GSD execution fallback

`scripts/gsd doctor`, all five required command-source resolutions, and
`go run ./cmd/agentcontractgen check` passed. #4001 is an issue foundation rather than a numbered
roadmap phase, and this non-Pi worker cannot run the project-local Pi workers. The GSD
discuss/plan/execute/verify/review prompts were resolved and are executed inline. This directory
is the resulting discussion, plan, TDD, verification, and review evidence.

## Skills used

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, and `golang-testing`.

## Non-applicable parity

No connector command, flag, help topic, generated manual, or website surface changes in this
foundation. CLI help/manual/website parity is therefore not applicable.

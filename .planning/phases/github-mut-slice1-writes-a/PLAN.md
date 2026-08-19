# GitHub mutation certification: slice 1 writes-a

## Task Delivery Header

- Issue: Refs #4015 — GitHub mutation certification slice 1 (writes-a).
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1 → main`.
- Delivery: A direct PR from `fm/cli-mut-slice1-writes-a` with committed, schema-validated live certification evidence and this phase's verification record.
- Working branch: `fm/cli-mut-slice1-writes-a`.
- Task: Execute the 274 fixed GitHub GraphQL direct-write commands listed in the assigned slice, one at a time, within the captain-authorized disposable boundary. Every certified command must have a provider-observed state assertion and independently proven cleanup.
- Verification: `go run ./cmd/connectorgen certification-matrix --check`; targeted connector CLI/engine tests; `git diff --check`; review of the PR's API-reported base.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Each attempted command has one honest classification | live | Per-command receipt records the fixed operation, provider response class, and classification; a local preflight error is not recorded as provider absence. |
| A certified mutation changed provider state | live | A declared or agent-derived read-back proves the specific produced state and records a plausible rejected wrong value. |
| Certified mutations leave no disposable residue | live | Direct GitHub provider DELETE followed by an independent 404/empty read-back records `verified_absent`. |
| Published evidence remains valid | live | `certification-matrix --check` accepts every committed evidence record without regenerating shared artifacts. |

## Scope and safety

- Target only `Polymetrics-Cert`, `Polymetrics-Cert/pm-cert-3993-20260810-wz0fru`, and `polymetrics-ai-certification`.
- Stop only for real money, real people outside the two disposable identities, public visibility, or a third-party target.
- Resolve the Classic PAT only at point of use from the macOS keychain; do not print, persist, or place it in an argument. Plan-minted, single-use approval tokens are not GitHub credentials and may use the direct command's `--approve` channel under the captain decision.
- Do not regenerate capability, flow, or status matrix artifacts in this lane.

## Inline GSD fallback

This is a non-interactive certification lane with an explicit autonomous brief. I resolved and generated prompts for `discuss-phase --auto`, `plan-phase --tdd --skip-research`, `execute-phase --interactive`, `verify-work`, and `code-review`. Compatible isolated GSD workers are unavailable and this task forbids waiting, so the lifecycle is executed inline and evidenced in this directory.

Required skills loaded for this connector/CLI/GraphQL lane: `github-issue-first-delivery`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-graphql`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, and `gsd-code-review`.

## TDD ledger

### Red

- Before a command is marked certified, its agent-derived assertion must be expressed as a provider read-back predicate that would reject a plausible wrong value.
- Before accepting any evidence file, run `go run ./cmd/connectorgen certification-matrix --check`; a malformed/proofless record is rejected and removed.

### Green

- A command becomes `certified` only after its plan/preview/approved run reaches GitHub, the state predicate is true, cleanup uses a direct provider DELETE where applicable, and the independent absence predicate is true.
- The evidence check accepts the committed evidence set without generated-artifact drift.

### Refactor

- No production connector code, generic write capability, or generated shared artifact is changed by this evidence-only lane.

## Execution sequence

1. Initialise an isolated local PM project and create/use the disposable GitHub credential without exposing its secret.
2. For every slice row, inspect its fixed GraphQL input schema and containment class; skip the four escape classes with `escape_needs_captain`.
3. Create a tagged fixture only when the mutation's schema and provider object model support a contained lifecycle. Plan, preview, self-approve, run, read back, direct-delete, and independently prove absence.
4. Persist only a schema-valid, non-secret evidence record immediately after each completed exchange; delete a rejected record.
5. At every 50 attempted commands, append one aggregate status line and commit/push the green batch.

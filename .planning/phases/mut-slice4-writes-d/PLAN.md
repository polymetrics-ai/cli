# GitHub mutation certification — slice 4 writes-d

## Task Delivery Header

- Issue: Refs #4015 — GitHub mutation certification slice 4.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A direct PR from `fm/cli-mut-slice4-writes-d` is open against the stated base with committed, schema-validated, live certification evidence.
- Working branch: `fm/cli-mut-slice4-writes-d`.
- Task: Exercise exactly the 145 commands in the supplied writes-d manifest serially. For each contained command, use `plan` → `preview` → token-approved `run`, assert a provider read-back, direct-provider cleanup, and a cleanup read-back. Commands outside the captain's boundary are recorded as `escape_needs_captain` and halt this slice.
- Verification: `go run ./cmd/connectorgen certification-matrix --check`, targeted CLI/connector tests as applicable, `scripts/verify-gsd-workflow`, and API read-back of the opened PR base.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A contained mutation is certified | live | Provider read-back shows the tagged mutation's stated effect; a deliberately plausible mismatching value would fail the asserted predicate. |
| A certified mutation is cleaned up | live | Direct GitHub API DELETE followed by a separate 404 or empty-collection read-back; the mutation CLI's success status is never used as evidence. |
| Every attempted command has one classification | live | Serial receipt/evidence maps every attempted manifest path to exactly one allowed status. |
| Uncontained effect is not executed | live | Manifest endpoint and captain ruling identify the escape; no provider mutation is issued. |

## GSD lifecycle

`scripts/gsd doctor`, all five `scripts/gsd sources` commands, and `go run ./cmd/agentcontractgen check` passed before planning. Generated prompts for `discuss-phase` and `plan-phase --tdd` were resolved. This is an inline/manual execution fallback: the certification is a single-worker serial external operation and compatible isolated GSD workers are unavailable in this runner. Required skills loaded: `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

## Serial execution plan

1. Build the local `pm` binary and inspect the exact GitHub mutation command contract.
2. For each manifest entry in order, first classify containment. Skip only `not_implemented` entries; stop on a captain escape.
3. For contained entries, create/reuse only a `pm-cert-` fixture in the approved org/repository/user boundary; run plan, preview, approved run, independent read-back, direct API cleanup and absence read-back.
4. Write only proof-bearing, redacted schema-valid evidence after both effect and cleanup are proven. Validate each record immediately. Do not regenerate shared matrices.
5. Run verification, review the diff, push the branch, open the direct PR, and API-read its base.


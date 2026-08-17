# Mutation-candidate lifecycle Slice 0 — context

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification.
- Base branch: `integration/4015-mvp-flat-r1` at `eba2658c5fd671a1eebfb71463cbe6a3045d3c65` before any edit.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A committed branch and open direct PR against the stated base, with local verification recorded and GitHub API base read-back confirmed.
- Working branch: `fm/cli-mutation-lifecycle-slice0-r1`.
- Task: Extend the landed generic certification-candidate projection to generate and classify every fixture-required `direct_write` and `reverse_etl` command. Classification must be declaration-derived, connector-owned where provider facts are needed, and fail closed. It must not execute a live mutation or record certification.
- Verification: Red/green unit tests in `./cmd/connectorgen`, connector-definition validation, candidate/sweep generation twice for byte stability, `connectorgen boundary`, the repository verification entry points, and a no-credential binary reachability check only where it cannot mutate.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| All fixture-required mutations project to candidates | live | A generator test counts 279 `direct_write` plus 577 `reverse_etl` rows from the current sweep and compares them with the generated candidate inventory. |
| Every mutation has one fail-closed escape classification | live | A test asserts the inventory has one classification per candidate and that the buckets sum to the generated mutation total. |
| Escape classifications are not permissive | live | Table-driven tests give synthetic paid-seat, outside-invitation, public-publication, third-party, contained, and unassessed declarations; only the contained declaration receives `contained`. |
| No live mutation is executed | fake | Slice 0 is intentionally generation and classification only. Tests load disk declarations and use no credential, HTTP transport, or reverse-ETL executor. |

## Locked decisions

- The safety test is containment, not reversibility. A destructive or irreversible verb does not itself escape the disposable boundary.
- The only escape categories are `real_money`, `real_people`, `public_visibility`, and `third_party_scope`; provider actions inside the disposable identity and organisation are `contained` when their declaration proves the target scope.
- An absent, malformed, duplicated, or unmatched classification is `unassessed`; it never defaults to `contained`.
- Connector-specific cohort/family identifiers and concrete classification evidence remain in `internal/connectors/defs/<connector>/`. Shared Go is generic and must pass `connectorgen boundary` without allowlist changes.
- Candidate identity, executor, command tokens, required flags, credential flag, input slots, and address derive from the declared surface and referenced operation/write declaration. A manual candidate is an exact-command override carrying a named reason, not a broad fallback.
- Slice 0 does not load credentials, make a provider request, create a fixture, run a reverse plan, preview, approve, execute, or publish evidence.

## Dependency and manual lifecycle fallback

The dependency PR #4214 owns the generic `connectorgen certification-candidates` projection for `direct_read`. It was still open when this context was created. Its landed implementation is required before this slice can extend the projection; this branch must not fork or parallel-reimplement that contract.

`scripts/gsd doctor`, command-source resolution, `agentcontractgen check`, and all five required generated GSD prompts were run. The direct-PR contract forbids planner/reviewer role spawning and this task is not a numbered roadmap phase, so the documented inline/manual GSD fallback applies. This context, PLAN, TDD ledger, run state, verification, and review record are the durable lifecycle evidence.

## Required skills

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, and `golang-documentation`.

## CLI help/manual/website parity

The expected change extends the internal `connectorgen` developer generator, not the user-facing `pm` command surface. Runtime `pm` help, CLI manual, website docs, and public command discovery are not applicable. The generator’s own command help and tests remain in scope and will be checked before delivery.

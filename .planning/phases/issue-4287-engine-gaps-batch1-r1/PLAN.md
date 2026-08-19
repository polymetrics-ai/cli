# Issue #4287: Engine batch-1 declaration gaps

## Task Delivery Header

- Issue: Refs #4287 — feat(engine): close batch-1 connector declaration capability gaps
- Base branch: `main`
- Merges into: `main`
- Delivery: Pull request open against `main` with its reported base read back through the GitHub API after local gates pass.
- Working branch: `fm/cli-engine-gaps-batch1-r1`
- Task: Close the five evidence-backed declarative engine gaps without connector-name branching or generic execution escapes. The Docker Hub sweep lane must be able to reclassify its five affected operations without changing its disable rules.
- Verification: Focused red/green tests for every capability and fail-closed path; changed-package tests; `go vet ./...`; build; `connectorgen validate`; `surface-sync --check`; connector boundary; lint and docs gates; `make verify`; Docker Hub branch reclassification evidence when the owning sweep lane is available; deep source review.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A REST operation uses its own declared paginator rather than a global next-URL paginator. | live | A local provider double observes declared page/size and start-index/count queries, while the global paginator remains unchanged for an operation without an override. |
| Pagination/source disagreement is refused before transport I/O. | live | A provider double request counter remains zero after preflight rejects the conflicting declaration. |
| HEAD status operations are typed and response-less. | live | A HEAD-only provider double observes the method and the result contains status metadata rather than decoded JSON records; invalid HEAD shape sends zero requests. |
| Additional JSON-family write media types are closed. | live | SCIM media type serializes and sends once; an unrecognized type is rejected with the request counter at zero. |
| Text export is bounded. | live | A declared cap limits streamed export, while an omitted/zero cap is refused before the provider double sees a request. |
| Secret response handling cannot leak output, logs, or warehouse records. | live | A canary response routed to a fake keychain is absent from captured output/log/warehouse sinks and present only at the keychain seam; any unsafe output policy is refused with zero sends. |
| Docker Hub can flip all five affected operations without a disable-rule edit. | fake | The definition/disposition branch belongs to the parallel sweep worker. This foundation branch will run its provided regeneration/check command against that branch once firstmate reports it ready; it will not modify Docker Hub’s definition. |

## Foundation Check

| Need | Existing evidence | Plan |
| --- | --- | --- |
| Direct-read executor boundary | Global bundle paginator and direct-read executor already exist. | Add an operation-scoped typed override that preserves the global fallback. |
| Response-less status boundary | `rest_read` currently admits only GET and POST. | Introduce a distinct typed status executor and validate its closed contract. |
| Write media validation | Direct writes already accept JSON and form data. | Extend only the finite documented JSON-family set. |
| Bounded streaming | Binary download has cap and safe root semantics. | Reuse bounded streaming primitives under a separate text-export contract. |
| Sensitive policy validation | Schema and validator already require non-inline input and approval. | Connect the existing policy to a secret-safe execution and keychain-only response destination. |

## TDD Slices

1. Per-operation pagination: introduce failing operation-override, disagreement, and global-fallback tests; then implement the narrow declarative override.
2. Typed status operation: add failing HEAD happy, invalid-shape, and no-JSON confusion tests; then implement the separate executor.
3. Closed write media types: add SCIM happy, unknown-type refusal, and existing JSON/form regression tests; then implement the allowlist.
4. Bounded text export: add cap happy, omitted-cap refusal, and over-cap cleanup tests; then implement the contract.
5. Sensitive live response: add keychain-only happy, unsafe-output preflight refusal, and no-leak sink tests; then connect the live policy and response destination.

## Lifecycle and Skills

- GSD commands resolved and generated inline: `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`.
- Inline/manual fallback reason: issue #4287 is absent from `.planning/ROADMAP.md` and this firstmate task forbids role spawning.
- Loaded skills: `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-cli`, `golang-documentation`, and `golang-lint`.

## Commit Checkpoints

1. Planning plus recorded red evidence.
2. Pagination green slice.
3. Status/media/text green slice.
4. Secret-response green slice.
5. Full verification and review fixes.

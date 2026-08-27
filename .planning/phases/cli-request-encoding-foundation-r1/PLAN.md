# Plan — source-backed request encoding foundation

## Required skills and lifecycle evidence

- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, and
  `golang-documentation`.
- Passed before planning: `scripts/gsd doctor`; `scripts/gsd sources` for
  discuss/plan/execute/verify/review; `go run ./cmd/agentcontractgen check`.
- Generated and executed inline: `discuss-phase 4367` and
  `plan-phase 4367 --tdd`. Inline execution is required by the single-worker
  brief; no GSD role is spawned.

## Source census and scope guard

The source-derived frozen fixture must assert 51 identities and exact citations:
47 GitLab operations at the pinned GitLab OpenAPI URL, two Sentry release-file
uploads at Sentry's dereferenced OpenAPI URL, Asana `createAttachmentForObject`,
and Jira `addAttachment`. Every row asserts source ID, provider operation ID,
method, path, URL, source location, selected media type, and expected primary
foundation. The fixture must reimport source-shaped documents; it must not read
or copy #4364's future manifest.

## TDD slices

1. **Red — source descriptor encoder disposition.** Add source-import tests
   that make the exact form/multipart cohort fail on the current universal
   request-encoding foundation; assert malformed encoder metadata, citation or
   source-identity mismatch, content-type mismatch, and an unrelated typed
   schema gap. Add the 51-row source-derived reconciliation assertion.
2. **Green — closed typed descriptor/projection admission.** Represent one
   selected JSON, urlencoded, or multipart encoder from the descriptor and
   propagate only its source-owned body/part metadata. Preserve encoder and
   schema gaps independently; no method-only lane selection or raw-body path.
3. **Red — engine request shapes.** Add no-network transport-spy tests for
   multipart text plus binary parts, form required/missing/duplicate/unknown
   fields, URL-encoded serialization styles, mismatch refusal, and request
   byte bounds. The tests initially fail at the new closed contract.
4. **Green — builder convergence.** Reuse the bounded multipart and prepared
   form request paths. Validate fields before transport and prove the spy sees
   the exact declaration-owned method, path, content type, body, part names,
   file semantics, and zero calls for invalid input.
5. **Refactor and generated-contract checks.** Make order and validation
   deterministic, retain source provenance, update source projection tests,
   run per-connector regeneration/checks without adopting provider commands,
   and report the measured promotion/residual counts by primary reason.
6. **Verify and frozen review.** Execute inline verify/review prompts, perform
   a deliberate temporary encoder-gate sabotage with focused red evidence,
   restore it, then run all direct-PR local gates and a fresh exact-SHA review.

## Commit/push checkpoints

1. Planning/TDD evidence.
2. Red source/reconciliation and runtime tests.
3. Green closed descriptor/projection and request builder behavior.
4. Refactor, generated checks, verification, and review fixes.

## CLI help/manual/website parity

This foundation must not create a provider command or caller-selectable flag.
Run `pm help <affected connector>`, bare namespace, and representative command
help if any generated surface changes. Otherwise record the explicit no-surface
result, run `make docs-check`, check docs/website generated parity, and leave
provider command/manual adoption to its owning lane.


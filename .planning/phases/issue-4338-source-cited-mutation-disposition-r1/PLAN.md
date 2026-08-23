# Plan — issue 4338 source-cited non-executable mutation disposition

## Task Delivery Header

- Issue: Refs #4338 — fix(connectorgen): add source-cited mutation dispositions
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, committed on
  `fm/cli-mutation-disposition-foundation-r1`, with scoped local gates green and
  API-reported base equal to `main`.
- Working branch: fm/cli-mutation-disposition-foundation-r1
- Task: Add a closed per-source-operation disposition for provider mutations
  that lack a complete executable action. It must retain a provider citation
  and source-traced runtime gap; it must not make any action or command look
  implemented or fold mutation coverage into read-only coverage.
- Verification: red/green behavioral tests through source projection and
  executable-coverage validation for separate Asana absent-action, Jira
  incomplete-contract, Sentry SCIM PATCH/dashboard POST, and Vercel-sized
  mutation-batch fixtures; complete-action, read-only-boundary, byte-stability,
  generated-artifact, lint, and repository gates.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| An absent-action mutation disposition retains a source-traced runtime gap | live | The real projection/validation path accepts an Asana-shaped locked mutation only when its cited disposition emits the named gap, while no action or CLI command becomes enabled. |
| An incomplete-action mutation disposition retains a source-traced runtime gap | live | The real projection/validation path accepts a separate Jira-shaped mutation only when the incomplete contract is explicitly disposed and keeps the provider citation/gap. |
| Complete mutations remain executable | live | A complete action/implemented command validates normally and rejects the mutation disposition. |
| Read-only cannot hide a mutation | live | POST, PUT, PATCH, and DELETE fixtures reject the read-only route and retain mutation validation failure without this disposition. |
| Existing projection stays stable | live | A pre-existing connector projection renders byte-identically when it has no disposition. |
| Four-connector scale stays per-operation | live | Asana, Jira, Sentry, and a Vercel-sized batch pass identical source-projection and executable-coverage paths, each retaining a citation/gap and creating no action. |

## TDD execution slices

1. **Red:** add separate behavioral real-path tests for Asana absent actions and
   Jira incomplete action contracts, plus complete-action, read-only rejection,
   and byte-stability controls. Run the focused command and record its failures.
2. **Green:** add the smallest source-cited mutation disposition model and the
   shared projection/coverage checks. It must fail closed outside a mutation or
   when a complete executable action exists.
3. **Refactor:** retain existing helper names and control output; no shared-file
   drive-by rewrite. Add the Sentry SCIM PATCH/dashboard POST and Vercel-sized
   source-cited mutation fixtures; re-run focused tests and generated/snapshot
   checks.
4. **Verification/review:** execute the GSD verify and inline review prompts,
   then repository-required gates individually under the shared 20-minute
   budget. Merge `origin/main` before push; re-run affected validation if the
   sibling read-only work lands.

## CLI help/docs/website parity

`connectorgen` internal projection behavior changes but no `pm` command,
connector command surface, help topic, manual, or website content changes.
Runtime help/docs/website parity is not applicable; generated connector/website
artifact checks remain required.

## Skills and lifecycle

Loaded: `golang-how-to`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`, and
`golang-structs-interfaces`.

Resolved command path: `scripts/gsd doctor`; `scripts/gsd sources` and
`scripts/gsd prompt` for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review`. Generated prompts are executed inline because
the compatible isolated GSD role runtime is unavailable.

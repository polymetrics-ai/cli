# Plan — issue 4336 provider dialect tolerance

## Task Delivery Header

- Issue: Refs #4336 — fix(connectorgen): tolerate bounded provider OpenAPI dialect gaps
- Base branch: main
- Merges into: main
- Delivery: Pull request open against `main`, committed on
  `fm/cli-provider-dialect-tolerance-foundation-r1`, with local verification
  green and API-confirmed base `main`.
- Working branch: fm/cli-provider-dialect-tolerance-foundation-r1
- Task: Preserve all affected source operations while supporting legitimate
  OpenAPI dialect syntax, retaining malformed provider contracts as explicit
  source-traced gaps, and keeping every resource guard finite.
- Verification: Red then green behavioral tests using each provider's real
  document through the importer; byte projection control; pathological-depth
  refusal; changed-package, generator/snapshot, static, and review checks.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Bitbucket, Notion, and Stripe retain their affected operations within a raised finite bound | live | Importer output contains each real provider operation and a measured finite guard still rejects a deeper constructed document. |
| Vercel `patternProperties` and Docker Hub `example` are handled correctly | live | Each real importer result retains the declared schema keyword and passes the relevant generic compiler/descriptor path. |
| Docker Hub dangling response ref and GitLab malformed path survive as source gaps | live | The operation remains, is merge-blocked, and its gap names both source pointer and response/path location. |
| Existing importer output is unchanged | live | A checked-in importable connector projection is byte-identical before and after the change. |
| No connector-local change lands | live | Changed paths exclude `internal/connectors/defs/`; source importer behavior is generic. |

## TDD execution slices

1. **Measure and red:** add real-document importer-path cases for the seven
   refusal sites plus byte-identical and pathological controls. Run them before
   production code and record each expected old failure and observed legitimate
   depth/reference chain.
2. **Green — finite bounds:** set only the minimum justified generic finite
   schema/reference limits; retain the existing pathological failure.
3. **Green — supported syntax:** implement bounded `patternProperties`
   representation and `example` annotation support without weakening unknown or
   dynamic-keyword validation.
4. **Green — retained malformed contracts:** convert only dangling local
   response references and missing required path parameters into the existing
   pointer/location source-gap mechanism; all other malformed reference/path
   forms remain rejected.
5. **Refactor and prove:** canonicalize gap ordering, rerun real importer cases,
   confirm byte-identical control, generation/snapshot/static gates, and review.

## CLI help/docs/website parity

`connectorgen source-import` changes internal importer interpretation only. It
does not add or change `pm` commands, flags, help text, manuals, website docs,
or generated help. These surfaces are not applicable.

## Skills and lifecycle

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, and `golang-structs-interfaces`.

Resolved command path: `scripts/gsd doctor`; `scripts/gsd sources` and
`scripts/gsd prompt` for `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review`. Prompts execute inline under the documented
single-worker fallback.

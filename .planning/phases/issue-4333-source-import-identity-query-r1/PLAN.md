# PLAN — source-import identity-bearing artifact query

## Task Delivery Header

- Issue: Refs #4333 — preserve declared identity-bearing artifact queries.
- Base branch: main.
- Merges into: main.
- Delivery: Pull request open against `main` with local required gates passed
  and GitHub-reported base verified.
- Working branch: fm/cli-source-import-query-identity-r1.
- Task: Let a v3 source document explicitly mark its fixed artifact query as
  identity-bearing so source-import can retrieve it, while retaining the
  no-query default and every URL/SSRF guard.
- Verification: Red/green behavioral source-import tests; focused package and
  CLI tests; formatting, vet, build, documented generator checks, and the
  separate `make verify` gates required by AGENTS.md.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Declared identity query is fetched | fake | A hermetic HTTP transport replaces only the socket layer; the real v3 lock parser, importer, cache fetcher, URL validation, request construction, response reader, digest check, and descriptor projection run. The assertion observes `version` in the GET request and the imported operation. |
| Capture query stays provenance-only | fake | The hermetic importer records every request from a v3 lock with `published_source.source_url?slug=` and asserts only the queryless artifact URL is requested; a real public provider is intentionally not contacted. |
| Legacy/default behavior is unchanged | fake | Two real v3 lock documents differing only by absent versus explicit false declaration import to byte-identical descriptors, and existing frozen GitHub artifact checks remain green. |
| Identity-query attacks remain rejected | fake | Hermetic raw v3 locks exercise credential-shaped, oversized, and excessive-key queries plus each existing URL and resolver guard. Live providers cannot safely host malicious URLs. |

## Required skills and GSD evidence

- Loaded: `golang-how-to`, `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`, and
  `golang-documentation`.
- Passed: `scripts/gsd doctor`; sources resolved for `discuss-phase`,
  `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `go run
  ./cmd/agentcontractgen check`.
- Manual inline fallback: generated GSD prompts are executed inline because
  this issue is unnumbered and the single-worker contract forbids role
  spawning here.

## Implementation design

1. Add `identity_query` to the v3 artifact contract. Its absence/false value
   preserves the current prohibition; true requires a non-empty fixed query.
   Reject the declaration on legacy and GraphQL artifacts.
2. Extract the existing bounded citation-query validation into a shared,
   source-import-only helper and apply it to opt-in artifact queries. Retain
   the current credential-key rejection and all base artifact URL guards.
3. Add an internal URL policy to the batch URL parser/request validator. Its
   default remains no query. The source-import HTTP fetcher is the only
   identity-query caller and receives the declaration-bound artifact object;
   redirect checks retain the default no-query policy.
4. Preserve the artifact object through the cache-to-source fetch hop, so the
   underlying HTTP fetcher sees the declaration rather than inferring policy
   from a raw URL.
5. Update the source-lock convention and behavioral test coverage. No runtime
   CLI flag, help topic, generated command manual, or website page changes:
   this is a lock-schema/importer contract documented in migration
   conventions; CLI parity is explicitly not applicable beyond the existing
   source-import contract test.

## TDD sequence

1. Add a raw v3-lock behavioral test whose artifact has
   `identity_query:true` and `?version=`. Drive it through the parser,
   importer, cache fetcher, and HTTP request path; assert request query and
   imported operation. Run it red before production edits.
2. Add behavioral regressions for citation capture-query stripping,
   absent/false declaration byte-identical projection, bounds/credential
   rejections, and all existing URL/DNS guards. Record the red failure.
3. Implement the smallest declaration-bound policy flow and shared bounded
   query validator; format and rerun the focused tests green.
4. Update the conventions and source-import command-contract assertions, then
   run required local gates. Execute verify and security/code review inline;
   plan/execution gaps if any.

## Commit and push checkpoints

- Planning/TDD evidence checkpoint.
- Red behavioral-test evidence checkpoint.
- Green implementation and focused tests checkpoint.
- Documentation and review-fix checkpoint, if needed.

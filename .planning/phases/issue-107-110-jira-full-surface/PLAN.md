# Plan: Issues #107-#110 Jira Full Surface

Parent: #81 Jira CLI feature parity parent roadmap  
Issues: #107 operation ledger, #108 direct reads, #109 advanced/body-variable applicability, #110 sensitive/admin policy  
Branch: `feat/81-jira-cli-parity`  
Worker mode: local critical path (no Pi subagent tool exposed in this harness)

## Required skills loaded

- gsd-core
- golang-how-to
- golang-cli
- golang-testing
- golang-error-handling
- golang-security
- golang-safety
- golang-spf13-cobra
- golang-structs-interfaces
- golang-context
- golang-concurrency
- golang-graphql
- golang-documentation

## GSD status

- `scripts/gsd prompt plan-phase issue-107-110-jira-full-surface --json` generated a repo-local Pi adapter prompt.
- `scripts/gsd prompt programming-loop init --phase issue-107-110-jira-full-surface --dry-run` is unavailable in this adapter (`unknown GSD command: programming-loop`, observed earlier in this Jira parent session); manual GSD/TDD fallback remains active.

## Scope

Model the official Atlassian Jira Cloud OpenAPI surface from `https://developer.atlassian.com/cloud/jira/platform/swagger-v3.v3.json` without credentialed Jira calls:

- Inventory all official operations and verify the expected count/method split: 620 total; GET 276, POST 135, PUT 119, DELETE 90.
- Add a Jira operation ledger (`operations.json`) for all official operations.
- Expand `api_surface.json` into operation-ledger mode so every endpoint is exactly covered by a stream/direct read/write/binary policy or a blocked duplicate/deprecated/disallowed/product-scope classifier.
- Add safe direct-read command metadata for bounded JSON GET endpoints.
- Record #109 as REST/body-variable only: Jira Cloud OpenAPI has no GraphQL surface; body-variable complexity is represented through typed REST operation/write metadata, not raw body escape hatches.
- Add write-policy metadata for reverse-ETL candidates while keeping execution behind plan → preview → approval → execute and typed/destructive gates.

## TDD slices

1. Red: embedded Jira full-surface test fails while `operations.json` is absent/incomplete and `api_surface.json` is still the small baseline.
2. Red: validation test asserts all operation-ledger rows are blocked by default when not executable and all executable refs resolve.
3. Green: generate/import a deterministic OpenAPI-derived ledger with normalized operation IDs, method/path, risk, output policy, and source URLs.
4. Green: expand `api_surface.json`, direct-read CLI metadata, and write metadata from the same reviewed ledger without raw generic write/direct-write commands.
5. Green: add tests for counts, method split, direct-read boundedness, write policy, sensitive/admin/destructive gates, no GraphQL requirement, and no secret leakage in metadata.
6. Refactor: keep generation deterministic; avoid new dependencies; preserve existing Jira streams and CLI commands.
7. Verify: targeted tests, `connectorgen validate`, docs/website generation if metadata changes, then full local gates.

## Implementation summary

- Full-surface Jira ledger implemented and verified: 620 operations, 333 reverse-ETL writes, 268 generated direct reads, 3 existing streams, and 16 explicitly blocked binary/file-upload/rest-query executor gaps.
- #109 resolved as REST-only: Jira has no GraphQL surface in the official OpenAPI; added `rest_query` metadata for safe POST read-query operations that need fixed body-variable read execution.
- #110 policy metadata added through write risk/confirmation fields plus operation `sensitive_policy` where secret-shaped inputs/effects are detected.
- No live Jira credentials or network calls were used beyond fetching the public OpenAPI document.

## Safety gates

- No secrets, live credentials, or credentialed Jira checks.
- No new dependencies without approval.
- No raw generic HTTP/GraphQL write, generic shell write, or SQL write escape hatches.
- Reverse ETL execution remains plan → preview → approval → execute; admin/destructive/sensitive actions carry typed confirmation/redaction metadata.
- Direct reads are GET-only, JSON-oriented unless explicitly binary-bounded, and must not add unbounded output destinations.
- Binary operations require max-byte/local-output policy and remain blocked or direct-read metadata only until safe executor support is verified.
- Parent PR merge to `main` remains human-gated.

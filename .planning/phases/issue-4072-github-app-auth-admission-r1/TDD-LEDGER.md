# Issue #4072 TDD Ledger: GitHub App auth admission

**Issue:** #4072 — `fix(engine): gate GitHub App token minting through shared rate admission`

**Parent:** #3754

**Fresh correction lineage:** **0/5**

**Recovered base:** `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`

**Canonical private finish-plan snapshot SHA256:**
`939f14f61defd993f8ad0335a5eb617d97083c9f73a6a75259d0e312ae8f408`

**Required skills used:** `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
`golang-database`, `golang-graphql`, `github-issue-first-delivery`,
`no-mistakes`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`.

## Scope Ledger

| In scope | Explicitly excluded |
|---|---|
| Engine custom-auth declared route admission | PostgreSQL/warehouse behavior |
| GitHub App installation-token POST | Connector certification/Shepherd |
| Local fake transport/coordinator tests | polling/apply #4041 and UDS child |
| Preserve GitHub existing rate policy | provider credentials, mutation, or CLI change |

## Test Matrix

| ID | Behavior | RED expectation | GREEN expectation | Status |
|---|---|---|---|---|
| A1 | no shared coordinator | raw GitHub token POST reaches recording transport before typed refusal | typed `shared_coordinator_unavailable`, zero sends | planned |
| A2 | lost shared coordinator | raw token POST reaches recording transport before lost-coordinate refusal | typed `shared_coordinator_unavailable`, zero sends | planned |
| A3 | granting coordinator | no declared route admission for the token POST | one `Decide`, one POST, one `Finish` | planned |
| A4 | declared-vs-actual path | declaration is not passed to admission seam | declared `POST /app/installations/{installation_id}/access_tokens`; actual escaped installation path sent | planned |
| A5 | secret boundary | cannot prove coordinator payload absence | reservation/observation/error evidence has no JWT, key, or minted token | planned |
| A6 | regressions | N/A | bearer, write-hook, ordinary REST, local admission, and GraphQL exclusion remain green | planned |

## RED

- **Planned command:** `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission' -count=1`
- **Expected failure:** test detects at least one physical installation-token POST
  before `require_shared` can return the typed refusal.
- **Commit:** pending
- **Observed failure:** pending — do not proceed to GREEN until this is a causal behavioral failure.

## GREEN

- **Planned command:** `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1`
- **Expected result:** zero-send missing/lost shared coordinator cases; one
  decision/send/finish granting case; focused regression matrix passes.
- **Commit:** pending
- **Observed result:** pending

## REFACTOR

- **Status:** pending; only if a cleanup is needed after GREEN and the focused
  matrix remains green.

## Deferred Validation Gate

No broad suite, race-heavy sweep, no-mistakes, push, PR creation, CI, or merge
is authorized until Firstmate releases the shared #4069 validation lane.

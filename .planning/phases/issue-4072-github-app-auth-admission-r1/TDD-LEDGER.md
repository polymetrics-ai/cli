# Issue #4072 TDD Ledger: GitHub App auth admission

**Issue:** #4072 — `fix(engine): gate GitHub App token minting through shared rate admission`

**Parent:** #3754

**Fresh correction lineage:** **0/5**

**Recovered base:** `7eea99bae` (`integration/4015-mvp-flat-r1`, including
#4122 / #3754's resolved-path Requester admission boundary).

**Canonical private finish-plan snapshot SHA256:**
`939f14f61defd993f8ad0335a5eb617d97083c9f73a6a75259d0e312ae8f408`

**Required skills used:** `golang-how-to`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-error-handling`, `golang-security`,
`golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`,
`gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
`gsd-verify-work`, `gsd-code-review`.

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
| A1 | no shared coordinator | raw GitHub token POST reaches recording transport before typed refusal | typed `SharedRateLimitUnavailableError`, zero sends | planned |
| A2 | unreachable shared coordinator | raw GitHub token POST reaches recording transport before refusal | typed `SharedRateLimitUnavailableError`, zero sends | planned |
| A3 | real shared coordinator | no token admission reaches physical send boundary | two processes share one budget: one token POST and one exhausted-budget timeout | planned |
| A4 | declared-vs-actual path | declaration is not passed to physical send boundary | actual escaped installation path is admitted by `Requester` at send | planned |
| A5 | secret boundary | cannot prove coordinator payload absence | coordination key/error evidence has no JWT, key, or minted token | planned |
| A6 | regressions | N/A | bearer, write-hook, ordinary REST, local admission, and GraphQL exclusion remain green | planned |

## RED

- **Planned command:** `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission' -count=1`
- **Expected failure:** test detects at least one physical installation-token POST
  before `require_shared` can return the typed refusal.
- **Commit:** pending (recorded with the RED test slice)
- **Observed failure:** `TestGitHubAppAuthRateAdmissionRequireSharedRefusesBeforeTokenSend`
  observed one physical token POST and `NewRuntime error = <nil>` while the
  copied GitHub app-installation policy required a missing shared coordinator.
  This is the intended causal behavioral failure, not a build failure.

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

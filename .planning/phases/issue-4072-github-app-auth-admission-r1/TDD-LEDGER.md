# Issue #4072 TDD Ledger: GitHub App auth admission

**Issue:** #4072 — `fix(engine): gate GitHub App token minting through shared rate admission`

**Parent:** #3754

**Fresh correction lineage:** **1/5** (correction 1 reserved for the recorded
configured-linter failure; no no-mistakes correction run has started)

**Recovered base:** `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`

**Canonical private finish-plan snapshot SHA256:**
`939f14f61defd993f8ad0335a5aeb617d97083c9f73a6a75259d0e312ae8f408`

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
| A1 | no shared coordinator | raw GitHub token POST reaches recording transport before typed refusal | typed `shared_coordinator_unavailable`, zero sends | pass |
| A2 | lost shared coordinator | raw token POST reaches recording transport before lost-coordinate refusal | typed `shared_coordinator_unavailable`, zero sends | pass |
| A3 | granting coordinator | no declared route admission for the token POST | one `Decide`, one POST, one `Finish` | pass |
| A4 | declared-vs-actual path | declaration is not passed to admission seam | declared `POST /app/installations/{installation_id}/access_tokens`; actual escaped installation path sent | pass |
| A5 | secret boundary | cannot prove coordinator payload absence | reservation/observation/error evidence has no JWT, key, or minted token | pass |
| A6 | regressions | N/A | bearer, write-hook, ordinary REST, local admission, and GraphQL exclusion remain green | pass |

## RED

- **Command:** `go test ./internal/connectors/hooks/github -run '^TestGitHubAppAuthRateAdmissionRequireSharedRefusesBeforeTokenSend$' -count=1`
- **Expected failure:** test detects at least one physical installation-token POST
  before `require_shared` can return the typed refusal.
- **Observed failure:** 2026-08-12 — failed causally at recovered base:
  `physical GitHub App token sends = 1, want 0 before shared admission refusal
  (NewRuntime error = <nil>)`. The failure is behavioral: the direct
  `http.DefaultClient` token exchange reached the secret-blind recording
  transport before resolver construction, rather than failing to compile or
  failing on test setup.
- **Expanded RED command:** `go test ./internal/connectors/hooks/github -run '^TestGitHubAppAuthRateAdmission' -count=1`
- **Expanded observed failures:** 2026-08-12 — no/lost coordinator tests each
  observed one premature physical token transport send; granting-coordinator
  test observed zero decisions where one declaration-aware decision is required.
- **Commits:** `9a44c9163` records the causal no-coordinator RED;
  `3f20bf7ba` records the expanded no/lost/grant/privacy RED matrix.

## GREEN

- **Command:** `go test ./internal/connectors/engine ./internal/connectors/hooks/github -run 'TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubDeclaredRateLimits' -count=1`
- **Expected result:** zero-send missing/lost shared coordinator cases; one
  decision/send/finish granting case; focused regression matrix passes.
- **Observed result:** 2026-08-12 — pass (`engine` and `github` packages).
  The complete GitHub hook package also passes with
  `go test ./internal/connectors/hooks/github -count=1`.
- **Commit:** this focused GREEN commit, `fix(4072): admit GitHub App token minting`.

## REFACTOR

- **Status:** complete. The recording transport remains secret-blind while
  recording the method/path required to prove the engine-managed physical send;
  no functional refactor followed GREEN.

## POST-GREEN BOUNDED VALIDATION

- **Functional matrix:** `GOMAXPROCS=2 go test -p 1 ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1` — pass on 2026-08-12.
- **Race matrix:** `GOMAXPROCS=2 go test -p 1 -race ./internal/connectors/engine ./internal/connectors/hooks/github -run '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)' -count=1` — pass on 2026-08-12.
- **Static analysis:** `go vet ./internal/connectors/engine ./internal/connectors/hooks/github` — pass on 2026-08-12.
- **GSD fallback:** the named issue phase is not present in the numeric roadmap and the canonical single-worker contract disallows role spawning. `scripts/gsd prompt verify-work issue-4072-github-app-auth-admission-r1 --auto` and the equivalent deep `code-review` prompt were resolved; their inline evidence is `UAT.md` and `REVIEW.md`.

## CORRECTION 1/5 — CONFIGURED-LINTER RED

- **Red:** `make lint` at broad-acceptance plan head `414228f02` exited
  non-zero on 2026-08-12. It reported exactly two `unused` declarations:
  `buildAuthenticator` (`internal/connectors/engine/auth.go:71`) and
  `buildCustomAuth` (`internal/connectors/engine/auth.go:280`).
- **Cause:** both private forwarding wrappers became unreferenced when the
  declared-route variants replaced their sole call paths.
- **Why this is the causal RED:** the acceptance condition is removal of dead
  private source. A test that invokes the wrappers would mask the configured
  linter finding and retain unnecessary API surface; the failing linter is the
  executable regression proof.
- **Green plan:** remove only the two unused wrappers, run `make lint`, then
  rerun the report-defined broad acceptance matrix. This reserves one fresh
  child correction; it does not interact with no-mistakes or the parked #3754
  lineage.
- **Green:** removed only the two private wrappers. `GOMAXPROCS=2 go test -p 1
  -count=1 -timeout 20m ./internal/connectors/engine
  ./internal/connectors/hooks/github -run
  '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)'`
  passed, followed by `make lint` with `0 issues` on 2026-08-12.

## Deferred Validation Gate

No broad suite, race-heavy sweep, no-mistakes, push, PR creation, CI, or merge
is authorized until Firstmate releases the shared #4071 validation lane.

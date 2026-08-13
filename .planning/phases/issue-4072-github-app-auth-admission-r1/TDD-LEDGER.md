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

## GENERATED ARTIFACT RED/GREEN — INHERITED CAPABILITY-MATRIX DRIFT

- **Ownership and boundary:** #4072 owns this child-local generated-only
  synchronization because its recovered base already carries the stale source
  locations. #4026 / merged PostgreSQL PR #4034 is generator precedent only;
  no #4034 ancestry, generator/source change, or PostgreSQL semantic change is
  imported.
- **Red:** at clean head `31be96f69d01a57e89b4a339fc45831297356f4c`,
  `go run ./cmd/connectorgen certification-matrix --check` exited 1 on
  2026-08-13 with `generated artifact ... capability-matrix.json has drift`.
  The identical check fails at recovered base `da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`.
  This is causal generator evidence, not a lint/test setup failure.
- **Masking condition:** the tracked base and head artifact bytes are the same
  (`cecf38c6516844acb9f8c0b0df7bacbd330cbba508ee4cdd5f0ea3269c18614f`),
  while `bb9fcacbfd1d2c97c4f5e0978ee3eb2cebe4776e` added a `connsdk` import
  that shifted six `internal/connectors/connectors.go` line numbers without
  refreshing the matrix.
- **Green plan:** run only `go run ./cmd/connectorgen certification-matrix`,
  then the `--check` command. The expected diff is exactly six
  `discovery_source` values: `capability:catalog`, `capability:cdc`,
  `capability:check`, `capability:query`, `capability:read`, and
  `capability:write` move from lines `49–54` to their shifted `50–55`
  positions. No connector capability, category, operation, executor, or
  baseline changes.
- **Expected invariant evidence:** regenerated base/head bytes match at
  `e63b906cb640b8fb4fc8fd46c1076b77b7dbced7889919d60527f9b4335d520a`;
  stripping all `discovery_source` fields and canonicalizing JSON produces the
  unchanged semantic SHA-256
  `bc5d14758c26755d83a9dc4dcbb715da31d95f67de38e352bc652b752c0819bc`.
  The generator source hash remains
  `bba4dea056e18ecc1231fe68cea8321dfc8a53d2b6ce58ac32ca02fa816a7bbf`.
- **Correction accounting:** this closes inherited generated drift and does
  not consume correction 2/5. The only substantive correction remains the
  already-recorded lint fix at 1/5.
- **Green:** on 2026-08-13, canonical
  `go run ./cmd/connectorgen certification-matrix` completed with
  `connectors=556 capability_complete=0 certified=0`; the follow-up
  `go run ./cmd/connectorgen certification-matrix --check` passed. The actual
  matrix SHA-256 is
  `e63b906cb640b8fb4fc8fd46c1076b77b7dbced7889919d60527f9b4335d520a`.
  `git diff --unified=0` reports exactly six additions and six deletions, all
  `discovery_source` values: catalog `50→51`, cdc `54→55`, check `49→50`,
  query `53→54`, read `51→52`, and write `52→53`.
- **Semantic invariant observed:** base and generated head both canonicalize
  without `discovery_source` to
  `bc5d14758c26755d83a9dc4dcbb715da31d95f67de38e352bc652b752c0819bc`.
  The checked generator is
  `cmd/connectorgen/certificationmatrix.go`; its SHA-256 remains
  `bba4dea056e18ecc1231fe68cea8321dfc8a53d2b6ce58ac32ca02fa816a7bbf` at
  recovered base, #4072, and #4034 precedent `be561871e6bb7d1a5b54d7687743ef8396a2cafe`.

## FULL LOCAL ACCEPTANCE — GENERATED GREEN CHECKPOINT `f52745f26`

All commands below passed on 2026-08-13 at generated-only checkpoint
`f52745f269fdf642a3315646b5c5ee798e959135` before the final evidence-only
handoff record:

- `test -z "$(gofmt -l ...five changed Go files...)"` and
  `git diff --check da8a8ff...HEAD` — pass.
- `go test -count=20 -timeout 20m ./internal/connectors/hooks/github -run
  '^TestGitHubAppAuthRateAdmission|^TestAuthenticatorGithubAppRefusesWithoutDeclaredRoute$'`
  — pass in 11.237s.
- `GOMAXPROCS=2 go test -p 1 -count=1 -timeout 20m
  ./internal/coordination ./internal/connectors/connsdk
  ./internal/connectors/engine ./internal/connectors/hooks/github` — pass.
- `GOMAXPROCS=2 go test -p 1 -race -count=3 -timeout 20m
  ./internal/connectors/engine ./internal/connectors/hooks/github -run
  '^(TestGitHubAppAuthRateAdmission|TestAuthenticatorGithubApp|TestRequireSharedGitHubWriteHook|TestGitHubWriteHook|TestGitHubDeclaredRateLimits|TestGitHubRateLimitAdmission)'`
  — pass (engine 6.287s, GitHub hooks 18.337s).
- `go vet ./internal/coordination ./internal/connectors/connsdk
  ./internal/connectors/engine ./internal/connectors/hooks/github` — pass;
  `go test -count=1 -timeout 20m ./internal/cli` — pass in 159.934s.
- `go build ./cmd/pm` — pass. The real binary is 148,131,618 bytes (141M),
  SHA-256 `0329d354f13187317612ff4d6ee162288723ec0838bd4ebd7c042632b2be7db2`.
  `./pm help connectors`, bare `./pm connectors`, and
  `./pm connectors inspect github --json` all pass using embedded/local
  metadata only.
- Individual repository gates pass: `make tidy-check`, `make lint` (0 issues),
  `make docs-check`, `make smoke-no-build`, `make agent-contract-check`,
  `make connectorgen-validate`, `make connectorgen-surface-sync`,
  `make github-parity-artifacts-check`, `make connectorgen-certification-matrix`,
  `make connector-runtime-preflight`, `make connector-canon-check`,
  `make connector-boundary`, and `make release-workflow-check`.
- `scripts/verify-gsd-workflow da8a8ff07aaf00e5c7965cd4d1d3c7252017d785`
  — pass. No `go test ./...`, `make verify`, provider call, real credential,
  parked #3754 operation, push, PR, or merge was used.

## Delivery Gate

The released bounded local/race/generator/build matrix is complete. The next
authorized validation action is the exact #4072 local-only no-mistakes vector
from `NO-MISTAKES-HANDOFF.md`, without `--yes`. Push, draft PR, CI, or any
parent-ref action remain separately blocked until the fresh parent-route gate
is proven; the parked #3754 run remains out of bounds.

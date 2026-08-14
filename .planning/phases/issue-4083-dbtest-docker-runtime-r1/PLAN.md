# Plan — Issue #4083 Explicit Docker runtime for dbtest

## Objective

Let the shared native database integration harness run through an explicitly
selected Docker or Podman direct local Unix socket, without changing any
connector behavior or relaxing resource ownership, endpoint, identity,
capacity, or cleanup guarantees. This foundation issue blocks #3976 but is
intentionally separate from the PostgreSQL connector lane.

## Scope fence

Allowed paths:

- `internal/connectors/native/dbtest/**`
- `internal/connectors/native/mysql/mysql_integration_test.go`
- `go.mod`
- `.github/workflows/security.yml`
- `AGENTS.md`
- `.planning/phases/issue-4083-dbtest-docker-runtime-r1/**`

Excluded: connector behavior, PostgreSQL connector implementation, external
database endpoints, generic container command surfaces, runtime product
commands, dependencies, credentials, remote endpoints, named connections,
global Docker/Podman defaults, and any merge.

## Design

1. Add a deliberately small runtime selector to `dbtest.Config` with only
   `docker` and `podman` values. Validate it before any command is created and
   retain the direct-local-Unix endpoint validator for both values.
2. Keep `CommandRunner` as the one injectable harness boundary. Its default
   implementation selects the declared client binary and passes the endpoint
   on every invocation (`docker --host`, `podman --url`), so neither global
   default is consulted or mutated.
3. Make identity parsing runtime-aware. Preserve Podman machine capacity
   handling. For Docker, bind the daemon ID and Docker image-store root to the
   configured socket and require a measurable image-store capacity; lack of
   safe capacity evidence is a refusal before a pull, not a skip.
4. Update the MySQL proof to require, when explicitly enabled,
   `POLYMETRICS_CONTAINER_RUNTIME` and `POLYMETRICS_CONTAINER_ENDPOINT`.
   An unset opt-in still skips; an incomplete/unsupported enabled choice fails
   with an actionable message naming `docker`, `podman`, and the direct Unix
   endpoint requirement.
5. Update README and the durable database-harness guidance in `AGENTS.md`.
6. Remediate the CI-discovered Go standard-library vulnerabilities by moving
   the module toolchain declaration and active security workflow pins together
   from Go 1.25.12 to the fixed Go 1.25.13. This is a toolchain patch only:
   do not add dependencies or change connector behavior.
7. Support a Docker daemon whose reported `DockerRootDir` belongs to a local
   VM (for example Colima) but cannot be measured from the client host. Only
   an explicitly configured, pre-cached pinned probe image may be used; the
   harness must never pull it. Run its `df` probe locked down against the
   daemon-side store, retain the direct-Unix endpoint and daemon-identity
   checks, and refuse malformed/unavailable capacity evidence before any
   database-image mutation.

## TDD execution

### Red

Add focused dbtest tests that require an explicit runtime, reject an unknown
runtime and all existing unsafe endpoint shapes for both runtimes, and assert
that each default client command receives its configured endpoint. Add the
enabled-MySQL configuration test/witness that exposes the prior false skip.
Run the selected test against the Podman-only baseline; it must fail for the
missing Docker/runtime-selection behavior, not compilation.

### Green

Implement the smallest runtime value/methods and runtime-aware target proof;
change only the reference MySQL test environment handling; then rerun the
same focused selectors until green.

### Refactor

Name shared resource-not-found, safe path, byte parsing, and target-proof
concepts independently of a particular runtime only where that avoids
duplicating safety logic. Run gofmt and inspect the scoped diff. No connector
code or standalone command path is permitted.

### CI remediation

Capture the existing Go 1.25.12 `govulncheck` failure first, then change every
active project/security pin to Go 1.25.13 and rerun the same scan. Historical
planning evidence remains historical; only the current issue's ledger and
verification checklist record this repair.

### Docker VM capacity follow-up

Add a focused failing test for an explicit Docker target whose `DockerRootDir`
cannot be `stat`ed by the host. The test requires a configured cached probe
image, exact locked-down `run --rm` capacity arguments, and strict numeric
POSIX `df -P -B1` parsing. Green is a daemon-side measurement that does not
pull or retain the probe image, while missing/malformed evidence remains a
pre-mutation refusal. Inspect the generated probe name and refuse an existing
container before claiming it for cleanup. Pass enabled MySQL runtime values
raw to the shared validators so whitespace and control characters remain
refused. Re-run the tagged MySQL reference proof through the
explicit Colima socket and record Docker as observed only on a passing result;
continue to record Podman as unavailable without inspecting its global default.

## Required skills and lifecycle

- `github-issue-first-delivery`
- `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`
- `golang-error-handling`, `golang-security`, `golang-safety`
- `golang-testing`, `golang-context`, `golang-concurrency`, `golang-database`
- `golang-continuous-integration`, `golang-dependency-management`, `golang-lint`
- `golang-documentation`, `no-mistakes`
- `gsd-discuss-phase`, `gsd-plan-phase --tdd`, `gsd-execute-phase`,
  `gsd-verify-work`, `gsd-code-review` (inline/manual fallback; role spawning
  is forbidden by the delivery contract)

## Verification plan

1. Capture focused red then green harness tests.
2. `go test -timeout 20m ./internal/connectors/native/dbtest`.
3. `go test -timeout 20m ./internal/connectors/native/mysql` and the tagged
   MySQL selector with opt-in absent/incomplete configuration witnesses.
4. If a direct local daemon exists, run the exact tagged MySQL proof once with
   Docker and once with Podman. Otherwise record each unavailable socket/daemon
   result; never claim a live run from a client binary alone.
5. `gofmt`, scoped `go vet`, `git diff --check`, `go build ./cmd/pm`, and
   required individual repository gates, followed by inline GSD verification
   and review.
6. Drive no-mistakes without `--yes` to the CI-green PR state; do not merge.
7. Run the exact CI `govulncheck` command with the fixed Go 1.25.13 toolchain
   and confirm the module/workflow pins agree.

# Context — MySQL container harness R1

## Intent

Deliver one reusable, opt-in Podman integration-test harness and prove it with exactly one new
Tier-3 native database connector: MySQL. The harness must start one isolated, pinned-image
container; seed deterministic multi-page data; exercise check, catalog, snapshot, incremental, and
binary-log CDC paths; and reclaim its container, named volume, and generated run-specific image
reference on every exit path.

## Constraints selected from the task

- The test is sequential and opt-in (`POLYMETRICS_DATABASE_INTEGRATION=1`). Its visible skip when
  that opt-in or an explicit `POLYMETRICS_PODMAN_CONNECTION` is absent is intentional. The documented
  command always supplies both, so it cannot report a false live pass.
- The harness never invokes an unscoped Docker command. It passes only a caller-supplied
  `--connection` on every child command; it neither changes a global default nor refers to a
  shared runtime.
- A host port is dynamically assigned by Docker and refused if it resolves to the database default.
  The connector receives host and port as separate configuration fields; no endpoint is logged.
- The MySQL image is `docker.io/library/mysql:8.4.11`, pinned by tag. The harness creates and always
  removes a unique run-specific tag; it never removes the shared pinned source image, whether that
  source was already cached or was pulled for the run.
- Podman is deliberately not used: three independent Podman machines on this host, including the
  Podman is used. The earlier "Podman machines will not start" conclusion was a symptom, not a
  cause: a 10-hour-hung `podman machine start` held the global machine-start lock, and
  Podman/applehv allows only one running VM at a time. Clearing that made a normal machine
  start work first time.
  Removing an image frees space inside the VM but leaves the host's sparse disk file
  inflated, so `POLYMETRICS_DATABASE_RECLAIM_DISK=1` trims the backing machine twice after
  container cleanup. One pass was measured insufficient.
  teardown and before the after-disk measurement. That destructive reset is opt-in and is intended
  only for the disposable default profile used by this test.
- The MySQL test database uses its isolated ephemeral server configuration. No credential or
  connection string is emitted, recorded, or placed in fixture data.
- The connector is Tier-3/dynamic-schema like PostgreSQL and is registered through the native
  MySQL factory, so production registry calls reach its wire-protocol check, catalog, read, and CDC
  implementation.
- MySQL binary logs are a distinct source mechanism, so the shared closed changefeed vocabulary
  gains `binlog_replication` in lockstep in Go validation and JSON Schema. This is the minimum
  shared declaration change necessary for an honest MySQL declaration, not a generic transport.

## Inline GSD fallback

`scripts/gsd doctor`, all five required source resolutions/prompts, and
`go run ./cmd/agentcontractgen check` succeeded. This work is an issue-shaped firstmate task rather
than a numbered roadmap phase, and compatible isolated GSD workers are unavailable/forbidden by the
single-worker delivery contract. Discussion, planning, execution, verification, and review are
therefore recorded inline in this phase directory.

## Required skills loaded

`golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`,
`golang-concurrency`, `golang-database`, `golang-documentation`,
`golang-dependency-management`, and `golang-pkg-go-dev`.

# REST write executor plan

## Scope and verified premise

`operations.schema.json` and the bundle loader accept `kind: "rest_write"`, but
the current execution chain only dispatches operation-backed `direct_read` and
`binary_download`. `commandrunner.Preflight` blocks every other operation-backed
command, and `connectorgen validate` repeats that limitation. This checkout has
2,431 `rest_write` declarations (not the 2,548 stated in the brief).

The shared confirmation seam from PR #3730 is already present in
`internal/connectors/engine/prepared_write.go` and `write_gate.go`. The new
executor must call `PreviewPreparedWrite` and `ExecutePreparedWrite`; it must
not add another approval mechanism.

The three historical data documents named in the task brief are absent from this
checkout and the supplied workspace, so their contents cannot be verified or
used as evidence.

## Delivery approach

1. Add a red integration test using a disposable fixture bundle and loopback
   server. It must prove the whole individual command lifecycle: typed form
   shaping, offline preview, destructive confirmation, a single-use grant,
   exactly one live request, response redaction, and no retry after a failure.
2. Add the public connector contracts and engine executor. Reuse the
   direct-read path for operation lookup, path/query/body shaping, API-surface
   binding, bounded JSON result handling, and output-policy redaction. Reuse
   `PreparedWrite`/`ExecutePreparedWrite` for preview binding and approval.
3. Add an explicit no-retry mode to the requester and use it exclusively for
   `rest_write`: no transient retry and no auth-refresh retry. No idempotent
   exception is enabled in this slice, so no retry citation is needed.
4. Carry the typed operation request through commandrunner and the existing
   connector-command plan/preview/approval/execute lifecycle. Persist all
   path, query, and body inputs in the plan hash; preserve `batchable:false` in
   the approval target while executing only one operation request.
5. Teach connectorgen validation and surface-sync to preflight only
   executable `direct_write` commands against supported `rest_write` metadata.
   Unsupported multipart, text/plain, and wildcard content types remain
   blocked until their separate typed payload contracts exist.

## Guardrails

- No production connector bundle is promoted or changed in this slice. A
  test-only bundle may demonstrate the executor.
- `rest_write` supports declared JSON, absent-content-type JSON, and
  `application/x-www-form-urlencoded` bodies in this slice. Multipart,
  text/plain, and wildcard media types fail closed before network dispatch.
- No generic raw body, HTTP-write, SQL-write, credential, or config path is
  added.
- `batchable:false` is represented as an optional operation declaration whose
  default is true; direct-write command execution is structurally one request,
  never a bulk loop.

## TDD and verification

The red test is recorded in `TDD-LEDGER.md` before production edits. Run scoped
package tests for `internal/connectors/engine`, `internal/connectors/connsdk`,
`internal/connectors/commandrunner`, `internal/app`, `cmd/connectorgen`, and
`internal/cli` as applicable, then the non-suite `make verify` gates listed in
AGENTS.md. The full `go test ./...` and `make verify` suite are CI work because
they exceed the per-command timeout.

## Required skills and workflow record

Skills used: `golang-how-to`, `golang-cli`, `golang-testing`,
`golang-error-handling`, `golang-security`, `golang-safety`,
`golang-design-patterns`, `golang-structs-interfaces`, `golang-context`,
`golang-documentation`, `no-mistakes`, and `gsd-programming-loop`.

`scripts/gsd doctor` passed, but the project-local adapter does not expose the
`programming-loop` command. This phase therefore uses the repository-permitted
manual GSD/TDD fallback. Inline work is deliberate: the active runtime policy
would normally route independent work to agents, but the controlling task and
the active Codex collaboration rule prohibit proactive delegation.

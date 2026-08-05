# Dual-Mechanism Connector Foundations (P0) Plan

Branch: `fm/cli-connector-mechanism-foundations-r1`

## Objective

Deliver the shared, connector-neutral foundation for the approved
dual-mechanism architecture: browser-mediated official OAuth, browser-session
capture without password automation, encrypted per-account local credential
storage, and declared connector mechanism metadata with an explicit web
connector risk-acceptance gate.

This is a foundation-only slice. It does not add WhatsApp, Twitter, Reddit, or
LinkedIn connector implementations, and it deliberately does not scaffold
`reddit-web` or an official LinkedIn connector.

## Decision Sources

- `karthik-agent-workspace/data/decisions/cli-dual-mechanism-connector-architecture-2026-08-03.md`
- `karthik-agent-workspace/data/decisions/cli-dual-mechanism-connector-decisions-2026-08-03.md`
- `karthik-agent-workspace/data/cli-dual-mechanism-connector-plan-r1/report.md` sections 3, 4.3, and 5

## GSD Mode

Manual fallback. `scripts/gsd prompt programming-loop init --phase
connector-mechanism-foundations-r1 --dry-run` reports that `programming-loop`
is not an available adapter command in this checkout. The required lifecycle
is therefore recorded here and in the TDD ledger rather than skipped.

## Recovery Context

This branch was recovered with prior P0 commits already present. Their original
red-test transcript is not recoverable, so this phase record does not claim
historical red evidence. The recovery audit found two remaining foundation
gaps to close test-first:

1. The controlled Rod session captured cookies but did not implement the
   `browserauth.Flow` contract that yields a `SessionCredential`.
2. The OAuth loopback configuration documented a loopback redirect but did not
   reject a non-loopback bind host.

## Slice Plan

1. Add failing tests for a controlled browser-session flow that returns only
   the declared minimum cookies and its configured CSRF metadata; prove it
   never gains password-entry methods.
2. Implement the minimum typed `driver.Flow` adapter over the existing
   controlled session interface. It must use the user-completed browser login,
   capture only declared cookies, and close the browser after capture.
3. Add a failing loopback test for a non-loopback redirect host, then restrict
   authorization-code callbacks to literal loopback IP addresses.
4. Add a failing top-level browserauth test for invalid dual/no credential
   outcomes and explicit re-authentication expiry checks, then enforce that
   invariant at the documented `browserauth.Login` boundary.
5. Add a failing synthesis test for the full declared mechanism block. Carry
   the upstream pin, breakage-review cadence, and local kill-switch reason to
   public metadata/definition/help instead of silently dropping declared
   fields between the engine and catalog.
6. Verify the existing vault, metadata, enable-gate, help/catalog, and
   no-password/no-argv guard coverage together with the new tests.
7. Run the required scoped Go and repository gates, commit the green recovery
   slice, and hand off to firstmate for the no-mistakes pipeline and PR flow.

## Non-Goals / Boundaries

- No password prompt, password typing, password storage, or argv credential
  input.
- No generic HTTP, SQL, shell, or browser-request escape hatch.
- No hosted relay, centralized credential custody, or cross-user session reuse.
- No connector bundle or connector-specific executor in this P0 slice.
- No branch merge, default-branch push, or draft-status change.

## Required Skills

`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, `golang-context`, `golang-concurrency`,
`golang-documentation`, `golang-lint`, `golang-naming`, `golang-code-style`,
`golang-popular-libraries`, `golang-dependency-management`,
`gsd-programming-loop`, and `no-mistakes`.

## Verification Plan

Run focused package tests first, then the affected CLI and connector packages,
`go vet`, binary build, command help/manual checks, and the individual
`make verify` gates specified by `AGENTS.md`. The full suite remains a CI gate
because this checkout's per-command timeout makes `go test ./...` and
`make verify` unreliable as single invocations.

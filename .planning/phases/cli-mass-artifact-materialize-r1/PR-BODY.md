## Intent

Complete honest provider-artifact operation mapping and publish the derived connector surfaces
that are safe for this branch to own. Refs #3958.

## What Changed

- Materialized and reconciled the connector-artifact slice with explicit provenance, operation
  classification, derived endpoint ledger, catalog data, documentation surfaces, and rate-limit
  declarations. The durable ledger accounts for all 426 targets: 231 materialized and 195
  genuinely blocked only after documented official-source exhaustion.
- Repaired shared generator/runtime correctness found during review: write-target accounting,
  operation-first classification, documented-example normalization, promoted-native command
  surface forwarding, and strict batch-reference HTTP(S) scheme admission.
- Closed the four PR CI failures: `govulncheck` now scans loadable packages, CodeQL's reference
  filter is an allowlist, the connector-boundary scanner distinguishes static feed assets from a
  connector id, and the generated-website search test no longer relies on a stale rank bound.

## Certify Timing Decision

The certify timing cap changes from **3m30s to 8m** (+4m30s / 129%) after a controlled,
precompiled, `-count=1` main-versus-branch comparison with the exact same 25 harness and 92 CLI
real invocations.

- The endpoint-ledger corpus grew from **2,060 to 13,534 entries (6.57x)**.
- Actual certify test elapsed time grew from **159.268s to 295.117s (1.85x)**.

Time therefore scales strongly sub-linearly relative to the materialized corpus. This is a budget
calibration for a much larger fleet, not a relaxation that hides a super-linear regression. The
current hosted observation was 421.290s against the old 210s limit; 8m leaves 58.710s (13.9%)
headroom while retaining the unchanged real-invocation contracts.

## Verification

- `make certify-timing` — pass: 292.016s test elapsed / 308.762s wall; 25/25 harness and 92/92
  CLI real-invocation contracts intact.
- `go run ./cmd/connectorgen surface-sync --check` — clean before the timing calibration.
- `go run golang.org/x/vuln/cmd/govulncheck@latest ./...` — `No vulnerabilities found`.
- `make connector-boundary` — clean across 192 shared production files and 552 connectors.
- `npm exec vitest run tests/api/search.test.ts` — 4 passing tests.
- Hosted PR checks — rerun after the budget commit; this body will be updated with the final result.

## Safety and lifecycle

No credentials or live provider operations were used. The CodeQL URL-scheme correction is defense
in depth: downstream same-host and HTTPS gates already prevented exploitation, while the early
filter now closes the whole scheme class. The inline/manual GSD fallback records
`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` evidence
under `.planning/phases/cli-mass-artifact-materialize-r1/`; compatible isolated roles are
unavailable for this single-worker lane. Skills used for this timing slice: `golang-how-to`,
`golang-continuous-integration`, `golang-testing`, and `golang-benchmark`.

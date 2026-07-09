# Issue #185 — Freshchat direct read

Parent issue: #180
Sub-issue: #185
Branch: `feat/185-freshchat-direct-read`

## Scope

Implement the safe Freshchat direct-read gap for `POST /users/fetch` without adding a generic HTTP client or mutation escape hatch.

This slice is intentionally narrow:

- add a typed direct-read body mapping for `pm freshchat user fetch --id ...`;
- allow direct-read POST only for a named Freshchat users-fetch output policy;
- keep max response bytes bounded and connector-relative endpoint validation intact;
- update Freshchat operation ledger/command metadata from blocked/planned to implemented direct-read coverage;
- test commandrunner body mapping and engine request execution with replay/httptest only.

## GSD/TDD

- GSD plan prompt: `scripts/gsd prompt plan-phase issue-185-freshchat-direct-read --skip-research`.
- Programming loop adapter command attempted: `scripts/gsd prompt programming-loop init --phase issue-185-freshchat-direct-read --dry-run`.
- Adapter result: unavailable (`unknown GSD command: programming-loop`); manual programming-loop fallback applies.

## Required skills used

- gsd-core
- golang-how-to
- golang-cli
- golang-testing
- golang-error-handling
- golang-security
- golang-safety
- golang-documentation

## Plan

1. Add red tests proving Freshchat `user fetch` remains blocked/planned and direct-read POST/body support is unavailable.
2. Extend the direct-read model narrowly for body-mapped POST reads with `freshchat_users_fetch` output policy.
3. Update Freshchat `cli_surface.json` and `api_surface.json` to cover `/users/fetch` as an implemented direct read.
4. Run focused validation and full verification.
5. Preserve safety: no credentialed Freshchat calls, no writes, no uploads, no raw HTTP command.

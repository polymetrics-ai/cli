# Plan — #103 Linear sensitive/admin policy

Date: 2026-07-09
Parent issue: #80
Parent branch/PR: `feat/80-linear-cli-parity`, draft PR #131

## Objective

Model Linear sensitive/admin/destructive operation classes as typed fixed-document reverse-ETL actions when an SDK document exists, with preview/approval/typed-confirmation policy and no inline secret/raw mutation escape hatches. Raw arbitrary GraphQL remains blocked by default.

## GSD / fallback

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed.
- `scripts/gsd prompt programming-loop issue-80-linear-complete-ops --skip-research` — unavailable (`unknown GSD command: programming-loop`), so manual PLAN → RED → GREEN → REFACTOR → VERIFY fallback is active.

## Required skills loaded

gsd-core, golang-how-to, golang-cli, golang-spf13-cobra, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-context, golang-concurrency, golang-graphql, golang-documentation.

## Safety

No secrets, no credentialed Linear checks, no new dependencies, no raw generic GraphQL command, no generic HTTP write tool, and no live Linear reverse-ETL execution. Implemented Linear writes remain plan → preview → approval → execute, use fixed GraphQL documents only, and tests use dry-run previews or httptest fixtures.

## TDD plan

1. RED: add focused tests proving the missing behavior for this slice.
2. GREEN: implement the smallest safe metadata/runtime change.
3. REFACTOR: align generated docs/website metadata and safety wording.
4. VERIFY: focused tests, connector validation, CLI help parity checks, then parent gates as time permits.

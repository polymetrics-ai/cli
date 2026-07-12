# Plan — #98 Linear help renderer

Date: 2026-07-09
Parent issue: #80
Parent branch/PR: `feat/80-linear-cli-parity`, draft PR #131

## Objective

Render connector command-surface help for pm linear, pm linear --help, and pm help linear from cli_surface metadata. Add runtime tests and docs parity evidence.

## GSD / fallback

- `scripts/gsd doctor` — passed.
- `scripts/gsd verify-pi` — passed.
- `scripts/gsd list --json` — passed.
- `scripts/gsd prompt programming-loop issue-80-linear-complete-ops --skip-research` — unavailable (`unknown GSD command: programming-loop`), so manual PLAN → RED → GREEN → REFACTOR → VERIFY fallback is active.

## Required skills loaded

gsd-core, golang-how-to, golang-cli, golang-spf13-cobra, golang-testing, golang-error-handling, golang-security, golang-safety, golang-design-patterns, golang-structs-interfaces, golang-context, golang-concurrency, golang-graphql, golang-documentation.

## Safety

No secrets, no credentialed Linear checks, no new dependencies, no raw generic GraphQL command, no generic HTTP write tool, and no Linear reverse-ETL execution. Any implemented Linear write remains plan → preview → approval → execute and uses fixed GraphQL documents only.

## TDD plan

1. RED: add focused tests proving the missing behavior for this slice.
2. GREEN: implement the smallest safe metadata/runtime change.
3. REFACTOR: align generated docs/website metadata and safety wording.
4. VERIFY: focused tests, connector validation, CLI help parity checks, then parent gates as time permits.

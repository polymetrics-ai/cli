# Trello parity wave 03 plan

## Scope

Implement connector-local Trello (`internal/connectors/defs/trello`) official REST API v1 parity for parent issue #3086 and subissues #3087-#3093. This worker runs in branch `fm/cli-trello-parity-wave03-r1`, fixture-only, with no live Trello calls, no credentials, no pushes, no PR, and no `/no-mistakes` pipeline.

## GSD command path and fallback

- `scripts/gsd doctor`: pass in this worktree.
- `scripts/gsd list`: pass; 69 commands exposed.
- `scripts/gsd prompt programming-loop init --phase trello-parity-wave03-r1 --dry-run`: unavailable (`unknown GSD command: programming-loop`).
- Fallback: use the repo-local Pi prompt contract from `.pi/prompts/pm-gsd-loop.md` plus `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` manually. Keep PLAN/TDD-LEDGER/VERIFICATION/RUN-STATE updated before production edits.

## Required skills loaded

`gsd-core`; `golang-how-to`; `golang-cli`; `golang-testing`; `golang-error-handling`; `golang-security`; `golang-safety`; `golang-design-patterns`; `golang-structs-interfaces`; `golang-context`; `golang-concurrency`; `golang-lint`; `golang-documentation`; CLI help/docs/website parity reference.

## Official audit evidence

- Official source fetched/re-indexed: `https://developer.atlassian.com/cloud/trello/swagger.v3.json`.
- Source SHA-256 at audit time: `b50fca38c5ea62025f9778482f89f11ae3da0dd983d31ba49401c4422e450b19`.
- OpenAPI: `3.0.0`; title `Trello REST API`; version `0.0.1`; server `https://api.trello.com/1`.
- Official source path count: 191.
- Official HTTP operation count: 261 (`GET=128`, `PUT=51`, `POST=45`, `DELETE=37`).
- Root distribution: members=45, cards=42, boards=41, organizations=26, enterprises=21, actions=16, checklists=12, lists=11, notifications=11, customFields=8, tokens=8, labels=5, plugins=5, webhooks=5, search=2, applications=1, batch=1, emoji=1.

## Implementation slices

1. **Red contract**: add Trello connectorgen tests proving the current four-row ledger is incomplete and the full official source count must be 261.
2. **Definition expansion**: generated connector-local `metadata.json`, `spec.json`, `streams.json`, `writes.json`, `api_surface.json`, `cli_surface.json`, `certification.json`, schemas, and sanitized fixtures from the official OpenAPI.
3. **Safety classification**: implemented executable boards/lists/checklists streams, fixed JSON direct reads for the remaining supportable GET endpoints, and typed mutations; blocked only duplicate field/filter accessors, raw `/batch`, token/application/enterprise admin surfaces, and token-management surfaces with evidence.
4. **Docs and generated surfaces**: updated Trello `docs.md`, docs connector manual/skill, catalog JSON/MD, website generated connector data, and CLI golden transcripts.
5. **Verification**: required fixture-only gates passed; exact results recorded in `VERIFICATION.md`.

## Expected post-change count target

Truthful final target from this audit: total 261 operations; 219 executable connector operations (3 fixture-backed ETL streams + 95 fixed JSON direct reads + 121 fixture-backed typed writes); 42 blocked/planned with exact evidence; 0 excluded/not-applicable; 0 certified (fixture-only, no live certification). Issue addenda use these actual post-change counts.

## Safety notes

- Trello `key` and `token` remain `x-secret`; fixtures do not contain real or secret-shaped values.
- Writes include Trello token as a redacted secret query parameter in typed write paths because the current declarative write schema has no separate query map; no shared runtime behavior is changed.
- Destructive deletes use `confirm: destructive`, redacted IDs, and idempotent `404` handling.
- Reverse ETL execution is not run; only fixture capture servers exercise engine request construction.
- `/batch` is blocked because it is a raw multi-URL Trello API escape hatch.
- Enterprise, token, and application compliance surfaces are blocked as admin/elevated or token-management surfaces requiring human-gated authorization/scope decisions.

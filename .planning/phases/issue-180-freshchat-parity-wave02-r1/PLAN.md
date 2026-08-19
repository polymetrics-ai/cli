# PLAN — issue #180 Freshchat parity wave02 r1

## Scope

Primary task: implement the complete documented Freshchat connector parity bundle for parent issue #180 and subissues #181-#187, limited to Freshchat-owned connector files, fixtures, docs, CLI metadata, and planning artifacts.

Allowed write scope:

- `internal/connectors/defs/freshchat/**`
- Freshchat-owned fixtures/docs/connector metadata in that directory
- Freshchat-related generated/manual CLI parity artifacts when required by connector registration (the root golden transcript needed the new Freshchat connector-command line)
- `.planning/phases/issue-180-freshchat-parity-wave02-r1/**`

Out of scope / hard stops:

- No shared runtime/foundation edits, no other connector edits, no generated hook/native import sets, no `go.mod`/`go.sum`.
- No provider credentials, live provider calls, live writes, certification claims, no merges/push/PR.
- Provider search/query and binary upload execution remain blocked unless existing definition-owned contracts already support them; record shared-foundation dependencies instead of claiming unsupported execution.

## GSD command path and fallback

- Ran `scripts/gsd doctor` and `scripts/gsd list` successfully.
- Attempted required `scripts/gsd prompt programming-loop init --phase issue-180-freshchat-parity --dry-run`; adapter returned `unknown GSD command: programming-loop`.
- Attempted alias shape `scripts/gsd prompt gsd-programming-loop ...`; same missing command.
- Manual fallback: loaded and followed `.pi/prompts/pm-gsd-loop.md` plus `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md` for the GSD programming loop. This fallback is recorded because the advertised shell command is unavailable in this checkout.

## Required skills loaded

`gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-lint`.

## Official inventory baseline

Official source: Freshchat API docs `https://developers.freshchat.com/api/`, ETag `W/"26e4fd8b1fe01578eae1dbaff6b69224"`, indexed locally for doc lookup.

Current parent graph: #180 plus subissues #181-#187. Source issue counts: 34 total operations; 15 ETL/read, 12 reverse ETL writes, 5 direct/provider-search/query, 2 binary/file, 0 CDC, 0 excluded/not-applicable.

Existing Freshchat bundle before edits: 18 streams, 13 write actions, 34 `api_surface.json` endpoint rows with 31 `covered_by` and 3 legacy `excluded` rows (`POST /users/fetch`, `POST /files/upload`, `POST /images/upload`).

Shared-foundation dependencies to preserve honestly:

- #2985 provider search/query boundary: blocks executable provider-search/read-body command support for `POST /users/fetch`.
- Binary/multipart upload execution has no connector-local safe contract in the current definition dialect for Freshchat; keep `/files/upload` and `/images/upload` represented as blocked binary operations.
- #2986/#2988 CDC foundation issues are not claimed for Freshchat because the official Freshchat API inventory has 0 CDC/changefeed operations.

## Implementation slices

1. **Red/validation slice:** capture current Freshchat validation/conformance state before edits.
2. **Operation ledger slice:** convert legacy excluded rows to operation-ledger `blocked` rows with `operation_ledger_version: 1`, source URLs, exact reasons, no exclusions.
3. **Reverse ETL safety slice:** ensure destructive/admin Freshchat actions are typed through `confirm: "destructive"`, risk text, redaction where useful, and existing plan -> preview -> explicit approval -> execute docs.
4. **Connector-local CLI metadata slice:** add Freshchat `cli_surface.json` covering all implemented ETL/write surfaces plus planned/blocked provider search and binary operations without raw escape hatches.
5. **Certification/evidence slice:** add fixture-only certification metadata for safe source defaults/live-unavailable classifiers only; keep certified=false/no live claim and no Freshchat write pairing claim.
6. **Docs/fixtures slice:** update `docs.md` to reflect exact operation counts, blocked dependencies, destructive/admin confirmation, certification status, and official source evidence.
7. **Issue addendum slice:** idempotently append captain-policy addendum to #180-#187 using `gh-axi`, preserving existing bodies and count tables.
8. **Verification and commit slice:** run focused validation/gates, update generated Freshchat-related CLI golden output, record results, commit branch, stop before no-mistakes/push/PR.

## Orchestration decision

Cycle plan: `local_critical_path`. A single worker owns one connector directory; parallel mutating subworkers would collide in the same Freshchat files. Read-only official-doc and issue recon used local/tool indexing only.

# PLAN — issue-3175 Chargebee parity wave05 r1

## Scope

Parent issue: #3175. Child issues: #3176, #3177, #3178, #3179, #3180, #3181, #3182.
Branch: `fm/cli-chargebee-parity-wave05-r1`.

Allowed write scope: `internal/connectors/defs/chargebee/**`, focused connector docs/generated surfaces (`docs/connectors/**`, `docs/cli/**` as generated), owned focused tests/fixtures for Chargebee, and this phase artifact. No shared runtime behavior changes, no dependencies, no live provider calls, no secrets, no certification claim, no push/PR/no-mistakes.

## GSD command path and fallback

- Ran `scripts/gsd doctor` successfully.
- Ran `scripts/gsd list` successfully.
- Attempted mandatory `scripts/gsd prompt programming-loop init --phase issue-3175-chargebee-wave05-r1 --dry-run`; adapter returned `unknown GSD command: programming-loop`.
- Used repo-local prompt path `scripts/gsd prompt plan-phase issue-3175-chargebee-wave05-r1 --skip-research` and this manual GSD programming-loop artifact as fallback for the missing programming-loop command. The adapter itself is healthy; the command is absent from the installed registry.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`
- `golang-testing`
- `golang-error-handling`
- `golang-security`
- `golang-safety`
- `golang-design-patterns`
- `golang-structs-interfaces`
- `golang-context`
- `golang-concurrency`
- `golang-documentation`

Policy/reference files read: `AGENTS.md`, required skill routing, GSD Pi adapter reference, CLI help/docs/website parity reference, issue-agent contract, parent-orchestrator contract/loop, universal runtime loop, migration handoff, migration conventions, connector architecture v2 design, Claude/Copilot review routing references.

## Source/audit plan

1. Re-audit the pinned official Chargebee OpenAPI source from `chargebee/openapi` commit `fbd261f5383317cdc98d00d448ba038cc0659df1`, whose spec `info.version` is `2026-07-21.2a6a65b3e1a8ff29840466a7bfdb5cdd778d0634`.
2. Reconcile against the issue/audit counts: 655 total documented operations = 432 REST operations + 223 OpenAPI webhook operations. Lane target: 125 ETL reads, 264 reverse-ETL writes, 18 direct/provider-query reads, 14 binary/file operations, 234 CDC/changefeed operations, 0 excluded.
3. Convert `api_surface.json` to operation-ledger mode (`operation_ledger_version: 1`), remove legacy `excluded` rows, and classify every official operation exactly once as implemented `covered_by` or blocked operation with exact lane notes/source evidence.

## Implementation slices

### Slice A — red/count tests

- Add a focused Chargebee operation-ledger/count test under `cmd/connectorgen` before production definition changes.
- Expected red state on current tree: 428 old rows, no `operation_ledger_version`, legacy exclusions, 32 stream rows, 36 write rows.

### Slice B — connector definition expansion

- Generate/curate Chargebee `streams.json`, schemas, and sanitized stream fixtures for all ETL/read operations supported by the declarative engine.
- Generate/curate Chargebee `writes.json` and sanitized write fixtures for all reverse-ETL operations supported by the declarative engine.
- Preserve existing curated streams/write shapes where present; add generated conservative schemas/fixtures for new operations.
- Mark destructive writes with `confirm: "destructive"`, typed closed schemas, redaction on path identifiers where appropriate, and idempotent 404 delete semantics where provider behavior supports missing-as-not-found.

### Slice C — direct/binary/CDC truthfulness

- Do not edit shared direct/binary/CDC runtimes.
- Direct/provider-query Chargebee operations are form-encoded POST or official paths with hyphenated path variables; the existing operation direct-read executor only supports JSON POST bodies and `{identifier}` path variables. Block unsupported direct ops with exact shared-runtime dependency notes.
- Binary/file operations remain blocked operation-ledger rows where no connector-local executable binary runner exists without shared runtime changes.
- Webhooks/hosted event surfaces remain CDC/changefeed blocked rows pending #2986/#2988.

### Slice D — docs/generated surfaces

- Update `docs.md` with exact official-source/version/count ledger, implemented vs blocked counts, safety boundaries, destructive confirmation, direct/binary/CDC blockers, and fixture-only uncertified status.
- Regenerate connector manuals/catalogs/skills with `pm docs generate` after `go build` makes the local `pm` binary available.

### Slice E — issue addendum and local verification

- Compute post-change counts from the actual files.
- Append the established captain-policy addendum idempotently to #3175-#3182 using `gh-axi issue comment`, with the marker `<!-- chargebee-parity-wave05-r1-captain-policy-addendum -->` and truthful local gate results.
- Run required gates: focused connectorgen validation and conformance for `chargebee`, focused CLI tests, `go build ./cmd/pm`, `make connector-boundary`, `make verify`, `git diff --check`.

## Safety gates

- No live Chargebee API calls or credential requests.
- No secret values in fixtures, docs, issue comments, or logs.
- No generic HTTP method/path/body/query, shell, file, raw API, or passthrough escape hatches.
- Reverse ETL remains plan -> preview -> approval -> execute; destructive actions require typed confirmation.
- No no-mistakes invocation beyond setup doctor, no push, no PR updates, no merge.

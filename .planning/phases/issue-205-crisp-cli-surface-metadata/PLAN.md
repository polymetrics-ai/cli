# GSD Plan — Issue #205 Crisp CLI surface metadata

Issue: https://github.com/polymetrics-ai/cli/issues/205
Parent: https://github.com/polymetrics-ai/cli/issues/204
Parent branch: `feat/204-crisp-cli-parity`
Planned branch: `feat/205-crisp-cli-surface-metadata`
GSD command path: `scripts/gsd prompt plan-phase 204 --skip-research` for parent planning; manual programming-loop fallback because `scripts/gsd prompt programming-loop ...` is not registered.

## Scope

Create the initial non-executable Crisp connector metadata scaffold:

- `internal/connectors/defs/crisp/metadata.json`
- `internal/connectors/defs/crisp/spec.json`
- `internal/connectors/defs/crisp/streams.json` with empty stream list only so the bundle loads before #207.
- `internal/connectors/defs/crisp/api_surface.json` from the official Crisp REST docs.
- `internal/connectors/defs/crisp/cli_surface.json` mapping provider-inspired command paths to safe Polymetrics intents with planned availability only.
- `internal/connectors/defs/crisp/docs.md` with fixed connector headings and honest known limits.
- Generated connector docs/catalog artifacts needed by `pm docs validate` for the new catalog entry.
- Catalog count tests/help text that must increase when the new declarative bundle is present.

No executable ETL streams, direct reads, writes, binary downloads, or reverse-ETL actions are implemented in #205.

## Source of truth

- Official docs: https://docs.crisp.chat/references/rest-api/v1/
- Parsed baseline: 220 non-HEAD operations.
- Method split target: `GET=91`, `POST=47`, `PATCH=44`, `PUT=12`, `DELETE=26`.

## Planned API-surface representation

Use `api_surface.json` `operation_ledger_version: 1` so every official operation is accounted for as a blocked metadata row until executable subissues replace rows with `covered_by` entries.

Initial #205 row mapping:

- GET rows -> `operation.model=direct_read`, `risk=low|medium`, `blocked_by_default=true`.
- POST/PUT/PATCH rows -> `operation.model=sensitive_reverse_etl` or `admin_reverse_etl` depending on title/path heuristics, `blocked_by_default=true`.
- DELETE rows -> `operation.model=destructive_action`, `blocked_by_default=true`.
- Binary-looking rows -> `operation.model=binary_read`, `blocked_by_default=true`.

This is intentionally non-executable inventory metadata. #208 owns exact final classification and replacement with `covered_by.stream`, `covered_by.direct_read`, `covered_by.write`, or `covered_by.binary_read` where implemented.

## CLI surface plan

Generate `cli_surface.json` command entries from each operation:

- `intent=etl` for GET rows that are likely durable list/search surfaces, availability `planned`.
- `intent=direct_read` for GET rows that are detail/check/download-style reads, availability `planned`.
- `intent=reverse_etl` for POST/PUT/PATCH/DELETE rows, availability `planned`, with explicit risk and approval text.
- No `intent=direct_write`, `raw_api`, generic HTTP write, shell write, or SQL write command.
- No command dispatch target (`stream`, `write`, or `operation`) until a later issue implements it.

## Red / green evidence

Red validation before production edits:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/crisp
```

Expected: failure because the Crisp bundle does not exist yet.

Green targeted validation after scaffold (connectorgen validates connector directories under a root; a direct connector path reports 0 checked connectors):

```bash
tmp=$(mktemp -d); cp -R internal/connectors/defs/crisp "$tmp/crisp"; go run ./cmd/connectorgen validate "$tmp"
```

Fleet validation:

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

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
- `golang-documentation`
- `golang-lint`

## Safety

- Public docs only; no credentials.
- No live Crisp API calls.
- No secrets in fixtures/docs; no fixtures in #205.
- No reverse ETL execution.
- No generic raw write surface.
- No new dependencies.

## Done criteria for #205

- Crisp official operation count in `api_surface.json` is 220 with the expected method split.
- `cli_surface.json` contains a planned safe intent for each operation or a safe docs-only classification.
- Targeted temp-root Crisp validation passes with 1 connector checked and 0 findings.
- Fleet validation passes.
- Generated connector docs/catalog validation passes.
- Full local `make verify` passes on the stacked branch.
- Parent ledger and #205 TDD/verification artifacts are updated with command results.

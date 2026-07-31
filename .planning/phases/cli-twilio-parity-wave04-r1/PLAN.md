# PLAN — Twilio parity wave04 r1

## Scope

Implement parent #3167 and child issues #3168-#3174 in this disposable worker branch for the `twilio` connector only. The official source of truth is Twilio public REST API v2010 (`twilio_api_v2010.json`, sha256 `a6753266b8b05a201e8658734e332ee51d07a0913f2d419335d87bdb287643a2`) plus the landed r2 audit record rank 61.

## GSD command path

- Ran `scripts/gsd doctor` successfully.
- Attempted required programming-loop prompt: `scripts/gsd prompt programming-loop init --phase cli-twilio-parity-wave04-r1 --dry-run`; adapter returned `unknown GSD command: programming-loop`.
- Fallback used: `scripts/gsd prompt gsd-plan-phase cli-twilio-parity-wave04-r1 --skip-research`, then manual universal programming loop with explicit TDD ledger, verification checklist, and RUN-STATE updates.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-cli`, `golang-spf13-cobra`
- `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-lint`
- References: required skill routing, CLI help/docs/website parity, GSD Pi adapter, issue-agent contract, connector migration conventions, connector architecture v2 design.

## Baseline evidence before production edits

- `gh-axi issue view --full` read for #3167-#3174.
- Official OpenAPI re-audit: 197 REST method/path operations; lanes `etl_read=96`, `reverse_etl_write=93`, `direct_read_query_search=0`, `binary_file=3`, `cdc_changefeed=5`, `excluded_not_applicable=0`; methods `GET=103`, `POST=62`, `DELETE=32`.
- Current `api_surface.json` has 197 endpoint rows and matches the OpenAPI method/path set exactly.
- Current local bundle has 103 streams, 94 write actions, 7 stream fixture files, 0 write fixture files, no certification contract.

## Slices

1. **Red coverage test**: add a connector-owned conformance test proving Twilio operation fixture coverage and lane arithmetic fail on current missing write/stream fixtures and missing operation ledger version.
2. **Ledger refresh**: update `api_surface.json` reviewed date/scope/operation ledger version and keep method/path coverage exact; document media/binary and event/notification lane classification.
3. **Fixture expansion**: generate sanitized stream fixtures for every stream and write fixtures for every write action. Keep values synthetic and non-secret-looking.
4. **Docs/surfaces**: update `docs.md`; regenerate connector manuals, skills, catalog/manual surfaces, website data/examples if the project generators touch them.
5. **GitHub issue addendum**: after final counts are computed, idempotently append the captain-policy addendum to #3167-#3174 with truthful counts.
6. **Verification**: run focused connectorgen validation/conformance, focused CLI tests, `go build ./cmd/pm`, `make connector-boundary`, `make verify`, and `git diff --check`.

## Safety constraints

- No live Twilio calls, no credential requests, no live writes, no certification execution.
- No new dependencies.
- No generic HTTP/path/body/query, shell, file, SQL, or passthrough escape hatches.
- Reverse ETL remains typed write schemas plus plan -> preview -> explicit approval -> execute; destructive deletes keep typed `confirm: destructive` and idempotent delete semantics.
- Do not edit shared runtime behavior.

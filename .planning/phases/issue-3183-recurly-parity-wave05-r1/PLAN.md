# Recurly official API parity wave05 r1 plan

## Scope

Parent #3183 with subissues #3184-#3190. Connector-local implementation under `internal/connectors/defs/recurly` plus owned fixtures, generated connector docs/catalog/website/help surfaces, and focused tests only. No live Recurly calls, credentials, pushes, PR updates, certification execution, VPS, Thaalam, or shared runtime behavior changes.

## GSD / skills evidence

- GSD adapter health: `scripts/gsd doctor` passed.
- GSD command prompt used: `scripts/gsd prompt plan-phase issue-3183-recurly-parity-wave05-r1 --skip-research`.
- Programming-loop adapter note: `scripts/gsd prompt programming-loop ...` and `scripts/gsd prompt gsd-programming-loop ...` both reported `unknown GSD command: programming-loop`; using manual GSD/TDD fallback recorded here.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, `golang-spf13-cobra`, `context-mode`.
- Required references read: `AGENTS.md`, required skill routing, GSD Pi adapter reference, CLI/help/docs/website parity reference, connector migration handoff, conventions, and architecture design.

## Authoritative inputs

- Issues read via `gh-axi issue view --full --comments`: #3183-#3190.
- Official sources re-audited: Recurly developer docs v2021-02-25 and official OpenAPI YAML v2021-02-25.
- Local landed audit path named by issues (`cli-official-api-parity-audit-r2/audit.json`) is absent from this checkout (`find`/`git ls-tree` found no such file); issue body counts and the official source are used as provenance, with this absence documented rather than fabricating a local audit record.
- Official source operation count from YAML: 197 operations (GET 97, POST 42, PUT 35, DELETE 23), matching issue parent total.

## Classification plan

- ETL streams: 93 GET operations (all non-binary, non-preview-renewal GET operations) with sanitized stream fixtures.
- Direct/provider query/search: 5 preview/read-query operations (`preview_invoice`, `get_preview_renewal`, `preview_subscription_change`, `preview_purchase`, `preview_gift_card`) as bounded `rest_read` operations and implemented direct-read CLI metadata where JSON runtime support exists.
- Binary/file: 3 official export/PDF operations recorded as bounded binary operations in `operations.json`; executable binary transfer is not widened in shared runtime.
- Reverse ETL writes: 96 POST/PUT/DELETE operations (excluding 4 preview POST read-queries), typed declarative write actions with closed schemas, destructive confirmation on deletes/destructive lifecycle mutations, and Recurly idempotency notes.
- Exclusions: none unless the official source proves a documented operation is deprecated/duplicate/disallowed and not in the parent issue count.

## Implementation steps

1. Generate/update Recurly bundle files from the official OpenAPI: `metadata.json`, `spec.json`, `streams.json`, `writes.json`, `operations.json`, `cli_surface.json`, `api_surface.json`, schemas, fixtures, `docs.md`, and `certification.json` if useful for fixture-only certification metadata.
2. Preserve no generic HTTP/body/query/shell/file passthrough: every command/action/operation is named, path-fixed, schema-gated, and bounded.
3. Add fixtures for every stream and write action; use synthetic values only.
4. Regenerate connector docs/catalog/website/generated surfaces with existing project commands.
5. Append the captain-policy destructive-operation addendum idempotently to #3183-#3190 using `gh-axi`, with post-change counts.
6. Run required gates: focused connectorgen validation for Recurly, focused conformance, focused CLI tests, `go build ./cmd/pm`, `make connector-boundary`, `make verify`, and `git diff --check`.
7. Commit the clean result on `fm/cli-recurly-parity-wave05-r1`; do not push or invoke `/no-mistakes`.

## Resume session — close the CLI command-surface gap

The previous wave committed a complete operation ledger (197 ops) but left only 8
`direct_read` commands in `cli_surface.json`: the 93 ETL streams and 96 typed
reverse-ETL writes had no individual `pm recurly <command>` entries, so the
connector failed the complete-parity bar (an ETL-only connector). This session:

1. Wrote `scripts/gen-recurly-cli-surface.py` (deterministic generator reading
   streams/writes/operations/api_surface) and `scripts/regen-recurly-surface.sh`
   to produce the full 197-command surface: 93 implemented ETL stream commands,
   96 implemented reverse-ETL write commands (required record fields mapped to
   typed leaf flags), plus the preserved 5 implemented direct-reads and 3 planned
   bounded binary operations. All 96 writes are executable (implemented); none
   required a partial/blocked disposition because every required mapping-path
   leaf is scalar/array-mappable.
2. Verified each command resolves against streams.json/writes.json/operations.json:
   `go run ./cmd/connectorgen validate` → 549 connectors, 0 findings (no
   `cli_surface_unknown_target`).
3. Regenerated Recurly MANUAL/SKILL, the connectors catalog (Recurly now shows
   93 streams / 96 writes with `read, write` capability), and the website
   connector data — scoped so only Recurly-owned files changed.
4. Added `cmd/connectorgen/recurly_full_surface_test.go` locking the invariant:
   every stream/write has a command, coverage counts (stream=93, write=96,
   direct_read=5, operation=3), and 194 implemented / 3 planned.

### Command-surface verification
- [x] Full surface: 197 commands; intents etl=93, reverse_etl=96, direct_read=8.
- [x] availability implemented=194, planned=3 (bounded binary metadata).
- [x] connectorgen validate 0 findings; conformance recurly PASS.
- [x] internal/cli suite, commandrunner, engine, defs, bundleregistry, conformance, connectorgen all PASS.
- [x] `pm docs validate` PASS; MANUAL/SKILL/catalog/website regenerated; scope limited to Recurly-owned files.

## Safety gates

- Secrets: all fixtures use synthetic placeholders; no credential prompts or values.
- Writes: reverse ETL only, no live execution; destructive actions use `confirm: "destructive"`.
- Direct/binary: bounded and fixed-target only; no raw method/path/body/query escape.
- Shared runtime: do not add provider-specific or generic shared runtime branches.

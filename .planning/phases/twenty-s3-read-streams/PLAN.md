# Twenty S3 read streams — PLAN

Issue: #280
Parent: #277
Branch: `feat/280-twenty-streams`
Base: `feat/277-twenty-connector-parity` @ `b4895064`

## GSD / skills

- GSD mode: repo-local Pi autonomous loop + per-issue GSD artifacts.
- GSD commands checked by orchestrator: `scripts/gsd doctor`, `scripts/gsd list`.
- Manual-GSD fallback note: `scripts/gsd prompt programming-loop ...` is not exposed in this repo adapter (`unknown GSD command: programming-loop`); this slice records equivalent PLAN / TDD-LEDGER / VERIFICATION evidence under `.planning/phases/twenty-s3-read-streams/`.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-documentation`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-testing`, `golang-context`, `golang-concurrency`, `golang-graphql`.

## Scope delivered

- Add 28 Twenty REST list streams to `internal/connectors/defs/twenty/streams.json`.
- Add 28 `GET /rest/<object>` `api_surface.json` rows covered by those streams.
- Add 28 `GET /rest/<object>/{id}` rows as `excluded.category=out_of_scope` because the current direct-read engine only supports `github_contents_*` output policies.
- Add focused Go coverage for the Twenty bundle stream count, schema references, cursor pagination, and read API-surface classification.
- Add sanitized two-page attachments fixture so conformance can exercise a representative stream now that streams exist.

## Out of scope

- `cli_surface.json` and help/manual/website parity: S6 #283.
- Writes and destructive actions: S4 #281, S5 #282.
- Generic direct-read engine support: follow-up engine capability outside S3.
- Credentialed live checks and reverse ETL execution.

## Slice invariant

Every committed state remains loader-valid: each stream has a matching `api_surface.json` coverage row, each schema ref points to the S2 schema file, and get-by-id rows are honestly excluded rather than dangling on unsupported direct-read machinery.

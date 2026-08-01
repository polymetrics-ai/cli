# Workday REST parity wave05-r1 plan

## GSD path and skills

- GSD command path: `scripts/gsd doctor` passed; `scripts/gsd prompt programming-loop init --phase issue-3231-workday-rest-parity --dry-run` failed because this repo-local adapter does not expose `programming-loop`, so this is a recorded manual-GSD fallback.
- Required skills loaded: `gsd-core`, `golang-how-to`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-testing`, `golang-documentation`, `context-mode`.

## Scope

Parent issue #3231 and subissues #3232-#3238 require connector-local Workday REST official API parity for the 2026.30 productionConfidenceLevel directory. No provider credentials, no live provider calls, no provider writes, no shared runtime behavior edits, no generic passthroughs.

Allowed files are the Workday REST bundle under `internal/connectors/defs/workday-rest/**`, generated Workday connector docs/catalog/website surfaces, issue-local planning artifacts, and status/report files needed by firstmate. Shared runtime changes are explicitly out of scope.

## Inventory source

- Official directory: `https://community.workday.com/sites/default/files/file-hosting/restapi/services2026.30.json`
- Official service specs: the 52 `specFilePath` entries in that directory manifest.
- Current fetch evidence: production services=52, HTTP operations=920 (`GET`=655, `POST`=154, `PATCH`=58, `DELETE`=33, `PUT`=20); archived services are not counted.

## Implemented result

1. Generated a deterministic Workday REST bundle from official specs.
2. `api_surface.json` partitions all 920 current production operations exactly once using service `basePath + path` for disambiguation.
3. Implemented 463 JSON resource GET operations as fixture-backed streams.
4. Implemented 174 Workday values/search GET operations as bounded direct reads with `clinical_json_redacted` output policy and 1 MiB caps.
5. Implemented 251 non-binary/non-generic-query mutation operations as typed write actions with closed root record schemas, path-field redaction, risk text, and destructive delete confirmation/idempotent 404 handling.
6. Marked 32 binary/file/fixed-query/generic-query current-contract gaps as blocked operation-ledger rows with official source evidence. No raw query/path/body or binary passthrough was added.
7. Added 463 stream fixtures, 251 write request-shape fixtures, Workday CLI direct-read metadata, fixture-only certification metadata, generated connector docs/catalog/website data, and updated CLI golden transcripts.

## Safety constraints

- `access_token` remains `x-secret`; no secret values in fixtures/docs/issues.
- No raw HTTP method/path/body/query, shell, local file, or passthrough operation is exposed.
- Binary/file operations stay blocked until a bounded binary executor and safe file policies are available.
- Reverse ETL writes are named actions only and depend on the existing plan -> preview -> approval -> execute flow.

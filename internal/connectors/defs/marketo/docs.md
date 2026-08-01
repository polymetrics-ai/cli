# Overview

Marketo is generated from the official AdobeDocs Marketo Engage REST Swagger 2.0 assets for asset, identity, lead database/bulk (`mapi`), and user-management APIs.

Post-change documented operation counts:

| disposition | count |
| --- | ---: |
| official operations | 327 |
| fixture-backed ETL/changefeed streams | 117 |
| bounded redacted direct reads | 28 |
| typed reverse-ETL writes | 158 |
| blocked binary/InputStream downloads | 10 |
| not-applicable identity token operations | 2 |
| blocked write-query operations | 11 |
| blocked dynamic custom-field body operations | 1 |
| certified live operations | 0 |

This connector does not expose a raw HTTP method/path/body command, generic query command, shell, arbitrary file command, or passthrough escape hatch.

## Auth setup

Configure `base_url` as the tenant-specific Marketo host root, for example `https://123-ABC-456.mktorest.com`. Do not include `/rest/v1`; each stream/action declares the official path from the Swagger operation.

Configure `access_token` as a secret value through the credential store or stdin/environment-backed credential workflows. Do not paste secret values into prompts, docs, issue comments, or fixtures. The two `/identity/oauth/token` operations are listed as not-applicable in `api_surface.json`; this connector does not request client IDs or client secrets and does not refresh tokens.

Some stream paths require connector config values such as `export_id`, `api_name`, `batch_id`, `program_id`, `list_id`, `lead_id`, or `field_api_name`. Use `--config key=value` with provider-safe identifiers when reading those specific streams.

## Streams notes

The bundle declares 117 fixture-backed streams. Cursor-paginated Marketo endpoints use `nextPageToken`/`moreResult`; offset-paginated asset/user endpoints use the documented offset and page-size parameters. Every stream has a fixture page under `fixtures/streams/<stream>/page_1.json` and a schema under `schemas/<stream>.json`.

Examples:

- `pm marketo etl get-all-channels --json --limit 25`
- `pm marketo direct get-are-leads-member-of-list --json --list-id 123 --id 456 --max-bytes 1048576`
- `pm marketo reverse update-email --preview --json ...`

## Write actions & risks

The bundle declares 158 typed write actions. Writes are only executable through reverse ETL planning:

1. create a plan from a typed command or reverse ETL mapping;
2. preview the resolved request and staged record count;
3. approve with the plan approval token;
4. execute the approved plan.

Delete/remove/cancel/discard/deactivate/unapprove operations require `confirm: destructive` in addition to the approval token. Multipart import actions are operation-specific typed uploads with bounded file parts; they are not generic file tools.

## Known limits

- Fixture replay and local conformance do not certify live Marketo behavior; certified count remains `0` until a separate live-safe certification run with approved credentials exists.
- The current shared direct-read executor accepts bounded JSON responses only. The 10 official InputStream export/failure/warning download endpoints remain blocked in `api_surface.json` until a shared bounded binary/file transfer executor is available.
- Optional write query parameters that are not required by the official Swagger operation are not encoded by the current declarative write contract. Write operations that require URL query parameters are blocked in `api_surface.json` until the shared declarative write contract has a structured typed query map; this avoids interpolating record values into action paths. `Sync Program Member Data` remains blocked because the official request body uses dynamic custom field names (`input[].{fieldApiName}`), which cannot be represented as a closed schema without an arbitrary-body escape hatch.
- Generated stream schemas use passthrough projection with a small common schema envelope because the Marketo Swagger response envelopes are broad and often resource-specific; fixtures prove request/pagination wiring without claiming live certification.

# Notion official OpenAPI audit — 2026-07-31

Public source fetched with no credentials:

- URL: `https://developers.notion.com/openapi.json`
- OpenAPI: `3.1.0`
- API title/version: `Notion API` / `1.0.0`
- Response bytes: `858733`
- Last-Modified: `Thu, 30 Jul 2026 23:41:58 GMT`
- SHA256: `4170a025f155ab721aa2e30451d1143ae79ad069b784b98d4fa420855f5d9d86`
- Paths: `34`
- Official HTTP operations: `49`

## Post-change official operation disposition

| total official ops | implemented | fixture-tested | blocked/planned | excluded/not-applicable | certified |
| ---: | ---: | ---: | ---: | ---: | ---: |
| 49 | 45 | 45 | 1 | 3 | 0 |

Additional connector-surface coverage: two fixture-tested legacy-compatible `/search` stream variants
(`databases`, `pages`) preserve the existing object-filtered stream contract. These are represented
as extra connector-surface rows in `api_surface.json` but are not counted as additional official
OpenAPI operations; the official `POST /search` operation is counted once via `search_results`.

## Blocked/planned official operation

- `POST /file_uploads/{file_upload_id}/send` (`upload-file`): multipart binary byte transfer is in
  scope, but remains blocked/planned until the shared binary payload approval/conformance runner can
  provide approved digests and live-safe redacted artifacts without broad local file exposure. No
  live file upload was attempted.

## Excluded / not applicable official operations

- `POST /oauth/token` (`create-a-token`)
- `POST /oauth/revoke` (`revoke-token`)
- `POST /oauth/introspect` (`introspect-token`)

Rationale: OAuth token lifecycle endpoints are auth-flow operations, not connector data operations;
executing or exposing them would require an approved auth-flow/scope foundation. They are represented
as `disallowed` operation rows and are not certified.

## Fixture evidence

- Executable official operations fixture-tested by conformance: `45`.
- Connector executable surfaces fixture-tested by conformance: `47` (`25` streams + `22` writes),
  including the two legacy-compatible `/search` stream variants.
- Required fixture-only gate: `go test ./internal/connectors/conformance -run 'TestConformance/notion' -count=1` passed after expansion.

No Notion workspace credentials, live provider reads/writes, live file uploads, or certification
runs were used.

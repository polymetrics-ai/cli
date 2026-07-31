# SUMMARY — issue-3062 Notion parity wave03 r1

Status: implementation and local verification complete; local commit pending.

Manual GSD fallback was used because the repo-local `scripts/gsd` registry did not expose the documented `programming-loop` command, despite `scripts/gsd doctor` passing.

## Delivered

- Re-audited public Notion OpenAPI (`https://developers.notion.com/openapi.json`): 49 official operations, sha256 `4170a025f155ab721aa2e30451d1143ae79ad069b784b98d4fa420855f5d9d86`, last-modified `Thu, 30 Jul 2026 23:41:58 GMT`.
- Expanded `internal/connectors/defs/notion` from a read-only partial bundle to fixture-backed read/write parity for supportable official data operations:
  - 25 streams with sanitized fixtures.
  - 22 typed reverse-ETL actions with sanitized fixtures.
  - `certification.json` explicitly records fixture-only, uncertified status.
  - `api_surface.json` records 51 connector rows: 49 official operations + 2 legacy-compatible object-filtered `/search` stream variants.
- Final official-operation counts: total=49, implemented=45, fixture-tested=45, blocked/planned=1, excluded/not-applicable=3, certified=0.
- Blocked/planned official operation: multipart `POST /file_uploads/{file_upload_id}/send` pending shared binary payload approval/conformance runner.
- Excluded/not-applicable official operations: OAuth token, revoke, and introspection endpoints.
- Extended the Notion StreamHook to support all executable GET/POST read routes, body-cursor pagination, query/body/path interpolation, single-object streams, stream-specific POST pagination keys, repeated-cursor fail-closed behavior, and projection.
- Added `connectorgen validate` support for single connector bundle directories (including malformed bundles missing metadata) and regression tests, so the required exact validation command now checks exactly one connector.
- Regenerated Notion MANUAL/SKILL and website generated connector catalog data; unrelated connector-doc churn from the all-doc generator was reverted.
- Appended the captain destructive/delete policy addendum idempotently to issues #3062-#3069 via `gh-axi` with actual post-change counts and no live certification claim.

## Verification

All required local fixture-only gates passed after local review fixes, including final `make verify`. No live Notion calls, credentials, provider writes, live file uploads, certification, push, PR, merge, `/no-mistakes`, VPS, or Thaalam actions were performed.

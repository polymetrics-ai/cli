# TDD ledger: Google Analytics Data API 24-operation parity

| Slice | Red evidence | Green evidence | Commit |
| --- | --- | --- | --- |
| Provider inventory | Official Google reference lists 11 v1beta + 15 v1alpha methods; the existing HOOK ledger contains a non-provider-derived 10/11-method narrative. | `api_surface.json` records exactly 24 semantic operations, their provider method/path, version/revision, retrieval date, and 20-read/4-write split. | Ledger-only checkpoint |
| v1alpha fixed GET reads | Table-driven tests for every added property quota, audience-list, recurring-audience-list, and report-task GET operation fail before dispatch/fixture support exists. | Fixture and `httptest` tests show only fixed GET paths, normalized property IDs, bounded/redacted JSON output, and no secret leakage. | Alpha direct-read checkpoint |
| POST/read and write policy | Operation/CLI validation fails when an official operation has no classifier or uses an executable command inconsistent with its API surface. | Every semantic provider operation has a closed-schema executable mapping or a specific #2985 / reverse-ETL gate; no raw request surface is present. | Operation-policy checkpoint |
| Docs and generated surfaces | Golden/catalog/manual data does not state the 24-operation inventory. | Connector docs/catalog/website/help data agree with final counts and fixture-only, uncertified status. | Parity checkpoint |

All tests remain connector-owned fixtures or local `httptest`; no credentialed checks, live writes, or certification are allowed.

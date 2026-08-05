# TDD ledger: Google Analytics Data API 24-operation parity

| Slice | Red evidence | Green evidence | Commit |
| --- | --- | --- | --- |
| Provider inventory | Official Google reference lists 11 v1beta + 15 v1alpha methods; the existing HOOK ledger contains a non-provider-derived 10/11-method narrative. | `api_surface.json` records exactly 24 semantic operations, their provider method/path, version/revision, retrieval date, and 20-read/4-write split. | `7c550c075` ledger-only checkpoint |
| v1alpha fixed GET reads | Table-driven tests for every added property quota, audience-list, recurring-audience-list, and report-task GET operation fail before dispatch/fixture support exists. | Fixture and `httptest` tests show only fixed GET paths, normalized property IDs, bounded/redacted JSON output, and no secret leakage. | `6daf9e150` alpha direct-read checkpoint |
| POST/read and write policy | `connectorgen validate` initially failed after the ledger named the unrepresented v1alpha operations. | All 24 typed operation records validate; every semantic provider operation has an executable mapping or a specific #2985 / reverse-ETL gate, with no raw request surface. | Implementation checkpoint |
| Docs and generated surfaces | GA command help golden coverage was absent for the new v1alpha command groups. | Connector docs/catalog/website/help data agree with final counts; golden transcripts cover bare GA help, group help, and exact command help. | Implementation checkpoint |

All tests remain connector-owned fixtures or local `httptest`; no credentialed checks, live writes, or certification are allowed.

# Context — DB → API PostgreSQL → GitHub route R1

## Task Delivery Header

- Issue: Refs #4015 — Production MVP — certification.
- Base branch: `integration/4015-mvp-flat-r1` (`2c48e4deb34128339fccbe5d4b7daad4e13a23e7`, confirmed before edits).
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: PR #4186 from `fm/cli-quadrant-db-to-api-postgres-github-r1` targets the exact base. `gh-axi pr list -R polymetrics-ai/cli --state open --base integration/4015-mvp-flat-r1 --head fm/cli-quadrant-db-to-api-postgres-github-r1 --limit 10 --fields url` returned only PR #4186, confirming the API-reported base selection.
- Working branch: `fm/cli-quadrant-db-to-api-postgres-github-r1`.
- Task: Prove through the shipped `pm` binary that PostgreSQL polling-watermark rows cross the durable warehouse and drive GitHub's two declared issue-label actions, with receipt/read-back/checkpoint ordering and real-provider evidence.
- Verification: targeted Go tests (including opt-in PostgreSQL binary integration), `go vet`, build, generated connector checks, lint, all other `make verify` gates individually, a real GitHub read-back in a controlled repository, and API verification of the PR base.

## Locked decisions from the task and definitions

- PostgreSQL remains the existing source adapter: `native_database/postgres_polling_watermark`, source delivery `at_least_once`, `source_ordered`, `deletes: not_available`.
- GitHub remains the only changed destination contract: `declarative_api/issue_label_destination`, acknowledgement `durable_warehouse`, idempotency `keyed`, and only its existing actions may run.
- Mapping is definition-derived, not configuration-invented: PostgreSQL rows must materialize the existing GitHub write-record fields `issue_number` and `labels`; `full_append` selects `add_issue_labels`, while the eligible keyed `incremental_upsert` mode selects `set_issue_labels`. `full_overwrite` remains explicitly ineligible for this API destination protocol.
- No generic HTTP, SQL, shell, action, or record mapping surface will be introduced. The existing set-replace/keyed consent requirements remain in force.
- The execution uses the GSD inline/manual fallback: compatible isolated Pi workers are unavailable and this repository's canonical single-worker contract forbids role spawning. Generated prompts were resolved for `discuss-phase --auto`, `plan-phase --tdd --skip-research`, `execute-phase`, `verify-work`, and `code-review`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or honest boundary |
| --- | --- | --- |
| PostgreSQL rows drive both declaration-owned GitHub label actions through the real `pm` binary | **live** | `TestPMBinaryExecutesLivePostgresWarehouseGitHubIssueLabels` uses a Docker PostgreSQL source and real GitHub HTTPS in retained private `karthik-sivadas/pm-parity-proof-db-to-api`: issue #1 is exactly `pm-db-api-live-add`; issue #2 is exactly `pm-db-api-live-set`. A separate authenticated labels API client reads both results. |
| Source rows cannot bypass durable warehouse / acknowledgement / read-back / checkpoint order | **simulated GitHub boundary** | `TestPMBinaryExecutesPostgresWarehouseGitHubIssueLabels` keeps the CI-safe `httptest` GitHub implementation while using live Docker PostgreSQL. It asserts durable WAL/Parquet receipt artifacts, destination GET read-back after POST/PUT, and checkpoint state. It proves production behavior through the HTTP boundary, not a real GitHub write. |
| Unsupported action, mode, and malformed source row are refused before write I/O | **unit + simulated boundary** | Typed app tests assert unsupported action, ineligible mode, null/malformed row, and `deletes: not_available`; fake provider event counts remain zero. They do not claim live GitHub refusal coverage. |
| Real provider proof is honest and does not expose credentials | **live** | The run exports `POLYMETRICS_GITHUB_TOKEN="$(gh auth token)"` only in the invoking shell; it is never printed, stored, committed, or added to the proof repository. |

## Discussion outcome

The existing GitHub-to-GitHub walking slice is intentionally not reusable unchanged: it requires equal source/destination connectors and compares a configured source issue before writing a configuration-synthesized destination record. That would not prove a database row mapped to an API action. This task replaces only that source-bound assumption with a closed, action-schema-derived warehouse-record mapping for the PostgreSQL source; GitHub's two-action destination, write approval, and warehouse mediation stay intact.

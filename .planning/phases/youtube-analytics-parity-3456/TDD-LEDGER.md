# TDD ledger — YouTube Analytics connector parity (#3456-#3463)

## Red / pre-edit validation

### 2026-07-31 — official operation inventory parity incomplete

Credential-free public discovery re-audit compared the two official discovery documents against existing `internal/connectors/defs/youtube-analytics/api_surface.json`.

Observed red evidence:

```text
RED youtube-analytics official operation inventory parity
official_count 16 local_count 7 missing_count 10 extra_count 1
local_classifiers {'covered_by': 3, 'excluded': 4}
official_methods {'POST': 3, 'PUT': 1, 'GET': 9, 'DELETE': 3}
missing_operations
POST /v2/groups youtubeAnalytics.groups.insert youtube_analytics_discovery_v2
PUT /v2/groups youtubeAnalytics.groups.update youtube_analytics_discovery_v2
GET /v2/groups youtubeAnalytics.groups.list youtube_analytics_discovery_v2
DELETE /v2/groups youtubeAnalytics.groups.delete youtube_analytics_discovery_v2
POST /v2/groupItems youtubeAnalytics.groupItems.insert youtube_analytics_discovery_v2
DELETE /v2/groupItems youtubeAnalytics.groupItems.delete youtube_analytics_discovery_v2
GET /v2/groupItems youtubeAnalytics.groupItems.list youtube_analytics_discovery_v2
GET /v2/reports youtubeAnalytics.reports.query youtube_analytics_discovery_v2
GET /v1/jobs/{jobId} youtubereporting.jobs.get youtube_reporting_discovery_v1
GET /v1/jobs/{jobId}/reports/{reportId} youtubereporting.jobs.reports.get youtube_reporting_discovery_v1
extra_operations
POST https://youtubeanalytics.googleapis.com/v2/reports:query
```

Expected green evidence after implementation:

- `api_surface.json` has 16 rows, no missing/extra official operations, no duplicates.
- Executable coverage rows: 7 ETL streams + 7 reverse-ETL writes.
- Blocked rows: 1 direct/provider query + 1 binary download, each with precise dependency/evidence.
- `capabilities.write` is true only after `writes.json` declares typed closed actions.
- Stream and write fixtures exist for every executable operation; dynamic replay remains honestly skipped if the existing custom OAuth auth hook prevents fixture replay.

## Green / verification log

### 2026-07-31 — implementation parity evidence

Credential-free public discovery comparison after edits:

```text
GREEN? official_count 16 local_count 16 missing 0 extra 0 classes {'covered_by': 14, 'operation': 2}
```

Implemented bundle evidence:

- Streams: `jobs`, `job`, `report_types`, `reports`, `report`, `groups`, `group_items`.
- Writes: `create_job`, `delete_job`, `create_group`, `update_group`, `delete_group`, `create_group_item`, `delete_group_item`.
- Blocked operation metadata: `reports_query` for Analytics `reports.query`; `download_report` for Reporting `media.download` binary bytes.
- Fixture pages/directories were added for all seven streams and all seven write actions; dynamic conformance remains skipped by the bundle-level custom OAuth hook marker, so static bundle validation plus hook tests/conformance skip evidence are the local proof path.

Green local gates recorded:

```text
go run ./cmd/connectorgen validate internal/connectors/defs -> connectorgen validate: 549 connector(s) checked, 0 findings
go test ./internal/connectors/conformance -run 'TestConformance/youtube-analytics' -count=1 -> ok
go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -> ok (golden transcripts updated for new youtube-analytics dynamic command)
go build ./cmd/pm -> ok
make connector-boundary -> outcome clean
make verify -> ok
git diff --check -> ok
```

Note: the requested single-connector spelling `go run ./cmd/connectorgen validate internal/connectors/defs/youtube-analytics` was attempted and currently fails in shared tooling because `connectorgen validate` treats its argument as a definitions root and scans the child directories `fixtures/` and `schemas/` as connector bundles. The repository-root definitions validation used by `make verify` passes with this connector included.

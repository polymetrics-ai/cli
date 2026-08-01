# TDD Ledger: GitLab parity wave02 r1 (#78)

## Red / validation-before-production

Planned failing validation: the pinned official GitLab OpenAPI source has 1,146 unique operations, while the current connector-local `api_surface.json` only has 11 rows and no `operations.json`, `cli_surface.json`, or `certification.json`. This mismatch is the red artifact for the connector-local parity ledger work.

Command to capture before production bundle edits:

```bash
python3 .planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_inventory_check.py
```

Expected initial result: fail with `local api_surface row count 11 != official operation count 1146`.

## Green target

- `internal/connectors/defs/gitlab/api_surface.json` has exactly 1,147 rows: one per official `(method,path)` operation from the pinned OpenAPI source plus one connector-local supplemental GET `/users` stream coverage row.
- `internal/connectors/defs/gitlab/operations.json` has exactly 1,146 typed metadata rows and no executable claims for rows that lack connector-local action/command evidence.
- `internal/connectors/defs/gitlab/cli_surface.json` has connector-local command metadata for implemented streams and planned typed operations; no generic raw API operation is exposed.
- Existing four streams remain fixture-backed and conformance-valid.
- Destructive/admin/DELETE operations are included in the ledger and marked with typed confirmation / blocked-planned evidence instead of being blanket-excluded as unsafe.
- Parent/subissue bodies get one idempotent captain policy addendum; existing bodies and count tables are preserved.

## Refactor / hardening target

- Deterministic generated JSON order: method/path order from the pinned OpenAPI with stable generated IDs.
- Bounded operation metadata: `max_bytes` on direct/binary/read metadata, output policies, risk, approval, mutation class, and source URL evidence.
- Docs explain fixture-only status and shared-foundation dependencies without claiming live certification.

## Evidence log

- [x] Red inventory mismatch captured: `python3 .planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_inventory_check.py` failed before production edits with `local api_surface row count 11 != official operation count 1146` and `local operations.json missing`.
- [x] Green inventory count captured after generation: `python3 .planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_inventory_check.py` passed with `PASS GitLab inventory parity: 1146 official operations plus 1 supplemental stream row represented`.
- [x] Surface count check captured: `python3 .planning/phases/issue-78-gitlab-parity-wave02-r1/traces/gitlab_surface_counts.py` reports 1,146 official operations, 1,147 api/CLI rows, lane counts 398 ETL/read, 637 reverse ETL write, 6 direct/query/search/metadata, 88 binary/file, 15 CDC/changefeed, and 2 excluded/not-applicable.
- [x] `connectorgen validate` result recorded: `go run ./cmd/connectorgen validate internal/connectors/defs --json` passed with 0 findings, including 0 GitLab findings.
- [x] GitLab conformance result recorded: `go test ./internal/connectors/conformance -run 'TestConformance/gitlab' -count=1` passed.
- [x] CLI targeted test result recorded: `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1 -timeout=10m` passed after regenerating GitLab-influenced golden root transcripts.
- [x] `go vet` / `go build` / `connector-boundary` / `git diff --check` result recorded: all passed (`go vet ./internal/connectors/... ./internal/cli/...`, `go build ./cmd/pm`, `make connector-boundary`, `git diff --check`).

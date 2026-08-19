# Verification checklist — YouTube Analytics connector parity (#3456-#3463)

## Required local gates

- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/youtube-analytics`
  - Attempted: shared `connectorgen validate` treats this path as a definitions root and scans `fixtures/` and `schemas/` as bundle directories, producing missing `metadata.json` findings for those child directories.
  - Equivalent repository-root connector validation passed and is the spelling invoked by `make verify`: `go run ./cmd/connectorgen validate internal/connectors/defs` -> 549 connector(s), 0 findings.
- [x] `go test ./internal/connectors/conformance -run 'TestConformance/youtube-analytics' -count=1`
- [x] `go test ./internal/cli -run 'Connector|Dynamic|Golden' -count=1`
- [x] `go build ./cmd/pm`
- [x] `make connector-boundary`
- [x] `make verify`
- [x] `git diff --check`

## Connector parity checks

- [x] Official discovery comparison returns 16 official rows, 16 local rows, 0 missing, 0 extra.
- [x] `api_surface.json` rows classify all operations as covered or blocked-operation; no legacy exclusions for supported official endpoints.
- [x] `streams.json` contains `jobs`, `job`, `report_types`, `reports`, `report`, `groups`, and `group_items`.
- [x] `writes.json` contains `create_job`, `delete_job`, `create_group`, `update_group`, `delete_group`, `create_group_item`, and `delete_group_item`.
- [x] Destructive write actions have required IDs, `body_type: none`, `path_fields`, `redact_fields`, and `confirm: destructive`.
- [x] Write schemas are closed and typed; no raw arbitrary JSON/body generic escape hatch.
- [x] Binary report download is not represented as a JSON stream.
- [x] Provider query and binary download blocked explanations cite concrete engine/runtime dependency.

## Documentation/help parity

- [x] `docs/connectors/youtube-analytics/MANUAL.md` regenerated/updated.
- [x] `docs/connectors/youtube-analytics/SKILL.md` regenerated/updated.
- [x] Connector catalog/generated docs updated as needed.
- [x] Website connector generated data updated as needed.
- [x] CLI golden transcripts updated for the new dynamic connector command listing.

## Issue/status handoff

- [x] Append captain-policy addendum idempotently to GitHub issues #3456-#3463 using `gh-axi`.
- [x] Update `.planning/phases/youtube-analytics-parity-3456/TDD-LEDGER.md` with green evidence.
- [ ] Append final `done: {summary}` to `/Users/karthiksivadas/karthik-agent-workspace/state/cli-youtube-analytics-parity-wave03-r1.status`.
- [x] Local commit created on `fm/cli-youtube-analytics-parity-wave03-r1`.

## Follow-up (resumed session)

- [x] `pnpm run gen:website-data` re-run against a clean install; found `website/data/connectors.generated.json`
      and `website/lib/connectors.catalog.data.generated.json` still carried the pre-parity `docs_md` string for
      `youtube-analytics` (stale relative to `docs.md`). Regenerated and committed
      (`8caa330b8 fix(website): regenerate youtube-analytics catalog docs_md`); re-running the generator
      afterward produces zero diff.

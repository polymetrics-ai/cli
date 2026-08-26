# Verification checklist — #4354

- [x] Candidate-to-main delta contains exactly seven Outreach artifact paths: `api_surface.json`, `cli_surface.json`, `docs.md`, custom-object audit, declaration disposition, source lock, and `sync_transport.json`.
- [ ] **Blocked:** source lock/disposition/API/CLI reconciliation is not validator-admitted. Strict source import rejects `source_kind`; the canonical descriptor and retained artifacts are absent. All 259 candidate rows are retained.
- [x] Six-lane evidence is recorded in `internal/connectors/defs/outreach/sources/outreach-six-lane-evidence.json`: ETL 96, direct-write 163, reverse-ETL 0 provider-class rows/declaration-pending, direct-read 0 N/A, binary download 0 N/A, binary upload 0 N/A.
- [x] Existing source-evidence failure is captured with PR #4350 as the shared owner and no bypass.
- [x] Focused red/green evidence is recorded in `TDD-LEDGER.md`; focused binary tests, generated Outreach skill currentness, and updated root-help golden transcript pass.
- [x] `pm` build and representative valid ETL, typed-write/reverse-ETL label, and destructive delete reach `missing --credential` in an isolated fixture-only project.
- [x] A local `httptest` proof rejects caller `--method`, `--path`, and `--source-url` before a request, then observes exactly `GET /api/v2/prospects` for the valid command.
- [ ] **Blocked:** `connectorgen validate` rejects `source_kind`; global operation evidence reports drift; the certification matrix says Outreach is not allowlisted; and certification sweep is missing. Exact commands/results are in `REVIEW.md`.
- [ ] **Blocked:** website catalog lists Outreach but `cliSurface: null`; no hand-authored `docs/cli` page exists. This must be fixed after source admission.
- [x] The generated Outreach skill and all nine root-help golden forms are current. A complete `go test -timeout 20m ./internal/cli` process completed, and the follow-up cache read returned `ok   polymetrics.ai/internal/cli (cached)`.
- [x] Independent clean-worktree usability proof at detached `1d64e22ce` in `/Users/karthiksivadas/karthik-agent-workspace/worktrees/cli-outreach-pilot-audit-r1`: build, initialized isolated project, and ETL/write/delete no-credential commands pass; the fixture-only binary test passes in 43.036s.
- [ ] Pending final `git diff --check` after recording the audit checkpoint.
- [ ] Pending draft PR base readback/API title/delivery record after push.

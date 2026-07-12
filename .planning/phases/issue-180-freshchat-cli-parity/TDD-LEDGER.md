# TDD Ledger — Issue #180 Freshchat CLI parity parent

## Policy

GSD programming-loop registry command is unavailable (`scripts/gsd: unknown GSD command: programming-loop`), so this parent uses the manual GSD fallback from `.pi/prompts/pm-gsd-loop.md` and `.agents/agentic-delivery/workflows/gsd-universal-runtime-loop.md`. Each behavior-changing sub-issue must create/update its own red/green/refactor ledger before production edits.

## Parent planning checkpoint

- Status: planned.
- Red evidence: not applicable for parent orchestration artifact creation.
- Green evidence: pending commit/push and draft parent PR creation.
- Refactor evidence: not applicable.

## Issue #181 planned TDD

- Red test: add an engine bundle test asserting `freshchat` loads a non-nil command surface from embedded defs; run `go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatCLISurface` and capture failure before adding `cli_surface.json`.
- Green implementation: add Freshchat `cli_surface.json` with safe app-intent mappings; rerun focused test plus `go run ./cmd/connectorgen validate internal/connectors/defs`.
- Refactor: run `gofmt -w` on any edited Go test file and verify no unrelated files changed.

## Later lanes

- #182: help renderer/docs tests must fail first for rendered Freshchat command help/docs before renderer/docs edits.
- #183: stream-runner tests must fail first for missing Freshchat stream command behavior or pagination/cursor gaps.
- #184: operation-ledger validation must fail first for missing/incorrect operation classifications.
- #185: complete; direct-read tests failed first for unsupported Freshchat bounded JSON direct-read policy, then passed after `freshchat_users_fetch` policy/metadata/docs/conformance updates.
- #186: complete; upload operation-policy tests failed first for missing Freshchat typed `file_upload` operation metadata, then passed after operations/docs/fail-closed command wiring.
- #187: complete; confirmation policy tests failed first for missing Freshchat admin/sensitive write confirmations, then passed after metadata/schema/docs updates.

## Parent CodeRabbit review-fix TDD

- Review-red evidence: CodeRabbit parent PR #226 run `321be408-2a1b-4ece-a55c-0e4333fc0b51` reported 11 actionable comments plus one nitpick against the integrated Freshchat range.
- Regression coverage added/updated:
  - `TestRunFreshchatUsersFetchRejectsMoreThanMaxIDs` for the Freshchat 100-id `users/fetch` cap.
  - `TestDirectReadFreshchatUsersFetchPOST` now asserts top-level and nested sensitive response keys are redacted.
  - `TestFreshchatParameterizedReplayFixturesDeclareReadQuery` and `TestReadRawRecordsWithReplayUsesFixtureReadQueryForConfigInput` cover parameterized replay inputs.
  - `TestBundleLoadEmbeddedFreshchatCLISurface` and `TestBundleLoadEmbeddedFreshchatFileUploadOperations` assert `max_items=100` and exact `max_bytes=10485760` metadata.
- Green evidence: focused package tests, full local gates, connectorgen validation, and CLI/docs parity checks pass after review fixes.
- Disposition note: the `api_key`-only help-test suggestion was treated as partially valid because runtime help intentionally names secret fields without values; the test now rejects concrete sample/serialized secret markers while allowing documented `api_key (secret)` field names.

## Parent incremental CodeRabbit follow-up TDD

- Review-red evidence: CodeRabbit incremental run `4fc19851-a5f4-406a-9af5-99fc3c344157` reported one minor and one nitpick finding against the review-fix range: malformed `--credential` flag metadata punctuation and duplicated Freshchat upload `availability=unsupported_local unsupported local workflow` wording.
- Regression coverage added: `TestCommandSurfaceRenderingNormalizesStructuredMetadata` covers structured flag metadata rendering and prevents non-local upload commands from duplicating unsupported-local wording.
- Green evidence: focused tests, docs validation, connectorgen validation, `go vet ./...`, `go test ./...`, `go build ./cmd/pm`, `make verify`, and docs grep all pass after the follow-up fix.

## Parent final verification

Final parent local gates pass after all sub-issue merges:

```bash
cd website && pnpm run gen:website-data
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

## Safety notes

No credentialed Freshchat calls, no reverse ETL execution, no destructive external actions, no dependency changes, and no raw generic write tools are allowed in TDD setup or verification.

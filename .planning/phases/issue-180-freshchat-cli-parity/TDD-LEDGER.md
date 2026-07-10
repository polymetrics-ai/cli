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
- #187: sensitive/admin policy tests must fail first for missing redaction/approval/typed confirmation classifications.

## Safety notes

No credentialed Freshchat calls, no reverse ETL execution, no destructive external actions, no dependency changes, and no raw generic write tools are allowed in TDD setup or verification.

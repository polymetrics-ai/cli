## Intent

Repair the two error-severity crash-consistency defects that shipped in #3882: a reverse-plan
lookup could discard legacy sync-mode normalization, and a newly created local-warehouse path
could acknowledge a durable checkpoint without flushing its newly created parent directory entries.

Refs #3882

## What Changed

- Route both `a.store.Load()` assignments through `normalizeLoadedState`, preserving all state
  compatibility migrations after every reload.
- Before every local-warehouse checkpoint acknowledgement, sync the raw directory, warehouse root,
  and every ancestor through the filesystem volume root, the fixed known-durable boundary.
  The chain is idempotent and does not depend on which writer observed a directory creation, so
  separately rooted projects sharing an allowed external warehouse each establish it themselves.
- Add regression coverage for the state-reload sequence, full observed directory-sync chain per
  writer, and a retry after an injected directory-sync failure.

## State-load audit

| Call site | Verdict |
| --- | --- |
| `internal/app/app.go:188` (`load`) | Normalizes the freshly loaded state. |
| `internal/app/app.go:2035` (`loadReversePlan`) | Normalizes the freshly loaded state before looking up the plan. |

No other `a.store.Load()` call site exists in `internal/app`.

## Why green CI missed this

These defects shipped in #3882 because prior unit coverage did not drive the complete
`Open → RunReverseETL(unknown plan) → RunETL` path from persisted legacy-shaped state, and it did
not observe filesystem directory-sync calls for an absent multi-level warehouse parent chain.
Ordinary happy-path runs remained green.

## TDD and verification

- Both new regression tests were run against unfixed behavior first; their failures are retained in
  `.planning/phases/cli-synccontract-durability-defects-r1/traces/red-run.txt`.
- Focused regression tests, `go test ./internal/app`, `go test ./internal/durability`,
  `go test ./internal/cli`, `go vet ./...`, and `go build ./cmd/pm` pass.
- Passed all separately-run non-suite verification gates: tidy, lint, docs, smoke, agent contract,
  connector generation/metadata, connector boundary, and release workflow.
- The durability tests prove that every writer synchronizes the complete raw-to-volume-root chain
  once before a successful checkpoint acknowledgement, including a retry after a failed chain
  sync. They observe sync coverage rather than emulate a real filesystem power loss; the platform
  primitive remains `durability.SyncDirectory`.

## GSD and skills

Manual inline GSD fallback: generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
`verify-work`, and `code-review` prompts under the single-worker contract. Required skills loaded:
`golang-how-to`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`,
`golang-error-handling`, `golang-security`, and `golang-safety`.

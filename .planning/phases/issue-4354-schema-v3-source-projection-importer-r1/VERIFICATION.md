# Schema-v3 source projection/importer foundation — verification checklist

## Verified implementation snapshot

- Production checkpoint: `b0449adb035475de23f706be7ad6ae889c659bcb`
- Fresh audit target: the same checkpoint, after commit, before publication.
- Usable-surface delta: **0**. The source-reference projection writes no
  `writes.json`, `cli_surface.json`, or `api_surface.json` field and does not
  create a `pm` command. The checked test constructs a matching implemented
  write/command and proves byte-for-byte that source projection leaves it
  unchanged.

### Source → declaration → evidence mapping

| Source form | Shared reader result | Projection behavior | Six-lane evidence |
| --- | --- | --- | --- |
| Retained Outreach schema-v2 declaration form (`source_kind: complete_machine_readable_specification_with_rendered_dynamic_supplement`) | Preserves each operation's provider URL, SHA-256, byte count, source location, method/path, and supplemental Custom Objects identity. It makes no source fetch. | Emits a v3 descriptor with no request, response, pagination, auth, or execution contract and `source_contract_unavailable`; materialization is skipped. | Every lane remains visible. A matching declaration may be `declared`, but none is `enabled`; the row carries `source_contract_unavailable`. |
| Schema-v3 `source_documents[].kind: source_reference` | Uses the same cited-only descriptor constructor and validates source identity/citation shape. | Same non-materializing `source_contract_unavailable` behavior. | Same source-backed row and disabled lanes. |
| Ordinary byte-backed v1/v2/v3 source | Unchanged retained-artifact reader verifies byte/canonical identity before parsing. | Existing fixed declaration projection remains authoritative. | Existing evidence and `missing_foundation` rollups remain unchanged. |

The reference form cannot truthfully name a missing executor shape because it
has no retained request/response contract from which to infer one. Its exact
remaining gap is `source_contract_unavailable`. Verified source contracts keep
their existing named `missing_foundation` handling; this change does not remove,
reclassify, or promote those rows.

## Local gates run

- [x] Red/green source-reference tests — commands and failure/success evidence
  recorded in `TDD-LEDGER.md`.
- [x] `go test -timeout 20m ./cmd/connectorgen -run '<source-reference set>'`
  — pass, including the normal `runSourceImportWithFetcher` write/check path,
  the six-lane evidence projection, malformed kinds, no-fetch assertion, and
  no materialization assertion.
- [x] Strict retained-artifact regression subset — pass:
  digest/byte mismatch, retained-copy mismatch, unsafe artifact destinations,
  and captured rendered-reference citation behavior.
- [x] `go test -timeout 20m ./internal/connectors/engine -count=1` — pass.
- [x] `go test -timeout 20m ./internal/connectors/commandrunner -count=1` — pass.
- [x] `go vet ./...` — pass (including fresh post-commit run).
- [x] `go build ./cmd/connectorgen` — pass.
- [x] `go build ./cmd/pm` through `make docs-check` — pass.
- [x] `make docs-check` — pass.
- [x] `make tidy-check` — pass.
- [x] `make lint` — pass, zero issues.
- [x] `make smoke-no-build` — pass with repository synthetic sample fixture;
  no provider credential or provider call was used.
- [x] `make connector-boundary` — pass.
- [x] `go run ./cmd/agentcontractgen check` — pass.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` —
  `553 connector(s) checked, 0 findings`.
- [x] `go run ./cmd/connectorgen surface-sync --check` —
  `553 connector(s) scanned`, no changes.
- [x] `go run ./cmd/connectorgen operation-evidence --check` — current
  artifact with `1525 rows; 5 rollups; fixed-100 passed`.
- [x] `make release-workflow-check` — completed its release assertions,
  including pin, Homebrew notification, target parity, tooling, size, and
  production-layout checks; the harness did not emit a final aggregate banner.
- [x] `git diff --check` — pass before commit and fresh against the audited
  commit.
- [x] Developer CLI help/doc parity — `go run ./cmd/connectorgen source-import
  --help` includes declaration-only behavior; migration and operation-evidence
  canon document it. `docs/cli/**` and `website/**` have no `source-import`
  user surface, so no user-facing `pm` help/manual/website update applies.

## Explicitly not claimed

- [ ] Full `go test -timeout 20m ./...`: CI authority. The local command is
  intentionally not used as one aggregate per repository guidance.
- [ ] Complete `go test ./cmd/connectorgen` aggregate: started locally, but the
  execution harness returned at its 30-second collection boundary before it
  supplied an exit status. Focused and retained-artifact subsets above passed;
  this aggregate is not claimed as a pass.
- [ ] Live Outreach/provider certification, credential use, or provider calls:
  prohibited and not attempted.
- [ ] A built-binary credential-boundary proof: not applicable because usable
  surface delta is exactly zero.

## Required local evidence

- [x] Focused Red/Green importer and projection Go tests with exact commands.
- [x] Scoped source-import write/check passes through the normal command
  runner without provider access and reports truthful source-backed counts.
- [x] `connectorgen validate`, `surface-sync --check`, and
  `operation-evidence --check` pass. The global generated evidence did not
  need a refresh because no production connector lock was re-pinned or bulk
  remapped in this foundation PR.
- [x] Six-lane mapping was inspected through the checked source-reference
  operation-evidence test; the only new cited-only state is
  `source_contract_unavailable`.
- [x] Usable-surface delta `0` is proven; no command was materialized.
- [x] Formatter, scoped tests, vet, builds, `git diff --check`, docs/lint and
  generator checks are recorded.
- [x] Full-suite and unconclusive aggregate evidence are explicitly listed.
- [x] Independent final-SHA audit records source → descriptor → lane mapping,
  generic-escape review, usable-surface delta, and the exact checked commit.
- [x] PR base read-back: [#4358](https://github.com/polymetrics-ai/cli/pull/4358)
  GitHub API reports `base=main` at
  `b33983927d863032dac8220949990506e812937d`, head branch
  `feat/4354-schema-v3-source-projection`, and creation head
  `cdaf4849eee5da74998dc097eb60d6ba7d81b7cd`.

## Remote review coverage

- PR: [#4358](https://github.com/polymetrics-ai/cli/pull/4358)
- Base/head at creation: `main` / `feat/4354-schema-v3-source-projection`
- Reviewed range requested: `b33983927d863032dac8220949990506e812937d...`
  `cdaf4849eee5da74998dc097eb60d6ba7d81b7cd`
- Primary route: `claude_auto` (non-draft PR opened by repository owner).
- Status: `pending`; no manual Claude command or Copilot fallback has been
  posted. The API currently reports no check run, review, or comment.
- Fallback: none unless the automatic review is skipped, fails, or is
  unavailable; then record the exact blocker before considering one backup
  route.

# Plan — Zoom Healthcare documented-operation parity, R1

## Delivery record

- Parent: [#3915](https://github.com/polymetrics-ai/cli/issues/3915); slice: [#3946](https://github.com/polymetrics-ai/cli/issues/3946).
- Scope owner: `internal/connectors/defs/zoom/**`, Zoom-local tests and fixtures, the Zoom-only generated endpoint projection, Zoom connector docs, and this phase evidence. A separately committed `connectorgen surface-reconcile --notes-contains` foundation is permitted because the existing generator could not scope reconciliation to the ledger's `provider_module=healthcare` provenance without rewriting 838 unrelated Zoom rows.
- Required skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, and `no-mistakes`.
- GSD lifecycle resolution: `scripts/gsd doctor`; `scripts/gsd sources discuss-phase|plan-phase|execute-phase|verify-work|code-review`; `scripts/gsd prompt discuss-phase cli-zoom-parity-healthcare-r1 --auto`; `scripts/gsd prompt plan-phase cli-zoom-parity-healthcare-r1 --tdd --skip-research`; `scripts/gsd prompt execute-phase cli-zoom-parity-healthcare-r1 --interactive`; `scripts/gsd prompt verify-work cli-zoom-parity-healthcare-r1 --auto`; `scripts/gsd prompt code-review cli-zoom-parity-healthcare-r1 --depth=standard`.
- Manual-GSD fallback: `gsd-sdk query init.phase-op cli-zoom-parity-healthcare-r1` returned `phase_found: false` despite `.planning/ROADMAP.md` existing, so official GSD cannot allocate this named provider-module phase. The generated workflows therefore exit before artifact creation. This directory is the required inline single-worker fallback; it preserves the same discuss → plan → execute → verify → review evidence without spawning roles forbidden by the parent contract.
- `go run ./cmd/agentcontractgen check` passed before planning.

## Source audit — completed before RED

The source of truth is Zoom's own Healthcare artifact, not the derived ledger:

| Item | Evidence |
| --- | --- |
| URL | `https://developers.zoom.us/docs/api/healthcare.md` |
| Retrieval | `2026-08-08T08:22:55Z` |
| HTTP / bytes | `200` / `13,783` |
| Artifact | OpenAPI `3.1.1`, API `2`, server `https://api.zoom.us/v2` |
| Reconciled operations | `GET /clinical_notes/notes`, `GET /clinical_notes/notes/{noteId}`, `PATCH /clinical_notes/notes/{noteId}` |
| Ledger delta | `0` — the three `provider_module=healthcare` blocked rows match method, path, title, and source URL exactly |

The prior stopped worker left five uncommitted Healthcare candidates in a different worktree. They
were inspected in place, never copied blindly. The two read operations, PATCH action, sensitive
field policy, and metadata capability change are supported by the live artifact. Two draft details
are rejected: `from` and `to` occur only in the `200` response schema, so they are not command
inputs; and the draft called a nonexistent `verifyWriteCommands` helper. The resumed test will make
the write preflight and 204 behavior explicit.

## Locked implementation decisions

- Implement exactly the Healthcare module's three documented operations: two bounded `direct_read`
  commands and one approval-gated `reverse_etl` action. Nothing is classified `unsafe_or_disallowed`.
- The read response contains clinical note content and EHR identifiers. Use the existing generic
  `clinical_json_redacted` output policy plus explicit `sensitive_policy.redact_fields` for
  `note_content`, the EHR identifiers, and note-owner/last-modifier identifiers. Do not change the
  shared engine policy. `surface-sync` defaults direct reads to `json_redacted`; the stricter
  clinical policy is therefore an intentional declared override, not copied endpoint metadata.
- Expose `--note-owner-user-id` and `--meeting-id` only for the list operation because the live
  operation prose explicitly says they are inputs. Do not expose `--from`, `--to`, paging, or
  `limit` request flags: the artifact presents those only as response fields and the runtime owns
  stream limits.
- Model PATCH as a typed `writes.json` action, `PATCH /clinical_notes/notes/{{ record.note_id }}`
  with the required boolean `is_note_completed`. It remains plan → preview → explicit approval →
  execute; the action is non-destructive but high-risk because it changes a clinical-note status.
- The `204 No Content` response is a successful action, not an omission. A Zoom-local fixture test
  will assert the exact PATCH path and JSON body and that the engine accepts the `204` status with
  one record written.
- Derivable `cli_surface.json` metadata, `api_surface.json` coverage state, and
  `internal/connectors/defs/operation_endpoint_ledger.json` are generated through
  `go run ./cmd/connectorgen surface-sync`; no generated output is hand-edited. The two direct-read
  ledger rows are generated with `surface-reconcile --notes-contains provider_module=healthcare`.
  The PATCH row's `covered_by.write` is the source ledger's required typed-action linkage; the
  reconciler intentionally owns direct-read rows only.

## Execution slices

1. **Plan checkpoint (this commit)** — preserve artifact evidence, source audit, scope fence,
   manual-GSD fallback, and the red/green plan.
2. **RED checkpoint** — change only Zoom's operation/reachability test and this TDD ledger:
   expected covered count `9 → 12`, locally-blocked count `1833 → 1830`, direct-read count
   `6 → 8`, and one reverse-ETL write. Add Healthcare fixture-execution assertions. Run the test
   against the unmodified bundle, capture its literal failure, commit and push before any bundle or
   documentation production edit.
3. **GREEN checkpoint** — author `operations.json`, `writes.json`, `cli_surface.json`, local
   fixtures, `metadata.json`, and `docs.md`; run `surface-sync` for derivable command metadata and
   the scoped reconciler for direct-read coverage/projection;
   update the scoped Zoom docs and any Zoom-only golden transcript delta. Run the targeted tests,
   validator, and binary help/live-reachability checks. Commit and push only after green.
4. **Verify/review checkpoint** — execute the generated GSD verify-work/code-review intent inline,
   record every command result and review disposition in this phase. Then update #3946 and #3915
   with the completed module count and exact resumption point for the next provider category.

## CLI help/manual/website parity

- [x] `pm help zoom`, `pm zoom`, `pm zoom healthcare`, and every new command's `--help` checked via the built binary.
- [x] Invalid/missing required flag behavior remains a usage error; the bare namespace remains contextual help and exits successfully.
- [x] `docs/connectors/zoom/{MANUAL.md,SKILL.md}`, `docs/connectors/README.md`, and generated/golden CLI output are updated only when the live generator shows Zoom-specific required drift.
- [x] `pm docs validate --connectors-dir docs/connectors` passes; no website-specific Zoom page exists, so `website/**` is not applicable.

## Verification plan

- RED: `go test -count=1 ./internal/connectors/defs/zoom/...` must fail before production edits.
- GREEN: focused Zoom test, `go run ./cmd/connectorgen surface-sync --check`, connector validation,
  conformance, commandrunner, and `internal/cli` tests; `go vet ./...`; then build `./cmd/pm`.
- Binary: run help for all three new routes and a synthetic-token request for each direct read
  against the live Zoom host, expecting Zoom's own structured `401`, never a local unknown-command
  error. The write command is verified through plan/preview only; it is never executed against
  Zoom.
- Scope: inspect `git diff --name-only` and the endpoint-ledger diff; both must be Zoom-only plus
  declared phase/docs artifacts. Zero `unsafe_or_disallowed` rows and zero invented paging flags.

## Canonical references

- `AGENTS.md`
- `docs/migration/HANDOFF-CODEX.md`
- `docs/migration/conventions.md`
- `docs/architecture/connector-architecture-v2-design.md`
- `.agents/agentic-delivery/contracts/issue-agent-contract.md`
- `.agents/agentic-delivery/contracts/parent-orchestrator-contract.md`
- `.agents/agentic-delivery/references/gsd-pi-adapter.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `.planning/phases/cli-zoom-parity-wave2-qss-r1/`
- `internal/connectors/defs/zoom/api_surface.json`
- `https://developers.zoom.us/docs/api/healthcare.md`

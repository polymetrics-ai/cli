# TDD Ledger — Gong CLI Parity Parent (#133)

## GSD/TDD setup

- `scripts/gsd doctor`: pass.
- `scripts/gsd verify-pi`: pass.
- `scripts/gsd list --json`: pass.
- `scripts/gsd prompt plan-phase 133 --skip-research --tdd`: prompt rendered.
- `scripts/gsd prompt programming-loop init --phase issue-133-gong-cli-parity --dry-run`: failed with `scripts/gsd: unknown GSD command: programming-loop`.
- Manual-GSD fallback recorded: use `/pm-orchestrate` and `/pm-gsd-loop` prompt bodies plus `gsd-universal-runtime-loop.md`; do not skip red/green/refactor evidence.

## Parent red/green strategy

Parent #133 is orchestration scope. Each sub-issue owns its behavior tests.

| Lane | Red evidence owner | Green evidence target |
|---|---|---|
| #144 operation ledger | `go test ./cmd/connectorgen -run GongAPISurfaceOperationLedger -count=1` fails against current 10-entry surface | exact 67-operation ledger, no legacy exclusions |
| #141 CLI surface metadata | metadata validation/help discovery test fails until `cli_surface.json` exists | validated Gong CLI command surface metadata |
| #142 help renderer/docs | help/docs parity test fails until metadata renders | runtime help/docs/website parity |
| #143 stream runner | stream-backed command test fails until runner handles Gong streams | stream commands execute through generic runner |
| #145 direct reads | bounded direct-read tests fail until operations are implemented | max-bytes/redaction/output-policy enforced |
| #146 body/binary engine | body/binary shape tests fail until fixed schemas and bounds exist | no raw JSON/body escape hatch; bounded binary policy |
| #147 sensitive/admin policy | preview/redaction/confirmation tests fail until policy exists | typed reverse-ETL writes with risk/redaction/confirmation |

## Current red slice

Active local critical path: #144. See `.planning/phases/issue-144-gong-operation-ledger/TDD-LEDGER.md`.

#144 status: red captured, green targeted checks passed. Full parent verification remains pending.

## Refactor notes

- Parent orchestration docs may change as sub-issues land.
- Operation-ledger rows remain metadata-only until executor lanes implement typed surfaces.


## 2026-07-10 integrated lanes #141/#142/#143/#145/#146/#147

Red:
- `go test ./internal/connectors/engine -run 'TestDirectReadJSONRedactedPolicy|TestDirectReadAvoidsDoubleVersionPrefix' -count=1` failed because `json_redacted` was unsupported.
- `go test ./cmd/connectorgen -run 'TestGongFullSurface|TestGongMetadata' -count=1` failed because Gong had no `cli_surface.json`, no `writes.json`, no `operations.json`, and `metadata.capabilities.write=false`.

Green:
- Added generic recursive JSON redaction and version-prefix-safe direct reads.
- Added Gong CLI surface, streams, direct reads, write actions, operation metadata, docs, and website catalog updates.
- Targeted verification passed (see VERIFICATION.md).

Refactor:
- Kept unsupported multipart/top-level-array payloads as typed blocked operation metadata with bounded schemas/policies instead of adding a generic raw upload/write path.

## 2026-07-10 engine-shape issues #252/#253/#254

Red:
- #252: operation direct-read POST JSON body construction, schema validation, output redaction, commandrunner typed body mappings, and validator safety tests were added before engine/runner changes.
- #253: top-level JSON array write body shape and schema-validation tests were added before write-body implementation.
- #254: stdlib multipart request construction, retry-safe file reopening, local path/size safety, engine multipart writes, and preview redaction tests were added before multipart implementation.

Green:
- Added schema-gated `OperationDirectReader` execution for typed GET/POST `rest_read` operations and commandrunner `body.*` flag mapping without raw body flags.
- Added `json_array` and `multipart` write body modes with meta-schema/validator coverage.
- Added reverse-plan payload identity binding for local file uploads (path hash plus size/mtime) so file changes invalidate approvals before execution.
- Implemented safe Gong commands for meetings integration status, flow steps/prospects POST reads, call media upload, CRM entities upload, and typed CRM entity-schema array writes.

Refactor:
- Remaining broad filter-shaped POST reads stay planned until safe typed filter flags are authored; blocker text now reflects command-specific mapping work rather than missing engine support.
- Multipart and top-level array support remain reverse-ETL/write-record scoped; no generic upload/raw JSON command was introduced.

## 2026-07-22 completion cycle — planned red/green ledger

Execution decision: `not_spawned_runtime_capability_missing`; implementation proceeds as coupled `local_critical_path` in the parent checkout.

Red targets:

- CLI: `pm help gong`, bare `pm gong`, and `pm gong --help` currently fail instead of rendering connector help.
- Approval integrity: same-size file content can change while size/mtime are restored, leaving the prior payload identity unchanged.
- Multipart bound: file growth after preflight `stat` can exceed `MultipartFile.MaxBytes` because the stream uses unbounded `io.Copy`.
- Upload TOCTOU: execute-time approval hashing and later multipart file opening are separate operations, so a same-size mutation in that gap could be sent unless the transport snapshots and compares the approved SHA-256 before opening the HTTP request.
- Gong completion: 10 POST direct-read commands, including `calls transcript`, remain `planned` and their API ledger rows remain blocked operations.

Green targets:

- Dynamic connector help resolves before app/project initialization and bare connector namespaces exit 0.
- Payload identity includes a content digest; execution propagates approved digests by record/field; multipart snapshots and verifies approved bytes before network send, copies at most `MaxBytes+1`, and fails closed before accepting an altered or oversized body.
- All 10 POST read-query commands have typed connector-authored flags, bounded `json_redacted` output, and `availability: implemented`; no raw JSON body flag exists.
- Gong API ledger reports 67 executable classifications and zero blocked operation rows.

Red captured 2026-07-22:

- `go test ./internal/cli -run TestDynamicConnectorHelpAndBareNamespace -count=1` — failed: `help topic "gong" not found`, bare namespace `missing connector command path`, transcript command still `availability=planned`.
- `go test ./internal/app -run TestPayloadIdentitiesBindSameSizeContentWhenMTimeIsRestored -count=1` — build failed because `PayloadIdentity.ContentSHA256` does not exist.
- `go test ./internal/connectors/connsdk -run TestRequesterDoMultipartRejectsGrowthAfterPreflightValidation -count=1` — failed: `DoMultipart error = nil, want stream-time max-bytes rejection`.
- `go test ./internal/connectors/connsdk -run TestRequesterDoMultipartRejectsChangedApprovedContentBeforeSend -count=1` — red captured: `unknown field ExpectedSHA256 in struct literal of type MultipartFile`; the test also requires zero HTTP calls on mismatch.
- Local Codex review found the CLI pre-clamped operation reads to 1 MiB and treated a flag value equal to `help` as help. `go test ./internal/cli -run 'GongTranscriptCommandAllowsDeclaredResponseCap|GitHubCommandSurfaceRunsDirectReadFile' -count=1` captured both reds: transcript rejected at 1,048,577 bytes and `--path help` sent no request. Both are now green.
- Final schema/example review found newly generated examples that omitted required typed POST filters. Generator red: `TestGongFullSurfaceCommandAndOperationCoverage` failed because the scorecard example omitted `--scorecard-id`; examples were corrected and the test is green.
- Pushed-head Security CI red: `govulncheck` reports GO-2026-5970 through existing indirect `golang.org/x/text` v0.36.0, fixed in v0.39.0, via the PostgreSQL connector. This is not introduced by Gong code, but it blocks readiness; use the smallest existing dependency upgrade and rerun security/full gates.
- Local Codex's legacy-plan compatibility finding is intentionally not changed: pre-digest upload approvals must fail closed after upgrade because their approved hash did not bind file content. The short-lived plan is invalidated rather than weakening the new upload guarantee.
- Final local Codex commit review produced three actionable parity findings: Gong advertised `--approve` as boolean instead of a token string, `pm help gong --json` emitted plain text, and flag-only `pm gong --credential ...` returned a missing-path error. Focused tests reproduced all three failures; `--approve` is now a string token, dynamic help uses the JSON manual envelope, and recognized flag-only namespaces render contextual help.
- Follow-up review probe showed `pm gong --bogus` also rendered help successfully. Red captured; unknown flags now return usage errors. The completed review additionally identified incomplete action-only flags (`--plan ... --preview`) as silent help; that red is now green. Mutation/plan lifecycle flags without a command path fail, while passive declared flags such as `--credential` retain contextual help.
- `go test ./cmd/connectorgen -run 'TestGongAPISurfaceOperationLedger|TestGongFullSurfaceCommandAndOperationCoverage' -count=1` — failed: 57/67 covered and 19/29 direct reads; 10 operation blockers remain.

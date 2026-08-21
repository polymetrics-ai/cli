# Verification: Gong release-0.3.0 live parity reconciliation

## Credential-free verification

- [x] Isolated worktree and preserved remote branch identity verified.
- [x] `no-mistakes doctor` passed; no daemon action was taken.
- [x] `scripts/gsd doctor` passed and command sources were resolved.
- [x] Current official OpenAPI refetch returned 69 operations with the current GET/POST/PUT/PATCH/DELETE distribution and a current strict source lock.
- [x] Exact method/path/operation-ID/deprecation comparison against the refreshed Gong source lock passed; the canonical inventory fingerprint is recorded in `SOURCE-AUDIT.md`.
- [x] Current foundation parent `c3f83cbf6eabbae00219566fb02719ca2d6c480d` and Batch 2/3 reconciliation completed without rewriting preserved history.
- [ ] `go run ./cmd/connectorgen source-import gong --check` is blocked by the provider-neutral artifact URL policy: Gong's official fixed source requires `?version=`, while the importer rejects query-bearing artifact URLs and the query-free route is 404.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/gong --json` remains blocked only by that same source-importer URL-policy dependency.
- [ ] Scoped `surface-sync --check` remains pending after source-import support accepts the fixed official URL. The unscoped command also has unrelated Aircall source-projection drift and is not claimed as Gong evidence.
- [x] Full direct-read reconciliation: all 30 implemented direct-read commands ran through the built binary in a fresh initialized project with no credential and each reached `missing --credential`; none was unknown or exact-endpoint blocked.
- [x] Focused Gong full-surface, commandrunner, and multipart conformance tests pass with `-timeout 20m`; the three multipart actions use the merged generic approval-digest path.
- [x] Batch 2/3 parity-map verification passes for 19 connectors / 5,127 documented operations; its regenerated foundation ledger contains zero Gong gap rows.
- [x] Built `pm` credential-free command sweep classified all 69 implemented Gong paths (30 direct reads, 27 reverse-ETL writes, 12 ETL streams) as `missing --credential`; it made no provider request and saw zero unknown, partial, or unbound results.
- [x] `pm help gong`, `pm gong`, and `pm gong calls get --help` render contextual help; manual, skill, and website generated artifacts were regenerated after the declaration changes.
- [ ] `go vet ./...`, `go build ./cmd/pm`, individual `make verify` static gates, and detached `make connector-boundary` pass.
- [ ] `go test -timeout 20m ./...` and `make verify` are attempted with a non-cutoff runner or truthfully left to CI.
- [ ] Inline code review is recorded in `REVIEW.md`; automated-review route/dispositions are recorded in PR #3552.

## Captain hard certification gate

- [x] All 69 official operations have an exact source-lock, enabled disposition, declaration/API-surface,
  and generated-CLI mapping; no provider-defined operation is disabled for scope, tier, safety, or
  destructive classification.
- [ ] Every enabled supported operation is reachable through the built CLI, persisted App path,
  runtime help/manual, and website projection. The built CLI/help/manual/website portion is green;
  persisted App-path live certification remains credential-blocked. Typed confirmation and approval
  guard writes; they do not reduce reachability.
- [x] ETL reconciliation is proven credential-free: all 12 declared stream commands reach the built
  binary's credential preflight and have exact stream/API bindings.
- [ ] Reverse-ETL reconciliation is proven through declaration-selected target/action mappings,
  plan, preview, explicit approval, apply, acknowledgement, and provider readback—or exact-source
  `not_applicable` evidence is recorded.
- [ ] Direct-read and direct-write reconciliation is proven against the real installed command
  paths at credential preflight. Live pagination, provider required-input behavior, and mutation
  readback remain in the disposable-credential stage.
- [x] Binary-download is exact-source `not_applicable`: every official Gong response contract is
  JSON or wildcard response metadata, with no binary response operation. Binary-upload has three
  exact multipart operations and focused generic conformance evidence.
- [x] Provider output-preservation evidence forbids read-field redaction declarations and stale
  redaction language; ordinary provider fields are retained and only configured credential values
  are masked with an explicit marker.
- [ ] Live certification uses the persisted App path with an approved non-echoing disposable
  credential reference, supported CRUD/application commands, cleanup, and bounded non-secret
  request/result fingerprints.
- [ ] No merge-ready claim appears in PR #3552 until every applicable item above is green.
- [x] Captain missing-foundation ledger is generated and drift-checked from Batch 2/3 source maps:
  `.planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/missing-foundation-gaps.json`.
  Gong has zero remaining rows; unrelated portfolio foundation gaps remain open.

## Live certification hold

No approved disposable Gong credential reference was provided or discovered. The live stage is
therefore intentionally not run. Required eventual evidence is: persisted App-path credential
use; reads, writes, application commands, pagination, required-input behavior, ETL,
plan/preview/approval/apply/readback reverse ETL, binary routes if declared, representative CRUD
with cleanup, and bounded non-secret request/result fingerprints. No browser session may replace
connector authentication.

## Accepted shared dependency

The remaining credential-free dependency is the provider-neutral `connectorgen source-import`
artifact URL policy. It must support the official, fixed, query-bearing Gong source URL without
opening arbitrary query input. No connector-specific importer bypass is present. The live
certification credential-reference gate remains independent.

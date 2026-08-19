# Verification: Gong release-0.3.0 live parity reconciliation

## Credential-free verification

- [x] Isolated worktree and preserved remote branch identity verified.
- [x] `no-mistakes doctor` passed; no daemon action was taken.
- [x] `scripts/gsd doctor` passed and command sources were resolved.
- [x] Current official OpenAPI refetch returned 69 operations with the current GET/POST/PUT/PATCH/DELETE distribution.
- [x] Exact method/path/operation-ID/deprecation comparison against Batch 2/3 Gong source lock passed.
- [x] Current-main, typed-destination, and Batch 2/3 foundation reconciliation completed.
- [x] `go run ./cmd/connectorgen validate --json` passes after reconciliation.
- [x] `go run ./cmd/connectorgen surface-sync --check` passes after reconciliation.
- [ ] Full direct-read reconciliation remains pending. The bundle-directory `surface-reconcile` invocation is not a valid scoped proof, so its zero-scan result is not claimed as evidence.
- [x] Focused Gong `connectorgen`, commandrunner, and conformance tests pass with `-timeout 20m`; multipart replay is explicitly excluded pending the shared F2/F4 foundation.
- [ ] Built `pm` credential-free direct-read sweep records every result classification and proves no `unknown command` or exact-endpoint preflight block.
- [ ] `pm help gong`, `pm gong`, and affected command help/docs/website generated-artifact checks pass.
- [ ] `go vet ./...`, `go build ./cmd/pm`, individual `make verify` static gates, and detached `make connector-boundary` pass.
- [ ] `go test -timeout 20m ./...` and `make verify` are attempted with a non-cutoff runner or truthfully left to CI.
- [ ] Inline code review is recorded in `REVIEW.md`; automated-review route/dispositions are recorded in PR #3552.

## Captain hard certification gate

- [ ] All 69 official operations have an exact source-lock, disposition, declaration/API-surface,
  and generated-CLI mapping; no provider-defined operation is disabled for scope, tier, safety, or
  destructive classification.
- [ ] Every enabled supported operation is reachable through the built CLI, persisted App path,
  runtime help/manual, and website projection. Typed confirmation and approval guard writes;
  they do not reduce reachability.
- [ ] ETL reconciliation is proven, or the official source audit records exact `not_applicable`
  evidence.
- [ ] Reverse-ETL reconciliation is proven through declaration-selected target/action mappings,
  plan, preview, explicit approval, apply, acknowledgement, and provider readback—or exact-source
  `not_applicable` evidence is recorded.
- [ ] Direct-read and direct-write reconciliation is proven against the real installed command
  paths, including pagination and required-input behavior.
- [ ] Binary-download and binary-upload reconciliation is proven through the closed operation
  runtime, or exact-source `not_applicable` evidence is recorded. Multipart upload remains blocked
  on the provider-neutral F2/F4 digest-binding foundation and is not claimed before it publishes.
- [ ] Provider output-preservation evidence retains every ordinary non-secret response field and
  uses an explicit marker where mandatory credential/transport-secret masking applies.
- [ ] Live certification uses the persisted App path with an approved non-echoing disposable
  credential reference, supported CRUD/application commands, cleanup, and bounded non-secret
  request/result fingerprints.
- [ ] No merge-ready claim appears in PR #3552 until every applicable item above is green.

## Live certification hold

No approved disposable Gong credential reference was provided or discovered. The live stage is
therefore intentionally not run. Required eventual evidence is: persisted App-path credential
use; reads, writes, application commands, pagination, required-input behavior, ETL,
plan/preview/approval/apply/readback reverse ETL, binary routes if declared, representative CRUD
with cleanup, and bounded non-secret request/result fingerprints. No browser session may replace
connector authentication.

## Accepted shared dependency

`cli-closed-operation-runtime-r1` owns the provider-neutral F2/F4 requirement to bind synthetic
fixture multipart payloads to the approval digests that runtime requires before dispatch. Gong
does not ship a bypass and does not claim multipart conformance until that published generic head
is integrated. The connector-local path and schema corrections are independently covered.

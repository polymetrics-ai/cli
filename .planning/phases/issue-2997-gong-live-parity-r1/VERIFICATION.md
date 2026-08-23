# Verification: Gong release-0.3.0 live parity reconciliation

## Credential-free verification

- [x] Isolated worktree and preserved remote branch identity verified.
- [x] `no-mistakes doctor` passed; no daemon action was taken.
- [x] `scripts/gsd doctor` passed and command sources were resolved.
- [x] Current official OpenAPI refetch returned 69 operations with the current GET/POST/PUT/PATCH/DELETE distribution and a current strict source lock.
- [x] Exact method/path/operation-ID/deprecation comparison against the refreshed Gong source lock passed; the canonical inventory fingerprint is recorded in `SOURCE-AUDIT.md`.
- [x] `origin/main` through `8127de418` is merged without rewriting preserved history; shared/runtime
  and unrelated connector conflicts resolve to main, while the PR diff remains Gong-owned.
- [x] PR #4335 merged at `8127de418`; the Gong source lock is a v3 `gong-v2` document with the
  exact fixed `?version=` query marked `identity_query: true`. The real scoped importer gets past
  URL validation and parses the official source.
- [ ] `go run ./cmd/connectorgen source-import gong --check` now stops at the provider-neutral
  source-import preflight for `GET /v2/all-permission-profiles` parameter 0:
  `unbounded request schema string has no maxLength`. It must retain or type-gap this ordinary
  provider input before descriptor projection; no Gong-local maximum/bypass is valid.
- [ ] `go run ./cmd/connectorgen validate internal/connectors/defs/gong --json` reports only the
  resulting missing `sources/gong-operation-descriptor.json` canonical descriptor.
- [ ] Scoped `surface-sync --check` remains pending on that missing canonical descriptor. The
  unscoped command also has unrelated Aircall source-projection drift and is not claimed as Gong
  evidence.
- [x] Full direct-read reconciliation: all 30 implemented direct-read commands ran through the built binary in a fresh initialized project with no credential and each reached `missing --credential`; none was unknown or exact-endpoint blocked.
- [x] Focused Gong full-surface, commandrunner, and multipart conformance tests pass with `-timeout 20m`; the three multipart actions use the merged generic approval-digest path.
- [x] On the branch merged through `8127de418`, credential-free focused checks passed again:
  `TestGongFullSurfaceCommandAndOperationCoverage`, `TestGongMetadataEnablesWriteCapability`,
  `TestGongCertificationDeclarationUsesOnlyOrdinaryRESTLiveCandidates`, the real
  `TestEveryImplementedCommandPassesRuntimePreflight`, and `TestConformance/gong`.
- [x] Batch 2/3 parity-map verification passes for 19 connectors / 5,127 documented operations; its regenerated foundation ledger contains zero Gong gap rows.
- [x] Built `pm` credential-free command sweep classified all 69 implemented Gong paths (30 direct reads, 27 reverse-ETL writes, 12 ETL streams) as `missing --credential`; it made no provider request and saw zero unknown, partial, or unbound results.
- [x] `pm help gong`, `pm gong`, and `pm gong calls get --help` render contextual help; manual, skill, and website generated artifacts were regenerated after the declaration changes.
- [x] Gong parameter import reconciled the three multipart operations (`17` scanned, `0` remaining drift); Gong certification candidates and the 71-row certification sweep are generated and current.
- [x] `TestGongFullSurfaceCommandAndOperationCoverage` is strengthened for the captain's collision
  policy. It first failed on three stale direct-read descriptions, then passed after all 30 direct
  reads and metadata declared preservation of an ordinary credential-equal provider value.
- [x] `go test -timeout 20m ./cmd/connectorgen -count=1` passed (146.952s); `make docs-check`,
  `make connector-boundary` (552 connectors, zero findings), `make lint` (zero issues),
  `go vet ./...`, and current Gong certification candidate/sweep/subject checks passed after
  regeneration.
- [x] `go vet ./...` and `go build ./cmd/pm` pass.
- [x] `tidy-check`, `docs-check`, `smoke-no-build`, `lint`, `agent-contract-check`,
  `connector-boundary`, and `release-workflow-check` pass. Boundary scanned 552 connectors with
  zero findings.
- [ ] `go test -timeout 20m ./...` completed but failed in unrelated
  `TestSkillsGenerateMatchesTrackedSkills` generated-skill drift (the failure names `pm-100ms`,
  not Gong). The test output also contained repeated unavailable local Redis endpoint diagnostics.
  No unrelated generated skills or runtime infrastructure were changed in this Gong branch.
- [ ] `make verify` was attempted. It reaches formatting, tidy, vet, then fails at the same
  repository-wide generated-skill drift before later gates; no Gong files outside this task were
  altered to mask that baseline failure.
- [ ] `make connectorgen-validate` is blocked by the resulting missing Gong source descriptor
  after generic source-import common-input preflight. `make connectorgen-surface-sync` is blocked
  by that same missing descriptor. `certification-matrix --connector gong --check` is blocked because Gong is
  not in the shared static allowlist; no allowlist change was made. Gong certification candidates,
  71-row sweep, and regenerated current subject all pass their `--check` modes.
- [ ] Latest scoped `go run ./cmd/connectorgen source-import gong --check` fetches the declared
  v3 identity-query document and returns only `GET /v2/all-permission-profiles` parameter 0
  `unbounded request schema string has no maxLength`. This is a provider-neutral descriptor-stage
  gap; it does not reopen the query policy or license a Gong-specific workaround.
- [ ] Inline code review is recorded in `REVIEW.md`; automated-review route/dispositions are recorded in PR #3552.

## Captain hard certification gate

- [x] All 69 official operations have an exact source-lock, enabled disposition, declaration/API-surface,
  and generated-CLI mapping; no provider-defined operation is disabled for scope, tier, safety, or
  destructive classification.
- [ ] Every enabled supported operation is reachable through the built CLI, persisted App path,
  runtime help/manual, and website projection. The built CLI/help/manual/website portion is green;
  the persisted App path now proves authentication, one bounded ETL read, one bounded typed direct
  read, input validation, and pagination validation. Five ETL cells, direct `targets list`
  required-input projection, and live writes remain uncertified. Typed confirmation and approval
  guard writes; they do not reduce reachability.
- [x] ETL reconciliation is proven credential-free: all 12 declared stream commands reach the built
  binary's credential preflight and have exact stream/API bindings.
- [ ] Reverse-ETL reconciliation is proven through declaration-selected target/action mappings,
  plan, preview, explicit approval, apply, acknowledgement, and provider readback—or exact-source
  `not_applicable` evidence is recorded.
- [ ] Direct-read and direct-write reconciliation is proven against the real installed command
  paths at credential preflight. The disposable credential stage now proves one ordinary typed
  direct read and cursor/required-input guards, but `targets list` required-input projection and
  operation-specific live coverage remain incomplete. Mutation readback remains unexecuted because
  no self-cleaning declaration-owned Gong pairing exists.
- [x] Binary-download is exact-source `not_applicable`: every official Gong response contract is
  JSON or wildcard response metadata, with no binary response operation. Binary-upload has three
  exact multipart operations and focused generic conformance evidence.
- [ ] Provider output-preservation declarations forbid Gong read-field redaction, and require all
  ordinary values—including an undeclared value that equals a credential—to be retained. Shared
  runtime collision masking still violates that rule; foundation issue #4321 owns the
  provider-neutral red/green fix. Explicitly declared secret fields remain maskable with a marker.
- [ ] Live certification uses the persisted App path with an approved non-echoing disposable
  credential reference, supported CRUD/application commands, cleanup, and bounded non-secret
  request/result fingerprints. The approved reference is now in use and scoped read proof is
  recorded, but full parity is blocked by the five ETL cells, source-projection dependency,
  unpaired writes, and paid-agentic exclusion described in `SOURCE-AUDIT.md`.
- [ ] No merge-ready claim appears in PR #3552 until every applicable item above is green.
- [x] Captain missing-foundation ledger is generated and drift-checked from Batch 2/3 source maps:
  `.planning/phases/issue-4289-parity-map-batches-2-3-r1/traces/missing-foundation-gaps.json`.
  Gong has zero remaining rows; unrelated portfolio foundation gaps remain open.

## Live certification boundary

The captain supplied an approved non-echoing disposable credential reference. The live stage now
uses only the built CLI, persisted App path, and repository certification harness. Required
evidence is: ordinary-REST authentication; reads, applicable writes/application commands,
pagination, required-input behavior, ETL, plan/preview/approval/apply/readback reverse ETL,
supported binary routes, representative CRUD with cleanup, and bounded non-secret
request/result fingerprints. No browser session may replace connector authentication. No Gong
agentic endpoint may be sent because it consumes paid credits; any mandatory such cell remains
uncertified pending a captain decision.

## Accepted shared dependency

The remaining credential-free dependency is provider-neutral `connectorgen source-import` common
input preflight: it must retain or type-gap an ordinary provider string without a declared
`maxLength` before descriptor projection. Query identity is now declaration-bound and exercised
through #4335's merged v3 contract; no connector-specific importer bypass or invented Gong input
bound is present. Separately, #4337 blocks a new proof-producing live run because the current
external-proof serializer would retain the account-scoped base URL argument verbatim.

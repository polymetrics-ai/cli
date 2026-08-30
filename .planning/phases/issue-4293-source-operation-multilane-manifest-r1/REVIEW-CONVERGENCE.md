# Review convergence — Issue #4293 source-operation multi-lane manifest

## Task Delivery Header

- Issue: Refs #4293 — Batch R1 source-operation multi-lane manifest and validator
- Base branch: `fix/4293-source-operation-multilane-manifest-r1` reviewed as immutable commit `27608b31ed0f3b138fe6218188ca02a084b4d8eb`
- Merges into: `fix/4293-source-operation-multilane-manifest-r1` → its existing parent branch → `main`; this review branch does not merge or cherry-pick.
- Delivery: Commit and push this frozen review ledger only after two independent Codex review passes; post the review proof on #4293. No PR or merge.
- Working branch: `codex/4293-mapping-review-r1`
- Task: Independently review mapping controls, source-lock lineage, all seven lane cells, schema/admission, artifacts, GraphQL root-field handling, CLI behavior, negative tests, and reachability without modifying the reviewed implementation.
- Verification: Frozen-SHA metadata; immutable-delta inspection; two separated Codex passes; targeted manifest/schema/CLI tests; applicable check commands; broad-suite failures recorded separately; staged `git diff --check`.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Review is bound to one immutable implementation | live | The target SHA, parent delta, merge base, worktree branch, changed-file inventory, and generated-file status are recorded before code discovery. |
| Every mandatory review lens has independent coverage | live | Two read-only Codex passes each record lens evidence, reachability traces, test evidence, and disposition against the same target SHA. |
| No live provider behavior is overstated | fake | Review-only task: local tests and structural checks are used because no credentialed provider I/O is authorized. Each result is recorded as static or local-test evidence, not certification. |

## Freeze record

- Immutable review SHA: `27608b31ed0f3b138fe6218188ca02a084b4d8eb`
- Reviewed branch at freeze: `fix/4293-source-operation-multilane-manifest-r1`
- Review worktree branch: `codex/4293-mapping-review-r1`
- Direct parent / review delta base: `dc481bac1a8b78d60ac0b4a2b0dfd1a9068ce8db`
- `main` merge base: `813f457a925f7ee3fe3bea101a43e445992c8552`
- Target subject: `fix(connectors): validate source operation lane mappings`
- Implementation delta (13 files / 1,300 additions): `cmd/connectorgen/{main.go,sourceoperationmapping.go,sourceoperationmapping_test.go}`, engine mapping loader/meta-schema/schema, connector-canon evidence documentation, and existing #4293 planning artifacts. No generated artifact change is in the direct target delta.
- Generated-file status: no generated file is changed in the direct parent → target delta.
- Immutable-tree rule: only this review artifact may change on the review branch. No mapping code, source lock, connector definition, runtime, or executor file is edited.

## Contracts and review method

- Issue #4293 is the source-accounting/mapping-layer contract. Runtime admission or execution claims are excluded except to trace reachability and prevent false promotion.
- `$firstmate-exhaustive-review` governs the frozen ledger, architecture map, mandatory lenses, findings, and fresh-context re-review. The captain explicitly substituted two fresh-context Codex passes for the skill's Claude audit; no Claude audit is requested.
- `$connector-lane-build-order` governs source lock → mapping/projection → definition artifacts → execution-witness separation and the seven-lane terminology.
- The skill-linked `docs/connector-terminology.md` is absent at this SHA. The review uses `docs/connector-canon/{INDEX,OPERATION-EVIDENCE,DECLARATION-ADMISSION}.md` as the current repository canon and records the absent path as a documentation fallback, not a finding against the reviewed implementation.
- Required Go skills are recorded in the final pass evidence: `golang-how-to`, `golang-security`, `golang-safety`, `golang-error-handling`, `golang-lint`, `golang-testing`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-graphql`, `golang-context`, and `golang-concurrency` as applicable to the reviewed paths.

## Pass 1 — frozen discovery

Status: complete — Codex context A, immutable target only. The provisional
finding set below is frozen before the separate Pass 2 re-review.

### Architecture and reachability map

| Surface | Exact target path | Reachability / boundary result |
| --- | --- | --- |
| Developer CLI | `cmd/connectorgen/main.go` → `runSourceOperationMapping` | The command is reachable only as `connectorgen source-operation-mapping <manifest> --check`; it is not a `pm` runtime command. |
| Mapping check | `cmd/connectorgen/sourceoperationmapping.go` | Reads the chosen manifest, source locks confined beneath that manifest directory, and the existing mapping-only declaration-admission parser. It never calls connector execution, credential, transport, source retention, or certification code. |
| Schema admission | `internal/connectors/engine/{bundle.go,metaschemas.go,schema/source_operation_mapping.schema.json}` | A closed JSON schema is embedded and then strict JSON decoding rejects unknown/duplicate members. This is authoring-only validation, not runtime admission. |
| Source cohort | `data/connector-canon/batch1-source-rigidity-r2-cohort-ledger.json` | The existing immutable cohort lists ten source locks and 4,341 REST source identities. The target neither adds a mapping manifest nor reads this ledger. |
| Existing projection | `data/connector-canon/batch1-source-rigidity-r2-operation-evidence.json` | It retains 4,341 rows but exposes only six classifications (`binary_download`, `binary_upload`, `direct_read`, `direct_write`, `etl`, `reverse_etl`), has no `sync_transport`, and is not a recognized input to the new command. |

### Provisional blockers

| ID | Severity | Exact evidence | Violated contract and impact | Proposed regression / disposition |
| --- | --- | --- | --- | --- |
| MAP-001 | blocker | The direct target delta adds only the checker, schema, tests, docs, and planning artifacts. `git ls-tree` finds no checked-in source-operation mapping manifest; `runSourceOperationMapping` accepts any caller-supplied path. The cohort ledger fixes ten locks / 4,341 identities while its `totals.projected_cells` is `0`. | #4293 requires a manifest that reconciles every Batch R1 ID exactly once and a deterministic check for that frozen denominator. At this SHA no command input contains that cohort, no expected ten-lock membership or 4,341 count is asserted, and no parent-denominator proof can run. | Add a tracked, source-lock-bound Batch R1 manifest (or an equivalently immutable cohort input) with all ten locks, 4,341 cited rows, and its required lane cells; make `--check` bind to that cohort or an equally fixed source of expected membership/count. Do not repair in this review. |
| MAP-002 | blocker | `sourceOperationMappingPathCheck` records a `hasETL` boolean and only rejects a non-`none` manifest pagination fact without `etl` (`sourceoperationmapping.go:271-291`). `sourceOperationMappingCellFindings` validates state/reason syntax only (`:466-507`); it never uses the locked HTTP method to require `direct_write` and `reverse_etl`, and schema `cells` has no minimum. | #4293 acceptance criterion 3 requires provider writes to retain independently explicit direct-write and reverse-ETL cells; criterion 6 requires missing applicable cells to fail. A locked `POST`/`PUT`/`PATCH`/`DELETE` operation can pass with no cells (or one only) as long as it declares pagination `none`. | Add red cases for every mutating method with either write lane removed and enforce the required independent dispositions from the locked operation identity. Keep it mapping-only; this does not require runtime behavior. |
| MAP-003 | blocker | The manifest fact model has no record-shape field (`sourceoperationmapping.go:46-68`). The mapping-only lock projection retains protocol, URL, location, provider ID, method, and path (`declarationadmission.go:306-318`), but not pagination, record shape, media, scope, or event facts. Fact checking only compares each fact citation URL with the source URL and checks nonempty values (`sourceoperationmapping.go:438-463`); it does not bind fact location or content to a locked source node. | #4293 requires cited pagination **and record shape**, path/scope, media, and event/cursor facts that determine applicability. A row can assert `pagination: none` or `event_cursor: none` with an arbitrary same-document locator, avoiding the only ETL requirement and any future sync/binary applicability assertion. This is not source-lock-bound validation. | Preserve/parse the applicable locked source facts or use a strict cited source-fact sidecar with exact node bindings; add red tests for false pagination/event/media facts and missing record shape. |
| MAP-004 | blocker | Artifact validation checks only a syntactically canonical relative string and the forward source-ID/lane reference (`sourceoperationmapping.go:306-314,549-554`). It neither resolves a file nor checks connector ownership/backlinks; `artifacts: []` and empty artifact `cells` are schema-valid. | #4293 requires each mapped cell to link to its connector-definition artifacts and rejects orphan/output-only projections. At this SHA a non-existent or unrelated canonical path can pass, and all mapping cells can have no artifact link at all. | Resolve artifact paths under the manifest/repository definition root, require regular connector-owned targets as appropriate, and verify required forward/reverse cell coverage with missing/nonexistent/unrelated/empty-artifact red cases. |

### Pass-1 lens ledger

| Mandatory lens | Status | Evidence / result |
| --- | --- | --- |
| Architecture/data flow | complete | Mapped CLI → manifest checker → schema/parser and confirmed no runtime/credential/provider-I/O branch. |
| Happy/bad/edge behavior | complete | Existing focused cases cover duplicate ID, missing row, ETL removal, orphan cell link, traversal, state-reason, supplemental lineage, and GraphQL mismatch; they omit cohort, write-lane, false-fact, record-shape, and artifact-target coverage identified above. |
| State machine/concurrency | not applicable | Check-only local file validation has no durable store, goroutine, retry, lease, callback, or transaction state. |
| Security/secret taint | complete | Source-lock containment resolves symlinks and rejects out-of-root locks; no credentials or provider requests are read. Artifact labels are not resolved, which is captured by MAP-004 rather than treated as a provider-I/O issue. |
| Retry/rate-limit/resume/idempotency | not applicable | The mapping checker has no network send, retry, cursor persistence, or write execution. |
| Output integrity | complete | Findings are sorted before output; successful summary derives counts locally. There is no generated manifest/report byte contract, which contributes to MAP-001 rather than a runtime-output claim. |
| Declaration reachability/closed surface | complete | CLI and embedded schema are reachable, but the checked-in Batch R1 cohort has no reachable manifest input; MAP-001 blocks the claimed control. |
| CLI/App parity | complete | Developer `connectorgen` usage/dispatch is present; no `pm`, App, credential, manual, completion, or JSON surface is in scope for this authoring-only command. |
| Provider semantics | complete | Canonical grouping protects GraphQL root identity; source facts that decide pagination/media/event semantics are only asserted, not validated (MAP-003). |
| Tests/evidence | complete | The target test file is synthetic-only and does not exercise the Batch R1 input, missing mutation lanes, fact truth, record shape, or actual artifact resolution. |

### Pass-1 local command evidence

- `scripts/gsd doctor`: pass.
- `scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}` and their generated prompts: inspected; manual inline review fallback recorded because this is a frozen review and the captain prohibited a Claude audit.
- `jq empty internal/connectors/engine/schema/source_operation_mapping.schema.json`: pass.
- Scoped Go commands were started while unrelated repository-wide `go test ./...` jobs held the shared Go workload; they were terminated before completion and are deliberately recorded as pending rather than green. They will be retried with captured completion after Pass 2/static review work.

## Pass 2 — fresh-context re-review

Status: complete — Codex context B, same immutable SHA
`27608b31ed0f3b138fe6218188ca02a084b4d8eb`, begun after the Pass-1 record
above was written. This pass independently re-read the issue contract, direct
delta, parser/schema, focused tests, cohort ledger, and target issue comments.

### Independent re-review result

- **MAP-001 confirmed.** The target issue's own completion-proof comment says
  that populating and invoking manifests for the ten Batch R1 locks is
  “remaining work” outside its reusable slice. That is incompatible with #4293
  itself, which owns the ten-connector source-accounting layer. The frozen
  cohort ledger contains the precise expected denominator (4,341), but the
  command has no call path to it and no tracked manifest can be invoked.
- **MAP-002 confirmed.** The sole semantic required-cell rule reads a
  manifest-provided pagination value. The checker has the locked protocol,
  method, and GraphQL provider root identity in memory, but never evaluates a
  mutating REST method or `Mutation.*` GraphQL root to demand both independent
  write cells. The target's synthetic `POST` fixture supplies both cells, so it
  demonstrates representability rather than refusal of a removed write lane.
- **MAP-003 confirmed, narrowed.** The data model cannot represent the
  required record-shape fact at all. Fact citations are structurally present,
  but mapping-only parsing retains no source-fact payload to check whether a
  claimed pagination/media/event fact reflects the pinned source. This makes
  the missing record-shape contract and the existing ETL applicability check
  insufficient for the issue's stated source-bound control.
- **MAP-004 withdrawn as a blocker.** The issue's narrower acceptance wording
  requires an artifact reference to resolve to an existing source-operation and
  lane cell, which the target does enforce. The target deliberately treats its
  artifact path as a non-dereferenced label; `..` is rejected, and there is no
  artifact file read to traverse. Lack of reverse/physical artifact validation
  is a future aggregate-manifest design question, not a confirmed defect in
  this reusable checker alone.

### Pass-2 lens ledger

| Lens | Independent result |
| --- | --- |
| Source-lock parsing / lineage | Pass for the reusable parser boundary: canonical relative lock paths, symlink containment, regular-file check, strict source-lock parsing, and supplemental same-route canonical lineage are present. It does not cure MAP-001's absent cohort artifact. |
| Canonical grouping / GraphQL roots | Pass: canonical representative must be self-canonical, same connector/protocol/method/path, with `ProviderOperationID` equality for GraphQL; focused test distinguishes `Query.widget` from `Mutation.widget`. |
| Seven lane cells | Blocked by MAP-002/MAP-003: the schema represents all seven but semantic admission only requires ETL from a self-reported pagination field, not source writes or other source-determined applicability. |
| Schema/admission | Pass for closed field/lane/state syntax and strict duplicate-member handling; blocked at required fact content/record-shape semantics by MAP-003. |
| Artifacts / traversal | Pass for the implemented reference-only contract: canonical label plus existing source-cell target and no `..` traversal. No broader artifact-file claim is attributed to this target. |
| CLI behavior / reachability | Pass for developer-only check-only dispatch/help and no runtime/credential route. Blocked for Batch R1 control reachability because no tracked cohort manifest is wired (MAP-001). |
| Negative tests | Blocked: no regression covers full-cohort absence, deleted direct-write/reverse-ETL cells, missing record shape, or contradictory source facts. |
| Runtime / Foundation Atlas | No runtime foundation demand. The target calls the existing authoring parser/schema seam only; no executor, transport, credential, source-retention, or certification implementation is reached. |

### Final finding set for the immutable target

| ID | Status | Classification |
| --- | --- | --- |
| MAP-001 | open blocker | missing mapping/control integration; no runtime foundation or code repair is authorized in this review. |
| MAP-002 | open blocker | missing mapping/control admission rule; no runtime foundation. |
| MAP-003 | open blocker | missing source-fact/model coverage; no runtime foundation. |
| MAP-004 | withdrawn | artifact path/backlink behavior is sufficient for the target's reference-only contract. |

## Broad-suite boundaries

Status: recorded separately; not a cause of the three mapping-control blockers.

- The target's own `VERIFICATION.md` records its pre-existing broad command
  result: `go test -timeout 20m ./cmd/connectorgen -count=1` failed after
  622.602 seconds in six existing Batch R1 parity tests:
  `TestImplementedCommandEndpointEquivalenceCoversExactFleet`,
  `TestOperationEvidenceGitLabSourceLockBridge`,
  `TestRetainedAsanaSourceImportRejectsReadProjectionDrift`,
  `TestRetainedAsanaMutationDispositionsCoverEveryDeferredSourceOperation`,
  `TestSourceProjectionGapCreatesCommandFromExistingClosedActionVariant`, and
  `TestSourceProjectionSourceCitedMutationDispositionLeavesExistingProjectionByteIdentical`.
- The same target record reports existing
  `operation-evidence --check` generated-artifact drift and 50 existing
  connector-definition validation findings. None was created or repaired on
  this review branch.
- This reviewer started the focused command with an isolated Go cache while
  unrelated repository-wide `go test ./...` workloads were compiling. It was
  intentionally cancelled before a result to avoid competing with the captain's
  Batch R1 work. It is **not a green result** and is not used to support any
  finding; the source/test inspection above is the review evidence.

## Final verification ledger

| Command / check | Result | Scope |
| --- | --- | --- |
| `scripts/gsd doctor` | pass | Repository GSD adapter and review workflow inputs. |
| `scripts/gsd sources …` plus generated phase prompts | inspected | Manual inline review fallback recorded; captain explicitly required Codex-only review rather than the skill's Claude audit. |
| `jq empty internal/connectors/engine/schema/source_operation_mapping.schema.json` | pass | Added schema is syntactically valid JSON. |
| `git diff --check` | pass | Review-artifact worktree diff has no whitespace errors. |
| Frozen direct-delta / tree / cohort inspection | pass as evidence | Exact target remains unchanged; it proves MAP-001/002/003 as documented. |
| `go test -timeout 20m ./cmd/connectorgen -run '^TestSourceOperationMapping' -count=1` | cancelled before completion | Do not treat as green; author-recorded focused green remains non-independent. |

## Verdict and required remediation

**BLOCK — no merge or parent integration claim for the #4293 acceptance
criteria at the frozen SHA.** The implementation is a reusable checker slice,
but it is not the source-operation multi-lane manifest/control required by the
issue.

1. **MAP-001 — establish the fixed Batch R1 denominator.** Add a tracked,
   canonical manifest or equally immutable cohort input that enumerates the
   ten locked connectors (`asana`, `bitbucket`, `circleci`, `dockerhub`,
   `gitlab`, `jira`, `notion`, `sentry`, `stripe`, `vercel`) and all 4,341
   source IDs. Bind `--check` to exact expected source membership/counts (not
   merely a caller-selected subset), preserve every row through unavailable
   overlays, and make the full input runnable in CI/review.
2. **MAP-002 — enforce source-backed write dispositions.** For every provider
   mutation (including a GraphQL `Mutation.*` root), require independently
   explicit `direct_write` and `reverse_etl` cells. The determination must be
   bound to source identity/facts rather than relying on the author to omit a
   cell; add red cases for each removed cell and a non-mutating `POST` boundary
   if the provider source distinguishes it.
3. **MAP-003 — model and bind applicability facts.** Add the missing
   record-shape fact and preserve exact source-node bindings for pagination,
   record shape, scope/path variables, media, and event/cursor evidence. Make
   the checker reject a contradictory/uncited source fact and use the validated
   facts to require ETL, sync, and binary dispositions where applicable.

These are mapping/control changes only. No shared runtime foundation, executor,
transport, credential path, source-retention behavior, or Go runtime repair is
authorized or required by this review.

---
reviewed_sha: 9e5329f34e015e39160bb8e951452bbd071a698a
depth: deep
scope_count: 22
status: issues_found
finding_count: 10
---

# Orchestration Deep Review

**Reviewed:** 2026-08-20T17:37:18Z
**Snapshot:** `9e5329f34e015e39160bb8e951452bbd071a698a`
**Disposition:** 4 BLOCKER, 6 WARNING
**Method:** immutable, discovery-only review. `HEAD` matched the requested SHA and the worktree was clean before this ledger was created. CodeGraph was attempted first, but this worktree has no `.codegraph/` index, so call paths were followed with read-only source inspection and focused tests.

## Scope

All 22 requested files were read in full:

1. `.planning/phases/issue-4305-rest-structured-body-r1/RUN-STATE.json`
2. `.planning/phases/synctransport-4303-destination-r1/RUN-STATE.json`
3. `data/cli-current-foundations-main-integration-r1/evidence-manifest.json`
4. `data/cli-current-foundations-main-integration-r1/input-manifest.json`
5. `internal/connectors/sync_transport.go`
6. `internal/connectors/sync_transport_test.go`
7. `internal/connectors/transportpolicy/policy.go`
8. `internal/connectors/write_result_output_test.go`
9. `internal/coordination/durable_store.go`
10. `internal/coordination/rate_parking.go`
11. `internal/coordination/rate_parking_test.go`
12. `internal/synccontract/commit.go`
13. `internal/synctransport/arrow_fast_path_controller.go`
14. `internal/synctransport/arrow_fast_path_pipeline.go`
15. `internal/synctransport/orchestrator.go`
16. `internal/synctransport/registry.go`
17. `internal/synctransport/types.go`
18. `website/content/docs/cli-reference.mdx`
19. `website/content/docs/etl.mdx`
20. `website/lib/docs.generated.ts`
21. `website/scripts/gen-docs-data.mjs`
22. `website/scripts/gen-docs-data.test.mjs`

Call paths outside the 22-file set were followed read-only where needed to verify composition, App dispatch and state CAS, rate-limit resumption, operation executors, CLI exit codes, generated surfaces, and tests.

## State and Call-Flow Map

```text
connector definitions
  -> DefinitionFactoriesFromRegistry
  -> RegisterDeclaredTransports (validate all; build; atomic batch registration)
  -> Registry.Preflight (source/destination/mode/stream/action/evidence/executor)
       ! delivery guarantees are not checked for cross-end compatibility [ORCH-B04]

App.RunETL
  -> special-route selection
       ! selected preflight errors collapse to false [ORCH-W07]
  -> Orchestrator.Run
  -> Destination.PlanDestination
  -> Source.ReadTransport(page)
  -> validate page/checkpoint/tombstones
  -> approval AuthorizeNextUnit
  -> durable warehouse Stage -> receipt Validate -> independent Reopen
  -> Destination.ApplyDestination -> provider durable acknowledgement/output
  -> independent destination ReadBack
  -> CommitAfterDownstreamAcknowledgement
  -> App stream-state CAS
       ! CAS is after provider mutation/readback [ORCH-B01]
  -> retire warehouse receipt

full_overwrite
  -> Begin private session
  -> repeat source -> approve -> stage -> reopen -> ApplyFullOverwrite
  -> approve publication -> PublishFullOverwrite -> ReadBackFullOverwrite
       ! readback failure invokes pre-publication Abort after publication [ORCH-B02]
       ! publication output is not copied into Result [ORCH-B03]
  -> checkpoint/App CAS -> receipt retirement
       ! losing stale writer may already have replaced the live target [ORCH-B01]

Arrow full_overwrite (serial or ordered pipeline)
  -> Begin Arrow session -> extract/transform/stage/COPY segments
  -> PublishArrowFullOverwrite -> readback -> checkpoint/App CAS
       ! same abort, result-loss, and late-CAS defects [ORCH-B01/B02/B03]

rate limit
  -> source 429 after last committed checkpoint
  -> persist ParkedRateLimitRun -> timer
  -> Claim(run ID) -> renew lease -> App resume from checkpoint
  -> Rearm on a new typed 429, or Complete on success
       ! transient coordinator/store failures lose their timer/retry path [ORCH-W05]
       ! isolation/claims are per run ID, not provider scope [ORCH-W06]
       ! malformed claim parameters can persist an unloadable file [ORCH-W10]
```

## Six-Surface Foundation Coverage

| Surface | Reachability / evidence | Verdict |
|---|---|---|
| ETL | Definition-owned sources and managed PostgreSQL target are registered without per-connector App edits; focused source-import, registration, preflight, App, and command tests pass. The live execution path still admits incompatible delivery declarations and mutates before the stale-state CAS. | **BLOCKED** by ORCH-B01 and ORCH-B04; route errors can also be obscured by ORCH-W07. |
| Reverse ETL | Multiple persisted declarative actions reach plan, preview, approval, apply, durable acknowledgement, provider readback, run persistence, and CLI output on the ordinary page path. Full-overwrite/Arrow output is lost and stale writers can duplicate or overwrite provider state. | **BLOCKED** by ORCH-B01 and ORCH-B03. |
| Direct read | `commandrunner` and source-import focused suites passed. Implemented `direct_read` is admitted by the command runner and no silent missing mapping was found in the reviewed composition. | **PASS** for the reviewed foundation. |
| Direct write | Structured REST body, exact query/path/header binding, form, GraphQL, SCIM, and declaration-bounded multipart direct writes have executable paths and focused tests. Successful ordinary provider output remains byte-complete while synthetic diagnostics are redacted. | **PASS** for declared `rest_write`/`graphql_mutation`; orchestration defects above apply if the action is used as a sync destination. |
| Binary download | `binary_download` has an executor source (`internal/connectors/engine/operation_kind.go:20`), commandrunner tests pass, and the input manifest records the qualified HEAD 200/404, exact text/binary, and no-artifact missing-path proof (`data/cli-current-foundations-main-integration-r1/input-manifest.json:50-65`). | **PASS**. |
| Binary upload | Multipart upload is available only through bounded `rest_write`. The distinct `file_upload` operation kind is declared (`internal/connectors/engine/operation_kind.go:22`) but intentionally has no certified executor annotation (`cmd/connectorgen/certificationmatrix_test.go:53-74`); the commandrunner returns an explicit “executor is not implemented” error (`internal/connectors/commandrunner/runner.go:419-429`), and the loader test prevents legacy `file_upload` from masquerading as the REST executor (`internal/connectors/engine/operation_multipart_test.go:271-276`). | **PARTIAL, EXPLICIT GAP**. The missing foundation is reported rather than silently unreachable. |

## Findings

### ORCH-B01 — Provider side effects precede the only stale-writer CAS

**Severity:** BLOCKER
**Evidence:**

- The ordinary page path performs provider apply and independent readback at `internal/synctransport/orchestrator.go:233-274`, then invokes `request.Commit` at `internal/synctransport/orchestrator.go:278-285`.
- Full overwrite publishes and reads back at `internal/synctransport/orchestrator.go:472-483`, then commits at `internal/synctransport/orchestrator.go:489-495`.
- Arrow serial publishes/readbacks before commit at `internal/synctransport/arrow_fast_path_controller.go:184-216`; the pipeline does the same at `internal/synctransport/arrow_fast_path_pipeline.go:125-157`.
- The App commit callback's only writer-conflict guard is the stream-state comparison at `internal/app/transport_dispatch.go:252-278`, after those side effects.
- Existing tests codify this order: a losing full-overwrite run has already applied and published (`internal/app/transport_dispatch_test.go:1383-1392`), and a losing page writer has already applied two pages before the state conflict (`internal/app/transport_dispatch_test.go:1629-1635`).
- Descriptors may validly declare no idempotency at all (`internal/connectors/sync_transport.go:50-55`), so replay is not necessarily safe.

**Impact:** Two App instances can read the same prior stream state, both mutate the provider, and only then discover which checkpoint lost. Append/non-idempotent actions duplicate writes. More seriously, a losing full-overwrite publisher can replace the live target after the winning publisher, fail its checkpoint CAS, and leave provider contents from the loser while durable stream state points to the winner. A crash after apply/publish but before CAS creates the same replay window.

**Root cause:** The durable stream state is used as an optimistic post-write CAS, not as pre-write execution ownership. The destination API carries neither a durable fencing generation nor an enforced idempotency identity tied to the warehouse receipt.

**Exact change plan:**

1. Add a durable, stream-scoped execution lease with a monotonically increasing fencing generation, acquired atomically against the expected stream state before source or provider I/O.
2. Carry the fence and stable warehouse receipt/idempotency key through every destination apply/publish API; destinations must reject stale fences.
3. Renew/reconcile the lease during long reads and Arrow COPY without holding the JSON/file lock across provider I/O.
4. Make checkpoint CAS validate and retire that same fence. On restart, reconcile an outstanding receipt/publication before any replay.
5. Refuse replay-capable modes when the destination cannot provide keyed idempotency or fencing (also ORCH-B04).

**Required tests:**

- **Happy:** one writer acquires the fence, applies exactly once, readbacks, commits, and retires the lease/receipt.
- **Bad:** two independently opened Apps start from the same checkpoint; the second is rejected before destination apply/publish, with exactly one provider mutation.
- **Edge:** cover expired lease takeover, cancellation while renewing, crash after apply-before-checkpoint, restart reconciliation, and a late full-overwrite publisher; assert provider contents and durable checkpoint always identify the same fenced generation.

### ORCH-B02 — A readback failure calls a pre-publication abort after publication succeeded

**Severity:** BLOCKER
**Evidence:**

- The interface explicitly defines `AbortFullOverwrite` as cleanup for *pre-publication* exits at `internal/synctransport/types.go:184-195`.
- The standard lifecycle leaves `published=false` in its defer at `internal/synctransport/orchestrator.go:352-361`; it sets the flag only after both `PublishFullOverwrite` and `ReadBackFullOverwrite` succeed at `internal/synctransport/orchestrator.go:474-485`.
- Arrow serial has the same state machine at `internal/synctransport/arrow_fast_path_controller.go:71-80` and `:186-204`.
- Arrow pipeline repeats it at `internal/synctransport/arrow_fast_path_pipeline.go:65-75` and `:127-145`.

**Impact:** If publication succeeds but receipt readback fails or times out, the deferred abort is invoked against an already-published session in violation of the interface contract. A destination that implements abort as shadow cleanup/rollback can delete, revert, or otherwise disturb the newly live target. Even when a current implementation happens to make abort close-only, the generic contract makes the orchestration unsafe for the next conforming destination.

**Root cause:** One boolean conflates “publication has occurred” with “publication was independently verified.”

**Exact change plan:** Split the lifecycle into `publicationAttempted`, `publishSucceeded`, and `readBackSucceeded`. Disarm pre-publication abort immediately after a successful publish. Treat readback failure after publication as an ambiguous/published reconciliation state, persist the acknowledgement/receipt, and resolve it through readback/reconciliation without invoking pre-publication cleanup.

**Required tests:**

- **Happy:** publish + readback succeeds; neither normal return nor defer calls abort.
- **Bad:** publish succeeds and readback returns an error; abort count remains zero and the publication receipt is retained for reconciliation.
- **Edge:** publish returns an error before durable publication (abort once), context expires between publish and readback, and restart completes readback without republishing; run all cases for standard, Arrow serial, and Arrow pipeline implementations.

### ORCH-B03 — Full-overwrite and Arrow publication outputs never reach run results

**Severity:** BLOCKER
**Evidence:**

- The public result contract requires every provider-returned response field, receipt, status, body, occurrence ID, and credential-equal byte to be retained (`internal/synctransport/types.go:516-545`, especially `:538-543`).
- The ordinary page path appends error output and successful acknowledgement output at `internal/synctransport/orchestrator.go:237-248`.
- Full overwrite calls `ApplyFullOverwrite` without any output channel at `internal/synctransport/orchestrator.go:436-451`, then receives a publication acknowledgement at `:474-483` but never appends its `Output` to `result.DestinationResults` before returning at `:504`.
- Arrow serial drops the publication acknowledgement output at `internal/synctransport/arrow_fast_path_controller.go:186-223`; Arrow pipeline drops it at `internal/synctransport/arrow_fast_path_pipeline.go:127-164`.
- App persists and returns only what the transport result contains (`internal/app/transport_dispatch.go:286-341`), so the omission cannot be recovered downstream.

**Impact:** Provider-generated IDs, statuses, headers, bodies, receipts, and occurrence identifiers disappear from both immediate and durable run output whenever a destination uses full overwrite or the binary fast path. Output is also unavailable when a full-overwrite apply/publish error carries a provider result. This violates the explicit ordinary-output completeness contract and makes status/readback consumers unable to audit the actual provider operation.

**Root cause:** Result collection was implemented only around `DestinationExecutor.ApplyDestination`; the session-oriented publication APIs were not wired into the shared output path, and `ApplyFullOverwrite` has no typed result return.

**Exact change plan:** Introduce one defensive-copy result collector used by ordinary apply, full-overwrite apply/publish, and Arrow publish. Extend session APIs (or a typed error-output interface) so apply failures can return provider output. Append acknowledgement output before readback/checkpoint so a later failure does not erase provider evidence. Preserve provider bytes verbatim; continue sanitizing only system-generated diagnostics.

**Required tests:**

- **Happy:** standard full overwrite, Arrow serial, and Arrow pipeline each return and persist an acknowledgement containing status, headers, body, provider ID, and occurrence ID.
- **Bad:** apply or publish returns a typed provider result with an error; the result is retained while the printable error remains credential-safe.
- **Edge:** readback and checkpoint failures after successful publication still retain output; credential-equal strings, numeric JSON, base64-equivalent bytes, duplicate headers, empty body, and unknown provider fields are byte-for-byte unchanged.

### ORCH-B04 — Delivery guarantees are declared but not enforced

**Severity:** BLOCKER
**Evidence:**

- `DeliveryGuarantees` declares idempotency, ordering, and delete behavior at `internal/connectors/sync_transport.go:41-70`.
- Validation checks only that each field is a known enum at `internal/connectors/sync_transport.go:296-312`; destination validation at `:360-423` does not relate those guarantees to modes, strategies, acknowledgement, or replay.
- Registry preflight checks descriptors, mode, stream/source binding, durable acknowledgement, action, executors, and conformance at `internal/synctransport/registry.go:131-218`, but never compares source and destination delivery guarantees.
- The ordinary orchestrator validates tombstone shape at `internal/synctransport/orchestrator.go:149-159` and passes the page onward; it does not reject a tombstone when either descriptor declares deletes unavailable.
- The execution architecture explicitly checkpoints after provider acknowledgement (`internal/synctransport/orchestrator.go:233-285`), yet `DeliveryIdempotencyNone` remains an admissible declaration (`internal/connectors/sync_transport.go:53-55`).

**Impact:** A transport can preflight with a source that emits tombstones and a destination that declares `deletes=not_available`, allowing deletes to be silently omitted or misapplied. Likewise, non-idempotent destinations are admitted to a restart/stale-writer path that can replay writes. Ordering declarations are descriptive rather than enforceable, so a mode can claim source ordering without a compatible destination path.

**Root cause:** The schema treats delivery properties as documentation/evidence metadata. There is no mode/strategy compatibility matrix and no per-page assertion that runtime behavior matches the declarations.

**Exact change plan:** Add an explicit delivery compatibility policy to `Registry.Preflight` keyed by mode and apply strategy. Reject replay-capable operation when the destination lacks keyed/fenced idempotency; reject incompatible ordering; reject delete-requiring modes/bindings when either side cannot support tombstones. At runtime, fail before staging/destination I/O if a source declaring deletes unavailable emits a tombstone or if a tombstone reaches a destination that cannot apply it.

**Required tests:**

- **Happy:** compatible keyed, ordered, tombstone-capable source/destination pairs run normally.
- **Bad:** `none` idempotency in a replayable route, unordered delivery in an ordered route, and a tombstone source paired to `not_available` destination all fail preflight with zero source/provider I/O.
- **Edge:** a source whose descriptor says `not_available` nevertheless emits a valid tombstone; runtime refuses before stage/apply. Exercise every canonical mode and apply strategy, including restart and deferred-checkpoint pages.

### ORCH-W05 — Transient rate-parking failures remove the only retry timer

**Severity:** WARNING
**Evidence:**

- `resumeDue` deletes the timer and, on a `Claim` error, only removes the in-memory resume lease before returning (`internal/coordination/rate_parking.go:600-621`).
- Claim loss, resume failure, ignored `ReleaseClaim` failure, and `Complete` failure all return without scheduling reconciliation at `internal/coordination/rate_parking.go:650-670`.
- Only the separate `rearmPending` branch schedules again (`internal/coordination/rate_parking.go:684-694`).
- The failure test asserts that the parked record remains after one failed resume (`internal/coordination/rate_parking_test.go:587-665`) but does not require a timer, bounded retry, or surfaced failure event.

**Impact:** A transient store/claim failure strands a due run until process restart. A successful provider resume followed by `Complete` failure is worse: this coordinator stops tracking progress while another coordinator or restart may later replay the provider work after the claim expires. The only emitted lifecycle events are park/resume; operators receive no durable explanation for the stuck record.

**Root cause:** `resumeDue` treats all errors as terminal but has no terminal failure state, retry class, reconciliation task, or event. The code cannot distinguish a safe pre-provider coordination failure from an ambiguous post-provider failure.

**Exact change plan:** Classify resume phases. Apply bounded backoff/retry only to pre-provider claim/store failures. Persist a reconciliation-needed state after provider success/ambiguous completion, so completing the parking record never replays provider I/O. Check and surface `ReleaseClaim` errors. Emit secret-safe typed failure/reconciliation events. Schedule at `max(now+backoff, resetAt, claimUntil)` and stop timers on `Close`.

**Required tests:**

- **Happy:** one claim/resume/complete removes the record and emits one resumed event.
- **Bad:** `Claim` fails once then succeeds on a bounded retry; `Complete` fails after a successful App resume and reconciliation removes the record without a second provider call; `ReleaseClaim` failure is observable.
- **Edge:** claim loss during renewal, cancellation/Close, restart during reconciliation, repeated store failure with a bounded ceiling, and no zero-delay retry loop.

### ORCH-W06 — Same-scope parked runs resume concurrently

**Severity:** WARNING
**Evidence:**

- Memory `Create` is unique only by `RunID` (`internal/coordination/rate_parking.go:110-126`); file `Create` is also keyed only by `run.RunID` (`internal/coordination/durable_store.go:194-215`).
- `Start` and `Park` schedule each persisted run independently (`internal/coordination/rate_parking.go:386-410` and `:466-474`).
- Claims are per run ID, not scope (`internal/coordination/durable_store.go:260-287`).
- App disables all rate-parking admission inside a resume context (`internal/app/app.go:3316-3320`), so one resumed request does not observe another live same-scope lease.
- The named concurrency test parks one run and checks only that new admissions are blocked (`internal/coordination/rate_parking_test.go:667-708`); it never parks two different run IDs for the same scope and fires both timers.

**Impact:** Two requests already in flight can both observe a 429 and persist distinct runs for the same provider scope. At reset, both timers/coordinators can claim their distinct run IDs and resume provider traffic concurrently, recreating the burst that caused the limit and amplifying duplicate-write/stale-CAS behavior.

**Root cause:** Admission is scope-aware, but durable uniqueness, scheduling, and leases are run-aware. There is no scope-level queue leader or serialization point.

**Exact change plan:** Model a durable per-scope queue/leader and acquire a scope claim atomically. Only one run for a scope may resume at a time; followers remain ordered/parked. A resume context may bypass admission only for its exact claimed scope leader and must still respect another live leader. Preserve concurrency across different scopes.

**Required tests:**

- **Happy:** two different scopes resume concurrently.
- **Bad:** two run IDs with the same scope/reset across two coordinators produce maximum provider concurrency one and deterministic follower handling.
- **Edge:** leader cancellation, claim expiry/takeover, restart with multiple same-scope records, different reset times, and a follower rearming after another 429.

### ORCH-W07 — Route selection hides actionable transport preflight failures

**Severity:** WARNING
**Evidence:**

- `shouldRunTransport` converts issue-label and declarative/managed-target `Registry.Preflight` errors to `false` at `internal/app/transport_dispatch.go:40-69`; it also converts connection, stream, contract, and action-selection errors to `false` at `:78-91`.
- The caller sees only the boolean at `internal/app/etl_mode_dispatch.go:51-73`.
- Contract modes then become the generic “no matching closed transport” error (`internal/app/etl_mode_dispatch.go:88-92`); legacy-compatible modes proceed into catalog and legacy destination I/O (`:94-106`).
- The discarded preflight errors would otherwise name missing source/destination declaration, executor registration, source binding, action, or conformance precisely (`internal/synctransport/registry.go:131-209`).

**Impact:** A declared route with a missing foundation or mapping can be reported as a generic unsupported mode, obscuring the corrective action. For a legacy-compatible mode, a malformed declared special route can fall into an unrelated legacy path and perform source/catalog work instead of failing closed before I/O.

**Root cause:** Route selection combines “this connection did not declare this special route” with “the route was declared but failed safety preflight” in a boolean API.

**Exact change plan:** Return a typed decision `(selected, reason, error)` from route selection. Only an explicit absence of a transport declaration may choose legacy behavior. Once both endpoints/connection select a special route, propagate every preflight, connection-contract, stream, action, and mapping error unchanged before any catalog/source/provider call.

**Required tests:**

- **Happy:** a truly one-sided legacy connection still takes the established legacy path; a valid special route executes transport.
- **Bad:** missing executor, source mapping, destination action, and conformance evidence return their precise typed errors with zero source/catalog/destination I/O.
- **Edge:** invalid issue-label connection metadata, a declarative source paired to a non-managed destination, a contract-mode route, and a legacy-compatible declared route cannot silently fall through.

### ORCH-W08 — Generated website data is deterministic but semantically out of sync with the live CLI

**Severity:** WARNING
**Evidence:**

- The website command map omits the live `pm etl transport postgres-managed-target` leaf at `website/content/docs/cli-reference.mdx:83-93`; the CLI implements and documents the leaf (`internal/cli/etl_transport.go:29-30`, `:208-209`).
- The ETL exit-code table omits policy exit code 7 at `website/content/docs/etl.mdx:376-388`, while runtime maps policy to 7 (`internal/cli/errors.go:164-183`) and the general CLI reference includes it (`website/content/docs/cli-reference.mdx:451-462`).
- ETL connection syntax incorrectly labels endpoints as `<credential>:<credential-name>` at `website/content/docs/etl.mdx:45-58`; the working CLI example uses `<connector>:<credential-name>` at `website/content/docs/cli-reference.mdx:193-204`.
- The generated artifact faithfully embeds the stale ETL and CLI-reference text at `website/lib/docs.generated.ts:47-52`.
- Generation serializes canonical MDX deterministically (`website/scripts/gen-docs-data.mjs:63-86`), but its only test is byte equality against that same input (`website/scripts/gen-docs-data.test.mjs:13-22`), so semantic CLI drift passes.

**Impact:** Operators and agents cannot discover one live approval-bound transport from the primary command map, may treat a policy refusal as an internal bug, and receive an invalid endpoint placeholder. The “docs validate is a release gate” claim at `website/content/docs/cli-reference.mdx:445-449` is therefore false for these semantic mismatches.

**Root cause:** Generation guarantees determinism, not parity with structured runtime command/exit-code/argument metadata. Manually maintained duplicate tables can drift independently.

**Exact change plan:** Correct the two MDX sources and regenerate `docs.generated.ts`. Export or generate the command map, exit-code table, and endpoint grammar from one structured CLI source. Add a parity test that enumerates runtime command leaves and categories and compares them with website data, rather than only regenerating from the same prose.

**Required tests:**

- **Happy:** regenerated content is byte-stable and includes every public runtime leaf, exit code, and correct placeholder.
- **Bad:** removing `postgres-managed-target`, code 7, or changing `<connector>` makes the parity test fail.
- **Edge:** aliases/hidden commands are explicitly classified, every URL is unique, section separators in `meta.json` are ignored deterministically, and generated content remains unchanged across two runs.

### ORCH-W09 — Evidence and phase status disagree and overstate output coverage

**Severity:** WARNING
**Evidence:**

- The evidence manifest records focused App and full App success, including “complete provider-result retention,” at `data/cli-current-foundations-main-integration-r1/evidence-manifest.json:24-43`.
- The TDD ledger simultaneously leaves source reachability pending, calls the full App package “in progress,” and leaves cross-bound rejection pending at `.planning/phases/cli-current-foundations-main-integration-r1/TDD-LEDGER.md:5-12`.
- The plan still marks focused combined checks and the independent evidence/report as incomplete at `.planning/phases/cli-current-foundations-main-integration-r1/PLAN.md:55-61`; verification still marks CLI parity and evidence hygiene unchecked at `.planning/phases/cli-current-foundations-main-integration-r1/VERIFICATION.md:23-33`.
- The #4303 run-state claims `complete_result_output: passed` at `.planning/phases/synctransport-4303-destination-r1/RUN-STATE.json:7-14`, but the scoped preservation tests exercise output sanitizers (`internal/connectors/write_result_output_test.go:12-74`) and the normal per-page transport path; the full-overwrite and both Arrow paths demonstrably omit acknowledgement output (ORCH-B03).
- The evidence manifest identifies `808896a28873c5f0479fa10e2f798da56f885b5e` as `composite_sha` at `data/cli-current-foundations-main-integration-r1/evidence-manifest.json:1-6`, not the reviewed evidence snapshot `9e5329f...`; it has no separate reviewed-evidence SHA field.

**Impact:** A downstream verifier can treat pending gates as complete and can infer output completeness across all orchestration modes from tests that do not cover those modes. The mismatch also makes it unclear whether claims describe the code composite or the later evidence-bearing review snapshot.

**Root cause:** Phase checklists, TDD state, run-state, and evidence manifest are maintained independently; evidence assertions are prose without machine-checkable links to the exact path/mode they cover.

**Exact change plan:** Generate phase gate status from one command ledger and update PLAN/TDD/VERIFICATION/evidence atomically. Give the manifest separate `code_sha` and `reviewed_sha` fields. Scope each assertion to named tests and modes. Remove the broad complete-output claim until ordinary, full-overwrite, Arrow serial, Arrow pipeline, success, error, and readback/checkpoint-failure cases exist.

**Required tests:**

- **Happy:** a schema/lint check confirms all “passed/green/complete” claims have a named command/test at the recorded SHA and all related checkboxes agree.
- **Bad:** a passed manifest claim paired with a pending TDD/verification row, or a `reviewed_sha` mismatch, fails validation.
- **Edge:** cached test results, deferred gates, provisional local component heads, and mode-limited assertions remain explicitly distinguishable without being promoted to global completion.

### ORCH-W10 — Rate-parking store methods can create invalid or unloadable state

**Severity:** WARNING
**Evidence:**

- Memory `Create` writes a run without calling `validateParkedRateLimitRun` (`internal/coordination/rate_parking.go:110-126`), while the file implementation does validate (`internal/coordination/durable_store.go:194-215`).
- Memory `Claim`/`RenewClaim` accept blank owner and zero/non-forward deadlines and write them at `internal/coordination/rate_parking.go:168-202`.
- File `Claim`/`RenewClaim` do the same at `internal/coordination/durable_store.go:260-305`.
- Persisted-state validation requires claim owner and deadline to be either both present or both absent (`internal/coordination/durable_store.go:369-387`). Thus `Claim(runID, "", now, nonzeroUntil)` can save a file that the next load rejects as “claim is incomplete.”
- The generic JSON update validates the loaded value before the callback through callers but saves the callback result without a post-update domain validation (`internal/state/store.go:116-139`).

**Impact:** A bad internal caller can corrupt the durable rate-parking file through an API call that returns success; all subsequent parking operations then fail to load it. Memory and file stores also disagree on whether an invalid parked record is accepted, undermining test fidelity.

**Root cause:** Store mutators rely on coordinator construction invariants instead of enforcing their own persisted-state contract, and validation is applied before rather than after mutation.

**Exact change plan:** Validate run ID, nonblank claim owner, nonzero timestamps, and `until.After(now)` at every store boundary. Make memory `Create` validate the run. Validate the complete next `rateParkingFileState` before returning it to `JSONStore.Update`, so no successful method can persist a value its own `load` rejects.

**Required tests:**

- **Happy:** identical valid operations behave the same for memory and file stores, and the file reopens successfully after every mutation.
- **Bad:** blank owner, blank run ID, zero deadline, non-forward deadline, and invalid run fields return errors and leave state byte-for-byte unchanged.
- **Edge:** failed validation after an existing valid claim, expired claims, duplicate creates, and reopen after each rejected call do not corrupt or partially mutate state.

## Explicit PASS Evidence

- **Immutable intake/provenance:** pre-review `HEAD` exactly matched `9e5329f34e015e39160bb8e951452bbd071a698a` and `git status --short` was empty. Every component SHA named in `input-manifest.json` was verified as an ancestor, and each listed preserving merge has that exact component head as its second parent. The two provisional local heads remain explicitly labeled as such (`data/cli-current-foundations-main-integration-r1/input-manifest.json:68-90`).
- **Atomic definition registration:** declarations and factories are validated before executor construction/registry mutation (`internal/synctransport/definition_composition.go:153-240`); tests cover multiple independent pairs and refusal before any registration (`internal/synctransport/definition_composition_test.go:12-100`). Missing foundations are named rather than silently skipped in this layer.
- **Plan/approval/apply order:** registry preflight occurs before planning, planning before source read, and each bounded mutation is authorized after page validation but before destination I/O (`internal/synctransport/orchestrator.go:54-100`, `:142-238`). No caller-controlled raw approval token is serialized in the transport request (`internal/synctransport/types.go:509-513`).
- **Durable acknowledgement guard:** acknowledgement construction has a private durability marker and commit validates it, candidate state, and acknowledgement time (`internal/synccontract/commit.go:20-29`, `:55-114`). The ordinary page path independently reopens warehouse work and reads provider state before invoking commit (`internal/synctransport/orchestrator.go:169-285`). ORCH-B01 concerns global writer ownership, not this local ordering.
- **Cancellation and bounded binary pipeline:** focused transport tests cover cancellation before apply, committing an already acknowledged page before returning cancellation, Arrow pipeline drain-before-abort, missing segment-store refusal before extractor I/O, and byte-credit release (`internal/synctransport/family_conformance_test.go:211-232`; `internal/synctransport/transport_test.go:772-870`, `:1316-1365`; `internal/synctransport/byte_credits_test.go:57-96`).
- **Ordinary provider output and secret semantics:** the normal apply path defensively preserves acknowledgement/error output (`internal/synctransport/orchestrator.go:237-248`). Scoped tests verify provider fields and credential-equal values remain exact while synthetic error text is redacted (`internal/connectors/write_result_output_test.go:12-74`). ORCH-B03 is restricted to the missing session-oriented paths.
- **Generated-doc determinism:** `generateDocsData` is a deterministic meta-order projection and writes one canonical artifact (`website/scripts/gen-docs-data.mjs:63-86`); the byte-equivalence test passed (`website/scripts/gen-docs-data.test.mjs:13-22`). The current `meta.json` resolves to 12 unique document URLs with no duplicates. ORCH-W08 concerns semantic parity, not reproducibility.
- **Focused verification commands passed at the reviewed snapshot:**
  - `go test -timeout 20m ./internal/connectors ./internal/coordination ./internal/synctransport`
  - `go test -timeout 20m ./internal/connectors/commandrunner`
  - `go test -timeout 20m ./internal/app -run '^(TestFoundationRollupPreservesMultiActionReverseETLComposition|TestRunETLTransportStaleWriterFinalizesLosingRun|TestRunETLTransportFullOverwriteStaleWriterAfterReceiptReadBackFinalizesLosingRun)$' -count=1`
  - `go test -timeout 20m ./cmd/connectorgen -run '^(TestSourceImport.*|TestValidate_CLISurfaceReverseETLRequiresRiskAndApproval|TestValidate_CLISurfaceDirectWriteStructuredRESTBodyRequiresClosedBoundedDeclaration|TestCheckCLISurfaceOperationHeaderMappingsRequiresExactDeclaredHeader|TestSyncBundleDirectWriteDerivesOperationContract)$' -count=1`
  - `node --test website/scripts/gen-docs-data.test.mjs`
  - `git diff --check`

Passing tests establish the listed positive contracts; they do not negate the adversarial interleavings and uncovered session paths described above.

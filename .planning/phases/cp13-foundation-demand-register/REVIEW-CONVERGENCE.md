# CP13 canonical complete carried finding ledger139

Frozen original reviewed source2e364954a6253442f50d86e0eb5e755b30b1978b/tree24d78d74914da61cea7e70ff3b2d93dcb2a8b137. Full138A72884bytes/SHA9407ccdb0aaeb017a0e523e30981811b3c48013e6c974497fdb5f01650981255; full138B118993bytes/SHA52b2c878e1d1513d3637853e92cd6f9d4d69b66845610ca86b7f172df853f458. Complete original reports/journals/probes/native finals remain immutable in canonical home. Three distinct invariants: CR-12 medium blocker; WR-138A-01/WR-138B-01 low warnings. All three are required corrections under139. CR-02 is an umbrella linked toCR-12, not a duplicate. CR-11/136 is independently closed; CP11/CP12 checkpoint acceptance remains carried.

The complete final finding narratives below are retained verbatim from the sealed reports. Their requested fixes are now authorized by139; their original candidate verdicts are not overwritten. No repair has begun at this ledger commit.

## Narrative Findings (AI reviewer)

### WR-138A-01 — Publish readiness JSON completely before exposing its pathname

**Classification: WARNING. Direct impact: low; test reliability. Final finding disposition: confirmed, should fix.**

**File:** `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_lock_contention_test.go:217-226`; producer at line 180, public caller at line 117.

**Reachable path and violated contract:** `TestConnectorgenMainSignalsOnlyAfterExactLockContention` starts the real-main child; after real `LOCK_NB` returns contention for the held directory inode, the child publishes its device/inode acknowledgment with `os.WriteFile`. The parent waits for pathname existence, then reads and unmarshals once. The F08R contract requires a bounded complete acknowledgment of that exact lock before signaling. `os.WriteFile` creates/truncates its destination before writing the JSON. If the child is descheduled between those operations, the parent reads an empty file and fails with `unexpected end of JSON input`, before its sound exact-inode check. A correct child can therefore spuriously fail the mandatory signal proof.

**Sibling invariant:** complete publication of interprocess readiness data. The same schedule exists in `cmd/connectorgen/vnext_publication_observation_test.go:101` → `:128-136` and `cmd/connectorgen/vnext_publication_witness_observation_test.go:151` → `:174-183`. Both parse once on successful ReadFile. The result read after child Wait does not share this readiness race. These siblings are one finding, not additional counts.

**Causality:** original CP11 miss, fix-created in the earlier F08R repair `69246943bdcb5c3cdc39c08a7cf1664f4af811aa`, as shown by blame on the actual producer/reader. It predates 094 and is unchanged by CP12/137. The underlying contention condition, real signal assertion and production lock behavior remain valid; this does not reopen their original production bugs.

**Independently expected/observed evidence:** one private overlay kept the real acknowledgment reader and split the real create/write sequence with a channel barrier. The created-but-unwritten schedule fails on the expected EOF; an actual complete `os.WriteFile` positive passes. The writer is released and joined by defer even on Fatal. This demonstrates a possible schedule, not its frequency or an assertion that any accepted historical receipt encountered it.

```text
go test -overlay /Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp11-joint-138a-2e364954-20260907T080056Z/probe-overlay.json -json -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCP11Review138AAcknowledgmentPublication$'
```

Terminal exit 1; Go package elapsed 0.998s; events 2026-09-07T08:17:13.849204Z–08:17:14.847512Z. Raw evidence `probe-contention-output.jsonl`: 3,423 bytes, SHA256 `c4c1f1520e68b15011235704efc9c949893dd1a83e24294a604aa2608063176c`. Overlay: 270 bytes, SHA256 `86d4b325bc1870247e0b5b5bf0343e6ee3f6283ac80b2e1928ecab94c282feb2`. Overlaid test: 10,264 bytes, SHA256 `cf93189c18b9ac6f1001e88d959e38472b1e677e94d852cc381960f9a4d0e3e3`. All are under the exclusive run directory. This probe is not a whole-package gate.

**Concrete fix and regression:** write each acknowledgment to a completed sibling temporary file and rename it to the ready pathname, or send a complete record over a pipe. Preserve the exact device/inode comparison, real LOCK_NB cut, bounded lifecycle and signal assertions. A deterministic regression should pause between temporary creation and complete publication and prove the parent waits and then accepts the correct complete payload. Apply the same protocol to the three siblings. No source fix was made during this frozen review.


## Narrative Findings (AI reviewer)

No structural pre-pass was supplied. Findings below follow direct source review, retained-source counterexamples and independent evidence reconciliation.

### CR-12 — BLOCKER: malformed response keys create successful-response ETL authority

**Impact:** medium. **Acceptance blocking:** yes. **Attribution:** initial_snapshot_miss in273644db, preserved through137; not caused by the local-shape helper. Related original umbrella: CR-02. This is one finding with four witnessed malformed spellings.

**File:** [source_lane_rules.go:401](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_rules.go:401), within401–421; classifier144–174. Affected sibling scope checks: [source_lane_collection.go:129](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_collection.go:129), [source_lane_bindings.go:269](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_bindings.go:269) and2082.

**Issue:** sourceResponseShape treats every response member beginning with `2` as successful. A retained read whose sole response key is `2invalid`, `20`, `2000` or `2ab`, with an otherwise ordinary object-array schema, reaches the actual inventory/facts/classifier/builder and advertises `applicable/mapped_unproven/source_record_collection`. Those keys cannot establish a successful response. The existing collection-scope and binding-coverage consumers use a different three-character test, which still admits `2ab`. Thus classification and scope accounting also disagree on malformed lengths.

| Sole response member | Independently expected ETL | Actual current builder |
| --- | --- | --- |
| `200` | applicable; positive control | applicable |
| `2XX` | applicable; supported range control | applicable |
| `2invalid` | undetermined | applicable |
| `20` | undetermined | applicable |
| `2000` | undetermined | applicable |
| `2ab` | undetermined | applicable |

Every case retained both literal keys `(fixture,primary,source.a)` and `(fixture,primary,source.b)`, observed normalized facts,14 cells in the literal seven-lane order, and applicable direct-read siblings. No accepted target, behavioral proof or implemented cell was created. This is a false authoring claim, not evidence of a runtime corruption incident.

**Evidence:** [private test](/Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp12-joint-138b-2e364954-20260907T080056Z/probes/status138b_test.go), [complete native output](/Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp12-joint-138b-2e364954-20260907T080056Z/probes/status138b-raw.jsonl), [command and custody receipt](/Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp12-joint-138b-2e364954-20260907T080056Z/probes/status138b-receipt.json). The overlay adds one nonexistent virtual test and replaces no production source. Actual Go exit1:7 run events,2 pass,5 fail including the parent. All four negative children fail the intended ETL assertion after the real builder frontier. Go package elapsed1.021s; the yielding transport did not separately measure total command wall time, so none is invented. All199 source hashes, HEAD/tree and tracked state match before/after.

**Fix:** use one closed successful-response-key predicate across shape, collection scopes, required coverage and no-body checks. Preserve exact200–299 and the already supported2XX range. Malformed/unsupported keys must not establish positive or exclusion authority; preserve anchored rows and visible uncertainty. Add permanent actual-builder/validator controls for malformed lengths/nondigits, valid200/299/2XX, non-success responses, binary/scalar consumers and malformed siblings beside valid success. Trace binding/proof scope consumers rather than repairing only the first prefix check.

**Current corpus limit:** the demonstrated malformed keys are private retained fixtures. No affected real provider row or runtime incident is claimed. The current structural-census results remain valid for their actual retained inputs.

### WR-138B-01 — WARNING: optional GraphQL null selector disagrees with its annotation schema

**Impact:** low. **Acceptance blocking:** not independently blocking; Firstmate owns disposition. **Attribution:** fix_created:481795060ab5fa619d87f03d540ee241ee146002, which introduced the annotation-envelope schema. The explicit null-positive reader test predates it at61d674d4. This is separate from137 and CR-11.

**File:** [source-lane-manifest.schema.json:1167](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/docs/connector-canon/source-lane-manifest.schema.json:1167), referring to the object-only definition at1030; [source_lane_rules_test.go:434](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_lane_rules_test.go:434).

**Issue:** the documented annotation envelope rejects `graphql:null`, although the optional Go pointer and permanent TestSourceLaneGraphQLCitationShapes deliberately accept that representation as carrying no facts. The same actual annotation passes the independent Draft202012 schema with the member absent or `{}`, but fails with null at `annotations/0/graphql`: `None is not of type 'object'`. The current normal receipt includes the passing permanent reader-shape test. No narrower canonical-input restriction is documented at SOURCE-LANE-MANIFEST.md:199–217.

**Fix:** allow null alongside the existing closed graphql_refs object; retain rejection of malformed citations and unknown fields. Add schema/reader parity checks for absent/null/empty/valid/malformed forms.

**Limit:** all six tracked annotations omit this selector; the complete current manifest and annotation envelope validate. This is authoring schema interoperability, not provider behavior, proof promotion or runtime data loss.


## Complete review148 frozen correction ledger — Firstmate150

Reviewed code b41c6eaaeceba77ac57b19c014083eaf76cc0ec5, tree48f6051c200a67d2e43c46564a9ec6fb6e66764f. Complete independent report104621bytes/SHA2569d6660ce319add752b55f64af4558e953ed60ac028ec6e038bcec871520381c7, fresh Astra/xhigh native01a07ca7-85d7-7d23-800c-d66003e1af5d, run cp11-cp12-cp13-final-148-b41c6eaa-20260907T160957Z. Owner read every byte and verified22 sealed artifacts plus original terminal event. Firstmate150 (18277bytes/SHA256c3f15cae614146ab70dfd7b15134ea31a3b7929a008dbc71fe01cfa1e25b908d) authorizes this entire coordinated correction wave. Original private reviewer artifacts remain immutable.

Firstmate separately accepted CP11 and CP12 at b41c6eaa; accepted original contracts and affected consumers remain protected. CP13 remains unaccepted. This artifact-only checkpoint precedes production corrections and permanent regression edits. All five findings are confirmed and dispositioned for correction, with their original severity and scope preserved; no extra generic audit or per-finding review. The following original finding narratives are retained verbatim from the complete report.

## Critical issues

### CR-148-01: Valid proof order can cause an empty-source declaration parse

Kind: authoring correctness defect, independently reproduced. Direct impact is failed generation, not runtime execution or data loss.

The selected 140/141 contract requires valid exact reviewed proof controls and deterministic admission independent of record ordering. The closed proofs schema does not assign authority to array order.

Entry-to-effect: `readSourceFoundationProofsObserved` iterates records; `sourceFoundationExecutions.validate` caches every input; a later record's `sourceFoundationPinnedDeclaration` calls `cache.get` for its owner. `source_proof_files.go:get` returns nil bytes on a hit. If the file was an earlier input but not an earlier parsed owner/test, the declaration map has no entry and `parser.ParseFile` receives an empty reader.

The two actual records overlap on `internal/synccontract/mode.go`: it is an engine proof input and the Mode proof's owner. Reversing the existing valid records should preserve both current observations, yet the code path appears to reject the Mode declaration. The default order happens to parse Mode first.

Affected siblings: any later foundation owner/test file previously consumed as another record's dependency; shared hashing callers do not require cached raw bytes and are not claimed affected.

Proposed correction: make parsed declaration acquisition independent of cache traversal order, retaining bounded original bytes or explicitly providing a charged, identity-checked reread; test both actual record orders and overlapping helper-to-owner/test transitions. Do not replace the empty reader with implicit filename reads.

Exact location: `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/source_foundation_atlas.go:141-151`, with `source_proof_files.go:127-131` and `source_foundation_execution.go:75-84` in the reached chain.

Terminal probe: `go test -overlay /Users/karthiksivadas/pm-cli-agent-workspace/data/review-runs/cp11-cp12-cp13-final-148-b41c6eaa-20260907T160957Z/probes/order-overlay.json -json -count=1 -timeout 20m ./cmd/connectorgen -run '^TestReview148ProofRecordOrder$'`. Exit 1; 3 run / 1 pass / 2 fail events including parent; package elapsed 1.755s. The original-valid-order child passed with both observations current. The reversed-same-valid-records child failed at the actual declaration consumer: `foundation declaration parse: internal/synccontract/mode.go:1:1: expected 'package', found 'EOF'`. The overlay adds one virtual test file and replaces no source file. Existing real-reader fixture copies the actual two reviewed records and original captures/inputs into `t.TempDir`; only array order differs.

Original tool-returned JSONL saved as [order-output.jsonl](probes/order-output.jsonl), 3,289 bytes, SHA256 `fea08a29d9dc9d1d13c57dc85942feb69a62e1d2a6cb36182b6821e7a7cff932`. Probe source 943 bytes/SHA256 `ab4d8503b315fe4e8f25669592a0029d6f3161f545d5bde6e98f08da1fac5a5a`; overlay 246 bytes/SHA256 `7f43bc06bfe7d901a93de66dde1654e301e3aa4d7c564ec281b25a3193fa45ba`. Whole-command walltime was not measured; package elapsed is not substituted for it.

Causality: newly introduced CP13 declaration consumer/cache interaction. Existing CP12 hash/result-only cache semantics are not thereby disproved. No source repair. Post-probe all 75 candidate path byte/hash pins still match and tracked Git status remains clean.

Disposition history: 2026-09-07T16:21:23Z static candidate; 2026-09-07T16:23:02Z confirmed with the original valid control and same-record permutation counterexample.

### CR-148-02: Missing/stale optional proof aborts independent demand reconciliation

**Classification:** BLOCKER; medium authoring correctness impact; CP13 acceptance blocking. No runtime lane promotion or production data loss is alleged.

**Locations/call path:** `cmd/connectorgen/source_foundation_register.go:79-81` propagates any proof observation error before constructing the register. `source_foundation_proof.go:118-120` treats a missing document as fatal; `source_foundation_execution.go:77-84` treats missing/stale inputs as fatal; `source_foundation_requirements.go` rejects missing referenced records before resolving an explicitly unresolved requirement. The actual command additionally rejects missing proof in `source_foundation_cli.go:282` and its input-pin loop before reaching the register.

**Existing contract:** Selected140 report lines123 and130-131 explicitly requires visible unavailable/stale proof while retaining source cell, known demand and next owner. Missing optional document or selected record must produce proof_unavailable. Only a purported resolved reuse is invalid. CP13-09 requires independent source/configuration work to continue. This is not a request to accept malformed documents, skip custody checks or weaken resolved proof.

**Independent expected behavior:** An otherwise valid two-source/fourteen-cell register with one explicitly unresolved demand remains available when the optional evidence document, a selected record or a dependency is absent/stale. Its proof cannot be current and its source/owner/membership must remain intact.

**Observed:** Private virtual Go tests use the actual source/Atlas/register fixture and first pass the valid complete register control in each case. Four single faults then return a zero register: missing document → "foundation proof document: missing"; missing selected record → "foundation requirement proof unknown: transport.sync-contract.v1.mode-vocabulary"; missing/stale go.sum → "foundation current input unavailable or changed". A fifth actual command case first passes source-demands, removes only optional proofs.json, then returns code1, zero stdout and "foundation command proof input invalid".

**Reproducer/custody:** `probes/unavailable148_test.go` and `probes/unavailable-overlay.json` add only a nonexistent virtual test file, using private temporary fixtures; no production overlay or repository mutation. Command: `go test -overlay <run>/probes/unavailable-overlay.json -json -count=1 -timeout 20m ./cmd/connectorgen -run '^TestReview148UnavailableProof'`. Actual terminal session35872, chunk1d8a0a, exit1; six run/six fail events including parent, zero pass events (each positive control completes before its negative within the same test). Package elapsed4.396s. Original full raw tool output remains in this review conversation; this journal quotes failure excerpts and does not label them a separately captured full receipt.

**Sibling scope and cause:** Initial CP13 consumer defect. Covers standalone proof observations, requirement construction, aggregate register and command preload, including present catalog with missing selected record and independent dependency changes. Existing CP12 optional lane-proof behavior and source-lane reducer are separate; neither was changed by this finding.

**Fix:** Represent missing/stale optional evidence as non-authorizing typed observations, preserve unresolved demands and full U/A/K/complement, and reject only resolved reuse that relies on unavailable evidence. Thread those states through command input custody and register schema/check. Keep cancellation, malformed/unsafe inputs and in-flight substitution refusals atomic. Add real command and register controls for each unavailable sibling.

### CR-148-03: Declaration file membership accepts an unrelated exact-contract proof

**Classification:** BLOCKER; medium authoring correctness impact. CP13-02/03/04 source-fit/adopter claims are affected; runtime lanes remain unpromoted.

**Location:** `cmd/connectorgen/source_foundation_requirements.go:179-202`, especially183-188; `source_foundation_register.go:171-190` derives adopter relations from these admitted tuples. A nonempty selector description plus a definition-file match never establishes that the selected source operation uses the exact proven mechanism/contract.

**Existing contract:** Firstmate148 explicitly forbids representing file membership as complete selector semantics. Selected140 requires a real source/config selector joined to the relevant existing owner/proof. The proof's own limitation says no direct operation behavior is proven. Its exact assertion is about a bundle declaring **check** status204.

**Actual counterexample:** The permanent positive fixture `sourceBindingREST118A` produces a real canonical `operation:widgets.get` GET /widgets operation with response status204. Its separate HTTP.Check is GET /check with **no** success_statuses. `engine.Check` reads HTTP.Check and `requesterWithCheckSuccessStatuses` returns the default requester when those statuses are absent (`internal/connectors/engine/read.go:2736-2760,2794`). Thus the selected /widgets declaration neither selects Check nor declares the Check-status contract. Nevertheless copying the Check proof's exact assertion into the requirement yields existing_shared_capability; removing operations.json yields connector_local_configuration. Both are accepted solely because operations.json belongs to the broad Atlas entry's file list.

**Independent probe:** `probes/selector148_test.go`, virtual test only. Real Atlas/two-current-proof controls and canonical source/binding controls complete first; shared and local cases both reach the production requirement builder and falsely resolve. Actual terminal session80727/chunkdbbe2f exit1, three run/three fail events including parent; package elapsed2.770s. Full output: `probes/selector-current-output.jsonl`. Command: `go test -overlay <run>/probes/selector-overlay.json -json -count=1 -timeout 20m ./cmd/connectorgen -run '^TestReview148ContractSelectorFit$'`.

**Disclosed fixture correction:** The first private probe expected no HTTP.Check; the shared minimal fixture actually declares a separate default /check. That run failed during setup, before the tested consumer, and is retained as `probes/selector-fixture-setup-output.jsonl` (session97603, exit1, elapsed1.930s). The corrected probe asserts the actual separate path and default status configuration. Neither run is owner pre-edit RED.

**Sibling scope/causality:** The same gate serves existing_shared_capability and connector_local_configuration, and every proof×binding adopter relation. The later144 file guard closes the original streams.json versus sync_transport.json mismatch but leaves this same-file/different-contract family. Permanent ExactBindingFit and SharedAndLocalAspects positives use this unrelated Check proof, so their passing receipts do not establish the promised relationship.

**Fix:** Verify an exact contract-specific selector/configuration relationship to the source requirement before resolving reuse or deriving an adopter. A broad Atlas entry/file match and quoted assertion are navigation only. Keep unrelated proof visible without resolved fit; add valid controls whose actual selected declaration executes the proven mechanism and counterexamples in the same allowed file using another mechanism/configuration.

## Warnings

### WR-148-01: Required combined auth/body source-fit control remains unproved

**Classification:** WARNING, acceptance-blocking missing mandatory evidence. This is a test-reliability/acceptance gap, not a claim of production criticality.

**Locations:** `cmd/connectorgen/source_foundation_aspects_test.go:5-78`; selected matrix `.planning/phases/cp13-foundation-demand-register/PLAN.md:82`; `source_foundation_fits_test.go:10-106`.

**Required and actual scope:** DemandMultipleAspects expressly requires a source body/auth requirement using an existing shared owner plus a local declaration deficit under the same source/lane. The later147 test selects one GET204 operation and its missing CLI command. Both requirements quote the Check-status assertion and cite responses. Neither requirement has request-body/auth facts, neither calls the facet-fit consumer, and no body/auth proof is selected. The independent auth tests exercise isolated scheme/placement observation; the body test exercises one separate canonical body match. Those controls do not establish the combined requirement promised by this matrix.

**Evidence:** Original normal1/1 and race1/1 receipts `requirement-shared-local-{normal,race}-147-01` are genuine later execution; their raw pins and source snapshots are verified in `probes/receipt-audit.json`. They are not initial RED and their names/counts are insufficient to close the missing auth/body behavior. CR-148-03 separately identifies the false Check-proof fit in their current fixture; this warning records the remaining required coverage even after that defect is corrected.

**Expected/fix:** Add a real retained-source/canonical-declaration fixture containing the requested body/auth requirements and a distinct local configuration deficit. Trace source facts, actual shared selector/mechanism proof and exact missing artifact to two independent results in one cell, with controls that fail if either requirement is dropped or the unrelated proof is borrowed. Capture actual terminal normal/race evidence without relabeling the existing later147 or prior setup failures.

### WR-148-02: Pending receiver condition disagrees between the schema and Go reader

**Classification:** WARNING; low authoring wire-consistency impact, not independently acceptance-blocking. No receiver approval or runtime authority is gained.

**Locations:** `cmd/connectorgen/source_foundation_assessment.go:92-100` decodes required decision.condition into a string; `source_foundation_requirements.go:132-138` only checks its resulting zero value. `docs/connector-canon/foundations/assessments.schema.json` decision first alternative requires condition equal to the explicit empty string.

**Observed/expected:** The published schema accepts the complete pending receiver decision and rejects omitted or null condition. The real assessment reader and requirement builder accept both omitted and null condition, normalizing either to empty string in output. Unlike the optional GraphQL contract repaired139, this new closed required member has no nullable/omission contract.

**Reproducer:** `probes/decision-parity148_test.go` and overlay add only a virtual test. In each case, a real retained-source/Atlas/assessment control with explicit empty condition reaches the builder successfully; the only mutation deletes or nulls condition. Both desired-refusal cases fail because the builder accepts them. Terminal session52218/chunk1b5532, exit1, three run/three fail events including parent; terminal Go package elapsed2.841s. Full raw output `probes/decision-parity-output.jsonl`. Separate installed Draft202012Validator applied to the actual published decision schema returns complete=true, absent=false, null=false. No network/schema retrieval or source write.

**Sibling scope/cause:** New CP13 wire decoder mismatch. Receiver ID/state already require nonempty exact values; exposure condition is nonempty and rejects both zero-value forms. The required decision_refs array has its own139/144 admission tests; those do not cover the nested empty-valued condition.

**Fix:** Preserve required/non-null member presence when decoding a decision, or explicitly reconcile the published grammar if omission is intentionally equivalent. Keep unknown-field rejection and pending/conditional meaning. Add the same complete/absent/null cases to real schema/reader parity checks.


### Owner timing clarification150

The review table labels772.983/622.740 as Go elapsed. They are original capture wall durations (772.9833497500222/622.7402766250016 seconds). Original raw receipts remain authoritative; no evidence rerun, original report rewrite or semantic disposition change.

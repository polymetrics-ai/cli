# CP15 frozen complete review170 and correction171 ledger

Reviewed candidate8d25bf0663da19b0ec828c64703a5ee9a707cbbe, tree dbd6598d45296c9ad94973f408be54ca279b7d46, original base29b790d741255ed6642a0bd5071d7b6e6cfc3bb8. Complete independent Astra/xhigh review170 found seven required corrections, one medium and six low impact. CR-164-01 resolved for its representation-shape contract. All seven original obligations, D1–D14, ten lenses and five safeguards remain accountable.

REPORT.md SHA256e1fa5198d2edea72f4f14d4e08a308c2d7bfd95b5e18ff807cf6043297372ce9; FINAL-SEAL.json SHA256e804688df75bf18e8fd18113e2d6151a355da4e0bf7e8c8c60a3b29438e40f58. Original files are immutable under private review-runs/cp15-final-170-8d25bf06-20260908T074209Z. Owner return cp15-review-170-return.md/.json preserves42 sealed artifacts and terminal native/probes. No repair is represented by this artifact-only checkpoint.

Firstmate171 adopts every finding for one coherent correction wave now. GroupA CR-170-01; GroupB CR-170-02/03; GroupC CR-170-04..07. Conditional CP16 scheduling is superseded by this seven-finding result. All remain OPEN until actual owner proof and independent corrected review. No approval policy relaxation, validation predicate change, runtime authoring reader, generic route, new dependency or test skip.

## Original final findings (verbatim body and metadata)

---
schema: "firstmate.review-finding.v1"
run_id: "cp15-final-170-8d25bf06-20260908T074209Z"
finding_id: "CR-170-01"
status: "confirmed"
severity: "medium"
acceptance_blocker: true
created_at: "2026-09-08T08:00:43Z"
updated_at: "2026-09-08T08:24:28Z"
source_sha: "8d25bf0663da19b0ec828c64703a5ee9a707cbbe"
parent_finding_id: "BD-CP15-168-01"
---

# CR-170-01 — Transport plan confirmation disagrees with its current policy resolver

Classification: **BLOCKER**. Kind: production correctness defect; medium demonstrated impact. Baseline age does not waive correction under Firstmate169/170 and captain156.

`internal/app/declarative_typed_destination_approval.go:87` seals every declarative typed destination transport plan with destructive confirmation, and the prepared approval target and physical-action set use the same contract. At `:451`, preview and authorization require both this stored policy and `App.confirmationPolicyForPlan` to be destructive. The latter (`internal/app/app.go:2808–2825`) only resolves ordinary command/action metadata and has no transport-mode case. An otherwise valid ordinary non-destructive apply action therefore creates a plan that immediately fails its own preview.

The private candidate probe [baseline-current.receipt.json](../baseline-current.receipt.json) reproduces the failure through real `PlanDeclarativeTypedDestinationTransport` → `PreviewDeclarativeTypedDestinationTransport` → `validateDeclarativeTypedDestinationPlan`, before the intended provider write. All five named baseline top-level tests fail. The multi-action pair has secondary count/subtest failures after its legitimate other-connector positive is rejected; those are siblings of this cause, not separate write/data-loss findings. Post-success receipt recovery, repeated full-append workset, and tombstone/read-back tests stop at preview and do not prove their later frontiers.

The source is unchanged from integrated R1; original failed current-App and derived34-file overlays remain evidence with their stated current-dependency/non-Go limitation. The independent probe uses exact current tracked source in an actual private cwd. No approval bypass or provider request is inferred from these failures.

Correction: make current confirmation policy resolve the authenticated transport plan's definition-owned transport contract consistently with creation, preview, sealing, authorization, and resume. Preserve destructive confirmation for this transport even when the underlying standalone action needs none; do not relax the validator or rewrite legitimate fixtures to conceal the mismatch. Re-run all five tests and add both typed-destination executor variants and tampered-plan negative controls.

Disposition: required correction; canonical shared CP16 owner under170/156 if the final transition rule is met. Final disposition is preserved in REPORT.md; the completed review has seven distinct required corrections.

**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/app/declarative_typed_destination_approval.go:451

Authority: Firstmate170 complete-review and baseline-disposition contract; adopted166/167 diagnostics where applicable; project mandatory meaningful regression coverage. This finding is a required correction, with demonstrated impact separate from acceptance blocking.

Final review disposition at 2026-09-08T08:24:28Z: confirmed, acceptance_blocker=true. No production repair was performed. Evidence and siblings above remain the basis of this judgment.


---
schema: "firstmate.review-finding.v1"
run_id: "cp15-final-170-8d25bf06-20260908T074209Z"
finding_id: "CR-170-02"
status: "confirmed"
severity: "low"
acceptance_blocker: true
created_at: "2026-09-08T08:04:12Z"
updated_at: "2026-09-08T08:24:28Z"
source_sha: "8d25bf0663da19b0ec828c64703a5ee9a707cbbe"
parent_finding_id: null
---

# CR-170-02 — Multipart allowlist rejection points to an absent media_policy field

Classification: **BLOCKER** for incorrect required diagnostic behavior; demonstrated impact **low**, not a security vulnerability. Kind: production diagnostic defect.

At `internal/connectors/engine/bundle.go:2590`, the `part.Type != "file"` branch for an authored `allowed_media_types` list returns `diagnosticAt("/media_policy", ...)`. That field is unrelated and may be absent. `Load` consequently directs the operator to `/actions/0/multipart/parts/0/media_policy` for a field part that contains only `allowed_media_types:["text/plain"]`. The original retained cause correctly names allowed_media_types; only the public coordinate is wrong.

Accepted contract:166/167 requires exact useful safe producer-owned property coordinates, retained by selected App/CLI projection; CP15-01/02. A present allowlist on a field is legitimately rejected, so changing acceptance would be the wrong correction.

The private [diagnostic-edges-v3 receipt](../diagnostic-edges-v3.receipt.json) runs the real loader against a healthy field-part control and the same declaration plus this one invalid allowlist. The control passes; the malformed case reaches the named semantic rejection and fails the independently expected `.../allowed_media_types` location. Output records exact file, field, reason/code and cause. Earlier multipart-location and multipart-location-v2 are retained setup failures (missing required kind, then record_schema); neither is production RED.

Affected sibling: the same helper is called from `validateOperationMultipartSemantics`, so operation-owned multipart field parts use the same wrong suffix. The provider_unrestricted `media_policy` branch is separate and should retain its correct media_policy coordinate.

Correction: locate the rejected allowlist at `/allowed_media_types` and use a corresponding useful fixed reason/code. Add both writes.json and operations.json selected-loader cases, preserving the healthy field and genuine media_policy cases; assert public JSON/text location and original cause. No source repair was made in this review.

Disposition: required correction; no predecessor or duplicate. Counts separately from missing semantic metadata in CR-170-03 because this producer supplies the wrong coordinate rather than no metadata.

**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/connectors/engine/bundle.go:2590

Authority: Firstmate170 complete-review and baseline-disposition contract; adopted166/167 diagnostics where applicable; project mandatory meaningful regression coverage. This finding is a required correction, with demonstrated impact separate from acceptance blocking.

Final review disposition at 2026-09-08T08:24:28Z: confirmed, acceptance_blocker=true. The semantic failure was repeated with the exact private test bytes immutably pinned before launch in [diagnostic-edges-sealed.receipt.json](../diagnostic-edges-sealed.receipt.json); both healthy controls pass. Original setup failures and first reached semantic receipt remain retained.


---
schema: "firstmate.review-finding.v1"
run_id: "cp15-final-170-8d25bf06-20260908T074209Z"
finding_id: "CR-170-03"
status: "confirmed"
severity: "low"
acceptance_blocker: true
created_at: "2026-09-08T08:04:12Z"
updated_at: "2026-09-08T08:24:28Z"
source_sha: "8d25bf0663da19b0ec828c64703a5ee9a707cbbe"
parent_finding_id: null
---

# CR-170-03 — Reachable change_apply rejection loses all useful semantic metadata

Classification: **BLOCKER** for incorrect required diagnostic behavior; demonstrated impact **low**, not a security vulnerability. Kind: production diagnostic defect.

`internal/connectors/sync_transport.go:680` still returns a plain error when a non-change-capture destination mode selects `change_apply`. `DestinationTransportDescriptor.Validate` has already accepted the canonical mode and known strategy, so this is a reachable semantic rejection, not a dead defensive branch. `declarationDiagnosticWithin` cannot add a path/reason to the plain cause, and `engine.loadBundle` falls back to `file=sync_transport.json, field=/, reason_code=bundle_invalid, reason=invalid bundle declaration`.

Accepted contract166/167 reserves that generic fallback for genuinely unclassified errors. Known declaration rules must report a fixed safe reason and exact useful field; the original cause alone is insufficient.

The private [diagnostic-edges-v3 receipt](../diagnostic-edges-v3.receipt.json) loads a valid full_append/append destination, then changes only its strategy to change_apply. The valid control passes. The invalid case reaches the exact expected internal message but the public tuple is generic/root. The source pointer is `/destination_transport/apply_strategies/0/strategy`; no authored value needs to be printed. Existing TestBundlePublicTransportBasics168 covers unknown strategy, missing mode and several siblings, but omits this known strategy/mode mismatch.

Correction: wrap this rejection at `strategyPath+"/strategy"` with a fixed code and useful reason, retaining the original error. Check the adjacent opposite mismatch for reachability before counting or changing it: destination change_capture is rejected earlier by `validateDestinationTransportModes`. Add current loader and public consumer coverage for the reachable mismatch without changing accepted strategy semantics.

Disposition: required correction. It is independent of CR-170-02's supplied-but-wrong multipart location and CR-170-01's approval execution defect.

**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/connectors/sync_transport.go:680

Authority: Firstmate170 complete-review and baseline-disposition contract; adopted166/167 diagnostics where applicable; project mandatory meaningful regression coverage. This finding is a required correction, with demonstrated impact separate from acceptance blocking.

Final review disposition at 2026-09-08T08:24:28Z: confirmed, acceptance_blocker=true. The semantic failure was repeated with the exact private test bytes immutably pinned before launch in [diagnostic-edges-sealed.receipt.json](../diagnostic-edges-sealed.receipt.json); both healthy controls pass. Original setup failures and first reached semantic receipt remain retained.


---
schema: "firstmate.review-finding.v1"
run_id: "cp15-final-170-8d25bf06-20260908T074209Z"
finding_id: "CR-170-04"
status: "confirmed"
severity: "low"
acceptance_blocker: true
created_at: "2026-09-08T08:13:53Z"
updated_at: "2026-09-08T08:24:28Z"
source_sha: "8d25bf0663da19b0ec828c64703a5ee9a707cbbe"
parent_finding_id: "BD-CP15-169-01"
---

# CR-170-04 — Polling help regression test rejects the current runtime-binding contract

Classification: **WARNING**. Impact: low; test reliability. Required correction.

`internal/cli/changefeed_cli_test.go:203-209` still requires the literal phrases "declaration alone" and "constructs an implemented declaration". The actual contextual help at `internal/cli/docs.go:493-500` describes a static binding and an implemented binding constructed for the selected catalog object, and explicitly requires runtime preflight of the selected source/apply executors. That preserves the accepted distinction; it does not deny dynamic eligibility. The test fails before checking its second positive condition.

The exact candidate was independently run by baseline-current, reproducing TestPollingHelpDistinguishesStaticDeclarationsFromDynamicRuntimeEligibility. The original full CLI command remains failed. This is an obsolete assertion contract, not demonstrated polling execution failure.

Fix the two public-help assertions to recognize the current binding vocabulary and test the substantive negative/positive preflight distinction, with whitespace-tolerant comparisons. Keep a falsifier that removing the static-only qualifier or dynamic eligibility makes the test fail; do not delete the test or weaken it to checking exit 0. Source age does not remove this required warning. It is distinct from the source-origin fixtures and ETL mode help below.


**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/cli/changefeed_cli_test.go:203

Authority: Firstmate170 complete-review and baseline-disposition contract; adopted166/167 diagnostics where applicable; project mandatory meaningful regression coverage. This finding is a required correction, with demonstrated impact separate from acceptance blocking.

Final review disposition at 2026-09-08T08:24:28Z: confirmed, acceptance_blocker=true. No production repair was performed. Evidence and siblings above remain the basis of this judgment.


---
schema: "firstmate.review-finding.v1"
run_id: "cp15-final-170-8d25bf06-20260908T074209Z"
finding_id: "CR-170-05"
status: "confirmed"
severity: "low"
acceptance_blocker: true
created_at: "2026-09-08T08:13:53Z"
updated_at: "2026-09-08T08:24:28Z"
source_sha: "8d25bf0663da19b0ec828c64703a5ee9a707cbbe"
parent_finding_id: "BD-CP15-169-02"
---

# CR-170-05 — Source-origin fixtures still demand a retired preflight contract

Classification: **WARNING**. Impact: low; test reliability and misleading boundary coverage. Required correction.

`internal/cli/cli_test.go:89-110` and `:191-229` both require the removed source-bound provider-operation refusal for the Asana custom_fields command. The current command is declared stream_etl with operation `get_custom_fields_for_workspace`, target `stream:custom_fields`; its execution bundle uses the ordinary configured base URL. `internal/cli/cli.go:1010-1035` invokes the current commandrunner PreflightRequest, then ordinary project/credential handling. Current production routing contains no PreflightSourceBound provider-truth checker; authoring/proof facts are not runtime inputs under the adopted architecture. Existing closed route and alternate-action origin checks are different, declared contracts (for example engine/operation_route.go:232-256).

The independent baseline-current command reproduces both failures: the first reaches missing-project handling; the persisted-config fixture reaches the deliberately deleted encrypted credential file. Neither evidence establishes a provider request or secret exposure. Treating these fixtures as a valid safety oracle for the current declarations gives two permanent red tests while failing to exercise a declared refusal.

Fix by replacing/rebinding the fixtures to an actual execution-declared origin restriction and its real preflight boundary, retaining a healthy positive and a physical request counter. If the purpose is the current Asana configurable stream, assert its accepted project/credential ordering instead. Do not restore an authoring-source reader or invent a blanket fixed-origin rule solely to satisfy the obsolete message. Preserve both original failed cases and observation linkage; they share one obsolete contract/root cause.


**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/cli/cli_test.go:102

Authority: Firstmate170 complete-review and baseline-disposition contract; adopted166/167 diagnostics where applicable; project mandatory meaningful regression coverage. This finding is a required correction, with demonstrated impact separate from acceptance blocking.

Final review disposition at 2026-09-08T08:24:28Z: confirmed, acceptance_blocker=true. No production repair was performed. Evidence and siblings above remain the basis of this judgment.


---
schema: "firstmate.review-finding.v1"
run_id: "cp15-final-170-8d25bf06-20260908T074209Z"
finding_id: "CR-170-06"
status: "confirmed"
severity: "low"
acceptance_blocker: true
created_at: "2026-09-08T08:13:53Z"
updated_at: "2026-09-08T08:24:28Z"
source_sha: "8d25bf0663da19b0ec828c64703a5ee9a707cbbe"
parent_finding_id: "BD-CP15-169-03"
---

# CR-170-06 — ETL help regression test is tied to obsolete wording

Classification: **WARNING**. Impact: low; test reliability. Required correction.

`internal/cli/cli_test.go:874-889` requires literal "Compatibility name for typed ... admission" wording and a lower-case sentence beginning "retains". Current help at `internal/cli/docs.go:996-1018` names every tested mode, explains the typed-only aliases through their behavior and refusal before source I/O, and starts the history sentence with "Retains". The current parser still maps the two compatibility names to full_overwrite and incremental_dedupe in `internal/synccontract/public_modes.go:34-57`; ParseSyncMode carries those contract modes while preserving the public spelling. This failed wording assertion alone does not demonstrate incorrect execution or lost mode admission.

baseline-current independently reproduces TestETLHelpListsAllSyncModes failing at the first old phrase; the additional obsolete literals remain reachable siblings in the same test.

Fix this help test to validate the current user-facing mode descriptions and typed-executor refusal semantics, using whitespace/case normalization only where appropriate. Retain explicit tests of compatibility-to-contract mapping at the parser boundary and a help falsifier for removing a mode or claiming unconditional execution. If clearer compatibility wording is desired in help, update runtime/docs/website together; that wording choice does not justify leaving a permanently failing test. Count this mode-help contract once, not once per missing substring.


**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/cli/cli_test.go:878

Authority: Firstmate170 complete-review and baseline-disposition contract; adopted166/167 diagnostics where applicable; project mandatory meaningful regression coverage. This finding is a required correction, with demonstrated impact separate from acceptance blocking.

Final review disposition at 2026-09-08T08:24:28Z: confirmed, acceptance_blocker=true. No production repair was performed. Evidence and siblings above remain the basis of this judgment.


---
schema: "firstmate.review-finding.v1"
run_id: "cp15-final-170-8d25bf06-20260908T074209Z"
finding_id: "CR-170-07"
status: "confirmed"
severity: "low"
acceptance_blocker: true
created_at: "2026-09-08T08:13:53Z"
updated_at: "2026-09-08T08:24:28Z"
source_sha: "8d25bf0663da19b0ec828c64703a5ee9a707cbbe"
parent_finding_id: null
---

# CR-170-07 — Diagnostic store tests can hang when their loader oracle fails

Classification: **WARNING**. Impact: low; test reliability. Required correction.

The new loader closures in `internal/connectors/manifeststore/bundle_diagnostics_165_test.go:43-45` and `:79-81` call t.Fatalf if the actual engine no longer returns the typed diagnostic. BundleStore invokes these closures from its separate `go s.load` goroutine (`bundle_store.go:216`). Fatal calls Goexit on that goroutine, so load never reaches its completion publication/close of pending.done. The test goroutine is blocked in Acquire with t.Context(), which is only canceled after the test ends. The very regression these two tests should diagnose therefore stalls until the global 20-minute timeout.

This is a concrete faulty failure path; ordinary passing runs cannot exercise it. Static control flow proves the lifecycle dependency. It is not a production deadlock claim about a normal returning loader.

Fix by having the loader return the actual error/result without calling Fatal and assert diagnostic identity from the main test goroutine after Acquire returns; use a captured result or buffered channel if original identity must be examined. Keep unexpected non-diagnostic failure and successful/no-error loader cases as bounded test controls. Both closures share this one test-harness root cause.


**File:** /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/connectors/manifeststore/bundle_diagnostics_165_test.go:44

Authority: Firstmate170 complete-review and baseline-disposition contract; adopted166/167 diagnostics where applicable; project mandatory meaningful regression coverage. This finding is a required correction, with demonstrated impact separate from acceptance blocking.

Final review disposition at 2026-09-08T08:24:28Z: confirmed, acceptance_blocker=true. No production repair was performed. Evidence and siblings above remain the basis of this judgment.


## Firstmate171 owner disposition checkpoint

The immutable seven originals above remain the finding authority. These are implemented dispositions pending broad validation and fresh exact-SHA review, not owner acceptance:

| Finding | Correction and actual focused evidence | Remaining gate |
|---|---|---|
| CR-170-01 / BD-CP15-168-01 | Shared current policy recognizes the closed typed-destination transport mode as destructive; standalone actions unchanged. Permanent two-executor App RED then unchanged-test GREEN; ten tamper controls and all five original recovery/result/tombstone tests pass. | Full App, selected race, independent review |
| CR-170-02 | Field allowlist rejection points to allowed_media_types with its fixed safe reason/code. Both action and operation real Load controls, selected App/CLI, cause preservation and isolated tuple falsifiers pass. Separate media_policy branch unchanged. | Affected full/race and independent review |
| CR-170-03 | Existing non-change-capture/change_apply refusal gains exact strategy coordinate and useful safe reason/code while preserving original cause. Valid append and selected App/CLI controls pass. Opposite mismatch remains rejected by earlier source-only mode validation. | Affected full/race and independent review |
| CR-170-04 / BD-CP15-169-01 | Test asserts normalized static binding versus dynamic binding plus actual preflight qualification. Removal/eligibility falsifiers pass. Production help unchanged. | Full CLI and independent review |
| CR-170-05 / BD-CP15-169-02 | Obsolete Asana origin refusal tests replaced with current configurable operation's actual project/vault ordering and healthy local returned-row/one-request control. Setup failures remain separately recorded. | Full CLI/race and independent review |
| CR-170-06 / BD-CP15-169-03 | ETL help asserts present modes, refusal/compatibility meaning, history fields and fixed alias mapping. Mode removal and misleading execution perturbations fail the oracle. | Full CLI and independent review |
| CR-170-07 | Async loader callbacks return actual errors without Fatal; main goroutine validates identity/cause after bounded completion. Plain non-diagnostic failure and success controls terminate. No production store lifecycle change. | Store race and independent review |

Original CR-164-01 remains protected by its existing permanent representation-shape regression and independently resolved170 disposition. Original seven CP15 obligations, D1–D14 and95+AM-169-01 assertion accounting remain required. None is replaced by a count of new tests.


Final171 local validation completes the remaining local gates in the checkpoint table: full current App/CLI normal, full affected packages normal/race, final App39 normal/race, CLI77 race, lint/vet/build/Atlas/generated/docs/smoke all pass with exact scope recorded in VERIFICATION.md. The one later test-fixture registration check and its final witness are explicit. Each CR-170-01..07 disposition is owner-implemented/locally-verified and awaits fresh171 independent judgment; no self-acceptance. All historical finding files and failed receipts remain unchanged.

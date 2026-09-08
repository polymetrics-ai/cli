---
phase: cp15-diagnostics-legacy
reviewed: 2026-09-08T10:13:00Z
depth: deep
files_reviewed: 51
files_reviewed_list:
  - cmd/connectorgen/source_visibility_shape_165_test.go
  - data/connector-canon/batch1-foundation-assessments.json
  - data/connector-canon/batch1-foundation-demand-register.json
  - data/connector-canon/batch1-source-lane-manifest.json
  - docs/cli/connectors.md
  - docs/connector-canon/foundations/catalog.json
  - internal/app/app.go
  - internal/app/bundle_public_diagnostics_167_test.go
  - internal/app/diagnostic_coordinates_171_test.go
  - internal/app/transport_confirmation_171_test.go
  - internal/cli/bundle_public_diagnostics_167_test.go
  - internal/cli/changefeed_cli_test.go
  - internal/cli/cli_test.go
  - internal/cli/diagnostic_coordinates_171_test.go
  - internal/cli/docs.go
  - internal/cli/errors.go
  - internal/cli/help_contracts_171_test.go
  - internal/cli/lazy_registry_test.go
  - internal/cli/source_visibility_shape_165_test.go
  - internal/cli/testdata/golden_transcripts.json
  - internal/cli/testdata/source-visibility-shape-165.json
  - internal/connectors/bundleregistry/bundle_diagnostics_168_test.go
  - internal/connectors/connectors.go
  - internal/connectors/connsdk/send_order_165_test.go
  - internal/connectors/database/definition.go
  - internal/connectors/engine/binary_read.go
  - internal/connectors/engine/bundle.go
  - internal/connectors/engine/bundle_diagnostics.go
  - internal/connectors/engine/bundle_diagnostics_165_test.go
  - internal/connectors/engine/bundle_execution_boundary_168_test.go
  - internal/connectors/engine/bundle_io_diagnostics_167_test.go
  - internal/connectors/engine/bundle_public_diagnostics_167_test.go
  - internal/connectors/engine/bundle_test.go
  - internal/connectors/engine/diagnostic_coordinates_171_test.go
  - internal/connectors/engine/direct_read.go
  - internal/connectors/engine/operation_headers.go
  - internal/connectors/engine/operation_multipart_test.go
  - internal/connectors/engine/operation_parameters.go
  - internal/connectors/engine/operation_route.go
  - internal/connectors/engine/polling_definition_test.go
  - internal/connectors/engine/rate_limits.go
  - internal/connectors/engine/schema.go
  - internal/connectors/engine/testdata/diagnostic_coordinates_171.json
  - internal/connectors/hooks/hookset/closed_inventory_165_test.go
  - internal/connectors/manifeststore/bundle_diagnostics_165_test.go
  - internal/connectors/manifeststore/bundle_store.go
  - internal/connectors/native/nativeset/closed_inventory_165_test.go
  - internal/connectors/polling_watermark.go
  - internal/connectors/source_visibility.go
  - internal/connectors/sync_transport.go
  - website/content/docs/cli-reference.mdx
findings:
  critical: 1
  warning: 1
  info: 0
  total: 2
status: issues_found
run_id: cp15-corrected-review-171-20260908T083959Z
source_sha: baedc5254266d0cbcc55fda3e86a4d8e64f5b57f
source_tree: 8a8f56541683620751c6ef521c642bc46931771c
diff_base: 29b790d741255ed6642a0bd5071d7b6e6cfc3bb8
reviewer_native: 01a08066-d46e-7710-b95d-251dc1a5ac1f
reviewer_model: gpt-6-astra
reviewer_effort: xhigh
---

# Corrected CP15 independent review171

## Narrative Findings (AI reviewer)

The complete review finds **two distinct required corrections**: one low-impact production diagnostic defect and one low-impact test-reliability warning. Six original170 findings are resolved. CR-170-06 has corrected its obsolete wording failure but remains incompletely repaired through CR-171-02. CR-164-01 remains resolved for its accepted representation-shape contract.

The critical frontmatter counter means one **BLOCKER-classified contract defect**; it does not describe a critical security vulnerability. No approval bypass, provider request leak, customer database operation, or data loss was demonstrated. The required warning counts alongside the production finding. These are independent causes, not multiple counts for examples or consumers.

Review scope is the entire original CP15 delta from `29b790d7`, including CR-164-01, through corrected `baedc525`, plus affected callers. All 67 changed paths are accounted for: the 51 source/test/docs/generated paths above and 16 phase evidence paths below. This is a fresh corrected-candidate examination with explicit reuse of unchanged contracts independently examined in170; it is not a claim to have replayed170's199 historical commands or freshly reread every unchanged large file.

## BLOCKER findings

### CR-171-01 — CLI target semantic rejection still loses exact diagnostic meaning

**Classification:** BLOCKER. **Impact:** low. **Kind:** production diagnostic defect. **Status:** confirmed, required correction. **Parent finding:** none; newly discovered reachable producer omission within original CP15 scope. This is distinct from repaired CR-170-03's transport-strategy producer.

**File:** `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/connectors/engine/bundle.go:3342`; sibling at3347. Relevant helper: `internal/connectors/engine/command_endpoint.go:18`.

`loadCLISurface` validates `foundation_gap.target` and `unsupported_disposition.target` after schema validation, but returns a plain wrapped `ValidateCommandEndpoint` error. The loader's new diagnostic boundary therefore emits `file=cli_surface.json`, `field=/`, `reason_code=bundle_invalid`, `reason=invalid bundle declaration`. A known, reachable canonical-path rule loses the producer-owned safe property and useful public explanation required by166/167.

The isolated `TestReviewerCLITargetDiagnostics171` loads healthy `/widgets` targets successfully for **both** owner variants. Changing only the target path to the synthetic `/../private-sentinel-171` reaches the real helper's `noncanonical path segment` rejection and preserves its original cause. Both malformed cases fail the independent exact-coordinate assertion: the actual field is `/`, rather than `/commands/0/foundation_gap/target/path` or `/commands/0/unsupported_disposition/target/path`. This is semantic RED with successful setup, not a fabricated schema-inaccessible branch. The current schema permits these strings; it does not reject dot segments first.

Reachable siblings belong to the same producer correction: invalid canonical path shape, whitespace, query/fragment/backslash, invalid percent encoding, decoded dot/separator segments, dangerous characters, and GRAPHQL operation-identity validation. Foundation target methods are only schema-constrained as nonempty strings, so its unsupported-method branch also needs a method coordinate. The unsupported-disposition method enum already catches unknown methods earlier; do not fabricate a second semantic failure there. Its8192 path bound is also schema-owned, while the foundation target's bound reaches this helper.

The full cause is already retained and public output is already safe. The defect is loss of useful public meaning and exact location, not changed acceptance or a proved secret leak. Store selection copies generation/digest freshly and preserves the producer tuple; App's SafeBundleError and CLI's bundle projection preserve that tuple rather than inventing missing metadata. Those consumer internals and the current permanent selected-consumer tests were examined. The new counterexample was executed at the real loader boundary; a second dedicated App/CLI execution of these exact new target fixtures was not launched. Public propagation is a direct source trace corroborated by current selected-consumer tests for the same diagnostic type, not claimed new dynamic consumer evidence.

**Fix:** attach producer-owned typed metadata for the actual failing target property. Prefer classifying method/path validation at its rule-owning helper, retaining its original error, then prefixing the exact command index and owner in the loader. Keep reasons/codes fixed and safe; never include arbitrary target values. Preserve all existing refusal predicates and healthy targets. Add both-owner loader tests and selected App/CLI tuple/cause/output controls, with a valid target and falsifiers for root/generic metadata. Do not weaken the schema or route validator to obtain green.

**Evidence:** [private command](probes/new-edge-controls/command.json), [terminal receipt](probes/new-edge-controls/receipt.json), [raw output](probes/new-edge-controls/output.jsonl); output SHA256 `05a6dbb98d6cd9a68f6aa807bd9e9a5c1a5dbd3fbac63f95cf771487404682a2`. Exact private test bytes are pinned before launch. Discovery does not itself provide owner repair GREEN.

## WARNING findings

### CR-171-02 — ETL help oracle accepts missing modes and false descriptions

**Classification:** WARNING. **Impact:** low. **Kind:** test reliability. **Status:** confirmed, required correction. **Parent finding:** CR-170-06; original observation BD-CP15-169-03. The weak helper was introduced in the correction wave; the retired-wording failure is fixed, but its substantive replacement contract is not complete.

**File:** `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/internal/cli/help_contracts_171_test.go:26`; falsifier selection at49; caller `internal/cli/cli_test.go` TestETLHelpListsAllSyncModes.

`etlHelpContract171` searches the entire normalized help string for mode-name substrings. Removing the complete `full_refresh_overwrite` heading and description still passes because `full_refresh_overwrite_deduped` contains the name. The same false positive exists for `incremental_append` versus its deduped name and `incremental_dedupe` versus the history name. Most descriptions have no independently asserted semantics at all.

The isolated `TestReviewerHelpOracleCompleteness171` obtains actual `cli.Run("help", "etl")` output and first proves the current oracle accepts it. Three bounded subtests remove each full mode section, including its description, and all demonstrate false acceptance. A fourth replaces the actual overwrite description with `Appends every record and keeps all existing rows on every run.`; the oracle still accepts it. Each mutation checks that its intended original section/text exists before replacement. This is a test-oracle counterexample, not a production RED or evidence that actual current help lies.

The current falsifiers remove only the history mode and generic refusal phrases. They cannot establish protection for every mode or distinguish overwrite from append semantics. The existing alias checks correctly examine five spellings through LookupPublicMode, and current App parser/mode tests preserve canonical mappings, typed executor refusal, cursor/key requirements and no source reads for incompatible modes. Those useful checks do not repair the missing help-description oracle. Count this once, not once per removed section or alias.

**Fix:** assert exact mode sections/headings with independently specified semantic expectations for each mode; prevent longer names from satisfying shorter ones. Retain whitespace/case tolerance where harmless. Falsify removal of each complete section and misleading descriptions for the protected behavior, including overwrite/append, cursor advancement, deduplication/history, and typed-executor refusal. Retain actual parser-boundary alias tests and healthy help output. A same-source expected string, exit0-only check, or one history-mode mutation is insufficient. Production help does not need to change to fix this warning.

**Evidence:** the same [private command](probes/new-edge-controls/command.json), [terminal receipt](probes/new-edge-controls/receipt.json), and [raw output](probes/new-edge-controls/output.jsonl). The help probe's pre-command SHA256 is `51123e742463ee020bb738d888c7837a7522cc8b33e39b421faa70546e02d84a`. The four demonstrated false positives have healthy actual-help controls.

## Identity, intake, and evidence authority

The frozen candidate is branch `fm/cli-top100-declaration-batch-r1-cp15`, HEAD `baedc5254266d0cbcc55fda3e86a4d8e64f5b57f`, tree `8a8f56541683620751c6ef521c642bc46931771c`. Base and actual merge base are `29b790d741255ed6642a0bd5071d7b6e6cfc3bb8`. HEAD/tree/branch, tracked cleanliness, original-plus-fix paths and the23 correction paths were independently checked against Git. The original candidate remains `8d25bf0663da19b0ec828c64703a5ee9a707cbbe`/`dbd6598d45296c9ad94973f408be54ca279b7d46`.

The fresh reviewer is actual native `01a08066-d46e-7710-b95d-251dc1a5ac1f`, parent `01a07bcc-454f-75a3-89f3-8bbb91572206`, role gsd-code-reviewer, gpt-6-astra/xhigh. Whitelisted native metadata was inspected; requested labels alone were not treated as proof. No reviewer child was launched. Implementer,164,166,170 and ancillary172 were not reused as this reviewer.

Canonical assignment:21984 bytes, SHA256 `dcd46de1aab3a91aefdd62d678393929f7ab32495a91a58f224acb4eb0b7df96`; candidate binding:22913 bytes, SHA256 `f1b0ae3da8e8f4ead96c0e3dd69fc34a3438c69d992a0f4ac3d50af1226ff2cb`. The complete unchanged assignment and full Firstmate171 implementation assignment were read. Native encrypted message storage prevents a separate plaintext byte-for-byte runtime-payload claim; the assignment file pin and supplied full dispatch were checked and this limit remains disclosed.

The complete170 report, seven final finding files, final seal, owner return, and completed Firstmate custody check were read. Report170 is50036 bytes/SHA256 `e1fa5198d2edea72f4f14d4e08a308c2d7bfd95b5e18ff807cf6043297372ce9`; its seal is15279 bytes/SHA256 `e804688df75bf18e8fd18113e2d6151a355da4e0bf7e8c8c60a3b29438e40f58`. The current handoff, source/receipt index, original seven obligations, PLAN/SUMMARY/VERIFICATION/TDD/convergence/migration/inventory records, captain156 and engineering/evidence/journal policies informed this review. All48 named binding pin occurrences checked at intake matched their supplied bytes/hashes.

Applicable project instructions and required skill routing were read. The exhaustive-review skill is pinned to SHA256 `93695083c2b48115089e8a37d613a6b4d2d60331ef2cecf49f162b516418330b`; Go how-to and applicable CLI/testing/error/security/safety/concurrency/context/documentation guidance were used. Source-lock vNext, Atlas, and CLI help/docs parity contracts were consulted. No `.codegraph` directory exists in this checkout. The project GSD code-review workflow was resolved/executed by the canonical owner through the recorded non-Pi adapter fallback; this fresh role performs the assigned independent review. Current Astra/captain instructions supersede obsolete Claude/additional-stage wording, without granting a new implementation or integration step.

## Each original170 finding

| Original ID and observation | Before witness | Current source and after witness | Independent disposition |
|---|---|---|---|
| CR-170-01; BD-CP15-168-01 | Two valid executor previews fail destructive-policy validation; all five original App cases stop before intended later assertions | `App.confirmationPolicyForPlan` now returns destructive for the exact typed-transport mode. Ordinary action resolution is unchanged. Creation, definition binding, plan seal, preview grant, confirmation, durable scope and resume were traced. Owner unchanged-test RED13/10→GREEN32/32; final App39 normal/race includes the original five and both executor variants. | **Resolved.** No validator relaxation or fixture-policy substitution. |
| CR-170-02 | Field multipart allowed_media_types points at media_policy for writes and operations | `bundle.go:2587` reports allowed_media_types with fixed useful code/reason; genuine media_policy restriction retains its own coordinate. Healthy write/operation controls and selected App/CLI cases pass; original cause graph remains inspectable. | **Resolved.** New CR-171-01 has another producer owner. |
| CR-170-03 | Full_append/change_apply reaches generic root diagnostic | `sync_transport.go:680` reports the exact strategy field and fixed safe rule; valid append control, original cause, field/code/reason falsifiers, selected App/CLI pass. Destination change_capture remains refused earlier, so its opposite semantic branch is not fabricated as a test. | **Resolved.** Acceptance/refusal semantics unchanged. |
| CR-170-04; BD-CP15-169-01 | Polling help test demands retired static/dynamic wording | `pollingHelpContract171` requires static binding alone insufficient, implemented per-object binding eligible only after the same runtime preflight, and rendered source/apply executors. Three falsifiers remove/change these distinctions and are rejected. | **Resolved.** Correct production needed an oracle correction, not product RED. |
| CR-170-05; BD-CP15-169-02 | Two Asana fixtures demand retired source-truth preflight | Current configurable stream contract: missing project before credentials; healthy local `/workspaces/workspace-171/custom_fields` returns exact field-171 after one request; deleting owned vault causes refusal without another request. Full CLI and selected race pass. | **Resolved.** No runtime authoring reader, fabricated fixed-origin policy, or production routing change. |
| CR-170-06; BD-CP15-169-03 | ETL help test demands retired phrases/capitalization | Current wording test passes and typed/parser controls remain useful; independent deletion/misdescription probes expose false-positive replacement helper. | **Partially repaired; required correction CR-171-02.** Original failure history and observation linkage retained. |
| CR-170-07 | Two asynchronous loader callbacks call Fatal and can strand Acquire | Both loaders return normally. Main test goroutine asserts after bounded Acquire; pending.done publication synchronizes captured error reads. Success and unexpected plain-error completion controls pass, including current package race. Production BundleStore lifecycle unchanged. | **Resolved.** Static invalid-callback trace plus bounded controls suffice; no intentional20-minute hang. |

The five original App cases retain separate downstream evidence rather than five findings:

| Original test | Reached current frontier |
|---|---|
| TestFoundationRollupPreservesMultiActionReverseETLComposition | Wrapper calls the actual persisted multi-action App test; its pass does not create another distinct criterion. |
| TestPersistedConnectionSelectsDeclarativeTypedDestinationAction | Correct selected action for two actions and another connector, one acknowledged write, ordered provider results and declared output semantics. |
| TestDeclarativeTypedDestinationPersistsProviderResultsAfterPostSuccessLocalFailure | Two actual successful POSTs followed by the deliberate missing-locator local failure; failed run preserves ordered provider results and reports loaded0. |
| TestDeclarativeTypedDestinationReopensRepeatedFullAppendWorksetAfterLocalReceiptFailure | Reopened App retains exact uncommitted stage receipt, distinct prior workset key, stable retried-workset idempotency key; actual3 calls/2 fixture mutations and committed receipt identity. |
| TestDeclarativeTypedDestinationTombstoneAppliesOnlyDeclaredDeleteAndReadsBackAbsence | Both physical delete actions visible in approval; two declared missing-ok deletes, zero creates, independent readback absence, unchanged2 output and one acknowledged checkpoint commit. |

Transport's new policy branch changes neither standalone action policy nor the current binding validator. Tampered missing seal, removed confirmation, mode, action and binding all still refuse. Plan creation grants no execution token; preview rederives definitions and verifies the authenticated seal before minting a grant. Authorization rederives source/destination/executor/action/mapping/digest and checks credential/configuration/scope identity. Existing authorization rejects token replay; per-unit callbacks recheck current scope and cancellation. Existing expiry/restart/recovery rules remain their prior bounded contracts, not an indefinite authorization guarantee.

## Seven original CP15 obligations

| Row | Verdict | Contract-to-evidence and applicability | Evidence limit |
|---|---|---|---|
| CP15-01 lazy selected diagnostics | **Required correction: CR-171-01** | Real metadata/lazy/store/Resolve/App/CLI paths retain healthy selection, exact selected identity, cancellation and unknown/wrong-case distinctions; D1/D2/D6–D8/D10/D11/D14 and current2303 plus selected consumers. Newly reachable malformed CLI target loses useful location. | Healthy inventory does not certify every selected bundle; generic tuple defect remains. |
| CP15-02 secret-free CLI/App parity | **Required correction: CR-171-01** | Exact malformed rate_limits syntax/semantic producer tests, selected generation/digest, full joined causes, CLI JSON/text/App and physical healthy controls remain applicable. Corrected multipart/strategy tuples pass current actual consumers. Public projection excludes arbitrary values and sibling text. | New target fixture executed at loader, consumer propagation source-traced; no new exact-target App/CLI dynamic claim. No unrelated provider-output redaction guarantee. |
| CP15-03 complete legacy inventory | **Reviewed** | Original390 structural rows and1039 lexical complement unchanged; all541 production path/function/callee identities match169 exactly as a multiset. Only resolveAuthAdmission line moves3639→3645. Actual registrations, wrappers, compound/follow-up/auth paths and explicit nonexecution dispositions retain170 examination; changed production adds no route. | Census is navigation plus source-backed disposition, not provider execution or proof from equal totals alone. |
| CP15-04 protected compatibility | **Reviewed** | Exact api_engine and seven protected factories plus49 generated hooks retained. Constructor/unknown-ID/wrong-identity controls, closed bundleregistry selection and170 source examination apply. No factory/extension owner changed. | No database startup/customer work, new removal authority, or certification of undeclared future rate behavior. |
| CP15-05 every physical send | **Reviewed** | Public/compound route→RequesterFor→requester/wrappers→retry/redirect→client.Do ownership unchanged. Independently expected caller→provider→route→server order and complete returned two-record bytes cover JSON/form/multipart/stream; refusal causes and zero server arrival preserved. GitHub follow-up/auth routes retain exact owners. | Accepted logical-request admission scope remains: safe-read net/http retransmission exception is not silently upgraded to per-retransmission re-admission. |
| CP15-06 retry/park/resume | **Reviewed** | Existing117 normal/209 race and170 physical-carry evidence retained for unchanged requester/rate/parking owners. Current repaired approval now reaches original persisted provider-result/restart/tombstone assertions. Strict writes, terminal RateLimitError, RateBudgetRefusalError, parked_rate_limit/resume and checkpoint/acknowledgement contracts remain unchanged. | No whole-machine crash/power-loss/shared-coordinator/provider-wide guarantee; full App normal/current selected race scope is exact below. |
| CP15-07 closed routing/truthful failure | **Reviewed, with shared diagnostic correction above** | Unknown/forbidden/malformed selection remains typed data error/internal_error, without selecting another action, generic transport, legacy fallback or source-mapping classification. CR164 source inspection remains separate from ordinary execution. No new registry or authoring reader. | Required inherited help warning CR-171-02 remains outside a claim of complete output/test acceptance; source visibility does not prove provider capability. |

## CR-164-01 continuity

**Resolved and preserved.** This finding concerns projected representation shape, not runtime revalidation of retained provider truth. The shared-pure-validator option was not required: current explicit alignment with the pure authoring validator is tested and remains unchanged by171.

The original canonical two-operation fixture has valid nonempty operation/command references and field mappings. The current `sourceArtifactProjectionShape` verifies reference/schema role placement, non-nil targets, legal empty schema root, nonempty config field coordinates, canonical parameter/pagination_parameter indices, bounded JSON citations, transport/schema separation, and actual duplicate/ancestor/root overlap by source document or target kind. Different owners and sibling coordinates remain valid. Both intended/admitted loops retain original nil target, missing role, transport mixture, duplicate mapping and malformed parameter cases. The CLI fixture is bound to actual canonical projection, not the current cohort's empty lists.

Current App/selected decoder/CLI malformed shape controls return SourceVisibilityDataError/source_visibility_invalid with zero resolver construction, App-open and approval-reader calls. Valid reference shapes pass decode; source_mapping_unproven remains the correct later outcome where mapping proof is absent. Current ordinary commands retain their separate execution path. No import cycle, new source-lock/proof reader, or second executor interpretation was introduced.

170's independently examined original owner RED/GREEN and tests retain their historical identity: original shape test SHA256 `187b05a0aaab79c3536abc9a91134a4d6d4c10f420a94c81d114bb87070677e7` across owner RED02/GREEN21, with later coordinate/citation/CLI controls separately recorded.171 changes no shape source, test, fixture or relevant parser contract; current CLI77 race and full CLI normal include its public path. Unchanged authoring-control evidence is reused rather than replayed to manufacture freshness. The withdrawn fabricated-implementation/provider-truth interpretation is not reopened.

## D1–D14 and assertion accounting

| Diagnostic row | Disposition and evidence |
|---|---|
| D1 lazy inventory/healthy unselected | Reviewed: TestCLILazyMalformedUnselected168, TestConstructionSelectedDiagnostic168 and App lazy controls; exact membership/zero unselected loads plus healthy selected construction. |
| D2 selected malformed/public boundary | Reviewed with CR-171-01 limit: selected engine/store generation, original syntax cause, safe App/CLI output and actual malformed-selection physical boundary/healthy returned bytes. |
| D3 duplicate rate rule | Reviewed: reachable duplicate policy index1, fixed reason/code and original semantic cause; unchanged predicate. |
| D4 closed rate state/source/scope | Reviewed: actual unknown-with-reason policy prohibition and healthy rules; real unknown/cancel ordering retained. |
| D5 policy/budget/cost/selector | Reviewed: exact indexed safe reasons and unchanged internal predicates; original95 migration obligations stay separate from new tests. |
| D6 file/read/parse/absence | Reviewed: selected required-file and metadata syntax/read distinctions; optional pure absence versus joined failure, original fs/parser causes. |
| D7 schema reference/compiler | Reviewed: failed reference names streams.json reference property, successfully read malformed schema names actual schema file; compile metadata and original errors retained. |
| D8 untrusted authored names | Reviewed: fixed trusted fields, sorted member ordinals, database known-schema/byte coordinates; synthetic raw keys do not enter public diagnostics. |
| D9 stream/write/multipart/operation | Reviewed with CR-171-01: corrected write/operation allowed_media_types metadata, genuine media_policy sibling and healthy controls. CLI target validator is another reachable producer missing useful metadata. |
| D10 selected identity | Reviewed: store makes a fresh diagnostic copy with selected connector/generation/digest and original full error as Cause; current selected tests reject mutation/reuse. |
| D11 wrapped/joined graphs | Reviewed: errors.Is/As at real consumers, complete sibling preservation privately, safe public projection, pure-versus-compound absence; incorrect field/reason/code falsifiers. |
| D12 App parity/order | Reviewed: connector/plan routes and credential-first endpoint retain ordering; new corrected tuples pass actual App tests. |
| D13 actual CLI output | Reviewed with required diagnostic/help corrections: JSON seven-field bundle, safe text, internal_error/exit1, already-reported suppression and no successful plain stdout; current corrected tuple consumer tests pass. |
| D14 LoadAll/source separation | Reviewed: exactly two healthy and two named malformed bundles, per-failure SyntaxError and safe metadata; malformed source-only data leaves healthy execution identity unchanged. |

`ASSERTION-MIGRATION-168.json` is byte-identical to170. All95 expressions remain accountable:82 retained cause assertions,7 restored public assertions and6 restored direct-helper expressions. The89 rows comprise82 cause and7 public rows; two restored groups account for the six direct expressions. The supplementary AM-169-01 remains separate, preserving selected errors.Is graph checks while replacing the obsolete public-sentinel expectation with safe malformed-JSON/file/exit assertions. This migration follows166 and remains justified. Current correction adds useful public coordinate controls; it neither removes original tests nor changes denominators. CR-171-01 shows why these tables are evidence maps rather than a universal proof of every known diagnostic producer.

## Whole-delta source and route impact review

| Source group and affected callers | Examination and reuse decision |
|---|---|
| engine bundle loader, schema compiler, rate rules, stream/write/operation validators, database/changefeed/polling/transport declarations | Original producer changes and170's complete traced classification reused where predicates and consumers are unchanged. Fresh examination covers current three production edits, the full diagnostic wrapper/selection graph, schema/optional-file ownership, multipart and transport siblings, and CLI endpoint omission. Plain helper errors were evaluated at real callers or earlier schema gates, not dismissed by grep count. |
| `bundle_diagnostics.go`, BundleStore, lazy registry/bundleregistry, App selected resolvers, CLI error envelope | Fresh copies preserve complete causes and selected identity. Store callback through reservation, load return, publication, done close and wait return inspected. New test closures return normally. Current consumer tuple tests and raw captures inspected. Original cancellation/unknown/healthy distinctions retain170 proof and current affected suites. |
| App transport creation/preview/authorization/persistence and standalone reverse action policy | Current resolver fix traced through `declarative_typed_destination_approval.go`, `issue_label_transport_approval.go`, `issue_label_warehouse_transport.go`, `transport_dispatch.go`, authenticated grant/seal consumers and original App tests. Exact mode branch cannot authorize a changed connection/action/binding because those are separately rederived and sealed. Current final normal/race exercises both executor variants and original later side effects. |
| Source visibility and connectorgen pure authoring projection | Original source/test changes, actual shape validator, nonempty reference tests and CLI consumers examined; historical withdrawn provider-truth claims are not inferred. No runtime authoring input. |
| Requester wrappers, native factories, generated hooks, closed routes and compound operations | Complete170 source-backed inventory and proof reused; all relevant production owners are unchanged. The three current production edits neither add an HTTP call nor change admission/retry/redirect code. Exact identity multiset, owner lists, physical-boundary proof and refusal positives independently reconciled. |
| CLI help/origin/mode tests, actual hand parser, docs/help/manual/site outputs | All correction tests and actual mode/source contract owners inspected. Polling and origin replacements are meaningful. ETL false-positive oracle is required warning. `newRootCmd`/hand-parser route ownership is unchanged; no Cobra/Viper substitution. Existing mode parser/refusal tests corroborate runtime meaning. |
| Atlas/catalog, assessments, source manifest/demand register | Catalog49→50 changes existing loader/reverse owner proof metadata and transport confirmation guarantee. No new shared foundation or runtime input. Independent recursive JSON comparison finds only2 source-manifest pin leaves and64 demand-register pin leaves, including source_manifest_content_sha256; every non-pin semantic value is identical. |

The16 phase paths excluded from the source counter are still reviewed evidence: ASSERTION-MIGRATION-168.json; CONTEXT.md; DISCUSSION-LOG.md; INVENTORY-RECONCILIATION-169.json and171.json; LEGACY-INVENTORY-SOURCE-PINS.json; LEGACY-ROUTE-INVENTORY.json; LEGACY-ROUTES.md; PLAN.md; RATE-BOUNDARIES-165.md; RATE-PROOF-RECEIPTS-165.json; REVIEW-CONVERGENCE.md; SUMMARY.md; TDD-LEDGER.md; VERIFICATION.md; cp15-diagnostics-legacy-UAT.md. Their historical pending/failed sections are superseded only by explicitly later scoped evidence, not silently rewritten as passes.

The complete inventory categories are49 hooks,8 factories,325 selected call occurrences,2 legacy commands,2 rate declarations and4 admission-vocabulary rows. Every original390 row and1039 lexical complement retains its source-backed disposition; no new N/A category closes an unexamined hole. The independent541 full production-call identity multiset extends beyond selected structural call rows. Protected factories are `native_database/{dynamodb,mysql,postgres}.v1` and `closed_typed/{bing-ads,faker,hubspot,tally-prime}.v1`; api_engine is the migrated shared execution owner. Hooks without current declared-rate files retain that explicit limit. GitHub's declared routes/auth/follow-ups select RequesterFor; postgres is declared not_applicable. Deferred/unsupported metadata remains nonexecuting, not a hidden executor.

Generated source accounting remains30401 cells:26 assessed and the exact30375 complement;4343 operations retain4341 primary and2 supplementary membership. Original programme denominators309/31 remain intact. Equal counts alone were not the comparison: complete JSON values and call identity multisets were compared. Atlas updates do not refresh old assertion evidence into current proof; existing stale/unavailable observations and source fit/adopter semantics are unchanged. No `implemented` label, classifier pass or visibility record is treated as provider/source-cell execution.

## Ten lenses

| Lens | Status | Result and limit |
|---|---|---|
| Architecture/data flow | Complete | Single execution-JSON reader, registry, hand parser and approved engines preserved; new producer gap reported. No connector-specific routing/fallback/authoring reader. |
| Happy/bad/edge behavior | Complete | Healthy setup controls, selected/unselected, absent/missing/null/unknown/duplicate and reachable semantic cases considered. Exact new canonical-path failure proved; schema-inaccessible opposite transport case excluded with source reason. |
| State/concurrency | Complete | Store copy/publication synchronization, bounded returning callbacks, plan/preview/grant scope and persisted recovery traced; no production callback lifecycle change. Reused overlapping claim/cancel evidence applies to unchanged owners. |
| Secret taint | Complete | Private raw parser/authored/joined context remains inspectable; public fixed reason/coordinate/identity projection omits arbitrary values and sibling causes. Synthetic sentinels only. No general provider-response redaction claim. |
| Retry/rate/resume/idempotency | Complete | Exact admission order and physical counts; strict mutation replay limits, parked/resumed/expiry/cancel/checkpoint consumers and repaired transport restart witnesses. Safe-read replay exception retained. |
| Output integrity | Complete | Safe JSON/text shape and error exits, already-reported behavior, no false plain success, provider-result/acknowledgement ownership; help oracle weakness reported. Existing error-output/short-write behavior is not promoted to a new transactional guarantee. |
| Declaration reachability/closed surface | Complete | Exact factories/hooks/legacy/auth/follow-up identities, reachable CLI target paths and mode intersections considered. No route authorized merely by metadata. |
| CLI/App parity | Complete | Current actual selected diagnostics, project/credential ordering, hand-parser/inspect/plan consumers, source inspection separation; new target consumer propagation is statically established with explicit dynamic limit. |
| Provider semantics | Complete within assigned scope | Retained provider facts/declaration-owned action/path/mapping and fixture results respected. Provider-live correctness, real customer DB and hosted/receiver behavior are evidence-backed N/A for this hermetic review. |
| Tests/evidence | Complete | Current34 raw captures, exact selected run/pass/terminal accounting, source applicability, RED/GREEN chronology, migration denominators and bounded oracle falsification. One required warning survives; green suites alone did not establish acceptance. |

## Five evidence safeguards and reached frontiers

1. **Contract-to-evidence table:** all seven rows, CR164, seven170 findings, D1–D14,95+AM169, complete inventory and affected consumer groups have explicit dispositions above. No finding/round/severity cap was used.
2. **Fallible helper internals:** loader Stat/Open/read/parse/schema/semantic stages through current-file attribution were traced; optional absence does not discard a joined failure. Store reservation/loader return/identity copy/publication/wait and main-thread test assertions were examined. Transport plan→sealed preview→grant→authorization→physical write→result/recovery/acknowledgement frontiers were traced. New CLI endpoint probes prove the semantic validator is reached after a valid control. Existing physical-send observations are at actual server/client boundaries, not outer counters.
3. **Compound identity and public output:** current real store/App/CLI tests check selected identity, complete errors.Is/As graphs and safe display separately. Plain unknown, cancellation, genuine malformed data, source mapping and optional absence are not conflated. Exact new target producer Cause is retained despite its missing public metadata.
4. **Independent expected members/bytes/state/history:** exact541 call multiset, factory/hook identities, full generated JSON semantic equality, known two-record wire bytes, provider-result identities, stable/distinct retry keys and durable receipt/checkpoint comparisons are used. Aggregate test/row totals never substitute for expected identity or source execution.
5. **Oracle falsification and source applicability:** current coordinate field/reason/code mutation controls pass; new exact mode-section/misdescription mutations demonstrate the surviving weak help oracle. Malformed target tests have healthy controls and real later-rule causes. Original RED/setup/zero-test distinctions remain explicit. Every new probe starts in the actual private copy with pre-command input pins and terminal raw output.

## Verification audit and historical limits

The current owner index contains233 captures:199 original plus34 current.170 independently audited the199 originals, including75 nonzero, and its unchanged input-specific conclusions are reused. This reviewer audited all34 current receipt/output captures, verified their recorded output hash/byte counts and completion, parsed selected JSON run/pass events, and compared exact `(Package, Test)` multisets. There are seven preserved current nonzero captures, so82 historical/current nonzero captures remain evidence; this is not82 defects.

| Current capture | Independently reconciled scope/result | Applicability |
|---|---|---|
| cp15-full-app-normal-171-01 |691 run/pass events, exit0;350.351s; raw `cc9b6857b079526419553d31ede69f8d6de544d037486ed5e035a016bf00a98c` | Full App before only final fixture Register error check. |
| cp15-full-cli-normal-171-01 |28806 run/pass events, exit0;632.091s; raw `dd3a22850ddbdc5735c5d9d5502e1ca4662164fefa2411e094b76cc3e9b39905` | Current CLI source/test bytes; later App test edit is outside this package. |
| cp15-engine-store-normal-171-01 / race-171-01 |Exact same2303 `(Package,Test)` run/pass events; exit0; raws `b586dfc5e9ded2f03d9ed197aa9cb2506f8b73ebe331120f23828121c3e2eb92` / `800752fe0ebbb5b64e6b3255c16e5eb47c1e4afdf4cdaa2c29d80fbb05c7637c` | Full engine, manifeststore, connectors, bundleregistry; final App test edit irrelevant. |
| cp15-final-app-focused-normal-171-01 / race-171-01 |Exact same39 events; all pass; raws `f768c0ef4f48d488a1933fdc38803bdac5fbd3e68c51c93da999e64f367316c1` / `5f92f513c1b43b06d841c6dbe948f2e82f5be36c685d92ab1a08c1ac571b8fd1` | Final test bytes, both executor variants, five original App cases and corrected diagnostic consumers. |
| cp15-final-cli-focused-race-171-01 |77 events all pass; raw `2683be45b0a37a9c5efbe569d7d6a645a668149424c2d297afa9093f922479d1` | Selected diagnostic/source-shape/origin/help cases; not full CLI race. |
| Final vet/lint/build |Affected vet plus final App vet; lint new-from-original-base affected packages exit0 after one fixture error-check repair; current binary builds | Not full-repository lint/CI or a changed dependency certificate. |
| Atlas/generated/docs/smoke/GSD |Atlas20 events pass; current source-lanes and source-demands deterministic checks exit0; docs validation and owned sample→warehouse→outbox smoke pass with three exact rows; final committed GSD evidence gate passes | Generator success not provider execution; source semantics independently compared; no full make verify/full CI claim. |

Independent command-input hash comparison against current files confirms that the four broad normal/race captures differ in exactly `internal/app/transport_confirmation_171_test.go`; production, module files and embedded execution inputs match. Final App normal/race and CLI race have no differing captured Go/module/internal inputs. The test-only diff checks the existing fixture Register error and does not alter its behavioral assertions. Broad evidence is therefore applicable for its named scope; the final39 explicitly cover the changed test. Captures bind dirty source input bytes where applicable, not merely their recorded HEAD.

Actual product RED precedes production corrections. Transport permanent test SHA256 `7ed8a2ceaad3189732f7db26944ec8c590fbe2dbf6000cb07675a446820f6e10` is unchanged between RED and first GREEN; both snapshots were checked and time order verified. Diagnostic test SHA256 `c80198a8695c26f994602efb5df95db57f5643007a5def84ed7e894d2b085611` is unchanged between corrected-fixture RED02 and first GREEN; later falsifiers are distinct additions.

The seven current nonzero results remain: transport behavioral RED; first diagnostic attempt with missing-risk setup failure alongside actual semantic failures; corrected diagnostic behavioral RED02; origin unused-import compilation failure; origin missing synthetic access_token setup; origin wrong workspace configuration setup; and lint's unchecked fixture Register result. None of the setup/compile failures is product RED. Polling/origin/help/store changes for correct production are explicitly test/oracle corrections. Passing current full App/CLI captures supersede the earlier failure for current scope without relabeling old full failures as green.

170's first two multipart probe versions were overwritten before immutable archival; those initial attempts remain setup failures with no repaired-source RED credit. Its later sealed semantic probes had real pre-command input custody.164's three terminal probes used the frozen-worktree cwd with overlays and therefore violated the later isolation requirement; that history remains disclosed and is not erased or reused as private-copy execution. No replay was performed merely to make that history look clean.

Unchanged native/database/connsdk/hook/physical/rate evidence retains170's independently examined contracts and exact limits. Earlier producer2416, engine2081, optional-absence35 and selected CLI63 evidence is not added to current2303/77 as though parent/subtest events were new criteria. Historical release/tidy/agent-contract/definitions/preflight/boundary checks retain actual unchanged source/module/configuration applicability; no full CI, provider-live, Linux/power-loss, remote runtime or customer DB certification is implied.

## Probe isolation, terminal custody and continuation

The only newly launched reviewer product command is `python3 <journal>/probe.py new-edge-controls go test -json -count=1 -timeout 20m ./internal/cli ./internal/connectors/engine -run '^TestReviewer.*171$'`. It is terminal exit1 after62.753 seconds, with the meaningful failures above. The raw output is17484 bytes/SHA256 `05a6dbb98d6cd9a68f6aa807bd9e9a5c1a5dbd3fbac63f95cf771487404682a2`. All selected tests and their parent accounting are present; this was not a compilation or zero-test failure.

Before launch, the committed candidate was archived into `<journal>/private-copy`. The baseline contains15528 pinned candidate files, matches the candidate source map by path/bytes/hash, and has no symlinks. Archive SHA256 is `7a9b74ead146d2e8920c5a8f2277b651c4facaa637e2c78f484a819126d89444`. The launcher was inspected and asserts its resolved cwd differs from `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli`; it records that condition, exact argv, all current private source/test inputs and relevant configuration before launch. GOCACHE, GOMODCACHE, GOTMPDIR and TMPDIR point under this journal; GOWORK is off. No overlay or frozen-worktree test/build/generator command was used. Initial archive setup encountered an unsupported Python tar filter argument before extraction/probes and was retried with explicit member validation; no product result was inferred from that setup attempt.

The reviewer did not modify production source or protected user/cache/session contents. No private native session bulk read or unrelated cache copy was performed. The role's authoring restriction was honored by having the owner mechanically maintain journals/private probe files from reviewer instructions; reviewer owns assertions and judgment and authors this REVIEW.md. No named Write tool is available in this runtime, so the available patch writer creates only this report. REPORT.md mirroring, stable finding extraction and FINAL-SEAL.json are mechanical owner completion steps, not another review or opportunity to narrow findings. Optional further exact-target App/CLI probes were considered but not launched; their dynamic limit is explicit above and they are not pending required evidence.

All launched reviewer Go commands are terminal and no reviewer child exists. Final independent custody checked all15528 tracked candidate files and the same15528 private-copy baseline files by path/bytes/SHA256 against the bound source pin map, with no mismatch. The probe receipt/output hash and isolated cwd assertions were rechecked; HEAD/tree remain exact, tracked status is clean, and git diff --check passes. Final report/probe/evidence bytes are to be sealed by the canonical owner when publishing this complete return. This report does not claim an owner seal already exists or certify whole-machine process isolation.

The final independent actionable set is **CR-171-01 and CR-171-02**, with CR-170-06 linked to the latter and all other170/164 dispositions as above. Correctness acceptance is not clean. Captain156 and the subsequently read Firstmate173 continuation authorize the complete two-correction result to proceed through normal R1 integration and CP16 with explicit owned carries, after terminal report/seal/carry/source binding. They do not require a separate current-CP repair/re-review or another SHA-only permission. Integration remains separate from correctness acceptance and must be bound by the canonical owner before CP16. This reviewer has performed no push/merge, main integration, repair, provider operation, new receiver or CP16 work.

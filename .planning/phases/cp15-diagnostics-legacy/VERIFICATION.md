# CP15 verification

Owner implementation and applicable local verification are complete for review; independent acceptance remains pending. PLAN.md is the sole seven-item plus CR-164-01 manifest. D1–D14 and all95 original assertion dispositions plus supplementary AM-169-01 are mapped below. Full App and CLI checks remain failed with explicit unresolved observations; no all-checks-green claim.

Current production source passes final engine2081 normal/race and physical publication230/canon checks. The sole later Go edit removes a redundant test type annotation: focused35 normal/race and final lint pass after it, with every assertion unchanged. CLI63 targeted race passes. Final source-demands check, regenerated manifest check, vet, binary build, smoke, help, docs, Atlas, definitions, runtime preflight, boundary, release, tidy and agent-contract results are individually mapped in the candidate receipt index, with exact source applicability below.

The following sections preserve earlier source-scoped chronology; their then-pending statements are superseded by the final169 reconciliation at the end. No provider-live, customer database, shared-service, final CI or acceptance claim.

## Producer verification after diagnostic checkpoint 781129d9

Database and polling-changefeed diagnostics now have actual malformed-load regressions with healthy controls, safe paths/reasons and complete causes. Database unknown-member coordinates use schema-owned dollar paths plus parser token-end byte offsets, without exposing member text. All current connectors/database/engine tests pass normal and race: 2416 test events each, including parent events. Original receipts are cp15-producers-normal-168-01 (raw647a0e6df37d074729837d930556586a18efb2d8ab3619b98545080426fe5d57) and cp15-producers-race-168-01 (raw4488218aa4fce09d684a96cedd959f9142310c39968ba07ea094a85489b17324); changed_inputs_after is empty. These bind post-checkpoint dirty source, not a new committed candidate. Selected manifeststore/bundleregistry normal81/81 also passes (cp15-selected-stores-normal-168-01, raw8d89462287220305f844558ec85cf33bfd678686f4d8d52b3fd9fb45b3e4f009).

Residual raw helper errors are classified by their actual loader call sites, not blanket grep exclusion:

| Producer | Loader disposition |
| --- | --- |
| declaredOperationRoutes blank/whitespace name | streams.schema.json base.routes.items.name pattern `^[a-z][a-z0-9_-]*$` refuses these before semantic loading; origin/version/duplicate branches retain typed semantic metadata |
| validateOperationParameterCLIName | validateOperationParameters wraps every failure with exact block/parameter index/cli_name and safe fixed reason |
| validateOperationRouteBase/Version | declaredOperationRoutes supplies fixed safe metadata and retains original helper errors |
| write multipart unsupported type; base64 source | writes.schema.json enum guards (`field/file`, `path/base64`) run before semantic validator |
| required_query empty any_of; sensitive input_mode/transform | operations.schema.json minItems1 and closed enums refuse before semantic validator |
| GraphQL default scalar parse helpers | loader caller attaches typed argument location and fixed safe reason; raw helper remains private cause |
| request-header value and parameter wire-value validators | Runtime request-value validation, distinct from bundle load; existing public request contracts are unchanged |
| requireOperationBinaryResponseContract and binary response media checks | Runtime/preflight response contract, not a bundle-load caller; loader uses separately typed binary declaration validators |
| DestinationBatch policy/max_records | Closed sync transport schema admits only per_record and 1..1000 before loader semantic validation |
| ValidateTransportExecutorFamily | Runtime executor-family compatibility consumer, not loader admission |

This table documents inspected residual branches, not final independent review or a waiver of other reachable siblings. D14 per-failure LoadAll metadata and final output/source reconciliation remain to close before handoff.

Required App package verification is currently FAIL, not green: cp15-consumers-normal-168-01 671/658 includes thirteen failing events at the existing typed-destination destructive-confirmation gate. Derived R1-base overlay probes reproduce the persisted-action and three independent sibling failures. Package-path setup errors are separately retained; corrected stores81 pass. Firstmate disposition requested under cp15-baseline-approval-contract-168 before altering approval policy/fixtures or assigning baseline debt. Full record: private cp15-168-approval-contract-block.md/.json.

## Firstmate169 disposition and complete diagnostic matrix

The decision key is resolved; **BD-CP15-168-01 remains unresolved**, owned by this canonical shared-code worker for CP16 before cumulative/reference-vertical acceptance. Firstmate169 permits a complete CP15 reviewable candidate with the explicitly failed App check. It does not prescribe severity, waive independent assessment, exclude inherited warnings from the final count, or authorize advance before review. Approval implementation and fixture semantics remain unchanged.

| Row | Actual current witness | Observation and limits |
| --- | --- | --- |
| D1 | TestCLILazyMalformedUnselected168; TestConstructionSelectedDiagnostic168; existing App lazy registry tests | Exact listed membership and zero unselected loads; healthy GitLab actual inspection/construction. Physical boundary proof is separate below. |
| D2 | TestAppPublicBundleDiagnostic167; TestCLISelectedStoreBundleDiagnostic168; TestConstructionSelectedDiagnostic168; TestSelectedMalformedBundlePhysicalBoundary168 | Actual malformed Load enters store-selected generation/digest through registry/App and CLI; original syntax/joined causes; construction/no-send and healthy returned-byte controls. |
| D3 | TestBundlePublicRateRules167/duplicate | Semantic duplicate policy index1, safe rule, retained cause; not a schema-intercepted duplicate hypothesis. |
| D4 | TestBundlePublicRateRules167 and TestBundleUnknownRateWithReasonRejectsPolicy168 | Closed state/source/scope/config rules and actual unknown-with-reason policy prohibition; earlier table label correction remains recorded. |
| D5 | TestBundlePublicRateRules167 and original migrated bundle tests | Policy/budget/cost/selector safe reasons, actual indexes, original cause predicates;95-expression dispositions remain explicit. |
| D6 | TestBundleDiagnosticIdentityAndCause165; TestBundlePublicDiagnostic167; TestBundleOptionalStatFailure167 | Metadata syntax/required-name/file/read distinctions and optional pure-absence versus compound failures; initial failed Stat observer attempts are not RED. |
| D7 | TestBundlePublicSchemaCompile168; TestBundleSchemaReferenceDiagnostic168 | Actual loaded schema filename/compile coordinate; never-selected reference stays at referencing field; underlying causes retained. |
| D8 | TestBundlePublicDiagnostic167; TestBundlePublicSchemaCompile168; TestBundlePublicDatabase168 | Sorted-member sentinel-safe locations for authored keys; fixed known properties; database unknown-member parser byte coordinate. |
| D9 | TestBundlePublicStreamHeaders167; TestBundlePublicMultipart167; TestBundlePublicWriteSemantics168 and sibling families | Static/vendor/media/closed/bounded multipart invariants and relative indexes; private direct-helper error contracts retained. |
| D10 | TestBundleDiagnosticSelectedGeneration165; TestBundleDiagnosticJoinedSelected168; selected App/CLI tests | Fresh selected generation/digest wrapper, original error identity preserved; success identity checks unchanged. |
| D11 | TestBundleDiagnosticJoinedSelected168; TestBundleOptionalStatFailure167; TestBundleCauseOracle165; TestPublicBundleOracleFalsification167 | Complete joined graphs, pure versus compound absence, wrong readable reason/identity rejection; public projection excludes sibling text. |
| D12 | TestAppPublicBundleDiagnostic167/connector; TestAppCredentialSelectionDiagnostic168 | App selectors preserve diagnostics; credential requirement precedes that endpoint’s selection. |
| D13 | TestCLIPublicBundleDiagnostic167; TestCLISelectedStoreBundleDiagnostic168; TestCLIBundlePreformattedAndReported168 | Actual dynamic command/inspect text and JSON, internal_error/exit1, structured selected bundle, safe wrapping, no duplicate already-reported output. |
| D14 | TestBundleMixedLoadAllAndSourceIndependence169; TestBundleLoadAllOneBadBundleDoesNotHideTheRest; D1/physical-boundary controls | Exactly two retained healthy and two named malformed bundles; each has safe metadata and original SyntaxError; malformed source-only inputs leave healthy execution identity unchanged. |

Each witness is named for independent checking, not counted as a distinct accepted requirement. Original RED/GREEN chronology and failed fixture corrections remain in TDD-LEDGER.md and immutable receipt captures. The current source package pair2416/2416 precedes only the added D14 test and authoring/docs updates; production Go bytes are unchanged. D14 receives separate normal/race proof. Final generated and selected consumer verification receipts remain to be added.

## Additional final-check observations169

Full CLI cp15-cli-normal-169-01 fails28796/28791. One obsolete sentinel-publication assertion is migrated as supplementary AM-169-01, preserving original95 expression dispositions. Current selected follow-up cp15-cli-migration-169-01 passes7/11, with the four baseline failures retained. BD-CP15-169-01 polling help, BD-CP15-169-02 source-origin refusal ordering (two tests), and BD-CP15-169-03 ETL help remain unresolved provisional observations, separately reported for Firstmate/final review. The derived34-file R1 Go overlay reproduces all four; it is not complete historical-tree proof. No severity, deduplication, CP16 assignment, or all-CLI-green claim is inferred. Exact tests, receipts and hashes: private cp15-cli-baseline-observations-169.json.

Final engine normal/race cp15-final-engine-{normal,race}-169-01 both pass2081/2081 after the one-cause optional-absence repair. Earlier2416 producer pair remains source-compatible for unchanged connectors/database; its engine result is superseded.

## Final169 reconciliation and source applicability

Final source-demands-check-169-02 exits0 with unchanged captured inputs; register SHA fe9b4bc8da544a1284ccdbaef8ded72b9b5c1fda67acf3e14f743d707d37e359. Universe30401, assessed26, exact complement30375 and known obligations26 remain unchanged. The latest generation changes only owner-source pins; coverage, known obligations, source fit and checks are byte-value equivalent. No stale proof is reminted.

Full engine normal/race2081 and physical publication230/canon results bind current production bytes. Subsequent test-only ST1023 correction is covered by35 normal/race. CLI63 race binds current CLI sources/tests. Connectors/database2416-group receipts bind unchanged respective producer source; engine component is superseded. Earlier selected store/App diagnostics remain meaningful through identical selected consumers plus current engine and CLI proofs; failed full-App approval and failed full-CLI baseline outcomes remain visible. Source manifest/Atlas/assessment bytes are unchanged since their successful checks; latest pure-absence helper does not alter those declarations. Runtime preflight/boundary/definitions results precede that helper repair; current actual physical publisher and optional-loader controls cover its affected transition, while these earlier results retain their precise historical pins. Tidy/release/agent-contract files are unchanged.

Final vet/build169-02, lint169-04, smoke169-02 and help-final169-01 pass. The final binary rebuild follows all production edits. Smoke asserts actual three returned outbox rows plus nonempty Parquet/owner evidence in an isolated fixture. Three actual help/namespace invocations pass; generated manual/golden and website diagnostic prose agree. Private final receipt index preserves every failed/green command, original hashes and differing current input paths; differing paths do not automatically imply acceptance or invalidate unrelated witnesses.

The original541-production-call multiset remains exact under the final961-Go-file AST scan. Portable390 structural and1039 lexical rows keep original pins and dispositions; INVENTORY-RECONCILIATION-169.json supplies current coordinates. Static census and positive rate tests together support owner coverage, not a claim that an AST selector proves all semantic completeness.

Official inline verify-work routes all eight SUMMARY obligations to independent judgment. No GSD auto-pass or phase acceptance is inferred. Firstmate169 explicitly schedules BD-CP15-168-01 only; BD-CP15-169-01..03 are separately reported for disposition/review. Final independent review must assess all inherited actionable observations and original CR-164-01 without a prescribed count or severity.

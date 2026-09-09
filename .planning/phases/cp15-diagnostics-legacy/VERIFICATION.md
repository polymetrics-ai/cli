# CP15 verification

Current Firstmate171 owner implementation and applicable local verification are complete for the preassigned fresh review; independent acceptance remains pending. PLAN.md is the sole seven-item plus CR-164-01 manifest. D1–D14 and all95 original assertion dispositions plus AM-169-01 remain accountable. All seven170 corrections are implemented together; historical full App/CLI failures below remain preserved and are superseded by the actual171 passes.

Full App691/691 and CLI28806/28806 normal passed. A later test-only Register error check is covered by final App39/39 normal/race and final lint/App vet. Production and CLI test bytes did not change after full normal checks. Full affected engine/store/connectors/registry2303/2303 normal/race and selected CLI77/77 race passed. These are exact event counts, not requirement counts or whole-repository race certification.

The following sections preserve source-scoped chronology, including failed setup and provisional carry statements. The final171 reconciliation at the end supersedes their then-current states. No provider-live, customer database, shared-service, final CI, no-mistakes, integration or acceptance claim.

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

## Firstmate171 corrected candidate verification in progress

Historical169 full App/CLI failures above remain original evidence, not current acceptance. The coherent seven-finding fixes and test-oracle migrations are recorded in REVIEW-CONVERGENCE and TDD-LEDGER. Focused current tests pass; whole App/CLI and affected race checks are running on frozen inputs.

Completed current production consumers: `cp15-engine-store-normal-171-01` passed2303/2303 events across full engine, manifeststore, connectors and bundleregistry packages (42.366s; output SHA256 b586dfc5e9ded2f03d9ed197aa9cb2506f8b73ebe331120f23828121c3e2eb92). `cp15-final-vet-171-01` passed current affected packages plus connectorgen/database, empty output legitimate. `cp15-final-lint-171-01` reports exactly one unchecked Register error in the new App fixture; deferred until active immutable-source captures end, then corrected fixture/lint verification required. No production failure is inferred from that lint finding.

Generated current source manifest is original renderer output, SHA256146782fd23f46e9f255478bcbf169c1e9593420e4edc11db207b0d65adcc57f1,101749993bytes. Only Atlas input hash/byte fields differ. Current demand register is original renderer output SHA256d6800f9c3443ce5f95f673a2e15c06b67a4f6ac7e657051a46c4cee7af2b7205,10241335bytes. Its64 differences are exclusively input/Atlas/assessment/source-manifest hash or byte pins. Coverage, known obligations, source-fit, examples, adopter relations, stale proof observations and checks remain exactly equivalent. Authoritative private comparison records are cp15-source-manifest-delta-171.json and cp15-demand-register-delta-171.json. Deterministic check commands are running; generation exit0 alone is not their result.

GSD sources for verify-work/code-review were resolved at pinned official20297a8ff941378b8615a5d3e8629e52c10a0f9d and full verify-work prompt executed inline under the existing non-Pi/Firstmate-owned role fallback. Full UAT disposition awaits the terminal evidence. Independent reviewer171 remains preassigned and unlaunched; this checkpoint is not acceptance.

## Final Firstmate171 owner reconciliation

All seven CR-170 corrections are implemented and locally verified. The original BD-CP15-168-01 and BD-CP15-169-01..03 failures remain historical; Firstmate171 explicitly required their correction here instead of provisional CP16 carry. REVIEW-CONVERGENCE records each original disposition and TDD-LEDGER preserves real RED, fixture/setup failures, test-oracle corrections and same-test GREEN limits.

| Current check | Exact scope | Result |
|---|---|---|
| cp15-full-app-normal-171-01 | Full App before final test-only setup error check | 691/691 events; 350.351s; raw `cc9b6857b079526419553d31ede69f8d6de544d037486ed5e035a016bf00a98c` |
| cp15-full-cli-normal-171-01 | Full CLI; final CLI test and production bytes | 28806/28806 events; 632.091s; raw `dd3a22850ddbdc5735c5d9d5502e1ca4662164fefa2411e094b76cc3e9b39905` |
| cp15-engine-store-normal-171-01 | Full engine/store/connectors/registry normal | 2303/2303 events; 42.366s; raw `b586dfc5e9ded2f03d9ed197aa9cb2506f8b73ebe331120f23828121c3e2eb92` |
| cp15-engine-store-race-171-01 | Same full packages race | 2303/2303 events; 229.132s; raw `800752fe0ebbb5b64e6b3255c16e5eb47c1e4afdf4cdaa2c29d80fbb05c7637c` |
| cp15-final-app-focused-normal-171-01 | Final App tests, original five plus both executor variants and selected diagnostics | 39/39 events; 39.411s; raw `f768c0ef4f48d488a1933fdc38803bdac5fbd3e68c51c93da999e64f367316c1` |
| cp15-final-app-focused-race-171-01 | Same final App39 events race | 39/39 events; 322.678s; raw `5f92f513c1b43b06d841c6dbe948f2e82f5be36c685d92ab1a08c1ac571b8fd1` |
| cp15-final-cli-focused-race-171-01 | Selected CLI diagnostics, source-shape carry, corrected origin/help and falsifiers | 77/77 events; 137.938s; raw `2683be45b0a37a9c5efbe569d7d6a645a668149424c2d297afa9093f922479d1` |
| cp15-final-lint-171-02 | New-from-original-base affected package lint | exit0; 10.945s; raw `e92606b0bf483111dff0a120c315ea165821348f31365020e2468a0059095c47` |
| cp15-final-vet-171-01 | Affected packages plus connectorgen/database before test-only cleanup | exit0; 4.396s; raw `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| cp15-final-app-vet-171-01 | Final App test cleanup | exit0; 1.857s; raw `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| cp15-source-lanes-check-171-01 | Current deterministic source manifest | exit0; 137.908s; raw `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| cp15-source-demands-check-171-01 | Current deterministic demand register | exit0; 197.963s; raw `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| cp15-atlas-verify-171-01 | Atlas symbol/proof selectors | 20/20 events; 20.956s; raw `ccca0a97c0275339abfb1fdb805320940a15ce244728902a5052756e7ef2b0fd` |
| cp15-build-171-01 | Current production binary in owned cache | exit0; 26.344s; raw `e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855` |
| cp15-smoke-171-01 | Owned isolated sample to warehouse to outbox; three exact returned rows/owned table | exit0; 6.367s; raw `53b182b706706d53efb0f0c75923f0cfe0be2ec42b1beb208e13be34cd504a2b` |
| cp15-docs-check-171-01 | Current binary connector docs validation | exit0; 13.306s; raw `bd93e0cd8de6b9962f4dfcc9936c53b10f83d0525bea43045089385fef059dc1` |

Current inventory reconciliation is INVENTORY-RECONCILIATION-171.json: all541 production call identities are unchanged;966 Go inputs include five added test files. The original390 structural rows and1039 lexical complement remain intact. Protected7 factories/49 hooks and unchanged rate/RequesterFor physical-send contracts retain original source-scoped proof and170 independent examination; this wave changes only transport confirmation policy and two diagnostic metadata branches, not runtime dispatch or request admission. Database diagnostic producer source, connsdk, hook factories, module/release/agent-contract/workflow files and user-facing docs are unchanged from their prior validated receipts. The complete final receipt index mechanically lists actual current input differences; source applicability is based on affected consumers, not whole-input equality alone. No unchanged costly199-command replay is claimed.

Atlas50 and generated manifest/register semantic comparisons preserve30401 cells,26 assessed/known obligations and30375 complement; no source/lane capability or stale proof credit changes. Current docs validation, full help/golden CLI cases and standalone sample reverse smoke preserve parity and ordinary action behavior. Original current carried CR-164-01 is independently resolved170 and remains protected by current connectors/CLI shape tests.

Final local Git/committed GSD evidence gate and exact candidate HEAD/tree are bound in the private cp15-candidate-171.json/cp15-review-171-binding.json and final receipt index after commit. Phase source is frozen before the preassigned independent Astra/xhigh review. Firstmate owns acceptance and subsequent integration.

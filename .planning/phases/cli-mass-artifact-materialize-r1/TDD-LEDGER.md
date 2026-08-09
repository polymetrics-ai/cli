# TDD ledger — 426 connector artifact sweep

## Static inventory contract

- Red: at recovery head, repository inventory found only 5 v2 provider-artifact bundles (`alpaca-broker-api`, `defillama`, `dockerhub`, `docuseal`, `flexmail`) and no canonical 426-target ledger. The sweep must not claim completion from an incomplete inventory.
- Green: the committed target ledger contains exactly 426 unique target rows with official-source provenance. The progress ledger reconciles recovered/materialized, explicit retry-pending, genuinely-blocked, and remaining rows exactly; `foundation_pending` remains a capability classification rather than an omission state.

## Bundle contract

- Red: a target with no `operations.json`/artifact provenance cannot demonstrate complete classified JSON surface.
- Green: each materialized or recovered target parses, loads, is discoverable, and records ETL, reverse-ETL, binary, direct-read, and direct-write classification counts; reachability is recorded where its foundation exists.

## Recovered seven-bundle compatibility

- Red: `go run ./cmd/connectorgen validate internal/connectors/defs` reported `additional property not allowed` for Jira and Workday REST `api_surface.json` rows using `covered_by.writes`. The recovered official operation surface cannot load without plural-write coverage support.
- Green: the narrow schema/loader/validator compatibility change accepts only the plural `covered_by.writes` form, keeps singular coverage unchanged, and the 551-bundle validation plus the recovered bundle load gate pass.

## Algolia artifact operation references

- Red: the first official-source batch dropped Algolia because its provider-published OpenAPI operation `DELETE /{path}` contains a non-empty `$ref` alongside local operation metadata; the extractor rejected that field before it could enumerate the documented method/path.
- Green: the parser accepts only a non-empty string operation reference while retaining the local method/path and metadata. `TestParseBatchOpenAPIArtifactAcceptsOperationReferenceWithLocalMetadata` passes, Algolia materializes with zero drops, and the post-generation 551-bundle static/preflight gate passes.

## Batch 002 registration repairs

- Red: Leadfeeder's documented aggregate method labels (`*` and `POST/PATCH/DELETE`) produced invalid `operations.json` IDs. ClickUp's required polymorphic custom-field `value` had no declared record-schema property, so generation could not mark the scalar CLI surface unavailable truthfully.
- Green: generated operation IDs sanitize every method label, proven by `TestMaterializedOperationIDSanitizesNonHTTPMethodLabels`; Leadfeeder's isolated retry validates. ClickUp declares the polymorphic `value` as an unconstrained JSON schema node, so its generated reverse-ETL command remains `not_implemented`/foundation-pending rather than inventing a scalar flag; its isolated retry validates.

## Xero stale direct-read coverage

- Red: Batch 003's staged Xero bundle reported eleven `covered_by.direct_read` targets without any implemented direct-read command; isolated `surface-reconcile` correctly refused to invent command coverage.
- Green: only those eleven stale claims were removed from the generated surface. The cached official artifact isolated retry validates, preserves the classified JSON, and is `foundation_pending` where runtime capability is absent.

## Primary-source retry discipline

- Red: a primary HTML/reference materializer drop had been counted as a terminal source blocker, which would omit an official-source alternative route.
- Green: each primary drop is recorded as `retry_pending` in `MATERIALIZATION-LEDGER.json` and `RETRY-QUEUE.json`, with the primary attempt plus pending official OpenAPI/Swagger, Postman/SDK, and reference-traversal routes. It cannot become genuinely blocked until those routes are exhausted and recorded.

## Preserved executable direct-read coverage

- Red: Batch 007's GitHub artifact replacement retained 173 existing `covered_by.direct_reads` claims but regenerated the CLI without their implemented commands, producing 173 invalid coverage references.
- Green: `TestBatchMaterializePreservesImplementedDirectReadCoverage` proves that each existing implemented direct-read or binary-download command named by a retained coverage entry is preserved. The isolated GitHub retry validates with 173 covered names and 173 implemented command names.

## Documented partner/enterprise artifact access

- Red: a cited official OpenAPI record for a partner- or enterprise-gated provider could not enter a manually evidenced static materialization manifest even when its provider-owned reference was publicly retrievable.
- Green: `TestReadBatchManifestAllowsDocumentedPartnerAccess` permits only `public`, `partner_gated`, and `enterprise` access models in an already evidenced manifest; automatic selection remains public-only and no credentialed execution is introduced.

## Official versioned artifact URLs

- Red: `TestBatchArtifactURLAndDestinationGuards` rejected the official Google Discovery URL `https://analyticsdata.googleapis.com/$discovery/rest?version=v1beta`, preventing static extraction of a cited provider artifact.
- Green: primary artifact retrieval admits only an explicit non-sensitive query-key allowlist (including `version`) and uses the same guarded request path as official-reference retrieval; token-bearing URLs remain rejected. The focused guard test passes.

## Google Discovery official artifacts

- Red: six exact-count Google Discovery REST documents were provider-owned machine-readable JSON but were rejected as non-OpenAPI, leaving their complete official operation inventories unavailable to the materializer.
- Green: `TestParseBatchGoogleDiscoveryArtifactEnumeratesNestedMethods` proves the bounded extractor recursively preserves documented HTTP method/path provenance while mapping reserved URI-template variables to the connector-safe path form. The 10-connector Discovery recovery batch validates with zero findings.

## Shared endpoint coverage and empty operation catalogs

- Red: the Discovery recovery encountered legacy bundles that intentionally map several ETL streams to one documented request; duplicate source rows either halted materialization or made generated CLI references ambiguous. A no-operation bundle also serialized `operations: null`, which the schema rejects.
- Green: `TestMaterializeAPISurfaceRetainsDuplicateCoveredStreamBindings` proves shared stream bindings merge into one `covered_by.streams` row while retaining all ETL references and connector-relative paths. `TestMaterializeOperationCatalogKeepsEmptyOperationsArray` proves a missing source catalog serializes as `[]`. The combined 10-bundle recovery validation is green.

## Registered ETL and reverse-ETL command retention

- Red: `TestBatchMaterializeGeneratesV2ProvenanceAndReachableSurface` failed after the generic materializer replaced the registered `widget list` and `widget create` paths with generated names. The same defect erased Gong's established `calls create` reverse-ETL registration during its artifact recovery.
- Green: the materializer retains one source CLI command per declared stream/write target, refreshes only its cited `api_surface` references, and rejects ambiguous duplicate target registrations. The focused materializer regression and Gong's isolated static validation pass with its registered paths preserved.

## Non-request OpenAPI operation metadata

- Red: `TestParseBatchOpenAPIArtifactIgnoresNonRequestOperationMetadata` failed because an official operation carrying `callbacks` rejected the entire provider request inventory; Finnhub separately proved the same issue for an `examples` field.
- Green: the extractor now ignores `callbacks` and `examples` as non-request metadata while never inventing callback deliveries as API endpoints. The focused parser regression and its existing fail-closed unsupported-container coverage pass.

## Derived implemented write flags

- Red: `TestBatchMaterializeConvertsLegacyExclusionsAndDerivesWriteFlags` failed after preserving a complete existing write command retained stale optional `record.name` metadata instead of deriving the required write contract from the bundle schema.
- Green: implemented writes retain their registered command path and summary while their flags, availability, risk, and output metadata are regenerated from the declared write schema. Existing `partial` commands retain their connector-owned contract. The focused write-flag and Gong full-surface regressions pass.

## Provider JSON compatibility

- Red: `TestParseBatchOpenAPIArtifactAcceptsJSONUnicodeEscapes` reproduced Twilio's valid JSON surrogate pair rejection by the YAML decoder; `TestParseBatchOpenAPIArtifactIgnoresProviderNeutralFreeTier` reproduced Finnhub's non-request `freeTier` annotation blocking its documented request inventory.
- Green: only after YAML decoding fails, valid JSON is decoded with `UseNumber` and re-marshaled before retrying YAML parsing; the operation inventory ignores the documented neutral `freeTier` annotation alongside other non-request metadata. Both focused regressions pass, and the isolated Twilio retry materializes and validates. Finnhub's separate `highUsage` field remains queued for an alternative-source retry rather than broadening the parser without a new focused proof.

## Swagger base-path preservation

- Red: `TestParseBatchSwaggerArtifactPrefixesBasePath` showed a standard Swagger 2.0 `basePath: /api` was discarded, so Whisky Hunter's documented `/auctions_data/` artifact could not reconcile with its existing `/api/auctions_data/` ETL bindings.
- Green: the parser validates and prefixes a Swagger base path before coverage reconciliation. The focused regression passes, and Whisky Hunter's one-connector retry validates, loads, and retains both shared stream bindings without inventing a command.

## OpenAPI server base-path preservation

- Red: ActiveCampaign's provider OpenAPI 3 fragments declare the v3 API root in `servers[0].url` (`https://{youraccountname}.api-us1.com/api/3`) while their `paths` are root-relative. The extractor returned `/contacts` rather than `/api/3/contacts`, causing all 61 legacy classifications to become false artifact-absent discrepancies.
- Green: TestParseBatchOpenAPIArtifactPrefixesServerBasePath first reproduced the root-relative result, then passes with the literal /api/3 request base path retained from a templated official server URL. The parser admits a server base only when every declared server has the same valid literal request path; ordinary OpenAPI provenance and Swagger basePath behavior remain unchanged. go test -count=1 ./cmd/connectorgen -run '^(TestParseBatchOpenAPIArtifactPrefixesServerBasePath|TestParseBatchSwaggerArtifactPrefixesBasePath)$' passes.

## 426-target rate-limit declaration completeness

- Red: find internal/connectors/defs -name rate_limits.json returned zero files, while the one-flow captain order requires exactly one provider-cited declaration for every 426 target. TestEverySweepTargetLoadsRateLimitDeclaration is added before production changes and must fail until every target has an embedded, schema-loadable declaration.
- Green: `scripts/materialize_cli_rate_limits.go` builds the immutable 426-entry `RATE-LIMIT-SOURCE-LEDGER.json` and writes no more than 40 declarations per atomic report. Its focused source-gap, bounded-selection, overwrite-refusal, and declared-over-unknown preservation tests pass. On a later rebase or batch resume, a valid provider-cited `declared` bundle record promotes the compact source ledger and is never replaced by a generated `unknown` fallback; reconciliation counts the on-disk declaration. The eleven B042 reports total 3 `declared`, 422 `unknown`, 1 `not_applicable`, and 426 files at this checkpoint. `TestEverySweepTargetLoadsRateLimitDeclaration` and `TestProductionDefinitionsEmbedEveryRateLimitDeclaration` load all 426 through production `defs.FS`; the reconciler refuses a ledger whose rate-limit state or file totals do not conserve the target manifest. `faker` alone is `not_applicable` because the target ledger records no external provider HTTP/API; every other non-declared policy carries the exact retained official-source publication gap rather than an inferred numeric limit.

## Complete root-reference fast path

- Red: `TestCompleteBatchHTMLReferenceRootAvoidsTraversalWhenCountIsMet` was absent while a 40-connector reference pass crawled unrelated provider links even where a root page already enumerated its entire immutable ledger count.
- Green: an `html_reference` root is now accepted without traversal only when its own explicit request inventory meets the connector's ledger count. Plausible's four-operation official root materializes and validates through that path; PyPI's incomplete normalized root remains retry-pending for a distinct official source rather than being overstated.

## HTML angle path placeholders

- Red: `TestNormalizeBatchHTMLOperationPathConvertsAnglePlaceholders` reproduced the provider-documented PyPI path `/pypi/&lt;project&gt;/json` being truncated rather than classified as the `{project}` request path.
- Green: only a complete HTML-reference root may translate a valid angle placeholder to its connector-safe `{name}` equivalent; malformed residual angle markup remains rejected. The focused regression passes and PyPI's isolated retry validates through the official root without broad reference traversal.

## Static reference assets

- Red: `TestParseBatchHTMLReferenceSkipsStaticAssetCandidates` reproduced a provider reference's `/build/css/api-doc.css` being selected as an API document solely because its filename contains `api-doc`, which prevented traversal of the linked OpenAPI export.
- Green: static stylesheet, script, source-map, font, and image suffixes are discarded before provider-host/admission checks. The focused regression passes alongside malformed, unsafe-query, and off-host selected-link fail-closed regressions; actual documentation and machine-artifact URLs remain candidates.

## Off-host navigation links

- Red: `TestParseBatchHTMLReferenceSkipsOffHostNonMachineNavigation` reproduced an external support URL containing generic `/api/reference` text aborting traversal before a same-host OpenAPI export could be read.
- Green: an off-host link fails closed only when it explicitly advertises a machine-artifact hint (`.json`, `.yaml`, `.yml`, OpenAPI, Swagger, or Postman); generic off-host navigation is skipped and never fetched. The focused regression passes together with the original off-host OpenAPI refusal.

## Exact existing-surface official-evidence fallback

- Red: 40 retry-pending connectors already carry a complete, classified source surface whose normalized endpoint count exactly equals the immutable official-survey count, but their current primary provider document is a reference index or an otherwise non-consumable artifact. Replaying reference traversal would neither add evidence nor finish JSON materialization.
- Green: an explicit `--existing-surface-evidence` materialization mode requires the cited official artifact bytes, an exact pre-existing endpoint-count match, and a source bundle with an API surface. It upgrades only that existing inventory to v2 endpoint-local provenance, regenerates operation/CLI JSON, and rejects a count mismatch rather than silently accepting partial evidence. The focused regression covers the success and mismatch cases.

## Explicit unversioned official-reference provenance

- Red: an official HTML reference that publishes no version marker was rejected solely because `artifact.version` is normally required, even when its preserved endpoint inventory exactly matched the immutable survey. A broad relaxation would allow ordinary unversioned artifacts to bypass provenance validation.
- Green: `TestBatchMaterializeAllowsExactUnversionedOfficialReference` proves an explicitly declared unversioned official HTML reference materializes only through `--existing-surface-evidence`, emits the `provider-publishes-no-version-marker` provenance value, rejects a non-evidence invocation, and rejects a 3-versus-4 immutable-count mismatch. `TestReadBatchManifestRejectsUndeclaredUnversionedArtifact` proves ordinary artifacts still require a nonempty version. `go test -count=1 ./cmd/connectorgen -run 'TestBatchMaterializeAllowsExactUnversionedOfficialReference|TestReadBatchManifestRejectsUndeclaredUnversionedArtifact|TestBatchMaterializeUsesExactExistingSurfaceEvidence'` passes.
- Skills: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, and `golang-safety`.

## Shared endpoint bindings in direct evidence materialization

- Red: the first B028 staging preflight found SearXNG's two ETL streams bound to one `GET /search` provider request emitted as duplicate v2 endpoint rows. The generated CLI can only resolve one coverage classifier for a method/path, so `search list` appeared to target the `reddit` binding.
- Green: direct evidence keeps the exact survey-entry count as its admission guard but deduplicates same method/path artifact rows before re-materialization; the existing coverage merger produces one `covered_by.streams` endpoint. The focused duplicate-binding regression and B028 staging validation demonstrate that no command registration is invented or lost.

## Ledger reconstruction and atomic replacement

- Red: an ad-hoc `jq` ledger update evaluated a candidate-selection expression with the wrong input scope after shell redirection had already opened the destination. The command therefore replaced `MATERIALIZATION-LEDGER.json` and `RETRY-QUEUE.json` with zero-byte files and left `RUN-STATE.json` with null counts, despite every connector bundle and batch report remaining intact.
- Green: `scripts/reconcile_cli_mass_artifact_ledgers.go --expected-materialized 194 --check` reads only the target manifest, retained batch reports/outcomes, the reconciliation record, and current bundle provenance. It rebuilt all three ledgers only after asserting 426 unique target names, a disjoint 232/194 queue/resolved partition, and `194 materialized + 232 retry_pending + 0 blocked = 426`. Its JSON writer stages validated non-empty temporary files before atomic rename; `TestStageValidatedJSONBytesLeavesDestinationUntouchedForInvalidCandidate` proves an invalid candidate cannot replace an existing ledger.

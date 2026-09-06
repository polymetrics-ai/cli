# Batch R1 vNext cutover verification

## Invariants

1. The runtime executes only rendered JSON. `source.lock.json` and provider provenance are authoring inputs, never runtime/admission inputs.
2. A native `mode=fixture` value cannot produce a successful check or canned record. It must reach ordinary validation before any provider request.
3. No native/hook compatibility reader can replace a declared engine execution route.
4. No generated or explanatory surface instructs authors or users to use a removed importer, certification, retention, fixture, compatibility, feature-flag, or second executor route.
5. Every migrated connector has exactly the seven declared lanes and reaches its normal credential/approval boundary without provider I/O.

## TDD ledger execution

| Slice | RED command/proof | GREEN command/proof |
| --- | --- | --- |
| Native fixture bypass | `go test -timeout 20m ./internal/connectors/native/alpha-vantage -run '^TestFixtureModeNoLongerBypassesCredentialBoundary$'` fails while fixture mode returns success. | The same test proves `mode=fixture` returns the ordinary missing-credential error without a request; the native residual guard is green. |
| Compatibility execution | The frozen review maps every `StreamHook`/`CheckHook` native delegation that can bypass declarative execution. | Targeted engine/hook/native tests prove only declared engine execution reaches the credential/approval boundary. |
| Source-lock render | A changed lock is absent or render drift is observed before its output is generated. | `go run ./cmd/connectorgen lock-render <connector>` followed by `--check` is byte-stable for each reference and migrated connector. |
| Credential boundary | A selected declared command has no connector-local witness before migration. | An isolated project with no stored credential returns ordinary missing-credential or approval-boundary output before provider I/O. |

## Exact residual scan

The scan is deliberately separated by data class. Test fixture values and provider evidence are not runtime paths; production Go, execution JSON, generated documentation, and website sources must not retain the forbidden behavior.

| Scope | Exact search expression | Required disposition |
| --- | --- | --- |
| Production Go: `cmd/connectorgen`, `internal/connectors/native`, `internal/connectors/hooks`, `internal/connectors/engine`, `internal/connectors/commandrunner`, excluding `_test.go` | `(?i)(fixtureMode|readFixture|mode[[:space:]]*[:=][[:space:]]*["']fixture|legacy[[:space:]_-]*(reader|shim|execution|admission)|compatib[a-z]*[[:space:]_-]*(reader|shim|execution|admission)|source[[:space:]_-]*(import|projection|admission)|certification[[:space:]_-]*(runtime|admission)|retention[[:space:]_-]*(runtime|admission)|feature[[:space:]_-]*flag)` | Zero execution/admission hits. Standard language-level compatibility or value fallback is classified separately and is not an execution route. |
| Execution definitions: `internal/connectors/defs/**/{spec,metadata,streams,writes,operations,cli_surface,sync_transport,rate_limits}.json` | `(?i)("mode"[[:space:]]*:[[:space:]]*"fixture"|fixture[[:space:]_-]*(mode|execution)|certification|retention[[:space:]_-]*(admission|runtime)|source[[:space:]_-]*(import|projection)|legacy[[:space:]_-]*(reader|shim))` | Zero operational hits; any remaining `source.lock.json` is provider provenance and is never embedded. |
| Human/generated surfaces: `docs`, `website`, `.agents`, `internal/cli` manuals, generated skills | `(?i)(mode[[:space:]]*=[[:space:]]*fixture|fixture[[:space:]_-]*(mode|execution)|source[[:space:]_-]*import|certification[[:space:]_-]*(gate|admission|runtime)|retention[[:space:]_-]*(gate|admission|runtime)|legacy[[:space:]_-]*(reader|shim|execution)|compatib[a-z]*[[:space:]_-]*(reader|shim|execution))` | Remove stale instruction. A retained provider-provenance mention is listed by exact path with why it cannot execute or admit a command. |

The final evidence records every nonzero result with path, quoted role, and classification; absence is never inferred from a partial page of search output.

## Required generation and checks

After source changes settle:

```text
go build ./cmd/pm
./pm docs generate --dir docs/cli
./pm skills generate --dir docs/skills --json
node website/scripts/gen-github-cli-surface.mjs
```

Focused cleanup checks:

```text
go test -timeout 20m ./cmd/connectorgen ./internal/connectors/defs ./internal/connectors/native/... ./internal/connectors/hooks/... ./internal/connectors/engine ./internal/connectors/commandrunner ./internal/cli
go build ./cmd/pm
go run ./cmd/connectorgen lock-render github --check
go run ./cmd/connectorgen lock-render gitlab --check
go run ./cmd/connectorgen lock-render asana --check
./pm docs validate --connectors-dir docs/connectors
```

Broader checks:

```text
go test -timeout 20m ./internal/connectors/engine ./internal/app ./internal/cli
git diff --check
```

The final delivery record also runs the secret/local-state scan for tokens, private keys, `.polymetrics`, vault, and `.env` paths; records available disk space; and records the exact reviewed source SHA, tests, residual result, connector witnesses, and named remaining gaps.

## Results

- G0 direct-parent amendment: `git fetch origin fm/cli-top100-declaration-batch-r1` refreshed the parent. `HEAD` and `origin/fm/cli-top100-declaration-batch-r1` both resolved to `d260b725ce6f53403961d7af1ef48ea6651cdd66`; `HEAD` is an ancestor of that remote tip. The immutable `main` merge base is `813f457a925f7ee3fe3bea101a43e445992c8552`, and #4325 fixes the delivery denominator at 4,341 primary retained source operations.
- G0 routing/local-state: #4325 comment `5500153864` and #4294 comment `5500165004` were read directly. The former certification tree was deleted at the frozen checkpoint. The untracked opaque `internal/connectors/certifications/.fingerprint-salt` has no tracked/index/history entry and is not ignored; it remains unread, unstaged, unmodified, and excluded from this work. It is recorded as local-state residue, never a retained certification path.
- G0 green gate: satisfied by planning commit and ordinary remote update `1655123262586b2eaa395aa75b0e54bd7c4558bd`; remote read-back matched exactly before N1 began.
- #4423 N1 RED: the former named defs proof could not compile because it read removed `Bundle` fields; the two Atlas proof selectors returned `no tests to run`, and explicit `go test -list` count assertions failed. This establishes the inherited handoff as red rather than accepting a zero-test exit status.
- #4423 N1 focused GREEN: `go test -count=1 -timeout 20m` passed for `TestRuntimeEmbedContainsExecutionJSONOnly`, `TestVNextSourceLockDeterministicallyRendersReferenceConnectors`, `TestVNextReferenceConnectorsDiscoverFromExecutionJSONOnly`, `TestVNextReferenceLockClosedSetRejectsUnrenderedArtifact`, and `TestVNextFoundationProofSelectorsResolve`. Each Atlas target was listed exactly once before execution. The full `cmd/connectorgen` and `defs` package suites passed. The closed-set reference test reads GitHub, GitLab, and Asana source locks and verifies their complete in-memory rendered output sets without writes or provider I/O.
- #4423 named remaining gap: `go test -count=1 -timeout 20m ./internal/connectors/engine` fails in unrelated `TestOperationRoutesFailClosedBeforeProviderIO`; N1 changes that package only by renaming the Atlas test selector. The named engine proof passed, but no broader-engine green claim is made. Render/generation, residual scan, secret/local-state scan, and final exact-SHA review remain pending for their later authorized gates.
- #4423 review: fresh-context manual re-review of `1655123262586b2eaa395aa75b0e54bd7c4558bd..eae04af74b6546f9c61130d426f06197723f4f22` found no N1 issue; its exact scope and safety analysis are recorded in `REVIEW-CONVERGENCE.md`. The review applies to the code-bearing N1 commit; the following evidence-only commit does not change production or test code.
- Independent N1 exact-range review over `1655123262586b2eaa395aa75b0e54bd7c4558bd..c5bf5c5d544e85dcca5eac3ebed45ba78ad7fb33` is **BLOCK** as an S1A unlock, not a passing review. It requires executable S1A RED gates for fixture credential bypass, production registry/binary surface coverage without skips, executor collisions, and hostile-origin credential taint before deletion. Its correction map is applied under already-accepted D1/D2; no new shared foundation is authorized.
- S1A cannot claim the source-lock authoring foundation atomically materializes a closed bundle or strictly admits immutable JSON until the approved D2 transaction and source-lock decoding repair actually land. The Atlas must carry those open gaps while the current renderer remains per-file and accepts trailing documents.
- S1A RED is executable and recorded before production deletion. Each exact selector count returned `1`. Alpha Vantage RED exited `1` because `mode=fixture` returned success and an untrusted `base_url` received one request. Bundle-registry RED exited `1` because the native overwrite erased Google Calendar's command surface and `Registry.Register` silently overwrote a duplicate name. No provider endpoint or credential store was used; the hostile-origin proof used only a local `httptest` server and does not log its synthetic test value.
- S1A first GREEN: the Alpha Vantage fixture/hostile-origin selectors, bundle-registry no-skip/collision selectors, and protected-database non-regression selector passed with `-count=1 -timeout 20m`. API bundle entries retain their rendered engine executors; PostgreSQL, MySQL, and DynamoDB retain their existing native registrations. This is not final cleanup or certification: legacy API native directories, delegating hook shims, stale generated wiring/docs, and the residual scan remain open.
- Amazon SQS exact disposition proof: absent-lock `lock-render --check` RED failed; source-lock render plus `--check`, `TestAmazonSQSDispositionBundleLoads`, and `TestEveryImplementedCommandPassesRuntimePreflight` GREEN passed. `pm connectors inspect amazon-sqs --json` shows the evidence-backed no-provider-action metadata; `pm amazon-sqs queue delete --json` exits policy-blocked before credentials or provider I/O. Inspection retains legacy stream/write data because the approved D2 closed-bundle publication transaction is not implemented; this remains R1-05, not a green claim for stale artifacts.
- Exact residual scan, production scope: the Captain-scoped 34 API native override registrations are removed, 34 unreachable API native package directories are deleted, and the retained Alpha Vantage boundary package has no fixture or caller-controlled-origin branch. The default registry admits rendered API executors plus only protected DynamoDB, MySQL, and PostgreSQL native registrations. The production scan has no API fixture/executor branch. Its remaining matches are the `rejectVNextLegacyExecutionEvidence` enforcement function and test, the Alpha Vantage regression-test name, and Captain-protected DynamoDB/PostgreSQL database fixture/transport paths. `RegisterSnapshotTransportSource` is a protected PostgreSQL compatibility-named test helper whose comment states the old page loop is unreachable; it is not an API execution route.
- Exact residual scan, tracked human-source scope: no stale API fixture instruction remains. Remaining matches are RLM fixture mode (a separate local RLM command), protected PostgreSQL generated manual/skill text, generated catalog descriptions for provider/test fixture facts outside this API-only S1A scope, and CLI test fixtures. Website build-cache output under `website/.next` is local generated state, excluded from website-source evidence and left untouched.
- Causal separation used a clean `git archive c5bf5c5d544e85dcca5eac3ebed45ba78ad7fb33` snapshot under `.cache/c5bf5c5-causal-baseline` and the same `-count=1 -timeout 20m` selectors. Engine's `TestOperationRoutesFailClosedBeforeProviderIO` failure is baseline-identical. App's typed-destination approval/I-O failures are baseline-identical. CLI's polling-help, source-bound-origin, and ETL-help failures plus local Redis connection-refused logs are baseline-identical environment/fixture failures.
- S1A-introduced app divergence: `TestProductionParkingCompositionSurvivesProcessKill` and two typed-destination drift tests passed at the baseline but failed after duplicate registration became reject-only. The tests now construct explicit replacement registries or isolated icon-valid test registries rather than silently overwriting a production name; each targeted selector passes and the current full app suite now contains only the baseline failures.
- S1A-introduced CLI divergence: the durable-parking helper silently failed after its sample override was rejected, and the nine root golden transcripts changed because Amazon SQS now has the visible source-lock disposition while engine-backed Google API surfaces are no longer hidden by native overrides. The helper now uses an isolated icon-valid test registry; `TestCLIDurableParkingAdmissionAndResumeAcrossKilledProcess` passes; `POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1` regenerated the changed root fixtures; and the current full CLI suite now contains only the baseline failures.
- Full API migration verification is pending by captain decision: every one of the 28 families must contribute its exact RED/GREEN command, source-lock render/check, seven-lane record, local-fake-server generic check/read/ETL witness, production-registry witness, and generated-surface result. Amazon's policy-blocked disposition remains a non-generic exception; DynamoDB/MySQL/PostgreSQL are protected non-regression controls. A new exact-SHA review will replace, not extend, the review of `ba1a58848`.
- Feishu RED/GREEN: the first local-only production-registry proof failed because the former bundle resolved `example.invalid`. The schema-4 lock now renders the declared Feishu/Lark host enum, fixed Bitable check/read routes, cursor pagination, and source-selected declared-route tenant authentication. `TestProductionFeishuGenericCheckAndRead` and `TestProductionFeishuRejectsFixtureAndCallerOriginBeforeIO` each selected exactly once and passed with `-count=1 -timeout 20m`; the latter observed zero rejected-origin/fixture requests. `go run ./cmd/connectorgen lock-render feishu --check` passed. This is a `runtime.provider-extension-seams.v1` reuse, not a new shared foundation, and it had no provider I/O or credential-store access.
- FreeAgent RED/GREEN: the local-only production proof first failed at the stale `example.invalid` route; the focused refresh-token proof then failed because the generic encoder sent no HTTP Basic client authentication. `source.lock.json` now selects the fixed FreeAgent OAuth2 endpoint, standard HTTP Basic client authentication, and fixed paginated read routes. The production check/read and rejected-origin/fixture proofs pass with one selector each, `TestOAuth2RefreshTokenBasicClientAuthentication` plus its invalid-mode rejection proof pass, and `lock-render free-agent-connector --check` passes. The Atlas records this as an `authoring.source-lock-vnext.v1` constrained extension; no provider I/O, hook/native executor, or caller-controlled origin remains.
- Freightview RED/GREEN: the initial local-only production proof failed at `example.invalid`. The source lock now renders only the fixed v2.0 client-credentials route, continuation-token pagination, and request-backed shipment fan-out. `TestProductionFreightviewGenericCheckAndRead` proves check/read plus a stamped quote fan-out; its origin/fixture companion observes zero requests. Both selectors count once and pass with `-count=1 -timeout 20m`, and `lock-render freightview --check` passes. This reuses the existing generic execution contract with no provider I/O or native/hook path.
- Google Analytics Data RED/GREEN: the production registry initially sent the legacy check to `example.invalid`; the new generic foundation RED tests showed positional rows were emitted unprojected and malformed header/value responses were accepted. Captain approved the bounded `runtime.response-header-projection.v1` contract. It is source-selected by GA4’s five schema-4 report streams, rejects unknown/duplicate/missing/mismatched values before emission, and projects only declared response names. Focused foundation tests plus GA4 production check/read and origin/fixture refusal pass; `lock-render google-analytics-data-api --check` is current. No hook, caller-origin, provider I/O, credential mutation, or database path is involved.
- Regression checkpoint after the GA4 foundation: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry` passed; `go test -count=1 -timeout 20m ./cmd/connectorgen` passed; `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings. The full engine package still has the recorded baseline failure `TestOperationRoutesFailClosedBeforeProviderIO`, whose expected source-traced route diagnostic differs from the current missing-route diagnostic; the new response-header selectors passed independently.

- S1A tenant/session/array-zip RED: `TestTenantOriginOverridesDefaultOperationRoute` failed against the static fallback route; `TestProductionMetabaseDeclaredSessionCheckAndRead` failed at `example.invalid/__legacy_hook/check`; `TestArrayZipProjection*` initially failed to compile because the bounded foundation was absent; and `TestProductionYahooFinancePriceArrayZipCheckAndRead` failed at Yahoo's legacy hook route. Every RED used an in-process local transport or compile failure, with no credential store or provider request.
- S1A focused GREEN: `go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/bundleregistry -run '^(TestTenantOriginOverridesDefaultOperationRoute|TestDeclaredSessionAuthRejectsDynamicTokenRoute|TestDeclaredSessionAuthUsesResolvedTenantOrigin|TestProductionMetabaseDeclaredSessionCheckAndRead)$'` passed. The Yahoo foundation and witness command `go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/bundleregistry -run '^(TestArrayZipProjectionCombinesDeclaredArrays|TestArrayZipProjectionRejectsMismatchedArrays|TestProductionYahooFinancePriceArrayZipCheckAndRead)$'` passed. These witnesses assert fixed local routes, bounded request counts, no password leak beyond the session exchange, selected existing-session header behavior, and two Yahoo OHLCV records only after validated equal-length zipping.
- S1A authoring and Atlas GREEN: `lock-render --check` passed for `metabase`, `prestashop`, and `yahoo-finance-price`; `connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings; the Foundation Atlas JSON/schema and unique-ID checks passed; and `TestFoundationAtlasSelectorsResolve` passed. The Atlas records `runtime.tenant-origin.v1`, `runtime.declared-session.v1`, and `runtime.array-zip-projection.v1` with their owners, closed selectors, constraints, and executable proofs.
- S1A generated surfaces: after `go build ./cmd/pm`, `./pm docs generate --dir docs/cli`, `./pm skills generate --dir docs/skills --json`, and `node website/scripts/gen-github-cli-surface.mjs` completed, `./pm docs validate --connectors-dir docs/connectors` passed. No credential resolution or provider I/O occurred.

- S1A final static-gate blocker: `go vet ./...` and `go build ./cmd/pm` passed, but `go run ./cmd/agentcontractgen check` exited 1 before `git diff --check` could run. Its exact blocker is duplicate project-agent name `pm-connector-worker` in the preserved local `.cache/c5bf5c5-causal-baseline/.claude/agents/pm-connector-worker.md` and the tracked `.claude/agents/pm-connector-worker.md`. The cache is pre-existing local residue; this task must neither delete, clean, move, inspect further, stage, nor modify it. Final static gates, exact-SHA review, commit, and parent-branch push remain blocked pending Firstmate's cache disposition.

- Final-remediation RED/GREEN: `TestCheckProjectionsIgnoresNestedCacheClaudeAgents` first failed because the inventory recursively inspected `.cache/.../.claude/agents`; it now passes with the root `.claude/agents` duplicate/unexpected-definition regression selectors. `go run ./cmd/agentcontractgen check`, `fm-ensure-agents-md.sh .`, and `git diff --check` pass. The nested cache and `.fingerprint-salt` remain untouched.
- Final local static gates: `go vet ./...`, `go build ./cmd/pm`, `make tidy-check`, generated docs validation, agent-contract check, connector validation, vNext lock checks for GitHub/GitLab/Asana, commandrunner implemented-command preflight, connector boundary, connector canon, and release-workflow checks pass. The boundary report is clean, with only its three recorded expires-on-2026-09-30 documentation exceptions.
- Final focused suites pass: full `cmd/connectorgen`, `internal/connectors/defs`, and `internal/connectors/bundleregistry`; the tenant/session/array-zip and agent-root-inventory focused selectors; and `TestFoundationAtlasSelectorsResolve`.
- Known pre-existing broad failures reoccur unchanged: engine `TestOperationRoutesFailClosedBeforeProviderIO`; app typed-destination approval/I-O expectations; CLI polling/source-bound-origin/ETL-help expectations with local Redis refused-connection logs; `internal/agentcontract`'s unchanged `TestRenderIsStableAndConnectorInheritsBase` connector-hash assertion; and lint findings confined to pre-existing `auth_declared_password_token.go`, `polling_watermark_test.go`, and unused `cmd/connectorgen` helpers. The current canonical contract and that hash test are byte-identical to `c5bf5c5d544e85dcca5eac3ebed45ba78ad7fb33`; no baseline artifact was rewritten.
- Final scans: `git diff --check` passes; the scoped secret scan finds no private-key, GitHub-token, OpenAI-token, or AWS-key pattern; working disk has 224 GiB free. The worktree has 0 staged files, 397 unstaged files, and 44 untracked paths; only the known cache and `internal/connectors/certifications/.fingerprint-salt` local residues remain excluded.

- Exact-SHA review remediation: Ashby now rejects `base_url` and `mode` before credential/transport work; PrestaShop now proves Basic access-key Check+Read over real tenant API resource paths, query windows, nested extraction, and incremental declaration; tenant origins reject arbitrary path prefixes; and Yahoo `chart.error` produces a typed pre-emission error with zero records. All named adversarial production/local selectors pass.
- Remediation generation: complete `npm --prefix website run gen:website-data` regenerated GitHub CLI, docs data, connector bundles/catalog, and connector data; `npm --prefix website run test:scripts` passed 34/34. `pm docs generate` and skills generation regenerated Ashby/PrestaShop manuals, skills, and catalogs; docs validation passes.
- Remediation gates: Ashby/PrestaShop/Yahoo render checks, full definition validation, full `cmd/connectorgen`/`defs`/`bundleregistry` suites, vet, build, agentcontract, diff, docs, command-preflight, boundary, canon, and release-workflow gates pass. The known baseline engine missing-route diagnostic still reoccurs unchanged.

- Second re-review remediation: all 71 Ashby ETL streams now have source-backed closed POST JSON bodies, body cursor contracts, success envelopes, and body-mapped ETL command flags. The known-larger local witness proves exactly two no-query pages; the success=false witness emits zero records. PrestaShop declares `runtime.offset-count-pagination.v1` and proves `limit=0,100` then `limit=100,100` with resource-keyed extraction and client filtering. Every network-capable engine connector method validates strict configuration before requester construction; hostile Ashby direct operations issue zero request.
- Documentation and session remediation: 21 migrated API docs inputs plus Ashby are semantic-parity-tested against rendered check/stream/body/pagination/error facts, and stale executor language is forbidden. Malformed or missing declared-session responses are rejected without a header or replay. Full docs, skills, catalog, and complete website data generation complete; website script tests pass 34/34.
- Second remediation gates: Ashby/PrestaShop/Yahoo lock checks and full definitions validation pass; full `cmd/connectorgen`, `defs`, `bundleregistry`, and `commandrunner` suites pass; vet, build, agentcontract, and diff checks pass. The same pre-existing engine missing-route diagnostic remains the only re-run engine-suite failure.

- Final second re-review gates: all Ashby/PrestaShop/Yahoo render checks, complete definition validation, focused second-review adversarial tests, full `cmd/connectorgen`/`defs`/`bundleregistry`/`commandrunner` suites, docs validation, complete website-data generation and 34/34 website script tests, vet, build, agentcontract, and diff checks pass. Scoped source and complete generated-data secret scans found no key/token pattern. Disk remains local-only; `.cache` and `.fingerprint-salt` remain untouched.
- Re-run baseline classifications are unchanged: engine missing-route diagnostic; app typed-destination assertions; CLI help/source-origin assertions plus Redis-refused logs; agentcontract connector-hash assertion; and lint findings outside the changed remediation paths. None were modified or reclassified as green.

- Final architecture remediation: operation-specific Ashby bodies/cursors/success envelopes and repeated-cursor refusal; PrestaShop server filter/sort plus exact combined provider windows; CloudTrail StartTime and non-success capped continuations; non-following password token redirects; calendar-safe date window budgets; and strict offset-count load validation all have focused local proofs.
- Atlas and publication: declared CloudTrail SigV4, response envelopes, offset-count, literal owners, proof membership, repaired nativeset owner path, semantic docs parity, fixture-free Ashby metadata, manuals, skills, catalog, and complete website data all validate. Source and generated-data secret scans are clean.
- Final architecture gates: lock checks, full definition validation, full `cmd/connectorgen`/`defs`/`bundleregistry`/`commandrunner`, vet, build, agentcontract, diff, docs, preflight, boundary, canon, release, and website scripts pass. The inherited engine/app/CLI/agentcontract-hash/lint baseline failures remain unchanged and recorded.

- Terminal B3–B6 correction validation: Canny’s generic production proof and deterministic render check passed; the generic paginator proof passed for authoritative full-page `hasMore=false` termination and exact cursor/offset/URL continuation resumes; strict source-lock duplicate/trailing-document and no-write-on-rejection proofs passed; the 28-family source-parent matrix and semantic docs proof passed. Complete generator, definitions, bundle-registry, commandrunner, and engine suites passed. The pre-existing app typed-destination and CLI polling/help/source-origin plus local Redis failures reoccurred unchanged and are not B3–B6 regressions. `go build ./cmd/pm`, GitHub/GitLab/Asana/Canny render checks, 553-definition validation, docs validation, `git diff --check`, and the full changed-artifact secret scan passed. No release automation was run.

## CP06 explicit construction verification

- Preflight base: local `HEAD`, remote parent, and PR #4294 API head are `cb33f2ef2584c7307f3c88b869c474796698b0fb`; API base is `main`; frozen checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7` is reachable. CodeGraph is absent; Go LSP is unavailable; built-in fallback searches, scoped `go list -deps -test`, and selector discovery were used instead.
- Lifecycle/skill evidence: `scripts/gsd doctor` resolved the adapter except for its recorded missing `issue-122-rebootstrap.md`; all lifecycle sources resolved, `go run ./cmd/agentcontractgen check` passed, and the CP06 Go/connector/review skill route is recorded in the TDD ledger. The named phase artifacts remain the inline/manual-GSD fallback.
- Preflight controls actually listed: factory controls `TestNewExecutorFactoriesRejectsDuplicateIDsBeforeConstruction|TestExecutorFactoriesRejectUnknownOrEmptySelectionBeforeConstruction`; registry/authority/compatibility controls; App controls `TestDefaultRegistryContainsOnlyBuiltins|TestOpenConstructsTheExplicitProductionRegistry`; and CLI `TestSentrySeerModelsCommandStopsBeforeProviderIOWithoutCredential`. `TestDefaultRegistryDoesNotUseProcessGlobalBuilder` selected zero tests and is explicitly retired as a verification selector.
- Reconciled READY contract: P1 generated index/store bridge must prove zero-decode list/lookup, one exact digest decode, bounded held identity, cancellation, and mismatch refusal; P2 must prove only the sealed executor/extension identities; P3/P5 must prove App/direct CLI identity plus zero filesystem/vault/approval/auth/request/constructor work on invalid selection; P4 must prove all 49 explicit hooks and disjoint 3+4 native/compatibility inventories. No CP06 Green result is recorded yet.
- Required focused GREEN commands after the individual RED slices: exact `go test -list` before each selector; `go test -count=1 -timeout 20m -race` for manifest index/store/factory construction; exact App/CLI counter selectors; `go test -count=1 -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$'`; `go vet` for changed packages; `go build ./cmd/pm`; generator byte-stability/check; `lock-render --check` for GitHub/GitLab/Asana; `connectorgen validate internal/connectors/defs`; Atlas JSON/schema/proof checks; and CLI implemented/missing-credential, partial, and unknown command witnesses.
- Baseline-only broad results observed before the block: full `internal/app` fails the typed-destination approval/I-O expectations named in `TestFoundationRollupPreservesMultiActionReverseETLComposition`, `TestPersistedConnectionSelectsDeclarativeTypedDestinationAction`, `TestDeclarativeTypedDestinationPersistsProviderResultsAfterPostSuccessLocalFailure`, `TestDeclarativeTypedDestinationReopensRepeatedFullAppendWorksetAfterLocalReceiptFailure`, and `TestDeclarativeTypedDestinationTombstoneAppliesOnlyDeclaredDeleteAndReadsBackAbsence`. Full `internal/cli` repeats `TestPollingHelpDistinguishesStaticDeclarationsFromDynamicRuntimeEligibility`, `TestSourceBoundOriginRejectsBeforeAppOrCredential`, `TestSourceBoundOriginRejectsPersistedCredentialConfigBeforeVault`, and `TestETLHelpListsAllSyncModes`, plus expected local Redis refusal logs. These are not CP06 green evidence and receive no repair in this blocked preflight.

## CP06 retrofitted impact verification — 2026-09-03

- Parent identity was re-read locally and through `gh-axi api repos/polymetrics-ai/cli/pulls/4294`: base `main`, head `fm/cli-top100-declaration-batch-r1`, head SHA `cb33f2ef2584c7307f3c88b869c474796698b0fb`, open. The frozen checkpoint remains an ancestor.
- Read-only/current-state checks: `go test -count=1 -timeout 20m -race ./internal/connectors/manifestindex ./internal/connectors/manifeststore` passed; `go test -count=1 -timeout 20m ./internal/connectors/hooks/...` reported 49 passing packages; `go test -list ... ./internal/connectors/bundleregistry ./internal/connectors/native/nativeset` failed because `registry_test.go` refers to deleted `loadDefinitions`. This is executable RED evidence, not a waived baseline.

| Impact lens | Verification result | Required proof before GREEN |
| --- | --- | --- |
| Architecture/data flow | `BuildRegistry` visibly loops all entries and calls `Acquire`. | Decode counter: list `0`, named lookup/execution `1`. |
| Affected callers | CLI metadata and App/identity have separate production construction calls. | One construction identity across CLI/App, including no duplicate full build. |
| Interfaces/configuration | Entry/handle lack metadata and generation/digest identity. | Typed mismatch/unknown/duplicate refusal before constructors. |
| Generated/docs surfaces | Generator and Atlas do not describe the current lazy-selection contract. | Check-mode byte equality plus Atlas/docs/skills parity proof. |
| Compatibility | 3 database + 4 compatibility IDs exist; stale registry test blocks package. | Exact disjoint inventory and selected-ID factory proof. |
| Security | Only pre-`os.Stat` closure ordering is tested. | Zero filesystem/vault/approval/auth/request/constructor counters on invalid selection. |
| Concurrency/bounds | Production `BundleStore` lacks cancellation/retry/identity coverage. | Race-safe shared-cancel, abandoned-retry, byte/count, digest/generation tests. |
| CLI/App reachability | Direct command injection improved; list/help/docs remain eager. | Implemented, partial, unknown, list, and named-entry behavioral witnesses. |
| Provider semantics | No execution JSON/source-lock diff was found. | Preserve engine-only declared request behavior; no provider activity. |
| Tests/evidence | Only partial focused suites are green; registry package does not compile. | Exact selector counts and all P1–P5 focused tests must pass. |

| Prerequisite | Current disposition |
| --- | --- |
| CP06-P1 production bridge | `blocked`: eager full decode and no verified held identity. |
| CP06-P2 selection ownership | `blocked`: no full row-to-factory/extension validation proof. |
| CP06-P3 construction root | `blocked`: duplicate/eager App and CLI paths remain. |
| CP06-P4 hook/compatibility fan-in | `blocked`: stale registry/Atlas proof references remain. |
| CP06-P5 before-I/O ordering | `blocked`: invalid-selection zero-I/O proof is absent. |

- Overall: `BLOCKED_PRE_IMPLEMENTATION`. Firstmate `impact-ready` is required before any CP06 production edit, test addition, generator write, commit, review, or push.

- CP06-P1 RED: `TestConstructionBuildsLazyRegistryWithoutDecodingList` was listed exactly once, then failed with `BuildRegistry() decoded 2 bundles ... want 0`. This is the required list-path behavioral failure; it replaces the obsolete deleted-`loadDefinitions` compile failure as the first implementation oracle.

## CP06 rate-limit preservation verification

- Required static identity: GitHub's declared rate execution JSON and PostgreSQL's explicit `not_applicable` JSON remain unchanged; no other connector receives a policy by inference.
- Required Harness A: mutate only a temporary `rate_limits.json` fixture and prove generated digest, byte charge, and closed-set output change deterministically.
- Required Harness B: construct generated GitHub through the production index/store/factory path, use a local fake 429/reset/admission witness, prove same-scope process-budget sharing and distinct-scope admission, and prove no provider endpoint/credential is used.
- Required Harness C: generation/digest mismatch refuses before factory execution; held handles retain the exact selected identity through construction.
- Required focused regressions: manifestindex/manifeststore/bundleregistry race tests; existing engine rate request-path tests; existing parking/resume tests; built `pm connectors inspect github --json` plus text output. No rate policy, source lock, provider state, release artifact, or live certification check may change.

## CP06 implemented-control results

- P1/P2/P4: `go test -count=1 -timeout 20m -race ./internal/connectors/bundleregistry ./internal/connectors/manifestindex ./internal/connectors/manifeststore ./internal/connectors/hooks/hookset ./internal/connectors/native/nativeset` passed.
- P3/P5: focused App controls for zero metadata decode, lazy transport composition, invalid-selection-before-stat, and zero phase measurement passed; `go test -count=1 -timeout 20m ./internal/cli -run '^TestDynamicConnectorCommandsUseLazyMetadata$'` passed.
- Rate Harness A: `go test -count=1 -timeout 20m ./cmd/connectorgen` passed, including `TestGeneratedManifestIndexDigestIncludesRateLimits` and Atlas selector resolution.
- Rate Harness B: `go test -count=1 -timeout 20m ./internal/connectors/bundleregistry -run '^TestLazyConstructionSharesGitHubRateAdmission$'` passed with local fake 429/reset only.
- Rate Harness C: `go test -count=1 -timeout 20m -race ./internal/connectors/manifeststore` passed, including generation/digest mismatch, generation separation, cancellation/retry, and held-identity controls.
- Existing rate regressions passed: `TestGitHubDeclaredRateLimits`, `TestGitHubRateLimitAdmissionPrecedesProviderAndIsolatesScope`, `TestParkedFullAppendRateResumePreservesCheckpointAndBatchSize`, and `TestParkedFullAppendRateResumeRearmsLatestCheckpoint`.
- Scoped broad App result: only the previously recorded typed-destination approval/I-O baseline failures remain (`TestFoundationRollupPreservesMultiActionReverseETLComposition`, `TestPersistedConnectionSelectsDeclarativeTypedDestinationAction`, `TestDeclarativeTypedDestinationPersistsProviderResultsAfterPostSuccessLocalFailure`, `TestDeclarativeTypedDestinationReopensRepeatedFullAppendWorksetAfterLocalReceiptFailure`, and `TestDeclarativeTypedDestinationTombstoneAppliesOnlyDeclaredDeleteAndReadsBackAbsence`). No lazy-construction/transport failure remains.

- Final scoped passes: `go vet` for CP06 packages; `go build ./cmd/pm`; `make smoke-no-build`; full `cmd/connectorgen`, `defs`, `engine`, `bundleregistry`, `manifestindex`, `manifeststore`, `nativeset`, and focused CLI/App suites passed.
- Generated/authoring passes: `connectorgen gen`; GitHub/GitLab/Asana `lock-render --check`; 553-definition validation; Atlas JSON/unique-ID validation; Atlas selector test; connector canon; implemented-command preflight; and whole-tree connector boundary returned zero findings.
- Surface passes: `pm help`, `pm connectors`, `pm github --help`, and GitHub text/JSON inspection exited successfully without credentials or provider I/O; docs validation passed.
- Recorded inherited failures, not CP06 green evidence: full `internal/app` retains the named typed-destination approval/I-O baseline; full `internal/cli` retains polling/help/source-origin baselines with local Redis refusal logs; full boundary-package test retains `TestScanFailsClosedWhenConnectorMetadataCannotLoad/invalid_cli_surface`; `make lint` retains unrelated existing engine and obsolete connectorgen/boundary helper findings after CP06's generator cleanup.

## CP07 GitHub reference verification

- Starting parent: `843a32de5f927b1235cc00883fa0c5e0f5ea8c5b` is both local and remote parent after PR #4294 API read-back.
- Scope: production-registry characterization only. No source lock/rendered GitHub execution JSON/provider request/credential mutation is permitted.
- Required witness: generated index → `NewRegistry` → GitHub `*engine.Connector`, exact executor/extension selection, process-local rate coordination, and no native/compatibility fallback.

- `TestGitHubReferenceUsesManifestSelectedAPIEngine` was listed exactly once and passed against the production registry. No CP07 source/definition/provider mutation was needed.

## CP08 PostgreSQL reference verification

- Starting parent: `c267f6ccb6988c6d0132f264e963c6701b8134f1` is local, remote, and PR #4294 API head.
- Scope: production-registry characterization only. PostgreSQL execution JSON/source lock/rate declaration and database behavior remain unchanged.
- Required witness: exact native executor/no extension, protected native adapter type, no rate coordination claim, and no API/compatibility fallback.

- `TestPostgresReferenceUsesManifestSelectedNativeDatabase` was listed exactly once and passed against the production registry. No CP08 definition, protocol, rate, credential, or provider mutation was needed.

## A1 correction-wave verification

- A1-01: exact table/fuzz matrix over seven modes and every closed axis; C3 real-axis classification, C4 gap, canonical digest serialization, and no-I/O dependency boundary.
- A1-02: generator/loader identity parity; same-name loaded mismatch; exact byte charge; factory/rate resolver non-invocation on mismatch; race/cancellation/lease regression.
- A1-03: real `cli.Run` root-router normal, reverse, and help paths use exactly one preflighted registry identity/decode; injected opener behavior remains explicit and separate.
- Final wave proof repeats rate, GitHub/PostgreSQL, hooks/3+4, credentials, render/validate/preflight, docs/help, smoke, boundary, vet/build, frozen self-review, and one final A1 exact-SHA re-review.

## A1 consolidated correction results

- A1-01: `go test -count=1 -timeout 20m ./internal/synccontract ./internal/syncplan ./internal/syncrun` passed; `go test -run '^$' -fuzz '^FuzzResolveNeverExecutesContradictoryModeAxes$' -fuzztime=5s -timeout 20m ./internal/syncplan` passed after 1.4M+ executions; `go list -deps ./internal/synccontract` contains no project dependency beyond the leaf package.
- A1-02: `go test -count=1 -timeout 20m -race ./internal/connectors/manifestidentity ./internal/connectors/manifeststore ./internal/connectors/manifestindex ./internal/connectors/bundleregistry` passed, including actual `engine.Load` identity parity, same-name loaded generation/digest/charge mismatch refusal, no factory call, cancellation, bounded cache, and generation-lease checks.
- A1-03: focused normal router/normal/help/reverse approval ownership, Cobra router, connector preflight, and tracked skill-generation tests pass; race variants pass. Full CLI remains red only on its recorded polling/help/source-origin baseline plus local Redis connection-refused logs, with no new registry-ownership failure.
- Retained rate/reference checks passed: generated rate-file identity test; GitHub same-scope fake 429/reset admission; GitHub declared rate contract; parked resume/rearm; GitHub API-engine and PostgreSQL native reference witnesses.
- Generated/structural checks passed: `connectorgen gen`; GitHub/GitLab/Asana renders; 553-definition validation; Atlas selector and JSON/unique-ID checks; commandrunner implemented preflight; whole-tree boundary (zero findings); agent contract; scoped vet; build; help/connectors/GitHub inspect smoke; docs validation; local warehouse/reverse smoke.
- Full App remains red only on the recorded typed-destination approval/I-O baseline. No source lock, rendered definition JSON, rate declaration, credential, provider, release, or certification residue changed.
- Current scoped gates passed: `make tidy-check`, `make docs-check`, `make smoke-no-build`, `make agent-contract-check`, `make connectorgen-validate`, `make connectorgen-vnext-locks`, `make connector-runtime-preflight`, `make connector-boundary`, `make connector-canon-check`, `make release-workflow-check`, `make vet`, `go build ./cmd/pm`, and `git diff --check`. The boundary report is clean with its existing documentation exceptions only.
- Current broad baselines: `go test -count=1 -timeout 20m ./internal/app` failed only the five named typed-destination approval/I-O controls recorded above; `go test -count=1 -timeout 20m ./internal/cli` failed only `TestPollingHelpDistinguishesStaticDeclarationsFromDynamicRuntimeEligibility`, `TestSourceBoundOriginRejectsBeforeAppOrCredential`, `TestSourceBoundOriginRejectsPersistedCredentialConfigBeforeVault`, and `TestETLHelpListsAllSyncModes`, with expected local Redis refused-connection logs. Neither package failure is resolved or called green by this wave.
- Current lint baseline: after removing two stale now-unused helpers from the changed engine loader, `make lint` reports 20 diagnostics only in unmodified `polling_watermark_test.go`, `continuation_cap_test.go`, `read.go`, boundary lexicon, and obsolete connectorgen helper files. No A1-changed path appears in the output; the unrelated lint debt remains outside this correction scope.

## A1 entry-capacity correction verification plan

- Baseline: Firstmate's independent review BLOCK is accepted at `701a0b45175f308400c938322fd1634a28efdaef`; no attempt is made to re-run or reinterpret that report.
- RED: the two-identity barrier selector must show that `Limits{Entries: 1, Bytes: 2}` admits a second distinct in-flight load before the correction, rather than treating an exit status as evidence.
- GREEN: the same selector must assert each locked snapshot satisfies `len(cache) + reservedEntries <= 1`, cancellation cannot admit B while A's loader is live, and post-completion B/A retries retain at most one entry.
- Affected checks: manifeststore race suite; manifestindex/bundleregistry identity and construction controls; scoped vet; Atlas JSON/owner/proof validation; agent-contract check; `git diff --check`.
- Exclusions: no provider endpoint, credential, rate daemon/declaration, source lock/render, CLI docs/help, release workflow, `.cache`, or certification residue check is part of this correction.
- Observed RED: `go test -count=1 -timeout 20m ./internal/connectors/manifeststore -run '^TestBundleStoreReservesEntryCapacityAcrossDistinctFlights$'` failed deterministically while the alpha barrier was held: B acquired successfully and the locked one-byte fixture snapshot reported retained plus reserved entries `2`. The failure is the review-reported entry-count gap, not an exit-status proxy.

## A1 entry-capacity correction results

- GREEN regression: `go test -count=1 -timeout 20m ./internal/connectors/manifeststore -run '^TestBundleStoreReservesEntryCapacityAcrossDistinctFlights$'` passed. A distinct B load is refused while A reserves the only entry; it remains refused after A's caller cancels while A's loader is barrier-held; after A terminally exits, B then A retry without a retained-plus-reserved count above one.
- Concurrency: `go test -count=1 -timeout 20m -race ./internal/connectors/manifeststore` passed.
- Affected construction identity: `go test -count=1 -timeout 20m ./internal/connectors/manifestidentity ./internal/connectors/manifestindex ./internal/connectors/bundleregistry` passed.
- Static and Atlas: `go vet ./internal/connectors/manifeststore ./internal/connectors/bundleregistry` was clean. `jq empty` and unique-ID validation of the Atlas catalog passed, as did `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestFoundationAtlasSelectorsResolve$'`.
- Delivery mechanics: `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and final `git diff --check` passed. The generated `verify-work` and `code-review` prompts were executed inline under the documented unavailable-runner fallback; the local review is recorded in `REVIEW-CONVERGENCE.md`. No provider, credential, rate-declaration, source-lock/render, parent-push, or CP09 activity occurred.
- Independent exact-SHA review PASS: Firstmate instruction `120.msg` accepted `c761e7e6f2d042c7560ab0c520dc9aa182110e6e` with zero blockers. A1-04 is eligible only for normal non-force publication to `fm/cli-top100-declaration-batch-r1`; CP09 begins after that parent update and `main` remains the PR merge base.

## CP09 strict source parsing and canonical graph verification

- Scope: `cmd/connectorgen` source-lock authoring only. The sole route remains schema-4 source lock → canonical graph → deterministic execution JSON → existing execution-only runtime. No connector source lock, rendered artifact, provider, credential, runtime executor, release path, `.cache`, or certification residue changed.
- Rejection: `TestRunLockRenderRejectsInvalidCanonicalGraphBeforeWriting` passed for unknown execution fields, wrong schema roles, unsupported encoder values, normalized aliases, missing record bindings, and non-object source identity. Each failure returns exit 1 before the sentinel output file changes.
- Admission/determinism: `TestVNextCanonicalGraphAllowsDirectOperationSchemaRoles` passed without adding CP10 semantic source joins. `TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering` passed with equal execution bytes and immutable graph digests; source operation/provider/CLI facts remain authoring-only and do not affect execution identity. `TestEverySourceLockBuildsCanonicalGraph` passed across every committed schema-4 lock.
- Package and authoring smoke: `go test -count=1 -timeout 20m ./cmd/connectorgen` passed. `lock-render asana --check`, `lock-render github --check`, and `lock-render gitlab --check` each observed a current bundle; full `connectorgen validate internal/connectors/defs` reported `553 connector(s) checked, 0 findings`.
- Atlas/docs/tooling: catalog revision 30 passed schema and unique-ID checks, `TestFoundationAtlasSelectorsResolve`, and `make docs-check`. `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and final `git diff --check` passed.
- Lifecycle/review: generated `execute-phase`, `verify-work`, and `code-review` prompts were fulfilled inline because the configured GSD worker/reviewer runtime is unavailable; this documented fallback does not replace a future Firstmate exact-SHA review if one is requested. CLI help/manual/website parity is not applicable: CP09 does not add or alter a CLI command, flag, output, help topic, or connector surface.
- Publication: after the empty pending-inbox check, `d11277378abe556323226e3f6998ce3caf6033dc` was committed and normally pushed to the declared parent branch. GitHub API read-back for draft PR #4294 confirms `base=main`, `head=fm/cli-top100-declaration-batch-r1`, and that exact head SHA. This does not publish a rendered connector bundle and does not start CP10.

## CP10 semantic source-execution admission verification

- Scope: CP10 changes only the authoring-side `cmd/connectorgen` in-memory admission path, its tests, canonical source-lock/Atlas documentation, and phase evidence. It does not modify a connector source lock, rendered execution bundle, generated global manifest/index, credential, provider route, runtime network call, filesystem staging/activation, `.cache`, or certification residue.
- RED: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextSemanticAdmissionRejectsSwappedGraphQLOperationBeforeWriting$'` initially exited `1`: an internally valid command could cross from `source:widgets.first` to a separate same-route `/graphql` operation and `runLockRender` replaced a sentinel. The test is now green and asserts the immutable source ID, exact command JSON path, and unchanged sentinel.
- GREEN semantic controls: `TestVNextSemanticAdmissionStagesRateIdentityAndManifestIndex`, `TestVNextSemanticAdmissionRejectsMalformedRateBeforeWriting`, `TestVNextSemanticAdmissionRejectsCrossConnectorManifestAndResolvesSuppliedSync`, and `TestVNextSemanticAdmissionRunsResolverAndPreflightBeforeWriting` pass. They prove deterministic rendered bytes/provenance/index identity; rate add/change/remove identity closure; malformed-rate no-write rejection; selected executor and generation-digest binding for an explicitly supplied complete sync plan through `syncplan.Resolve`; and real resolver/preflight rejection of bad route/flag bindings before output replacement.
- Regression and authoring checks: `go test -count=1 -timeout 20m ./cmd/connectorgen` passed. `go run ./cmd/connectorgen lock-render asana --check`, GitHub, and GitLab each reported their bundle current; `go run ./cmd/connectorgen validate internal/connectors/defs` checked `553 connector(s), 0 findings`; `TestFoundationAtlasSelectorsResolve` passed.
- Static/docs checks: `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, `make docs-check`, `go run ./cmd/agentcontractgen check`, and final `git diff --check` passed.
- Scoped changed-file secret scan found no private-key, GitHub-token, OpenAI-token, or AWS-key pattern. `.cache/` and `internal/connectors/certifications/` remain untracked, unread, unstaged, and unmodified local residue.
- CLI parity: not applicable to a public surface. CP10 adds no command, flag, output, help topic, manual, website contract, or connector surface. The required scoped `go test -count=1 -timeout 20m ./internal/cli` re-run failed only the recorded inherited `TestPollingHelpDistinguishesStaticDeclarationsFromDynamicRuntimeEligibility`, `TestSourceBoundOriginRejectsBeforeAppOrCredential`, `TestSourceBoundOriginRejectsPersistedCredentialConfigBeforeVault`, and `TestETLHelpListsAllSyncModes` controls, with the known local Redis refused-connection logs; no CP10 admission selector failed.
- Lifecycle/review: generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts were resolved. The established phase lacks the required adapter ROADMAP and `scripts/gsd doctor` has the pre-existing missing `issue-122-rebootstrap.md` prompt, so the equivalent TDD, verification, and review procedures were completed inline; no unavailable worker/reviewer role was fabricated. Independent review remains deferred to the coherent phase boundary as Firstmate instruction `122.msg` requires.

## N2 boundary-review CP09/CP10 correction verification

- Scope/no-write: instruction `124.msg` permits only effective request/response schema joins, complete closed hook-bearing manifest staging, and canonical staged provenance from parent `56ec3d9d7dc1d726203b0ef0c03ddec3209b8dde`. No source lock, rendered/global output, provider/credential I/O, certification, filesystem staging/activation, release path, `.cache`, or CP11 behavior changed.
- RED/GREEN: swapped existing request and response schemas initially let `runLockRender` exit `0` and replace a sentinel; after the join, `TestVNextSemanticAdmissionRejectsSwappedEffectiveSchemaBeforeWriting` rejects both at the authored role pointer before replacement. GitHub initially failed manifest binding because staged extension was empty; `TestVNextSemanticAdmissionStagesGitHubProductionHookEntry` now proves exact generated entry/index equality, `hook/github.v1`, and the closed selected hook. Reordered equivalent operations initially changed staged field paths; `TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering` now proves full provenance equality while a malformed reordered lock reports its authored `/operations/0/schema_refs/record`.
- Focused regression: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextSemanticAdmissionRejectsSwappedEffectiveSchemaBeforeWriting|TestVNextSemanticAdmissionStagesGitHubProductionHookEntry|TestVNextCanonicalGraphIgnoresIrrelevantOperationOrdering|TestVNextSemanticAdmissionRejectsUnboundDirectOperationSchemaRoles)$'` → `ok`. The direct-only role control proves no untyped REST response/request role is admitted merely as provenance.
- Package/affected suites: final `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (51.157s); `go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/manifestindex ./internal/connectors/native/nativeset ./internal/syncplan` → `ok`.
- Authoring/docs/static checks: `lock-render --check` passed for Asana, GitHub, and GitLab; `connectorgen validate internal/connectors/defs` reported `553 connector(s) checked, 0 findings`; catalog revision `32` passed JSON/unique-ID validation and `TestFoundationAtlasSelectorsResolve`; `make docs-check`, `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, and `go run ./cmd/agentcontractgen check` passed. A scoped changed-file secret scan found no private-key, GitHub-token, OpenAI-token, or AWS-key pattern.
- Lifecycle/review: generated `execute-phase`, `verify-work`, and `code-review` prompts were fulfilled inline under the recorded unavailable Pi-worker/manual-GSD fallback. The read-only self-review is in `REVIEW-CONVERGENCE.md`; no local blocker remains. This does not satisfy or replace the fresh Firstmate exact-SHA review required before CP11.

- Publication: correction commit `b4dd03e8c2e113f1e791f5844db253135cdb5c9e` was normally pushed as `HEAD` to `origin/fm/cli-top100-declaration-batch-r1`. GitHub API read-back for draft PR #4294 reports `base=main`, `head=fm/cli-top100-declaration-batch-r1`, `sha=b4dd03e8c2e113f1e791f5844db253135cdb5c9e`, and `state=open`. The final exact-SHA review candidate is recorded in Firstmate state after this evidence checkpoint; CP11 remains prohibited.

## CP11 B1 transactional publication verification plan

- Impact-ready: Firstmate instruction `125.msg` authorizes only #4427/B1 from immutable `36e4d980de0d51d92fe74a68306845643596a6cb`; the SHA-bound ten-row plan map is `READY` before production edits. CP12 is not started.
- Bounded surface: CP11 will exercise only temporary connector publication roots through the existing semantic stage. It will prove a closed staged artifact set, integrity/digest equality, same-filesystem durable activation, journal recovery, old/new reader consistency, writer serialization, and generation-handle-safe pruning. It will not publish real source locks or connector artifacts, call a provider, resolve a credential, alter App/CLI runtime routing, or delete an author-owned file.
- Evidence contract: a RED must expose stale optional output, a mixed old/new observation, journal interruption, writer interleave, or unsafe prune; GREEN must observe an exact old or new complete set after recovery rather than only exit status. Fault tests use explicit cut-point hooks and barriers, never wall-clock sleeps.
- Planned local gates: selector-list proof; focused `cmd/connectorgen` tests and race run; package test; source-lock deterministic in-memory parity; Atlas/docs checks; `go vet ./cmd/connectorgen`; `go build ./cmd/pm`; `go run ./cmd/agentcontractgen check`; `git diff --check`; scoped changed-file secret scan; generated inline/manual `execute-phase`, `verify-work`, and `code-review` evidence; normal non-force push and GitHub PR base/head/SHA read-back.

## 2026-09-04 — CP11 B1 transactional publication verification

- Scope/no-I/O: publication consumes an already admitted CP10 staged value only
  through hermetic temporary roots. It did not materialize a checked-in
  connector/source lock or execution bundle; it made no provider, credential,
  database, App/CLI-runtime, release, `.cache`, or certification access.
- RED/GREEN: the TDD ledger records executable failures for orphan recovery,
  rejected active validation, repeat-set journaling, ambiguous generation
  identity, unsafe artifact paths, unowned pruning, and a symlinked
  `generations/` root that deleted an author-owned target. Each is green under
  an observed temporary-root control; the last requires a nonsymlink root
  before recovery/staging/check/open/prune.
- State-machine/concurrency proof:
  `go test -race -count=1 -timeout 20m ./cmd/connectorgen -run
  '^(TestVNextGenerationPublisher.*|TestVNextPublicationGenerationIDDelimitsArtifactNamesAndBytes|TestRunLockRenderPublishesOnlyClosedGeneration)$'`
  → `ok` (7.547s). It covers all declared durable fault cut points, exact
  old/new recovery, same-parent staging, stale optional artifact removal,
  serialized writers, no mixed bundle/index view, and held-reader prune
  deferral.
- Package and authoring proof:
  `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (52.811s);
  `make connectorgen-vnext-locks` → `ok` (reference in-memory parity plus
  temporary closed publisher); `bash scripts/tests/connector-canon.sh` →
  `connector canon check: ok`; and
  `go test -count=1 -timeout 20m ./cmd/connectorgen -run
  '^TestFoundationAtlasSelectorsResolve$'` → `ok`. `jq empty
  docs/connector-canon/foundations/catalog.json` also passed after the
  revision-33 owner/proof update.
- Affected runtime packages:
  `go test -count=1 -timeout 20m ./internal/connectors/defs
  ./internal/connectors/engine ./internal/connectors/manifestindex` → all
  `ok`; those packages are read-only consumers and were not modified.
- Static/documentation contract:
  `go vet ./cmd/connectorgen`, `go build ./cmd/pm`, `make docs-check`,
  `go run ./cmd/agentcontractgen check`, and
  `go run ./cmd/connectorgen validate internal/connectors/defs` all passed;
  definition validation reported `553 connector(s) checked, 0 findings`.
  `go run ./cmd/connectorgen --help` retained
  `lock-render <connector> [--defs <dir>] [--check]`. Website metadata was
  regenerated with `node website/scripts/gen-docs-data.mjs`; its two-node test
  suite passed.
- CLI/manual/website parity: no PM user-facing command, flag, output, help
  topic, or runtime behavior changed, so PM CLI manual parity is not
  applicable. The unchanged `connectorgen` help was exercised; canonical
  authoring docs, Guide, architecture/migration references, website install
  and connector-delivery content, and generated website metadata were updated
  for the closed-generation/check-only contract. No `docs/cli` connectorgen
  manual exists.
- Required unrelated integration baseline:
  `go test -count=1 -timeout 20m ./internal/cli` failed only in its recorded
  polling-help, source-bound-origin, ETL-help, and local Redis
  connection-refused baselines. CP11 changes no `internal/cli` source; the
  exact observed failures are retained in the command transcript and local
  review.
- Lifecycle/review: all resolved GSD command sources and generated
  discuss/plan/execute/verify/review prompts were fulfilled inline under the
  recorded unavailable-compatible-worker/roadmap fallback. The read-only
  self-review is recorded in `REVIEW-CONVERGENCE.md`; it found and corrected
  the symlinked-generation-root deletion before the final tests. This does not
  replace the mandatory Firstmate exact-SHA independent review after push.
- Patch hygiene: `git diff --check` passed. A scoped scan of every changed
  source, test, script, canon, Atlas, documentation, website-generated, and
  phase-evidence file found no private-key, GitHub-token, OpenAI-token, or
  AWS-key signature. The scan deliberately excluded untouched untracked
  `.cache/` and certification residue.

## CP11 exact-review correction verification plan

- Exact authority: instruction `126.msg` permits five corrections only from
  `c4f0bc3728dda318ea3d01f78de7aa299b6135cb`; CP12 remains prohibited.
- Observable temporary-root proof: test root/lock symlink refusal with
  external sentinels; unknown-stage preservation across every public operation;
  renamed/self-consistent generation, unexpected-directory, and nonempty-lease
  rejection; oversized/duplicate control-document refusal; and context
  cancellation with unchanged `CURRENT`/journal plus successful retry.
- Required execution: focused RED/GREEN selectors after each behavior;
  race publisher suite; full `./cmd/connectorgen`; affected engine/index/
  commandrunner packages; `make connectorgen-vnext-locks`; canon/Atlas JSON
  and selector checks; vet; `go build -o /dev/null ./cmd/pm`; docs validation;
  agent-contract and definitions validation; diff/secret checks; and the
  documented `internal/cli` baseline run.
- Parity: no PM command/flag/help/manual/website behavior is intended. Exercise
  unchanged `connectorgen --help`; update only canonical authoring/Atlas claims
  if the corrected filesystem, strict-control, or cancellation contract changes.
- Review/delivery: append the inline manual-GSD result and read-only
  self-review to phase evidence; normally push one checkpoint; verify PR #4294
  API base=`main`, expected head ref, draft/open state, and exact SHA; then
  pause unchanged for Firstmate's new exact-SHA CP11 review.

## 2026-09-04 — CP11 exact-review correction verification

- Authority/scope: instruction `126.msg` permits only the five publication
  corrections from `c4f0bc3728dda318ea3d01f78de7aa299b6135cb`. All observable
  proofs use hermetic temporary publication roots. No checked-in connector or
  source lock was materialized; no provider, credential, database,
  certification, runtime-route, cross-connector, or CP12 work occurred.
- RED/GREEN record: `TDD-LEDGER.md` records the observed root/stage/tree/control
  REDs, the blocking-Flock manual RED fallback, and exact GREEN dispositions.
  `REVIEW-CONVERGENCE.md` records the required inline/manual local review and
  its non-substitution for Firstmate's pending independent review.
- Correction selector GREEN:
  `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSymlinkedConnectorRootWithoutTouchingTarget|TestVNextGenerationPublisherRefusesSymlinkedLockWithoutTouchingTarget|TestRunLockRenderRefusesSymlinkedConnectorRootWithoutTouchingTarget|TestVNextGenerationPublisherPreservesUnownedStageAcrossMutations|TestVNextGenerationPublisherRecoversDurablyOwnedStage|TestVNextGenerationPublisherRejectsSelfConsistentRenamedGeneration|TestVNextGenerationPublisherRejectsUnexpectedDirectoryAndNonemptyLease|TestVNextGenerationPublisherRejectsDuplicatePublicationControlMembers|TestVNextGenerationPublisherRejectsOversizedIntegrityControl|TestVNextGenerationPublisherContextCancelsContendedLockWithoutStateChange|TestRunLockRenderContextCancelsContendedPublicationAndRetries)$'`
  → `ok` (3.902s).
- Race/concurrency GREEN:
  `go test -race -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisher.*|TestVNextPublicationGenerationIDDelimitsArtifactNamesAndBytes|TestRunLockRenderPublishesOnlyClosedGeneration|TestRunLockRenderContextCancelsContendedPublicationAndRetries)$'`
  → `ok` (10.795s).
- Changed-package GREEN:
  `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (56.628s);
  `go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/manifestindex ./internal/connectors/commandrunner`
  → `ok`.
- Authoring/canon GREEN: `make connectorgen-vnext-locks` → `ok`;
  `go run ./cmd/connectorgen validate internal/connectors/defs` →
  `553 connector(s) checked, 0 findings`; `jq empty
  docs/connector-canon/foundations/catalog.json`,
  `bash scripts/tests/connector-canon.sh`, and
  `go test -count=1 -timeout 20m ./cmd/connectorgen -run
  '^TestFoundationAtlasSelectorsResolve$'` → `ok`.
- Static/documentation GREEN: `go vet ./cmd/connectorgen`,
  `go build -o /dev/null ./cmd/pm`, `go run ./cmd/agentcontractgen check`,
  `make docs-check`, `go run ./cmd/connectorgen --help`, and
  `git diff --check` passed. The help output preserves the existing
  `lock-render <connector> [--defs <dir>] [--check]` syntax; PM/manual/website
  parity is not applicable because no PM surface changed.
- Hygiene: a scoped changed-path scan found no private-key, GitHub-token,
  OpenAI-token, or AWS-key marker. `make docs-check`'s transient `pm` binary
  was removed; `.cache/` and `internal/connectors/certifications/` remain
  untracked, unread, and untouched.
- Required baseline (not green, unchanged scope):
  `go test -count=1 -timeout 20m ./internal/cli` failed after 234.747s with
  existing `TestPollingHelpDistinguishesStaticDeclarationsFromDynamicRuntimeEligibility`,
  `TestSourceBoundOriginRejectsBeforeAppOrCredential`,
  `TestSourceBoundOriginRejectsPersistedCredentialConfigBeforeVault`, and
  `TestETLHelpListsAllSyncModes` failures, plus local Redis connection-refused
  logs for `127.0.0.1:1` and `127.0.0.1:2`. CP11 changes no `internal/cli`
  source, so this is recorded baseline evidence rather than repaired or called
  green.
- Delivery gate: correction commit/push, PR #4294 API base/head/SHA read-back,
  Firstmate-state update, and a new independent exact-SHA review remain pending;
  CP12 remains prohibited.

## CP11 correction-review repair verification plan (instruction 127)

- Scope/no-I/O: operate only on temporary publication roots and the shared
  connector-boundary/Atlas authoring code. No checked-in connector generation,
  source lock, rendered artifact, provider, credential, database,
  certification, runtime route, or CP12 behavior is exercised.
- RED/GREEN: record each observed descriptor-replacement, component-boundary,
  acquisition-cancellation, journal-order, and Atlas-map failure before its
  minimal repair; no former test is relabeled as a RED result.
- Focused proof: run the new destructive-path sentinel test, scanner
  positive/negative fixture, cancellation barrier, both journal-cut recovery
  tests, and Atlas mutation/resolution validator. Run their package suites and
  the publisher race selector after refactor.
- Affected static gates: `go run ./cmd/connectorgen boundary . --json` must
  clear the reviewed-range false positive; run `go vet` for
  `./cmd/connectorgen` and `./internal/connectors/boundary`, PM build,
  connector-canon/Atlas schema checks, Foundation Atlas selectors, definition
  validation, vNext lock checks, docs/agent-contract checks, diff/secret scan.
- Parity: `connectorgen lock-render` flags/output remain unchanged. Exercise
  `go run ./cmd/connectorgen --help`; PM CLI/manual/website changes are not
  applicable unless a user-visible surface changes unexpectedly.
- Delivery: update review evidence with an inline/manual GSD disposition,
  normal non-force push only, API-read PR #4294 base/head/SHA, record Firstmate
  state, archive the instruction only after all work, and pause for fresh
  exact-SHA CP11 review.

## CP11 correction-review repair verification — instructions 127–129

- Focused publication/lock-render/Atlas Green: source-descriptor replacement,
  source-mutation refusal before generation creation, prepared-journal cuts,
  post-acquisition cancellation, and Atlas selector mapping tests passed.
- Changed package Green:
  `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (61.431s);
  `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (325.701s).
- Boundary Green: the focused Cal.com component test passed, and
  `go run ./cmd/connectorgen boundary . --json` returned `outcome: clean`,
  `connectors_loaded: 553`, `checked_files: 280`, zero findings, and only the
  six pre-existing declared exceptions.
- Authoring/documentation Green: `make connectorgen-vnext-locks` → `ok`
  (30.213s); definition validation reported 553 connectors and zero findings;
  `go vet ./cmd/connectorgen ./internal/connectors/boundary`, PM build,
  agent-contract validation, `make docs-check`, Atlas JSON/schema checks,
  `go mod tidy -diff`, help rendering, scoped secret scans, and `git diff
  --check` passed. `make docs-check`'s generated `pm` binary was removed.
- Module closure: direct descriptor-relative Unix calls make the already-pinned
  `golang.org/x/sys v0.47.0` a direct dependency. `go mod tidy` also corrected
  the existing `golang.org/x/text` checksum/version closure to its selected
  `v0.38.0`, which was necessary for the race suite to resolve.
- Required unaffected-package baselines remain non-green and were not repaired:
  `go test -count=1 -timeout 20m ./internal/connectors/boundary` fails only
  `TestScanFailsClosedWhenConnectorMetadataCannotLoad/invalid_cli_surface`
  because the fixture expects `cli_surface.json` decoding although
  `loadLexicon` reads metadata only; the actual scanner is clean. The required
  `internal/cli` package fails only its recorded polling-help, source-bound
  origin, ETL-help, and local Redis-refusal baselines. No changed CP11 path
  reaches either failure.
- Manual GSD fallback: the adapter sources and all five generated prompts were
  resolved; its known missing `issue-122-rebootstrap.md`, absent phase roadmap,
  and unavailable compatible isolated reviewer require inline execution,
  verification, and review. This remains evidence only, not a replacement for
  Firstmate's fresh exact-SHA review.
- Delivery gate: commit, ordinary push to
  `fm/cli-top100-declaration-batch-r1`, PR #4294 API base/head/SHA read-back,
  state update, archive `127.msg`, then pause for Firstmate review. CP12
  remains prohibited.

## CP11 final repair verification plan — instructions 130–131

- F1: separate `CURRENT` and `JOURNAL` temporary-replacement tests must prove
  identity comparison occurs after the fault barrier and before rename. The
  retained prior control bytes, moved original, and replacement are observable
  fixtures; a refusal without those assertions is insufficient.
- F2: run stage, generation/lease, rollback, and control-cleanup replacement
  tests through Recover, Prune, and Publish rollback. Each asserts the original
  validated object and the replacement survive intact.
- F3: run a post-acquisition two-publisher barrier. While the first holds its
  original inode, the second must not mutate through a replacement lock path;
  after restoring the bound path and releasing the first, retry must pass.
- F4: run a physical staged implemented-command positive witness and a modified
  staged command negative witness that fails specifically at
  `preflight staged command`, then run Atlas selector/mapping validation.
- Gates: focused `cmd/connectorgen` selectors, full and race package suites,
  vNext lock target, source-lock definition validation, documentation/Atlas
  JSON checks, `go vet`, PM build, agent contract check, module tidy/diff
  checks, and the documented actual boundary scanner. The inherited boundary
  fixture and `internal/cli` baselines remain background-only unless a changed
  path proves causal connection.
- Delivery: self-review the exact candidate, ordinary push, PR #4294 API
  base/head/SHA read-back, external status update, archive each instruction,
  then pause unchanged for a new Firstmate exact-SHA review. CP12 remains
  prohibited.

## CP11 final repair verification — F1–F4

- Scope: only `cmd/connectorgen` publication/Atlas proof code, source-lock authoring canon/Atlas documentation, and this phase evidence changed. No provider, credential, database, runtime, source-lock payload, generated execution JSON, PM CLI contract, `.cache/`, or certification-residue path was read or modified.
- F1 Green: `TestVNextGenerationPublisherRefusesReplacedAtomicControlTemporary` passed for both temporary `CURRENT` and `JOURNAL` identity barriers. The failure path retains prior control bytes plus both moved-original and replacement fixtures.
- F2 Green: post-validation stage, generation/lease, and committed-control replacement tests passed; the rollback-specific generation replacement test passed. Recover, Prune, Publish rollback, and control cleanup all refuse without removing either object.
- F3 Green: `TestVNextGenerationPublisherRefusesReplacedLockAfterAcquisition` covers an original held inode, deleted companion anchor, visible replacement lock, second-operation refusal, unchanged `CURRENT`/journal/generation transaction, restored hard-link pair, and serial retry. `TestVNextPublicationOpenLockRefusesExistingLockWithoutAnchor` had an executable RED against the predecessor anchor-bootstrap behavior and passes with the fixed no-reconstruction rule.
- F4 Green: `TestVNextPublicationProofContractRejectsCompoundMapping` had an executable RED against predecessor behavior and passes with exactly-one guarantee enforcement. Atlas selector resolution accepts only the source-lock-vNext record. Physical publication witnesses pass for a valid implemented `widgets get` command and refuse the post-admission invalid staged command specifically at `preflight staged command "widgets get"` before selection state exists.
- Final focused/corpus gates:
  - `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestFoundationAtlasSelectorsResolve|TestVNextPublicationProofContractRejectsCompoundMapping|TestVNextGenerationPublisherPhysicallyPreflightsImplementedStagedCommand|TestVNextGenerationPublisherRefusesPhysicallyStagedCommandPreflight|TestVNextGenerationPublisherRefusesReplacedAtomicControlTemporary|TestVNextGenerationPublisherRefusesReplacedValidatedStageCleanup|TestVNextGenerationPublisherRefusesReplacedValidatedGenerationCleanup|TestVNextGenerationPublisherRefusesReplacedRollbackGenerationCleanup|TestVNextGenerationPublisherRefusesReplacedCommittedJournalCleanup|TestVNextGenerationPublisherRefusesReplacedLockAfterAcquisition|TestVNextPublicationOpenLockRefusesExistingLockWithoutAnchor)$'` → `ok` (3.235s).
  - `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (60.036s).
  - `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (328.097s).
  - `make connectorgen-vnext-locks` → `ok` (30.882s); it includes all `TestVNextGenerationPublisher.*` witnesses and deterministic reference-lock rendering.
  - `go run ./cmd/connectorgen validate internal/connectors/defs` → `connectorgen validate: 553 connector(s) checked, 0 findings`.
  - `make docs-check` → `ok`; `go run ./cmd/connectorgen --help` retained the existing lock-render syntax.
  - `go vet ./cmd/connectorgen`, `go mod tidy -diff`, and `go run ./cmd/agentcontractgen check` → `ok`.
- CLI help/manual/website parity: not applicable because no PM or connectorgen command, flag, output, namespace, generated manual, or website surface changed. The existing generator help was exercised as a regression check; source-lock canon and Atlas documentation changed because the authoring publication behavior changed.
- Manual GSD verification fallback: `verify-work` and `code-review` prompts were generated after the earlier discuss/plan/execute prompts. The adapter has no compatible authorized isolated worker/reviewer, so deterministic tests and this inline self-review are the local evidence; they do not substitute for Firstmate's required fresh exact-SHA review after the normal push.

## CP11 exact-review continuation verification plan — instruction 132

- F1: run separate late-bound private-source replacements for `CURRENT` and `JOURNAL`; refusal must preserve prior control bytes, moved original temporary, and replacement.
- F2: run late stage, pruning-generation, lease-only, rollback-generation, `CURRENT`, and `JOURNAL` replacement tests. Each must prove the operation neither deletes the moved validated object nor the later replacement.
- F3: while the first publisher holds the retained connector-directory lock, replace both sibling lock names with a self-consistent pair. The second publisher must not reach a mutation fault point; after the first completes, recovery and a new publish must work.
- F4: Atlas mapping checks must name physical closed-publication and durable-marker tests, the new lease/current/matched-pair refusal witnesses, and the existing staged command physical preflight witnesses. No unrelated Atlas record is changed.
- Gates after GREEN: focused normal and race selectors, complete normal and race `cmd/connectorgen`, `make connectorgen-vnext-locks`, definition validation, docs/Atlas JSON, vet, build, agent-contract, tidy, exact diff, read-only self-review, normal push, PR API base/head/SHA read-back, external status update, archive `132.msg`, and a fresh Firstmate exact-SHA pause. No standalone adversarial program, provider, credential, or CP12 operation is permitted.

## CP11 exact-review continuation local results — instruction 132

- **Focused GREEN:** the F1–F4 publication/Atlas proof selector passed in 7.962s, including both late atomic controls, all late quarantine cleanup boundaries, matched-pair serialization, physical closed-tree and durable-stage witnesses, invalid-lease refusal, and Atlas selector resolution.
- **Normal package GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 61.981s.
- **TDD provenance:** the corresponding executable RED/GREEN observations are recorded in `TDD-LEDGER.md` under the instruction 132 actual section. F4 is a documented mapping-evidence correction: the report's source-level proof mismatch is the RED evidence; the added physical/durable tests are the GREEN proof. No fabricated failing runtime test was used.
- **Race package GREEN:** `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 329.199s.
- **Corpus/definition/Atlas GREEN:** `make connectorgen-vnext-locks` → `ok` in 32.693s; `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings; `TestFoundationAtlasSelectorsResolve` passed.
- **Maintenance GREEN:** `make docs-check`, `go vet ./cmd/connectorgen`, `go mod tidy -diff`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and `git diff --check` passed.
- **Read-only smoke boundary:** top-level `go run ./cmd/connectorgen --help` rendered usage. `lock-render --help` is not a supported subcommand form and returns its existing usage error; a real checked-in connector `lock-render --check` stops at the expected absent `generations/` directory because this checkpoint deliberately does not materialize a publication corpus. The temporary-root package proofs above exercise publish/check behavior without changing definition state.

## CP11 final F1 continuation verification plan — instruction 133

- F1: run the new final-source-validation replacement test for `CURRENT` and `JOURNAL`; each subtest must observe an identity refusal and preserve prior control bytes, moved original temporary, and the mismatched replacement.
- F4: run `TestFoundationAtlasSelectorsResolve` after updating only the F1 mapping.
- Gates after GREEN: focused normal/race selectors and complete normal/race `cmd/connectorgen`; `make connectorgen-vnext-locks`; definition validation; Atlas JSON/docs; vet; build; agent-contract; tidy; exact diff; inline self-review; ordinary push; PR API base/head/SHA read-back; external status update; archive `133.msg`; then fresh Firstmate exact-SHA pause. No standalone adversarial program or provider/credential action.

## CP11 final F1 continuation local results — instruction 133

- **Focused GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesFinalReplacedAtomicControlTemporary|TestFoundationAtlasSelectorsResolve)$'` → `ok` in 1.568s. The `CURRENT` and `JOURNAL` subtests execute the barrier strictly after final source validation and before namespace installation, then prove identity refusal, prior bytes/inode, moved original temporary, and a quarantined replacement.
- **Complete package GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 65.113s; `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 328.540s.
- **Corpus/definition/Atlas GREEN:** `make connectorgen-vnext-locks` → `ok` in 33.542s; `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings; `jq -e . docs/connector-canon/foundations/catalog.json` and `TestFoundationAtlasSelectorsResolve` passed.
- **Maintenance GREEN:** `make docs-check`, `go vet ./cmd/connectorgen`, `go mod tidy -diff`, `go build ./cmd/pm`, `go run ./cmd/agentcontractgen check`, and `git diff --check` passed.
- **Read-only smoke boundary:** `go run ./cmd/connectorgen --help` rendered the real command usage. No provider, credential, database, source-lock materialization, `.cache`, certification-residue, release-automation, or CP12 action occurred.

## CP11 durable F1 continuation verification plan — instruction 134

- Matrix: execute every listed authority/backup/install/replacement/restore/clear cut for existing `CURRENT`, existing `JOURNAL`, no-prior `CURRENT`, and no-prior `JOURNAL`. Each restart must resolve authority before ordinary control decoding and produce only a valid prior control, valid intended completion after authority clear, or the valid absent first-publication state.
- Mismatch: inject a private-source replacement strictly at the final source barrier; at every post-exposure cut, assert restart retains the replacement, restores/syncs the prior or absent target, removes the repair authority, and does not accept malformed or syntactically valid substitute payload as ordinary control.
- F4: map the changed source-lock-vNext publication guarantee only to durable restart witnesses; retain the synchronous final-boundary test as a non-mapping regression.
- Gates after GREEN: focused normal/race selectors and complete normal/race `cmd/connectorgen`; `make connectorgen-vnext-locks`; definition validation; Atlas JSON/docs; vet; build; agent-contract; tidy; exact diff; design-level crash-cut self-review; ordinary push; PR API base/head/SHA read-back; external status update; archive `134.msg`; then fresh Firstmate exact-SHA pause.

## CP11 durable F1 continuation local results — instruction 134

- **Focused GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesFinalReplacedAtomicControlTemporary|TestVNextGenerationPublisherRecoversInterruptedReplacedControlBeforeDecode|TestVNextGenerationPublisherRecoversEveryControlRepairDurableCut|TestVNextGenerationPublisherRecoversEveryRetainedReplacementCut|TestFoundationAtlasSelectorsResolve)$'` → `ok` in 10.675s. It includes the malformed substitute pre-decode proof, all initial/replacement/restoration/clear durability barriers, both controls, and both existing/no-prior paths.
- **Complete package GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 78.969s; `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 346.125s.
- **Corpus/definition/Atlas GREEN:** `make connectorgen-vnext-locks` → `ok` in 49.058s; `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings; `jq -e . docs/connector-canon/foundations/catalog.json` and `TestFoundationAtlasSelectorsResolve` passed.
- **Maintenance GREEN:** `make docs-check` (including `go build ./cmd/pm`), `go vet ./cmd/connectorgen`, `go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, and `git diff --check` passed.
- **Read-only smoke boundary:** `go run ./cmd/connectorgen --help` rendered the command usage. No provider, credential, database, source-lock materialization, `.cache`, certification residue, release automation, CP12, reset, rebase, or force-push action occurred.
- **Review/lifecycle evidence:** `REVIEW-CONVERGENCE.md` records the manual crash-cut self-review. The resolved GSD prompts were executed inline because Firstmate prohibited role spawning; the external fresh exact-SHA review remains mandatory after the ordinary push.

## CP11 monotonic-authority redesign verification plan — instruction 136

- Demonstrate the real lock/path divergence with a cooperative publisher plus same-permission external `rename`/`unlink` test action; do not substitute a manual-unlock product behavior.
- Assert immutable prepared authority creation before exposure; identity-bound create-only phase transitions; source and target substitution refusal/retention; existing/no-prior `CURRENT` and `JOURNAL`; fresh-process recovery at every persisted authority/phase/public-control cut; recovery before public control trust; cooperative writer serialization; and cleanup identity invariants.
- Require a state-machine table, counterfactual/disconfirming result, focused package selectors, full/race `cmd/connectorgen`, vNext lock corpus, definition/Atlas/docs/static/contract/build checks, a read-only full-range self-review, normal push, PR API read-back, and fresh exact-SHA pause.

## CP11 monotonic-authority redesign local results — instructions 136–137

- **Scope and state machine:** only the `cmd/connectorgen` F1 publication protocol, its proof tests, source-lock-vNext/Atlas authoring documentation, and CP11 evidence changed. `PLAN.md` records the required state table and counterfactual: `prepared` → `installed` or `replacement_retained` → `restored`, with the immutable prepared authority retired and fsynced before disposable private phase/backup cleanup. There is no manual-unlock product surface, legacy authority fallback, F2/F3 change, CP12 action, provider/credential/database I/O, source-lock materialization, cache/certification-residue access, reset, rebase, or force-push.
- **Executable RED provenance:** the initial locked `CURRENT`/`JOURNAL` source/target substitution test failed all four cases with `private prepared recovery authorities = 0, want one`. Review then exposed two additional executable gaps: the pre-exposure prepared-authority replacement test failed `prepared-authority substitution installed a public control before refusal`, and the post-prepared-authority-sync crash matrix failed all eight new cuts because the predecessor had no such durable retirement boundary. `TDD-LEDGER.md` records each command, observed failure, and minimal repair without relabeling a former passing test.
- **Focused F1/Atlas GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRecoversEveryControlRepairDurableCut|TestVNextGenerationPublisherRecoversEveryRetainedReplacementCut|TestVNextGenerationPublisherPrivatePreparedAuthoritySurvivesPublicControlSubstitution|TestVNextGenerationPublisherWritesBoundImmutableRecoveryPhaseChains|TestVNextGenerationPublisherFinalTargetSubstitutionCannotClearPrivateAuthority|TestVNextGenerationPublisherFinalNoPriorTargetUnlinkCannotClearPrivateAuthority|TestVNextGenerationPublisherRefusesSubstitutedPreparedAuthorityBeforePublicControlExposure|TestVNextGenerationPublisherRefusesSubstitutedPreparedAuthorityBeforePublicControlTrust|TestVNextGenerationPublisherRefusesFinalReplacedAtomicControlTemporary|TestVNextGenerationPublisherRecoversInterruptedReplacedControlBeforeDecode|TestVNextGenerationPublisherSerializesMatchedLockAnchorReplacement|TestFoundationAtlasSelectorsResolve)$'` → `ok` in 17.256s.
- **F1 coverage proven:** the focused matrix holds the real cooperative lock while same-permission `rename`/`unlink` mutates public controls; covers `CURRENT`/`JOURNAL`, existing/no-prior state, source and target substitutions, immutable prepared bytes and predecessor-bound create-only phase chains, retained replacements, fresh publisher recovery before public decode at every backup/prepared/install/phase/replacement/restore/authority-retirement/clear persistence cut, final cleanup identities, and cooperating-writer serialization.
- **Changed-package GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 90.003s; `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 352.581s. `make connectorgen-vnext-locks` → `ok` in 58.622s.
- **Authoring/runtime/static GREEN:** `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings; `make connector-canon-check`, `make docs-check`, `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/pm`, `go mod tidy -diff`, `make tidy-check`, `go run ./cmd/agentcontractgen check`, Atlas schema/unique-ID checks, and `go run ./cmd/connectorgen --help` passed. The temporary `pm` binary generated by `make docs-check` was removed.
- **Boundary/preflight GREEN:** `make connector-boundary` returned `outcome: "clean"` for 281 files and 553 connectors, with only the six existing declared exceptions. `go test -count=1 -timeout 20m -run '^TestEveryImplementedCommandPassesRuntimePreflight$' ./internal/connectors/commandrunner` → `ok` in 22.290s.
- **Lint disposition:** full `make lint` reported 43 repository-wide diagnostics (20 `errcheck`, 5 `staticcheck`, 18 `unused`), including pre-existing lines in `vnext_publication.go` and unrelated packages, but no `vnext_publication_repair.go` or new-test finding. The exact changed-line check, `golangci-lint run --new-from-rev f36b5d0a275ed27fd5f4da242ba192e43f8066d5`, → `0 issues`; it caught and prompted repair of the new control-record close-error handling before that final pass.
- **Intentional local exclusions:** `make smoke-no-build`, release-workflow checks, and the unrelated `internal/cli` suite were not run because their credential/database or release paths are expressly prohibited by instruction 136 and no `internal/cli` behavior changed. This is a scoped omission, not a green claim.
- **Delivery gate:** instruction 137 confirms CP11 must receive fresh independent exact-SHA review before CP12. Remaining CP11 actions are read-only full-range self-review, whitespace/secret scan, scoped commit, ordinary push, PR #4294 API base/head/SHA read-back, status update, archive `136.msg`, then pause unchanged for that review.

## CP11 F1 Design B linearization verification plan — instruction 139

- **Authority/scope:** start only from `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`, parent `f36b5d0a275ed27fd5f4da242ba192e43f8066d5`, under the authoritative F1 report. Change only CP11 publication/recovery/check code/tests, source-lock-vNext canon/Atlas witnesses, and phase evidence. CP12, rebase/reset/force-push, checked-in generation materialization, provider/credential/database I/O, `.cache`, and certification residue remain excluded.
- **Required candidate RED:** direct lock-ignoring actor plus real descriptor operations falsify candidate post-validation install rename, restore rename, no-prior public unlink, and final-public-validation→sole-authority-retirement. Assertions require lost/authority-free identity behavior, not merely failure text.
- **Required GREEN matrix:** every `CURRENT`/`JOURNAL` × prior-present/absent × install/restore/logical-absence transition through no-replace capture, create-only selection, terminalization, fresh recovery, real `lock-render --check`, late replacement/unlink, A/B/C substitution, malformed private graph, unsupported/collision errors, successor cleanup substitution, and cooperating-writer serialization.
- **Pass criteria:** every publisher-observed displaced public inode remains reachable under its bound capture identity; decode yields only selected prior/intended identity or fails closed; pending/divergent shared check is nonzero, emits no success stdout, changes no filesystem entry, and cleanup cannot weaken active successor terminal.
- **Planned gates:** focused RED/Green tests; `go test -count=1 -timeout 20m ./cmd/connectorgen`; `go test -race -count=1 -timeout 20m ./cmd/connectorgen`; `make connectorgen-vnext-locks`; definition validation; `make connector-boundary`; Atlas/canon/docs; vet/build/contract; diff/secret scan; exact-range self-review; normal push and GitHub API base/head/SHA read-back.

## CP11 F1 Design B final local verification — instruction 139

- **Focused protocol witnesses:** terminal-transition matrix; every reachable
  terminal-authority durable cut; unsupported/collision/source-absence
  no-replace failures; topology and private-identity refusal; repeated A/B/C
  substitutions; retained predecessor cleanup safety; cooperating writers; and
  all three actual `lock-render --check` paths passed under
  `-count=1 -timeout 20m`.
- **Changed package:** `go test -count=1 -timeout 20m ./cmd/connectorgen`
  passed in 124.842s. `go test -race -count=1 -timeout 20m
  ./cmd/connectorgen` passed in 416.263s. The focused retained-predecessor
  witness also passed under `-race`.
- **Corpus and authoring:** `make connectorgen-vnext-locks` passed; `go run
  ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors
  with 0 findings; `jq -e . docs/connector-canon/foundations/catalog.json`,
  `TestFoundationAtlasSelectorsResolve`, `make connector-canon-check`, and
  `make docs-check` passed.
- **Static and boundary:** `go vet ./cmd/connectorgen`, `go build
  -o /tmp/connectorgen-cp11 ./cmd/connectorgen`, `make tidy-check`, and `go run
  ./cmd/agentcontractgen check` passed. `make connector-boundary` reported
  `outcome: "clean"` for 284 files and 553 connectors, with only its existing
  declared documentation exceptions.
- **Platform check:** native Darwin compilation exercised
  `RenameatxNp(..., RENAME_EXCL)` through the package suite. A direct Linux
  cross-compile of the exact `unix.Renameat2(..., RENAME_NOREPLACE)` API passed.
  A full Linux `./cmd/connectorgen` cross-compile is not claimed: the embedded
  `go-duckdb` dependency failed first with `undefined: Conn` without a Linux
  cgo toolchain.
- **Generated residue:** `make docs-check` created the local `pm` binary; this
  run removed it immediately. `.cache/` and
  `internal/connectors/certifications/.fingerprint-salt` remain untouched.

## CP11 Astra B-01/B-02 correction verification plan

- Frozen subject: `8214bd91403ce620773b61caf674faa540ee1701` over `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`; the candidate is BLOCKed by B-01/B-02. This section starts with planned proof only. It must not relabel source-derived Astra findings as executed RED until the named commands fail.
- B-01 observable RED/GREEN criteria:
  - stale-stage recovery, stale-generation prune, and failed-active-validation rollback each use the real quarantine-restoration pathname. After B is quarantined and C is installed exactly before restoration, RED loses C on the candidate; GREEN keeps C public, B private, and moved A reachable while returning refusal.
  - activation creates a late empty destination at `before_stage_rename`. RED replaces it; GREEN returns a typed destination conflict while both stage and collision remain.
  - no test accepts only an error return: device/inode and sentinel bytes/tree entries are asserted.
- B-02 observable RED/GREEN criteria:
  - a first-head (`CURRENT`) and second-head (`JOURNAL`) base crash each persist valid `prepared.json` and both required directory syncs but no base terminal. RED ordinary fresh `Recover` refuses. GREEN terminalizes the valid base, creates the marker only after both heads are terminal, and permits a normal retry.
  - the real `runLockRender(..., "--check")` path remains exit 1 with empty success stdout and an unchanged snapshot during the interruption. It does not recover or write.
  - malformed/missing prepared, predecessor/successor, graph topology, and terminal-divergence controls retain pre-decode refusal.
- Focused commands, in order:
  1. `go test -list 'TestVNextGenerationPublisher(RefusesSecond|RefusesFinalGeneration|ResumesInterrupted)' ./cmd/connectorgen`
  2. `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation)$'`
  3. `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation|TestVNextGenerationPublisherRefusesLateReplacedValidatedStageCleanup|TestVNextGenerationPublisherRefusesLateReplacedValidatedGenerationCleanup|TestVNextGenerationPublisherRefusesLateReplacedRollbackGenerationCleanup|TestVNextGenerationPublisherBootstrapsRetainedTerminalAuthorities|TestVNextGenerationPublisherCheckRefusesMalformedPrivateTransactionBeforePublicDecode|TestRunLockRenderCheckRefusesPendingPrivateAuthorityWithoutWriting|TestVNextGenerationPublisherRecoversEveryTerminalAuthorityDurableCut)$'`
  4. `go test -race -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation|TestVNextGenerationPublisherRecoversEveryTerminalAuthorityDurableCut)$'`
  5. After coherent GREEN only: `go test -count=1 -timeout 20m ./cmd/connectorgen`, `go test -race -count=1 -timeout 20m ./cmd/connectorgen`, `make connectorgen-vnext-locks`, `go run ./cmd/connectorgen validate internal/connectors/defs`, `make connector-canon-check`, `make docs-check`, `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/connectorgen`, `go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, `git diff --check`, and a changed-path secret/local-state scan that excludes unread preserved residue.
- Platform boundary: native Darwin package execution is required for the no-replace primitive. Linux helper compilation may be recorded separately, but no Linux runtime filesystem claim is made without a compatible cgo toolchain.
- Delivery boundary: no `make verify`, full-project test run, release check, no-mistakes action, provider call, credential access, or normal parent push occurs in this local correction wave. A fresh Firstmate-managed Astra review of the corrected exact SHA is mandatory before publication or CP12.


## CP11 Astra B-01/B-02 correction actual RED

- `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation)$'` → exit 1 in 3.250s, as required before any B-01/B-02 behavior correction.
- B-01: each real stage/prune/rollback quarantine caller overwrote actor-created regular C with B after the production restoration absence check; the activation caller accepted and replaced its late empty final-generation destination. The tests assert bytes and inodes, so the observed failure is destructive behavior rather than an error-status proxy.
- B-02: both durable bootstrap cuts passed the shared check-only nonzero/no-success-output/unchanged-tree assertions, then fresh ordinary recovery refused the valid pre-marker base record with `publication control authority marker is missing from pending repair`. No recovery or retry success is claimed at RED.
- Scope at RED: only `cmd/connectorgen` temporary-root test instrumentation/witnesses and phase evidence changed; no provider, credential, database, runtime-reader, source-lock, rendered-output, `.cache`, certification, release, or CP12 path was touched. Green and all package/race/static/canon gates remain pending.

## CP11 Astra B-01/B-02 correction focused GREEN

- `jq -e 'type == "object"' docs/connector-canon/foundations/catalog.json` returned `true`; `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation|TestFoundationAtlasSelectorsResolve)$'` → `ok` in 5.461s.
- The real A/B/C witness passes all stale-stage, stale-generation, and rollback callers: C's public inode/bytes remain, B's quarantined inode/bytes remain, moved A remains identity-bound and readable, and the no-replace conflict remains discoverable through `errors.Is(err, fs.ErrExist)`. The final-generation activation collision likewise retains both original inodes and returns `fs.ErrExist`.
- The durable bootstrap witness passes both CURRENT-first and JOURNAL-second cuts: shared check-only render is still nonzero/no-success-output/tree-stable before recovery; fresh exclusive recovery appends only the strict base terminal, creates the remaining head/marker, preserves originally absent controls, and a normal render retry succeeds.
- Continuity: the planned selector containing the new witnesses, legacy stage/generation/rollback substitution checks, bootstrap terminal authority, malformed-private authority, check-only pending authority, and terminal durable-cut matrix → `ok` in 47.655s. The planned race selector covering the three correction witnesses plus the terminal durable-cut matrix → `ok` in 36.321s.
- Atlas evidence: existing mappings did not cover the new post-absence restoration, final activation, or pre-terminal bootstrap facts. The narrow canon/catalog maintenance registered the physical/durable tests and `TestFoundationAtlasSelectorsResolve` passed; no source lock, rendered execution JSON, connector route, runtime reader, credential, provider, database, `.cache`, certification, release, or CP12 path changed.
- At the focused checkpoint, full package/race, corpus/canon/docs/static gates, changed-path secret/local-state scan excluding unread residue, local exact-range self-review, and Firstmate-managed Astra disposition remained pending. The completed local gates are recorded below; only Firstmate review remains external.

## CP11 Astra B-01/B-02 correction final local verification

- **Package:** `go test -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 157.507s. `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` in 462.893s.
- **Publication/canon:** `make connectorgen-vnext-locks` → `ok`; `make connector-canon-check` → `connector canon check: ok`; `go run ./cmd/connectorgen validate internal/connectors/defs` → 553 connectors, 0 findings. Catalog JSON parsed with `jq`, and the mapped `TestFoundationAtlasSelectorsResolve` passed.
- **Static/CLI/docs:** `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/connectorgen`, `go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, `git diff --check`, and `go run ./cmd/connectorgen --help` passed. `make docs-check` validated connector docs; its generated root `pm` binary was removed and absence was confirmed.
- **Scope/safety:** the literal-secret scan of exactly the changed source/canon/catalog/evidence paths returned no matches. The only broader identifier hit was a local directory-name `token` variable. No source lock, rendered execution JSON, connector, provider, credential, database, runtime reader, `.cache`, certification residue, PR, push, or CP12 operation occurred.
- **Local delivery boundary:** this evidence establishes a local candidate only. Commit normally on the existing branch, record the resulting exact SHA for Firstmate, and await a fresh Astra review before every further mutation.

## 2026-09-06 — CP11-R3-01 local verify-work record

- **Scope and candidate boundary:** this is the single deduplicated FIFO-reader repair from the frozen CP11 review range `36e4d980de0d51d92fe74a68306845643596a6cb..0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e`. The implementation delta is confined to `cmd/connectorgen/vnext_publication_dir.go` and its temporary-root regression in `cmd/connectorgen/vnext_publication_test.go`; the accompanying plan/TDD records are the GSD evidence. It adds no source lock, rendered execution artifact, runtime route, connector executor, credential/provider/database operation, cache/certification input, public CLI surface, or CP12 work.
- **GSD verify-work obligation:** `node scripts/gsd prompt verify-work 4427` was generated and read. Under the exact inline fallback in `.agents/agentic-delivery/references/gsd-pi-adapter.md` **Agent requirements** (generate a prompt for a non-Pi runner, then execute it inline when a compatible isolated runtime is unavailable), this section performs the requested acceptance verification directly. It is not a claim that a Pi `/gsd-verify-work` process ran. The fresh independent Astra exact-SHA review remains the separately required code-review gate.
- **Behavioral proof:** the pre-fix selector exited 1 in 2.932s because bounded children replacing a stale-stage owner and an enumerated admission member with no-writer FIFOs exceeded their one-second limits before descriptor type validation. The same selector is now green in 2.200s after the final test-cleanup adjustment: stale-stage `Recover` and public `CURRENT` recovery refuse promptly, real `lock-render --check` rejects FIFO `JOURNAL` with exit 1 and no success output, and the actual `vNextPublicationDirectoryFS.Open` path allows a nested schema directory/regular artifact before rejecting an enumerated-to-FIFO member. Every case preserves the FIFO under `Lstat` and permits a subsequent exclusive operation lock.
- **Repair and preserved contracts:** `openRegular` adds `O_NONBLOCK` to the existing descriptor-relative `O_NOFOLLOW` read before `Stat()` validates the exact opened descriptor. The new `openFilesystemMember` applies the same nonblocking descriptor validation to the semantic-admission filesystem adapter, allowing only a regular member or a genuine directory for nested traversal. Writes, create/exclusive behavior, leases, identity checks, byte limits, and existing error paths are unchanged. The continuity selector retains B-01 no-clobber activation/restoration, B-02 strict interrupted-base recovery, and check-only authority tests; full package runs exercise that selector.
- **Changed-package GREEN:** `go test -count=1 -timeout 20m ./cmd/connectorgen` passed in 132.575s and `go test -race -count=1 -timeout 20m ./cmd/connectorgen` passed in 409.519s. `make connectorgen-vnext-locks` passed; `make connector-canon-check` reported `connector canon check: ok`; `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with zero findings; and `make connector-runtime-preflight` passed against the real commandrunner preflight.
- **Static/repository gates:** `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/connectorgen`, `go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, and `git diff --check` passed. `make docs-check` passed. `make connector-boundary` reported `outcome: clean`, 284 files checked, 553 connectors loaded, zero findings/warnings, and only its six existing declared exceptions. `make release-workflow-check` passed all pinned-action, release-target, tooling, size-budget, and production-layout assertions.
- **Lint disposition:** repository-wide `make lint` remains red on 51 pre-existing diagnostics, including historical `vnext_publication.go` defer-close findings and unrelated checked-out source; none are introduced by this FIFO repair. The changed-line proof `golangci-lint run --new-from-rev=HEAD ./cmd/connectorgen/...` initially found one unchecked close in the new test, which was repaired; its final result is `0 issues`.
- **Smoke classification and result:** Firstmate's scope clarification was applied before this required gate. `make smoke` creates one new `mktemp` project and passes that root to every `pm` command; its source is the deterministic in-process `sample` connector, while warehouse and outbox paths are confined below the same root. It creates no existing credential-store entry and makes no provider, customer-database, or network request. The successful run proves a nonempty connection-owned Parquet table and owner record plus a nonempty local outbox JSONL after plan, preview, stdin approval, and reverse execution. The exact disposable smoke root and ignored locally built `pm` artifact were moved to Trash after their assertions, so both remain recoverable and no smoke residue is in the worktree.
- **Platform/baseline boundaries:** a CGO-disabled Linux cross-compile cannot build the project because the embedded DuckDB dependency fails first (`undefined: Conn`); that is recorded as a build-environment limitation, not Linux runtime proof. The separately required `go test -count=1 -timeout 20m ./internal/cli` ran for 223.511s and reproduced only four inherited failures — static polling-help text, two source-bound-origin preflight expectations, and ETL-help text — along with expected local Redis connection-refused logs. No failing test names or source paths are in this repair delta.
- **Next gate:** stage only the four reviewed files, commit the coherent local CP11-R3-01 repair, bind its exact code SHA here and in the Firstmate status, then obtain the fresh independent Astra/xhigh review over the full original CP11 range plus this repair/evidence delta. CP11 is not accepted and CP12 remains prohibited.

## 2026-09-06 — CP11 F-07 historical signal provenance correction

This is a record correction, not a current rerun. The earlier statement that
the pre-routing RED command/output/edit was unavailable is superseded by the
metadata-bound recovery at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-F07-signal-provenance-recovery-7294373166db75466e2c92269f7887f51ceaddc6.md`.

- Worker record 3224/event 3223/physical line 3224, SHA-256 `65acae633258da13e38c4a9e0a64d532009bc2555b35ffc681a31b8d8828a14f`, ran `gofmt -w cmd/connectorgen/vnext_publication_test.go && go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestConnectorgenMainPreservesNonConsumingSignalTermination$'` and exited 1: `validate` did not terminate for the parent plus `interrupt`, `terminate`, and `repeated-interrupt`; package result was `FAIL polymetrics.ai/cmd/connectorgen 12.383s`.
- Record 3230/event 3229/physical line 3230, SHA-256 `9e4d444101f045b996d3c19ebe5e4d61ecd4abc2b2ab256b7d583bd2d88a3aa9`, records the actual `main.go` routing change: non-`lock-render` calls `run`, and only `lock-render` creates `NotifyContext` through `runMain`.
- Record 3238/event 3237/physical line 3238, SHA-256 `b0501c5a52f0fa821bd27907990bae074199e1bfb2563cb4cab3ea9f8c11fcff`, ran the combined non-consuming/lock-render selector after that edit and exited 0 / `ok polymetrics.ai/cmd/connectorgen 15.417s`.

The failing snapshot was an uncommitted worktree, so no full Git/tree SHA is
available; later commit `7294373166db75466e2c92269f7887f51ceaddc6` must not be
called the failing state. This correction does not establish provider-live
behavior, a forced second-signal guarantee, or cancellation in the middle of
a publication transaction.

## 2026-09-06 — CP11 F-01–F-08 current local validation

- **Focused evidence:** Group 1 F-04/F-08 active selector, Group 2 F-01/F-02/F-03 selector, and Group 3 identity/durable-caller matrix are retained in `GROUP1-EVIDENCE.md`, `GROUP2-EVIDENCE.md`, and `GROUP3-EVIDENCE.md`. Group 3 includes the actual committed/new-selected `BeforePrune` cut, successful Publish final prune, fresh old-selected/rejected-new empty/nonempty variants, immediate rollback variants, stage ownership, Check read-only behavior, and the F-03 Publish-initial-recovery close caller.
- **Final package/race source state:** after correcting two new F-02 test cleanup error checks, focused F-02 passed in 1.274s and new-only lint reported `0 issues.`. `go test -count=1 -timeout 20m ./cmd/connectorgen` then passed in 263.677s; `go test -race -count=1 -timeout 20m ./cmd/connectorgen` passed in 691.666s.
- **Current static proof:** formatter diff, `go vet ./cmd/connectorgen`, both `cmd/connectorgen` and `cmd/pm` builds, `go mod tidy -diff`, generated agent-contract check, and diff check passed. In this same coherent wave, the source-lock publication test gate, connector canon check, 553-definition validation, runtime preflight, docs check, clean 284-file/553-connector boundary report, and release workflow check passed. No provider, credential, database, cache/certification input, no-mistakes pipeline, push, merge, or CP12 operation occurred.
- **Remaining gate:** commit this reviewed coherent local wave, bind exact SHA/tree and changed paths, then freeze source for the separately prescribed Astra/xhigh whole-range review and Luna/low final mechanical index. These local results do not accept CP11.

## 2026-09-06 — CP11 frozen code boundary for final review

- **Frozen behavior code:** `e77cd390f45eb938917dc3c39882bda34aae09a6`, tree `4c37521b11a6f1a22bd6fa5ea5e043d4982329f8`, on `fm/cli-top100-declaration-batch-r1`. The published head `8214bd91403ce620773b61caf674faa540ee1701` is its verified ancestor. The preserved initial CP11 candidate is `0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e`; the final reviewer must inspect that full original range plus this code delta, not only this commit.
- **Frozen behavior paths:** `cmd/connectorgen/vnext_publication.go`, `vnext_publication_dir.go`, `vnext_publication_repair.go`, their existing tests, and the new capture-identity, durable-matrix, observation, and resource-error fixtures. Canonical plan/TDD/verification/review records and `GROUP1-EVIDENCE.md`, `GROUP2-EVIDENCE.md`, and `GROUP3-EVIDENCE.md` accompany the code. The inherited `DISCUSSION-LOG.md`, TDD-order correction, premature-F01 snapshot, source-lock canon, Foundation Atlas, and `main.go` remain part of the full original CP11 review range as applicable.
- **Frozen validation:** final current `cmd/connectorgen` normal/race results are 263.677s/691.666s after new-only lint reached zero; the concrete source-lock/canon/553-definition/preflight/docs/boundary/release/static results are recorded above. No source/test behavior may change during the final review pass. A later artifact-only commit may bind this code SHA to read-only reports; it must not be called a behavioral retest or obscure this identity.
- **Final review gates:** Firstmate's specified fresh GPT-6 Astra/xhigh whole-range review and Luna/low final mechanical index receive this exact code SHA/tree, GROUP1/2/3 evidence, `TDD-ORDER-CORRECTION.md`, and preserved exact-baseline fixture/overlay provenance. Their reports are evidence only; CP11 remains unaccepted and no push/CP12/no-mistakes action follows automatically.

## CP11 Group 2 F-03 verification completion tracker — steer 058

The post-repair focused Group 2 selector and its intermediate same-worker
package receipt are not final Group 2 verification. The receipt's complete
returned event is preserved at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group2-intermediate-package-raw-result.json`
(physical `10897`, raw SHA-256
`4fd7390eef0b432f8f5f983d3924f9360bde2993d51586fa48bc6658634e0255`): package
`271.387s`, wall `273.672594250s`, then-uncommitted source state only.

At creation, before a Group 2 completion statement, the required work was to
execute and record the remaining
resource-backed F-03-A pre-record/Write/short-Write/Sync/real-Close and
post-record directory-sync frontiers with valid bootstrap/successor graph
classes; every F-03-B listed producer/consumer compound-cause case; and the
F-03-C CURRENT/JOURNAL plus stage/generation reachability cases. The case,
selector, hook, real resource, pre-restoration graph/identity/residue, pending
nonmutating Check, fresh recovery/retry, and normal/race result belong in
`GROUP2-F03-REPAIR-EVIDENCE.md`. Until then this section is an explicit
incomplete verification tracker, not a later-source attribution or CP11
acceptance.

### CP11 Group 2 F-03 executed verification — steer 058/060

The tracker cells are now executed at the current test source
`cmd/connectorgen/vnext_publication_group2_original_test.go`. Its explicit
seams are `vNextPublicationControlRecordHooks`, the record/directory-sync
fault points in `vnext_publication.go` and `vnext_publication_repair.go`, the
owned raw opened-file Close in `vnext_publication_dir.go`, and the
read-control completion/Close points in `vnext_publication.go`; each is
nil/direct in production.

- `TestCP11F03ARepairPreparationFrontierMatrix` proves all seven preparation
  cut classes in three actual transitions (JOURNAL absent→present, CURRENT
  present→present, JOURNAL present→logical-absence), retaining the appropriate
  phase-free graph or no unexposed cleanup, identity-bound anchors/public
  state, nonmutating Check, and fresh Recover/Check/retry. Its Sync/Close test
  performs the real operation before injecting its completion result.
- `TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation`
  preserves absent/absent base bootstrap; the current
  `TestCP11F03ARepairBasePresentPresentAuthorityRecovers` does both CURRENT
  and JOURNAL present/present bases through interruption and terminal-head
  recovery.
- `TestCP11F03BRepairCompoundCausesRemainInspectable` covers every listed
  compound resource/consumer and pure absence. The four
  `TestCP11F03CRepair*` selectors cover temporary and generation/stage
  quarantine A/B ownership, exact nonempty B bytes and residue before
  fixture-only cleanup, meaningful identity refusal, and fresh retry.

Current exact results: F-03-A normal `20.804s`; its three bounded race classes
`10.200s`, `10.089s`, `11.507s`; base normal/race `2.617s`/`4.095s`; and B/C
normal/race `8.442s`/`11.749s`, all `ok`. Earlier complete physical
`11540/11549/11563/11577` receipts are preserved as the prior two-class source
state. The F-03 dynamic proof is complete; Group 2's source/static/final-review
and CP11 acceptance gates remain open.

### CP11 Group 2 execution binding; Group 3 verification boundary — steer 061

- **Exact completed checkpoint:** `54746816735a964d0177a7a64646d29561f08180`
  binds the current F-03 source/test/evidence tree. It is not a new independent
  review or CP11 acceptance. The unchanged Group 1 F-04-R/F-08-R normal/race
  results (package `9.198s`/`15.924s`) remain their existing post-e77 evidence;
  they were not rerun solely for this record.
- **Accurate F-03-C wording:** the executed B substitutions are directories—an
  empty directory and a directory carrying foreign bytes. The preserved
  `10897` full-package and `11540/11549/11563/11577` receipt files describe
  only the recorded earlier source states.
- **Next verification:** Group 3 must add proportional focused normal and race
  coverage only for the changed public nested-quarantine and durable-witness
  tests. Its oracle must observe raw control/object identities and real private
  authority before restoring fixture-owned interference; it must not promise
  all-or-nothing recursive deletion or automatic restoration. After Group 3,
  execute the specified coherent current-source package/race/static boundary
  and submit the whole original range plus final SHA to fresh independent
  Astra/xhigh review. No e77 report is reused as that review.

# Batch R1 vNext source-lock cutover

## Task Delivery Header

- Parent issue: Refs #4325 — Batch R1 scalable synchronization execution amendment
- Current child: Refs #4425 — A1 manifest-selected executor registry
- Base branch: `main` (API-confirmed PR #4294 base)
- Merges into: `main` through the existing PR #4294
- Delivery: one writer pushes only independently green, ordinary fast-forward commits to the existing `origin/fm/cli-top100-declaration-batch-r1` head; do not open a per-slice PR, force-push, or merge.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`
- Task: CP06 establishes explicit `app.Open` executor construction and a sealed compatibility adapter. It removes process-global default-registry authority, rejects duplicate factory/connector identities before construction, and preserves only named compatibility inventory; it does not activate API or database manifest selections, migrate a connector, alter source locks, add an executor, or create a second execution route.
- Verification: record RED/GREEN evidence for process-global selection and duplicate construction; run focused registry/App/command-boundary tests plus race/vet, the existing renderer/definition checks, residual and secret/local-state scans, and the frozen-SHA self-review procedure in `REVIEW-CONVERGENCE.md`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Production construction has no process-global default-builder authority | live | A registry construction test installs a process-global builder and proves `app.Open` still calls the explicit factory path; prior behavior can be selected by import/init order. |
| Duplicate constructor identities fail before object construction | live | Instrumented fake constructors prove a duplicate ID returns a typed error with zero constructor calls. |
| Compatibility is sealed and command paths remain reachable | live | The production registry's named compatibility inventory reaches the existing credential or typed-block boundary without a config/name fallback. |

## Non-negotiable architecture

- `source.lock.json` is immutable authoring and evidence input only. It is never embedded in or read by runtime.
- One pipeline owns generation: vNext source lock → canonical per-operation descriptor with shared schema registry and request/response refs → deterministic execution files (`metadata`, `spec`, streams and schemas, writes, operations, CLI surface, optional sync transport and rate limits).
- Runtime reads only execution JSON. Legacy importers, retained artifacts, source projections, certification records, root declaration ledgers, source locks, hashes, and retention-only admission are removed from runtime/admission with **no compatibility reader, fallback, feature flag, or second route**.
- Existing engine, commandrunner, REST/GraphQL/multipart encoders, credential/approval boundary, DuckDB/warehouse path, and sync transport machinery remain when they execute declared artifacts.
- Diagnostics may report authoring omissions, but only malformed/missing required execution JSON, ambiguous bindings, missing actual encoders/executors, invalid bounded routes/schemas, invalid invocation/approval/auth, and incompatible sync executor/mode may block runtime.
- Each rendered connector explicitly declares supported and empty lanes. A source operation may feed multiple lanes. Direct operations remain direct; binary, ETL, reverse-ETL, and sync transport semantics stay distinct.

## Deletion scope

1. Remove legacy `connectorgen` source import/projection/materialization/retention, declaration-admission, operation-evidence, and certification command paths and their runtime-coupled helpers/tests.
2. Remove source-lock, retained artifact, certification, API-surface evidence, and global admission-ledger embedding/loading from `defs` and `engine`.
3. Replace direct-read and deferred visibility gates that consult source/certification ledgers with execution-bundle self-consistency checks.
4. Keep source facts only in connector-local vNext locks and deterministic authoring tests; keep execution facts only in rendered JSON.
5. Update Foundation Atlas entries in the same change: retire the two legacy source foundations and record the vNext authoring renderer. No new runtime foundation is planned.

## TDD slices and test matrix

| Slice | Happy | Bad | Edge |
| --- | --- | --- | --- |
| JSON-only runtime | GitHub, GitLab, and Asana discover and reach the credential/approval boundary with provider I/O disabled. | A selected connector with malformed or missing required execution JSON is rejected. | One malformed connector does not suppress unrelated healthy connectors. |
| No legacy admission | Commands bind from operations/CLI execution JSON without source locks, importers, certification, retained evidence, or root ledgers. | Ambiguous command/operation bindings and missing encoders still fail. | A documented deferred command reports its concrete foundation gap rather than being hidden by certification or retention state. |
| Deterministic renderer | Re-rendering a vNext lock is byte-stable and equals checked-in execution JSON. | Unknown schema refs, invalid bounded routes, and contradictory lane bindings fail authoring validation. | One source operation populates multiple supported lanes without turning a direct operation into a warehouse pipeline. |
| Lane coverage | Direct read/write, binary download/upload, ETL, reverse ETL, sync transport, and explicit empty lanes are surfaced. | A sync mode without its actual executor is rejected. | Optional sync transport/rate-limit files remain absent when explicitly unsupported and deterministic when configured. |

## Rollout order

1. G0: freeze the direct-parent delivery amendment, parent/base/denominator, and local certification-tree disposition; commit and normally push the planning-only checkpoint.
2. #4423 N1: commit characterization tests red-first for GitHub, GitLab, and Asana, then restore the executable proof baseline without a runtime behavior change.
3. Implement the vNext canonical descriptor/renderer and remove every legacy runtime/admission dependency locally only through the later authorized child sequence.
4. Materialize and verify GitHub, GitLab, and Asana; commit and push the green reference cohort only when its later child gate is reached.
5. Migrate green connector-local cohorts in order: Bitbucket, CircleCI, Docker Hub, Jira, Notion, Sentry, Stripe, Vercel only after the prerequisite children are green.
6. Preserve source mapping if an actual shared executor is absent and report the exact gap before implementing any new genuine shared runtime foundation.

## G0 direct-parent delivery amendment

- Authority: #4325 comment `5500153864` (2026-09-01T20:41:45Z) requires the full issue lifecycle and exactly one code writer, with ordinary fast-forward commits directly to the existing Batch R1 parent branch. #4294 comment `5500165004` records that routing correction; the PR body was not changed after `gh-axi pr edit` failed before mutation on deprecated `projectCards`.
- Immutable delivery denominator: **4,341 primary retained source operations**, as published on #4325. N1 is a proof-baseline repair only; it neither changes that denominator nor re-pins a source lock, execution manifest, generated connector artifact, or runtime behavior.
- Immutable delivery base: after `git fetch origin fm/cli-top100-declaration-batch-r1`, both `HEAD` and `origin/fm/cli-top100-declaration-batch-r1` resolve to `d260b725ce6f53403961d7af1ef48ea6651cdd66`; its merge base with `origin/main` is `813f457a925f7ee3fe3bea101a43e445992c8552`. This continuation does not rebase or recreate any excluded local work.
- Certification-tree disposition: `HEAD` and the index contain no `internal/connectors/certifications/` path. The frozen checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7` deleted the former tracked legacy certification tree. The sole current untracked item, `internal/connectors/certifications/.fingerprint-salt`, has no Git history and is not ignored; its opaque local provenance is not repository ownership. It remains unstaged, unread, unmodified, and out of scope for G0 and N1. No certification route is restored or retained.

## Lifecycle record

- Restarted from the authoritative remote checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7`, proved reachable from `origin/fm/cli-top100-declaration-batch-r1` before any edit. The excluded `/private/tmp/cli-batch1-vnext-legacy-cutover-r1` worktree is neither a source nor a recovery target.
- The prior TDD ledger's pending rows conflict with the claimed reference-cohort green checkpoint. Those claims are not carried forward as executable evidence: this continuation starts with a new native fixture-bypass RED proof and records its matching GREEN result before production cleanup.
- `scripts/gsd doctor` was run at restart and exited 1 solely because `.gsd/prompts/issue-122-rebootstrap.md` is absent. `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` passed. The generated `discuss-phase` and `plan-phase --tdd` prompts require `.planning/ROADMAP.md`, which this established custom phase does not contain. The documented inline/manual-GSD fallback therefore records discussion, plan, TDD, execution, verification, and review evidence in this phase directory; it is not a lifecycle waiver.
- Loaded guidance: repository skill routing; project and Firstmate `connector-lane-build-order`; `go-engineering` and its ETL/security guidance; `tdd`; GSD Pi adapter; and CLI help/docs/website parity. `skill://golang-how-to`, the repo-named task-specific Go skills, and `skill://firstmate-exhaustive-review` are unavailable in this session, so `go-engineering` is the recorded substitute rather than a false claim that those skills were loaded.
- Architecture inputs: `docs/connector-canon/SOURCE-LOCK-VNEXT.md`, current vNext loader/renderer tests, reference locks, the connector lane contract, this plan, and `VERIFICATION.md`. Historical source/certification/retention rules are deletion inventory, not current runtime authority.
- Independent Firstmate review over `1655123262586b2eaa395aa75b0e54bd7c4558bd..c5bf5c5d544e85dcca5eac3ebed45ba78ad7fb33` returned **BLOCK** as an N1-to-S1A unlock. It confirms the N1 proof repairs, but it does not certify N1 or authorize the old fixture/native/registry paths. Captain applies the already-accepted D1/D2 decisions through the S1A correction map.
- S1A must add and execute RED gates before production deletion: exact-count fixture credential-boundary refusal, no-skip production registry/binary implemented-command sweep, executor-identity collision rejection, and hostile-origin/no-secret-send. Then correct the Atlas to name open whole-bundle publication and strict source-lock-decoding gaps unless the approved D2 transaction implements them, make execution manifests authoritative, remove API native overwrites and delegating hooks, and preserve execution-JSON-only defs plus existing credential, approval, and bounded-request controls.
- Captain-selected full migration supersedes the partial API cleanup checkpoint: retain Amazon SQS as the approved bounded non-generic disposition, preserve DynamoDB/MySQL/PostgreSQL unchanged, and migrate the 28 removed API hook/native executor families — Apify Dataset, Ashby, AWS CloudTrail, Babelforce, Basecamp, Bunny Inc, Canny, Copper, Dixa, Fastbill, Feishu, Free Agent, Freightview, Google Analytics Data, Google Classroom, Google Pagespeed, Less Annoying CRM, Lokalise, Mendeley, Mercado Ads, Metabase, Mode, My Hours, Pocket, Prestashop, Rootly, SafetyCulture, and Yahoo Finance — through source-lock-selected generic execution.
- Each family slice must author provider route/auth/response facts without credentials or provider mutation, declare all seven lanes truthfully, prove generic-engine check/read/ETL through a local fake server and returned production registry, remove `__legacy_hook`, fixture-mode, and caller-controlled-origin metadata, and regenerate its docs/skills/catalog surface. The Atlas change must include a global selector resolver plus owner, guarantee, selector, and proof updates. No push or review occurs until the full migration receives one fresh exact-SHA independent OMP review.
- Feishu Foundation Atlas discovery: **reuse** `runtime.provider-extension-seams.v1`; the existing source-selected `custom` authentication seam admits its sole fixed token request through `DeclaredRouteRequester`, while the rendered engine bundle owns Bitable check/read/pagination. It changes no shared runtime contract and introduces no alternate executor.
- FreeAgent Foundation Atlas discovery: **constrained extension** of `authoring.source-lock-vnext.v1`; the schema-4 auth descriptor now selects standard OAuth2 refresh-token HTTP Basic client authentication in the existing generic encoder. The engine has no connector-name branch, all credentials remain auth-bound, and no hook/native executor or second route is introduced.
- Freightview Foundation Atlas discovery: **reuse** `authoring.source-lock-vnext.v1` and its existing declared `oauth2_client_credentials` path. The lock selects fixed client-credentials, cursor, and request-backed fan-out contracts; no shared runtime behavior, hook, or native executor changes.
- Google Analytics Data Foundation Atlas decision: Captain resolved `[key=ga4-response-header-mapping]` by approving `runtime.response-header-projection.v1`. The bounded generic foundation maps only each source-declared response header to an equal-position row value, rejects unknown/duplicate/missing/mismatched values before any record emits, and has no hook, raw transform, caller origin, credential mutation, or database path. GA4 source locks select it for the five report streams without a connector-name branch or capability demotion.
- PageSpeed Foundation Atlas decision: Captain approved `runtime.cartesian-config-fanout.v1` only for the bounded PageSpeed HTTPS URL × mobile-or-desktop strategy product, fixed declared repeated categories, source-lock cardinality limits, and hard pre-I/O request budgets. It is not an arbitrary transform or multi-axis fan-out language; the rendered bundle selects it without a native reader, hook, caller origin, or fixture route.
- Mendeley Foundation Atlas decision: Captain resolved `[key=mendeley-per-stream-accept]` by approving `runtime.per-stream-headers.v1` for source-declared static vendor `Accept` media types only. The foundation rejects config/secret interpolation, caller injection, protected and transport-control names, non-vendor media types, and unbounded maps before I/O; it merges the stream value deterministically with global headers. Mendeley selects its four fixed media types without a hook or native reader.
- My Hours Foundation Atlas decision: Captain resolved `[key=my-hours-password-token-auth]` by approving `runtime.declared-password-token.v1`: a source-declared fixed HTTPS `POST /tokens/login` JSON `{email,password}` exchange, top-level `accessToken` extraction, per-runtime bearer cache, no refresh, no replay, no caller origin/control, and no generic password grant or custom/native auth route.
- Tenant-origin Foundation Atlas completion: Captain-approved `runtime.tenant-origin.v1` now selects only a declared connection config key and optional fixed append path. The resolver rejects unsafe origins before I/O and is shared by ordinary runtime construction and default stream-route resolution; PrestaShop and Metabase select it without a caller route or hook.
- Metabase Foundation Atlas completion: Captain-approved `runtime.declared-session.v1` accepts only the declared static HTTPS `/session` shape, sends the fixed JSON `{username,password}` exchange through the resolved tenant runtime, extracts top-level `id`, disables exchange retries, caches per runtime, and writes only `X-Metabase-Session`. Metabase selects the ordinary session-token header first when present, otherwise this bounded exchange.
- Yahoo Finance Foundation Atlas completion: Captain-approved `runtime.array-zip-projection.v1` permits only declared scalar copies and equal-length declared response arrays. It emits one record per index or fails before emission; it has no transform language, hook, native executor, caller origin, credential, or provider-specific runtime branch. Yahoo Finance selects it for documented chart OHLCV arrays.
- Exact-SHA review remediation: Ashby now reuses the fixed-origin source-lock path with strict configuration; PrestaShop reuses `runtime.tenant-origin.v1` plus standard Basic authentication and the generic offset reader; Yahoo selects the review-required `runtime.declared-response-error.v1`. The new error foundation is a bounded source-declared object check before extraction, never a connector-name branch, hook, transform, retry, or provider-specific executor.
- Second exact-SHA remediation: `runtime.declared-response-error.v1` now also admits only a source-declared boolean success envelope; Ashby selects it for `success=false` responses before records emit. Captain's PrestaShop correction authorizes `runtime.offset-count-pagination.v1`, a narrow `limit=offset,count` ETL-only window with no caller pagination controls. The shared connector network boundary validates declared configuration before every engine-owned provider call.




## Commit checkpoints

1. G0 planning-only direct-parent amendment and local-state disposition; ordinary push to the existing remote head.
2. #4423 N1 characterization/red-test contract and green executable proof baseline, with no runtime behavior change.
3. Later authorized legacy removal plus vNext renderer green.
4. Later GitHub, GitLab, and Asana rendered bundles and proof green; normal push to the existing remote head.
5. Later deterministic connector-local rollout commits and pushes for the remaining Batch R1 cohort.
6. Final verification, code review, generated docs/help checks, and no-legacy dependency audit.

## CP06 pre-implementation impact map

- Issue / parent: Refs #4425 — A1 manifest-selected executor registry; parent Refs #4325.
- Parent ref / PR base+head: `fm/cli-top100-declaration-batch-r1` → `main` through PR #4294. On 2026-09-03, local `HEAD`, `origin/fm/cli-top100-declaration-batch-r1`, and the GitHub API head all resolved to `cb33f2ef2584c7307f3c88b869c474796698b0fb`; API base is `main`, state is open, and frozen checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7` is an ancestor of `HEAD`.
- Analysis base SHA / last green code SHA: `cb33f2ef2584c7307f3c88b869c474796698b0fb`. The existing dirty CP06 source/test/generator work predates this reconciliation; it remains preserved, uncommitted, and is neither a GREEN candidate nor authorization to stage a production path.
- Brief revision: Firstmate inbox `105.msg` and `107.msg`; impact protocol `/Users/karthiksivadas/karthik-agent-workspace/data/cli-preimplementation-impact-analysis-r1/report.md`.
- Intended observable delta: a generated compact entry binds every execution bundle to exactly one closed executor identity; metadata callers read only that index, and executable callers acquire one digest-bound bundle through the App construction root. Duplicate, empty, unknown, or multiple executor/extension identities fail with typed errors before a constructor, credential, filesystem, or provider action.
- Non-goals / forbidden actions: no connector migration, source-lock/rendered-execution mutation, provider request/credential, release work, generic executor/service/analyzer, runtime selector, compatibility fallback, or database capability change. Existing engine bundle decoding remains the sole execution decoder.
- Expected changed paths: planning only until this disposition. The first code wave is restricted to `cmd/connectorgen/{gen.go,gen_test.go}`, generated `internal/connectors/manifestindex/index_gen.go`, `internal/connectors/manifestindex/{index.go,index_test.go}`, `internal/connectors/manifeststore/{store.go,store_test.go}`, `internal/connectors/bundleregistry/{manifest_loader.go,factories.go,hooks.go,registry.go,*_test.go}`, `internal/connectors/hooks/*` only for the exact 49 factory cutovers plus `hookset/hookset_gen.go`, `internal/connectors/native/nativeset/{factories.go,nativeset_test.go}`, `internal/app/{app.go,app_test.go}`, the affected `internal/cli` construction/help/docs/skills callers and tests, this Atlas entry, and these four phase artifacts. Any other path is an Impact delta and requires a new map row before edit.
- Expected exported symbols: generated immutable `manifestindex.Entry`/index access; a digest-bound store handle; a concrete `bundleregistry` construction value and closed executor/extension factories; and an App construction seam. Their exact exported names remain provisional until the fallback caller inventory is rerun immediately before changing them; no compatibility alias or process-global replacement is permitted.
- Tool availability: CodeGraph=absent (the MCP reported no `.codegraph` index); Go LSP=unavailable (only `clangd` is configured). Fallback completed with exact built-in searches over `internal/app`, `internal/cli`, `internal/connectors`, and `cmd/connectorgen`; scoped `go list -deps -test`; `go test -list` controls; base-source reads; and the exact Firstmate impact report. `scripts/gsd doctor` resolved the adapter except for its known missing `issue-122-rebootstrap.md`; all five lifecycle sources resolved and `go run ./cmd/agentcontractgen check` passed.
- Reused exact-SHA evidence: CP04 `218ba977e`, CP05 code `6be9e9b02`, CP05 review evidence `cb33f2ef2`; those prove only local primitives. They do not prove a production index/store bridge, selected construction, or zero-I/O rejection.

### Entry and data flow

| Step | Owner file:symbol | Input -> output | I/O/store/resource | Invariant |
| --- | --- | --- | --- | --- |
| Generated compact index | `cmd/connectorgen/gen.go:genManifestIndex` → generated `manifestindex/index_gen.go` | existing rendered execution JSON only → sorted connector, generation, digest, metadata, exactly-one executor, and exact extension IDs | generation reads checked-in execution JSON; check mode must be byte-stable | runtime never reads `source.lock.json`; runtime configuration, connector-name branching, import order, and registration order cannot select an executor |
| Metadata use | `manifestindex.Index:ListPage,Lookup` and CLI metadata facade | compact entry → list/help/discovery metadata | zero full bundle decoding for list/lookup | bounded page and continuation; unknown/duplicate/oversized entries fail before allocation or I/O |
| Single-bundle bridge | `bundleregistry` manifest loader over `manifeststore` | exact entry `{connector,generation,digest}` → held bundle handle | one bounded cache entry; loader delegates only to `engine.Load(defs.FS, connector)` and validates the held identity | no alternate JSON reader, raw source lock, or cross-generation/digest return |
| Executable construction | `app.Open` → concrete `bundleregistry` construction | validated catalog + one held bundle → one connector | construction validation precedes project/vault/approval/state/coordinator work | App and direct CLI command use the same entry/generation/digest/executor; factory errors are returned, never panicked |
| Extension fan-in | `cmd/connectorgen/gen.go` → 49 explicit hook package factories → `hookset.Factories` | exact generated factory inventory → immutable extension catalog | no blank import or global hook registry is consulted | duplicate/empty/unknown extension IDs fail before executor construction |
| Native and compatibility | `nativeset` typed native factories plus sealed adapter | exact selected native ID or one of four closed compatibility IDs → connector | no configuration/name fallback | `dynamodb`, `mysql`, and `postgres` are native executor identities; only `bing-ads`, `faker`, `hubspot`, and `tally-prime` are compatibility rows |

### Caller/interface/artifact inventory

| Kind | Caller/owner file:symbol | Dependency | Expected effect | Migration/action |
| --- | --- | --- | --- | --- |
| production caller | `internal/app/app.go:Open,OpenForReverseExecution,DefaultIssueLabelTransportIdentity` | current project I/O then independent bundle registry | selection presently follows local I/O and can reconstruct a second registry | validate immutable construction before `os.Stat`; route presentation identity through index metadata and execution through one held construction |
| production caller | `internal/cli/cli.go:{rootManual,dynamicConnectorWithCommandSurface,runConnectors,appRegistry}`, router, docs, and skills consumers | full `bundleregistry.New` for list/help/preflight and later independent App open | metadata and executable paths can decode/construct different registries | list/discovery use compact metadata; detail loads one identified bundle; direct preflight consumes the App-held selection rather than constructing a second registry |
| test caller | `internal/app` explicit/open/process helpers; `internal/cli` opener/process helpers; `bundleregistry`, `commandrunner`, `manifestindex`, `manifeststore`, and `nativeset` tests | exported registry, hook, native, and opener seams | old globals hide caller drift | migrate every fallback-search hit to explicit typed construction or a sealed test seam; no global-builder test escape |
| generated owner/output | `cmd/connectorgen/gen.go`, 49 hook packages, `hooks/hookset/hookset_gen.go` | init-derived hook registrations and no index output | current generator cannot own selection or remove init authority | generate sorted compact index and explicit factory table; delete legacy `RegisterHooks`/`HooksFor` registrations and prove deterministic output |
| compatibility consumer | `internal/connectors/native/nativeset/factories.go` | three databases plus four legacy native connectors are currently name-selected together | a compatibility adapter could win a manifest selection | split explicit native database factories from the four-row adapter; validate disjoint inventories and invoke the adapter only through its closed executor identity |

### Mandatory impact matrix

| Lens | State | Exact evidence | Intended/possible effect | Required control or behavioral test |
| --- | --- | --- | --- | --- |
| Architecture and data flow | direct | Base `bundleregistry.New` eagerly calls `engine.LoadAll(defs.FS)`; CP04/CP05 have no production caller | a cache/factory-only refactor would retain a second eager route | generated index → held loader → App construction witness, with decode/constructor counters |
| Affected callers | direct | fallback searches found App, CLI root/help/list/direct/docs/skills, generated hookset, process helpers, and tests | partial migration leaves a process-global or second construction root | rerun exact caller inventory immediately before exported changes and migrate every production caller |
| Interfaces and configuration | direct | base default builder, `engine.RegisterHooks/HooksFor`, name-native selection, and unchecked `Entry.Executor` | config/name/order can select behavior or panic | closed executor/extension IDs and typed validation errors before every constructor |
| Generated, help, docs, skills, website artifacts | direct | `connectorgen gen` owns hook wiring; CLI derives manuals and skills from constructed connectors | generated output or public inventory can drift | byte-stable index/factory generation; metadata parity checks; mark website/content N/A only with zero inventory change evidence |
| Compatibility and migration | direct | `nativeset` contains three database and four legacy factories | protected databases can be conflated with fallback compatibility | exact 3+4 disjoint inventory witness; a selected manifest always wins and no connector-name fallback remains |
| Security and secret taint | direct | existing fixture/hostile-origin and missing-credential zero-request controls; App currently opens vault before registry | selection error could reach local secret or provider paths | factory/index validation spies assert zero vault/auth/request/constructor calls and errors expose no secret/config value |
| Concurrency, cancellation, and resource bounds | direct | CP05 mutex/LRU/single-flight store is test-only; builder/hook globals are mutable | duplicate flights, eviction, or global mutation can alter selection | digest-bound held handle, bounded count/bytes, release/cancellation/race proof, no new background goroutine |
| CLI and App reachability | direct | CLI preflights one registry then `App.Open` reconstructs another; root help iterates constructed connectors | differing generation/digest/executor or command boundary | same-entry App/direct-command witness; list zero-decode; implemented/missing-credential, partial, and unknown command controls |
| Provider and connector semantics | none_with_evidence | no source-lock, route, request encoder, pagination, response mapping, approval, or provider transport path is in the allowlist | factory fan-in could accidentally lose a hook semantic | exact 49-factory inventory parity plus existing production registry/preflight witnesses; no connector artifact diff |
| Focused behavioral tests and evidence | direct | factory RED controls exist; the recorded `TestDefaultRegistryDoesNotUseProcessGlobalBuilder` selector is stale, while `TestDefaultRegistryContainsOnlyBuiltins` and `TestOpenConstructsTheExplicitProductionRegistry` list exactly once | compile-only or zero-match proof can mask no-op construction | list every selector before execution; assert exact decode/constructor/I-O counts, typed causes, held identity, and command boundary outcome |

### P1–P5 prerequisite reconciliation

| ID | Concrete owner and change | Closed contract before the first GREEN edit | RED/GREEN observable |
| --- | --- | --- | --- |
| CP06-P1 — production bridge | `manifestindex` owns the generated compact index; `manifeststore` is corrected from raw test bytes to a digest-bound acquired-handle cache; `bundleregistry` owns the only loader and delegates to `engine.Load(defs.FS, entry.Connector)` | an entry names one logical generation/digest; list and lookup decode none; acquire verifies the exact entry and returns one held bundle; release makes only unheld entries evictable | 10,000-entry list/lookup has zero decodes; concurrent same-entry acquire decodes once; stale/mismatched digest, oversize, and cancellation fail without returning another bundle |
| CP06-P2 — generated selection ownership | `connectorgen gen` emits the sorted index from existing rendered execution JSON and the sealed executor/extension vocabulary; it never reads a source lock at runtime | every row is exactly one of `api_engine.v1`, `native_database/dynamodb.v1`, `native_database/mysql.v1`, `native_database/postgres.v1`, or one of the four explicit `closed_typed/*` compatibility IDs; extension IDs are explicit generated rows, never user/config/name/init selectors | duplicate/empty/multiple/unknown executor or extension entry returns a typed refusal before factory invocation; generation/check output is byte-stable |
| CP06-P3 — one construction root | `app.Open` owns executable construction; CLI metadata callers consume the compact index and direct command preflight uses the App-held selection instead of a second `bundleregistry.New` | no full registry construction for metadata listing; one App/direct command sees the same connector, generation, digest, and executor; manual detail may acquire only its named entry without constructing an executor | list performs zero full decode; direct command and App counter witness one identical held identity and preserve implemented/partial/unknown outcomes |
| CP06-P4 — explicit 49-hook fan-in | each current hook package exposes its explicit factory; `hookset` is the sorted generated closed inventory; delete `engine.RegisterHooks`, `engine.HooksFor`, blank-import activation, and the init registrations they require | no production package can register/overwrite a hook at import time; exact 49 IDs are validated once while building immutable extension factories | reversed factory order leaves selection unchanged; every generated ID constructs its same hook; duplicate/unknown IDs fail before the API factory; generator drift is detected |
| CP06-P5 — before-I/O ordering | `app.Open` first creates and validates the immutable catalog/factory construction, then performs project/vault/approval/state/coordinator I/O; connector construction occurs only after the selected entry is validated | malformed selection, duplicate factories/extensions, and incompatible compatibility inventory are typed errors before `os.Stat`, vault open, approval authority, state read, credential, or provider access | injected file/vault/auth/request/constructor counters remain zero for each invalid selection; valid App/direct command retains the existing credential/typed-block boundary |

### Preflight disposition

- RED oracle and exact selector: the preserved factory RED remains historical input only. The stale App selector was corrected by `go test -list` to `TestDefaultRegistryContainsOnlyBuiltins|TestOpenConstructsTheExplicitProductionRegistry`; each new P1–P5 behavioral selector must be listed before its individual RED execution and must assert decode/constructor/I-O or identity state, not compilation or exit status.
- Smallest GREEN path: implement the P1/P2 bridge and index as one same-issue prerequisite correction; then make the P3/P5 App/CLI construction cutover; then remove the P4 global hook path and split the native/compatibility inventory. Each is a red→green vertical slice; no connector mapping, source lock, or provider path changes.
- Verification commands: plan-only preflight completed with exact GitHub/base checks, GSD source resolution, contract check, fallback caller searches, scoped dependency listing, and test listing. After individual GREEN slices, run only the exact focused selector/race/vet/generator/commandrunner/CLI-App commands recorded in `TDD-LEDGER.md` and `VERIFICATION.md`; do not run a write-capable generator before its RED/implementation slice.
- Stop findings: none remain unknown. The preserved dirty code is excluded from the candidate until it is reconciled to this allowlist and independently red/green-proven; any surprise path, legacy global, source-lock read, second decoder, or pre-I/O violation returns this map to `BLOCKED_PRE_IMPLEMENTATION`.
- Disposition: READY — all ten rows are direct or `none_with_evidence`, and CP06-P1 through CP06-P5 now have a bounded same-issue owner, route, and observable proof. This authorizes only the planned first RED; it is not a CP06 production GREEN claim.

## CP06 retrofitted impact revalidation — 2026-09-03

- Authority: Firstmate inbox `109.msg` pauses all CP06 production edits pending an explicit `impact-ready` resolution. This record supersedes the earlier CP06 `READY` disposition as a production-edit authorization.
- Frozen identity: local `HEAD`, `origin/fm/cli-top100-declaration-batch-r1`, and PR #4294's API-reported head are `cb33f2ef2584c7307f3c88b869c474796698b0fb`; API base is `main`; `0b214b79eeb871238ce8454cd7b896e71e2746a7` remains an ancestor.
- Preserved local state: 119 tracked paths differ from that parent. The local cache tree and certification residue remain untracked, unread, unstaged, and outside this analysis.
- Tool evidence: CodeGraph has no repository index; LSP exposes no Go server. The fallback used focused source reads, symbol searches, `go test -list`, and scoped package tests. `scripts/gsd doctor` retains only the known missing `issue-122-rebootstrap.md`; all five lifecycle sources and `go run ./cmd/agentcontractgen check` resolved.
- Scope boundary: no source lock, rendered execution JSON, provider request, credential, database behavior, release artifact, or generic executor may change while this disposition is blocked.

### Actual changed-surface map

| Lens | State | Current evidence | Required minimal in-scope resolution |
| --- | --- | --- | --- |
| Architecture and data flow | blocked | `bundleregistry.Construction.BuildRegistry` loops `index.List()` and calls `store.Acquire` for every entry; `engine.Load` is therefore reached fleet-wide. | Replace fleet construction with an index-backed named acquisition path; list must decode zero bundles and named execution must acquire one selected entry. |
| Affected callers | blocked | `internal/cli:appRegistry` feeds root help, connector listing, dynamic help, docs, skills, and commands; `app.Open` independently calls `bundleregistry.NewRegistry`; `DefaultIssueLabelTransportIdentity` does so again. | Inventory and migrate every production caller to one construction instance or an index-only metadata surface; retain explicit test-only registry injection. |
| Interfaces and configuration | blocked | `manifestindex.Entry` contains connector, generation, digest, executor, and byte charge only; `BundleHandle` exposes only `*engine.Bundle`; `New()` still panics for presentation callers. | Generate the safe metadata and held identity required by list/lookup, validate generation/digest against the loaded execution generation, and return typed construction errors at executable boundaries. |
| Generated, help, docs, skills, website artifacts | blocked | `connectorgen.genManifestIndex` emits only identity/executor fields; its test checks one alpha row. The altered Atlas still names deleted `loadDefinitions` and eager `engine.LoadAll`; CLI docs/skills still call `appRegistry`. | Make generator, checked output, Atlas owner/proofs, and CLI-derived surfaces describe the same index and selection contract; prove check-mode byte stability before any generation write. |
| Compatibility and migration | blocked | `nativeset` splits three database and four compatibility executor IDs, but `BuildRegistry` still dispatches all entries eagerly and the registry package does not compile because its stale `loadDefinitions` test remains. | Prove a generated 3+4 disjoint inventory maps each selected ID to exactly one factory and remove only stale test assumptions; no connector-name runtime fallback. |
| Security and secret taint | blocked | `openWithRegistry` invokes its closure before `os.Stat`, but no invalid generated-selection path is injectable through `app.Open` and no vault/approval/auth/request counters cover it. | Add an observable invalid-selection witness that proves zero filesystem, vault, approval, auth, request, and constructor work after validation fails. |
| Concurrency, cancellation, and resource bounds | blocked | Raw `manifeststore.Store` has CP05 cancellation tests; production `BundleStore` has oversize, held-capacity, and shared-load tests only. It keys cache identity by connector and does not expose/verify digest or generation. | Give the production held-store path the count/byte, shared-cancellation, abandoned-flight retry, and exact-identity witnesses required by CP06-P1. |
| CLI and App reachability | blocked | `runMaybeConnectorCommandWithRegistry` now passes its preflight registry to `OpenWithRegistry`, but that registry was already fleet-built by `appRegistry`; metadata callers remain eager. | Prove list zero-decode, one named lookup/decode, and one shared App/direct-command selection while preserving implemented, partial, and unknown outcomes. |
| Provider and connector semantics | none_with_evidence | The preserved diff changes no connector source lock or rendered execution JSON; construction continues to delegate declared execution to `engine.Load`. | Keep this row closed only if later construction changes remain outside provider route/auth/pagination/response and protected database semantics. |
| Focused behavioral tests and evidence | blocked | Index/store race tests pass and `./internal/connectors/hooks/...` reports 49 passing packages, but `./internal/connectors/bundleregistry` fails to compile: `registry_test.go` still references deleted `loadDefinitions`. | Rebuild the exact selector inventory around observable list/decode/identity/pre-I/O behavior; no prior CP06 “green” claim is valid until the registry package and all P1–P5 witnesses pass. |

### CP06-P1 through CP06-P5 disposition

| Prerequisite | State | Minimal resolution before `impact-ready` |
| --- | --- | --- |
| CP06-P1 — production bridge | blocked | Make generated entries carry safe list metadata plus exact generation/digest identity; make the production store acquire and hold one verified selected bundle rather than eagerly decoding the fleet. |
| CP06-P2 — generated selection ownership | blocked | Validate one closed executor identity and all explicit extension IDs for every generated row against the complete factory inventory before any constructor; retain only generation-time sealed database/compatibility mappings. |
| CP06-P3 — one construction root | blocked | Replace production `appRegistry`/`app.Open`/identity duplicate construction with an explicit shared construction root that supports index-only metadata and named execution acquisition. |
| CP06-P4 — explicit hook fan-in | blocked | Preserve the completed removal of global hook registration, then add the exact 49-factory and duplicate/unknown rejection proof and repair stale registry/Atlas tests that still name the removed route. |
| CP06-P5 — before-I/O ordering | blocked | Extend the existing closure-before-`os.Stat` proof to malformed/duplicate selection errors and prove no filesystem, vault, approval, credential, request, or connector construction occurs first. |

- Disposition: `BLOCKED_PRE_IMPLEMENTATION: CP06-P1-production-bridge, CP06-P2-complete-selection-validation, CP06-P3-single-construction-root, CP06-P4-stale-registry-atlas-proofs, CP06-P5-zero-I-O-witness`.
- Next action: preserve all local code exactly and await Firstmate's explicit `impact-ready` resolution. Planning evidence only is authorized.

### CP06 rate-limit architecture preservation overlay

- Authority: Firstmate inbox `111.msg`; research record `/Users/karthiksivadas/karthik-agent-workspace/data/cli-rate-limit-architecture-preservation-r1/report.md`.
- Impact: direct. GitHub is the sole current declared API rate document (10 policies); PostgreSQL is explicitly `not_applicable`; every other current bundle remains absent or unknown. No rate policy, source lock, rendered rate declaration, provider resource, credential, or live-provider activity is in scope.
- RL-03: `rate_limits.json` remains optional execution JSON in the closed rendered file set. Changing or removing only that file must change the generated entry's byte charge and digest; connectorgen must not parse or validate rate policy semantics.
- RL-04: store identity is exactly `{connector,generation,digest}`. A connector-correct but generation/digest-mismatched load returns no handle and no factory execution. Count/byte accounting, flights, and held identities use that same key.
- RL-06: `api_engine.v1` receives the exact selected `engine.Bundle`, including `Bundle.RateLimits`; native and sealed compatibility construction consume that selected generation and may not reload `defs.FS`.
- RL-07: invalid index/factory/extension selection returns before filesystem, vault, approval, authentication, request, constructor, or provider work and exposes no protected value.
- RL-09: two same-scope selected GitHub constructions share the existing process budget; a different declared scope remains independently admissible. This reuses existing requester/rate coordination; it adds no rate service, daemon, policy layer, shared-coordination requirement, or provider work.

| Harness | Observable contract |
| --- | --- |
| A — `TestGeneratedManifestIndexDigestIncludesRateLimits` | A rate-file-only edit changes generated digest and byte charge; its closed set is detected. |
| B — `TestLazyConstructionSharesGitHubRateAdmission` | One generated GitHub selection crosses index → held bundle → `engine.New` → local fake 429/reset/admission; same scope gets one pre-reset send, a different scope remains admissible. |
| C — `TestBundleStoreRejectsGenerationOrDigestMismatch` | Generation/digest identity and held handle are exact; mismatched loads do not reach a factory. |

| CP06 prerequisite | Rate-preservation addition |
| --- | --- |
| CP06-P1 | Generated digest/byte charge includes rate files; store/cache/flight/handle identity is `{connector,generation,digest}` and remains held through executor/App lifetime. |
| CP06-P2 | Index validation rejects mismatched identity before any executor or extension factory. |
| CP06-P3 | App/CLI named construction shares the selected generation and existing process rate registry rather than constructing a private rate layer. |
| CP06-P4 | The generated GitHub hook factory remains selected through the index and uses the engine requester path; no import-time fallback or direct request path returns. |
| CP06-P5 | Invalid selection has a zero-I/O witness through the first project stat and all downstream vault/auth/request boundaries. |

- CP06 is not green until Harnesses A, B, and C pass with the existing focused rate requester and parking/resume proofs. CP11 owns atomic generation publication/pruning; CP06 does not add it.

- Rate-overlay allowlist extension: only the protected native and sealed compatibility constructor files under `internal/connectors/native/{dynamodb,mysql,postgres,bing-ads,faker,hubspot,tally-prime}/` may change, solely to consume the already-selected `engine.Bundle`. Their provider protocol, declared artifacts, lane behavior, and rate declarations remain unchanged.

### CP06 implementation status after impact-ready

- P1: complete locally — generated safe metadata/command summaries list without bundle decode; `BundleStore` is count/byte bounded, cancellation-safe, keyed by `{connector,generation,digest}`, returns canonical identity, and supports generation holds after cache release.
- P2/P4: complete locally — generated entries carry one closed executor and optional explicit hook extension ID; construction validates every executor/extension before a loader or constructor; generated hook fan-in contains 49 explicit factories; protected database and four compatibility adapters consume the selected bundle.
- P3/P5: complete locally — App opens without connector/transport fleet decode, lazy transport composition occurs only when a declared route needs it, direct command/root-help metadata reads use index projections, and invalid registry construction precedes the injected project-stat boundary.
- Rate overlay: Harnesses A, B, and C are implemented locally. API selection keeps GitHub's declared process-local rate bundle, same-scope selected constructions share admission after a local fake 429/reset, and a different scope remains independent. No rate declaration or source lock changed.
- Still required before any candidate checkpoint: finish the scoped smoke/build/docs/renderer checks, frozen self-review, and all delivery gates. Do not commit or push until that work is complete.

- Generated-output guardrail: `internal/connectors/boundary/{classify,ownership}.go` and its focused test now recognize `manifestindex/index_gen.go` as a literal generated shared index. This is required to keep provider literals out of handwritten generator/runtime code; the full connector-boundary scan is the proof.

## CP07 — GitHub manifest-selected API engine reference

## Task Delivery Header

- Issue: Refs #4425 — A1 manifest-selected executor registry; parent Refs #4325.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `843a32de5f927b1235cc00883fa0c5e0f5ea8c5b`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing PR #4294.
- Delivery: one ordinary fast-forward reference-proof checkpoint on the existing parent branch; no per-slice PR.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: prove GitHub's generated entry selects `api_engine.v1` plus only its explicit `hook/github.v1` extension, constructs `*engine.Connector` through the production registry, retains declared rate coordination, and cannot fall back to a native/compatibility connector.
- Verification: exact selector discovery, focused production-registry test, generated-index/render checks, existing GitHub rate admission proof, commandrunner preflight, local CLI help/inspect smoke, and frozen local self-review.

### Scope and TDD disposition

- No source lock, rendered connector artifact, provider route, credential, rate declaration, or provider call may change. CP07 consumes CP06's generated-index foundation; it does not migrate GitHub.
- Starting state is already structurally selected by CP06: `GeneratedEntries()` contains GitHub `api_engine.v1` and `hook/github.v1`. A new RED that claims the opposite would be fabricated. Record this as an explicit manual-TDD fallback, then add a production-registry witness whose assertion would fail if a native/compatibility fallback reappeared.
- Expected paths: the CP07 witness test plus these phase artifacts only, unless its RED reveals a direct CP06 invariant regression. Any other source path is a stop.

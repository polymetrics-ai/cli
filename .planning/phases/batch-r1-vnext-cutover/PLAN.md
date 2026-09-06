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

## CP08 — PostgreSQL manifest-selected native database reference

## Task Delivery Header

- Issue: Refs #4425 — A1 manifest-selected executor registry; parent Refs #4325.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `c267f6ccb6988c6d0132f264e963c6701b8134f1`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing PR #4294.
- Delivery: one ordinary fast-forward PostgreSQL reference-proof checkpoint, then the Firstmate-controlled coherent A1 review gate.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: prove PostgreSQL's generated entry selects exactly `native_database/postgres.v1` with no extension; the production registry returns the protected native adapter from the selected bundle, keeps `rate_limits.json` explicitly not applicable, and has no API/compatibility fallback.
- Verification: exact selector discovery, production-registry witness, native/nativeset suite, generated index/render check, implemented-command preflight, and local self-review.

### Scope and TDD disposition

- No PostgreSQL definition, source lock, rate declaration, database protocol, credential, container harness, provider action, or other connector may change. This is selection proof only.
- As with CP07, CP06 already generated the selected state. The manual-TDD fallback records no fabricated RED; the witness would fail on an API/compatibility fallback, an unexpected extension, or a false rate coordination claim.

## A1 consolidated correction wave — review disposition 114

- Review base/candidate: `9f96054ed85d6470306bc5b033e805b89b36c7c6..62e04650f44878da013a893bbeccc22c3b9b690c`; corrected parent is `346200f659e9ae22d0be42a83ee10f75d658f0b6`.
- Authority: Firstmate inbox `114.msg` and `data/cli-a1-62e-phase-review-r1/report.md`. One Terra-max correction wave only; no CP09, connector migration, provider/release work, rate declaration change, or per-micro review.
- A1-01: complete #4424's pure closed plan/resolver input and result contract. Add all seven mode mappings, retry/idempotency/receipt/acknowledgement/checkpoint axes, exact generation/artifact/executor/foundation/evidence digests, typed real-axis C3 classification, C4 foundation gaps, canonical serialization, exhaustive table matrix, and fuzz coverage. `synccontract` remains a dependency leaf.
- A1-02: add one shared immutable execution-identity helper used by generator and `engine.Load`; loaded bundles expose connector/generation/digest/byte charge. Store compares loaded identity before cache/handle/factory, rejects same-name mismatch with `ErrBundleIdentityMismatch`, and retains exact-key flights/cancellation/leases/rate behavior.
- A1-03: make `cli.Run` construct one explicit registry and thread it through root manual, dynamic help, connector preflight, normal App open, and reverse App open. Injection is an explicit test-only opener mode; variadic length must not decide construction ownership.
- Required retained invariants: RL-03/04/06/07/09; 49 generated hooks; 3+4 native/compat inventory; source-lock execution-only boundary; credential/approval/provider-I/O rejection; existing docs/help/manual parity.

| Correction | RED oracle | GREEN observable |
| --- | --- | --- |
| A1-01 | `incremental_upsert` with append and incomplete identity/durability axes is accepted. | Every legal seven-mode row resolves exactly once; every incompatible row is C3 with its true axis; digest bytes survive canonical JSON; fuzz finds no executable contradiction. |
| A1-02 | Index A accepts same-name loader result B and labels it identity A. | Loader B produces typed mismatch, zero handle/factory/rate resolver; loaded identity equals generated A before engine/native/compat construction. |
| A1-03 | Normal `cli.Run` gets preflight registry A then default `app.Open` registry B; `--help` re-resolves. | Production-equivalent normal/reverse/help routes observe one constructor/decode identity and preserve credential/approval boundaries. |

- Focused end gate: report-prescribed synccontract/syncplan/syncrun matrices+fuzz; identity/store/factory race tests; normal CLI identity/decode tests; GitHub/PostgreSQL references and rate admission; hooks/3+4; commandrunner; generated/Atlas/docs/help/smoke/boundary; one frozen self-review then one Firstmate-managed exact-SHA A1 re-review.

### A1 correction implementation status

- A1-01 complete locally: version-2 pure plan model has separate source/destination bindings, exactly one executor per role, all required digest bindings, seven canonical mode axes, and C3/C4 result discriminants. The resolver is deterministic and fuzzed; it remains a leaf with no runtime wiring.
- A1-02 complete locally: generator and engine share one execution-identity algorithm. Store accepts only identity-bearing loaded bundles that exactly match generated connector/generation/digest/charge; mismatch occurs before cache, handle, factory, or rate resolution.
- A1-03 complete locally: root-run construction has explicit production vs test override mode. The normal router, dynamic help, App, reverse App, connector docs, and skills reuse the one preflight registry. Skills generation traverses a lazy registry once rather than resolving the fleet twice.
- Local green gates are complete: closed-plan table/fuzz, identity/store/factory race, normal router ownership, rate/reference/hook/native witnesses, command preflight, renderer/definition/Atlas/docs/help/smoke/boundary, scoped vet/build, and frozen self-review. The known full-App/full-CLI/lint results are classified in `VERIFICATION.md`; none is repaired or reclassified by this wave.
- Pending only: candidate commit, then the Firstmate-managed final exact-SHA A1 re-review. CP09 remains prohibited.

## A1-04 entry-capacity correction — review disposition 118

- Review parent/candidate: `346200f659e9ae22d0be42a83ee10f75d658f0b6..701a0b45175f308400c938322fd1634a28efdaef`.
- Authority: Firstmate inbox `118.msg`. The independent review BLOCK found a concrete resource-bound violation. This is one narrow A1-02 correction, not CP09, a connector migration, a source-lock/render change, or a rate/CLI/factory redesign.
- Atlas disposition: **constrained_extension** of `definitions.bundle-loader.v1`. Update its supported guarantee and proof list in the same change because the existing BundleStore contract gains an in-flight entry-capacity invariant. No captain approval is needed for the declared seam; no new foundation is introduced.
- Owner and allowed production paths: `internal/connectors/manifeststore/bundle_store.go` only. The likely state is a count-reservation field paired with the existing byte reservation, reserved under `mu` when a distinct `bundleFlight` is installed and released exactly once in its terminal loader path. Same-key waiters consume no additional count capacity.
- Red: add a deterministic barrier-based two-identity test using `Limits{Entries: 1, Bytes: 2}`. While identity A's loader is held, identity B must fail with `ErrBundleCapacity` and the locked snapshot must show retained plus reserved entries at most one. After A's caller cancels but the loader remains held, B must still fail. Once A terminally exits, B retry must load and retain one entry; retrying A after B is releasable must preserve the same bound.
- Green: reserve one entry slot atomically with every newly installed distinct flight; release that slot exactly once with the matching byte reservation, including canceled and failed loaders. Preserve eviction, handle refs, generation leases, exact identity validation, and same-key flight sharing.
- Refactor/review: keep the reservation accounting local to BundleStore; do not export diagnostics, add goroutines, alter loader cancellation semantics, or touch connector, rate, generation/digest, factory, App, or CLI paths. Read the post-change diff against the exact candidate and rerun the focus set before reporting the new SHA for Firstmate's fresh independent review.

| Acceptance criterion | RED observable | GREEN observable |
| --- | --- | --- |
| Distinct in-flight loads honor `Limits.Entries` | Two one-byte identities begin concurrently under `Entries: 1, Bytes: 2`. | B is refused before its loader starts while A owns the only reserved entry slot. |
| Cancellation cannot create an overlapping distinct flight | Canceling A removes its waiter while its blocked loader remains live, and B can start. | B remains refused until A's loader terminally releases its reservation. |
| Completion and retry restore exactly one slot | A completion leaves a leaked reservation or B retry inserts beyond the cap. | B retry succeeds after A finishes, and every snapshot has cached plus reserved entries no greater than one. |
| Existing A1 contracts stay intact | Any focused retained witness changes. | Identity, rate, construction, connector, and CLI focused tests remain green without production-surface changes. |

- Planned verification: exact RED/GREEN selector in `./internal/connectors/manifeststore`; package race suite; focused `manifestindex`/`bundleregistry` construction and identity tests; `go vet ./internal/connectors/manifeststore ./internal/connectors/bundleregistry`; Atlas JSON/proof validation; `git diff --check`; read-only exact-candidate self-review. No broad parent push or CP09 start.

### A1-04 implementation status

- RED was observed exactly as the review predicted: with A barrier-held, B loaded under `Entries: 1, Bytes: 2`, producing a retained-plus-reserved count of two.
- GREEN reserves `reservedEntries` under the existing store mutex at distinct-flight admission, includes it in the entry-capacity eviction/refusal loop, and releases it in the matching terminal load path beside the byte charge. Last-waiter cancellation still cancels the loader but cannot release capacity before that terminal path.
- The exact barrier regression, full manifeststore race suite, affected identity/index/registry suites, scoped vet, Atlas JSON/selector proof, CLI build, and agent-contract check are green. No source lock, rendered execution JSON, rate, connector, factory, App, or CLI behavior changed.
- Independent exact-SHA review PASS: Firstmate instruction `120.msg` accepts `c761e7e6f2d042c7560ab0c520dc9aa182110e6e` with zero blockers and closes A1-04. Preserve that candidate; normal non-force publication to `fm/cli-top100-declaration-batch-r1` is authorized. CP09 starts only after this parent update, without reset or rebase to `main`.

## CP09 — N2 strict source parsing and canonical graph

## Task Delivery Header

- Issue: Refs #4426 — N2 semantic source-lock admission; parent Refs #4325.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `988dd16c3d206a28d3e7b16f8a0d805c4163f7ca`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing draft PR #4294.
- Delivery: one ordinary non-force strict-parser green checkpoint published to the declared parent after focused verification and code review; CP10 semantic admission remains separate and unstarted.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: extend the sole schema-4 source-lock decoder into one strict typed canonical source graph. It must reject trailing roots, duplicate members, unknown execution fields, wrong schema roles, invalid encoders, alias collisions, and missing structural bindings before any rendered-file replacement; preserve every valid source identity and deterministic render bytes under irrelevant input ordering.
- Verification: parser/graph RED-GREEN tests with no-write sentinels; deterministic equivalent-lock render comparison; affected `cmd/connectorgen` suite, source-lock validation/check-only render, Atlas proof, scoped vet/build, `agentcontractgen check`, `git diff --check`, and a goal-backward code review.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Non-canonical and structurally invalid source locks are refused before output replacement | live | Table-driven decoder/render tests preserve a sentinel generated file after each rejected trailing, duplicate, unknown, role, encoder, alias, or binding case. |
| A valid lock becomes one typed canonical graph without losing source identity | live | Equivalent valid input orderings yield canonical descriptors and rendered closed sets with equal bytes and retain every operation/source identifier. |
| CP09 keeps the execution boundary closed | live | Check-only render/validation uses the source decoder without provider I/O or a runtime executor; no source lock or rendered execution artifact is published by this checkpoint. |

### CP09 scope, Atlas, and lifecycle disposition

- Authority: Firstmate instruction `120.msg`; A1-04 exact review PASS is published at the declared parent head above. CP09 is the first half of #4426 only; CP10 owns source-to-execution/resolver/preflight semantic admission.
- Foundation Atlas classification: **constrained_extension** of `authoring.source-lock-vnext.v1`, through its declared canonical-descriptor and rendering seam. Its owner, contract guarantee, constraints, and proof list must be maintained in this change if the canonical graph contract changes; no new shared runtime foundation, runtime source-lock reader, connector branch, executor, provider source fetch/re-pin, or generated-file publication is allowed.
- GSD lifecycle: `scripts/gsd doctor` has the recorded repository blocker for missing `.gsd/prompts/issue-122-rebootstrap.md`; resolve the required commands and run the documented inline/manual fallback in order (`discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, `code-review`). No compatible isolated worker is authorized by the canonical contract.
- Required skills: `go-engineering`, `tdd`, and `connector-lane-build-order` are loaded. The repository-mandated `golang-how-to` skill is unavailable in this runtime; record that fact rather than claiming its use. CodeGraph has no repository index and Go LSP is unavailable.

### CP09 implementation status

- The existing `decodeStrictJSON` remains the sole decoder. Canonicalization now retains an authoring-only typed graph, validates typed execution nodes and source identity shapes, verifies structural role placement, enforces normalized command alias uniqueness, and invokes the existing loader against a read-only in-memory rendered view before a renderer can replace output.
- Structural-only boundary: request and response references may legitimately belong to a direct operation; only record references require the stream's exact schema. CP09 does not compare provider citation routes or invent semantic source-to-execution joins; those remain CP10 work.
- Atlas and canonical source documentation are updated in the same change. The complete `cmd/connectorgen` package, all-lock graph corpus, three CLI check-only renders, 553-definition validation, catalog/docs checks, vet, CLI build, and agent-contract check are green; exact commands/results are recorded in `TDD-LEDGER.md` and `VERIFICATION.md`.
- Publication: committed as `d11277378abe556323226e3f6998ce3caf6033dc` (`feat(connectorgen): canonicalize source locks strictly`) and normally pushed to `origin/fm/cli-top100-declaration-batch-r1`. PR #4294 API read-back confirms `base=main`, `head=fm/cli-top100-declaration-batch-r1`, `draft=true`, and the same head SHA. No reset, rebase, source-lock artifact publication, or CP10 work occurred.

## CP10 — N2 semantic source-execution admission

## Task Delivery Header

- Issue: Refs #4426 — N2 semantic source-lock admission; parent Refs #4325.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `85c28e70e4c8f811ea342a1f1054e09759cde1c1`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing draft PR #4294.
- Delivery: one ordinary non-force semantic-admission green checkpoint published to the declared parent after focused verification and inline/manual code review; the independent review remains at the coherent phase boundary.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: turn the CP09 canonical graph into a complete in-memory staged generation with exact source-to-schema, stream, action, operation, command, manifest/index, rate-file, and supplied sync/Atlas admission facts. Reuse the real engine loader, commandrunner preflight, manifest index, native selection inventory, and syncplan resolver; refuse broken joins with the immutable source operation ID and JSON field path before any output replacement.
- Verification: RED-GREEN semantic-admission tests with no-write sentinels; deterministic staged-set and rate-identity controls; exact GraphQL same-route operation-binding control; affected `cmd/connectorgen` suite, in-memory loader/preflight/resolver controls, Atlas proof, scoped vet/build, `agentcontractgen check`, `git diff --check`, and a goal-backward code review.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A source node has one exact runtime schema, stream, action, operation, and command binding | live | A one-fact-at-a-time mutation returns the source operation ID plus JSON field path and leaves the output sentinel unchanged; valid bindings retain an ordered provenance row for every projected node. |
| Source command identity cannot drift across same-route GraphQL operations | live | Two fixed GraphQL operations share `/graphql`; swapping one command's operation ID is refused at that command field rather than admitted by route equality. |
| Staged execution/manifest/index identity is closed and deterministic | live | Two admissions of the same graph yield equal file bytes, digest/byte charge, manifest entry, index lookup, and provenance; adding/removing/changing only `rate_limits.json` changes/restores the closed identity. |
| Runtime and synchronization admission reuse existing authority | live | Every implemented command is passed through `engine.Load` and `commandrunner.Preflight`; supplied source/destination/Atlas facts are accepted only when the real `syncplan.Resolve` result is executable. |

### CP10 scope, Atlas, and lifecycle disposition

- Authority: Firstmate instruction `122.msg` accepts CP09 and authorizes immediate serial CP10 in #4426/N2. It explicitly preserves CP09 boundaries: no connector work, schema-4 lock authoring, rendered-artifact/global-output publication, provider I/O, credential path, publication algorithm, or source migration.
- Foundation Atlas classification: **constrained_extension** of `authoring.source-lock-vnext.v1` through its declared canonical-descriptor, staged-generation, and proof seams. Reuse real runtime admission APIs; do not add a shared runtime foundation, connector-specific fallback, opaque wrapper, second reader, or copied encoder/resolver rule.
- GSD lifecycle: `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`. The generated `discuss-phase` prompt is recorded and the remaining prompts will be fulfilled inline because the established phase lacks the adapter's required ROADMAP and the known doctor blocker remains `.gsd/prompts/issue-122-rebootstrap.md`. No compatible isolated worker/reviewer is authorized.
- Required skills: `go-engineering`, `tdd`, `connector-lane-build-order`, and `connector-migration-exact-sha-review` are loaded. Repository-mandated `golang-how-to` is unavailable; it is not claimed. CodeGraph has no repository index and Go LSP is unavailable.

### CP10 plan

1. Retain the one strict decoder and CP09 graph. Add only an in-memory semantic staging value: rendered bytes, loaded bundle identity, manifest/index input, deterministic source-to-execution provenance, and optional supplied sync/Atlas admission facts.
2. Bind source operation IDs to loaded execution identities, not route text: schemas by exact registry path, streams/actions/operations by their declared names, commands by normalized path and exact stream/write/operation target. Preserve arbitrary provider source facts raw; validate only their declared bindings without inventing or deleting facts.
3. Use `engine.Load`, an in-memory `engine.New`, `commandrunner.Preflight`, `manifestindex.New`, `nativeset.ManifestSelections`, and `syncplan.Resolve` rather than local approximations. A source lock with only one endpoint must not synthesize a saved-sync counterpart; sync resolution receives explicit supplied endpoint and Atlas facts.
4. Keep the staged set in memory. CP11 alone owns filesystem staging, activation, recovery, and global publication.

### CP10 implementation status

- `vNextStagedGeneration` retains the exact rendered byte set, loader identity, selected manifest/index entry, deterministic provenance, and only explicitly supplied sync results. It has no filesystem staging or activation path.
- Semantic admission now uses the real in-memory `engine.Load`/`engine.New`, `engine.ResolveImplementedCommandBinding`, `commandrunner.Preflight`, shared generator executor selection, `manifestindex.New`, and `syncplan.Resolve`. It does not read a global generated manifest, connector credential, provider, or transport.
- Exact parent-operation binding rejects a source command that crosses to a same-route GraphQL operation. Known source facts validate only their declared GraphQL/transport or unsupported-target counterpart; unknown fields stay raw. Declared `rate_limits.json` is accepted solely through the runtime loader and changes the staged identity.
- Focused RED/GREEN and complete package results are recorded in `TDD-LEDGER.md`; pending cleanup verification, inline/manual `verify-work`, and code review are recorded only after observation.

## N2 boundary-review correction — CP09/CP10

## Task Delivery Header

- Issue: Refs #4426 — N2 exact-SHA admission correction; parent Refs #4325.
- Base branch: `fm/cli-top100-declaration-batch-r1` at `56ec3d9d7dc1d726203b0ef0c03ddec3209b8dde`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing draft PR #4294.
- Delivery: one ordinary non-force correction checkpoint published to the declared parent, with the exact review-block contracts proved and an inline/manual self-review recorded; a new Firstmate-managed exact-SHA review then controls CP11 authorization.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: close only the N2 review's three authoring-admission defects: bind every admitted request/response schema reference to the loaded runtime schema that consumes it; derive a complete in-memory production manifest entry including a selected hook extension; and canonicalize provenance independently of irrelevant authored operation ordering while retaining authored positions for diagnostics.
- Verification: one RED-GREEN no-write schema-swap test per request/response role; a GitHub local complete-manifest/hook/preflight witness; reordered-operation byte/digest/provenance equality plus authored diagnostic location; full affected generator suite, check-only reference renders, definition validation, Atlas/docs checks, scoped static gates, and goal-backward self-review.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every admitted request/response reference joins its effective runtime schema | live | Swapping either role to a different existing shared schema makes `lock-render` reject at the immutable source operation's authored `/schema_refs/<role>` path while leaving a generated-file sentinel byte-identical. |
| Staged manifest matches production hook selection | live | GitHub's source lock stages an entry exactly equal to its closed generated manifest entry, including `hook/github.v1`; construction and preflight use that selected local hook without provider or credential I/O. |
| Provenance is independent of irrelevant source operation ordering | live | Reversing same-rank source operations retains equal rendered bytes, identity, complete provenance, manifest, and index, while a malformed reordered input still names its authored source-array pointer in the error. |

### N2 correction scope, Atlas, and lifecycle disposition

- Authority: Firstmate instruction `124.msg` accepts the independent N2 boundary-review BLOCK at `56ec3d9d7dc1d726203b0ef0c03ddec3209b8dde` as three bounded CP09/CP10 corrections. CP11 remains prohibited until the follow-up exact-SHA review returns and Firstmate explicitly authorizes it.
- Foundation Atlas classification: **constrained_extension** of `authoring.source-lock-vnext.v1`. Reuse the current `engine.Load`, engine connector, `hookset.Factories`, native selection inventory, manifest index, commandrunner preflight, and sync resolver. No runtime foundation, connector-specific branch, second selection source, source-lock reader, or generated global manifest publication is authorized.
- GSD/manual fallback: `scripts/gsd doctor` reports only the established missing `.gsd/prompts/issue-122-rebootstrap.md`; all five lifecycle command sources and `discuss-phase`/`plan-phase --tdd` prompts resolved. The custom phase has no adapter ROADMAP, and no compatible isolated worker/reviewer is authorized, so the equivalent TDD, verification, and review procedures are recorded inline.
- Required skills: `go-engineering` (fundamentals and production), `tdd`, `diagnose`, `connector-lane-build-order`, and `connector-migration-exact-sha-review` are loaded. Required `golang-how-to` and `golang-testing` are unavailable in this runtime and are not claimed; CodeGraph has no repository index and Go LSP is unavailable.

### N2 correction plan

1. Add one real runtime-schema counterpart for every admitted schema role. A request reference must equal the loaded write record schema or a direct operation's declared runtime request schema; a response reference must equal the loaded stream schema. A role with no actual consumer is refused rather than recorded as provenance-only.
2. Model production selection once from the closed hook factory and native executor inventories. Use the selected hook to construct the in-memory engine connector and record its exact extension in the staged manifest/index. Do not read or synthesize a global manifest.
3. Retain each source operation's authored index exclusively for errors. Assign canonical order after source-ID sorting and use it for every staged provenance field path, then prove reordered inputs have identical staged provenance.
4. Keep all work in authoring-time memory and preserve `runLockRender`'s no-write-before-admission boundary. CP11 continues to own physical staging, activation, recovery, and publication.

## CP11 — B1 transactional connector-generation publication

## Task Delivery Header

- Issue: Refs #4427 — B1 transactional connector-generation publication; parent Refs #4325.
- Authorization / exact base: Firstmate instruction `125.msg` accepts the fresh N2 correction review PASS at immutable `36e4d980de0d51d92fe74a68306845643596a6cb`. CP11 starts only from that head; CP12 is not started.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing draft PR #4294.
- Delivery: one ordinary non-force CP11 checkpoint pushed only to `origin/fm/cli-top100-declaration-batch-r1`, then pause unchanged for a new Firstmate exact-SHA review. No per-slice PR, rebase, merge, or release action.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: publish one already semantically admitted connector generation as a closed, connector-local staged set. The set contains execution bytes plus manifest, provenance, Atlas-reference, compact-index, proof, and integrity metadata; it is validated, fsynced, activated through one `CURRENT` pointer, recovered through a durable journal, and pruned only when no generation handle holds it.
- Verification: RED/GREEN fault-state, stale-optional-file, reader/writer, writer/writer, digest, journal-recovery, and held-generation prune tests using temporary local roots and barriers only; scoped generator/static checks; read-only self-review; normal push and PR API read-back.

### CP11 SHA-bound ten-row impact map

- Frozen identity: local `HEAD` is `36e4d980de0d51d92fe74a68306845643596a6cb`, the sole authorized CP11 start point. The retained checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7` remains historical ancestry, not an alternate implementation base.
- Authority: `125.msg` explicitly authorizes #4427/B1 only after the N2 exact-SHA PASS. It forbids CP12, cross-connector transactions, provider or credential I/O, actual connector/source-lock materialization, and deletion of author-owned files.
- Foundation classification: constrained extension of `authoring.source-lock-vnext.v1`. CP10's `vNextStagedGeneration` is the sole semantic input; publication must not add another source-lock decoder, renderer, executor selection, or runtime request path.

| Lens | State | Exact evidence | Intended/possible effect | Required control or behavioral test |
| --- | --- | --- | --- | --- |
| Architecture and data flow | direct | `cmd/connectorgen/vnext_lock_cli.go:runLockRender` currently replaces rendered files independently; CP10 exposes one complete in-memory `vNextStagedGeneration` in `vnext_admission.go`. | Per-file replacement can expose an old index beside a new optional rate file or bundle. | One connector-local publisher stages the CP10 closed set under a same-filesystem generation root, validates it from disk, writes integrity metadata, then swaps only durable `CURRENT`; a generation resolver reads exactly that pointer. |
| Affected callers | direct | `runLockRender`, its temp-root tests, `Makefile:connectorgen-vnext-locks`, `scripts/tests/connector-canon.sh`, and canonical authoring docs name the current render/check contract. | A partial command/doc/check migration could retain flat per-file publication as an executable route. | Route generation publication and `--check` through the same publisher/checker; update only authoring command checks and docs. `defs.FS`, `bundleregistry`, App, CLI runtime routing, and CP12 projection remain unchanged in B1. |
| Interfaces and configuration | direct | `parseVNextLockArgs` exposes connector, `--defs`, and `--check`; `manifestidentity.ForFS` already defines the closed execution digest. | An unsafe connector/file/generation name, symlink, cross-device stage, or ambiguous pointer can escape the closed set. | Strict connector-relative artifact paths and generation IDs; connector-scoped lock outside the generated set; same-parent staging; typed `CURRENT`, journal, integrity, and held-generation handle APIs; reject malformed pointers, nonregular files, symlinks, duplicate/stale members, and digest mismatch before activation or use. |
| Generated, help, docs, skills, website artifacts | direct | `SOURCE-LOCK-VNEXT.md`, `INDEX.md`, `IMPLEMENTATION-PROCEDURE.md`, `REMOTE-REPRODUCIBILITY.md`, Foundation Atlas revision 33, Make target, and connector-canon script describe byte checks of flat outputs. | Documentation could claim atomic closed publication while scripts still check individual files, or expose a stale operator recipe. | Update the source-lock canon, Atlas owner/proof contract, and authoring verification references to the new staged-generation check. Public PM help, manuals, website docs, and skills are N/A unless command syntax/output changes; record the inspection result. |
| Compatibility and migration | direct | Committed reference locks and flat execution JSON are currently a deterministic in-memory parity corpus; instruction `125.msg` forbids actual connector/source-lock materialization. | Moving existing connector artifacts or accepting their flat layout as a second active-generation reader would exceed B1 and reintroduce a fallback. | Publish only hermetic temporary test roots in this checkpoint. Preserve author-owned `source.lock.json` and checked-in execution data untouched; the new resolver accepts only `CURRENT`-named staged generations and has no flat fallback. |
| Security and secret taint | direct | Source locks are authoring-only; current authoring path has no credential or provider call; generated roots may contain attacker-controlled filesystem names. | Unsafe paths or error strings could access outside the connector root or expose source/config contents. | Use fixed internal filenames, reject traversal/symlinks/nonregular files, never serialize secret/config values in journal/error output, and prove semantic/stage failures occur before any provider, credential, or source materialization action. |
| Concurrency, cancellation, and resource bounds | direct | Current writer has no connector lock, generation lease, recovery journal, or directory durability protocol; `manifeststore.GenerationLease` documents the later executor hold concept. | Concurrent writers can interleave and pruning can remove an old generation held by a reader. | Connector-scoped advisory lock serializes writers; readers hold an OS-backed generation lease before releasing the shared pointer lock; pruning skips held generations. Barrier tests assert old-complete or new-complete reads only, no sleeps, unbounded retry, cross-connector lock, or background worker. |
| CLI and App reachability | none_with_evidence | `defs.FS` embeds the immutable flat execution snapshot; `bundleregistry.NewConstruction` loads only that embedded snapshot. `125.msg` forbids actual connector/source-lock materialization. | Repointing App/CLI to an unmaterialized generation tree would create a second runtime route or alter connector behavior. | Limit B1 runtime proof to the new local generation resolver over temporary publication roots. Do not modify App, CLI executable routing, embedded defs, connector behavior, help surface, or credentials; later authorized work must integrate any real materialized generation through one route. |
| Provider and connector semantics | none_with_evidence | CP11 consumes only CP10's already-admitted byte set and metadata; its allowed source has no HTTP/requester/credential path. | A semantic validator could accidentally re-render source locks, select provider origins, or invoke connector transport. | Validate staged bytes with the existing loader/admission facts only; fixture tests use no provider URL, credential, network, or database. No connector lock or rendered execution artifact in the repository is written. |
| Focused behavioral tests and evidence | direct | #4427 requires RED stale-optional, per-file-mix, writer-interleave, and reader index/bundle mismatch controls; CP10 already supplies deterministic in-memory admission fixtures. | Exit-status-only checks can hide a torn active set or unsafe prune. | First create failing temp-root state-machine tests, then prove complete old/new reads, closed file-set/digest equality, every journal/durability cut point recovery, failed stage/active validation rollback, two-writer serialization, and an old handle blocking prune. Run selector listing, package/race tests, scoped vet/build, docs/Atlas/contract checks, and goal-backward self-review. |

### CP11 impact-ready disposition

- Tool/lifecycle evidence: CodeGraph is unavailable because this repository has no `.codegraph` index; Go LSP is unavailable. Targeted source reads cover the current writer, CP10 stage, manifest identity/index/store, embedded runtime boundary, authoring checks, and issue #4427. `scripts/gsd doctor` reports only the established missing `.gsd/prompts/issue-122-rebootstrap.md`; all five lifecycle sources plus `discuss-phase` and `plan-phase --tdd` prompts resolved. The custom phase has no adapter ROADMAP or compatible isolated worker, so the required GSD/TDD execution, verification, and review evidence is recorded inline.
- Required skill route: `go-engineering` fundamentals/production/advanced, `tdd`, `connector-lane-build-order`, and `connector-migration-exact-sha-review` are loaded. Repository-required `golang-how-to` is unavailable and is not claimed; `go-engineering` is the documented replacement for the bounded filesystem, concurrency, error, and test work.
- B1 implementation allowlist: new connector-generation publication/resolution code and tests under `cmd/connectorgen`, the lock-render command contract/tests, exact authoring docs/check callers, the `authoring.source-lock-vnext.v1` Atlas record, phase evidence, and no other production path unless the first RED proves an unavoidable in-allowlist owner omission. No `internal/connectors/defs`, App, CLI runtime, connector lock, source lock, or rendered execution JSON changes are authorized.
- READY — all ten lenses are direct or `none_with_evidence`; the only active writer path has a bounded owner, artifact layout, recovery state machine, and observable RED/GREEN proof. This authorizes the first CP11 RED test only.

### CP11 implementation design and inline GSD execution

- Publication root: only `<defs>/<connector>/generations/` is publisher-owned.
  A candidate writes below a direct `.stage-*` child, includes the admitted
  execution bytes plus manifest/provenance/Atlas/index/proof metadata and
  publisher-owned `integrity.json`/`.lease`, and is renamed to a
  digest-addressed sibling only after staged semantic validation and fsync.
  `source.lock.json`, existing flat execution artifacts, `defs.FS`, and every
  App/CLI runtime path remain outside the writer.
- Selection/recovery: typed `CURRENT` binds `{generation, integrity_digest}`.
  Typed `JOURNAL` records `{old,new,state}` before selection; current and
  journal replacement fsync their parent. Recovery accepts only a complete
  validated old or new pointer, restores old on failed activation, and retains
  malformed/unowned paths rather than deleting them.
- Reader/pruning: `Open` validates `CURRENT`, acquires a shared generation
  lease while holding the connector lock, then returns a pointer-bound handle.
  Pruning is connector-local and requires both an integrity-validated
  publisher-owned closed tree and a nonblocking exclusive lease; it cannot
  remove a held or unowned generation. `Check` uses only a shared existing lock
  and never recovers, prunes, creates, or rewrites publication state.
- No broad transaction or materialization: locks never cross connector roots;
  staging consumes the already-admitted CP10 value and uses no credential,
  provider, transport, database, source-lock reparse, actual corpus render, or
  runtime registration path.
- GSD/manual fallback: `scripts/gsd doctor` reports the known missing
  `issue-122-rebootstrap.md`; `sources` resolved all five mandatory commands
  and generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`,
  `verify-work`, and `code-review` prompts. The custom phase has no adapter
  `ROADMAP.md` and no compatible isolated GSD worker/reviewer, so the required
  discussion, plan, execution, verification, and review are performed inline
  and evidenced in this phase directory. This is a documented runtime fallback,
  not a lifecycle waiver.
- Skills: loaded `go-engineering`, `tdd`,
  `connector-lane-build-order`, `connector-migration-exact-sha-review`,
  `frontend-design`, and `vercel-react-best-practices`. The required
  `golang-how-to`, Go-specialist skill names, and `web-design-guidelines` are
  unavailable in this runtime; they are not claimed. The website integration
  and CLI parity references were read; no React/UI implementation or PM CLI
  surface is changed.

## CP11 exact-review correction — Firstmate instruction 126

## Task Delivery Header

- Issue: Refs #4427 — transactional connector-generation publication correction; parent Refs #4325.
- Base branch: `main` (existing draft PR #4294 base, API-verified again after delivery).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing draft PR #4294.
- Delivery: one normal non-force correction push to
  `fm/cli-top100-declaration-batch-r1`, PR #4294 base/head/SHA API read-back,
  then no branch change until Firstmate's next exact-SHA review disposition.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: correct the five CP11 exact-review blockers from immutable
  `c4f0bc3728dda318ea3d01f78de7aa299b6135cb`: descriptor-relative
  no-follow confinement, durable stage ownership, content-address validation,
  strict bounded control documents, and cancellable lock acquisition.
- Verification: one executable RED/GREEN control per finding; race/fault and
  full changed-package checks; canon/Atlas/static checks; read-only local
  review; normal push and PR API base/head/SHA read-back.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Connector root and publication lock cannot escape `--defs` | fake | A hermetic temporary definitions root is required because instruction 126 forbids materializing a real connector. Symlinked connector-root and lock tests assert all operations refuse and external sentinels remain byte-identical. |
| Only durably marked publisher stages are removed | fake | A temporary `.stage-author-owned` directory proves `Recover`, `Publish`, `Open`, and `Prune` preserve its sentinel and refuse ownership; a publisher-created marker permits recovery cleanup. |
| Every selected generation is its actual closed-tree content address | fake | A temporary copied/renamed self-consistent tree, unexpected empty directory, and nonempty lease each make `Check`/`Open` refuse before a reader observes it. |
| Every control document is bounded, no-follow, and unambiguous | fake | Temporary malformed `CURRENT`, journal, integrity root, and integrity-file payloads assert deterministic refusal, unchanged durable state, and no external-target read. |
| CLI/publisher lock waits honor cancellation | fake | A second descriptor holds the actual temporary connector lock; each context-bearing operation returns its cancellation error with unchanged pointer/journal, releases its acquisition handle, and succeeds after release. |

### Correction design and impact disposition

- Authority/base: instruction `126.msg` accepts the exact-review BLOCK and
  authorizes only these five corrections from
  `c4f0bc3728dda318ea3d01f78de7aa299b6135cb`. CP12 remains prohibited.
- Foundation Atlas classification: **constrained_extension** of
  `authoring.source-lock-vnext.v1`. The existing publisher remains the sole
  authoring owner; no runtime, connector, source-lock, renderer, or transport
  route is added.
- Confinement: retain a descriptor for the supplied definitions root and open
  connector/generation subroots relative to it. Reject connector-root and lock
  symlinks before access; source-lock and publication-control reads plus
  control-file replacement/removal stay beneath those rooted descriptors.
  Stage cleanup first proves a real non-symlinked directory and strict marker;
  `runLockRender` reads its author-owned source lock through the confined
  connector root.
- Stage ownership: write, sync, and retain a typed private marker immediately
  after a random direct-child stage directory is created. Recovery/prune
  removes only a regular directory whose strict marker binds connector,
  generation, and stage identity; any unknown, malformed, or symlinked stage
  is preserved and makes the operation fail closed.
- Generation integrity: stream every validated artifact through the same
  length-framed content-address calculation used for publication. Require its
  result to equal directory, integrity, and pointer generation values; derive
  the exact permitted directory set from artifact paths, and require an empty
  regular lease.
- Control and cancellation: use one rooted no-follow bounded reader plus
  `decodeStrictJSON` duplicate rejection for all typed control documents.
  Thread `context.Context` from the CLI signal boundary through every
  publisher operation and use a nonblocking advisory-lock retry that returns
  cancellation without creating, replacing, pruning, or leaking state.
- Scope: preserve all connector/source locks, flat execution corpus, `defs.FS`,
  App/CLI runtime routing, provider/credential/database access, certification,
  cross-connector transactions, and author-owned files. No PM command syntax,
  help, manual, or website surface changes are expected; authoring canon/Atlas
  claims change only where these behavior guarantees become true.
- Lifecycle/skills: `scripts/gsd doctor` has only the established missing
  `issue-122-rebootstrap.md`; all five command sources and generated
  discuss/plan/execute/verify/review prompts resolved. The phase has no
  compatible isolated GSD worker/reviewer, so GSD/TDD/review executes inline
  and is recorded here. Loaded `go-engineering` (advanced/production),
  `tdd`, `connector-lane-build-order`,
  `connector-migration-exact-sha-review`, `frontend-design`, and
  `vercel-react-best-practices`; required unavailable names
  `golang-how-to` and `web-design-guidelines` are not claimed.

## CP11 correction-review repairs — Firstmate instruction 127

## Task Delivery Header

- Issue: Refs #4427 — transactional connector-generation publication correction; parent Refs #4325.
- Base branch: `main` (existing draft PR #4294 base; read back through GitHub API after delivery).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through draft PR #4294.
- Delivery: one normal non-force correction push from
  `fm/cli-batch1-vnext-cutover-r2` to
  `fm/cli-top100-declaration-batch-r1`, API-confirmed PR base/head/SHA, then
  a frozen candidate awaiting a fresh Firstmate exact-SHA CP11 review.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: from immutable correction parent
  `4fedb3875cbe7071799aed0e9b6ce1e34257f95e`, repair the five review blockers:
  whole-operation descriptor confinement, component-boundary scanner matching,
  post-acquisition cancellation, prepared-journal-before-rename recovery, and
  complete Atlas guarantee-to-positive/negative-proof mapping.
- Verification: truthful RED/GREEN per repair; deterministic root-replacement,
  journal-cut, cancellation-acquisition, and scanner tests; race/full changed
  package checks; boundary/canon/Atlas/static checks; inline manual GSD review;
  normal push and PR API read-back.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Publication mutation stays in the original connector inode after pathname replacement | fake | A hermetic temporary definitions root is necessary because instruction 127 forbids checked-in connector materialization. A hook replaces `<defs>/acme` with an external sentinel tree while the publisher holds descriptors; every destructive path must leave the sentinel byte-for-byte intact and mutate/recover only the moved original inode. |
| Connector policy scanning respects identifier components | live | A real scanner fixture rejects the accidental `vNextCanonicalCommand`/`cal-com` match while finding a real `calCom...Policy` identifier; the repository boundary command then has no reviewed-range false positive. |
| Cancellation wins if it races a successful advisory-lock acquisition | fake | A barrier-controlled temporary lock releases and cancels at acquisition; the canceled call must not reach the mutation hook or alter CURRENT, JOURNAL, or the generation tree, and a later retry must succeed. |
| Recovery persists prepared journal before final stage rename | fake | Faults at prepared-journal-before-rename and final-rename-before-CURRENT cuts recover a complete old or new selection with no unowned deletion. Temporary roots are required because the checkpoint never materializes the real corpus. |
| Every changed publication guarantee names positive and negative proof | live | Atlas schema/validator tests mutate the real catalog fixture to omit either mapping and fail; the committed source-lock publication entry maps every declared guarantee to resolving positive and negative test symbols. |

### Discussion, design, and impact disposition

- Firstmate's report fixes the decision boundaries: the existing binding canon
  wins over the implementation, so the durable `prepared` journal is synced
  before the final stage rename. No alternate publication state machine is
  introduced.
- Foundation Atlas classification: **constrained_extension** of
  `authoring.source-lock-vnext.v1`. The repair deepens its existing publication
  owner and its proof authority only; it adds no connector-specific runtime
  route, source-lock reader, provider call, credential access, or transaction.
- Descriptor design: retain one no-follow definitions/connector/generations
  descriptor set and connector lock for each public operation. A small
  descriptor-relative Unix filesystem helper opens each final component
  atomically with no-follow semantics; all publication reads, file creation,
  tree walks, renames, leases, syncs, cleanup, and pruning use those retained
  descriptors. This is required by the existing Darwin/Linux release surface;
  Windows is not a release target.
- Boundary design: replace unrestricted compact-alias substring matching with
  component-aware matching at the shared lexicon root. Do not rename
  `vNextCanonicalCommand` to hide the finding.
- Cancellation design: after a successful nonblocking `Flock`, immediately
  recheck `ctx.Err()`, unlock if canceled, and return before the caller can
  execute a hook or mutate any durable state.
- Atlas design: add an explicit per-guarantee mapping with one resolving
  positive and one resolving negative proof. The schema requires the mapping,
  and `TestFoundationAtlasSelectorsResolve` rejects missing, duplicate,
  unknown, or nonresolving mappings.
- Scope: source locks, rendered execution artifacts, `defs.FS`, PM/runtime
  routing, provider/credential/database I/O, cross-connector transactions,
  author-owned files, `.cache/`, and certification residue remain untouched.
  `lock-render` syntax/help, PM help, manual, and website surfaces are
  unchanged; authoring canon/Atlas claims change only after their corresponding
  behavior exists.
- GSD/manual fallback: `scripts/gsd doctor` retains only the known missing
  `issue-122-rebootstrap.md`; command sources and generated
  `discuss-phase`/`plan-phase --tdd` prompts resolved. No compatible isolated
  Pi/GSD worker exists, so the required discussion, TDD plan, execution,
  verification, and review run inline and are recorded in this phase.
- Skills: loaded `go-engineering` (advanced and production), `tdd`,
  `connector-lane-build-order`, and `connector-migration-exact-sha-review`.
  Required names `golang-how-to` and the listed Go-specialist skills are not
  installed in this runtime and are not claimed. CLI parity, GSD adapter,
  task-header, source-lock canon, and Atlas maintenance references were read;
  no website/UI work is applicable.

### TDD execution order

1. Add a root-replacement destructive-path regression; observe the existing
   absolute-path escape; replace publication filesystem access with retained,
   atomic no-follow descriptors until GREEN.
2. Add the component-boundary scanner regression; observe the false
   `cal-com` finding; repair the shared lexicon matcher until genuine policy
   identifiers still match.
3. Add a lock-release/cancel acquisition-barrier regression; observe the
   cancellation loss; recheck context and unlock after successful `Flock`.
4. Change the journal-cut recovery proof to the binding order; observe the
   pre-rename journal expectation fail; move prepared-journal sync before the
   final rename and prove both recovery cuts.
5. Add failing Atlas mapping validation for an omitted positive/negative proof;
   extend schema, catalog, and validator, then resolve every source-lock
   publication guarantee before GREEN.

### CP11 correction-review repair execution record — instructions 127–129

- Intake: Firstmate instructions `127.msg`, `128.msg`, and `129.msg` were read
  in numeric order. `128.msg` bounds the Atlas change to
  `authoring.source-lock-vnext.v1` publication guarantees only; unrelated Atlas
  records remain unchanged. `129.msg` confirms that this correction proceeds
  without a new planning detour.
- Descriptor confinement: one operation retains no-follow connector and
  generations descriptors through publication, recovery, opening, and pruning.
  The lock-render path first admits its source lock through the retained
  connector descriptor without publisher state, then acquires the connector
  lock, rereads the same source bytes through that descriptor, and refuses a
  mutation before it creates `generations/`.
- Recovery order: `prepared` is written and fsynced before the same-directory
  stage-to-generation rename. Recovery handles both a pre-rename owned stage
  and a renamed generation before `CURRENT` by restoring the complete old
  selection or clearing a first publication.
- Boundary scope: compact display aliases now match one or more complete
  identifier components. `calCom...Policy` remains a policy match; the
  incidental `cal`/`com` characters inside `Canonical`/`Command` do not.
- Atlas scope: `publication_guarantees` and their one-to-one positive/refusal
  mappings are optional at the catalog schema level and enforced only for
  `authoring.source-lock-vnext.v1`. No other foundation was migrated to that
  convention.

## CP11 final repair disposition — Firstmate instructions 130–131

## Task Delivery Header

- Issue: Refs #4427 — transactional connector-generation publication repair; parent Refs #4325.
- Base branch: `main` (the existing draft PR #4294 target; the API base/head/SHA will be read after the normal push).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through draft PR #4294.
- Delivery: one ordinary non-force correction candidate from immutable parent `f7a325aec3594635acbd27e39099640283ca3663`, pushed only to `fm/cli-top100-declaration-batch-r1`, followed by API read-back and an unchanged pause for another Firstmate exact-SHA review.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: repair only F1–F4: bind temporary `CURRENT`/`JOURNAL` control files through atomic replacement; bind validated stage, generation, lease, and control identities through cleanup; retain one lock inode as the complete-operation serialization domain; and make the source-lock-only Atlas claims map to behavior-specific witnesses.
- Verification: deterministic temporary-root RED/GREEN tests for both controls, stage/generation/control cleanup, and post-acquisition lock replacement; a physical staged implemented-command preflight refusal; full and race `cmd/connectorgen`; source-lock/Atlas, documentation, static, and PR API checks.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A replaced control temporary never changes `CURRENT` or `JOURNAL` | fake | A temporary root and deterministic publisher hook replace the temporary inode after fsync. The call must refuse, retain the prior control bytes, and preserve both the moved original and unrelated replacement; the instruction forbids checked-in materialization. |
| Cleanup removes only the validated stage, generation, lease, or control object | fake | Barrier tests replace each named temporary-root object after ownership/integrity/lease/control validation. Recover, prune, and rollback must refuse without deleting either object. |
| A renamed lock pathname cannot open a second publication domain | fake | A post-acquisition barrier replaces `.connectorgen.lock` while the first operation retains its inode. A second operation must refuse while the first is held; the original state remains unchanged and a restored bound lock permits retry. |
| Atlas claims name exact physical publication behavior | fake | A local staged bundle with an implemented command is deliberately altered after semantic admission so only physical `commandrunner.Preflight` can reject it; no provider, credential, or checked-in connector input is used. |

### Scope, lifecycle, and TDD order

- Intake: `130.msg` was read completely before any new design or production edit. Per `131.msg`, it was acknowledged in the external status record and moved to `handled/` before this planning record. CP12 remains prohibited.
- Scope is limited to `cmd/connectorgen`, source-lock authoring canon/Atlas evidence, and this phase's GSD records. Do not alter source locks, generated execution JSON, runtime routing, PM CLI surface, provider/credential/database behavior, `.cache/`, or certification residue.
- Foundation Atlas classification: constrained extension of `authoring.source-lock-vnext.v1`. Publication-proof rules remain exclusive to that entry; no whole-catalog migration or generic locking framework is introduced.
- Lifecycle: `scripts/gsd doctor` has only the established missing `issue-122-rebootstrap.md`; all five command sources resolve. `discuss-phase` and `plan-phase --tdd` prompts were generated and executed inline. No compatible isolated worker/reviewer is authorized, so execution, verification, and code review will be recorded as the required manual fallback.
- Skills: `go-engineering` (advanced and production), `tdd`, `diagnose`, and `connector-migration-exact-sha-review` were loaded. Required `golang-how-to` is unavailable in this runtime and is not claimed. CodeGraph has no repository index and Go LSP is unavailable.
- RED 1 / F1: replace each fsynced `CURRENT` and `JOURNAL` temporary entry before its final rename; observe prior controls and both objects can be changed by the current pathname-based rename.
- GREEN 1: retain the file descriptor identity until a descriptor-relative identity-checked rename succeeds; cleanup verifies the same identity and never removes a replacement.
- RED 2 / F2: replace a marker-proven stage, integrity/lease-proven generation, and validated control between validation and removal; observe the current cleanup can reopen and remove a later pathname object.
- GREEN 2: retain the validated descriptor and identity through removal, revalidate the exact parent entry immediately before destructive unlink, and refuse a replacement without touching either object.
- RED 3 / F3: replace `.connectorgen.lock` after the first `Flock`; observe a second operation can lock the new inode independently.
- GREEN 3: bind the lock pathname to a publication-local stable inode identity for the complete operation, reject a discontinuity before mutation, and prove a second operation cannot enter a parallel transaction.
- RED 4 / F4: show the old compound Atlas mappings neither name the valid-stage/unowned-generation witnesses nor reach staged implemented-command preflight.
- GREEN 4: one guarantee per mapping, exact valid/refusal witnesses, and a staged malformed implemented-command refusal that reaches the physical preflight boundary.

### CP11 final repair execution — F1–F4

- F1 RED: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherRefusesReplacedAtomicControlTemporary$'` failed with both `CURRENT` and `JOURNAL` replacement paths returning `<nil>`. GREEN: the same test passed after retaining the temporary descriptor through `renameBound` (1.219s).
- F2 RED: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesReplacedValidatedStageCleanup|TestVNextGenerationPublisherRefusesReplacedValidatedGenerationCleanup|TestVNextGenerationPublisherRefusesReplacedCommittedJournalCleanup)$'` failed because Recover, Prune, and committed-journal cleanup returned `<nil>` after post-validation replacement (1.494s). GREEN: the same three tests passed (1.525s), and `TestVNextGenerationPublisherRefusesReplacedRollbackGenerationCleanup` passed (1.234s).
- F3 RED: with predecessor anchor reconstruction temporarily restored, `TestVNextPublicationOpenLockRefusesExistingLockWithoutAnchor` failed because an existing unanchored lock opened successfully. GREEN: the anchor is created only with `O_EXCL` lock creation; the restored test passed (1.005s). `TestVNextGenerationPublisherRefusesReplacedLockAfterAcquisition` additionally removes the anchor, replaces the visible lock while the first inode remains held, proves the second publisher refuses, then restores the original hard-link pair before a serial retry.
- F4 RED: with predecessor compound-mapping acceptance temporarily restored, `TestVNextPublicationProofContractRejectsCompoundMapping` failed because a two-guarantee mapping returned `<nil>` (4.11s). GREEN: one-guarantee enforcement and the real source-lock-only catalog selector test passed. The staged physical witnesses use `operationDirectReadLockForSemanticAdmissionTest`: the unmodified implemented `widgets get` command publishes/checks, while a post-admission `unsupported.bogus` flag reaches and fails `preflight staged command "widgets get"` with no `CURRENT`, `JOURNAL`, or generation member.
- The F3 companion hard link is publication-local, is never rebuilt for an existing lock, and does not create a generic locking framework. No connector source lock, generated execution JSON, PM CLI surface, provider/credential/database path, `.cache/`, or certification residue was touched.
- Lifecycle execution prompt: `scripts/gsd prompt execute-phase batch-r1-vnext-cutover` was generated and followed as the required inline/manual fallback. The adapter still cannot provide an authorized compatible isolated worker; this does not replace the required Firstmate exact-SHA review.

## CP11 exact review repair continuation — Firstmate instruction 132

## Task Delivery Header

- Issue: Refs #4427 — transactional connector-generation publication repair; parent Refs #4325.
- Base branch: `main` (existing draft PR #4294 target, verified by the GitHub API after every ordinary push).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: one ordinary non-force CP11 F1–F4 correction candidate over immutable review subject `d3661661dbd1646376e0fbae6d73ab658532a153`, API-confirmed on draft PR #4294, then an unchanged Firstmate exact-SHA-review pause.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: repair only the late temporary-control rename boundary, quarantine-bound stage/generation/lease/control removal, a matched lock-and-anchor replacement domain, and behavior-appropriate source-lock-vNext-only Atlas witnesses. CP12 remains prohibited.
- Verification: repository-only deterministic RED/GREEN tests at each late boundary; focused/full/race `cmd/connectorgen`; vNext corpus, definition, docs, static, contract, exact-diff, self-review, and PR API gates.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A late replacement cannot substitute the actual `CURRENT` or `JOURNAL` rename source | fake | The instruction forbids checked-in materialization; a temporary-root test replaces the source at a post-identity/pre-rename fault point and asserts old control, moved original, and replacement persist after refusal. |
| Cleanup cannot unlink a late stage, generation, lease, `CURRENT`, or `JOURNAL` replacement | fake | Temporary-root recovery/prune/rollback tests move a validated object aside, install a replacement at the exact pre-quarantine boundary, and assert both objects survive refusal. |
| A matched sibling lock/anchor replacement cannot make a second publisher enter | fake | A two-publisher temporary-root test installs a matching replacement pair while the original transaction holds its serialization inode, then asserts the second does not mutate, completion stays serial, and recovery succeeds. |
| Every source-lock-vNext publication claim has a behavior-appropriate witness | fake | Atlas validation resolves actual repository test symbols; physical/durable behavior is witnessed through publisher tests rather than in-memory comparators or unsynced fixture setup. |

### Discussed scope, design, and TDD order

- Intake: `132.msg` and its exact-SHA report were read in full. This is an in-scope completion of the existing root/control/lock/generation/stage identity contract, not new product work. `.cache/` and certification residue remain unread and untouched; no source lock, generated execution JSON, runtime route, provider, credential, database, reset, rebase, branch rewrite, or CP12 work is allowed.
- Lifecycle: `scripts/gsd doctor` has only the known missing `issue-122-rebootstrap.md`; all five command sources and their generated prompts resolved. Firstmate forbids role spawning and no compatible isolated worker/reviewer is available, so discussion, plan, execution, verification, and review use the required inline/manual fallback. `go-engineering`, `diagnose`, `tdd`, and `connector-migration-exact-sha-review` were loaded; required `golang-how-to` and specialist `golang-*` skills are unavailable in this runtime and are not claimed.
- F1 RED then GREEN: place each atomic control source in a retained private temporary directory; make the actual `renameat` source that fixed child, not a connector-root temporary pathname. Add a fault point after source identity observation and before the rename syscall, where a replacement preserves old control, original temporary, and replacement after refusal.
- F2 RED then GREEN: move every validated cleanup candidate into a retained private quarantine directory before destructive traversal. Validate the moved root and every bound child identity, including `.lease`; on mismatch restore the moved replacement without deleting either object. Cover late stage, pruning-generation, lease-only, rollback-generation, `CURRENT`, and `JOURNAL` replacement cases.
- F3 RED then GREEN: use the retained connector-directory inode as the operation's `Flock` domain, so a matching replacement `.connectorgen.lock`/anchor pair is only an unrelated sibling pair and cannot serialize a second publisher. Prove first completion, blocked second mutation, and post-completion recovery.
- F4 RED then GREEN: replace in-memory/unsynced/incomplete Atlas witnesses with physical closed-publication, durable-marker, late lease/current, and matched-pair tests. Keep validation exclusive to `authoring.source-lock-vnext.v1`; do not migrate the catalog.

### Execution evidence

- **F1 RED:** with the new post-identity/pre-rename barrier installed against the predecessor pathname source, replacing its child changed the prior `CURRENT` and `JOURNAL` controls. **GREEN:** `renameBound` now takes the retained private source directory and rechecks `control` immediately before `renameat`; `TestVNextGenerationPublisherRefusesLateReplacedAtomicControlTemporary` preserves the prior control, moved original, and replacement for both controls.
- **F2 RED:** late stage/generation/lease/control replacement respectively reached an ownership error, returned `<nil>`, or surfaced only active-validation failure rather than an identity refusal under direct removal. **GREEN:** private quarantine movement plus post-move root/lease rechecks make all six late-bound tests refuse while retaining both objects.
- **F3 RED:** `TestVNextGenerationPublisherSerializesMatchedLockAnchorReplacement` admitted the second publisher when a matching sibling pair replaced the original file pair. **GREEN:** the Flock now uses a duplicate of the retained connector-directory descriptor; the focused test proves no second mutation before first completion and a successful serial retry.
- **F4 evidence correction:** the review correctly classified the old closed-set comparator and fixture-written marker as nonphysical witnesses. New physical closed-tree and publisher-written durable-stage tests, plus the lease/control/matched-pair refusals, are mapped only in `authoring.source-lock-vnext.v1`.
- Focused F1–F4/Atlas proof suite passed: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisher(ActivatesClosedSetAndDefersHeldPrune|DefersHeldGenerationPrune|RefusesPruneWithInvalidLease|Refuses(Replaced|LateReplaced)AtomicControlTemporary|RefusesLateReplacedValidatedStageCleanup|RefusesLateReplacedValidatedGenerationCleanup|RefusesLateReplacedGenerationLeaseCleanup|RefusesLateReplacedRollbackGenerationCleanup|RefusesLateReplacedControlCleanup|SerializesMatchedLockAnchorReplacement|CheckRefusesPhysicalClosedSetMutation|RecoversPublisherWrittenDurableStage|SerializesWritersAndReadersSeeWholeGeneration|RollsBackFailedActiveValidationWithoutOrphan|RecoversEveryDurableCutPoint)|TestVNextPublicationOpenLockBindsConnectorDirectory|TestRunLockRenderPublishesOnlyClosedGeneration)$'` → `ok` (7.962s). The normal complete `cmd/connectorgen` suite also passed (61.981s); race and remaining gates remain recorded in the verification checklist.
- Complete race suite passed: `go test -race -count=1 -timeout 20m ./cmd/connectorgen` → `ok` (329.199s). Corpus, definition, documentation, static, contract, build, and exact-diff gates are recorded in `VERIFICATION.md`; only the required ordinary commit/push, API read-back, inbox archival, and Firstmate exact-SHA pause remain.

## CP11 final F1 repair continuation — Firstmate instruction 133

## Task Delivery Header

- Issue: Refs #4427 — transactional connector-generation publication repair; parent Refs #4325.
- Base branch: `main` (existing draft PR #4294 target, read back through the GitHub API after ordinary push).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: one normal non-force F1-only correction candidate over immutable review subject `958a07a778fba6264d1aec567efa5d8c853eefa2`, API-confirmed on draft PR #4294, then a fresh Firstmate exact-SHA-review pause.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: repair only the final `CURRENT`/`JOURNAL` source check/use transition and its dependent `authoring.source-lock-vnext.v1` refusal witness; F2 and F3 remain accepted and unchanged, and CP12 remains prohibited.
- Verification: deterministic final-source-barrier RED/GREEN for both controls; focused/full/race `cmd/connectorgen`; vNext corpus, definition, docs/Atlas, static, contract, build, exact-diff, self-review, and PR API gates.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A source replacement after final validation cannot overwrite `CURRENT` or `JOURNAL` irrecoverably | fake | Checked-in publication generations are deliberately absent. A temporary-root publisher test injects a replacement strictly after final source validation and immediately before the namespace transition, then asserts prior control bytes, moved original temporary, and replacement all persist after refusal. |
| The F1 Atlas refusal mapping names that exact final-boundary witness | live | `TestFoundationAtlasSelectorsResolve` resolves the one updated source-lock-vNext mapping to the executable final-boundary test; an unregistered or wrong mapping fails the real validator. |

### Discussed scope, design, and TDD order

- Intake: `133.msg` and `data/cli-batch1-cp11-repair-r3-final-review-r1/report.md` were read in full. The review accepts F2/F3 and authorizes no work beyond F1 plus its mapping. The known GSD issue prompt remains absent; generated discuss/plan/execute/verify/review prompts and this inline/manual fallback are recorded because Firstmate forbids role spawning.
- Skills: `go-engineering`, its advanced concurrency guidance, `diagnose`, `tdd`, and `connector-migration-exact-sha-review` were loaded. Required `golang-how-to` and specialist `golang-*` skills are unavailable in this runtime and are not claimed.
- RED: introduce an existing-suite hook strictly after the final private-source identity validation and immediately before its namespace installation. With the predecessor protocol, the hook swaps the source, `renameat` installs the replacement, and the prior control changes despite the later identity error.
- GREEN: preserve the prior control as a verified hard-link backup in a retained private quarantine before installation. After the final barrier, install the candidate, verify the installed inode, and on mismatch hard-link the installed replacement into quarantine before atomically restoring the prior control. A clean expected install removes only the verified private backup. No F2/F3 path changes.
- F4: replace only the F1 mapping negative witness with the new final-source-barrier test and update the corresponding source-lock publication guarantee/canon wording. Do not change any other Atlas entry or mapping.

### Execution evidence

- **F1 RED:** `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherRefusesFinalReplacedAtomicControlTemporary$'` failed for both `CURRENT` and `JOURNAL`: after the new final-source-validation barrier swapped `control`, the predecessor installed `"final unrelated replacement"` over the prior control.
- **F1 GREEN:** a pre-transition verified hard-link backup holds the prior control in a private quarantine. The final barrier now precedes the namespace installation, which verifies the installed inode. A mismatched installed replacement is hard-linked into quarantine before an atomic prior-control restore; the test proves prior bytes and inode, original temporary, and quarantined replacement for both controls.
- **F4 GREEN:** only the F1 source-lock-vNext guarantee/mapping changed to `TestVNextGenerationPublisherRefusesFinalReplacedAtomicControlTemporary`; the focused old/new control, durable-cut, and Atlas selector suite passed in 5.657s.

## CP11 durable F1 repair continuation — Firstmate instruction 134

## Task Delivery Header

- Issue: Refs #4427 — transactional connector-generation publication repair; parent Refs #4325.
- Base branch: `main` (existing draft PR #4294 target, verified through the GitHub API after normal push).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: one normal non-force F1-only durable-repair candidate over immutable review subject `2f381433f95fb180b00eb258539fc71bf6256737`, API-confirmed on draft PR #4294, then a fresh Firstmate exact-SHA-review pause.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: make the final `CURRENT`/`JOURNAL` source-substitution repair a typed, fsynced, restart-recoverable control transaction; update only its dependent source-lock-vNext durable witness/canon. F2/F3 remain accepted and untouched; CP12 remains prohibited.
- Verification: a bounded durable-cut matrix; test-only source/fault barriers and restart recovery for existing and no-prior controls; focused/full/race `cmd/connectorgen`; corpus, definition, docs/Atlas, static, contract, build, exact-diff, crash-cut self-review, and PR API gates.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A restart before repair resolution cannot accept a substituted public control | fake | Repository publication generations are temporary-root-only. An existing suite uses real local descriptor/fsync/rename state, injects a replacement and crash at each durable repair cut, restarts a new publisher, and asserts prior control or valid no-prior state—not the substitute. |
| Recovery retains an observed replacement and durably restores a prior control | fake | The same temporary-root suite asserts the prior inode or valid absent state, the forensic replacement, cleared repair authority, and successful ordinary recovery after a fresh publisher process model. |
| The Atlas durable guarantee names restart-capable proofs only | live | `TestFoundationAtlasSelectorsResolve` rejects an unregistered or non-mapped durable witness; its resolved mapping must name restart tests rather than the synchronous-only fault test. |

### Discussed scope, protocol classification, and TDD order

- Intake: `134.msg` and `data/cli-batch1-cp11-repair-r4-final-review-r1/report.md` were read in full. The exact review accepts F2/F3 and blocks only the F1 crash interval plus its F4 durable witness. No CP12, source-lock/rendered-output materialization, provider, credential, database, `.cache`, certification residue, reset, rebase, force push, or broader Atlas work is permitted.
- Atlas classification: **constrained extension** of `authoring.source-lock-vnext.v1`, owner `cmd/connectorgen/vnext_publication.go`. The existing publication guarantee/owner/proof seam is extended with one typed control-repair protocol and its real restart witnesses; no new shared foundation or runtime route is introduced.
- Lifecycle: `scripts/gsd doctor` retains only the known missing `issue-122-rebootstrap.md`; all five command sources resolved. `discuss-phase` and `plan-phase --tdd` prompts were generated and executed inline. Firstmate forbids role spawning and the adapter cannot provide an authorized compatible isolated worker/reviewer, so the remaining lifecycle uses the required inline/manual fallback without replacing the required external exact-SHA review.
- Skills: `go-engineering`, advanced Go guidance, `diagnose`, `tdd`, and `connector-migration-exact-sha-review` were loaded. Required `golang-how-to` and specialist `golang-*` skills are unavailable in this runtime and are not claimed.
- RED/GREEN vertical order: first record and exercise a restart after the pre-exposure authority cut; then the mismatched-install/replacement-retention/restoration cuts for each control; then no-prior paths; finally remap the Atlas guarantee only after every listed durable witness is green.

## CP11 F1 monotonic recovery-authority redesign — Firstmate instruction 136

## Task Delivery Header

- Issue: Refs #4427 — CP11-only F1 corrective redesign; parent Refs #4325.
- Base: immutable review candidate `f36b5d0a275ed27fd5f4da242ba192e43f8066d5`; existing draft PR #4294 remains `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: one scoped ordinary non-force corrective commit and push after truthful RED/GREEN, a full read-only self-review over `f36b5d0a..new`, then a fresh independent exact-SHA pause.
- Scope: replace only the mutable-path post-exposure F1 authority transition with a monotonic recovery-state protocol. CP12 is prohibited; F2/F3 behavior, provider/credential/database I/O, source-lock materialization, cache/certification residue, reset, rebase, and force push remain excluded.
- Required evidence: concrete current-state-machine trace; earlier workspace + connection-ID sync allocation/identity evidence; real same-permission rename/unlink divergence; source/target substitution; existing/no-prior `CURRENT`/`JOURNAL`; every persistence/control cut; concurrent cooperating writers; cleanup identity checks; state-machine table; counterfactual/disconfirming check; focused/full/race/static proof.
- Programme order (instruction `137.msg`): this CP11 repair remains the immediate bounded task and must finish its independent exact-SHA review before CP12. CP12–CP16 and the later Batch One connector proof sequence are future dependency-ordered phases, not scope for this candidate.

### Pre-inspection skill and lifecycle route

- Before inspecting Go source, loaded `skill://go-engineering`, `references/fundamentals.md`, `references/advanced.md`, and `references/production.md`; those supply the requested error, safety/security, strict-control-decoding, descriptor/identity, concurrency, race, and testing guidance.
- Loaded `skill://diagnose` for a deterministic divergence feedback loop and `skill://tdd` for vertical RED→GREEN behavior slices. `skill://golang-how-to` was requested as required by the repository but is unavailable in this runtime; no unavailable `golang-*` skill is claimed.
- Read `.agents/agentic-delivery/references/required-skills-routing.md` and `gsd-pi-adapter.md`. The already resolved `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` path remains required; Firstmate's no-role-spawn constraint requires its documented inline/manual fallback while preserving the external exact-SHA review.

### Design investigation contract

- **Fact to establish:** a cooperative publisher `Flock` serializes compliant writers but cannot prevent a same-permission external actor from `rename`/`unlink` against names below the connector root. The investigation must demonstrate that divergence in the real temporary-root test seam rather than infer it from lock ownership.
- **Fact to establish:** prior Polymetrics workspace + connection-ID sync code/history supplies an allocation and identity invariant reusable for a private state namespace only if exact symbols, commits, and tests demonstrate it.
- **Working hypothesis, not yet accepted:** retain one immutable prepared recovery authority at a private workspace/connection-scoped namespace, append identity-bound create-only phase records, and derive recovery from the latest verified record. This must eliminate mutable post-exposure replacement of the sole restart authority without introducing user-visible manual-unlock behavior.

### Investigation evidence and selected protocol

| Item | Evidence | Consequence |
| --- | --- | --- |
| Current F1 transition | `writeAtomicLocked` calls `beginControlRepairLocked` after the final private-source identity check, then installs `CURRENT`/`JOURNAL`; `recoverLocked` calls `recoverControlRepairLocked` before either public control is read (`cmd/connectorgen/vnext_publication.go`, `writeAtomicLocked`, `recoverLocked`). | A durable authority must exist before public exposure and must be found before public control decoding. |
| Mutable sole authority | Candidate `f36b5d0a` writes `.connectorgen-control-repair.json` through `writeControlRepairLocked`, then `updateControlRepairLocked` replaces that pathname after replacement retention and public restoration (`cmd/connectorgen/vnext_publication_repair.go`, `writeControlRepairLocked`, `updateControlRepairLocked`, `resolveControlRepairLocked`). | The post-exposure state transition cannot retain this update mechanism. |
| Lock boundary | `acquireOperation` locks a duplicated connector-directory descriptor; `vNextPublicationAssertLockBound` intentionally prevents sibling lock files becoming a second cooperative domain (`vnext_publication.go`, `acquireOperation`, `vNextPublicationAssertLockBound`). | This protects only cooperating writers; it cannot prohibit `rename` or `unlink` by another same-permission actor in the directory. |
| Real divergence | New `TestVNextGenerationPublisherPrivatePreparedAuthoritySurvivesPublicControlSubstitution` held the real operation lock, used `os.Rename` plus replacement write against `CURRENT`/`JOURNAL` and the final source, then crashed. Its RED command failed all four cases because no `.connectorgen-control-repair-*` private authority existed: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestVNextGenerationPublisherPrivatePreparedAuthoritySurvivesPublicControlSubstitution$'`. | The old root pathname is not a private restart authority; the regression must stay descriptor/fresh-process based. |
| Reusable allocation and identity model | `internal/app.(*App).migrateWarehouseIdentity` allocates opaque workspace/connection/stream identities and rejects duplicates; `allocateUniqueIdentity` reserves each generated candidate. `warehouse.LocationFor` derives `<workspace>/<connector>/<connection>` only from checked components, while `Location.EnsureOwnership` rejects an owner mismatch. | Reuse the invariant, not the warehouse code: identity selects a private region structurally, and an existing different identity fails rather than being rewritten or merged. |
| History and proof | Final history commit `fcff76a7305fe469c3903c33a89bc47912852ac6` (`fix(warehouse)!: nest materialized tables under their owning connection (#3901)`) records the cross-connection overwrite RED and path-isolation GREEN. Its retained proofs include `TestSecondConnectionDoesNotDestroyFirstConnectionRows`, `TestConnectionIdentityIsOpaqueAndNotDerivedFromNameOrCredential`, `TestSafePathPartRejectsRatherThanRewrites`, and `TestEnsureOwnershipRefusesAnotherConnectionsDirectory`. | F1 has no workspace or connection object and must not import warehouse state. Its dedicated connector-local private transaction directory uses the same random `Mkdirat(..., O_EXCL)` descriptor-allocation pattern as `vNextPublicationCreateQuarantine`, without reusing warehouse state or the quarantine object. |

**Selected design — evidence-backed, not a new product surface:** retain the pre-exposure authority as immutable `prepared.json` inside one identity-bound, random, connector-local private transaction directory. Append phase files with `O_CREAT|O_EXCL`; each phase binds the prepared file identity/digest and its predecessor. Immediately after the final atomic-source barrier, reassert the private transaction and prepared identities before public installation. Recovery scans only bound private transactions, verifies the complete monotonic chain, and derives the latest verified state before it can inspect `CURRENT` or `JOURNAL`. During cleanup, first recheck identities and retire/sync the prepared authority while the public target is valid; only then delete disposable private phase/backup members, so no pending authority can outlive its restoration material. The transaction directory is never selected by a public source or target pathname. No manual-unlock command, compatibility fallback, provider behavior, or warehouse dependency is introduced.

| State | Durable private input | Public target interpretation | Next transition |
| --- | --- | --- | --- |
| `prepared` | immutable prepared authority plus optional bound prior | untrusted | observe expected/prior/other target; append `installed`, `replacement_retained`, or `restored` |
| `installed` | prepared + identity/digest-bound installed phase | must still be the recorded expected inode | clear only when verified; otherwise retain the observed replacement |
| `replacement_retained` | prepared + bound replacement phase and hard link | substitute is forensic only | restore bound prior or valid absence, sync public root, append `restored` |
| `restored` | prepared + verified phase chain | must be the recorded prior or absence | identity-bound private cleanup; any later divergent public target fails closed |

**Counterfactual/disconfirmation:** had the mutable root record been adequate, the real locked rename tests would have found one private authority and fresh recovery would not need a separate namespace. All four RED cases found zero such authorities. Conversely, the selected protocol is rejected if a source or target rename can change the private directory identity, prepared digest, or phase predecessor without recovery failing before public-control decoding.

**GSD route actually used:** `scripts/gsd doctor` found every adapter component except the known absent `.gsd/prompts/issue-122-rebootstrap.md`; `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` to `.gsd/commands.json`, `.gsd/upstream.lock.json`, and `.gsd/official-docs/COMMANDS.md`. The five corresponding `scripts/gsd prompt` invocations were generated. Firstmate prohibits the roles those prompts require, so this CP11 pass follows their documented inline/manual fallback and records each discuss/plan/TDD/execute/verify/review result here rather than claiming a spawned role.

## CP11 F1 durable terminal-authority linearization — Firstmate instruction 139

## Task Delivery Header

- Issue: Refs #4427 — CP11-only F1 control-publication linearization correction; parent Refs #4325.
- Base branch: `main` (existing draft PR #4294 target; API base will be read back after the normal candidate push).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing draft PR #4294.
- Delivery: one ordinary non-force Design B correction from immutable candidate `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8` and parent `f36b5d0a275ed27fd5f4da242ba192e43f8066d5`, normally pushed to the existing parent branch; then pause unchanged for a fresh Firstmate-managed OMP exact-SHA review. CP11 is not claimed complete and CP12 is not started.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: replace CP11's final authority-retirement and public-control check/use protocol with retained terminal authority heads, successor-bound append-only records, descriptor-relative no-clobber capture, create-only selection, and authority-first recovery/checking for `CURRENT` and `JOURNAL`.
- Verification: candidate RED counterfactuals; temporary-root, descriptor-relative, lock-ignoring actor matrix; real CLI `lock-render --check` read-only refusal; focused/full/race `cmd/connectorgen`; source-lock corpus, Atlas/canon, static, contract, self-review, normal push, and PR API base/head/SHA read-back.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every selected public control has one retained terminal authority | fake | CP11 forbids checked-in control materialization. Temporary roots use real descriptors, fsyncs, and transaction directories; fresh recovery must identify exactly one terminal head per protected target. |
| A late public occupant is captured without clobbering | fake | A direct same-permission actor ignores the real `Flock`, mutates `CURRENT` or `JOURNAL` at post-validation barriers, and the test asserts each moved inode remains reachable in a distinct bound capture slot. |
| Check never decodes pending or divergent public control state | fake | The real `lock-render --check` path runs against a temporary root; it must return nonzero without its success line or any filesystem mutation before public payload decoding. |
| Unsupported no-replace filesystems never silently clobber | fake | Deterministic syscall-error injection asserts typed unsupported/concurrency errors, retained predecessor authority, and unchanged public occupant rather than a plain-rename fallback. |

### Design B decision, trigger, mask, symptom, and TDD order

- Authority: `139.msg` authorizes only this CP11 correction over `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`; the authoritative design/evidence is `/Users/karthiksivadas/karthik-agent-workspace/data/cli-batch1-cp11-f1-linearization-research-r1/report.md`. No CP12, rebase/reset/force-push, source-lock/rendered-output materialization, provider/credential/database I/O, `.cache`, or certification residue work is authorized.
- Trigger: a same-permission actor that ignores cooperative `Flock` replaces or unlinks `CURRENT`/`JOURNAL` after the final `fstatat` identity observation but before a clobbering `renameat` or public `unlinkat`, or after final public validation but before candidate retirement of sole `prepared.json` authority.
- Mask/symptom: cooperating writers serialize; existing hooks precede a later validation; ordinary crash cuts omit late pathname mutation; authority counting ignores authority-less residue; and public-only `--check` masks no-prior `JOURNAL` behind old `CURRENT`. Candidate `4fa9a5b8…` can clobber/unlink a late occupant, retire the only authority, and let fresh recovery/check trust a schema-valid third inode or false-success the masked check.
- Selected protocol: protected-mode marker after durable bootstrap heads; exactly one retained terminal head per target; successor `prepared.json` binds predecessor terminal identity/digest; append/fsync capture-intent, captured, selected, and terminal phases; platform no-replace public capture then post-move classification; create-only `linkat` selection; logical absence without public unlink; authority-first recover/read-only check; descriptor-authorized decode; four-attempt `retry_required`; successor-only cleanup.
- Platform/error rule: Linux `Renameat2(..., RENAME_NOREPLACE)` and Darwin `RenameatxNp(..., RENAME_EXCL)` normalize `ENOSYS`, `EINVAL`, `ENOTSUP`/`EOPNOTSUPP`, and `EXDEV` to one typed unsupported-transition error; collisions are typed conflict/retry. No plain clobbering fallback.
- RED first: candidate tests execute post-last-validation install rename, restore rename, no-prior public unlink, and final-public-validation→sole-authority-retirement counterfactuals with fresh recovery and real CLI check. GREEN then implements one vertical terminal-authority slice at a time. Atlas/canon map only actual post-validation and real-check witnesses.
- Lifecycle/skills: `scripts/gsd doctor` has the known missing issue prompt; all lifecycle sources resolve and discussion/plan prompts were generated. Firstmate forbids role spawning, so documented inline/manual fallback applies. Loaded `go-engineering` fundamentals/advanced/production, `diagnose`, `tdd`, and exact-SHA connector review guidance; required `golang-how-to` was attempted but unavailable.

## CP11 F1 Design B implementation result — instruction 139

- Implemented the report's version-3 protected-mode protocol: bootstrap creates
  terminal heads for both controls before the marker; every later transition
  creates an immutable, random successor transaction with predecessor-bound
  `prepared.json`, private anchors, bounded capture attempts, and
  identity/digest-chained exclusive-create phase records.
- `CURRENT` and `JOURNAL` public changes now use descriptor-relative
  no-replace capture and create-only selection. A present control is linked from
  a retained anchor; logical absence is observed rather than produced by public
  `unlinkat`. Linux and Darwin use their native no-replace primitives and fail
  closed on unsupported filesystems or collisions.
- Recovery and `--check` are authority-first. They scan strict private state
  before public parsing; every authorized public read revalidates retained
  marker, transaction, prepared, phase, capture, and predecessor identities
  after its graph scan and before decoding the selected descriptor.
- Automatic predecessor retirement is intentionally absent. A durable successor
  can supersede a predecessor, but the implementation retains prior authority
  rather than introduce a same-UID-unsafe cleanup mutation. The cleanup
  substitution proof therefore verifies retention and refusal, not an
  unreachable deletion barrier.
- Scope remains CP11-only: no CP12, source-lock/rendered corpus materialization,
  runtime route, provider/credential/database I/O, `.cache`, or certification
  residue access. Local implementation and verification are complete; the next
  delivery action is the one ordinary push and fresh Firstmate exact-SHA review.

## CP11 B-01/B-02 Astra correction — recovery authorization

## Task Delivery Header

- Issue: Refs #4427 — CP11 transactional publication correction; parent Refs #4325.
- Base branch: `main` (existing draft PR #4294 target).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main` through existing draft PR #4294.
- Delivery: one coherent local CP11 correction candidate over immutable review SHA `8214bd91403ce620773b61caf674faa540ee1701`, with executable B-01/B-02 RED and GREEN evidence and a fresh Firstmate-managed Astra exact-SHA re-review. No push, CP12 start, rebase, reset, force-push, new PR, provider, credential, database, source-lock materialization, `.cache`, or certification-residue action occurs before that review disposition.
- Working branch: `fm/cli-batch1-vnext-cutover-r2`.
- Task: correct only B-01 no-clobber quarantine restoration/final-generation activation and B-02 resumable valid pre-marker base authority recovery. Retain strict malformed/private-identity/graph refusals, retained terminal authority, existing descriptor confinement, and the sole schema-4 authoring-to-execution route.
- Verification: execute deterministic real-filesystem temporary-root RED witnesses against the candidate path; execute matching GREEN witnesses, relevant package and race suites, Atlas/canon validation, scoped static checks, diff/secret-residue scan, and a frozen local self-review. The external Astra re-review owns final acceptance.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A second public replacement cannot be overwritten while a mismatched cleanup candidate is restored | live | The real publisher quarantines B, observes the public name absent, then a lock-ignoring actor creates C immediately before the production restore syscall. C's inode and bytes remain public; B remains in its private quarantine; moved A remains reachable; the operation refuses. Temporary roots are required because CP11 forbids checked-in generation materialization. |
| Final generation activation cannot overwrite a late destination | live | At the real `before_stage_rename` seam, a fresh empty final-generation directory is created. Activation refuses, the staged source remains intact, and the colliding destination inode remains intact. |
| A valid interrupted bootstrap authority resumes only under the exclusive publisher lock | live | A crash after durable base `prepared.json` plus both directory syncs and before terminal append leaves shared `lock-render --check` nonzero and byte-for-byte read-only. A fresh ordinary publisher then terminalizes both heads, writes the marker, preserves the recorded public state, and successfully retries. |
| Invalid bootstrap authority remains fail-closed | live | Existing malformed, missing-prepared, successor, fork, gap, and divergent-control tests continue to refuse before public control decode; no permissive bootstrap bypass or manual-unlock surface is introduced. |

### Astra B-01/B-02 frozen ledger and impact map

- Authoritative review: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-batch1-cp11-astra-review-r2/report.md`, B-01/B-02. The source-derived counterexamples are hypotheses until the exact RED witnesses run; inherited 8214 green results are historical only.
- Review/base discipline: local `HEAD` is frozen candidate `8214bd91403ce620773b61caf674faa540ee1701`, immediate parent `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`; the retained remote ancestry checkpoint `0b214b79eeb871238ce8454cd7b896e71e2746a7` was previously proven and is not an alternate implementation base.
- Foundation classification: `constrained_extension` of `authoring.source-lock-vnext.v1`. Existing owners are `vNextGenerationPublisher`, descriptor-relative directory operations, and the terminal-control authority graph. No new shared foundation, protocol, runtime reader, importer, cache, recovery command, or compatibility path is allowed.

| Lens | State | Exact owner/evidence | Risk if unchanged | Bounded control |
| --- | --- | --- | --- | --- |
| Architecture and data flow | direct | `activateStageLocked`, `removeTreeQuarantinedLocked`, and `vNextPublicationRestoreQuarantined` mutate descriptor-relative publication namespaces; `ensureControlAuthorityLocked` bootstraps terminal heads before the marker. | A late public destination can be overwritten, or a durable valid base authority can become unrecoverable. | Use the existing descriptor-relative no-replace primitive for every B-01 destination mutation; append only the valid pending base terminal under the existing lock. |
| Affected callers | direct | Stage recovery calls cleanup through `removeTreeQuarantinedLocked`; prune and active-validation rollback call it through `removeGenerationLocked`; publication calls `activateStageLocked`; Publish/Open/Recover call `recoverControlRepairLocked`. | A narrow helper patch can miss a reachable stage, generation, rollback, or ordinary recovery path. | Test stage, prune generation, rollback generation, and activation callers; restart through a fresh publisher rather than a helper-only resolver. |
| State and durable ordering | direct | Base `prepared.json`, transaction sync, connector sync, phase append, and marker creation are ordered in `createControlRepairLocked` and `ensureControlAuthorityLocked`. | A pre-marker durable record can be either ignored unsafely or remain permanently pending. | Permit only a no-predecessor, phase-empty, prior-equals-intended, anchor-validated base record to append its terminal phase; re-scan before marker creation. |
| Filesystem integrity | direct | `renameFrom` clobbers while `renameNoReplaceFrom` is the established Darwin/Linux no-clobber primitive. | An absence check followed by plain rename has a check/use window; stage activation can replace an empty final directory. | Remove the B-01 `renameFrom` sinks; reject typed destination-exists/unsupported errors without a plain-rename fallback and retain every reachable inode. |
| Concurrency and cancellation | direct | The retained connector-directory `Flock` serializes cooperating writers but does not constrain a same-permission pathname actor. | A test that relies only on the lock hides the exact race. | Hold the real publisher lock and mutate names through test hooks at post-observation/pre-syscall boundaries; run the affected suite under `-race`. No goroutine, background worker, retry loop, or lock-domain change is authorized. |
| Security and authority | direct | Strict private transaction identities and terminal heads are the authority before public decode. | Relaxing bootstrap validation could make malformed or successor state a public-control authority. | Keep strict decoder, phase graph, anchor, predecessor, and divergent-control refusals unchanged; shared check stays non-mutating while markerless authority is pending. |
| Error and conflict behavior | direct | `vNextPublicationNoReplaceRenameError` exposes collision and unsupported primitive errors; `errVNextPublicationControlConflict` carries retry-required selection conflicts. | Converting a collision to success or a generic cleanup could hide retained state and violate the no-overwrite contract. | Preserve typed errors and wrapped context; collision leaves the public object and quarantined candidate intact, with no manual cleanup. |
| Tests and executable evidence | direct | Existing cleanup witnesses inject only before quarantine; terminal-cut tests begin from a completed bootstrap. | Existing green tests cannot falsify B-01's second replacement or B-02's pre-terminal cut. | Add deterministic, named RED tests at the exact production seams with inode/byte/tree assertions and fresh publisher recovery. |
| Canon and Atlas proof | direct | `SOURCE-LOCK-VNEXT.md` and `authoring.source-lock-vnext.v1` promise preservation at every cleanup boundary and durable bootstrap authority. | Current claims exceed executed coverage. | After GREEN, map each new behavior-granular physical/durable proof and distinct refusal proof; state resumable valid bootstrap without widening malformed-state acceptance. |
| Runtime/CLI and external I/O | none_with_evidence | CP11 remains authoring-only. `lock-render --check` is used only against a local temporary root; embedded runtime reads execution JSON, not `CURRENT`. | Repointing runtime, reading credentials, or materializing a real connector would create a second route and exceed scope. | No `internal/` runtime, checked-in definition, source lock, generated execution JSON, provider, credential, database, or public help change. |

### Diagnosis, decisions, and lifecycle

- B-01 trigger: after a mismatched public candidate B is moved into private quarantine, restoration observes the public pathname absent. A second actor-created regular C between that observation and `renameat` is the initiating substitution. Existing one-replacement tests mask it because they stop before quarantine or never create C. The visible symptom is C's lost link/bytes despite an identity-refusal result. A control where C remains public would falsify the claimed clobber root.
- B-02 trigger: an interruption occurs after base `prepared.json`, transaction sync, and connector sync but before the base terminal append. The absence of the marker is the masking condition that causes `ensureControlAuthorityLocked` to reject rather than resolve. The visible symptom is ordinary fresh recovery permanently refusing a valid durable bootstrap. A valid pending base that cannot pass strict scan/anchor checks must still refuse; that control falsifies any claim that the new path is permissive.
- No product decision remains. The complete frozen Astra ledger authorizes one coordinated B-01/B-02 correction wave only; L-01 control-cache cleanup, CP12, connectors, release work, and broad review programmes remain excluded.
- GSD/manual fallback: `scripts/gsd doctor` reports only the established absent `.gsd/prompts/issue-122-rebootstrap.md`; all five lifecycle sources and generated `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` prompts resolved. Firstmate prohibits worker/reviewer fan-out, so their required discussion, plan, TDD, execution, verification, and local review procedures are recorded inline. This does not replace the fresh Firstmate-managed Astra re-review.
- Required guidance loaded: `diagnostic-reasoning`, `firstmate-exhaustive-review`, `connector-lane-build-order`, `connector-migration-exact-sha-review`, `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, `golang-testing`, `golang-documentation`, and `golang-lint`. `go-engineering` is unavailable at the installed path and is not claimed; the concrete Go skill route covers its filesystem, durability, concurrency, error, security, and testing concerns. The available LSP is `clangd`; no Go language server is configured, so symbol work uses scoped source/search fallback.
- Executable RED observed before Green behavior edits: `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^(TestVNextGenerationPublisherRefusesSecondPublicReplacementAtQuarantineRestore|TestVNextGenerationPublisherRefusesFinalGenerationActivationCollision|TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation)$'` exited 1 in 3.250s. Stale-stage recovery, generation pruning, and rollback each restored B over C (`got="second public replacement"`, wanted C); activation returned `<nil>` after replacing the late destination; CURRENT-first and JOURNAL-second durable base interruptions each made fresh `Recover` refuse `publication control authority marker is missing from pending repair`. The B-02 test reached its real read-only `lock-render --check` snapshot assertion before that refusal.
- Focused GREEN: restoration and activation now use only `renameNoReplaceFrom`; each stage/prune/rollback witness preserves moved A, public C, and quarantined B by inode and bytes, and surfaces `fs.ErrExist` through the refusal. A strict phase-empty/no-predecessor base is terminalized only under the existing exclusive lock, then the graph is re-scanned before a missing head/marker is completed. Both CURRENT-first and JOURNAL-second restart paths preserve recorded absence, terminalize both heads, and complete a normal `lock-render` retry; shared check remains non-mutating during interruption.
- Canon/Atlas maintenance is necessary: the prior source-lock-vNext mappings named one-substitution cleanup witnesses and no pre-terminal-bootstrap witness, so they did not cover the exact corrections. The canon now states no-replace final activation and strict resumable bootstrap; the Atlas registers the three physical/durable witnesses, maps all three cleanup guarantees to the A/B/C witness, maps final activation collision, and maps the bounded bootstrap transition. No connector/runtime artifact changed.
- Final local gates GREEN: complete `cmd/connectorgen` normal package passed in 157.507s and complete race package passed in 462.893s. `make connectorgen-vnext-locks`, `make connector-canon-check`, 553-definition validation, docs validation, `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/connectorgen`, `go mod tidy -diff`, `agentcontractgen check`, generator help, `git diff --check`, and the changed-path literal-secret scan passed; the generated `pm` docs-check binary was removed.
- Impact status: local correction evidence is complete and ready to freeze as one local SHA. No push, PR mutation, CP12, `.cache`, certification-residue, provider, credential, database, or runtime action is authorized; only fresh Firstmate-managed Astra exact-SHA review may unblock continuation.

## 2026-09-05 — CP11 Codex successor review header

### Task Delivery Header

- Primary issue: [#4427](https://github.com/polymetrics-ai/cli/issues/4427), transactional connector-generation publication; parent [#4325](https://github.com/polymetrics-ai/cli/issues/4325).
- Delivery branch and PR: `fm/cli-top100-declaration-batch-r1` → `main` through existing draft [#4294](https://github.com/polymetrics-ai/cli/pull/4294). The local branch tracks that same remote branch and is two commits ahead only with the preserved CP11 candidate; no replacement branch, rebase, force-push, PR mutation, or push is authorized.
- Immutable implementation subject: `36e4d980de0d51d92fe74a68306845643596a6cb..0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e`; published/rejected predecessor `8214bd91403ce620773b61caf674faa540ee1701`; code-bearing correction `403b9973ae29cb596b4829f92a214aad7cad805f`. This header is review evidence only and does not change that source subject.
- Workspace/custody: isolated successor worktree `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli`, sole programme writer, clean before this evidence record. The frozen historical worktree, source-lock cache, and certification residue remain untouched. `no-mistakes doctor` found the shared daemon healthy with no active run for this branch; no pipeline is started during CP11 review.
- GSD/manual fallback: the current adapter resolves the five lifecycle command sources; its direct launcher requires the installed Node path and `node scripts/gsd ...` works. `doctor` reports the pre-existing missing `.gsd/prompts/issue-122-rebootstrap.md`; no compatible project-local Pi worker exists. The successor therefore records the required discussion/plan/TDD/verification/review evidence inline and uses separately assigned read-only Astra contexts. This does not waive independent audit or fresh exact-SHA review.
- Review dependencies: required repository routing; `firstmate-exhaustive-review`; retained `connector-lane-build-order`; `diagnostic-reasoning`; `go-engineering`; and the applicable Go safety, security, error, context, concurrency, testing, lint, design, and structure guidance were loaded. `connector-migration-exact-sha-review` is absent from the declared and retained skill locations; the explicit exhaustive-review and exact-SHA requirements govern its coverage rather than being claimed as loaded. CodeGraph is absent for this repository.
- Current gate: CP11 is blocked on the merged frozen review ledger's one deduplicated filesystem-reader root, `CP11-R3-01`. B-01 no-clobber restoration/activation and B-02 strict interrupted-base recovery are not acceptance by historical evidence alone; they have been independently inspected and remain subject to a fresh final exact-SHA review. CP12 remains prohibited.

### Successor audit disposition

- A distinct Astra/xhigh merged-ledger audit completed read-only against the same `36e4d980de0d51d92fe74a68306845643596a6cb..0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e` source range. It affirms `CP11-R3-01` as one Medium cause, adds the bypassing `vNextPublicationDirectoryFS.Open` semantic-admission filesystem sibling, and establishes no additional independent root or product/architecture decision.
- The authorized next group is one bounded CP11 filesystem-reader repair only: preserve descriptor-relative no-follow behavior, make potentially blocking ordinary reads nonblocking before type validation, type-validate the exact opened descriptor before exposing it, and preserve legitimate directory traversal, exclusive writes, leases, byte bounds, and current refusal/error semantics. Do not mistake a pathname precheck for the final boundary.
- The repair must use a bounded subprocess RED witness for the candidate's FIFO hang, then GREEN actual recovery, real `lock-render --check`, and the semantic-admission filesystem path. The full original CP11 change plus all repair deltas requires a fresh independent exact-SHA review after verification. No provider/credential/database operation, runtime route, generated connector materialization, CP12 work, push, or no-mistakes run is authorized in this repair group.

### CP11-R3-01 GSD discussion and TDD plan (inline/manual fallback)

- **Generated workflow context:** `node scripts/gsd sources` resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`; `node scripts/gsd prompt discuss-phase 4427` and `node scripts/gsd prompt plan-phase 4427 --tdd` were read. The official Pi workflow requires `.planning/ROADMAP.md`, which this custom issue phase does not provide, and no compatible project-local Pi worker is available. The established `doctor` limitation remains missing `.gsd/prompts/issue-122-rebootstrap.md`. Inline discussion/planning is the named fallback, not a lifecycle waiver.
- **Exact fallback source and execution boundary:** `.agents/agentic-delivery/references/gsd-pi-adapter.md`, **Agent requirements**, requires a non-Pi runner to generate `scripts/gsd prompt <command>` and expressly permits executing that generated prompt inline when compatible isolated runtime agents are unavailable or the canonical single-worker contract forbids spawning roles. The matching canonical source is `.agents/agentic-delivery/canonical/delivery-contract.json`, `state_machine.steps` `execute_tdd`, which requires the worker to “Work inline; do not spawn a specialist or GSD role.” We generated the five prompts through `node scripts/gsd`; only the discuss/plan/execute obligations have been completed inline so far. `verify-work` and `code-review` are generated procedures, not falsely claimed automatic Pi gates, and will be completed and recorded after local gates before the required fresh Astra review.
- **Discussion decisions:** this is a local publication availability/refusal bug, not provider semantics or a product choice. Keep descriptor-relative confinement, `O_NOFOLLOW`, same-descriptor identity/type checks, the single authoring `connectorgen` route, strict private authority, byte bounds, and current error wrapping. Do not add a generic filesystem/runtime abstraction, source-lock reader, public flag, manual recovery, alternate executor, dependency, or retry loop. CP11 changes no PM help/manual/website/user command surface, so CLI parity documentation is not applicable beyond preserving `connectorgen` success-output ordering.
- **RED state:** add a subprocess-only test harness that can safely demonstrate the candidate's no-writer FIFO hang without stranding the parent Go test. The child invokes real recovery/check/admission paths against temporary roots; parent timeout is the observable defect. The harness snapshots metadata only, never reads a FIFO, and establishes release through a follow-up operation after the child returns or is terminated.
- **GREEN state:** make ordinary descriptor-relative read opens nonblocking before their type is inspected, then reject an unpermitted type from that exact descriptor before any `io/fs` reader is exposed. Apply the same final contract to `vNextPublicationDirectoryFS.Open`, retaining only genuine directory traversal required for nested schema paths. Preserve create/exclusive writers and lease locking instead of applying a blanket flag change.
- **Regression matrix:** stale-stage owner FIFO through `Recover`; public CURRENT/JOURNAL FIFO through fresh recovery and real `lock-render --check`; regular-to-FIFO final-read substitution; semantic-admission `fs.FS` delegated-open substitution after enumeration; regular nested schema traversal; regular reads; missing/symlink rejection; exclusive write and generation lease behavior. Assertions prove prompt refusal/no success output, unchanged object and active identities/bytes, release of the connector lock, and no provider/credential/database work.
- **Verification plan after GREEN:** focused normal selectors, focused race selectors that do not intentionally hang, `go test -count=1 -timeout 20m ./cmd/connectorgen`, its race counterpart, `make connectorgen-vnext-locks`, connector canon/definition/docs/static checks, `go vet ./cmd/connectorgen`, `go build -o /dev/null ./cmd/connectorgen`, `go mod tidy -diff`, `go run ./cmd/agentcontractgen check`, and `git diff --check`. Full suite/no-mistakes/publication remain later programme gates. A new full-range independent Astra exact-SHA review follows all local GREEN evidence.

## CP11 R3-02–07 coordinated audited repair — Firstmate instruction 012

### Task Delivery Header

- Issue: Refs #4427 — CP11-only corrective repair; parent Refs #4325. Existing parent PR remains `fm/cli-top100-declaration-batch-r1 → main` at https://github.com/polymetrics-ai/cli/pull/4294.
- Reviewed code baseline: `3a455877cdd9686ba6f04341960a3c31196909bd`; the preceding artifact-only frozen-ledger checkpoint is `c6194254560ff874ac63e69a6c80dfe9ab06b5e2`. No source behavior changed in that artifact commit.
- Scope: the complete audited causal groups R3-02/06/07 (bounded descriptor ownership, validated reader binding, stable held-generation cleanup authority), R3-03 (actual command signal ownership), R3-04 (all 31 introduced configured-lint diagnostics and their error contracts), and R3-05 (truthful type-safe FIFO preservation witnesses). Preserve R3-01, B-01, B-02 and all prior CP11 authority/no-clobber contracts.
- Exclusions: CP12+, a new shared foundation, live/provider/credential/database exercise, raising resource limits, garbage collection, cache cleanup, reset/rebase/force-push, publication, and the known pre-range final-programme debt.

### GSD lifecycle and discussed impact

- The non-Pi inline route was prepared after the ledger commit with `node scripts/gsd doctor`, `node scripts/gsd sources {discuss-phase,plan-phase,execute-phase,verify-work,code-review}`, `node scripts/gsd prompt discuss-phase 4427`, `plan-phase 4427 --tdd`, `execute-phase 4427`, `verify-work 4427`, and `code-review 4427`; `go run ./cmd/agentcontractgen check` is current. Doctor's missing `.gsd/prompts/issue-122-rebootstrap.md` is pre-existing. Per the project adapter, prompt generation plus inline single-worker execution is the named fallback, not a waived lifecycle.
- R3-02 retains every historical authority descriptor through `scanControlAuthorityLocked`; retain immutable identities/digests but close traversal descriptors, reopen through the connector descriptor, and recheck identity before every dependent use. Partial prepared-hook return ownership must release a returned state exactly once.
- R3-06 validates A then reopens its generation path: return the validated directory descriptor or fail closed. R3-07 cannot rely on the replaceable `lease` pathname alone: reader/cleanup contention must bind to stable generation identity and cover prune plus recovery/publish cleanup siblings with consistent ordering.
- R3-03 scopes signal interception and stop/cleanup to a returning `lock-render` entry frame; non-consuming commands retain default OS signal semantics. Existing lock-render lock-acquisition/pre-mutation cancellation and retry are preservation controls unless a separate defect is actually demonstrated. R3-04 resolves all 31 audited diagnostics semantically without suppressions. R3-05 replaces read-based FIFO snapshots with no-follow type-aware identity/regular-data assertions and corrects scope claims.

### TDD order and acceptance boundary

1. Establish fresh bounded-child Group-A REDs: effective post-initialization `RLIMIT_NOFILE` exhaustion on valid retained history; actual Open A→B post-validation substitution; held A with moved/replaced lease through real Prune and recovery/publish cleanup. Each asserts returned bytes/identities, not status alone.
2. Repair the descriptor/identity group together and run normal/race continuity controls, including partial prepared-hook cleanup and private authority substitution. Do not delete history or change process limits.
3. Establish built-main Group-B SIGINT/SIGTERM/repeated-signal REDs for non-consuming siblings, then preserve existing lock-render lock-acquisition/pre-mutation cancellation, no mutation/success, and post-unlock retry without expanding to mid-transaction cancellation.
4. Resolve Group-C lint/error contracts and FIFO proof snapshots. The configured linter is the RED witness for dead code; R3-05 is evidence strengthening and must not fabricate a production fault.
5. Only after focused GREEN, changed-package/race and relevant authoring/preflight/canon/Atlas/boundary/docs/static/release evidence, freeze one code-bearing local SHA for Firstmate's required fresh Astra full-range exact-SHA review. New causal findings go to Firstmate; no piecemeal acceptance.

## 2026-09-06 — CP11 F-01–F-08 coordinated TDD plan

### Task Delivery Header

- Issue: Refs [#4427](https://github.com/polymetrics-ai/cli/issues/4427) — transactional connector-generation publication; parent Refs [#4325](https://github.com/polymetrics-ai/cli/issues/4325).
- Base branch: `main` through existing [PR 4294](https://github.com/polymetrics-ai/cli/pull/4294).
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: locally committed coordinated CP11 repair and exact-SHA evidence, then Firstmate-managed independent review; no push or merge in this scope.
- Working branch: `fm/cli-top100-declaration-batch-r1`.
- Task: repair the audited F-01–F-08 publication/proof group without relaxing B-01/B-02/R3 contracts; obtain actual RED/GREEN and a fresh independent whole-range review.
- Verification: focused behavioral selectors and retained B/R3 controls; normal/race `cmd/connectorgen`; current canon/Atlas/source-lock/preflight/static/boundary checks; review at final exact code SHA.

| Slice | RED before behavior change | GREEN acceptance and retained controls |
| --- | --- | --- |
| F-04 safe snapshot oracle | Bounded child exposes regular→FIFO, regular→symlink, and directory replacement at the oracle's classification/open boundary. | Same descriptor supplies identity and bytes or refuses boundedly; no FIFO/symlink target read; nested schema and semantic-admission controls remain. |
| F-08 real process harness | Current early exit/readiness/Wait path is held/withheld under a bounded parent and demonstrates non-deterministic readiness or unreaped child without hanging the suite. | Every successful Start arms kill/reap, every lock releases, normal/cleanup waits bound; actual lock contention readiness, real OS signal, no success/mutation, and retry are observed. |
| F-01 capture identity | Durable pending CURRENT/JOURNAL capture A→B swap immediately before each actual dependent open moves/accepts B on the candidate. | Checked descriptor is retained or rechecked before enumerate/candidate/rename/sync; B refuses before public mutation and A/B/control remain identifiable. |
| F-02 recursive ownership | Repeated actual nested A→B refusal leaks an open child descriptor or exact ownership accounting under a bounded no-GC child; Fstat error is a separate control. | Every acquired child closes once, refusal preserves actual objects/quarantine semantics, normal cleanup remains unchanged. |
| F-03 writer/owned cleanup | Actual temporary Close error after Write/Sync is silently successful or loses required typed/owned cleanup state. | Close error is surfaced without false rollback claim; primary/secondary errors remain inspectable; linked anchor and safe owned cleanup/recovery/retry are accounted for. |
| F-05/F-06 proof matrix | Equal-byte distinct B defeats prior weak identity oracle; current durable fixtures show actual prepared rather than committed state and omit rejected-new restart/final prune. | Exact A/B/C/control identities precede restoration; named durable/caller matrix decodes CURRENT/JOURNAL at every cut and exercises fresh restart/final prune. |
| F-07 canonical evidence | Prior records claim signal RED unavailable. | Dated PLAN/TDD/VERIFICATION/REVIEW correction names recovered records/commands/hashes, old uncommitted-tree limit, and does not call final code the failing state. |

Required skills recorded for this plan: `golang-how-to`, `golang-cli`,
`golang-testing`, `golang-error-handling`, `golang-safety`, `golang-context`,
`golang-concurrency`, `golang-security`, and `golang-lint`; the pinned
`firstmate-exhaustive-review` gate and project CLI parity reference were also
read. CLI help/manual/website parity is **not applicable**: no user-facing
command, flag, help text, generated manual, website content, or PM/App route is
added or changed; the tests exercise existing internal `connectorgen` behavior
only. Preserve stdout success ordering and no-success-on-refusal instead.

The plan is the documented non-Pi inline fallback for generated GSD
`discuss-phase` and `plan-phase --tdd`; generated `execute-phase`,
`verify-work`, and `code-review` are run/recorded in sequence after this plan.

### 2026-09-06 — F-07 canonical provenance correction

The earlier planned/unavailable-trace assertion is retained as superseded
history. The authoritative bounded recovery is
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-F07-signal-provenance-recovery-7294373166db75466e2c92269f7887f51ceaddc6.md`.
It binds the real pre-routing test to worker record 3224/event 3223/physical
line 3224, SHA-256
`65acae633258da13e38c4a9e0a64d532009bc2555b35ffc681a31b8d8828a14f`:
`gofmt -w cmd/connectorgen/vnext_publication_test.go && go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestConnectorgenMainPreservesNonConsumingSignalTermination$'`
exited 1 with package `FAIL` in 12.383s. The parent and the `interrupt`,
`terminate`, and `repeated-interrupt` subtests failed because `validate` did
not terminate under global signal interception.

The recorded source-edit boundary is record 3230/event 3229/physical line
3230, SHA-256
`9e4d444101f045b996d3c19ebe5e4d61ecd4abc2b2ab256b7d583bd2d88a3aa9`:
`main` exits through `runMain`; non-`lock-render` arguments call `run`; only
`lock-render` creates `signal.NotifyContext`. Record 3238/event 3237/physical
line 3238, SHA-256
`b0501c5a52f0fa821bd27907990bae074199e1bfb2563cb4cab3ea9f8c11fcff`,
then ran the combined non-consuming/lock-render selector to exit 0 with
`ok polymetrics.ai/cmd/connectorgen 15.417s`. The failing tree was an
uncommitted tested state: no full Git/tree SHA was recovered, and later
`7294373166db75466e2c92269f7887f51ceaddc6` is not its identity. This is a
historical evidence correction only; F-08 remains independent and no
second-signal or mid-transaction guarantee is asserted.

## 2026-09-06 — CP11 e77 seven-finding coordinated repair plan (steer 051)

### Delivery and GSD record

- **Immutable pre-fix authority:** behavioral `e77cd390f45eb938917dc3c39882bda34aae09a6` / tree `4c37521b11a6f1a22bd6fa5ea5e043d4982329f8`; audited ledger `3256d930c6e919ee55d25d78c8b90e7a33f9eca5`; provenance-only binding `d22f2ce42498d466728729399f231d644f2f2ca5`. The complete seven-entry audited ledger is fixed scope.
- **Inline GSD fallback:** `scripts/gsd doctor` found the pre-existing missing `.gsd/prompts/issue-122-rebootstrap.md`; lifecycle sources/prompts resolve. The adapter permits inline discuss → plan --tdd → execute → verify → review under this single-worker contract. This is a named limitation, not a waiver of TDD, audit, or fresh review.
- **Skills:** `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-context`, `golang-concurrency`, and `golang-lint`; required routing, GSD adapter and exhaustive-review gate apply. No user-facing CLI syntax/help/manual/website surface is planned.
- **Exclusions:** no provider/credential/customer-DB activity, protected `.cache` access, runtime/receiver foundation, CP12, no-mistakes, push, release, or merge.

### Dependency order and behavioral witnesses

1. **Group 1 — F-04-R/F-08-R:** record bounded helper A→FIFO/symlink/directory A→B controls and directory-open-before-flock negative control; then make observations descriptor-bound and add a real EWOULDBLOCK acknowledgement before actual-main signal. Assert exact SIGINT/SIGTERM, direct waits, no success/mutation and retry while retaining TreeSnapshot, routing and cleanup fixes.
2. **Group 2 — F-03-A/B/C:** record genuine failed-preparation, compound-cause and failed-allocation A→B failures before production edits. Repair the preparation graph, typed meaningful error producers/consumers and owned-allocation cleanup together, preserving strict bootstrap, pure absence, no-clobber A/B/C, history and one Close owner.
3. **Group 3 — F-02-P/F-05/F-06-P:** reuse trustworthy Group 1 observations for nested public quarantine recovery/prune and every named durable caller/cut. Observe raw controls, roots/content and real private authority before restoration, complete empty/nonempty B rows, and correct the false Group 3 no-authority claim without rewriting historical evidence.
4. **Freeze:** record only observed canonical/Atlas proof changes; run focused normal/race, one full `cmd/connectorgen` normal/race and required static/canon/preflight/docs/boundary-scanner/release gates. Commit/freeze the behavior candidate, then use the supplied fresh Astra/xhigh and Luna/low prompts.

### Group 1 execution update (2026-09-06)

- **F-04-R/F-08-R complete:** The original unsafe F-04 witness execution was
  preserved before the helper changed; the exact record is
  `GROUP1-F04-ORIGINAL-CONTROL.md`. The safe helper now binds observations to
  actually opened no-follow/nonblocking descriptors. No earlier executed
  pre-edit F-08 pre-flock negative was found. Its recorded pre-flock gate is a
  later repair-harness control, with inert test-only EWOULDBLOCK/EAGAIN
  acknowledgement; it uses SIGINT and SIGTERM only after acknowledgement and
  asserts exact non-consuming default signals separately. The focused Group 1
  normal/race selectors passed in package 9.198s/15.924s; hashes and the
  race-only test-bound adjustment are in `GROUP1-EVIDENCE.md`.
- **Next:** Group 2 pre-edit behavioral acceptance controls for F-03-A/B/C;
  no Group 2 production repair happens before their actual failed-preparation,
  compound-cause and owned-allocation A→B observations are recorded.

### Group 2 original-control update (2026-09-06)

- **F-03-A/B/C RED recorded before repair:**
  `GROUP2-F03-ORIGINAL-CONTROLS.md` binds the 2.687s bounded fixture command,
  test-only seam/source identities, and durable/error observations. Its
  defect-reproduction pass is RED evidence: it shows prepared-only authority
  after the post-record frontier, typed-cause flattening in real compound
  close/write paths, and temporary/quarantine A→empty-B replacement deletion.
  The complete F-03-B sibling table remains in scope for the coordinated repair;
  no source repair has begun at this record.
- **Desired-behaviour RED before repair:** after the preserved defect controls,
  the intended F-03-A/B/C assertions failed in 2.567s before source behaviour
  changed. The compile-only setup attempt is explicitly excluded. See
  `GROUP2-F03-REPAIR-EVIDENCE.md` for the exact failure classes and the bounded
  uncommitted-source chronology limit; it does not replace the original-control
  commit or recast later GREEN results as RED.
- **Coordinated GREEN:** post-record complete authority now remains coherent
  through recovery; incomplete known preparations clean record-before-anchors while
  unknown/replaced occupants fail closed. Shared compound producers preserve
  typed causes through consumers without converting absence-plus-completion into
  success. Temporary and quarantine allocation clean only identity-proven A,
  preserving empty/nonempty B. The focused selector passed in 4.389s; the full
  sibling disposition and remaining validation gates are in
  `GROUP2-F03-REPAIR-EVIDENCE.md`.
- **Intermediate package receipt:** the returned same-worker event at physical
  `10897` is preserved in
  `/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-group2-intermediate-package-raw-result.json`
  (raw SHA-256
  `4fd7390eef0b432f8f5f983d3924f9360bde2993d51586fa48bc6658634e0255`). It
  records the then-uncommitted Group 2 `go test -count=1 -timeout 20m
  ./cmd/connectorgen` as exit 0 / package `271.387s`, wall `273.672594250s`.
  It is an intermediate result only: it neither binds later edits nor replaces
  the complete final three-group normal/race validation or independent review.
- **Follow-up proof gate (steer 058):** Group 2 is not complete after the
  focused selector. `GROUP2-F03-REPAIR-EVIDENCE.md` now carries the frozen
  case-to-contract matrix for F-03-A's pre-record/Write/short-Write/Sync/Close
  and post-record sync frontiers across valid bootstrap/successor states;
  F-03-B's remaining resource-backed compound-cause consumers; and F-03-C's
  CURRENT/JOURNAL and stage/generation reachability. Add only post-repair
  coverage for those cells, observing graph/identity/residue before fixture
  restoration, nonmutating pending Check and bounded fresh recovery/retry.

### CP11 Group 2 F-03 post-repair proof completion — steer 058/060

The F-03 resource-backed proof gate is now executed, not inferred from the
intermediate package receipt. The exact test source is
`cmd/connectorgen/vnext_publication_group2_original_test.go`; its test-only
seams are `vNextPublicationControlRecordHooks`, the post-record and
directory-sync fault points in `vnext_publication.go`/
`vnext_publication_repair.go`, the raw opened-file close seam in
`vnext_publication_dir.go`, and read-control completion/close seams in
`vnext_publication.go`. Nil production paths remain direct descriptor calls.

- F-03-A: `TestCP11F03ARepairPreparationFrontierMatrix` executes the full
  record frontier at JOURNAL absent→present, CURRENT present→present, and
  JOURNAL present→logical-absence. `TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation`
  supplies absent/absent base bootstrap and
  `TestCP11F03ARepairBasePresentPresentAuthorityRecovers` supplies both
  CURRENT/JOURNAL present/present bases. They observe complete prepared
  graph/phase/anchor identity, a nonmutating pending Check, and fresh
  Recover/Check/retry; actual Sync/Close followed by injected completion is
  intentionally not described as an OS/power-loss error.
- F-03-B: `TestCP11F03BRepairCompoundCausesRemainInspectable` covers actual
  open/parent-close and parent-close/opened-file-close ownership, opened
  control read/close and pure absence, all shared record writers, predecessor,
  stage, and capture consumers with both relevant causes observable.
- F-03-C: its four named `TestCP11F03CRepair*` selectors cover Publish,
  CURRENT/JOURNAL temporary, stale-generation Prune, and stale-stage
  quarantine across empty/nonempty B, asserting A/B identity/type/bytes and
  residue before fixture cleanup and then fresh recovery/retry.

Current results are F-03-A normal `20.804s`, base normal/race
`2.617s`/`4.095s`, state-class race `10.200s`/`10.089s`/`11.507s`, and F-03-B/C
normal/race `8.442s`/`11.749s`; the primary evidence has the exact commands and
preserved earlier physical `11540/11549/11563/11577` receipts. This closes the
F-03 dynamic matrix only. It does not grant CP11 acceptance or replace Group
2's source/static/final-review boundary.

### Retained invariants

No-replace B-01 A/B/C/public-C, strict B-02 bootstrap, bounded metadata/live descriptors, capture-open identities, validated returned directory, directory lifetime lock plus lease integrity, and F-07 historical provenance stay protected. A passing defect-reproduction probe is a RED witness, not GREEN; proof strengthening never fabricates a production defect.

### CP11 Group 2 checkpoint binding and Group 3 execution — steer 061

- **Checkpoint, not acceptance:** `54746816735a964d0177a7a64646d29561f08180`
  (`fix(connectorgen): complete CP11 Group 2 authority proof`) is the coherent
  Group 2 source/test/evidence checkpoint. It follows `a9924626` and
  `8d133782`; its behavior paths are
  `vnext_publication.go`, `vnext_publication_repair.go`,
  `vnext_publication_dir.go`, and
  `vnext_publication_group2_original_test.go`. The corresponding canonical
  records are this plan, `TDD-LEDGER.md`, `VERIFICATION.md`,
  `REVIEW-CONVERGENCE.md`, and `GROUP2-F03-REPAIR-EVIDENCE.md`.
- **Bound execution record:** the Group 2 test-only seams are
  `vNextPublicationControlRecordHooks`, post-record/transaction/connector
  sync points, the opened-file-after-parent-close seam, and read-control
  completion/close seams; nil production paths stay direct descriptor calls.
  The current F-03 matrix results are A normal `20.804s`, its three race state
  classes `10.200s`/`10.089s`/`11.507s`, base normal/race `2.617s`/`4.095s`,
  and B/C normal/race `8.442s`/`11.749s`, all `ok`. The earlier physical
  `10897` full-package `271.387s` and `11540/11549/11563/11577` two-class
  receipts remain historical source-state evidence, never results for this
  checkpoint. F-03-C's actual interference is a replacement **directory** B,
  empty or holding foreign bytes—not a regular-file substitution.
- **Historical diagnosis, not reinvocation:** the e77 Astra report and ledger
  audit remain the fixed seven-entry diagnosis through `e77cd390`; they are
  not a new review of `a9924626` or this checkpoint. Group 1's post-e77
  `69246943` witness/contention changes remain completed execution evidence.
  F-03-A/B/C, F-04-R, and F-08-R await only the required fresh whole-range
  exact-SHA review; they are not restarted as new intake.
- **Authorized Group 3 execution:** implement only F-02-P and F-05/F-06-P.
  Use actual public Recover→owned-stage cleanup and Prune traversal after
  root quarantine to observe nested post-identity/pre-open A→B replacement
  and an opened nested-child identity failure, accounting separately for
  root, quarantine, child, lock, bounded descriptors, allowed prior sibling
  deletion, and residue. Extend each listed durable caller/cut with
  descriptor-safe raw CURRENT/JOURNAL bytes/type/inode, logical state,
  selected/rejected roots/content, and the real retained
  transaction/prepared/phase/anchor graph before fixture restoration. Preserve
  only fixture-owned interference, public-C/no-replace, legitimate selection
  advance, and bounded fresh recovery/retry. These are proof obligations:
  do not manufacture a production RED when an observation-only test passes on
  prior production.

### CP11 Group 3 proof execution update — steer 061/062

- **Steer 063 object-kind correction:** the earlier shorthand must not merge
  distinct contracts. Group 2 F-03-C has replacement-directory B cases;
  Group 3 F-05/F-06 has empty/nonempty replacement regular `.lease`-file B
  cases; F-02-P has nested-directory replacement. The preserved completed
  Group 3 receipts remain evidence for their respective changing source
  states, with no wording-only rerun.

- **Changed test-only paths:** `vnext_publication_resource_error_test.go`
  adds serial public Recover-owned-stage, public Prune stale-generation, and
  Publish-final-prune generation rows. Each crosses quarantine before a nested
  `removeTree` post-identity/pre-open A→B directory replacement or actual
  opened-child identity failure. `vnext_publication_durable_matrix_test.go`
  makes CURRENT/JOURNAL reads descriptor-bound and adds raw-control,
  generation-content, and private authority graph witnesses. The established
  caller matrix in `vnext_publication_test.go` consumes the durable witness.
  No production source, runtime route, connector declaration, provider,
  credential, or CLI surface changed.
- **Actual scope:** the nested rows run four bounded no-GC repetitions for each
  caller/fault pair, separately account for root/quarantine candidate/nested
  child and public-operation lock descriptors, record partial candidate residue
  before fixture-only reconstruction, and use fresh recovery. The durable
  witness captures raw CURRENT/JOURNAL regular-file bytes and inode, logical
  head, selected/rejected/stale generation tree, marker, transaction directory,
  prepared record, phase and anchor identity at the actual refusal. It invokes
  public Prune, Recover, Open, Publish-initial-recovery, prepared and committed
  new-selected recovery, final Publish prune, rejected-new recovery, immediate
  rollback, owned-stage cleanup, and non-destructive Check.
- **Variant boundary:** Group 2 F-03-C remains the replacement-*directory* B
  proof. The Group 3 durable caller rows replace the actual `.lease` regular
  member, so their empty and nonempty B cases are regular-file substitutions;
  both cases execute each unheld public schedule rather than borrowing a shared
  sink control. This is not a claim that a directory can stand in for a lease.
- **TDD framing:** F-02-P's old direct-child descriptor RED remains historical.
  The new public nested rows and F-05/F-06 witnesses are missing-observation
  proof controls; their first passing execution on repaired production is not a
  fabricated production RED. An early fixture draft correctly exposed that
  nested failure leaves quarantine residue rather than promising rollback; the
  final test records that residue before restoring only its owned fixture backup.

### CP11 artifact-only current-gate bind — Firstmate 064

The behavior-bearing checkpoint remains exactly
`7481d1770a21cc95869fd10bf281f632af48c089` / tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`. After its test/source freeze,
fresh static/admission documentation gates passed: the exact source-lock,
canon, definition, Foundation Atlas, docs, and release Make/Go recipes now
record exit 0. Runtime preflight also exited 0, with an explicitly **cached**
Go result; it is not claimed as fresh runtime execution. Their literal
commands, outputs, source references, and exact-candidate package/race receipt
separation are in `GROUP3-EVIDENCE.md` and the external current-gate report
with SHA-256 `a9bcdf60d0fb4945e096216727a39344ae87816e18abde8b6bdb71022e2bc908`.

The required `make connector-boundary` invocation has no retrievable
exit/output from its concurrent wrapper. Its observed process completion does
not constitute a pass, no duplicate run was performed, and its disposition
requires Firstmate recovery or explicit rerun authority. This artifact-only
bind neither accepts CP11 nor changes the required fresh whole-range Astra
exact-SHA review.


## 2026-09-06 — Firstmate 068/070/073 evidence correction; audit pending

This dated correction supersedes only earlier current-state claims; all
original receipts and historical text remain preserved. Behavioral/test source
remains `7481d1770a21cc95869fd10bf281f632af48c089`, tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`. Current descendant `afde575a`
changes evidence only. No test, production or dependency change accompanies
this correction.

The original current boundary scanner receipt was recovered by Firstmate068:
exit 0, wall 254.579597042s, clean 284 files/553 connectors, zero findings or
warnings, six existing exceptions. Its preserved JSON has SHA-256
`744d0d129e15c8eccbaf723dda0ca96487c8babc25ce3c2d53d6feea25ed5849`;
raw physical 13951 has SHA-256
`ec88ff8c01c57e8647206c78ff003698932b66be9420a737536073ddd75c07b7`.
The complete private 064 report after its 068 appendix has SHA-256
`a59da5a5752e5c2149ca474cf5b9ac5761fa494fe7ae470ee04084a5680ba779`;
its older `a9bcdf60...` hash is prior chronology, not a conflicting result.
All eight supplementary gates therefore have successful results, with runtime
preflight still explicitly cached. No scanner or unchanged test was rerun.
The separate boundary-package failure remains CP29 debt. The final normal
319.928s/race 783.834s receipts are intended-source pre-commit results;
new-only Group 3 lint zero does not establish whole-original-range lint green.

The complete fresh independent Astra/xhigh review returned CHANGES REQUIRED
at 7481 (SHA-256
`2d92ce239d19509aa1838c23d5d0b9f31f4e5784232a512ee2b0f93e70ca571c`).
Its five entries are unadjudicated: 7481-01 successful-create/helper-completion
ownership, 7481-02 compound EEXIST retry, 7481-03 allocation A/B proof,
7481-04 expected durable-state oracle, and 7481-05 nested phase/evidence.
Earlier Group 2 complete and Group 3 provisional-remediation language is
challenged by those findings and cannot be treated as current closure.
F-04-R/F-08-R are resolved by that review; other reviewed obligations retain
the report's partial dispositions. The whole original report and all five
entries/prior-seven dispositions are inputs to the independent complete-ledger
audit in REVIEW-CONVERGENCE.md. No repair, CP11 acceptance, CP12 advance,
provider work, no-mistakes run, push or merge follows from this record.


## 2026-09-06 — Complete Firstmate073 audit disposition

The independent fresh Astra/xhigh complete-ledger audit returned **CHANGES
REQUIRED: six actionable entries (one Medium, five Low)** against behavioral
candidate `7481d1770a21cc95869fd10bf281f632af48c089`. Its full 315-line original
text, all dispositions, ten lenses, ownership/consumer and eleven-caller/cut
evidence tables are preserved verbatim in [REVIEW-CONVERGENCE.md](REVIEW-CONVERGENCE.md)
under the dated complete-audit return. Original report SHA-256:
`bc109e85fdde9d1958b2cde7874a3f7b30b8e5d06b1b0c2764088fb34fa3e0a0`.
This supersedes current closure implications in earlier entries; it does not
rewrite their historical source, test receipts or original attribution.

7481-01/02 require common-record creation/partial-record ownership repair
(including prepared, marker and phase siblings) and complete compound-collision
error classification. 7481-03/04/06 require the previously stipulated exact
allocation A/B and fresh recovery witnesses, independent expected controls/
roots/stable history across every actual caller/cut including held-reader and
Check, and the omitted compound-error controls. The prior F03 dynamic matrix
is not completely closed. Missing proof does not establish renewed B deletion
or a current recursive descriptor leak.

7481-05 adjudicates the nested Publish fixture as initial recovery, not final
prune. Its extra final-prune label and post-Stat retained-child identity claim
are unsupported. Mandatory public nested Recover-stage/Prune-generation proof
and the separate actual final-prune lease matrix remain. No extra nested
final-prune runtime guarantee is required. F04R/F08R remain resolved; the
other old-seven obligations retain the audit's partial dispositions.

The new external four-cell marker/phase overlay exited 1 (package 1.079s,
wall 5.26s), proving desired fresh-recovery failures on unchanged source.
It is not production GREEN. Current normal/race and eight supplementary
receipts, cached preflight, recovered 068 scanner success, historical F07
chronology and CP29 debt retain their exact original limits. Current
original-base-range lint attribution is still an acceptance gate.

Only evidence aggregation/correction has occurred. No production/test repair
has started. The complete audited set now awaits Firstmate's separately
authored coordinated repair/proof scope; CP11 remains unaccepted and CP12
has not begun.


## 2026-09-06 — Firstmate079 ownership design and acceptance preparation

The captain approved the review-loop recommendations and continuation of CP11.
Firstmate079 supersedes the 077/078 research hold. Reviewed production, tests
and dependencies remain `7481d1770a21cc95869fd10bf281f632af48c089`, tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`, original base
`36e4d980de0d51d92fe74a68306845643596a6cb`. The complete audited ledger is
committed at `053853368a1514eaadf0b2411ab8740959559797`; all six findings remain
open. Existing delivery: https://github.com/polymetrics-ai/cli/pull/4294.

One fresh native `/root/cp11_ownership_design_079` was launched explicitly with
`gpt-6-astra`, `reasoning_effort=xhigh`, `fork_turns=none`, using the entire
Firstmate079 specialist prompt and the documented worker-owned GSD/native
route. Its task is a bounded common-record ownership/result contract and
collision-outcome design with regression boundaries, not another exhaustive
review or implementation. Current GSD adapter sources/reference and required
Go/review/evidence routing were reconciled before launch. No new runtime,
dependency, provider or review-policy change is authorized by this preparation.

[CP11-ACCEPTANCE.json](CP11-ACCEPTANCE.json) is the single current acceptance
manifest. It binds original 051/058/063 contracts, immutable source paths,
original desired RED/oracle and package receipts, and each obligation's caller,
actual phase, object, first side effect, owner, independently expected state,
permitted transitions, assertion location or explicit gap. Its draft contains
119 addressable obligations across six findings, including retained bounded
coverage. That count is accounting, not a test count or acceptance claim.
All rows remain open or partial; no test name or package PASS closes a row.
The manifest preserves all eleven cleanup schedules plus separate Check,
DIRECTORY B versus regular lease-file B, held readers, explicit compound
classes, and the optional nature of extra nested final-prune proof.

Next: read and verify the complete design report/source/hash, then update the
pending manifest design reference. Firstmate supplies the coherent implementation
brief after reading that result, before any production/test/dependency edit.
The installed discuss → plan --tdd → execute → verify/gap → independent review
lifecycle remains mandatory for that work. No RED/GREEN is claimed for this
JSON/planning-only preparation; no tests were rerun and no no-mistakes run owns
the branch. Original-range lint and final coherent candidate gates remain open,
with CP29 debt and every historical receipt distinction preserved.

Firstmate080 was read and acknowledged. The captain-approved
`QUALITY-EXECUTION-PLAN.md` and `ENGINEERING-EXECUTION-POLICY.md` now govern
future execution. CP11 still requires the original-base full range; later
accepted-baseline impact scopes and cumulative CP16/26/31 review are explicit
approved policy. The architect's current scope is unchanged. Manifest references
bind the received policy documents; their future QE actions are not claimed
complete by this preparation. No generic checker/framework is introduced here.

### Firstmate079 design completion / 081 compact evidence binding

The full Astra/xhigh ownership design was read and bound in
[CP11-ACCEPTANCE.json](CP11-ACCEPTANCE.json);
[CP11-ACCEPTANCE-SUMMARY.md](CP11-ACCEPTANCE-SUMMARY.md) is the compact
per-finding/caller projection for Firstmate-authored prompts. The report SHA-256
is `35f8926266fdae784e88d429eeea0b27adcd114a1550ca47cd7dece3f592a417`.
It selects explicit partial-open ownership, proven-owned incomplete-record cleanup,
complete-authority retention, and separated collision/completion outcomes.
There is no unresolved architecture/product question in that report. Its four
dependency groups remain design input for the separate implementation brief,
not authorization to edit production/tests. All architect commands are terminal;
no child/session or new GREEN is claimed. External test-content and overlay hashes,
virtual added paths, tested production SHA, and transcription versus original-tool
receipt classes are explicit in the manifest. The 119 rows preserve six audited
findings and existing variants; they are not 119 new defects or test requirements.

## Firstmate083 coordinated implementation — current TDD plan

Authority: complete Firstmate083 brief and adopted design079 (hash recorded in
CP11-ACCEPTANCE.json). Owner Astra/medium; no child roles. Intake7459a0e9,
behavior/test/dependency7481d177. No competing no-mistakes run.
GSD sources resolved for discuss/plan/execute/verify; generated plan-phase --tdd
and execute-phase prompts read. Completed079 discussion/design is reused.
Documented inline Codex execution applies because this canonical assignment
forbids additional roles; prompt generation itself is not verification.
Required Go how-to/testing/error-handling/security/safety/design-patterns/
structs-interfaces/lint guidance and existing delivery routing apply.

| Group | Findings / affected symbols | Expected behavior and intended RED | Planned verification |
| --- | --- | --- | --- |
| A | 03/04, durable/allocation witness helpers | Retained pre-cut identities/bytes/history reject readable replacement root, equal-byte different inode, actual B→C, lost history and wrong phase; preserve original oracle counterexample as RED | `go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestCP11Expected' -v`, then focused race |
| B | 01/02, openFile result, record writer and three callers, CreateTemp/classifiers | Actual created fd survives parent completion; owned incomplete bytes removed, complete authority retained, compound collision returned; desired existing probe cases RED before production edits | focused ownership/allocator normal/race, all seven phase semantics and valid state classes |
| C | 03/04/06, G2/DM/PT/RE callers and narrow fault seams | All stipulated allocation/cut/compound/resource-instance assertions use independent expected state; explicit missing proof remains open until actual assertions pass | proportional selected normal/race, exact receipts |
| D | 05/all, claim labels/Atlas/evidence/readiness check | Actual initial-recovery label; immutable baseline catches omitted IDs, stale receipts and fake/zero execution; semantic review still required | negative gate fixtures, stable final whole package normal/race, required supplementary and original-range lint |

Red: pending new Group A oracle regression execution; original073 RED receipts
remain immutable, not claimed as current executions.
Green: pending. No repaired capability or CP11 acceptance is claimed.
Production/test changes begin only after this plan record. Each completed command
will retain exact source/test/dependency binding and raw results. The 119 existing
rows and all prior resolved protections remain accountable by manifest reference.

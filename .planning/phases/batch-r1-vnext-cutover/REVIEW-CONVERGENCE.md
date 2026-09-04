# Batch R1 vNext cutover review convergence

## Immutable initial review

- Base SHA: `813f457a925f7ee3fe3bea101a43e445992c8552` (`origin/main` merge base).
- Review SHA: `0b214b79eeb871238ce8454cd7b896e71e2746a7`.
- Branch at freeze: `fm/cli-batch1-vnext-cutover-r2`, tracking `origin/fm/cli-top100-declaration-batch-r1` at the same code SHA.
- Original change surface: 21,464 paths from base to review SHA. Planning/review artifacts created after this freeze do not alter reviewed runtime behavior.
- Generated status: the cutover commit contains generated execution bundles. Any changed source lock must regenerate and check its execution JSON before the next review.
- Delivery: direct update of the existing Batch R1 remote branch only. No new PR, merge, rebase, force push, credential creation, or provider I/O.

## Architecture and execution map

| Surface | Owner and evidence | Review obligation |
| --- | --- | --- |
| Immutable authoring | `internal/connectors/defs/<connector>/source.lock.json` schema 4; `cmd/connectorgen/vnext_lock.go:24-207` canonicalizes one descriptor, strips provider-only CLI evidence, and validates seven lanes. | Locks contain provider facts and source-backed mappings only; no runtime/admission fact is permitted. |
| Deterministic renderer | `cmd/connectorgen/vnext_lock.go:210-276`; `cmd/connectorgen/vnext_lock_cli.go`; `cmd/connectorgen/vnext_lock_test.go`. | Render only metadata/spec/schemas/streams/writes/operations/CLI/optional execution JSON; prove byte identity. |
| Runtime embedding and loader | `internal/connectors/defs/defs.go`; `internal/connectors/defs/vnext_runtime_inventory_test.go`; `internal/connectors/engine/bundle.go:1253-1338`; `internal/connectors/bundleregistry/registry.go:28-31`. | Embedded runtime consumes execution JSON. Authoring locks, sources, certification, root admission, and retention artifacts must not enter `defs.FS`. |
| Direct command and provider boundary | `internal/connectors/commandrunner`; engine read/write/binary executors; CLI hand parser and `internal/cli`. | Declared operation binding, required flags/schema, credential preflight, approval, bounded encoding, and response handling remain closed before provider I/O. |
| Native/hook compatibility path | `internal/connectors/native/**`; `internal/connectors/hooks/**`; `internal/connectors/engine/hooks.go:186-210,276-305`; generated hookset imports. | A native/hook reader must not create a fixture or legacy second executor for a source-lock vNext connector. |
| Generated/help/public surfaces | `internal/cli/docs.go`, `docs/cli`, `docs/connectors`, `docs/skills`, `website/content/docs`, `website/docs.generated.ts`, `website/scripts/gen-github-cli-surface.mjs`. | Explain only the source-lock pipeline and preserve approval/credential safety; regenerate after source changes. |

## Frozen findings

| ID | Lens | Finding and reachability trace | Violated invariant | Status |
| --- | --- | --- | --- | --- |
| R1 | Declaration reachability; credential boundary | `internal/connectors/native/alpha-vantage/alpha_vantage.go:67-81` returns success for `config.mode=fixture` before its credential check; `Read` also selects `readFixture`. The same pattern is present in the native fixture-mode inventory and declared in connector specs. | A caller can bypass ordinary credential-bound execution and receive a canned alternative result. | Open; cleanup slice A. |
| R2 | Architecture/data flow; output integrity | Native hook packages, for example `internal/connectors/hooks/alpha-vantage/hooks.go:17-52`, return `handled=true` after delegating `Check`/`ReadStream` to native code. | A source-lock vNext connector must not be redirected to a legacy compatibility reader or second executor. Scope and references must be reconciled before deletion. | Open; discovery classification. |
| R3 | Tests/evidence | The inherited TDD ledger records pending RED/GREEN while the handoff calls the reference cohort green. | Delivery evidence must be executable, exact, and non-contradictory. | Fixed in planning; new RED/GREEN proof required before production edits. |

## Mandatory-lens evidence ledger

| Lens | Initial-SHA evidence and required adversarial cases | State |
| --- | --- | --- |
| Architecture/data flow | Independent audit maps lock -> renderer -> `defs.FS` -> loader -> registry -> commandrunner/engine and native/hook registrations. Native factories overwrite execution-bundle connectors and 29 direct-native hooks remain globally registered. | complete — R1-02, R1-03, R1-09 |
| Happy/bad/edge behavior | Audit covers stale output, contradictory lane availability, reference mismatch, malformed route/encoder/sync, trailing and duplicate JSON, partial loader errors, and missing optional interfaces. | complete — R1-05, R1-06, R1-10, R1-13 |
| State machine/concurrency | Per-file render publication has crash/two-writer mixed-generation cut points; `sync.Once` retains an unexplained partial registry; PostgreSQL fixture transport can persist fixture state. | complete — R1-02, R1-05, R1-09, R1-10 |
| Security/secret taint | `mode=fixture` bypasses credentials. Alpha Vantage accepts arbitrary HTTP(S) `base_url` then sends its query API key to that origin. No secret value was read. | complete — R1-04 |
| Retry/rate-limit/resume/idempotency | Native reads/direct shims bypass the engine's request budgets, paginator/resume, rate-limit/error mapping, and output path. No live provider retry was exercised. | complete — R1-09 |
| Output integrity | Expected-file byte determinism passes, but stale output survival, mixed bundle writes, native projection/redaction parity, and selected-bundle diagnostics fail. | complete — R1-05, R1-09, R1-10 |
| Declaration reachability/closed surface | Only three of 553 bundles have locks; registry replacement makes 38 Google Calendar implemented commands unreachable and the fleet preflight skips missing command surfaces. | complete — R1-02, R1-03, R1-07 |
| CLI/App parity | CLI and App share the shadowed registry. Runtime says 557 connectors while help says 556; docs mention a nonexistent `lock-check` command. | complete — R1-03, R1-11 |
| Provider semantics | Review used immutable declarations and hermetic code only. Native code owns actual HTTP/pagination/auth behavior for overwritten names; live provider semantics remain unverified by task authority. | complete — R1-04, R1-09 |
| Tests/evidence | Two Foundation Atlas proofs match no tests; the third fails to compile. The renderer test is synthetic and the TDD ledger has no executable RED/GREEN evidence. | complete — R1-01, R1-08 |

## Independent exact-SHA audit intake — BLOCK

- Independent audit source: Firstmate-isolated Codex review, delivered 2026-09-01 at `data/cli-batch1-vnext-independent-codex-audit-r1/report.md` in the supervisor workspace.
- Audited immutable code SHA: `0b214b79eeb871238ce8454cd7b896e71e2746a7`.
- Verdict: **BLOCK**. The evidence below is a frozen discovery record, not connector certification and not a claim that any repair is green.
- Safety: the audit used no credentials and performed no provider I/O. This task must keep the required vNext-only route and must not add a shared runtime foundation without a new captain decision.

| ID | Severity | Frozen finding and reachable path | Required disposition | State |
| --- | --- | --- | --- | --- |
| R1-01 | critical | `internal/connectors/defs` cannot compile because `defs_test.go` references removed `engine.Bundle` fields. | Repair stale test references; make the named Atlas proof compile and pass. | open |
| R1-02 | critical | `bundleregistry.New` loads execution bundles then `nativeset.RegisterInto` overwrites same-name connectors; 29 native hooks are still process-wide alternate paths. | Remove shadow executors/compatibility shims or obtain a captain decision for a declaration-selected native protocol foundation; reject collisions. | open — architecture decision required for genuine native protocols |
| R1-03 | high | 29 `definitionConnector` wrappers erase optional command interfaces; Google Calendar's 38 implemented commands become unknown and preflight silently skips them. | Restore interface reachability and assert every implemented command through the returned production registry/binary. | open |
| R1-04 | high security | Public fixture selection bypasses credentials; Alpha Vantage accepts arbitrary origin and attaches the API-key query secret. | Remove production fixture/origin controls; inject test transport and add hostile-origin/no-secret-send proof. | open |
| R1-05 | high | Renderer checks only expected outputs, leaves stale runtime files, and publishes one file at a time without a bundle transaction. | Close the generated file set and choose an approved whole-bundle publication mechanism before implementation. | open — architecture decision required |
| R1-06 | high | Canonical descriptor preserves raw execution blobs; schema references/lane state do not prove binding, availability, or executor satisfiability. | Strengthen the existing typed vNext descriptor contract and validate rendered in-memory bundle admission before publish. | open |
| R1-07 | high | Only Asana, GitHub, and GitLab locks exist among 553 bundles; the eight scheduled follow-on connectors are unmigrated and gates hard-code the trio. | Preserve the eight-connector rollout, migrate each lock, and derive cohort gates from the declared inventory. | open |
| R1-08 | high | Required RED is absent: two Atlas proof names select zero tests and the third is blocked by R1-01. | Add checked-in red-first reference-lock characterization with closed-set/runtime assertions; repair proof names. | open |
| R1-09 | high | Native reads and direct shims bypass engine projection, request budgets, continuation, rate-limit/error mapping, and output policy. | Remove ordinary HTTP native paths; any genuine native protocol requires the R1-02 captain decision and full parity proof. | open — architecture decision required for genuine protocols |
| R1-10 | medium | `bundleregistry.New` discards a non-empty `LoadAll` error and returns a healthy subset, hiding selected malformed-bundle diagnostics. | Retain and surface per-connector typed load failures without hiding healthy bundles. | open |
| R1-11 | medium | CLI/docs output is stale: runtime list 557 vs help 556, docs name nonexistent `connectorgen lock-check`, and fixtures are described as non-runtime. | Regenerate help/manual/website from returned execution registry and add parity proof. | open |
| R1-12 | medium | CLI schema/runner retain loadable evidence-admission values (`deferred`, `unsupported_with_provider_evidence`, `foundation_gap`, `unsupported_disposition`). | Remove latent compatibility admission branches or record an explicit current exception with tests. | open |
| R1-13 | medium | Source lock decoder accepts trailing JSON and duplicate members resolve last-value-wins. | Require EOF and reject duplicate members before typed decoding. | open |

### Captain decision packet

Only the following are genuine architecture decisions; no other review finding is being escalated as a decision:

1. **D1 — genuine native protocol treatment (R1-02/R1-09):** whether any remaining non-HTTP native protocol deserves a separately approved, execution-JSON-selected shared foundation. Without approval, the cleanup must remove native overlays/shims and mark unsupported lanes truthfully rather than retaining a second executor.
2. **D2 — closed whole-bundle publication (R1-05):** choose the durable publication contract for a connector's generated JSON set (for example, a lock-and-staged same-directory replacement with rollback and directory durability). Per-file rename cannot satisfy the frozen invariant.

The current worker must not choose either architecture. All other findings have direct in-scope RED/GREEN repair paths after the frozen ledger checkpoint is committed.

## Discovery and fix protocol

No production source changes begin until every lens above is complete, blocked, or not applicable and the frozen findings are committed. Each coherent finding group then records its RED command, GREEN command, changed paths, and newly introduced behavior. The final record is a fresh-context review of the final code SHA; any post-review code-bearing commit invalidates that review.

## #4423 N1 fresh-context exact-SHA re-review

- Review range: `1655123262586b2eaa395aa75b0e54bd7c4558bd..eae04af74b6546f9c61130d426f06197723f4f22`.
- Scope checked: `git diff --check` passed. The code-bearing paths are the three Go **test** files only; no runtime Go source, source lock, rendered execution JSON, credential store, provider client, or generated documentation changed. The two planning files record the reproducible RED/GREEN evidence.
- Behavior checked: the repaired `defs` proof now compiles against current `engine.Bundle` execution fields; each Foundation Atlas proof name is a top-level `func Test...(t *testing.T)`; each exact selector was counted once and executed; GitHub, GitLab, and Asana locks are decoded, canonicalized, rendered twice in memory, and byte-compared against their committed closed execution file sets. The negative test proves an unrendered known execution artifact is rejected by the test comparator.
- Safety checked: all new reads are fixed repository-relative test reads. There is no credential lookup, provider request, mutable global, goroutine, context creation, generated-file write, or test fixture/origin escape. Errors from root/file/AST traversal are checked and fail the test with local context.
- Result: no actionable N1 finding. This is a proof-baseline repair for R1-01 and R1-08 only; R1-05 and every other frozen finding remain open. The unrelated full-engine failure in `TestOperationRoutesFailClosedBeforeProviderIO` remains explicitly recorded rather than waived.

## Review-runner blocker

The required role-separated discovery was requested through the worker task fan-out with five independent lenses (runtime route, native/hook route, security/approval, docs surface, and TDD evidence). The runtime refused the request: `OMP task fan-out is disabled in a Firstmate worker; Firstmate alone manages Herdr worker lifecycle.` This worker must not bypass that custody boundary by starting or driving Herdr. The frozen review cannot advance to a complete finding set, production fix wave, or certification until Firstmate supplies the authorized review route or records an accepted manual-review fallback with equivalent independent contexts.

## 2026-09-02 independent N1 exact-SHA review — BLOCK

- External review range: `1655123262586b2eaa395aa75b0e54bd7c4558bd..c5bf5c5d544e85dcca5eac3ebed45ba78ad7fb33`; code-bearing SHA: `eae04af74b6546f9c61130d426f06197723f4f22`.
- Verdict: **BLOCK as an N1-to-S1A unlock, not a passing N1 review.** N1 correctly repairs its narrow proof selectors and leaves no production-code delta, but the review found the Atlas still overclaims atomic materialization, source-lock decoding accepts trailing documents, the production registry overwrites rendered executors and its command sweep skips erased surfaces, and the fixture/origin RED contract is absent while legacy tests require fixture success.
- Captain applies previously accepted D1/D2 decisions with the API-only S1A partition: API connectors use rendered execution JSON and the generic runtime; API native overwrites, delegating hook shims, public fixture/origin selectors, and hidden second executors are forbidden; PostgreSQL, MySQL, and DynamoDB native database registrations and behavior remain unchanged; whole-bundle publication must be staged, locked, atomic, durable, and rollback-capable.
- Before production deletion, S1A must add executable RED gates with exact selector counts for fixture credential-boundary refusal, full production registry/binary implemented-command reachability without skips, duplicate executor-identity rejection, and hostile-origin/no-secret-send. Then it must correct the Atlas to name unresolved R1-05/R1-13 gaps unless the approved D2 transaction actually implements them. A new independent exact-SHA review is required after the S1A code-bearing change and before publication.

## 2026-09-03 — CP05 local frozen-SHA self-review

- Candidate SHA: `6be9e9b02e14f9619a9e8f23dac142dd358579b6`; parent: `218ba977e`; review mode: required Firstmate-directed local, read-only review before the checkpoint push. No production code changed while the candidate was reviewed.
- Scope: three paths only — `internal/connectors/manifeststore/{store.go,store_test.go}` and this phase's TDD ledger. The store has no caller outside its own package/tests, so it neither creates an execution route nor changes source-lock rendering, runtime embedding, connector selection, CLI/App commands, credentials, provider I/O, or generated surfaces.
- Commands and result: `git diff --check`; `go test -count=1 -timeout 20m -race ./internal/connectors/manifeststore`; `go test -count=1 -timeout 20m ./internal/connectors/manifestindex ./internal/connectors/manifeststore`; and `go vet ./internal/connectors/manifestindex ./internal/connectors/manifeststore` passed. Broad `go test -count=1 -timeout 20m ./internal/connectors/...` was run: CP05 packages passed, while the deferred baseline failures were `TestScanFailsClosedWhenConnectorMetadataCannotLoad/invalid_cli_surface`, `TestReviewMutationQueryControls`, `TestReviewDeleteRetryEvidenceIsOperationScoped`, `TestReviewStreamContractsAreProviderShaped`, and `TestRecoveredLegacyStreamMetadataPreserved`. Firstmate instruction `103.msg` explicitly defers them to final integrated Batch 1 verification.

| Lens | Disposition |
| --- | --- |
| Architecture/ownership and data flow | Pass. `manifeststore` accepts only CP04's immutable `manifestindex.Index`; lookup precedes loading. It has no loader for source locks or other execution artifact formats, and no caller currently changes the existing runtime route. |
| Happy/bad/edge behavior | Pass. Tests observe cached hits, deterministic LRU count and byte eviction, oversize no-retention, unknown-key no-load refusal, caller-mutation isolation, and loader error propagation. |
| Concurrency/cancellation/cache bounds | Pass. One owned context is shared only for the indexed key's active flight; a first cancelled waiter leaves later waiters live, while an abandoned flight is cancelled, removed, and may retry. Positive count/byte limits are mandatory; bytes retained never exceed the admitted payload budget. Race testing passed. |
| Secret/credential taint | Pass. The only loader input is `manifestindex.Entry`; no credential lookup, logging, serialization, provider request, or secret-bearing field exists. The review and tests used no credentials. |
| Retry/resume/idempotency | Pass. The store retries nothing itself. A later explicit `Load` after a failed or abandoned read creates a new load; no provider operation, checkpoint, continuation, or replay path is introduced. |
| Errors/output integrity | Pass. Invalid limits, nil loader/context, and unknown connectors fail deterministically before a loader starts. Successful callers receive private copies; the owned cached bytes cannot be mutated through a returned result. |
| Closed declaration/CLI/App reachability | Pass. The package has no callsites outside its own tests and does not alter rendered declarations, CLI parsing/help, App catalogs, executors, approval, or credential boundaries. |
| Compatibility/generated/docs parity | Not applicable. No public CLI, generated execution JSON, lock, docs, or website surface changed. Package documentation declares ownership and bounds. |
| Resource limits | Pass. Cache payload retention is bounded by entries and bytes, with LRU eviction; active loads are keyed only by the finite immutable index and are cancelled after the final waiter leaves. |
| Behavioral test evidence | Pass. Focused unit, package, race, and vet commands above assert observable loader-call counts, returned bytes, cache isolation, and cancellation/retry transitions. |

- Findings: no CP05 code finding. The five broad-suite failures named above are externally deferred baseline evidence, not waived or repaired by this microcheckpoint. No correction phase is required; the focused suite remains the final code evidence for candidate `6be9e9b02e14f9619a9e8f23dac142dd358579b6`.
- Final local validation after this read-only review: reran `go test -count=1 -timeout 20m -race ./internal/connectors/manifeststore`, `go test -count=1 -timeout 20m ./internal/connectors/manifestindex ./internal/connectors/manifeststore`, and `go vet ./internal/connectors/manifestindex ./internal/connectors/manifeststore`; all passed. `lock-render --check` passed for GitHub, GitLab, and Asana; `go run ./cmd/connectorgen validate internal/connectors/defs` checked 553 connectors with 0 findings; `git diff --check` passed. A targeted secret scan of CP05 source/evidence found no private-key, GitHub, OpenAI, or AWS token patterns. The only untracked local state remains `.cache/` and `internal/connectors/certifications/`; neither was read, staged, or modified. Disk available: 211 GiB.

## 2026-09-03 — CP06 Go skill route and impact map

- Firstmate instruction `104.msg` arrived during CP06, after the initial immutable-path inspection and RED proof were already recorded. Before further implementation and review, the installed `golang-how-to` orchestrator and the selected `golang-structs-interfaces`, `golang-design-patterns`, `golang-safety`, `golang-concurrency`, `golang-context`, `golang-testing`, `golang-security`, `golang-error-handling`, and `golang-data-structures` skills were read from the Firstmate CLI workspace. No skill artifact is imported or copied into production.
- Impact map: `internal/connectors/connectors.go` held the process-global default builder; `internal/connectors/bundleregistry/registry.go` constructed engine/native connectors through import/init and name-selected factories; `internal/connectors/engine/hooks.go` exposes the legacy global hook registry; generated `hooks/hookset` blank imports activated it; `internal/app/app.go` opened the production registry indirectly; `internal/cli` process fixtures depended on the global builder. CP06's exact scope is explicit constructor ownership, duplicate/unknown pre-construction rejection, generated direct hook factories, and a named native compatibility adapter. It does not alter a connector source lock, rendered execution JSON, provider transport, approval/credential behavior, or executor selection in a connector manifest; CP07/CP08 own those selections.
- Applied constraints: factories return concrete connectors and validate IDs before construction; factory maps are private and preallocated; production uses direct generated hook closures rather than context values, mutable global selection, import order, or registration order; context remains request-scoped and no function/dependency is stored in it; test injection passes typed opener functions directly; nil/duplicate/unknown factory paths fail closed before I/O; no user-origin, credential, or secret flows into selection.

## 2026-09-03 — CP06 local frozen-diff self-review

- Review subject: the uncommitted CP06 candidate on `fm/cli-batch1-vnext-cutover-r2`, based on `cb33f2ef2584c7307f3c88b869c474796698b0fb`. Firstmate `110.msg` requires this local ledger before a candidate commit; therefore this record freezes the reviewed changed-path set rather than inventing a code SHA before staging.
- Scope: generated metadata/index/selection, generation-aware bounded bundle store, explicit executor/extension/native construction, App/CLI lazy construction, rate-limit preservation Harnesses A–C, Atlas/generator boundary ownership, and the exact tests/evidence files. No source lock, rendered execution artifact, provider route, credential, provider resource, release artifact, or certification path is in the reviewed diff.
- Review method: current-source read of construction/index/store/App/CLI/generator/adapter paths; changed-path census; exact focused/race suites; local built-binary help/namespace/inspection smoke; generator/lock/definition/boundary checks; scoped broad App/CLI results; and `git diff --check`. CodeGraph and Go LSP remain unavailable, so the CP06 impact-map fallback searches and tests remain the source-intelligence record.

| Mandatory lens | Disposition |
| --- | --- |
| Architecture/data flow | Pass. Generated execution JSON produces a compact metadata/executor/extension index; `Construction` validates selection, `BundleStore` acquires one identity, and `engine.New`/native adapters consume that selected bundle. No source-lock runtime reader or second executor route was introduced. |
| Happy/bad/edge behavior | Pass. List is zero-decode; named lookup is one-decode; unknown extension, reserved builtin, unknown executor, malformed metadata, oversize, held-capacity, generation/digest mismatch, and canceled flight controls are executable. |
| State machine/concurrency | Pass. Store flights/cache/holds key by `{connector,generation,digest}`; waiter cancellation and retry are race-tested; generation leases survive cache-handle release and are explicitly releasable. No new background goroutine is owned by construction. |
| Security/secret taint | Pass. Invalid construction occurs before injected project stat; no credential or provider action occurs in the controls. Rate scope uses existing opaque coordination identity. Tests use local fake HTTP only and no secret value is logged or stored. |
| Retry/rate-limit/resume/idempotency | Pass. Harness B proves two production GitHub selections share a same-scope local admission after fake 429/reset while a different scope remains independent. Existing GitHub rate and parked-resume proofs pass; no retry, shared-policy, or parking semantic changed. |
| Output integrity | Pass. Command summaries are generated from execution `cli_surface.json`; root help, namespace help, dynamic connector help, and GitHub text/JSON inspection execute successfully. No raw response/secret output path changed. |
| Declaration reachability/closed surface | Pass. `ManifestSelections` is a closed native/compatibility inventory; runtime validates every selected ID against factories; generated index is explicitly classified as a shared generated output; whole-tree boundary scan has zero findings. |
| CLI/App parity | Pass. `App.Open` and CLI metadata construction do not decode the fleet; transport composition is lazy at a declared execution route and preserves the pre-open registry ownership for test replacement seams. Local smoke completes plan → preview → approval → execute. |
| Provider semantics | Not applicable with evidence. No connector definition/rate file/source lock changed. GitHub 429/reset evidence is an in-process `httptest` response; no provider endpoint was contacted. |
| Tests/evidence | Pass with recorded inherited baselines. Focused/race/package/build/docs/generator/preflight/boundary/smoke checks pass. Full App/CLI/boundary/lint failures are the named pre-existing baselines recorded in `VERIFICATION.md`; none is introduced by CP06. |

- Findings: no new CP06 finding. The self-review found and corrected three in-scope issues before this freeze: missing generated Catalog/CDC metadata parity, reserved-builtin index acceptance, and handwritten provider-literal generator/boundary classification.
- Frozen-diff result: zero unresolved CP06 blocker. Candidate commit and ordinary parent push remain separate delivery steps; no external review was requested or performed.

## 2026-09-03 — CP07 GitHub reference local self-review

- Review subject: the five-file uncommitted CP07 proof diff on parent `843a32de5f927b1235cc00883fa0c5e0f5ea8c5b`.
- Scope: GitHub generated-index / production-registry characterization plus phase evidence. It adds no production route, source lock, rendered JSON, credential, provider call, rate policy, database behavior, or release surface.

| Lens | Disposition |
| --- | --- |
| Selection/data flow | Pass. The witness follows generated entry → `NewRegistry` → `*engine.Connector`, asserting the exact engine and hook IDs rather than connector-name fallback behavior. |
| Security/rate | Pass. It observes only safe declared process-local coordination; it does not create a credential or send a request. |
| CLI/App/provider parity | Pass. Existing CP06 built-binary help/inspect and command-preflight evidence remains applicable because CP07 changes no runtime surface. |
| Tests/evidence | Pass. Exact witness, race registry suite, GitHub render check, commandrunner preflight, and build passed. Manual-TDD fallback is explicit because CP06 already supplied the selected state. |

- Findings: none. No external review is required for this microcheckpoint; Firstmate's CP08 coherent A1 review gate remains controlling.

## 2026-09-03 — CP08 PostgreSQL reference local self-review

- Review subject: the five-file uncommitted CP08 proof diff on parent `c267f6ccb6988c6d0132f264e963c6701b8134f1`.
- Scope: generated-index / production-registry PostgreSQL characterization and evidence only. It adds no source lock, execution artifact, native database protocol, container, credential, provider action, rate declaration, or release behavior.

| Lens | Disposition |
| --- | --- |
| Selection/data flow | Pass. The witness follows generated entry → `NewRegistry` → `nativepostgres.Connector` and asserts no extension/API/compatibility selection. |
| Security/rate | Pass. PostgreSQL's explicit not-applicable rate declaration yields no coordination claim; no credential or network action occurs. |
| Native boundary | Pass. Existing CP06 selected-bundle native constructor proof remains applicable; the CP08 witness does not restore a second bundle load. |
| Tests/evidence | Pass. Exact witness, race registry/native suites, commandrunner preflight, and build passed. The manual-TDD fallback is explicit because CP06 already supplied the selected state. |

- Findings: none. CP08 completes the coherent A1 reference-proof pair; request Firstmate's controlling A1 review disposition before a CP08 parent push.

## 2026-09-03 — A1 coherent exact-SHA review request

- Requested range: `9f96054ed85d6470306bc5b033e805b89b36c7c6..62e04650f44878da013a893bbeccc22c3b9b690c` (CP02 through CP08), published normally on `fm/cli-top100-declaration-batch-r1`.
- Required reviewer context: independently review the entire CP02–CP08 range, not only CP08; validate the closed sync contracts/resolver, generated index/store identity, explicit factory and hook fan-in, App/CLI pre-I/O/lazy construction, GitHub API-engine reference/rate admission, PostgreSQL native reference/no-rate claim, generated boundary ownership, and all recorded inherited baselines.
- Worker disposition: Firstmate must launch and record the independent exact-SHA phase review. This worker must not start CP09 or create a second reviewer path until the review disposition is handled.

## 2026-09-04 — A1 correction-wave local self-review

- Review subject: the uncommitted A1 correction wave based on reviewed parent `346200f659e9ae22d0be42a83ee10f75d658f0b6`; review authority is Firstmate report `data/cli-a1-62e-phase-review-r1/report.md`.
- Scope: A1-01 pure `synccontract`/`syncplan` closure, A1-02 generator/engine/store loaded-identity binding, A1-03 one production CLI registry construction, focused tests, Atlas ownership, and GSD evidence. No source lock, rendered definition JSON, rate declaration, provider route, credential, release artifact, `.cache`, or certification residue changed.

| Lens | Disposition |
| --- | --- |
| Closed resolver model | Pass. Seven canonical mode rows bind the progression/apply/object/key/delete/retry/idempotency/receipt/acknowledgement/checkpoint axes. `Resolve` returns an executable plan, C3 with the actual invalid axis, or C4 with the declared foundation reference before construction or I/O. |
| Loaded identity and bounds | Pass. Generator and `engine.Load` share `manifestidentity.ForFS`; the store requires generated and loaded connector/generation/digest/byte charge equality before cache insertion or handle exposure. Same-name mismatches cannot reach factories or rate resolution; flight, cancellation, eviction, and lease tests pass under race detection. |
| CLI/App ownership | Pass. Production mode owns one registry from `cli.Run` through root manual, dynamic help, direct preflight, normal App, and reverse App. Test override mode is explicit rather than inferred from variadic candidates. The normal/help/reverse counter witness resolves one selected connector. |
| Rate, factory, and native safety | Pass. GitHub local fake 429/reset admission, GitHub API-engine, PostgreSQL native/no-rate, 49-hook, and closed 3+4 compatibility controls pass. No rate declaration, provider request, native protocol, or compatibility fallback changed. |
| Generated and public surfaces | Pass. Generation, reference lock checks, 553-definition validation, Atlas JSON/proof checks, commandrunner preflight, docs validation, built help/namespace/inspect, local warehouse/reverse smoke, and whole-tree boundary check pass. |
| Error and lint review | Corrected during review: moving execution identity out of `engine/bundle.go` left `readFileString` and `dirExists` dead; they were removed and the engine/identity/store/index/registry suites rerun. The remaining 20 lint diagnostics are confined to unmodified historical files and remain a recorded baseline. |
| Broad suites | Recorded, not waived. Full App and CLI reruns reproduce only the existing typed-destination and polling/help/source-origin baselines listed in `VERIFICATION.md`; engine passes. |

- Local result: no remaining A1-01/A1-02/A1-03 finding after the dead-helper cleanup. `git diff --check`, scoped vet, build, agent-contract, generation, safe gates, and focused/race suites pass. The external final exact-SHA A1 re-review remains mandatory; do not begin CP09.

## 2026-09-04 — A1 entry-capacity correction local self-review

- Review subject: the uncommitted narrow correction based on exact reviewed candidate `701a0b45175f308400c938322fd1634a28efdaef`.
- Workflow: generated `scripts/gsd prompt verify-work batch-r1-vnext-cutover` and `scripts/gsd prompt code-review batch-r1-vnext-cutover` paths were followed inline. The current runner cannot spawn the GSD reviewer role, so no second local reviewer was fabricated; Firstmate's mandatory fresh independent exact-SHA review remains the controlling gate.

| Lens | Disposition |
| --- | --- |
| Entry capacity and eviction | Pass. `reservedEntries` is incremented in the existing mutex-held distinct-flight reservation, participates in the same capacity/eviction loop as retained cache entries, and decrements in the matching terminal loader path. No same-key waiter consumes a second slot. |
| Cancellation and retry | Pass. `releaseWaiter` still cancels an abandoned loader but does not free byte or entry reservation; the loader terminal path frees both together. The barrier regression proves B is refused before and after A's caller cancellation, then B and A retry only after capacity becomes available. |
| Identity, rate, and construction | Pass. The diff touches no identity comparison, loader result, factory, rate, connector, App, or CLI path. Existing affected identity/index/registry suites and full manifeststore race suite pass. |
| Foundation/documentation | Pass. `definitions.bundle-loader.v1` declares the strengthened retained-plus-flight entry guarantee and names the executable regression proof; JSON/unique-ID and Atlas selector validation pass. |
| Safety and scope | Pass. No provider, credential, source lock, rendered JSON, release, cache residue, or certification path is read or changed. `go vet`, build, agent-contract, and `git diff --check` pass. |

- Findings: none. Commit this reviewed correction locally, report its exact SHA to Firstmate, and do not start CP09 or push the parent branch before Firstmate relays a fresh independent review disposition.

## 2026-09-04 — A1-04 independent exact-SHA review PASS

- Authority: Firstmate instruction `120.msg`.
- Reviewed candidate: `c761e7e6f2d042c7560ab0c520dc9aa182110e6e`.
- Disposition: **PASS — zero blockers.** The candidate closes A1-04. Its exact reviewed content remains immutable; publish it only as a normal non-force update to the declared continuation parent `fm/cli-top100-declaration-batch-r1`.
- Release boundary: no reset or rebase to `main`; `main` remains only PR #4294's eventual merge base. CP09 strict source parsing is authorized only after the normal parent update.

## 2026-09-04 — CP09 strict source parsing local self-review

- Review subject: the uncommitted #4426/N2 CP09 parser/graph, test, source-canon documentation, Atlas, and GSD evidence change based on normally published parent `988dd16c3d206a28d3e7b16f8a0d805c4163f7ca`.
- Workflow: `scripts/gsd prompt execute-phase`, `verify-work`, and `code-review` were resolved and completed inline. The configured runtime cannot supply the canonical isolated GSD worker/reviewer, so no counterfeit second reviewer was spawned; the manual fallback is recorded in the phase evidence.

| Lens | Disposition |
| --- | --- |
| Decoder and publication boundary | Pass. `decodeStrictJSON` remains the one duplicate/trailing/unknown-field decoder. Graph construction and `engine.Load` complete before `runLockRender` reaches a write; sentinel tests prove refusal preserves existing output. |
| Role semantics | Pass after review correction. CP09 enforces only structural lane compatibility: record→matching stream, request→write or direct operation, response→stream or direct operation. The initially considered source-route comparison and direct-operation schema rejection were removed because they are CP10 semantic admission, not parser work; the direct-role regression now protects that boundary. |
| Determinism and identity | Pass. Same-rank lane/command rendering orders by immutable source identity. Equivalent operation ordering produces equal execution bytes and graph digest while graph-retained provider, CLI, and operation facts remain absent from execution identity. |
| In-memory validation safety | Pass. The read-only virtual filesystem carries rendered byte slices into `engine.Load`; it invokes no connector construction, credential, transport, resolver, preflight, provider request, or filesystem publication. Every committed schema-4 lock builds successfully through that path. |
| Documentation and Atlas | Pass. `authoring.source-lock-vnext.v1` revision 30 names the graph owner, constraints, non-goal, and proofs; `SOURCE-LOCK-VNEXT.md` describes structural role placement without overstating CP10 joins. Catalog schema/unique-ID, Atlas selector, and docs checks pass. |
| Tests and static gates | Pass. Focused rejection/admission/determinism/corpus controls, full `cmd/connectorgen`, reference check-only renders, 553-definition validation, vet, build, agent contract, docs check, and `git diff --check` pass. |

- Findings: none after the structural-role scope correction. Do not start CP10, alter source locks/rendered artifacts, or publish before processing the inbox/current-state gate.
- Publication: the inbox/current-state gate was empty before commit. The reviewed checkpoint was committed as `d11277378abe556323226e3f6998ce3caf6033dc` and normally pushed to `fm/cli-top100-declaration-batch-r1`; PR #4294 API read-back reports that exact draft parent head over `main`. No external review was requested or fabricated, and no CP10 work follows from this checkpoint.

## 2026-09-04 — CP10 semantic source-execution admission local self-review

- Review subject: the uncommitted CP10 authoring-side admission, tests, source-lock canon/Atlas documentation, and GSD evidence change based on normally published parent `85c28e70e4c8f811ea342a1f1054e09759cde1c1`.
- Workflow: generated `execute-phase`, `verify-work`, and `code-review` prompts were resolved and fulfilled inline. The established phase cannot meet the adapter's ROADMAP prerequisite and the configured runtime cannot provide its isolated GSD worker/reviewer; no counterfeit reviewer was spawned. Firstmate instruction `122.msg` explicitly defers independent review to the coherent phase boundary.

| Lens | Disposition |
| --- | --- |
| Authoring/runtime data flow | Pass. The CP09 graph remains the only source-lock representation. Admission renders bytes into the existing read-only virtual FS, calls `engine.Load` and a no-I/O `engine.New`, then returns only an in-memory staged value. It neither registers a connector nor reads source locks from runtime dispatch. |
| Exact semantic joins | Pass. Schemas, stream/write/operation nodes, command surfaces, implemented binding resolution, and source known facts are matched against the loaded bundle. Same-route GraphQL commands require the owning immutable source operation, not route equality. Unknown provider fields remain raw. |
| Manifest and sync identity | Pass after self-review correction. A supplied sync plan now requires both its source executor/digest and `GenerationDigest` to equal the staged manifest identity before `syncplan.Resolve`; the stage does not fabricate a destination, Foundation Atlas fact, target, or digest. |
| Boundary and security | Pass. `commandrunner.Preflight` runs only against the no-I/O engine connector; no credential, transport, provider request, filesystem staging, activation, or generated global index path is added. Rate validation comes only from the existing loader and is covered by a no-write sentinel. |
| Determinism and output integrity | Pass. Output bytes are loader-identified, staged provenance is sorted and duplicate-refused, and the one-entry manifest index retains the exact loaded identity. The failing GraphQL, malformed-rate, resolver, and preflight controls leave the pre-existing output sentinel unchanged. |
| Documentation and Atlas | Pass. The source-lock canon describes the semantic boundary without granting runtime source-lock access. Catalog revision 31 names the existing authoring foundation, real owner APIs, constraints, no-goals, and all CP10 proof selectors; no shared runtime foundation is claimed. |
| Tests and scoped gates | Pass. Focused RED/GREEN controls, full `cmd/connectorgen`, reference `--check` renders, 553-definition validation, Atlas selector resolution, vet, build, docs check, and agent-contract check passed. The required full `internal/cli` re-run reproduces only its recorded polling/source-origin/ETL-help baseline failures and local Redis connection-refused logs. |

- Findings: no remaining CP10 blocker. The review corrected the source-graph comment to describe its now-authorized in-memory admission role and added generation-digest equality for supplied-sync identity closure before final validation. No connector source lock, rendered execution artifact, provider behavior, credential path, release path, `.cache`, or certification residue changed.

## 2026-09-04 — N2 boundary-review CP09/CP10 correction local self-review

Read-only review scope: `cmd/connectorgen/vnext_admission.go`, `vnext_graph.go`, the focused admission/graph tests, `SOURCE-LOCK-VNEXT.md`, and the `authoring.source-lock-vnext.v1` Atlas entry.

| Lens | Disposition |
| --- | --- |
| Effective schema authority | Pass. Request roles compare canonical staged JSON with every actual loaded write/direct request schema; response roles compare with the actual loaded stream schema. Existing-but-swapped schemas and direct-only roles with no typed consumer fail at the authored `schema_refs` pointer before `lock-render` replacement. |
| Closed runtime selection | Pass. The existing native selection remains authoritative; the generated closed `hookset.Factories` inventory supplies at most one exact connector hook. The selected hook is passed to `engine.New`, and its exact extension is retained in the staged manifest/index. The GitHub real-lock test compares the full entry with `manifestindex.GeneratedEntries` without provider or credential work. |
| Deterministic provenance | Pass. Source ID sort assigns `CanonicalIndex` only after structural validation. Every staged operation-derived provenance field uses it; all errors retain `Index`, the author-visible input position. The reorder proof checks full provenance equality and a malformed reordered input still identifies `/operations/0/...`. |
| Boundary/no-write safety | Pass. Admission remains entirely in memory until `runLockRender` receives a successful stage. No second lock reader, manifest source, provider request, credential path, global output, certification, staging, activation, or CP11 path was added. |
| Compatibility and documentation | Pass. Full generator tests and three current reference lock checks pass. The canon and Atlas name the current direct-response limitation rather than inventing a response executor, retain source-lock runtime isolation, and list executable focused proofs. CLI help/manual/website surface parity is not applicable: no user-facing command, flag, output, or connector contract changed. |

- Findings: no remaining local correction blocker. The manual GSD fallback is used because no compatible Pi worker/reviewer runtime is available and the established adapter blockers remain recorded; it does not replace the required fresh Firstmate exact-SHA independent review. `.cache/` and certification residue were neither read nor changed.

- Reviewed correction range: `56ec3d9d7dc1d726203b0ef0c03ddec3209b8dde..b4dd03e8c2e113f1e791f5844db253135cdb5c9e`. The independent review's F1/F2/F3 defects are covered by focused RED/GREEN proofs, the full generator package, affected runtime packages, reference lock checks, Atlas/docs checks, scoped static gates, and this read-only review. Fresh external exact-SHA review remains mandatory for the final published evidence checkpoint.

## 2026-09-04 — CP11 B1 transactional publication local self-review

- Review subject: the uncommitted #4427/B1 authoring-only publisher,
  `runLockRender` wiring, temporary-root proofs, source-lock canon/Atlas/docs
  parity, generated website metadata, and CP11 evidence. The permitted
  production boundary is CP10's admitted in-memory staged generation; no
  checked-in connector root was materialized.
- Workflow: the resolved `execute-phase`, `verify-work`, and `code-review`
  procedures were fulfilled inline. The established adapter has no compatible
  isolated worker/reviewer or phase roadmap, so this is a recorded manual-GSD
  fallback—not an independent review or a substitute for Firstmate's
  exact-SHA gate.

| Lens | Disposition |
| --- | --- |
| Transaction, durability, recovery | Pass. Per-connector exclusive publication locks serialize writers. A same-parent `.stage-*` tree fsyncs all files/directories, a typed old/new journal is durable before `CURRENT`, pointer replacement is atomic and parent-fsynced, and every injected cut point recovers to one validated old or new closed generation. |
| Read/check and runtime boundary | Pass. `Check` acquires only an existing shared lock and rejects pending recovery rather than writing it. `Open` binds one validated `CURRENT` generation before a reader lease. Publication reuses the existing loader/selection/preflight only for physical staged validation; `defs.FS`, App/CLI routing, source-lock runtime reads, and connector behavior remain unchanged. |
| Closed-set and metadata integrity | Pass. The immutable generation ID frames every name/payload, integrity binds sorted digests and byte counts, and closed-tree validation rejects missing, unexpected, nonregular, and symlinked members. Omitted optional artifacts cannot survive in the selected set, so readers cannot combine an old index with new execution bytes. |
| Ownership, filesystem safety, and taint | Pass after one self-review correction. Recovery originally followed a symlinked `generations/` root and deleted an author-owned staged target; the new red/green control requires `Lstat`-verified nonsymlink generation roots before recovery, staging, check, open, or prune. Pruning additionally requires a closed-tree integrity proof and an exclusive generation lease, retaining unowned and held directories. No provider, credential, database, source-lock, `.cache`, or certification input was read. |
| Concurrency and resource bounds | Pass. The barrier proof prevents writer interleave and mixed generation observation; the reader lease blocks stale prune. File digesting streams through a fixed buffer; control metadata has a 1 MiB limit; no cross-connector lock or transaction exists. |
| CLI/docs/website parity | Pass. `connectorgen lock-render` syntax is unchanged and its observed help remains accurate. Canonical authoring docs, user guide, architecture/migration references, website delivery/install pages, and generated website docs all describe the closed generation/check-only contract. PM CLI help/manual parity is not applicable because no PM command, flag, output, or runtime behavior changed. |
| Proof and baselines | Pass for the changed authoring path: full `cmd/connectorgen`, race state-machine suite, source-lock target, canon/Atlas checks, affected runtime packages, vet/build/docs/agent-contract/definition validation, and website generation/tests passed. The separately required `internal/cli` suite reproduces only the recorded polling-help, source-bound-origin, ETL-help, and local Redis-connection-refused baselines; CP11 does not modify that package. |

- Findings: no remaining local CP11 blocker. The symlinked-generation-root defect
  found during this self-review was converted into an executable red/green
  regression before the final package/race checks. No actual connector/source
  lock or flat execution artifact was materialized, and no author-owned file
  is deleted by the publisher.
- Required next gate: commit and normally push this exact candidate, verify PR
  #4294 base/head/SHA through the API, then pause for Firstmate's independent
  exact-SHA review. CP12 remains prohibited.

## 2026-09-04 — CP11 exact-review correction local self-review

- Review subject: instruction-126's five bounded corrections in
  `cmd/connectorgen/{main.go,vnext_lock_cli.go,vnext_publication.go}`,
  their temporary-root tests, the source-lock canon, Foundation Atlas record,
  and CP11 evidence. Reviewed from the authorized
  `c4f0bc3728dda318ea3d01f78de7aa299b6135cb` base discipline; CP12 and all
  connector/runtime/provider/credential surfaces remain out of scope.
- Workflow: no compatible isolated GSD reviewer is available, so this is the
  required inline/manual read-only review. It is not independent review and
  does not replace the next Firstmate exact-SHA gate.

| Lens | Disposition |
| --- | --- |
| Root and control confinement | Pass. Connector roots and publication locks are `os.Root`-relative and reject visible symlinks before access. Source-lock/control reads, control-file atomic replacement, and control deletion remain beneath the connector descriptor. Root/lock external-sentinel tests cover every publisher entrypoint and lock-render. |
| Stage ownership and recovery | Pass. A stage marker is written and directory-synced before candidate files, binds version/connector/generation/stage, remains with the generation, and is strictly decoded before cleanup. Missing, malformed, mismatched, or symlinked `.stage-*` entries are preserved and force refusal; a durable valid stage still recovers. |
| Closed-tree identity | Pass. Validation streams the sorted declared artifact bytes through the same length framing as publication and compares the result to `CURRENT`; exact file and directory membership plus an empty lease are required. A copied/renamed tree with rewritten pointer/integrity metadata cannot self-certify. |
| Strict metadata and resource bounds | Pass. One bounded rooted reader fronts CURRENT, JOURNAL, integrity, and the marker; strict duplicate-member/trailing-value rejection reuses the source-lock decoder. Read growth is capped at 1 MiB plus one byte; the control record is rejected before an unbounded decode. |
| Cancellation and resource release | Pass. The production CLI signal context reaches context-bearing publish/check/recover/open/prune entrypoints. Nonblocking advisory-lock retry checks cancellation without a spawned waiter; the temporary-root tests prove deadline return, unchanged CURRENT/JOURNAL, released descriptor/lock state through a successful retry, and the real lock-render path. |
| Runtime and secret boundary | Pass. Publication remains authoring-only over temporary roots; it adds no `defs.FS` reader, source-lock runtime reader, App/PM routing, connector source lock/materialization, provider call, credential lookup, database access, certification read, or cross-connector transaction. No secret value was read or emitted. |
| Canon/Atlas and CLI parity | Pass. Canon and Atlas now name only the implemented confinement, durable stage, exact-tree, strict-control, and cancellation guarantees; catalog revision 34 resolves each owner/proof selector. `connectorgen lock-render` syntax/help is unchanged, so PM/manual/website updates are not applicable. |
| Proof and baseline | Pass for the changed surface: correction selector suite, race suite, full `cmd/connectorgen`, vNext lock target, canon/Atlas checks, affected packages, vet/build/docs/contract/definition validation passed. `internal/cli` remains its recorded unrelated polling/source-origin/ETL-help and local Redis-refusal baseline. |

- Findings: no new local CP11 correction blocker. The remaining mandatory gate is
  a fresh Firstmate independent exact-SHA review after the ordinary correction
  push; do not begin CP12 or amend the reviewed candidate beforehand.

## 2026-09-04 — CP11 correction-review repair local self-review

- Review subject: instruction 127's five accepted corrections and instruction
  128's source-lock-only Atlas scope: descriptor-relative publication,
  lock-render source admission, component-boundary matching, cancellation,
  journal ordering, proof mapping, canon, and evidence. CP12 and all runtime,
  connector, provider, credential, and local-residue paths were excluded.
- Workflow: `execute-phase`, `verify-work`, and `code-review` prompts were
  generated and performed inline. No compatible isolated Pi reviewer is
  available; this read-only review is not independent review and does not
  replace the required Firstmate exact-SHA gate.

| Lens | Disposition |
| --- | --- |
| Descriptor confinement and source admission | Pass. Root, connector, generation, lock, control, lease, stage, rename, pruning, and cleanup paths are no-follow descriptor-relative. Lock-render admits through the retained connector descriptor, rereads under its acquired lock, rejects mutation before `generations/`, and preserves the original inode if the visible connector pathname is replaced. |
| Transaction and recovery order | Pass. The durable prepared journal precedes final rename, the generation descriptor is synced after rename, and both durable cuts recover a complete old or new selection without an unowned deletion. |
| Cancellation and ownership | Pass. Successful nonblocking Flock rechecks cancellation and unlocks before hooks or generation state; owned markers and integrity/lease checks remain prerequisites for removal. |
| Boundary correctness and cost | Pass. Compact aliases are assembled only from complete identifier components, avoiding `canonical` substring matches. Component splitting runs only for lexemes that declare a compact contains alias, avoiding an allocation across the common path. |
| Atlas and documentation scope | Pass. Only `authoring.source-lock-vnext.v1` received publication guarantee mappings. The schema permits the field without imposing it on other Atlas records; the source-lock selector test enforces exact one-to-one resolving pairs. Canon text matches the pre-admission reread and journal ordering. |
| Runtime, secret, and parity boundary | Pass. No source lock is added to runtime, no PM/CLI flag/help/output changes, no provider/credential/database I/O, no connector materialization, and no local residue access occurred. `connectorgen --help` still reports the existing lock-render syntax; canonical authoring docs changed because the authoring behavior changed. |

- Findings: no remaining local blocker. Full generator normal/race suites, vNext
  publication gate, definition validation, actual boundary scan, vet/build,
  docs, contract, Atlas, module-tidy, whitespace, and scoped secret checks
  support this disposition. The two documented unrelated package baselines are
  not called green.
- Required next gate: commit and normally push the exact candidate, API-read
  PR #4294 base/head/SHA, then pause unchanged for Firstmate's independent
  exact-SHA CP11 review. Do not start CP12.

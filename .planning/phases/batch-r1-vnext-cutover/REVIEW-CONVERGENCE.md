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

## 2026-09-04 — CP11 final repair review intake (instructions 130–131)

- Immutable review subject: `f7a325aec3594635acbd27e39099640283ca3663`;
  Firstmate's fresh repair review is **BLOCK**. Only F1 temporary-control
  identity, F2 validated-object cleanup identity, F3 lock-inode serialization,
  and F4 behavior-granular source-lock Atlas proofs are authorized.
- The review report is
  `data/cli-batch1-cp11-repair-final-review-r1/report.md`. Its explicit
  regressions—not a replacement design inferred before intake—are the test
  contract. `130.msg` was fully read and archived under the `131.msg` stop
  instruction before this record.
- Required next disposition: exact RED/GREEN witnesses, a narrow self-review,
  one normal candidate push, and another Firstmate-managed independent
  exact-SHA review. No CP12, review-role spawning, rebase, force push, or
  release action is allowed.

## 2026-09-04 — CP11 final repair manual self-review (F1–F4)

- Review mode: inline/manual fallback after the generated `execute-phase`,
  `verify-work`, and `code-review` prompts. Firstmate prohibits review-role
  spawning and no compatible isolated reviewer is available. This is self-review
  only; a new Firstmate exact-SHA review remains required after push.

| Lens | Disposition |
| --- | --- |
| F1 temporary controls | Pass. `writeAtomicLocked` retains the temporary descriptor through `renameBound`; a failed rename cleanup is itself identity-bound, so a replacement cannot be unlinked. |
| F2 destructive cleanup | Pass. Stage and generation roots remain open through ownership/integrity/lease validation and `removeTreeBound`; controls retain/load an exact identity through validation and `removeRegularBound`. The obsolete unbound regular-removal path was deleted. |
| F3 serialization | Pass after self-review correction. The companion anchor is linked only with `O_EXCL` lock creation; reconstructing it for an existing pathname would allow a replacement inode to establish a second domain. The post-acquisition test covers a removed anchor as well as a visible lock replacement. |
| F4 Atlas and staged preflight | Pass. The one-guarantee validator remains gated to `authoring.source-lock-vnext.v1`; valid owned-stage and unowned-generation tests map only to their corresponding claims. The physical fixture reaches real `commandrunner.Preflight` for an implemented command without provider or credential I/O. |
| Boundary and maintenance | Pass. Publication-local descriptor helpers add no generic framework or dependency. Canon/Atlas/README language names the anchor lifecycle, temporary/control cleanup identities, and behavior-granular proof rule. No CLI/help/runtime/provider scope widened. |

- Findings fixed during review: (1) an existing lock without an anchor must
  refuse instead of reconstructing an anchor for a possible replacement inode;
  (2) `removeRegular` was obsolete after all cleanup moved to
  `removeRegularBound`. No remaining local finding.
- Evidence: focused F1–F4/Atlas tests, full normal generator suite, final full
  race generator suite, vNext corpus gate, 553-definition validation, docs,
  vet, tidy, and agent-contract checks are recorded in `VERIFICATION.md`.
- Required next gate: one ordinary commit and push of this exact candidate,
  PR #4294 API base/head/SHA read-back, status update, then pause unchanged for
  Firstmate's independent exact-SHA review. CP12 remains prohibited.

## 2026-09-04 — CP11 exact-review continuation intake (instruction 132)

- Fresh independent exact-SHA review BLOCKed `d3661661dbd1646376e0fbae6d73ab658532a153`; the report is `data/cli-batch1-cp11-repair-r2-final-review-r1/report.md` and was read in full. It authorizes only F1 late rename-source binding, F2 late cleanup/quarantine binding, F3 matched sibling lock-pair serialization, and F4 behavior-appropriate source-lock-vNext Atlas witnesses.
- The review's concrete defect statements, not the prior candidate's passing suite, are the completion contract. In particular, a pathname precheck before `renameat`/`unlinkat` is not identity binding; a missing-anchor test is not a matched-pair test; and a resolving Atlas symbol is not evidence that its mapped physical/durability claim actually ran.
- Manual workflow fallback remains required: the GSD adapter's known issue prompt is absent, Firstmate forbids role spawning, and no compatible isolated reviewer is authorized. The final local review will be explicitly self-review only and cannot replace the next Firstmate exact-SHA review.

## 2026-09-04 — CP11 exact-review continuation local self-review (instruction 132)

- **F1:** `writeAtomicLocked` now creates the source only as `control` in a retained private directory. `renameBound` rechecks that exact source before descriptor-relative rename; deferred cleanup cannot unlink a late replacement.
- **F2:** regular controls and tree cleanup use the same quarantine protocol: retain validated descriptors, move the public entry into a private descriptor-confined directory, recheck candidate and lease identities, then delete only the private candidate. A mismatch restores the candidate when safe or leaves both objects intact.
- **F3:** the operation lock is a duplicated retained connector-directory descriptor. No production code reads, creates, or trusts `.connectorgen.lock`/anchor siblings, so a self-consistent replacement pair cannot split the Flock domain.
- **F4:** the source-lock-vNext Atlas now maps only physical/durable behavior witnesses for closed trees, publisher-written stages, late cleanup, and matched pairs. The in-memory comparator is no longer a publication-guarantee witness.
- **Scope/maintenance:** no runtime route, source lock, rendered execution JSON, provider, credential, database, cache, certification residue, release automation, or CP12 surface changed. Full normal/race package, corpus, definition, docs, static, contract, build, and diff gates are recorded in `VERIFICATION.md`. No additional local finding remains; fresh Firstmate exact-SHA review is still mandatory after ordinary push.

## 2026-09-04 — CP11 final F1 continuation intake (instruction 133)

- Fresh independent review BLOCKed `958a07a778fba6264d1aec567efa5d8c853eefa2` on one remaining F1/F4 defect. Its exact report was read in full. F2 quarantine cleanup and F3 connector-directory serialization are accepted; reopening either is out of scope.
- The defect is a final source check/use race: `renameBound` checked the private source a second time, then consumed its pathname with `renameat`; the old hook ran before the second check, so the claimed refusal never exercised the final boundary. The correction must preserve a prior control, verify the installed inode, and restore both sides on mismatch.
- Inline/manual GSD remains required because the adapter's known issue prompt is absent and Firstmate forbids role spawning. Local review cannot replace the required fresh exact-SHA review after normal push.

## 2026-09-04 — CP11 final F1 continuation local self-review (instruction 133)

- **F1 transition:** the final hook follows the last private-source identity assertion and immediately precedes the installation call. Existing-target installation hard-links the verified prior control into a private quarantine, verifies the installed inode, hard-links any mismatch before atomically restoring the exact prior inode. The no-prior path uses a non-overwriting link and retains a mismatch privately rather than replacing an unexpected public target.
- **F1 witness:** the new test drives both `CURRENT` and `JOURNAL` through that exact hook, proves the restored inode is the original control, and locates the mismatched payload under a quarantine rather than merely checking an error string. The old pre-final-boundary witness remains unmodified; it no longer anchors the F1 Atlas guarantee.
- **F4 mapping/canon:** only the F1 source-lock-vNext guarantee text, proof-test registration, negative mapping, and its canon explanation changed; F2/F3 mappings and production code remain untouched.
- **Scope/verification:** manual source review found no remaining F1 defect. Final normal/race package, focused witness/Atlas, corpus, definition, docs, static, contract, build, tidy, whitespace, and command-smoke evidence is recorded in `VERIFICATION.md`. Fresh Firstmate exact-SHA review remains required after one ordinary push.

## 2026-09-04 — CP11 durable F1 continuation intake (instruction 134)

- Fresh independent review BLOCKed `2f381433f95fb180b00eb258539fc71bf6256737` because the prior backup, substituted public control, restore, and record-free return have no durable transaction authority. The synchronous F1 seam is correct but cannot establish crash safety. F2/F3 are accepted and must not change.
- Required repair: a typed fsynced control-repair authority must identify target, intended/prior/replacement identities, repair storage, and phase; it must be recovered before any ordinary `JOURNAL`/`CURRENT` parse. A pending authority makes the public control non-authoritative, and its recovery must retain a replacement, durably restore prior/absence, then clear authority only after that resolution.
- The review gate is the matrix in `TDD-LEDGER.md`, not the prior synchronous-only pass. The final local review must enumerate every crash cut and independently verify that no cut makes a substitute ordinary authority.

## 2026-09-04 — CP11 durable F1 crash-cut self-review (instruction 134)

- **Pre-exposure:** `writeAtomicLocked` only reaches the final private-source barrier after `beginControlRepairLocked` has descriptor-bound the optional prior, fsynced its quarantine, and atomically fsynced a strict typed repair authority under the retained connector root. A pre-authority backup fault leaves the ordinary public control untouched; every later cut has authority.
- **Authority ordering:** `recoverLocked` calls `recoverControlRepairLocked` immediately after lock binding and before its first `JOURNAL` read or `CURRENT` decode. A malformed pending public control therefore cannot turn a parse failure into an ordinary recovery decision. An invalid authority or bound storage mismatch refuses before public-control interpretation.
- **Existing/no-prior resolution:** for a bound prior, recovery retains any non-prior public inode first, then restores the verified hard-link prior and fsyncs the connector namespace. For no-prior publication it retains any installed public inode, unlinks it descriptor-bound, and fsyncs the valid absent namespace. The same resolver handles a live mismatch and a fresh restart.
- **All durable cuts:** test-only barriers cover initial backup sync; prepared-authority sync; raw install; public install sync; replacement-link sync; replacement-authority sync; public-restoration sync; restored-authority sync; and authority-clear sync. Every barrier is run against `CURRENT` and `JOURNAL`, with and without a prior. Fresh recovery proves valid old/new-or-absent state, cleared authority, and retained substitute where present.
- **Clear/forensics:** authority removal follows public namespace sync and, on mismatch, replacement retention plus repair-state sync. Cleanup may only remove an empty repair directory after authority clear; a nonempty replacement directory persists as forensic evidence. F2/F3 routes, provider I/O, credentials, and connector execution are not part of this correction.

## 2026-09-04 — CP11 monotonic-authority redesign intake (instruction 136)

- Captain authorizes exactly one F1 redesign from immutable candidate `f36b5d0a275ed27fd5f4da242ba192e43f8066d5`. The review target is the post-exposure mutable pathname used to replace the sole restart authority, not F2/F3 or CP12.
- Review must disconfirm the false equivalence between cooperative `Flock` serialization and adversarial same-permission pathname integrity. It must also verify that a replacement cannot erase, replace, or redirect the only recovery authority, including across fresh restart cuts.
- The required end review is read-only over the complete candidate-to-new range, records concrete evidence versus hypothesis, and stops rather than recursively redesigning if a remaining F1 invariant needs CP12 scope.

## 2026-09-04 — CP11 F1 redesign research disposition (instruction 136)

- **Evidence, not assumption:** candidate `f36b5d0a` has the right pre-decode ordering but persists its sole authority at `.connectorgen-control-repair.json` and replaces that name in `updateControlRepairLocked` after target exposure. The retained directory-descriptor `Flock` serializes cooperative publishers, but the new locked source/target test performed same-permission `os.Rename` successfully in all four `CURRENT`/`JOURNAL` cases.
- **Concrete precedent:** warehouse history `fcff76a7305fe469c3903c33a89bc47912852ac6` and `migrateWarehouseIdentity`/`allocateUniqueIdentity`/`LocationFor`/`EnsureOwnership` prove the applicable principle: generated structural identity creates a private region, and a conflicting identity is refused rather than rewritten. F1 has no warehouse workspace/connection domain; importing one would be unrelated scope. Its dedicated transaction uses the same random, descriptor-bound allocation pattern as the existing quarantine, without reusing that object.
- **Reviewable decision:** replace the mutable root authority with immutable prepared state inside that private transaction region and append only create-only identity/digest-chained phases. Recovery must derive the latest verified phase before public controls; source/target substitution cannot name or replace the transaction authority. No manual-unlock surface or legacy fallback is permitted. The final read-only review must specifically attempt to disconfirm the private namespace, phase-chain, pre-decode, cleanup, and public-substitution invariants over the full `f36b5d0a..new` range.

## 2026-09-04 — CP11 monotonic-authority redesign local self-review (instructions 136–137)

- **Review range and mode:** read-only review of the complete `f36b5d0a275ed27fd5f4da242ba192e43f8066d5..candidate` range: all ten scoped paths, the final production transition/cleanup code, all new F1 matrices, the source-lock-vNext canon, and the sole affected Atlas entry. This inline review is the required local fallback because Firstmate prohibits role spawning; it does not substitute for the fresh independent exact-SHA review after the ordinary push.

| Lens | Disposition |
| --- | --- |
| Authority monotonicity | Pass. The old root `.connectorgen-control-repair.json` writer/update route is gone. A unique connector-local directory holds immutable `prepared.json`; phase records are exclusive-create, sequence-bounded, digest/identity/predecessor-bound, and strictly verified as one contiguous chain. |
| Pre-exposure / public boundaries | Pass. Prepared creation and parent fsync precede public installation. The final source barrier reasserts private transaction/prepared identity immediately before installation; public target substitutions at preparation and final cleanup retain/restore rather than clear the authority. The real `Flock`/same-permission rename-unlink witness distinguishes cooperative serialization from pathname integrity. |
| Restart / durable ordering | Pass after review correction. Recovery resolves a validated private authority before any ordinary `JOURNAL`/`CURRENT` decoding. The new authority-retirement cut proves `prepared.json` is fsynced away before phase/backup disposal; earlier cuts retain all material necessary to restore prior or valid absence. |
| Cleanup / forensic retention | Pass. Every private/public deletion or final directory removal retains the expected descriptor identity. Replacement records remain private forensic state; empty completed transactions are removed only through their bound connector entry. |
| Concurrency and external behavior | Pass. The existing matched-lock-pair serialization proof remains green. No new command, flag, runtime route, connector behavior, manual-unlock mechanism, compatibility reader, credential/provider/database path, F2/F3 behavior, or CP12 surface was introduced. |
| Canon / Atlas proof | Pass. `authoring.source-lock-vnext.v1` alone records the revised contract, owner symbols, durable positive proof, and distinct pre-exposure refusal proof. Canon says exactly the same private-transaction/authority-retirement ordering. |
| Quality gates | Pass except the known repository-wide full-lint debt: full `make lint` reports 43 existing findings, while the exact `--new-from-rev f36b5d0a275ed27fd5f4da242ba192e43f8066d5` lint check reports zero. Focused/full/race package, corpus, definition, boundary, preflight, canon, docs, vet, build, tidy, contract, Atlas, help, and secret checks are recorded in `VERIFICATION.md`. |

- **Findings resolved during this review:** (1) a substituted prepared inode could be detected only after public installation, so the final pre-install private identity assertion and regression were added; (2) phase/backup cleanup preceded durable authority retirement, so retirement was reordered and its crash cut added; (3) new close-error handling was made explicit after changed-line lint identified it.
- **Conclusion:** no remaining CP11-local blocker found. Instruction 137 keeps CP12–CP16 and later connector proof work out of this candidate. Next step is a scoped ordinary commit/push and Firstmate-managed independent exact-SHA review; do not start CP12.

## 2026-09-04 — CP11 F1 Design B linearization review intake (instruction 139)

- **Review subject:** correction begins only from `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`, parent `f36b5d0a275ed27fd5f4da242ba192e43f8066d5`. The authoritative report mandates Design B with Design A's no-clobber capture primitive.
- **Root finding:** candidate final checks do not linearize later destructive public `renameat`/`unlinkat`, then delete sole prepared authority. A same-permission direct actor can destroy a late occupant or make recovery/check trust an authority-free third inode; public-only check masks pending no-prior `JOURNAL` behind old valid `CURRENT`.
- **Review predicates:** no production clobbering rename over public control; no public unlink to select absence; protected mode never has zero terminal heads; successor durability precedes predecessor retirement; authority-first read-only check precedes public decode; malformed repair-prefix directory refuses; typed no-replace errors have no fallback; repeated substitutions retain distinct identities; cleanup cannot weaken current terminal; same-UID privilege boundary is documented.
- **Proof standard:** candidate REDs execute after actual final validation. Green tests use real descriptor-relative operations, fresh publisher recovery, direct lock-ignoring actor, inode assertions, and production `lock-render --check`; exit status, bytes alone, authority counts, or test-only checker are insufficient.
- **Process:** Firstmate forbids role spawning, so generated GSD workflow is recorded inline. This local intake cannot replace fresh Firstmate-managed OMP exact-SHA review after bounded correction push.

## 2026-09-04 — CP11 F1 Design B local exact-range self-review

- **Subject and method:** read-only review of
  `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8..WORKTREE` across the publisher,
  descriptor helpers, platform no-replace shims, adversarial tests, canon,
  Atlas, and CP11 evidence. GSD sources and the generated discuss/plan/execute/
  verify/review prompts were re-resolved; the established inline/manual fallback
  remains necessary because Firstmate forbids role spawning. This is not the
  required independent exact-SHA review.
- **Authority and recovery:** protected mode cannot have a missing terminal
  `CURRENT` or `JOURNAL` head. Every successor binds the predecessor's exact
  terminal descriptor/digest/selection; fork, cycle, gap, malformed, pending,
  retry-required, and divergent authority fails before public decoding.
- **Namespace safety:** public mutation uses only no-replace capture and
  create-only linking. The design does not retain a public clobbering rename or
  public unlink path for controls. Captured inodes, prior/intended anchors, and
  terminal authority remain reachable across fresh recovery.
- **Read safety:** review found a post-scan private-state replacement window in
  the first Design B reader. It was fixed by graph-wide private-identity
  revalidation immediately before every authorized public read and covered by a
  direct shared-check substitution witness. Public controls are decoded from
  the already-open descriptor only after that check.
- **Cleanup and privilege boundary:** no automatic predecessor deletion exists,
  deliberately making the report's optional cleanup move/post-move barriers
  unreachable rather than unsafe. The retained-predecessor substitution proof
  verifies no authority is deleted. Canon documents that this is integrity
  protection against public-name interference, not authentication against an
  arbitrary same-UID private-state attacker.
- **Verification and scope:** final normal/race package suites, corpus,
  definition, Atlas/canon/docs, vet/build/tidy/contract, and boundary gates
  pass as recorded in `VERIFICATION.md`. No runtime reader, source lock,
  rendered execution artifact, provider/credential/database path, CP12 change,
  `.cache`, certification residue, reset, rebase, or force-push entered scope.
- **Conclusion:** no remaining CP11-local finding. Commit and ordinary push one
  coherent candidate, read back PR #4294 base/head/SHA, archive the handled
  instruction, and pause unchanged for Firstmate's fresh exact-SHA review.

## 2026-09-05 — CP11 Astra B-01/B-02 frozen correction intake

- Immutable reviewed candidate: `8214bd91403ce620773b61caf674faa540ee1701`; immediate parent `4fa9a5b8cdecdfc07afe54ee3eddb7d19719f5b8`; review source `data/cli-batch1-cp11-astra-review-r2/report.md`. This is the complete current CP11 blocking ledger. L-01 is explicitly optional and excluded.
- Custody: Firstmate owns the independent Astra re-review. Native OMP task fan-out is disabled and no direct reviewer/fixer role may be created. The local worker records the required inline GSD/TDD/review procedure and will freeze one coherent corrected SHA without pushing it.
- B-01, source-derived finding pending executable RED: `vNextPublicationRestoreQuarantined` observes absence then calls clobbering `renameFrom`; a second C at that public name can be overwritten while B is restored from quarantine. `activateStageLocked` likewise plain-renames a stage over a late empty final-generation directory. Reachable callers are stale-stage recovery, stale-generation prune, failed-active-validation rollback, and final activation. Violated contract: every cleanup/activation boundary preserves every observed identity. Required exact proof is C-public/B-private/A-reachable inode-and-byte preservation, not only an error.
- B-02, source-derived finding pending executable RED: a strict, fsynced base `prepared.json` with no terminal phase is valid private authority but `ensureControlAuthorityLocked` rejects it before ordinary recovery. Reachable path is first publication bootstrap after `prepared`/transaction-sync/connector-sync and before base terminal append; visible behavior is permanently failing ordinary fresh recovery. Violated contract: valid durable bootstrap preparation is resumable under the exclusive lock without weakening read-only check or malformed-state refusal.

| Mandatory lens | B-01/B-02 finding coverage | State before RED |
| --- | --- | --- |
| Architecture/data flow | Quarantine restore, activation, authority bootstrap, recovery, check, and retry paths are mapped in `PLAN.md`. | complete source map |
| Happy/bad/edge behavior | A/B/C replacements, final empty destination, CURRENT/JOURNAL first/second base head, valid/malformed/missing-prepared and retry states are enumerated. | executable proof pending |
| State machine/concurrency | B-01 uses a lock-ignoring actor at the exact syscall seam; B-02 covers pre-marker durable cut, fresh restart, and retry. | executable proof pending |
| Security/authority | No public overwrite fallback; valid base only; strict private graph before public decode; check stays read-only. | complete contract map |
| Error/retry/resume | No-replace collision/unsupported stays typed; ordinary recovery resumes valid base and retains conflict reconciliation. | executable proof pending |
| Output/filesystem integrity | Tests assert inode/bytes/tree reachability for A/B/C and selected authority state, not exit status alone. | executable proof pending |
| Declaration/runtime boundary | CP11 remains authoring-only; no runtime `CURRENT` reader, importer, compatibility path, credential, provider, or database path changes. | complete |
| Canon/Atlas evidence | Existing claims are retained as contract, not passed evidence; behavior-granular proof mappings update only after GREEN. | pending GREEN |
| Tests/evidence | Existing one-substitution and post-bootstrap matrices are retained as regressions but do not cover B-01/B-02. | executable RED required |

- Disconfirming controls: a candidate C-preservation outcome disproves B-01; a malformed/missing-prepared recovery outcome invalidates the proposed B-02 narrowing. Both are recorded as findings, not normalized into the correction.
- Local review prohibition before fix: no behavior code is changed before a deterministic test-only fault seam and its executable RED are recorded. The later local self-review must inspect the complete `8214bd…new` range, new fault points, all recovery/cleanup callers, canon/Atlas mappings, and every preserved refusal. It is not the required fresh Astra review.


## 2026-09-05 — CP11 Astra B-01/B-02 executable RED update

- The frozen source-derived ledger is now executable: the exact three-selector command exited 1 in 3.250s. B-01 lost C in each real stage/generation/rollback restoration caller and accepted the final-generation activation collision; B-02 refused fresh recovery for both valid durable bootstrap cuts after the read-only check assertion passed.
- This is evidence of the blockers, not a local review verdict. Production behavior remains unmodified pending the bounded Green patch. After Green, inspect the full `8214bd91403ce620773b61caf674faa540ee1701..local-candidate` range, then freeze the SHA for Firstmate-managed Astra re-review without pushing or starting CP12.

## 2026-09-05 — CP11 Astra B-01/B-02 focused GREEN update

- B-01 is corrected without a new primitive: all restore and activation destinations use the established descriptor-relative no-replace path. The physical A/B/C and final-destination witnesses pass, including retained typed collision causes.
- B-02 is corrected through the narrow markerless-base transition only. Exclusive recovery validates a unique phase-empty/no-predecessor equal-prior/intended record, appends its terminal, re-scans, then completes missing base authority/marker; check remains non-mutating and malformed/private/graph refusals remain in the continuity selector.
- Atlas/canon mapping was changed only because the existing one-substitution and post-bootstrap proof names could not substantiate the new exact boundaries. `TestFoundationAtlasSelectorsResolve`, focused normal, continuity, and focused race passed. This is not the final local self-review or the required Firstmate Astra verdict.

## 2026-09-05 — CP11 Astra B-01/B-02 local correction review

- **Reviewed scope:** `cmd/connectorgen/vnext_publication.go`, `vnext_publication_repair.go`, their physical/durable tests, `SOURCE-LOCK-VNEXT.md`, the existing `authoring.source-lock-vnext.v1` catalog record, and CP11 planning evidence. No connector definition, rendered execution artifact, runtime reader, provider, credential, database, `.cache`, or certification-residue path is in the correction.
- **B-01 decision:** `restoreQuarantinedLocked` validates B, observes public absence, then uses `renameNoReplaceFrom`; a C created after that observation cannot be replaced. The refusal retains both its identity cause and the typed collision cause. `activateStageLocked` uses the same primitive, so an unvalidated late final-generation directory remains rather than being overwritten. The witnesses exercise real stage cleanup, prune, rollback, and activation paths with A/B/C inode and byte checks.
- **B-02 decision:** markerless recovery permits one state only when scanner-validated private identity, no predecessor, zero phases, and equal prior/intended all hold. It appends the existing strict committed terminal phase, closes/re-scans, creates a missing base head only afterward, and writes the marker only after both terminal heads exist. Successors, nonempty phases, malformed/missing private records, graph failures, and read-only check remain fail-closed.
- **Proof/documentation:** prior mappings only named one-substitution and post-bootstrap tests, so they could not substantiate the new exact boundaries. The catalog now registers/matches the A/B/C, final-activation, and strict-base tests one guarantee at a time; the canon describes the same bounded protocol. No new foundation or compatibility route is introduced.
- **Verification/finding:** focused, continuity, focused race, full normal/race package, canon, definition, docs, static, help, diff, and explicit changed-path scan gates passed as recorded in `VERIFICATION.md`. No local actionable finding remains. This review is evidence only; the candidate remains BLOCKED until Firstmate-managed Astra exact-SHA review.

## 2026-09-05 — CP11 Codex successor merged immutable review ledger

### Review identity and boundaries

- **Implementation reviewed:** `36e4d980de0d51d92fe74a68306845643596a6cb..0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e` (11 linear commits, 37 changed paths). The candidate is exactly two commits after published/rejected `8214bd91403ce620773b61caf674faa540ee1701`; `403b9973ae29cb596b4829f92a214aad7cad805f` contains the code correction and `0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e` is decision evidence only.
- **Independent inputs:** separately assigned Astra/xhigh publication and bootstrap contexts reviewed the entire range read-only under the supplied B-01 and B-02 lenses. Neither ran tests, builds, generators, linters, provider/credential/database work, GitHub mutation, or project edits. The reports agree on the single root below; duplicate labels `PUB-01` and `R3-B02-01` are one causal finding, not two.
- **Historical execution boundary:** the phase's B-01/B-02 RED/GREEN and normal/race package results are retained writer-recorded evidence only. They are not newly run or independently certified here. This static ledger does not claim provider-live, native Linux runtime, power-loss, or full-programme certification.

### CP11-R3-01 — Medium blocker: a FIFO can indefinitely block a required regular-file read

- **Visible symptom:** a local `connectorgen lock-render`, `Recover`, or `lock-render --check` can remain blocked while retaining the connector publication lock, instead of refusing a nonregular control/stage member. It emits no bounded typed refusal and context cancellation does not interrupt the already blocked syscall.
- **Initiating trigger:** a same-permission actor places a no-writer FIFO at a descriptor-relative member that is later consumed through `openRegular` (for example a stale stage `.connectorgen-stage.json`, public `CURRENT`, or `JOURNAL`). No provider, credential, private-authority forgery, or cross-connector action is needed.
- **Masking condition:** all current regular-only fixtures are regular files and existing cancellation tests cover lock acquisition rather than a blocked open. `O_NOFOLLOW` protects a symlink boundary but does not make FIFO `O_RDONLY` opens nonblocking.
- **Reachable source path:** `vNextPublicationDirectory.openRegular` calls `openFile` at `cmd/connectorgen/vnext_publication_dir.go:204`; `openFile` calls `unix.Openat` without `O_NONBLOCK` at line 175; only after it returns does `file.Stat()` reject a nonregular object at lines 208–216. `removeStageLocked` → `validateStageOwnerLocked` → `vNextPublicationReadControl` reaches it during `Recover`/publication. `recoverControlRepairLocked` → `vNextPublicationControlStateMatchesPublic` and `vNextPublicationReadAuthorizedControlLocked` reach it for public `CURRENT`/`JOURNAL` reconciliation and `--check`. `vNextPublicationDirectoryFS.Open` at line 455 is a distinct required sibling: it delegates raw `openFile` to `fs.ReadFile` consumers during `vNextValidatePublishedStage`/`engine.Load` after member enumeration, bypassing `openRegular`; a regular-to-FIFO substitution at that final delegated open can otherwise still block or expose a nonregular descriptor.
- **Invariant violated:** malformed or unowned publication members must fail closed within bounded local I/O, preserve the object and active generation, release the operation lock, and never turn a path precheck into a substitute for same-descriptor validation. The candidate's declared nonregular-file refusal occurs too late to uphold that liveness boundary.
- **Required RED/GREEN proof:** use a bounded subprocess because the candidate can hang. Publish a valid temporary baseline, inject a no-writer FIFO at the actual member immediately before the read, invoke production recovery/publication and real `lock-render --check`, and assert prompt nonregular refusal, no success output, unchanged FIFO/private-authority and active-generation identities/bytes, then release of the lock for a subsequent operation. Cover representative stage-owner and public-control callers, a regular-file-to-FIFO substitution, and the actual semantic-admission filesystem path after enumeration but before its delegated `Open`. A pathname `Lstat` precheck alone is insufficient; the final opened descriptor must be the object type-validated.
- **Audited constrained repair direction:** a descriptor-relative no-follow, nonblocking ordinary read followed by exact-opened-descriptor type validation is required. `vNextPublicationDirectoryFS.Open` must preserve legitimate nested directory traversal while rejecting FIFOs and all other disallowed members before exposing a reader. The repair must preserve portable Darwin/Linux flag/error behavior, ordinary-file reads, missing/symlink refusals, exclusive writes, leases, byte bounds, and error wrapping; it does not promise bounded completion of arbitrary device I/O or add a new runtime foundation, decoder, or recovery command.

### Resolved findings and non-blocking scope

- **B-01 (resolved by source inspection, pending fresh exact-SHA confirmation):** stage activation and quarantine restoration now end in `renameNoReplaceFrom`. The new real A/B/C stage/prune/rollback witnesses place C after the previous absence observation and require C public, B quarantined, A reachable, and `fs.ErrExist`; activation retains both the stage and a late destination collision. No plain-rename fallback is present at either final public mutation boundary.
- **B-02 (resolved by source inspection, pending fresh exact-SHA confirmation):** markerless recovery accepts exactly one scanner-validated, phase-empty, no-predecessor base whose prior equals intended, appends its terminal under the existing exclusive lock, rescans, then establishes the other base head and marker. The new CURRENT-first/JOURNAL-second witness exercises fresh recovery, non-mutating `--check` during interruption, preserved public absence, and normal retry. It does not justify treating malformed, missing-prepared, successor, gapped, forked, or terminal-divergent graphs as recoverable.
- **L-01:** the unused operation-control cache is optional, has no observed decision consumer, and remains excluded from this correctness group.

### Mandatory lens disposition

| Lens | Disposition |
| --- | --- |
| Architecture/data flow | Complete — authoring-only `lock-render` admission/publisher/authority paths traced; no runtime CURRENT reader or connector route added. |
| Happy/bad/edge behavior | Complete — regular, missing, symlink, FIFO, replacement, malformed controls, B-01 A/B/C, and B-02 interrupted base paths were audited; CP11-R3-01 remains an actionable defect. |
| State, concurrency, cancellation | Complete — FIFO can hold shared/exclusive publication locks before bounded recovery transitions; normal lock/graph behavior was traced. |
| Security and taint | Complete — the hostile local object boundary is the defect; no credential/provider path is introduced. |
| Retry, rate, resume, idempotency | Complete — FIFO blocks before bounded recovery can refuse; provider semantics are not applicable to CP11. |
| Output/bytes/filesystem integrity | Complete — candidate lacks a bounded nonregular refusal; B-01 inode/byte preservation remains source-verified. |
| Declaration and closed-surface reachability | Complete — no new declared operation, generic executor, runtime reader, source lock materialization, or compatibility route. |
| CLI/App parity | Complete/applicable scope — `connectorgen` success output follows successful check/publish only; PM/App surfaces are unchanged. |
| Provider semantics | Not applicable with evidence — source locks, transports, credentials, and database behavior are outside this authoring-only range. |
| Tests/evidence | Complete static audit — candidate has no FIFO regression, and inherited B-01/B-02 execution is carefully bounded above. |

### Required next gate

- This ledger was frozen at `0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e` and independently audited by a separately assigned Astra/xhigh context. The audit affirms CP11-R3-01, adds the `vNextPublicationDirectoryFS.Open` semantic-admission sibling, and establishes no other independent root. The next authorized action is one GSD/TDD repair group with meaningful bounded-subprocess RED evidence before production code. The full original range plus all repair deltas must receive a fresh independent final exact-SHA review after GREEN; CP12 remains prohibited.

### CP11-R3-01 repair record — pending fresh independent exact-SHA review

- **Root-cause group implemented:** `openRegular` now performs the existing descriptor-relative, no-follow read with `O_NONBLOCK` before validating the exact opened descriptor as a regular file. The semantic-admission sibling is closed by `openFilesystemMember`, which opens no-follow/nonblocking and permits only regular files or real directories before `vNextPublicationDirectoryFS.Open` hands the descriptor to `io/fs`. No create/write/lease call site changed.
- **Meaningful RED/GREEN:** the pre-change bounded subprocess selector exited 1 after a stale-stage owner and enumerated admission member FIFO each exceeded one second before the prior blocking open returned. The repair's final focused selector passes and observes prompt refusal, FIFO preservation, lock release, actual `lock-render --check` no-success behavior for `JOURNAL`, and valid nested directory traversal. Full normal/race package evidence, boundary/preflight/canon/docs/static/release checks, lint-new-line proof, smoke classification, and the inherited `internal/cli` baseline are detailed in `VERIFICATION.md`.
- **New edge/function inventory for final review:** `openFilesystemMember` is the only new production helper; it introduces one allowed-type branch for a directory required by nested staged schemas. `openRegular` adds no state transition, retry, or fallback, only a nonblocking flag before its existing `Stat()` gate. The regression introduces bounded child-process scenarios for stale stage ownership, public `CURRENT`, public `JOURNAL` check, and adapter-after-enumeration substitution. There is no provider-live, native Linux runtime, power-loss, or whole-programme certification claim.
- **Disposition:** CP11-R3-01 is locally repaired with an observable regression, not accepted. B-01 and B-02 remain source-preserved dispositions pending the same independent final review. No scope decision, new foundation, or additional root was discovered; CP12 remains prohibited until a fresh Astra/xhigh context reviews the complete original range plus this committed repair/evidence delta at its exact SHA.

## 2026-09-06 — CP11 R3-02–07 audited frozen ledger (artifact-only checkpoint)

### Exact reviewed identity, custody, and evidence boundary

- **Reviewed code SHA:** `3a455877cdd9686ba6f04341960a3c31196909bd`.
- **Full immutable range:** `36e4d980de0d51d92fe74a68306845643596a6cb..3a455877cdd9686ba6f04341960a3c31196909bd` (14 commits, 37 paths; base is the recorded merge base).
- **Writer:** sole Codex Terra/xhigh project owner. The original fresh Astra/xhigh discovery review did not implement the FIFO repair; the separate GPT-6 Astra/xhigh frozen-ledger auditor did not implement or discover the original candidate; GPT-5.6 Luna/low performed read-only mechanical reconciliation only. All three worked read-only and made no provider, credential, customer-database, test, build, generator, lint, GitHub, or no-mistakes action.
- **Independent inputs preserved outside this candidate:** exact discovery ledger, independent Astra audit, audited causal ledger, and corrected Luna table are retained in Firstmate task data under `data/cli-batch1-pi-takeover/`. This canonical record preserves their operative findings and is the local committed repair input.
- **Certification boundary:** writer-recorded Darwin FIFO 2.200s, normal 132.575s, and race 409.519s evidence is not independent execution. Neither those results nor this audit establish full-suite/full-lint green, native Linux runtime, power-loss, provider-live/customer database behavior, hosted Verify/security, release, or no-mistakes certification.

### Full causal finding/disposition ledger

| ID | Severity | Causal classification | Audited disposition and invariant |
| --- | --- | --- | --- |
| CP11-R3-01 | Medium | Original full-range FIFO root, fixed by `3a455877` | **Production resolved.** Descriptor-relative no-follow/nonblocking opening validates the same descriptor before reading; semantic-admission `FS.Open` receives the matching regular-or-directory gate. R3-05 owns remaining preservation-proof gaps. |
| CP11-R3-02 | Medium | Earlier CP11 discovery miss | **Confirmed.** `scanControlAuthorityLocked` retains every historical transaction directory descriptor (`2 + 4p` for `p` changed publications) for graph identity, so healthy history can persistently reach `EMFILE`. Group with R3-06/R3-07 and consequential ownership/error handling; preserve every authority record. |
| CP11-R3-03 | Medium | Earlier CP11 discovery miss | **Confirmed.** Process-wide `signal.NotifyContext` intercepts SIGINT/SIGTERM for validate/boundary/ownership/gen although they discard the context; `main`'s deferred stop is bypassed by `os.Exit`. Scope interception to lock-render's consumption contract and test production entry point. |
| CP11-R3-04 | Medium | One full-range configured-lint group | **Amended and confirmed.** 31 introduced diagnostics (21 errcheck, 3 staticcheck, 7 unused) are a gate failure; 20 remain proven pre-range programme debt. Error/sync/close handling needs semantics, not blanket ignores/suppression/config weakening. |
| CP11-R3-05 | Low | `fix_created:3a455877cdd9686ba6f04341960a3c31196909bd` evidence gap | **Amended and confirmed.** FIFO tests lack type-safe before/after FIFO inode and relevant generation/private-authority/control snapshots; adapter path does not acquire a publication lock. This is not an unresolved production FIFO claim. |
| CP11-R3-06 | Medium | Earlier CP11 reader-identity miss | **Added.** `openLocked` validates generation A, closes it, reopens generation name B, then returns B with A's file list. Bind/transfer the validated descriptor or refuse; this is authoring API scope, not a PM runtime CURRENT reader claim. |
| CP11-R3-07 | Medium | Earlier CP11 reader-lifetime miss | **Added.** A held reader's replaceable empty lease inode can be moved and replaced with empty B, allowing cleanup to lock B and delete the held generation. Stabilize reader/cleanup lock identity across all cleanup siblings. |

#### Source and sibling map

- **R3-02:** `vnext_publication_repair.go:194,206,213,492,499,540,679,717,734,766`; publication transitions `vnext_publication.go:350,358,374,380`. Graph state descriptors support prepared/phase/capture/anchor/predecessor identity and public-read revalidation. Check/Open/Recover/Prune/Publish all reach scans. A fault-path sibling returns non-nil state plus error after `AfterControlRepairPrepared` and current callers discard it without close; it joins the ownership repair without explaining ordinary history growth.
- **R3-03:** `main.go:40-68,165`; `gen.go:358-392`; `boundary.go:15`; `ownership.go:15`. Default termination must remain for non-consuming command branches. Lock-render cancellation remains lock-acquisition/pre-mutation only, not a new full mid-transaction cancellation promise.
- **R3-04:** introduced paths are `cmd/connectorgen/vnext_publication.go`, `vnext_publication_dir.go`, `vnext_publication_repair.go`, their tests, and `main.go:48`. `main.run` predates CP11 but dispatch commit `4fedb3875cbe7071799aed0e9b6ce1e34257f95e` removed its sole call. Durability-sensitive sites include `writeAtomicLocked:812-817` and failed-creation cleanup `createControlRepairLocked:1236-1247`.
- **R3-05:** `vnext_publication_test.go:2912-3096`, PLAN/TDD records and `VERIFICATION.md:679,686`. A FIFO-safe snapshot must not use `snapshotVNextPublicationTreeForTest` unchanged because its non-directory `os.ReadFile` path would open the FIFO.
- **R3-06:** `vnext_publication.go:449-479,548-555,1184-1207`; local actor replacement occurs after A validation before B re-open/lease binding.
- **R3-07:** `vnext_publication.go:465-470,627-630,1511-1567,1865-1923`; real Open/Prune and Recover/Publish cleanup siblings must hold an identity-stable domain. The late-lease existing test is not the early-empty-replacement case.

### Preserved resolved contracts

- **B-01:** final no-replace quarantine restoration and stage activation preserve A/B/C and late collision identity/bytes with typed `fs.ErrExist`, across stage/prune/rollback/activation callers and Darwin/Linux fail-closed shims.
- **B-02:** only strictly scanned, phase-empty, predecessor-free equal-prior/intended interrupted base authority resumes; revalidate, terminalize, rescan, complete heads/marker. Preserve CURRENT-first/JOURNAL-second check refusal/recovery and all malformed/private graph refusals.
- **R3-01:** retain nonblocking final-descriptor type validation, real nested semantic directory traversal, no-success refusal ordering, byte bounds, strict decoding, and existing lease/create behavior.
- No repair may add a provider route, credential/database action, generic executor, runtime source-lock/CURRENT reader, authority deletion/GC, path-only precheck, hidden fallback, or new shared runtime foundation.

### Mandatory-lens status at this freeze

| Lens | Audited status |
| --- | --- |
| Architecture/data flow | **Complete.** Command dispatch, admission, authority/public controls, generation, Check/Open/Recover/Prune, callbacks and cleanup mapped; R3-06 validation-to-reader handoff added. |
| Happy/bad/edge | **Complete.** Absent/empty/malformed/duplicate/trailing/oversized/partial/symlink/FIFO/replacement/cancellation/interruption cases mapped; required witnesses below remain RED work. |
| State/concurrency | **Complete.** Publication/authority transitions, bootstrap, locks, handles, leases, cleanup and lock order mapped; R3-02/03/06/07 actionable. |
| Security/secret taint | **Complete.** Local object identity is the hostile boundary; no secret/provider taint added. Arbitrary same-UID destruction of all private authority remains outside the retained threat promise. |
| Retry/rate/resume/idempotency | **Complete/applicable.** Healthy-history retry persists R3-02 failure; no provider rate/replay behavior changed. |
| Output/filesystem integrity | **Complete.** Strict control bytes and success/refusal ordering retained; R3-05/R3-06/R3-07 identify missing preservation/identity proof. |
| Declaration/closed surface | **Complete.** CP10 physical admission/preflight remains sole route; no runtime/provider surface added. |
| CLI/App parity | **Complete/applicable.** Connectorgen command branches mapped; R3-03 affects siblings. PM/App routes and four historical red CLI obligations remain outside CP11 behavior. |
| Provider semantics | **Not applicable with evidence.** No provider operation, transport, credential, database, paging, redirect, or remote idempotency behavior changed. |
| Tests/evidence | **Complete static audit.** Existing/required witnesses and causal limits mapped; static audit is distinct from executed evidence. |

### Required coordinated RED/GREEN matrix

| Group | Required observable RED | Required GREEN and preserved negative controls |
| --- | --- | --- |
| A: R3-02, R3-06, R3-07 | Actual-publication retained history in fresh controlled-`RLIMIT_NOFILE` child reaches valid-history descriptor exhaustion; barrier replaces validated A before actual Open; early empty lease replacement while old Open is held permits cleanup loss/broken old read. | Same effective limit completes Check/Open-read-release/Recover/Prune/next publish with returned bytes/history identities; descriptor/transaction/prepared/phase/capture/predecessor replacement refuses; partial prepared-hook cleanup and lock reacquisition prove release; Open binds validated A or refuses; held generation survives all cleanup siblings then prunes after release. |
| B: R3-03 | Production-main non-consuming command remains alive after real SIGINT/SIGTERM/repeat signal, not merely direct context cancellation. | Real signals terminate without success output and without timeout-SIGKILL ambiguity; bounded local-I/O readiness; real flock-held lock-render returns cancellation exit 1/no mutation then succeeds after release. |
| C: R3-04/R3-05 | Full configured lint exposes all 31 in-range diagnostics; current FIFO tests lack safe preservation oracles. | Semantic error/durability/dead-code repair yields zero in-range configured diagnostics without suppression/coverage deletion, leaving 20 debt explicit; FIFO-safe metadata snapshots prove relevant inode/bytes/authority/generation/control preservation and scope lock release to locking cases. |

### Mechanical evidence corrections

- The repair code commit `3a455877cdd9686ba6f04341960a3c31196909bd` changes exactly six paths: `cmd/connectorgen/vnext_publication_dir.go`, `cmd/connectorgen/vnext_publication_test.go`, and phase PLAN/TDD-LEDGER/VERIFICATION/REVIEW-CONVERGENCE. The old `TDD-LEDGER.md:381` six-member fixture is unrelated CP10 test material and is withdrawn as commit-path evidence. The four-file staging wording in `VERIFICATION.md:686` is wrong and requires a later bounded evidence correction.
- The authoritative pre-range source prefixes are `cmd/connectorgen/vnext_admission.go:390` and grouped `cmd/connectorgen/orderedjson.go`, not Luna's initially expanded `internal/connectors` prefixes. The twelve grouped helper identities are unavailable in retained output and must not be invented.
- The four named `internal/cli` failures and 20 pre-range diagnostics are red final-programme obligations, not CP11 green evidence. Old-checkout-looking lint output is cache/presentation metadata, not wrong-target execution, supported by worktree/module/config/path evidence and matching polling blob `cb34243703a64123c1f9b116e7152249c43da98e`.

### Next permitted repair and final-review gate

The full audited repair order is (A) descriptor/reader ownership R3-02/R3-06/R3-07 with consequential close/cleanup semantics; (B) actual-entry signal behavior R3-03 plus `main.run`; (C) remaining full-range in-range lint and R3-05 truthful preservation evidence. Establish RED before relevant code edits, then coherent GREEN/continuity checks. Record every new helper/state/edge for a fresh GPT-6 Astra/xhigh context that implemented none of the repair. That context must review the complete original range plus all repair deltas at final exact code SHA. CP11 remains unaccepted until it returns a zero-blocker verdict with actual CP11 evidence; no later code-bearing commit may rely on an earlier exact-SHA review.

## 2026-09-06 — CP11-R3-07-LATE-LEASE adjudicated sibling (artifact-only checkpoint)

### Exact frozen decision and causal history

- **Decision authority:** Firstmate accepted the independent GPT-6 Astra/xhigh bounded adjudication at [`cp11-late-lease-astra-xhigh-c6194254560ff874ac63e69a6c80dfe9ab06b5e2-21e53fd1a36d4cb5a523c7af70e0f0d64f72fe3a94d86b2896b1d631db5e868f.md`](/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-late-lease-astra-xhigh-c6194254560ff874ac63e69a6c80dfe9ab06b5e2-21e53fd1a36d4cb5a523c7af70e0f0d64f72fe3a94d86b2896b1d631db5e868f.md), SHA-256 `87dffb471993c91a4c9946d218409e4ef328f6ff1f697a991e15c5d0847d1ade`. The disposition is **preserve/fix**, without a new captain or product decision.
- **Frozen source state:** HEAD `c6194254560ff874ac63e69a6c80dfe9ab06b5e2`; complete eight-path tracked diff SHA-256 `21e53fd1a36d4cb5a523c7af70e0f0d64f72fe3a94d86b2896b1d631db5e868f`; audited comparison `3a455877cdd9686ba6f04341960a3c31196909bd`; full review base `36e4d980de0d51d92fe74a68306845643596a6cb`.
- **Scope distinction:** `68fe9dd363225003b8ec2081380a3b8b35afb1cda9014d3d3106ae824c27ca21` identifies only the frozen `vnext_publication.go`/`vnext_publication_test.go` projection. It records the uncommitted deletion of the old late-lease hook/test and removal of the lease binding; it is not the full tracked state and is retained as causal evidence.
- **Historical REDs:** the pre-deletion full package failed `TestVNextGenerationPublisherRefusesLateReplacedGenerationLeaseCleanup` after `245.872s` because `Prune()` returned nil after late lease replacement rather than refusing identity. The deletion snapshot subsequently failed `TestFoundationAtlasSelectorsResolve/authoring.source-lock-vnext.v1` after `240.844s`, because the catalog retained that required executable proof. Neither result is a pass or a basis to weaken coverage.

### CP11-R3-07-LATE-LEASE — Medium, fix-created Group A sibling

The switch to stable generation-directory Flocks correctly fixes reader lifetime, but it also removed the previously captured `.lease` member identity from `removeGenerationLocked`'s pre/post-quarantine bindings. This is a distinct local closed-tree integrity contract. A directory Flock serializes cooperative holder lifetime; it does not prove that a noncooperating late rename/write has not exchanged a member after generation validation and before quarantine.

- **Required coexistence:** retain the directory-inode shared/exclusive Flock for `Open` and all cleanup callers; never restore the replaceable `.lease` inode as the sole lock domain. Independently open the existing empty regular `.lease` through the retained descriptor, capture its same-opened-descriptor identity, and pass that one bounded binding through the shared pre/post-quarantine validation. Its descriptor must close with the same result/error ownership discipline as the R3-02 repair.
- **Protected late boundary:** the old witness opens/validates generation G, captures `.lease` A, then at the nil-production `after_generation_lease_identity` barrier moves A to an extra name inside G and creates distinct B at `.lease`. The actor is after member capture but before generation entry/quarantine movement. G's root identity remains unchanged, so root-only checks are insufficient. On post-quarantine binding mismatch, refusal/restoration must leave both A and B reachable; a later public C must remain public while the A/B tree stays quarantined with typed no-replace collision causes.
- **Independent held-reader boundary:** the retained three-caller `Prune`/`Publish`/`Recover` witness holds a directory lock while an **empty** B replaces the lease and proves old reader bytes, retained G, and ordinary post-release cleanup. It stops at directory-lock contention and is complementary, not an equivalent proof of the unheld late-member boundary. Keep both empty and nonempty B variants at the late-member boundary so identity, not merely size, causes refusal.

### Normative proof obligation and caller matrix

- `docs/connector-canon/SOURCE-LOCK-VNEXT.md` requires both post-quarantine lease-member revalidation/refusal and separately reader-held lifetime behavior. `docs/connector-canon/foundations/catalog.json` maps the declared guarantee *“the generation lease that authorizes a prune is rechecked after quarantine before deletion; a late lease replacement is refused without deleting either lease object”* to positive `TestVNextGenerationPublisherActivatesClosedSetAndDefersHeldPrune` and negative `TestVNextGenerationPublisherRefusesLateReplacedGenerationLeaseCleanup`.
- `TestFoundationAtlasSelectorsResolve` parses every registered proof declaration and requires every publication guarantee to have exactly one registered positive/negative mapping. Retire neither the guarantee nor its negative obligation. If a name changes, update its test declaration and behavior-granular mapping in the same repair; do not map the held-reader test into this slot.
- The common destructive sink is `removeGenerationLocked`/`removeTreeQuarantinedLocked`. Required coverage is explicit `Prune`; no-journal recovery; committed/new-selected recovery; rejected-generation recovery; ordinary `Publish` (including its initial recovery); failed active-validation rollback; and `Open`'s transitive recovery. The exact prune sink must exercise the late member substitution. Reuse existing public `Prune`/`Recover`/`Publish`/rollback fixtures for the other paths and record actual committed-journal and rejected-generation-cut coverage. Read-only `Check` is not a destructive sibling.
- Preserve B-01 `renameNoReplaceFrom` restoration, B-02/private-authority refusal, R3-01 FIFO/type-safe snapshots, R3-02 bounded descriptor ownership, R3-06 validated-generation handoff, and R3-03 signal behavior. No new runtime foundation, provider/credential/database operation, authority history deletion, generic executor, path-only precheck, or mid-transaction cancellation expansion is authorized.

### Required executable continuation

Restore the old test/barrier surgically without reverting the coordinated repair files. First run the restored/equivalent witness against the current nil-binding logic and observe missing identity refusal or object loss; a hook-hit-only failure is insufficient. Then restore the independently bound lease-member guard, prove A/B inode/type/byte and generation/control identity preservation for empty/nonempty B, and rerun the Atlas selector plus affected held-reader/resource/identity/FIFO/signal suites. The corrected test must remain in the final complete-range Astra/xhigh review together with its temporary deletion and the two failed full runs as causal history.

## 2026-09-06 — fresh Astra/xhigh final review of repaired CP11 code

**Verdict: changes required; CP11 remains unaccepted.** A fresh read-only
`gpt-6-astra`/xhigh context that did not implement the repair reviewed code
`7294373166db75466e2c92269f7887f51ceaddc6` over the complete immutable range
`36e4d980de0d51d92fe74a68306845643596a6cb..7294373166db75466e2c92269f7887f51ceaddc6`.
It read the required briefs, historical phase records, audits, canon/catalog,
and mechanical validation index; inspected all 37 range paths and the complete
publication/authority/directory/rename implementations and tests. It made no
mutation and ran no tests, builds, generators, linter, provider/database, or
no-mistakes operation. The review started from a clean worktree at the named
code SHA. This record is an artifact-only report; it neither changes that code
nor makes a provider-live/full-programme certification.

### Findings requiring one coordinated follow-up repair scope

| ID | Severity/classification | Location, invariant, and observable regression |
| --- | --- | --- |
| F-01 | Medium; remaining R3-02/R3-06 identity-handoff sibling | `cmd/connectorgen/vnext_publication_repair.go:1523`; sibling validation paths `vNextPublicationValidateCaptureLocked:981` and `vNextPublicationValidateCapturedCandidateLocked:1045`. Capture A is validated and closed, then its pathname is reopened without checking that opened directory against recorded `capture.Identity`; a lock-ignoring actor can replace it with B before a public control is moved into B. Bind/retain the checked descriptor through dependent validation/capture. A regression must install distinct B immediately before the dependent open and require refusal before public-control movement plus unchanged A/B identity/content. |
| F-02 | Low; `fix_created:7294373166db75466e2c92269f7887f51ceaddc6` | `cmd/connectorgen/vnext_publication_dir.go:419`. `removeTree` removed `defer child.Close`; its `Fstat` and identity-mismatch returns at lines 425/428 bypass the lone close at 431. A nested replacement during recursive stage/generation cleanup leaks the descriptor. Repeat actual cleanup refusals with deterministic replacement; preserve refusal/objects and prove descriptor count bounded without finalization. |
| F-03 | Medium; remaining R3-04 writer-error-contract miss | `cmd/connectorgen/vnext_publication.go:839` (`writeAtomicLocked`). The writable temporary's successful `Close()` error is discarded after `Write`/`Sync`; `writeCurrentLocked`, `writeJournalLocked`, publication, and recovery can miss the outcome. Inject writable-close failure after write/sync and require observable error with recoverable durable state. `vnext_publication_repair.go:1277` failed-creation cleanup also discards cleanup/sync/close outcomes while returning a primary error; it is a related disposition, not independently asserted false success. |
| F-04 | Low; residual `fix_created:3a455877cdd9686ba6f04341960a3c31196909bd` R3-05 oracle gap | `cmd/connectorgen/vnext_publication_repair_test.go:1674`. The snapshot skips stable nonregular `Lstat` values but then uses `os.ReadFile(path)` at 1684; a regular-to-FIFO/symlink replacement between classification/open can block or read unrelated bytes, and recursion has the same reopen window. The oracle needs no-follow/nonblocking opened-descriptor identity before bytes; a replacement regression must refuse boundedly without reading replacement/external target. This is not a remaining production R3-01 FIFO defect. |
| F-05 | Low; repair-wave proof gap | `cmd/connectorgen/vnext_publication_test.go:74,163`. `OpenKeepsValidatedDirectory` proves returned A bytes but not A inode continuity/B identity or bytes. Held-generation checks prove old reader bytes and only an empty B pathname, not B inode or displaced lease-A before restoration. Cover Prune/Publish/Recover with exact A/B identity/bytes and selected state before restoration; equal-byte distinct B must fail. Core reader handoff/directory locking was statically found correct. |
| F-06 | Low; late-member recovery-cut matrix gap | `cmd/connectorgen/vnext_publication_test.go:2326`; `TDD-LEDGER.md:880`. The matrix omits fresh recovery of a rejected `journal.New` generation through `vnext_publication.go:1376`; immediate active-validation rollback is not that restart cut. Leave prepared/interrupted publication with old selection/finalized rejected new generation, restart real `Recover`, perform late lease substitution, and assert identity refusal, A/B and old-selection preservation, and recoverable state. |
| F-07 | Medium; R3-03 evidence gap | `TDD-LEDGER.md:860`; `PLAN.md:1038`. The required pre-repair production-main SIGINT/SIGTERM/repeated-signal RED remains planned; no actual command, tested identity, and observed failure are recoverable. `runMain` was statically found correctly scoped; broader/full writer GREEN does not constitute the required RED. Recover genuine provenance if available or retain this gap truthfully; do not manufacture a RED from current green behavior. |

### Preserved dispositions and lenses

- B-01 no-replace restoration/activation and typed collision causes remain; no overwrite fallback was found. B-02 remains strict phase-empty/predecessor-free/equal-prior-intended recovery with malformed/private/read-only refusals. R3-01 production descriptor admission is retained; F-04 is test-oracle only.
- R3-02 historical handle accumulation and prepared-hook state ownership are repaired; F-01/F-02 remain. R3-03 code is repaired; F-07 is proof provenance. The 31 introduced configured diagnostics remain dispositioned without configuration weakening; F-02/F-03 are remaining semantic concerns. R3-06 validated descriptor handoff and R3-07 directory lock plus independent late-member binding are retained; F-05/F-06 prevent acceptance.

| Mandatory lens | Status |
| --- | --- |
| Architecture/data flow; happy/bad/edge; state/concurrency; security/taint; retry/resume/idempotency; output/filesystem integrity; declaration/closed surface; CLI/App parity | Complete static inspection; F-01–F-07 are the remaining contracts. |
| Provider semantics | Not applicable: no provider operation, transport, credential, database, paging, or remote idempotency behavior changed. |
| Tests/evidence | Blocked for acceptance by F-04–F-07 and the remaining error-contract defects. |

Writer-recorded local hermetic execution remains attributed rather than independently rerun: Group-A REDs (Open `4.063s`, held reader `6.606s`, bounded descriptor `99.621s`); late-member nil-binding RED `2.485s`; focused GREEN `2.479s`; public-C `3.155s`; four-selector matrix `7.245s`; broader CP11 `121.026s`; normal `253.295s`; race `676.216s`; admitted definitions 553/0 and boundary clean exit 0 (284 files, 553 connectors, empty findings/warnings, six documented exceptions). `connectorgen-vnext-locks` is cached evidence, not a fresh run. Scoped lint retains the 15 established pre-range package items; four `internal/cli` failures and 20 broader lint items stay CP29/final-programme obligations. No provider-live, customer database, whole-suite, hosted, Linux/power-cut, shared-receiver, no-mistakes, or full-programme certification exists.

The older four-file staging instruction is historical, not a discrepancy in the ten-file `72943731` commit; contemporaneous six-path `3a455877` history and the unrelated CP10 fixture distinction remain corrected. Firstmate must author the next coherent repair brief for F-01–F-07. No piecemeal fix, CP12 advance, or acceptance is authorized by this failed exact-SHA review.

## 2026-09-06 — CP11 F-01–F-08 complete independent ledger audit

### Exact identity, custody, and scope

- **Verdict:** changes required; CP11 remains unaccepted. The completed independent GPT-6 Astra/xhigh audit reviewed the full immutable range `36e4d980de0d51d92fe74a68306845643596a6cb..7294373166db75466e2c92269f7887f51ceaddc6`, code tree `a748946dc72cf3b93f4eae51e1972fa98b1f4780`, base tree `bf52fee265e592fa9d18d49fcbffec0ecd8b335a`. It is a 37-path review; `72943731` has ten paths and `3a455877` has six, not four.
- Prior artifact-only `6c780efe3807411a790ba8c1f4eeb21310f15122` changes only this review record. The complete private audit is `/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-F01-F07-astra-xhigh-audit-7294373166db75466e2c92269f7887f51ceaddc6.md`, SHA-256 `8cce3d9ae85d3f24ad1bf2ab9691fda5715fe5bbb0a16ba41592815ef593497e`.
- The auditor was read-only: no project/Git mutation, test/build/generator/lint/GSD action, provider, credential, database, GitHub, or no-mistakes action. Writer-recorded tests remain writer-attributed; static inspection is not provider/runtime/full-programme certification.
- Original fresh review, R3/B ledgers, late-lease continuation, validation index, current PLAN/TDD/VERIFICATION, canon/Atlas and corrected F-07 provenance were read. This ledger amends dated records; it does not erase them or authorize repair itself.

### Complete deduplicated ledger

| ID | Disposition and classification | Required invariant / complete sibling scope |
| --- | --- | --- |
| F-01 | Confirm, Medium; existing R3-02/R3-06 identity-handoff miss | Capture use must bind the actual opened directory to recorded `capture.Identity` through validation, candidate lookup, rename, and sync. |
| F-02 | Confirm, Low; `fix_created:7294373166db75466e2c92269f7887f51ceaddc6` | Each `removeTree` child descriptor closes exactly once on normal, Fstat-error, and identity-refusal exits. |
| F-03 | Confirm/expand, Medium; R3-04 resource/error group | Writable Close, known-owned failed-creation cleanup, predecessor-anchor registration, and compound typed causes are truthfully accounted for. A primary error does not excuse meaningful durability/owned-cleanup loss. |
| F-04 | Confirm with causal qualification, Low; residual R3-05 proof gap | Snapshot identity and regular bytes come from the same opened no-follow/nonblocking descriptor; recursion uses verified directories. It is not a production R3-01 reopening. |
| F-05 | Confirm/extend, Low; proof sibling | A/B inode/type/bytes and selected-control state are asserted before restoration in reader, held-generation, public-caller, and rollback fixtures. The actual reader/lock production repair remains correct. |
| F-06 | Confirm/amend, Low; durable/caller proof gap | Prepared/new-selected, true committed/new-selected, successful Publish final prune, and old-selected/rejected-new fresh restart are separate cuts; immediate rollback/initial recovery are not substitutes. |
| F-07 | Amend; unavailable-RED claim withdrawn | Recovered actual production-main RED/edit/GREEN fulfills the historical obligation; records retain the exact uncommitted-tested-state limit. |
| F-08 | Add, Low; `fix_created:7294373166db75466e2c92269f7887f51ceaddc6` | Signal-process tests arm child cleanup/reap and lock release immediately, use bounded normal/cleanup waits, and prove real lock contention without sleep. |
| B-01/B-02 | Retained resolved contracts | No-replace A/B/C restore/activation and strict markerless phase-empty/predecessor-free/equal-prior-intended bootstrap recovery remain unchanged. |
| R3-01–R3-07 | Retained with the F siblings above | Keep descriptor regular/directory admission, bounded live handles while history stays retained, lock-render-only signals, no lint weakening, validated reader handoff, directory lifetime lock plus independent lease-member integrity. |

### Exact resource and caller map

- **F-01:** `vNextPublicationValidateCaptureLocked` (`vnext_publication_repair.go:981`), `vNextPublicationValidateCapturedCandidateLocked` (`:1045`), and `completeControlCaptureLocked` (`:1509/:1523/:1526`) validate A then reopen its pathname without comparing the opened descriptor. A lock-ignoring A→B swap can make `CURRENT`/`JOURNAL` move into B before later private revalidation fails. `state.openTransaction` and `vNextPublicationReadEmptyControlCaptureLocked` demonstrate the required checked-open pattern. Retain a verified descriptor or recheck every dependent reopen; no generic filesystem layer, path-only check, or retained graph descriptor.
- **F-02:** `removeTree`/`removeTreeBound` (`vnext_publication_dir.go:402/:419/:423/:425/:428/:431`) return before close after successful `openDirectory` on Fstat/mismatch. Reachable stage and generation callers are Prune, no-journal/selected-new/rejected-new recovery, successful publication cleanup, immediate rollback, and Open recovery. Top-level refusal retains no-replace restoration; nested traversal failure can leave root quarantined and prior removed siblings absent. No all-or-nothing recursion, finalizer/GC, global-limit, or history-deletion workaround.
- **F-03:** `writeJournalLocked`/`writeCurrentLocked`/`writeAtomicLocked` (`vnext_publication.go:791/:807/:819/:839–843`) discard writable temporary Close after selection may advance. `createControlRepairLocked` (`vnext_publication_repair.go:1259/:1269–1284/:1297–1308`) must retain bound-anchor/directory/transaction/connector cleanup outcomes and register a linked predecessor anchor before a later fallible Close. Combined close paths including `openFile` (`vnext_publication_dir.go:184`) retain meaningful original plus secondary typed causes. Read-only primary-error policy can remain documented; unknown paths are never deleted without identity.
- **F-04:** replacement snapshot `vnext_publication_repair_test.go:1651–1698`, especially `Lstat`/recursion/`os.ReadFile` at `:1674/:1680/:1684`, can observe A then read B, follow symlink, block on FIFO, or recurse through replacement. Production `openRegular`/`openFilesystemMember` are separately retained.
- **F-05:** `OpenKeepsValidatedDirectory` (`vnext_publication_test.go:37–77`) needs returned/retained A and B identity/bytes. `HeldGenerationUsesStableCleanupLock` (`:80–181`) needs displaced lease A and B identity/bytes before restore for Prune/Publish/Recover. Preserve stronger late-prune at `:2189`; strengthen public-caller `:2326` and rollback `:2688`. Publish compares to its intentional selected-state cut.
- **F-06:** `recoverLocked` rejected-new removal at `vnext_publication.go:1352–1388`, notably `:1376`, is distinct from immediate rollback. Existing `AfterCommitSync` is prepared-JOURNAL/new-selected because it runs before committed JOURNAL install. True committed/new-selected interrupts after `writeJournalLocked`; successful Publish final prune begins with old active then publishes new; old-selected rejected-new leaves old CURRENT, prepared JOURNAL Old/New and final new tree after stage rename/parent sync, restarts fresh, substitutes lease at `AfterGenerationLeaseIdentity`, refuses, then recovers after fixture restoration.
- **F-08:** non-consuming tests `vnext_publication_test.go:333–365` start children before cleanup ownership; lock-render `:396–411` sleeps then waits unboundedly while holding lock. Test seam is narrow/inert in production; child runs actual `main` and receives OS signal, never a context-only surrogate or new CLI/runtime switch.

### Observable RED/GREEN and factual limits

- First repair safe F-04/F-08 infrastructure: bounded real children prove regular→FIFO, regular→symlink and directory replacement at the snapshot classification/open boundary; observe coherent retained object or bounded refusal without FIFO/symlink read. Immediately arm kill/reap and defer lock release on every started child; bounded withheld-readiness proves reaping but timeout SIGKILL is never success. Retain FIFO handshake/exact terminating signal; lock-render requires exit 1/context cancellation, no success output, preserved selected/control/authority/generation state, release and retry.
- Then actual F-01/F-02/F-03 REDs precede production changes. Legitimate pending capture CURRENT/JOURNAL fixtures replace A with B at each real dependent open and observe unauthorized B validation/move. Repeated nested cleanup A→B substitutions count live descriptors without finalization and preserve A/B; separate open-descriptor Fstat error is a control. Real temporary-writer/known-owned creation injections Close after real Write/Sync, preserve injected/original typed errors, selected control/journal/authority/generation, CLI no-success, recovery and retry. No mock publisher, generic writer layer, invented nil-error short write, or arbitrary unknown deletion.
- F-05 equal-byte/different-inode B is an oracle control, not a production RED. F-06 decodes CURRENT/JOURNAL at every interruption; explicit Prune, no-journal Recover/Open, Publish initial recovery, prepared/new-selected, genuine committed/new-selected, final prune, old-selected/rejected restart, immediate rollback, stage ownership and non-destructive Check remain named distinct rows. Keep empty/nonempty B, held A and public C where their contracts apply.

### F-07 recovered historical evidence

The authoritative recovery is `/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-F07-signal-provenance-recovery-7294373166db75466e2c92269f7887f51ceaddc6.md`. The prior unavailable-command/failure/edit claim is superseded:

1. Worker record 3224/event 3223, physical line 3224, SHA-256 `65acae633258da13e38c4a9e0a64d532009bc2555b35ffc681a31b8d8828a14f`, at `2026-09-05T20:44:45.666Z` ran `gofmt -w cmd/connectorgen/vnext_publication_test.go && go test -count=1 -timeout 20m ./cmd/connectorgen -run '^TestConnectorgenMainPreservesNonConsumingSignalTermination$'`. It exited 1: parent 11.47s plus `interrupt` 2.45s, `terminate` 1.68s and `repeated-interrupt` 1.72s all failed to terminate validate; package FAIL was 12.383s.
2. Record 3230/event 3229, physical line 3230, SHA-256 `9e4d444101f045b996d3c19ebe5e4d61ecd4abc2b2ab256b7d583bd2d88a3aa9`, at `20:44:51.574Z` records actual `runMain`: non-lock-render delegates to `run`; only lock-render creates `NotifyContext`.
3. Record 3238/event 3237, physical line 3238, SHA-256 `b0501c5a52f0fa821bd27907990bae074199e1bfb2563cb4cab3ea9f8c11fcff`, at `20:45:14.884Z` ran combined non-consuming/lock-render selector and exited 0, `ok polymetrics.ai/cmd/connectorgen 15.417s`.

Firstmate's scoped observation and supervision record corroborate the sequence. The failing worktree was uncommitted: no full tree SHA exists, `72943731` is not its identity, and unrelated old-worktree files are not reconstructed. Canonical PLAN/TDD/VERIFICATION must record that dated correction while retaining the old claim as superseded history. F-08 does not reopen correct signal scope or create forced-second-signal/mid-transaction cancellation guarantees.

### Ten lenses, preserved boundaries, and next order

| Lens | Audit disposition |
| --- | --- |
| Architecture/data flow | Complete: main/runMain, lock-render, authority/capture, identity/lease guards, cleanup, output and recovery traced; F-01–F-03 remain. |
| Happy/bad/edge | Complete: malformed/partial controls, substitutions, FIFO/symlink, cancellation, bootstrap/recovery/no-clobber traced; F-04–F-06/F-08 are proof gaps. |
| State/concurrency | Complete: connector/generation locks, independent lease, capture graph and prepared/selected/committed/rejected transitions traced; no mid-transaction expansion. |
| Security/taint | Complete: strict no-follow paths and private/public identity remain; no credential/provider sink; F-01 is local identity boundary. |
| Retry/resume/idempotency | Complete for publication; F-03/F-06 require truthful error/cut evidence. |
| Output/filesystem integrity | Complete; F-03 Close and F-04/F-05 observations remain. |
| Declaration/closed surface | Complete: seven lanes/staged loader/preflight/Atlas remain authoring-only; no fallback/foundation/importer. |
| CLI/App parity | Complete for scope: connectorgen only; PM/App/approval/runtime unchanged. |
| Provider semantics | Not applicable: provider transport, credentials, database, paging and remote idempotency unchanged. |
| Tests/evidence | Complete static audit but acceptance blocked by F-01–F-08; fixtures were traced to actual cut/objects. |

Preserve B-01 typed no-replace A/B/C/public-C behavior, B-02 strict bootstrap, R3-01 descriptor admission/nested directories, R3-02 bounded live descriptors with history retained, R3-03 lock-render-only registration, R3-04 31 introduced diagnostics without lint weakening plus 20 global/15 package pre-range debt, R3-05 proof scope, R3-06 validated reader, R3-07 directory lifetime plus independent late-member guard. Four historical `internal/cli` failures remain CP29/final-programme obligations. Cached `connectorgen-vnext-locks` remains cached; no provider-live, customer database, whole suite, hosted verification/security, Linux/power-cut, no-mistakes, release, or merge certification is claimed.

Correct dated `VERIFICATION.md` “51 pre-existing”: 31 introduced diagnostics belong to complete CP11 range, `3a455877` has six paths, and old four-file/CP10-fixture text is not commit evidence. Preserve pre-range ownership records without inventing twelve grouped orderedjson names.

**Required order:** commit this ledger; repair F-04/F-08 infrastructure; run real F-01/F-02/F-03 REDs; repair their one coherent resource/error group without weakening authority/locks/type/no-replace; complete F-05/F-06 proof and caller cuts; correct F-07/canonical mechanical history; run focused behavioral and retained B/R3/package/race/canon/Atlas/preflight/static gates; freeze exact code SHA; obtain fresh independent whole-range Astra/xhigh review. No per-finding acceptance, scope reduction, or finding/round cap applies. No unresolved product or architecture decision remains.

## 2026-09-06 — CP11 F-07 canonical-record enactment

The requested PLAN/TDD/VERIFICATION correction is now recorded. The old
unavailable-provenance assertion is preserved above as superseded historical
review evidence; it is withdrawn, not silently rewritten. The authoritative
recovery remains
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-F07-signal-provenance-recovery-7294373166db75466e2c92269f7887f51ceaddc6.md`.

It identifies the actual pre-routing RED as worker record 3224/event 3223,
physical line 3224, SHA-256
`65acae633258da13e38c4a9e0a64d532009bc2555b35ffc681a31b8d8828a14f`:
the non-consuming signal selector exited 1 with `FAIL` in 12.383s. Record
3230/event 3229/line 3230, SHA-256
`9e4d444101f045b996d3c19ebe5e4d61ecd4abc2b2ab256b7d583bd2d88a3aa9`,
is the source-edit boundary that scopes `NotifyContext` to lock-render.
Record 3238/event 3237/line 3238, SHA-256
`b0501c5a52f0fa821bd27907990bae074199e1bfb2563cb4cab3ea9f8c11fcff`,
is the combined selector's exit 0 / `ok` in 15.417s. The transcript gives
the full parent/interrupt/terminate/repeated-interrupt failure sequence and
the exact commands in the linked recovery record.

No full tree SHA was recovered because the RED was an uncommitted tested
worktree state. `7294373166db75466e2c92269f7887f51ceaddc6` remains later
reviewed code, never the red fixture identity. F-08 remains a separate
harness correction; this record grants no extra signal or cancellation
contract. CP11 remains unaccepted pending the complete Group 3 proof matrix,
whole-package/race/static gates, freeze, and fresh independent exact-SHA
review.

## 2026-09-06 — CP11 code freeze and final review input

- The complete coordinated F-01–F-08 behavior wave is frozen at code commit
  `e77cd390f45eb938917dc3c39882bda34aae09a6`, tree
  `4c37521b11a6f1a22bd6fa5ea5e043d4982329f8`. This follows the preserved
  candidate `0ff2ac7d96c5b67a76da6228631b9e0d057a9e4e`; existing published
  head `8214bd91403ce620773b61caf674faa540ee1701` is an ancestor. The review
  subject remains the complete original range and all repair deltas, not an
  isolated final commit.
- Current execution includes full `cmd/connectorgen` normal/race PASS at
  263.677s/691.666s after final new-only lint zero. The reviewer receives
  GROUP1/2/3 evidence, TDD order correction, exact baseline fixture/overlay
  provenance, and this exact code identity. The final artifact binding is
  documentation only; no behavior source/test change is permitted while the
  review and mechanical index run.
- The outcome must retain every full-lens finding/disposition or a reasoned
  zero-blocker result. A report alone does not accept CP11, publish, or begin
  CP12.

## 2026-09-06 — CP11 e77 final-review discovery aggregation (frozen intake)

This intake records the complete independent Astra discovery ledger before the
required separate ledger audit. It does **not** accept CP11, mark a finding
fixed, authorize a test/source change, or alter the frozen behavioral identity:
`e77cd390f45eb938917dc3c39882bda34aae09a6`, tree
`4c37521b11a6f1a22bd6fa5ea5e043d4982329f8`. The original review range is
`36e4d980de0d51d92fe74a68306845643596a6cb..e77cd390f45eb938917dc3c39882bda34aae09a6`
(46 paths); e77 itself changes 15 paths. Published
`8214bd91403ce620773b61caf674faa540ee1701` remains an ancestor. The current
artifact-only review-input binding is
`e27caeac904dd1f5521c67f045655bc564e88ff4`, tree
`f89d41cb7c9c41c9b95515c9428176f837d9e9d8`; this intake is documentation
only and is not a behavioral retest.

The immutable discovery source is the complete read-only Astra report
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-final-astra-e77cd390-full-range-review.md`,
SHA-256 `8b862d9ebe1e1644c7fc27e5c028b799c1ac4762a27e269489b098656302b3c2`.
Its private overlay/fixture bytes remain preserved and must not be rewritten.
The report's verdict is **CHANGES REQUIRED**: one Medium and five Low entries.
The worker generated the installed `scripts/gsd prompt code-review 4427`
instructions and applies its documented inline Codex review fallback; the
required new independent Astra/xhigh ledger auditor has not yet returned.

### Complete merged discovery ledger

| ID | Proposed severity / causal classification | Reachable invariant and observed or static evidence | Bounded acceptance to carry into the coordinated repair plan |
| --- | --- | --- | --- |
| F-03-A | Medium; existing-range missed durability/ownership sibling, rooted in `8214bd91403ce620773b61caf674faa540ee1701` | `createControlRepairLocked` can retain `prepared.json` after a real prepared writer completion or post-record transaction/connector Sync failure while its `!keep` cleanup removes known prior/intended anchors. The private real-Sync/injected-completion probe returned the injected error plus `ENOTEMPTY`, left a prepared-only transaction, and fresh Recover refused the missing intended anchor while old CURRENT stayed intact. | Cover real record Write/Sync/Close and both post-record directory Sync frontiers across bootstrap/successor, present/absent prior and intended-present/logical-absence. Preserve a coherent recoverable graph or remove the complete identity-proven unexposed preparation; prove fresh Recover/Check/retry and no foreign A/B/C removal. |
| F-03-B | Low; unresolved existing-range F-03 error-contract root | Meaningful original open plus parent Close, parent Close plus file Close, writable prepared/phase record Close, and predecessor Close retain only one typed cause or flatten the other at `vnext_publication_dir.go:185,191`, `vnext_publication.go:209,1214`, and `vnext_publication_repair.go:466–469,1320–1325`. `ReadControlBound` absence handling must not turn joined absence plus teardown failure into clean absence. | Fault actual compound paths and assert both inspectable causes, one Close owner, truthful durable/public state and retry. Retain only explicitly justified harmless read-only teardown suppression; do not use a generic error or filesystem framework. |
| F-04-R | Low; `fix_created:e77cd390f45eb938917dc3c39882bda34aae09a6` proof-helper sibling, not a production R3-01 regression | `FileWitness` observes `Lstat(A)` then pathname `ReadFile`; `DirectoryWitness` reads by pathname after observing root A. A replacement can yield A metadata with B bytes, follow a symlink, or block on FIFO. The repaired production `TreeSnapshot` remains accepted. | Make the witness contract descriptor-bound at its actual open/observation boundary; demonstrate regular→FIFO, regular→symlink and directory A→B bounded refusal or coherent retained A, plus normal nested regular bytes. Treat controls as oracle proof, never a fabricated production RED. |
| F-05/F-06-P | Low; incomplete existing F-05/F-06 proof, with false e77 explanatory claim | Durable cuts execute, but common control assertions decode values rather than retain raw CURRENT/JOURNAL bytes and inode identity; selected roots/content and real private prepared/phase/anchor graph are not retained at each cut. `GROUP3-EVIDENCE.md:50–53` incorrectly says ordinary Publish creates no authority; source and the reviewer probe show retained authority after normal Publish. Rejected-new/immediate rollback have empty/nonempty B; true committed/final prune/public-caller rows have nonempty B only. | For each actual destructive caller/cut, account for legitimate advanced controls while asserting allowed bytes/inodes, generation root/active content, relevant real private authority and fixture-owned restoration. Retain A/B/C, returned A, held lease and fresh-object limits; correct the no-authority assertion and explicitly state executed variants. |
| F-02-P | Low; uncompleted existing audited F-02 proof obligation, not an observed remaining leak | The GC-disabled 24-iteration direct child test and opened-child Fstat-error control prove the immediate descriptor fix, but neither runs a nested owned stage/generation through real quarantine/public recovery or prune with nested replacement/failure. | Exercise public Recover/Prune nested quarantine traversal after ownership, inject nested post-identity/pre-open replacement and opened-child identity failure, account for descriptor/lock ownership and A/B/quarantine residue. Preserve public C/fresh recovery and explicitly allow already-owned earlier siblings to have been removed; do not promise all-or-nothing recursion. |
| F-08-R | Low; incomplete F-08 repair, with lsof readiness seam `fix_created:e77cd390f45eb938917dc3c39882bda34aae09a6` and older broad Signaled assertion | Current lsof confirms connector-directory open before source admission/canonicalization and `acquireOperation`, so SIGINT can arrive before flock contention; non-consuming cases accept any signaled exit instead of the signal sent. Immediate cleanup/direct Wait/withheld-readiness reaping remains resolved. | Acknowledge the exact nonblocking flock EWOULDBLOCK on the real opened directory before signalling actual main; add a pre-flock-delay negative control; require expected SIGINT/SIGTERM, exit 1/no success, exact selected/control/private-authority preservation, bounded waits and retry. No new CLI/runtime/cancellation contract. |

### Ten-lens discovery disposition

| Mandatory lens | Frozen discovery disposition |
| --- | --- |
| Architecture/data flow | Complete mapping of CLI/admission/retained roots and locks/stage/authority/current/journal/reader/prune; F-03-A is the failed-preparation authority discontinuity. No new runtime reader/foundation. |
| Happy/bad/edge | Complete for malformed controls and filesystem substitutions; F-03-B and F-04-R retain compound-error and hostile-observation proof gaps. Redirect/provider behavior is unchanged and not in scope. |
| State/concurrency | Complete durable-cut, lock, capture/predecessor and no-replace mapping; F-03-A, F-02-P, F-05/F-06-P and F-08-R retain the stated observable gaps without adding same-UID mutation or all-or-nothing recursion requirements. |
| Security/secret taint | Complete: bounded descriptor-relative inputs retain no-follow defenses and introduce no credential/provider route. F-04-R is test-helper observation scope only. |
| Retry/rate-limit/resume/idempotency | Complete for local publication/recovery; F-03-A blocks safe retry. Provider rate/replay/checkpoint behavior is unchanged and N/A. |
| Output integrity | Complete selected-generation/content/stdout-after-success tracing; F-03-B preserves typed error truth and F-04-R/F-05/F-06-P limit observation proof. Raw transport envelopes are unchanged/N/A. |
| Declaration reachability/closed surface | Complete: no declaration/source-lock/seven-lane/runtime mapping change; staged physical preflight/Atlas evidence remains authoring-only and closed. |
| CLI/App parity | Complete for connectorgen-only scope; PM hand parser, App persistence, plan/preview/approval and runtime selectors are unchanged. F-08-R is test proof, not a new CLI surface. |
| Provider semantics | Not applicable: no provider transport, credential, database, paging, GraphQL or remote idempotency behavior changed. |
| Tests/evidence | Complete discovery with changes required. The report distinguishes fresh commands, cached/history, private defect-reproduction probes, uncommitted historical F-07 RED, pre-range boundary/CLI/lint debt, and unproven Linux/power/provider/no-mistakes/release claims. |

### Preserved and adjudication boundaries

- B-01 no-clobber A/B/C/public-C restoration, B-02 strict markerless bootstrap,
  R3-01 descriptor admission, R3-02 bounded live descriptors/history retained,
  R3-03 lock-render-only signal registration, R3-04 lint attribution,
  R3-05 proof scope, R3-06 returned validated reader, and R3-07 directory
  lifetime lock plus independent lease-member guard remain protected.
- F-01 capture identity, F-02 immediate child descriptor ownership, the
  pre-prepared F-03 link-registration fix, the main F-04 oracle, meaningful
  F-05/F-06 durable distinctions, F-07 historical provenance correction, and
  F-08 immediate child cleanup are accepted partial dispositions. They are not
  rewound by the six listed siblings.
- The report's `internal/connectors/boundary` failure is attributed to the
  exact-base package-source overlay under current dependencies, not CP11; the
  archival premature-F01 patch is the sole exact-range diff-check whitespace
  artifact. Neither fact certifies a full baseline suite or green boundary
  package.

### Next audit gate

No source/test behavior may change while the new independent ledger audit
checks this complete discovery set. It may amend, withdraw, add or group items;
only its complete frozen verdict will be incorporated and committed as the
pre-repair canonical ledger. There is no unresolved product or architecture
decision in this intake, no push, and no CP12/no-mistakes authorization.

## 2026-09-06 — CP11 e77 independent complete-ledger audit (frozen pre-repair record)

The independent Codex GPT-6 Astra/xhigh auditor completed the required
whole-ledger adjudication without changing project files or refs. The
authoritative complete original return, recovered from the exact worker rollout
message `amsg_01a07495-52ae-7a00-a473-530109963ffb` at physical line 9330,
is preserved at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-e77-audit-original-return.md`,
SHA-256 `9d68c63c7d6b525dab6010997a9eb36dc43890ca9cfd657523bbee25e47e24ba`;
adjacent `cp11-e77-audit-original-return-provenance.json` binds its message,
session-meta CLI source and exact cwd provenance. The later
`cp11-e77-complete-ledger-audit.md` copy differs only by a terminal newline
(SHA-256 `6d0d5d80bd6355d2f578649f47dd36af5d9b441007a432b8a805dcac358010f2`)
and is non-authoritative for this binding.
It audits the immutable behavioral code
`e77cd390f45eb938917dc3c39882bda34aae09a6`, tree
`4c37521b11a6f1a22bd6fa5ea5e043d4982329f8`, over original base
`36e4d980de0d51d92fe74a68306845643596a6cb`; the full range is 46 paths and
e77 is 15 paths. Published `8214bd91403ce620773b61caf674faa540ee1701`
remains an ancestor. The prior final-review report remains immutable at
SHA-256 `8b862d9ebe1e1644c7fc27e5c028b799c1ac4762a27e269489b098656302b3c2`.

**Frozen audit verdict: CHANGES REQUIRED.** All six submitted entries survive
with the amendments below, and the audit adds F-03-C. This is the complete
pre-repair ledger: one Medium and six Low entries. It supersedes the preceding
"next audit gate" pending wording, but neither accepts CP11 nor authorizes a
repair before this documentation record is committed and Firstmate supplies
the coordinated repair prompt.

### Audited complete disposition ledger

| ID | Verdict / severity / classification | Audited invariant, trigger and evidence | Required bounded repair/proof scope |
| --- | --- | --- | --- |
| F-03-A | Confirmed and amended; Medium; existing-range failed-preparation ownership root present in `8214bd91403ce620773b61caf674faa540ee1701` | After a valid prepared record can name anchors, record Close or post-record transaction/connector Sync failure can invoke `!keep` anchor cleanup while leaving `prepared.json`, stranding a self-created authority graph. The preserved real-Sync/injected-completion probe reproduced error plus `ENOTEMPTY`, prepared-only residue, failed fresh Recover and unchanged old CURRENT; its pass is a defect witness, not GREEN. | Cover pre-record, Write/partial, Sync, real Close, transaction Sync and connector Sync frontiers across valid bootstrap and successor logical classes. Retain coherent authority or remove only a complete identity-proven unexposed preparation; distinguish absence-only and malformed records, preserve meaningful causes/one Close owner, and prove pending Check/fresh Recover/retry without foreign A/B/C removal. |
| F-03-B | Confirmed, amended and sibling-expanded; Low; existing-range compound-error contract gaps | `%v`/single-cause branches at directory open/close, roots, opened-control identity, writable marker/prepared/phase records, predecessor link, writable staged files, owned-stage cleanup and capture Sync/Close lose typed meaningful causes. A producer repair must not make `ReadControlBound` treat absence plus completion failure as clean absence; conflict consumers need compatibility review. | Fault real compound paths with real opened resources, retain both inspectable causes through relevant public callers and accurate durable state, cover shared writable records/staged and owned-stage siblings, pure absence versus absence-plus-failure, bounded release/retry. Harmless read-only teardown may remain explicitly policy-bounded; no generic framework. |
| F-04-R | Confirmed and amended; Low; `fix_created:e77cd390f45eb938917dc3c39882bda34aae09a6` test-observation sibling only | New `FileWitness`/`DirectoryWitness` can pair A metadata with B pathname bytes, follow a symlink or block on FIFO. This does not regress accepted production `TreeSnapshot`. | Use the existing descriptor-bound, no-follow/nonblocking observation contract at the witness boundary; add bounded FIFO/symlink/directory replacement negative controls plus nested regular positive evidence. Refusal or coherent retained A must precede unrelated reads; label as oracle proof, not production RED. |
| F-05/F-06-P | Confirmed and amended; Low; incomplete accepted proof obligation and false Group 3 authority assertion | Major durable cuts are real, but control assertions lack raw bytes/inodes and selected/rejected root/content/private authority identity before restoration. Ordinary Publish does retain private authority (six transaction directories in the independent baseline); Group 3's contrary sentence is false. Returned A, held A/B lease, committed cut, final prune and rejected-new distinctions remain accepted. | At each named caller/cut observe expected logical advancement plus exact control bytes/type/inode, roots/content and relevant private transaction/prepared/phase/anchor identities before fixture restoration. Complete meaningful empty/nonempty B rows for unheld destructive generation cases, retain A/B/C and no-clobber, distinguish fresh publisher object from process restart, and correct derivative documentation. |
| F-02-P | Confirmed and amended; Low; missing nested public-quarantine proof, not a remaining FD leak | Immediate child ownership before identity failure is accepted from GC-disabled direct tests and opened-Fstat control. No equivalent witness covers nested owned stage/generation public Recover/Prune quarantine traversal. | Use real public cleanup after root ownership/quarantine with nested post-identity/pre-open replacement and opened-child identity failure; account for descriptors, locks, A/B/quarantine residue and bounded fresh recovery. Earlier owned sibling deletion is permitted; keep public-C/no-replace and do not invent a stage lease or all-or-nothing recursion. |
| F-08-R | Confirmed and amended; Low; e77 lsof readiness proof gap plus older broad signaled assertion | Opening the connector directory occurs before source admission and flock, so lsof readiness does not prove EWOULDBLOCK contention; any-signaled exit does not prove sent SIGINT/SIGTERM. Immediate cleanup/direct Wait/withheld-readiness reaping remains accepted. | Emit a test-owned inert acknowledgement only after the exact retained-directory flock returns EWOULDBLOCK/EAGAIN under the parent lock; pre-flock open is a negative control. Then signal actual main and prove exact signal, exit 1/no success, selected/control/private-authority preservation, bounded waits and retry. No new CLI/runtime/second-signal/mid-transaction contract. |
| F-03-C | Added and confirmed; Low; existing-range cleanup-ownership root in `958a07a778fba6264d1aec567efa5d8c853eefa2` | Failed temporary-control and quarantine allocation cleanup re-observes a pathname after owned A was moved and uses B's current identity as authority to delete B. Private overlay reproductions cover Publish real EEXIST and Prune real-open/injected-Fstat-completion paths; both delete distinct empty B while preserving moved A. They are defect witnesses, not GREEN. | Retain/establish authority from the originally owned opened object before cleanup; fail closed when it cannot be proven. Exercise temporary control and quarantine public paths with empty/nonempty B, preserve displaced A and exact B identity/bytes, typed primary/completion/cleanup errors and bounded retry/recovery. Cover shared CURRENT/JOURNAL temporary and stage/generation quarantine reachability without a new allocator/foundation. |

### Audited adjudications and preserved boundaries

- Withdrawn: ordinary Publish creates no private authority; all prepared-only
  failures are inherently corrupt; e77 created F-03-A; F-02-P demonstrates a
  remaining leak; F-04-R reopens production R3-01; every post-Publish field
  must equal its pre-Publish state; nested cleanup is all-or-nothing; and a
  fresh publisher object proves another process.
- Retained: B-01 A/B/C/public-C no-clobber, B-02 strict markerless bootstrap,
  R3-01 descriptor admission, R3-02 bounded live descriptors/history,
  R3-03 lock-render signal scope, R3-04 truthful lint attribution, R3-05
  limited proof scope, R3-06 validated returned reader and R3-07 directory
  lifetime lock plus lease-member integrity. F-07's recovered historical
  12.383s uncommitted RED, routing edit and 15.417s GREEN remain accepted;
  `72943731` is not the failing tree.
- The boundary package failure remains strongly pre-range under exact-base
  package-source overlay, never a green boundary-package claim. Exact-range
  diff-check whitespace is confined to the preserved premature-F01 patch;
  excluding only that archival artifact yields the code whitespace result.
  Full/race, Linux, provider, release and delivery gates were not re-labeled
  as auditor execution.

### Complete ten-lens audit disposition

| Lens | Status |
| --- | --- |
| Architecture/data flow | Complete; F-03-A/C are failed-cleanup ownership discontinuities on the traced CLI→authority→publication→cleanup graph. |
| Happy/bad/edge | Complete, changes required; malformed controls, substitutions, allocation failure and cancellation were assessed; F-03-B/F-04-R retain error/oracle gaps. |
| State/concurrency | Complete, changes required; durable cuts and no-clobber/lock integrity are mapped; nested ownership and real contention remain open. |
| Security/secret taint | Complete; local filesystem authority only, no credential/provider sink; F-03-C is empty-directory ownership deletion, not content exfiltration. |
| Retry/rate-limit/resume/idempotency | Complete for changed local behavior; F-03-A blocks normal recovery and F-03-C can permit retry after B deletion; provider rows N/A. |
| Output integrity | Complete, changes required; physical content/control/returned A and success output traced; typed compound causes and coherent witnesses remain incomplete. |
| Declaration reachability/closed surface | Complete; no source-lock, lane, runtime mapping or provider declaration change. |
| CLI/App parity | Complete for changed connectorgen boundary; PM/App persistence, approval and runtime paths unchanged/N/A. |
| Provider semantics | Not applicable; no transport, credential, response, paging or rate behavior changed. |
| Tests/evidence | Complete audit, acceptance incomplete; observed F-03-A/C probes, proof gaps, historical provenance and platform/full-suite limits are distinguished. |

### Dependency-ordered next gate

1. Commit this exact audited documentation-only record before any source/test
   repair.
2. Firstmate must author the coordinated repair prompt for all seven entries;
   no per-finding fix loop or scope reduction is permitted.
3. That future wave begins with trustworthy witness/contention controls, then
   genuine RED evidence for the ownership/error group, coherent repair,
   nested/durable proof completion, accurate records/Atlas references,
   proportional gates, a new frozen code SHA and fresh full-range review.

No product or architecture decision remains unresolved. This audit grants no
source acceptance, push, no-mistakes, CP12, provider/live, release or merge
authority.

## 2026-09-06 — CP11 Group 2 F-03 proof-completion status (steer 058)

The frozen F-03-A/B/C ledger remains the repair boundary. The current focused
post-repair controls and intermediate full-package receipt demonstrate only
their named cases; they do not close the audit or permit a final review. The
receipt is the original returned event at physical `10897` (raw SHA-256
`4fd7390eef0b432f8f5f983d3924f9360bde2993d51586fa48bc6658634e0255`) and is
explicitly limited to its then-uncommitted source state.

The remaining review-required proof is tracked case-by-case in
`GROUP2-F03-REPAIR-EVIDENCE.md`: F-03-A's full preparation-frontier and valid
state-class graph, F-03-B's resource-backed compound producer/consumer paths
and pure-versus-compound absence, and F-03-C's CURRENT/JOURNAL temporary plus
stage/generation quarantine reachability. Each added test must retain exact
case/seam identity, actual resource/Close ownership, observable state before
fixture restoration, meaningful error causes, pending Check and permitted
fresh recovery/retry. Post-repair coverage is not a retroactive RED. No
independent exact-SHA verdict is requested or implied while these cells remain
incomplete.

### 2026-09-06 — CP11 Group 2 F-03 proof-cell disposition (steer 058/060)

The former incomplete proof cells are now dispositioned as **covered**, not as
an independent exact-SHA review verdict. The active proof source is
`cmd/connectorgen/vnext_publication_group2_original_test.go`; the only added
seams are inert direct-descriptor test boundaries:
`vNextPublicationControlRecordHooks` and record/directory-sync fault points in
`vnext_publication.go`/`vnext_publication_repair.go`, the raw
opened-file-after-parent-Close hook in `vnext_publication_dir.go`, and
read-control completion/Close hooks in `vnext_publication.go`.

| Former proof gap | Covered executable proof |
| --- | --- |
| F-03-A frontiers and graph classes | `TestCP11F03ARepairPreparationFrontierMatrix` covers each record frontier in JOURNAL absent→present, CURRENT present→present, and JOURNAL present→logical-absence; `TestVNextGenerationPublisherResumesInterruptedBaseAuthorityPreparation` covers absent/absent base; `TestCP11F03ARepairBasePresentPresentAuthorityRecovers` covers both CURRENT/JOURNAL present/present bases. Each observes ownership graph/identity, nonmutating Check, fresh recovery and retry. Real Sync/Close precedes injected completion; no natural I/O/power-loss inference is made. |
| F-03-B concrete consumers | `TestCP11F03BRepairCompoundCausesRemainInspectable` covers the real open/parent-close, parent/opened-file-close, read/close/pure-absence, common record writer, predecessor, stage, and capture paths with joined observable causes and actual close ownership. |
| F-03-C actual temporary/quarantine callers | The four `TestCP11F03CRepair*` selectors cover Publish, CURRENT/JOURNAL transition, stale-generation Prune, and stale-stage A→B cases over empty/nonempty B; they assert A/B identity/type/bytes/residue before fixture-only cleanup plus the defined fresh recovery/retry. |

Current normal/race evidence is F-03-A `20.804s` normal and bounded classes
`10.200s`/`10.089s`/`11.507s` race; base `2.617s`/`4.095s`; B/C
`8.442s`/`11.749s`, all `ok`. The four physical `11540/11549/11563/11577`
receipts remain exact earlier two-class evidence and are not attributed to the
later state. This resolves only the identified F-03 proof gaps. A fresh
whole-range review and the remaining Group 2/static/final CP11 gates are still
required.

### 2026-09-06 — independent e77 ledger intake and current disposition

The immutable independent Astra/xhigh review at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-final-astra-e77cd390-full-range-review.md`
(SHA-256 `8b862d9ebe1e1644c7fc27e5c028b799c1ac4762a27e269489b098656302b3c2`)
and its complete ledger audit at
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-e77-complete-ledger-audit.md`
(SHA-256 `6d0d5d80bd6355d2f578649f47dd36af5d9b441007a432b8a805dcac358010f2`)
returned **CHANGES REQUIRED** for the frozen range through `e77cd390`.
Their scope is preserved; neither report is retroactively made an acceptance
of this later repair tree.

| Audit ID | Current disposition |
| --- | --- |
| F-03-A, F-03-B, F-03-C | The coordinated Group 2 repair/proof slice addresses the named ownership/error roots: complete prepared authority is retained before later sync frontiers; meaningful compound causes have direct consumers; and temporary/quarantine cleanup stays identity-bound to A. The resource-backed current results above are the response evidence. A new independent exact-SHA review must still accept or reject them. |
| F-04-R | Provisionally remediated after the e77 review by `69246943`: the durable-matrix file/directory witness holds the checked descriptor, and the bounded FIFO/symlink/directory post-open child control is recorded in `GROUP1-EVIDENCE.md`. The audit predates that commit; fresh exact-SHA review still decides acceptance. |
| F-05/F-06-P | Provisionally remediated after e77 by the current Group 3 durable witnesses: descriptor-bound raw CURRENT/JOURNAL bytes/type/inode, selected/rejected/stale trees, and real marker/transaction/prepared/phase/anchor graph are asserted at the public caller cuts; ordinary Publish authority text is corrected. Empty/nonempty `.lease` B execute the unheld schedules. Fresh exact-SHA review still decides acceptance. |
| F-02-P | Provisionally remediated after e77 by `TestVNextPublicationPublicNestedQuarantineBoundsChildOwnership`: public Recover-owned-stage, Prune, and final-Publish cleanup cross quarantine into nested replacement/opened-child failure, account bounded descriptors and candidate residue, then fixture-only restore/retry. No remaining leak or all-or-nothing rollback is claimed; fresh exact-SHA review still decides acceptance. |
| F-08-R | Provisionally remediated after the e77 review by `69246943`: `LockContention` fires only after actual `LOCK_NB` contention on the retained descriptor, the child proves directory-open is insufficient before release, and non-consuming status compares the exact sent signal. `GROUP1-EVIDENCE.md` records bounded normal/race results; fresh exact-SHA review still decides acceptance. |

These are finding dispositions, not a finding cap, scope reduction, or an
authorization for CP12, no-mistakes, publication, provider work, release, or
merge. The historical e77 reports are the frozen diagnosis, not a renewed
review event for later Group 1/2 commits. Group 3 alone is the remaining
authorized execution group; the resulting coherent tree still requires fresh
independent whole-range exact-SHA review.


## 2026-09-06 — Firstmate 070/073 preserving owner and complete 7481 audit intake

### Task Delivery Header

- Issue: Refs #4427 — transactional connector-generation publication; parent Refs #4325.
- Base branch: `main`, verified by the GitHub pull API on 2026-09-06.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`; main merge remains captain-owned.
- Delivery: preserve the complete independent ledger audit and commit its evidence locally; no push, repair, CP12 advance or programme acceptance in this intake.
- Working branch: `fm/cli-top100-declaration-batch-r1`.
- Existing PR: https://github.com/polymetrics-ai/cli/pull/4294.
- Task: aggregate the entire fresh review and all prior-seven dispositions, obtain Firstmate's supplied independent Astra/xhigh complete-ledger audit, and preserve every disposition for Firstmate's coordinated response.
- Verification: exact Git/remote identity; original report/handoff SHA-256; code/test/dependency freeze; complete original report and ten-lens/contract-to-evidence audit with original command custody.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Preserving owner and delivery identity | live | `pwd -P` and Git top-level both resolve to the assigned isolated worktree; branch/HEAD/tree and actual pull API head/base match 067/070. |
| Runtime and branch custody | live | Current ancestor Codex PID 81393 arguments select `gpt-6-astra`, `model_reasoning_effort=medium`; process ancestry includes Herdr. Doctor is healthy and AXI status reports zero current-branch/repository runs. |
| Immutable original review and full aggregation | live | File hashes match 067 and 070; all five entries and seven dispositions below are transcribed verbatim from the full original review, which remains unchanged. |
| Complete independent adjudication | live, pending | A new native Astra/xhigh context must return the complete original-range audit; this draft supplies no acceptance. |
| No provider or customer-data exercise | fake | Existing hermetic fault-injection receipts/probes are the authorized evidence scope; no provider credentials or customer database access is authorized. |

### Verified source and review custody

Worktree `/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli` remains at
`afde575a1368fe86b5ab5acb5a3d39a294a20e96`, tree
`41bead7def7472a0383b1081b1134c9e19803624`, 28 local commits ahead of the
published `8214bd91403ce620773b61caf674faa540ee1701`. API base is `main` at
`813f457a925f7ee3fe3bea101a43e445992c8552`. Only the protected untracked
`.cache/` existed before this evidence edit; it remains untouched.
Behavior/test identity is `7481d1770a21cc95869fd10bf281f632af48c089`, tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`, over original base
`36e4d980de0d51d92fe74a68306845643596a6cb`. The descendant changes only
GROUP3-EVIDENCE, PLAN, TDD-LEDGER and VERIFICATION. This intake changes evidence
only. The 067 handoff reports all old native children and command sessions
terminal; none is revived or adopted.

Complete original fresh report:
`/Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-astra-7481d177-fresh-full-range-review.md`,
380 lines / 52,454 bytes, SHA-256
`2d92ce239d19509aa1838c23d5d0b9f31f4e5784232a512ee2b0f93e70ca571c`.
It returned from fresh Astra/xhigh context
`cp11_astra_7481d177_fresh_review_r1`, thread
`01a07552-ef8d-7c01-b366-f4f79190c1c1`. Verdict: **CHANGES REQUIRED**,
exactly five unadjudicated entries (one Medium and four Low). No self-acceptance
or 90% estimate survives this intake. Handoff 067 hash is
`ee71e0fbe4afbb0f996ce38923e706d77ec8abcdf207005165251cbacf432afe`.

### Workflow, skills and independent role routing

Read required routing, issue/parent contracts and GSD adapter. Loaded
`no-mistakes`, `golang-how-to`, Go security/safety/error-handling/lint/testing,
`connector-lane-build-order`, and the pinned `firstmate-exhaustive-review` gate.
The approved `REVIEW-EVIDENCE-REQUIREMENTS.md` is required in full for the audit.
`scripts/gsd sources code-review` and `scripts/gsd prompt code-review
batch-r1-vnext-cutover` resolve the installed workflow. Doctor retains the
known unrelated missing `.gsd/prompts/issue-122-rebootstrap.md`. The documented
non-Pi inline/native route is used beneath this canonical worker. Firstmate073
explicitly authorizes one independent native audit context and supersedes the
generic local no-role rule for this assignment; it supplies the entire prompt.
No new prompt scope, fixer, Pi role, pipeline or adapter is introduced.
The requested fresh role is `cp11_7481_complete_ledger_audit_astra_r1`, model
`gpt-6-astra`, effort `xhigh`, no inherited conversation. Actual launch receipt
and complete verdict will be recorded after launch/return. Captain's current
Astra and no-intermediate-push policy supersedes the imported gate's older
Claude/push wording. Historical Terra/Luna records retain their real identity.

### Complete original fresh findings and prior-seven dispositions — verbatim intake

The following is verbatim review material, not the owner's adjudication. Its
source-line references bind 7481. The auditor must also read the whole original
report, including prior B/R3/F resolutions, ten lenses and proof custody.

## Complete actionable ledger

### 7481-01 — Medium — prepared creation can still strand a self-created, unrecoverable transaction

**Disposition:** F-03-A remains open at an earlier actual-creation frontier.

**Locations:** [record writer](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_repair.go:461), [openFile completion](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_dir.go:193), [failed-preparation cleanup](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_repair.go:1296).

The common writer reports created=false whenever directory.openFile returns an error. But openFile first performs real Openat(O_CREAT|O_EXCL), then closes its temporary parent descriptor. If that Close reports an error, it closes the already-opened file and returns nil/error. The file has already been created. The writer therefore loses creation ownership before its new created=true assignment at line 466.

createControlRepairLocked receives preparedCreated=false, removes its known prior/intended anchors, and tries to remove the transaction directory. The newly created empty prepared.json prevents removal. A later fresh scan treats that file as prepared authority, fails decoding EOF, and refuses normal recovery before public control processing.

The reviewer probe performed that exact sequence using the existing CloseDirectory hook: real parent Close first, designated completion error second, armed only at BeforeControlRepairRecord for the prepared-authority label. Publish returned the completion error plus directory-not-empty; a new transaction contained only a zero-byte prepared.json. Prior CURRENT bytes stayed unchanged, but both Check and a fresh publisher's Recover returned “decode publication control repair prepared authority: EOF.” This is a desired-behavior RED, not a hypothetical power loss or a public-name attacker.

**Causality:** the broad failed-preparation root existed in the original range, including 8214. a9924626 introduced the created flag and its false return at this still-fallible helper boundary; 54746816 added record hooks without moving the ownership frontier. This is an incomplete repair of an existing-range miss, not a newly introduced healthy-publication defect and not something created by artifact-only evidence commits.

**Sibling scope:** every prepared writer using openFile has this earlier creation boundary, including CURRENT/JOURNAL successor and bootstrap classes. The common writer is also used by marker/phase creation; those callers must be inspected when changing its ownership/error result, without claiming that arbitrary malformed private state must be auto-repaired. Current Write/partial/Sync/record-Close and post-record directory-Sync cases do not exercise a parent-Close failure after successful creation.

**Required bounded closure:** establish known ownership at the first successful creation, including helper post-create failures. Preserve a coherent recoverable preparation or completely remove only the proven-owned, unexposed preparation. Do not delete an unknown replacement, weaken malformed-state refusal, or merely preserve an unreadable record and call it recoverable. Add this real-open completion frontier to the existing state-class matrix, retain both primary/completion and cleanup causes, assert prior controls and owned graph before restoration, and prove nonmutating Check, fresh recovery and normal retry where permitted.

### 7481-02 — Low — EEXIST retry discards newly joined completion/cleanup errors

**Disposition:** residual F-03-B consumer defect on the F-03-C allocator path.

**Location:** [failed temporary allocation retry](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication.go:1016).

After an exclusive temporary control open fails, lines 1020–1021 join the owned temporary-directory Close and identity-bound cleanup outcomes into err. Line 1022 then tests only errors.Is(err, fs.ErrExist). A joined EEXIST plus a meaningful Close or cleanup failure still matches, so the entire joined result is discarded and allocation retries. A subsequent successful allocation can make public Publish return nil.

The reviewer probe opened actual A, created its real control blocker, moved A, installed distinct nonempty directory B, then let the real control Openat fail EEXIST. It actually closed the owned A directory and appended a designated completion error. Publish returned a valid pointer and nil error. Exact A/B inodes and their “A blocker”/“B bytes” contents were preserved, proving that the ownership repair is useful but the completion outcome is still lost.

**Causality:** the broad EEXIST consumer originates in 4fedb387/958a07a7. a9924626 introduced the joined producer directly before that unchanged consumer. The lost-error root predates this wave, while the new joined-error-to-old-retry compatibility edge is fix-created by a9924626. This is not a renewed B-deletion claim.

**Sibling scope:** writeCurrentLocked and writeJournalLocked share the allocator; publication and recovery reach both. All error classification consumers were searched alongside the modified producers. The recursively pure NotExist predicate correctly prevents a joined absence/completion error becoming clean absence in ReadControlBound. Direct errno-only Mkdir/open/identity consumers are not equivalent to this joined path and were not falsely reported.

**Required bounded closure:** retry only a genuinely retryable collision whose meaningful owned completion/cleanup has succeeded, or otherwise preserve the complete error through the public caller. Assert errors.Is/As visibility for both the collision and injected completion/cleanup, one Close owner, exact A/B state and bounded retry. Do not turn all harmless read-only teardown into a new durability requirement or add a generic error framework.

### 7481-03 — Low — failed-allocation proof still substitutes existence and other control-write cuts for the audited A/B contract

**Disposition:** F-03-C's production A-bound cleanup is improved, but its declared executable completion is not established.

**Locations:** [CURRENT/JOURNAL fixture](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_group2_original_test.go:442), [replacement assertion](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_group2_original_test.go:527), [public temporary observer](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_group2_original_test.go:554), [public quarantine observer](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_group2_original_test.go:698).

The public Publish EEXIST fixture retains only A/blocker existence, B existence and optional foreign bytes. It does not compare retained A/B to captured pre-cleanup identities, and it ends after Publish rather than proving the audited fresh recovery/retry. The public stale-generation Prune quarantine fixture similarly checks existence/optional payload and old-generation existence, not exact original A/B identity continuity or fresh recovery.

The later CURRENT/JOURNAL fixture records A, moves its directory and creates B, but creates no control blocker and returns nil from AfterTemporaryOpen. The subsequent real control creation through retained A succeeds. The observed identity refusal is a later writeAtomic cleanup boundary, not failed allocator EEXIST ownership. It is useful coverage, but cannot establish equivalence to the missing failed-allocation sibling schedule.

The later shared assertion verifies exact A but only that the current B is a directory distinct from A; it never retains the B identity created by the actor. A distinct C at B's pathname can satisfy that check. The stale-stage quarantine case adds legitimate fresh recovery but inherits this B-continuity gap. Empty and nonempty directory variants do run; emptiness/type variants are not substitutes for exact B identity.

**Causality:** the public original-control skeleton is from 8d133782, the repaired existence checks from a9924626, and the new sibling fixture/assertion from 54746816. This is incomplete execution of the audited F-03-C proof obligation, including a repair-created boundary substitution. It is separate from 7481-02 because the absent identity/schedule/restart assertions remain even after truthful errors are fixed.

**Required bounded closure:** use descriptor-safe expected witnesses for both A and the actual actor-created B before the attempted cleanup. Execute the actual failed temporary allocation and failed quarantine allocation through the two audited public paths, with their meaningful empty/nonempty directory variants, exact causes and residue. Prove the required fresh recovery/retry after only explicitly fixture-owned restoration. Retain the CURRENT/JOURNAL successful-allocation cleanup test as complementary coverage, or clearly narrow its claim rather than calling it failed-allocation proof.

### 7481-04 — Low — durable “preservation” helper observes current state but does not compare it with required expected state

**Disposition:** F-05/F-06-P remains open despite real additional observations and executed variants.

**Locations:** [control/private helper](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_durable_matrix_test.go:222), [generation-root helper](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_durable_matrix_test.go:289), and their public-caller/rollback/Check call sites.

The helper reads CURRENT/JOURNAL and then compares a second witness with bytes just read at the same post-operation cut. It requires a regular file and a nonzero inode, not the expected retained control identity or expected raw bytes from the relevant earlier cut. It scans the current private graph, compares transaction identities to that freshly scanned graph, and checks that a fresh serialized tree contains prepared/phase names. It discards those snapshots. It does not compare the pre-existing marker/transaction/prepared/phase/anchor history to an expected retained history with explicitly permitted new records.

For selected/rejected/stale generation roots it requires a directory and a nonempty serialized snapshot, but compares neither expected root identity nor expected content. Even an encoded empty snapshot has nonzero byte length. A current coherent graph or a nonzero inode does not establish intermediate preservation.

The reviewer proof control retained a real old generation across a new publication, moved old root A, then installed an unrelated B directory at the old generation pathname containing only a foreign file. The actual durable-cut helper returned normally. The control then deliberately failed its desired rejection assertion. This demonstrates an oracle false positive; it does **not** demonstrate that production Open, validation or cleanup accepts B.

The strong A/B lease and returned-reader witnesses remain useful and accepted. The held-reader matrix still checks decoded expected controls plus its A/B witnesses without the required expected private/root/raw-control crosswalk. Later successful fixture restoration/fresh recovery cannot retrospectively establish those missing pre-restoration comparisons.

**Causality:** the F-05/F-06 preservation obligation is older than this wave. These self-referential helper checks were introduced by 7481d177; the new evidence text overstates what they establish. The former “ordinary Publish has no authority” sentence is now explicitly corrected and is not retained as a separate finding.

**Required bounded closure:** capture exact expected pre-existing controls, generation roots/content and private identities at the appropriate durable boundary; compare after refusal, explicitly permitting legitimate CURRENT/JOURNAL/phase advancement and newly appended transactions. Distinguish unchanged historical authority from newly created authority. Add an oracle control that would reject a substituted expected root/history rather than merely finding a readable current tree. Apply the contract to every named cut below, including held-reader and read-only Check, without requiring a successful Publish or rollback to retain a pre-Publish CURRENT inode that it legitimately replaces.

### 7481-05 — Low — new nested “final Publish prune” row actually fails during initial recovery

**Disposition:** F-02-P's required public Recover/Prune recursion is now exercised; this added cut claim is fix-created and false.

**Locations:** [nested Publish fixture](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication_resource_error_test.go:193), [actual Publish order](/Users/karthiksivadas/.treehouse/cli-6bae67/2/cli/cmd/connectorgen/vnext_publication.go:395), GROUP3-EVIDENCE.md's 061/062 and 063 tables and derivative canonical status.

The fixture publishes nested old, holds old while publishing current, then releases old before invoking guarded Publish(next). At entry, old is already an unheld stale generation. publishLocked calls recoverLocked before staging next; no-journal recovery removes stale generations. The installed nested hook therefore refuses during this initial recovery and Publish never reaches its later successful-selection/final-prune cut.

The test asserts the shared nested hook and an error, not the expected new CURRENT/committed JOURNAL or a final-prune-only boundary. The weak generic control helper from 7481-04 cannot detect the wrong schedule. Calling the row “successful Publish final-prune cleanup” in the table is therefore unsupported.

The actual public Recover-owned-stage and Prune-stale-generation nested rows are meaningful: they cross real quarantine, execute nested replacement and actual-open post-Stat failure, compare quarantined root identity and replacement-case A/B identities/bytes, repeat four times with GC disabled, reconstruct only the test fixture, then check fresh recovery and root-scoped descriptor release. The direct 24-refusal test and immediate child defer remain accepted. No current production FD leak is asserted.

The post-Stat failure row checks nested payload but does not retain/compare the nested child's pre-failure identity despite the table's “retained-child identity” wording; either assert that identity or describe the narrower observation honestly. Already removed owned siblings and partial quarantine residue remain permissible.

**Causality:** the incorrect fixture and final-prune label were introduced together in 7481d177. This is a new proof-schedule/evidence defect, not proof that F-02's original leak persists, and not a reason to demand all-or-nothing recursive deletion.

**Required bounded closure:** make the named nested Publish row reach the true final prune with asserted selected/committed controls and relevant resource state, or explicitly withdraw the unsupported extra final-prune claim while preserving the genuinely executed mandatory Recover/Prune proof. Keep the separate real final-prune lease matrix; it is not nested recursive ownership evidence. No new process-restart or runtime guarantee is required.

## Seven-entry audit disposition and sibling search

| Audited entry | Current disposition |
| --- | --- |
| F-03-A | Partially repaired, not closed: coherent retention after successful record completion and later transaction/connector Sync is present; known partial records can be removed identity-bound. Earlier successful-create/parent-Close frontier remains 7481-01. Strict valid absence/bootstrap remains accepted. |
| F-03-B | Partially repaired, not closed: meaningful producers now join errors and pure absence is distinguished; EEXIST consumes and loses joined completion/cleanup in 7481-02. |
| F-03-C | A-bound production cleanup retained; original empty-B deletion is not reasserted. Error visibility and exact failed-allocation proof remain incomplete under 7481-02/03. |
| F-04-R | Resolved: file witness derives type/identity/bytes from one opened no-follow/nonblocking descriptor; directory witness retains A while reading dependent metadata. Bounded FIFO/symlink/directory controls and ordinary nested positive are real. |
| F-05/F-06-P | Partially repaired: all principal logical cuts and meaningful empty/nonempty regular lease-B variants now execute; ordinary authority existence is corrected. Expected intermediate raw/control/private/root preservation remains 7481-04. |
| F-02-P | Core public nested Recover/Prune coverage added and no current leak found. New final-Publish and retained-child evidence claims require 7481-05; shared durable-state assertions inherit 7481-04. |
| F-08-R | Resolved: actual EWOULDBLOCK/EAGAIN acknowledgement on the retained locked directory precedes signalling; pre-flock directory-open is explicitly insufficient; exact sent OS signal, bounded lifecycle, no-success/state/retry assertions retained. |

The F-03 search covered actual open plus parent Close; parent Close plus already-opened-file Close; definitions/connector root completion; opened-control identity/read plus Close; writable marker/prepared/phase Write/Sync/Close; predecessor link registration and Close; stage file Write/Sync/Close and known-owned stage creation cleanup; capture identity/primary error plus Sync/Close; pure and joined absence; and conflict/collision consumers. The known predecessor anchor is registered before its fallible Close. Record/stage/capture primary failures are not flattened into generic strings. Read-only teardown policies were not recast as blanket new durability promises.

The F-01 search covered all three recorded capture-dependent opens for both CURRENT and JOURNAL, transaction/predecessor reopens and empty-capture admission. The retained helper compares actual Fstat identity with recorded capture identity before dependent enumeration/candidate use/rename/sync. The six actual-open tests place replacements at their real dependent opens, assert exact A/B/public state, restore only fixtures and use fresh recovery. No current capture-handoff defect was found.

The F-02/F-05/F-06 search covered public and internal cleanup sinks, nested children, early and late lease replacements, root replacements, public C collision, failed allocation, rollback, ordinary and interrupted publication, held handles and Check. The different object kinds remain separate: F-03-C B is a directory, F-05/F-06 B is a regular lease file, F-02 B is a nested directory.



### Current gate correction and proof custody supplied to the audit

The full current 064 report with 068 appendix has SHA-256
`a59da5a5752e5c2149ca474cf5b9ac5761fa494fe7ae470ee04084a5680ba779`.
The older `a9bcdf60...` hash in the untouched original review predates that
appendix. The recovered original boundary result file has SHA-256
`744d0d129e15c8eccbaf723dda0ca96487c8babc25ce3c2d53d6feea25ed5849`;
physical 13951 / raw including-newline SHA-256
`ec88ff8c01c57e8647206c78ff003698932b66be9420a737536073ddd75c07b7`
records exit 0, wall 254.579597042s, clean 284 files/553 connectors, zero
findings/warnings and six existing exceptions. No scanner rerun occurred.
This supersedes only the stale unresolved-scanner cells, not the separate
boundary-package `invalid_cli_surface` failure and original-package-source
under current dependencies CP29 debt.

Final intended-source package normal 319.928s / race 783.834s and focused
32.789s / 45.628s remain their original pre-commit receipts. Earlier
315.166s/763.610s and 30.270s/42.177s remain different source states. The
current package-receipts file SHA-256 is
`738193c2a30663995bb0e8904b131c244e8909f70f7a52b39b1840384ed27d6a`.
All eight supplementary gates have successful current results; runtime
preflight is explicitly cached. No refreshed whole-original-range lint
receipt exists: Group 3 precommit new-only zero is not whole-range green.
Historical 20 global/15 package diagnostics and 31 introduced diagnostics,
CLI/engine/App/contract/Redis obligations and Linux CGO limitation remain
separately bounded debt, not current acceptance.

Original review probes retain test SHA-256
`76cd08d998041871dcda4b576c6f09776b4848f7559a77caa87549e67dc3f7ef`,
additional-nonexistent-test-path overlay SHA-256
`bf669b5c23665f21aa6d04e7341a2466e640225afabf23d9db07a99500a6a8e3`,
and transcribed receipt SHA-256
`70da86171316ddce5f93ae168b7dc186275cdfc4aad6e9ea8036e0600107e2ea`.
One three-probe command exited 1, package 2.560s/wall 5.59s: two desired
production regressions and one oracle counterexample. A transcribed receipt
is not an original rollout export, and failed assertions are not repaired GREEN.

Pending audit dispositions include derivative Group 2 completion claims and
Group 3 expected-state and nested-final-prune claims. Dated 068 boundary
corrections and pending-five-finding notices have now been appended to
PLAN/TDD/VERIFICATION/GROUP3; their original historical text is preserved.
The original fresh report is unchanged; audit disposition precedes any
coordinated repair/proof scope.
The complete source membership, seven lanes/modes, architecture and CP12–31
programme remain unchanged and unaccepted beyond the prior bounded CP01–10.

### Independent audit launch receipt

Firstmate074 inbox message 072 was read and acknowledged; it confirms the
preserving owner session `01a07580-22bd-7e33-be7c-94728fc61408` and authorizes
continuation under 073. Native `collaboration.spawn_agent` accepted a fresh
`/root/cp11_7481_complete_ledger_audit_astra_r1`, explicitly
`model=gpt-6-astra`, `reasoning_effort=xhigh`, `fork_turns=none`, using the
complete Firstmate073 specialist prompt. `list_agents` reports the owner and
this one auditor running. This is the worker-owned native execution of the
installed GSD review/audit fallback, not a new Firstmate root or repair agent.


## 2026-09-06 — Firstmate073 complete independent audit returned

The one authorized fresh native Astra/xhigh auditor
`/root/cp11_7481_complete_ledger_audit_astra_r1` returned its complete final
answer and saved report. The owner read all 315 lines / 52,818 bytes and
independently verified SHA-256
`bc109e85fdde9d1958b2cde7874a3f7b30b8e5d06b1b0c2764088fb34fa3e0a0`.
The full original report is preserved verbatim between the markers below.
Its immutable range is `36e4d980de0d51d92fe74a68306845643596a6cb` through
`7481d1770a21cc95869fd10bf281f632af48c089`, candidate tree
`a2e583336ffa8ad86a0de95110259342bfa6dab0`. The report's observed evidence-only
HEAD is `afde575a1368fe86b5ab5acb5a3d39a294a20e96`; this subsequent ledger
aggregation does not create a new behavioral review candidate.

**Current verdict: CHANGES REQUIRED; six adjudicated actionable entries,
one Medium and five Low.** The original five survive with explicit
amendments; 7481-06 adds an omitted existing F03B compound-proof obligation.
There are two production invariants needing repair and four proof/evidence
entries. This is the complete frozen response set, not a repair authorization.

| Finding | Audited disposition and remaining bounded obligation |
| --- | --- |
| 7481-01, Medium | Confirmed common-record ownership defect, including earliest successful creation before parent Close and prepared/marker/phase partial-record siblings. Strict malformed-state refusal and coherent known-owned cleanup remain required. |
| 7481-02, Low | Confirmed compound allocation error lost by collision retry. Preserve pure collision behavior, meaningful typed causes, one Close owner and actual A/B ownership through the complete public caller chain. |
| 7481-03, Low | Confirmed failed-allocation proof gap. Retain actual A and actor-created DIRECTORY B, correct direct CURRENT/JOURNAL later-cleanup labels, and supply existing public/direct/stage/generation recovery obligations. No current B-deletion defect is inferred. |
| 7481-04, Low | Confirmed expected-state oracle gap, including F03A stable-history siblings, all eleven cleanup caller/cut rows, held-reader and separate nonmutating Check. Legitimately advanced controls/phases differ from stable pre-existing history. |
| 7481-05, Low | Confirmed unsupported extra nested final-Publish claim: that fixture reaches initial recovery. Mandatory nested Recover-stage/Prune-generation and actual final-prune lease proof remain. Correct the label and post-Stat child-identity overclaim; extra nested final-prune execution is optional. |
| 7481-06, Low | Added omitted existing F03B proof: Write/short-Write plus Close, failed-link plus Close, capture primary plus Sync/Close, opened-control identity/stat/read plus Close, actual staged completion and per-resource Close ownership. No additional production defect is asserted. |

The report preserves every old-seven disposition: F04R/F08R resolved;
F03A/F03B/F03C/F05-F06P/F02P remain partial in the specified respects.
B01/B02, R3, actual F01 capture opens, returned validated A, directory lifetime
locking, late lease and original F07 RED/edit/GREEN chronology remain protected.
All ten lenses and the complete ownership/consumer and actual caller/cut
contract-to-evidence tables are included verbatim below.

The additional marker/phase overlay completed exit 1, package 1.079s /
wall 5.26s, with four desired fresh-recovery failures. Its actual test,
overlay and original tool command/result capture hashes were independently
verified against the report. No project test was added or changed; no supplied
probe or completed normal/race/supplementary suite was rerun. The original
three-probe transcription remains distinct from the new original tool-result
capture and from a rollout export. All auditor commands are terminal and it
created no child. Firstmate075/076 heartbeat messages (inbox 073/074) were
forwarded to this same context and acknowledged, without new technical scope.

Current original-base-range lint evidence remains a CP11 acceptance gate;
Group3 new-only zero does not close it. Historical 20 global/15 package and
31 introduced diagnostic groups remain their exact separate attributions.
The corrected 068 boundary scanner success remains current; boundary-package
and other CP29 obligations are not promoted. The preserved patch archive is
the only recorded original-range diff-check exclusion; this new evidence
slice has a separately scoped whitespace check.

The owner will commit this complete evidence-only aggregation on
`fm/cli-top100-declaration-batch-r1` and report its exact identity to Firstmate
for a separately authored coordinated repair/proof brief. CP11 remains
unaccepted; CP12, implementation repair, no-mistakes, publication and merge
remain outside this audit completion. Existing delivery:
https://github.com/polymetrics-ai/cli/pull/4294.

<!-- BEGIN VERBATIM CP11-7481 COMPLETE AUDIT bc109e85 -->
# CP11 complete-ledger audit — 7481, Astra/xhigh, round 1

**CHANGES REQUIRED. Frozen set: six actionable entries — one Medium and five Low.** All five submitted entries survive with the bounded adjudications below. **7481-06** adds an omitted, already-required compound-error proof obligation. There are two production invariants requiring repair and four proof/evidence entries. Marker and phase record failures are confirmed siblings of 7481-01, not separate findings.

This is the complete independent judgment for Firstmate's coordinated response. It is not implementation authorization, acceptance of a future repair, publication, CP12, no-mistakes, release, or merge certification. No individual finding was sent for early repair.

## Identity, authority and scope

| Field | Binding |
| --- | --- |
| Primary / parent issue | 4427 / 4325, CP11 |
| Existing delivery | https://github.com/polymetrics-ai/cli/pull/4294 |
| Worktree | /Users/karthiksivadas/.treehouse/cli-6bae67/2/cli |
| Branch | fm/cli-top100-declaration-batch-r1 |
| Original base and verified merge base | 36e4d980de0d51d92fe74a68306845643596a6cb |
| Behavioral/test candidate | 7481d1770a21cc95869fd10bf281f632af48c089 |
| Candidate tree | a2e583336ffa8ad86a0de95110259342bfa6dab0 |
| Observed evidence-only HEAD | afde575a1368fe86b5ab5acb5a3d39a294a20e96 |
| HEAD tree | 41bead7def7472a0383b1081b1134c9e19803624 |
| Verified published ancestor | 8214bd91403ce620773b61caf674faa540ee1701 |
| Original range | 52 paths, 20,672 insertions, 160 deletions |
| Current merged REVIEW-CONVERGENCE SHA-256 | 8dd3142eb4a3abc3efe8dd10ab6099ec7571ed424c631e06c2c2497fdba5e97c |

The earlier intake hash was 8fb9b29a37a3d3f0104bc245793fbd90ebdedb63c64f9f1f72266da93c44c20c. The owner subsequently corrected the intake paragraph to say that the dated 068 boundary corrections and pending-five-finding notices had already been appended. I read that update. The change is evidence chronology, not a changed disposition or behavioral candidate.

HEAD's descendant changes are four planning files only. The observed worktree difference from 7481 is limited to PLAN, TDD-LEDGER, VERIFICATION, GROUP3-EVIDENCE and REVIEW-CONVERGENCE under .planning/phases/batch-r1-vnext-cutover. The explicit comparison of cmd, internal, Makefile, go.mod, go.sum, scripts, docs and website to 7481 exited 0. The pre-existing untracked .cache/ was neither read nor changed. No project file, index, commit or ref was modified by this auditor.

I used the pinned [firstmate-exhaustive-review skill](/Users/karthiksivadas/pm-cli-agent-workspace/data/review-system/.agents/skills/firstmate-exhaustive-review/SKILL.md) and the complete approved [REVIEW-EVIDENCE-REQUIREMENTS.md](/Users/karthiksivadas/pm-cli-agent-workspace/data/batch-one-execution/REVIEW-EVIDENCE-REQUIREMENTS.md). The captain's model and no-publication overrides were applied; the complete-ledger and ten-lens requirements were retained. Project AGENTS, required-skill routing, task delivery header, delivery contract, issue contract, GSD adapter and CLI parity references were read. Go how-to, error handling, security, safety, testing, lint, CLI, context and concurrency guidance informed the review. CodeGraph was absent, so source inspection used Git and targeted reads without indexing.

The full 380-line fresh review, complete merged convergence record, full prior e77 audit and original-return provenance, coordinated 051 repair brief, expanded 058 obligations, 063 object-kind/caller correction, 067 handoff, and CP11 GROUP1/GROUP2/GROUP2-F03/GROUP3/TDD/VERIFICATION evidence were reviewed. Mechanical indexes supplied locations and receipt bindings; they did not establish correctness.

All source references below refer to 7481. Abbreviations used in the tables are exact paths:

- P: cmd/connectorgen/vnext_publication.go
- D: cmd/connectorgen/vnext_publication_dir.go
- R: cmd/connectorgen/vnext_publication_repair.go
- G2: cmd/connectorgen/vnext_publication_group2_original_test.go
- DM: cmd/connectorgen/vnext_publication_durable_matrix_test.go
- RE: cmd/connectorgen/vnext_publication_resource_error_test.go
- PT: cmd/connectorgen/vnext_publication_test.go

## Complete finding dispositions

### 7481-01 — Medium — confirmed, with marker/phase siblings

**A successful exclusive record creation can be reported as uncreated; other common-writer callers abandon self-created incomplete records.**

D:193 openFile performs real Openat with O_CREAT|O_EXCL before closing its temporary parent. If that Close fails, it closes the newly opened raw descriptor and returns nil plus an error. R:461 sets created=true only after openFile returns successfully. Its error return therefore reports created=false even though an empty record already exists.

R:1288 createControlRepairLocked uses preparedCreated to decide whether to remove the unexposed record before disposing of registered anchors. On this boundary it removes anchors while leaving the empty prepared.json. The supplied external probe demonstrates the retained empty record and unchanged public CURRENT; fresh Check/Recover refuse EOF. The original post-complete-record repair and later Write/Sync/Close tests did not cover this earlier side effect.

The common writer's other callers are substantive siblings. R:709 WriteAuthorityMarker and R:1459 appendControlRepairPhaseLocked discard created and return on a failed writer. Both can leave their own malformed record after a partial Write, or after successful creation followed by the parent's Close completion failure. I ran one bounded additional-test overlay against unchanged source. Four cases — marker/phase multiplied by parent-Close-after-create/actual partial Write — captured the actual created regular-file identity, found that same object with zero or one byte, and demonstrated failed fresh recovery with EOF. These are production desired-behavior REDs. They do not show that arbitrary malformed private state should be accepted.

The closure unit is the entire common record writer and its prepared, authority-marker and phase callers. Record the first successful creation and ownership before helper completion can hide it; preserve every meaningful cause and exactly one actual Close owner. On failure, either retain coherent authority or remove only a proven-owned, unexposed incomplete record through its bound identity. Do not remove required anchors first, use a pathname as ownership, erase a replacement occupant, or discard established history. A complete durable phase may already authorize a capture or selection; it is not interchangeable with an incomplete unexposed phase.

Valid classes remain distinct: absent/absent bootstrap can legitimately have no anchors; present/present bootstrap must use equal states; successor CURRENT/JOURNAL cases include present/absent prior and intended state as actually reachable. Recovery must continue refusing unknown malformed graphs. Required proof includes the newly identified earliest-create boundary and partial Write, existing record/transaction/connector completion boundaries, pending nonmutating Check, and bounded fresh recovery/retry with expected controls/history at each applicable class.

**Causality:** the broad failed-preparation ownership root predates the latest wave and exists in published 8214. a992 introduced the created result but placed its frontier after the fallible helper; 547 added hooks without closing it. Marker/phase incomplete-record handling is an older reachable common-writer sibling, not a new production defect introduced by 7481's tests.

### 7481-02 — Low — confirmed

**A retry classification drops a compound allocation outcome and can report successful Publish.**

P:987 CreateTemp first creates and opens A, obtains A's identity, and then exclusively opens its control file. Its failed-control-open path now joins the owned directory Close and identity-bound cleanup outcomes into err. P:1022 still decides retryability with errors.Is(err, fs.ErrExist). A matching collision is not evidence that it is the only cause. A real Close or identity/cleanup failure is lost when the next allocation succeeds.

The supplied public Publish probe reached the real exclusive-open collision and real Close followed by injected completion, retained actual A/B identities and bytes, and got nil from Publish. The current G2 public temporary replacement fixture also expects nil despite the identity-bound cleanup refusal; this expectation needs reconciliation with the repaired public error contract.

The producer reaches writeAtomicLocked, writeCurrentLocked and writeJournalLocked. Their consumers include prepared JOURNAL creation, CURRENT selection, committed JOURNAL replacement, rollback restoration and recovery restoration; Recover is also reached from Open, Prune and Publish's initial recovery. Quarantine allocation does not have the same post-join collision retry: its after-open failure returns the joined outcome. A pure Mkdir collision may retry. Intentional forensic retention/ENOTEMPTY must retain its stated policy; it is not automatically another durability error. A real added Close failure or ownership refusal cannot disappear behind the collision.

The pureNotExist predicate checks all joined children, and ReadControlBound treats compound absence as found/error. I found no corresponding new false-absence defect. I also traced typed control conflict and selection-install handling: no additional reachable compound-to-success conflict defect was established. Read-only teardown policy is not converted into a blanket durability requirement.

**Required repair:** classify the complete outcome at the retry boundary, preserve typed causes and bounded attempts, demonstrate pure collision versus the meaningful compound through the actual producer and public caller, and retain A/B ownership.

**Causality:** a992 introduced this joined-error-producer to old-retry-consumer edge. The broader history of discarded cleanup errors is older; neither “entirely new in 7481” nor “unchanged old error loss” is an adequate attribution.

### 7481-03 — Low — confirmed required proof, amended production implication

**Failed-allocation coverage does not establish the stipulated actual A/B identities and recovery for every caller.**

G2:428/554 public Publish and G2:604/698 public stale-generation Prune do reach real failed allocations, with empty/nonempty replacement DIRECTORY variants. Their assertions mainly establish existence and optional bytes. They do not retain both actual A and actor-created B identities before cleanup or supply the promised fresh recovery/retry for those fixtures.

G2:442's direct CURRENT/JOURNAL rows move A and install B at AfterTemporaryOpen but create no control blocker. The exclusive control-file allocation succeeds through retained A. The later error comes from deferred temporary-root cleanup after transition work. This is useful complementary cleanup coverage; it does not execute failed allocation for those explicit callers.

G2:527's shared assertion checks expected A, then checks only that B is a directory different from A and, when supplied, has particular payload bytes. Unrelated C can pass. The empty-directory variant does not independently prove the retained actor-created B identity or all its expected contents. G2:618's stage quarantine fixture includes a retained A and fresh retry, but shares the deficient B assertion.

The exact 058 F03-C obligation already required CURRENT/JOURNAL temporary uses, stage/generation quarantine reachability, actual identities/types/bytes/residue before restoration, meaningful errors and bounded fresh recovery/retry. The public and direct rows cannot silently substitute for each other. Either execute the missing failed-allocation rows or present a precise equivalence for independent adjudication that preserves the explicit contract. The existing direct successful-allocation/later-cleanup rows may be retained with narrower labels.

**No current production B-deletion defect is inferred from this proof gap.** The old empty-directory deletion and the new cleanup/error contract are separate facts. a992's bound cleanup is a real repair; incomplete assertions do not undo it.

**Causality:** residual old F03-C verification, with later direct fixtures claiming a failure phase they do not reach.

### 7481-04 — Low — confirmed, with F03A expected-history and caller omissions

**The central durable oracle verifies self-consistency of the observed state instead of preservation of independently expected state.**

DM:222 reads CURRENT/JOURNAL and their witnesses after the operation, then compares the same current objects again. It scans current private authority and compares entries to their own current records. DM:289 requires a root directory and nonempty serialized snapshot; it is not given the expected root identity, expected bytes or expected prior history.

The supplied oracle counterexample moved expected root A away, installed an unrelated readable directory B and invoked the helper. The helper accepted B. The probe then deliberately failed the oracle-rejection assertion. This is an oracle false positive, not evidence that production accepted B.

The defect extends to witnesses used in the F03A proof: G2:330 records old transaction names only; AssertNoPreparedGraph skips those names and does not require them to remain. AssertPendingPreparedGraph derives prior/intended identities from the post-operation record. These checks can establish useful internal consistency but cannot prove that independently expected old authority, controls and anchors survived. That is the same expected-state invariant, not an additional production finding.

DM's new empty lease-file B helper captures B only after invoke returns. Distinct-from-A cannot identify the actor's B at the destructive boundary. Several older rows do retain stronger actual lease A/B witnesses; they must not be weakened to the central helper.

There are also explicit caller omissions. The held-reader row in PT:101 checks decoded controls without the durable expected-history helper. The stage cases around PT:1993 and PT:2285 have marker/root/identity observations of differing strength but do not supply the promised prior raw-control/private-history comparison. Check at PT:788 retains raw CURRENT bytes, but stale-root existence and a post-observed authority helper are not the complete independently expected cut. An assertion absent from those actual rows cannot be supplied by another row's passing name.

**Required proof:** retain expected raw control types/inodes/bytes, actual root identities/content, stable authority marker/prepared/phase/anchor history, and actual actor-created A/B witnesses before the operation or at its proper durable cut. Separate stable pre-existing history from legitimate new transactions/phases. Specify expected new CURRENT/JOURNAL rather than requiring normal Publish or rollback to preserve an inode it legitimately replaces. Add bounded negative controls proving the important helpers reject unrelated roots, same bytes/different expected identity, and missing/replaced stable history. Apply the corrected comparison to every caller/cut in the table below, including held-reader and Check.

**Causality:** residual F05/F06 and F02 proof incompleteness, plus inadequate oracle strengthening in the latest proof wave. The F03A helper sibling is another existing proof omission under the same invariant.

### 7481-05 — Low — confirmed as an unsupported extra claim; no invented final-prune requirement

**The nested Publish fixture reaches initial recovery, not final prune.**

RE:137 builds an old generation with nested content, holds it while publishing current, then releases the handle. At entry to Publish(next), old is already an unheld stale generation. P:391 calls recoverLocked before staging next, selecting CURRENT, committing JOURNAL or entering final prune. The recursive child hook therefore fires while initial no-JOURNAL recovery prunes old. The fixture's nested final-prune claim is false.

The existing contract's mandatory public nested rows are Recover removing an owned stage and Prune removing a stale generation. Those rows reach real quarantined recursion and preserve useful no-GC descriptor/resource proof. The direct recursive descriptor repair also remains valid. The contract does not require adding another nested final-prune schedule merely to save an extra label. **The bounded closure is to withdraw/correct the nested final-Publish claim and derivative evidence.** A genuinely reached additional final-prune nested row is optional.

The separate DM successful-Publish final-prune lease tests start from old as the current generation and assert the new CURRENT with committed JOURNAL at refusal. They do reach final prune. They prove lease identity behavior; they are not nested recursive-descriptor proof.

For the post-Stat nested case, the hook performs real Stat and then returns an injected error; it is not a natural Stat syscall failure. The fixture retains nested bytes and outer root identity, but does not retain a pre-operation nested-child identity for the claimed identity-preservation assertion. Narrow that assertion to what is observed, or add that exact child witness. No actor-created replacement B exists in the post-Stat-error case. Earlier legitimate deletions and retained quarantine residue are allowed; recovery proof must reconstruct only fixture-owned known content. Neither all-or-nothing recursive deletion nor a stage/Check generation lease is required.

**Causality:** this fixture/claim was added in 7481. It is not an original unchanged production leak.

### 7481-06 — Low — added: omitted mandatory F03B compound proof

**The latest ledger marks the explicitly expanded error matrix covered although several stipulated compound cases are absent.**

This is an omission in the fresh five-entry review, deduplicated as one remaining F03B proof invariant. It does not assert a newly demonstrated production error-loss defect beyond 7481-02. The exact 058 instruction explicitly rejects replacing stipulated cases with source changes, a package PASS, or another member of the class.

G2:783's actual assertions establish several useful compound cases, but:

- All three record labels execute short Write alone, and separately full Write plus real Sync/injected completion plus real Close/injected completion. They do not execute Write(error)+Close or short-Write+Close together. Manually deleting the raw test record afterward does not prove public recoverability.
- The predecessor case and RE:816 exercise successful linking followed by Close failure. They do not exercise failed link plus Close in the same invocation. Registration-before-Close is correctly implemented and must remain.
- The capture case starts with a successful identity/read path, injects a post-Sync completion and then Close completion. It does not preserve an earlier primary failure together with both Sync and Close causes. The BeforeControlRepairCaptureClose seam is not exercised by that matrix.
- The opened-control case injects a read-completion error after the bounded read succeeded. It does not execute the separately stipulated identity/stat/read failure boundaries plus Close. Pure absence and compound missing-parent-close cases are real, useful covered rows.
- The staged “sync and close” case injects BeforeFileSync, so it is a pre-Sync primary failure plus actual Close completion. It cannot be reported as a completed Sync failure. Bind another exact test if it covers the required writable completion class, or complete that narrow missing row.
- One-close ownership is counted explicitly for records/predecessor, but is not independently counted throughout all claimed compound controls. A Close hook performing a real Close is useful evidence; it is not by itself the whole per-resource owner accounting.

**Required proof:** complete or precisely bind those existing narrow compound cases, keep every meaningful cause visible at the relevant consumer/public caller, account for the actual resource and one Close owner, and observe the contract's state before restoration. Shared cases are appropriate; no Cartesian fault explosion, generic error framework or natural-I/O simulation is required. Public coherence after partial record creation belongs to 7481-01, and missing expected state belongs to 7481-04; do not duplicate those as further findings.

**Causality:** incomplete 051/058 repair verification and overbroad later coverage disposition. It is not new architecture or a newly created production bug. The current “F03 dynamic matrix closed” statement must be corrected.

## Contract-to-evidence: ownership, completion and classification

Every row binds to candidate 7481 and the exact paths defined above. “Current package” means the owner receipts identified below, not a new run by this auditor. “Gap” means unresolved acceptance evidence.

| Invariant and source | Public caller / actual phase; already-completed side effect | Resource owner and required state transition | Actual evidence and disposition |
| --- | --- | --- | --- |
| Root/operation ownership, P:225–338 | Publish/Recover/Open/Prune/Check entry; definitions and connector roots opened before completion error | openConnectorRoot must close its owned roots and preserve meaningful multiple failures; no success handle after failed setup | G2 “definitions and connector close” performs both real closes and asserts both causes. Current package covers it. No additional production finding. |
| openFile failure, D:193 | Common record, temporary and read-control callers; failed Openat can still be followed by parent Close failure | Temporary parent is helper-owned; preserve open and Close errors; no descriptor transfer on failure | G2 failed-open+parent-close and parent-close+opened-file-close rows use real operations. Success-create ownership is incomplete: 01. |
| Earliest record creation, D:193 → R:461 | CURRENT/JOURNAL preparation and common marker/phase creation; O_EXCL creation succeeds before parent Close | Helper initially owns fd; caller must receive truthful creation/identity state or helper must complete bound cleanup. Error is not no-side-effect evidence | Supplied prepared probe and four new marker/phase probes fail desired fresh recovery. 01 mandatory repair. |
| Record Fstat/Write/short Write, R:461 | Prepared record before authority exposure; marker or next phase at its own append cut | Preserve created identity; one actual writable Close; incomplete bytes must not become accepted authority; remove only proven-owned unexposed state | G2 frontier matrix plus raw common-writer tests cover useful pieces. Actual partial marker/phase Writes newly fail recovery. 01 and 06; unknown identity is never deletion authority. |
| Record Sync/Close, R:461 | Bytes already written; successful real operation may be followed by injected completion error | Preserve all meaningful causes and distinguish valid complete record from incomplete preparation; no double Close | Existing F03A Sync/Close and G2 common-record Sync+Close controls are real-resource coverage. Write+Close siblings remain 06. |
| Anchor linking and predecessor completion, R:1288 | Successor preparation; predecessor anchor may already be linked when Close fails | Register successful link before fallible Close; remove only registered own unexposed links; retain predecessor and required history | Source does register link first; RE:816 and G2 successful-link Close coverage retained. Failed-link+Close proof remains 06. |
| Prepared cleanup and post-record directory Sync, R:1288 | Prior/intended anchors and complete prepared record may already exist before transaction/connector Sync failure | Complete owned cleanup or coherent graph; never remove anchors while retaining a record that requires them | G2 F03A matrices and base tests supply real completion injection and fresh recovery. Earliest creation remains 01; expected stable-history oracle remains 04. |
| Authority marker, R:709/1178 | Both valid base heads may exist; marker write fails before ensure returns | Retain coherent bases/history; own incomplete marker cannot strand ordinary recovery; complete marker remains authoritative | New marker zero-/one-byte identity-bound probe failures, 01. No permission to parse arbitrary malformed marker as valid. |
| Phase append, R:1459 | Prior prepared/phases exist; public/capture state depends on phase being appended | Preserve prior valid history and completed transitions; incomplete new phase differs from completed phase plus directory Sync error | New phase zero-/one-byte probes prove 01 at first bootstrap phase. All phase callers must retain the same bounded contract; no speculative deletion of completed phases. |
| Capture setup, R:1493 | Capture Mkdir/open/identity/read may already succeed; Sync and Close run after earlier errors | Capture descriptor has one owner; primary, Sync and Close causes survive; no capture intent from invalid/replaced directory | Six checked-open F01 tests remain strong; G2 post-Sync+Close is partial compound coverage. Earlier-primary+Sync+Close remains 06. |
| Recorded capture reopen, R:1008/1575/1643 | Validation, completion and selection reopen a recorded directory | Compare actual opened descriptor to recorded identity before dependent read/link/rename; preserve A, actor B and public bytes | Capture identity tests cover CURRENT/JOURNAL across six actual opens, distinct pre-recorded A/B, unchanged public bytes, empty captures and fresh recovery after restoring A. F01 retained. |
| Temporary Mkdir/open, P:987 | A exists before its control-file allocation; opening/identity acquisition can fail | Retain only demonstrably owned cleanup; do not delete a guessed path after failed identity acquisition | a992 bound cleanup corrects old B deletion. Pure Mkdir EEXIST remains bounded retry; failed control allocation compound error remains 02. |
| Temporary collision consumer, P:1016–1024 → P:890 | Real O_EXCL collision, then real Close/cleanup outcome, before retry | Preserve meaningful noncollision causes and A/B; only genuinely retryable outcome can become another attempt | Supplied Publish probe returns nil erroneously. Public G2 nil expectation needs correction. 02; full caller chain traced, not a search-only judgment. |
| CURRENT/JOURNAL atomic writer, P:848/864/890 | Temp Write/Sync, repair transition and public selection can precede deferred temp Close | Return completion failure even if selection advanced; recovery follows actual durable cut, not automatic rollback | RE atomic Close, rollback validation+Close, restoration-callers and CLI-output tests remain useful current-package coverage. 02 feeds them; 03/04/06 qualify missing proof. |
| Absence classification, P:1262–1347 | Missing control Open may be joined with parent Close; existing read may already complete before Close | Only pure absence means not found; compound errors remain observable; read-only policy stays bounded | pureNotExist walks joined children; ReadControlBound preserves found/error for compound absence. G2 pure/compound controls covered; earlier read-failure compounds remain 06. |
| Conflict/selection classification, R:1643/1702/1871 | No-replace link/selection or terminal phase may already complete | A conflict must not erase unrelated completion failures or authorize clobber; outer completion still propagates | Actual chain examined. No additional compound-conflict false success established. Keep compatibility check when changing error producers. |
| Stage creation/write/activation, P:646/761/783/745 | Owned stage, marker/files and closed-set validation precede no-replace activation | Writable resources close, failed owned cleanup reports meaningful causes, destination collision preserves both identities; unknown malformed stages remain refused | Staged primary+Close and failed-stage cleanup cases exist; stage oracle and completion-label limitations are 04/06. PT final activation collision retains exact stage/destination. No generic cleanup of unknown stages authorized. |
| Quarantine allocation, P:1037 | Quarantine directory A created/opened/identified before injected allocation failure | Close A and remove only matching empty A; retain replacement B and compound outcome | Stage/generation F03C rows reach real public paths. Bound cleanup is repaired; exact B and full fresh recovery proof remain 03. No post-join EEXIST retry sibling here. |
| Quarantine move/restore, P:1086–1200 | Public object may already have moved into private candidate; public C may appear | Compare candidate/bindings; restore only with no-replace; preserve A/B/C on conflict | B01 stage/generation/rollback second-public-replacement tests retain exact candidate and public C bytes/inodes; activation collision also covered. Retained invariant. |
| Recursive removal, D:434–500 | Child fd is opened before Stat or identity failure; earlier siblings may already be deleted | Immediate deferred Close owns every opened child; no GC credit; preserve mismatched child and allowed partial residue | RE direct replacement/Fstat tests and mandatory public Recover-stage/Prune-generation repeated nested tests. Immediate repair retained. 05 narrows wrong-phase/child-identity claims; 04 covers expected durable state. |
| Lease and returned generation, P:493/508/1688 | Reader obtains validated directory A and lease; cleanup takes directory lifetime lock plus checked lease | Returned FS stays bound to A; held directory prevents removal even after lease-path replacement; destructive use compares actual lease identity | Returned-reader payload/descriptor tests, held-reader and late-lease matrices retained. Required full raw control/root/history comparison remains 04. |
| Nonmutating Check, P:445/461 | Existing shared lock/read-only validation; no recovery/prune/repair | Exact expected pre-cut state remains unchanged; refusal cannot mutate authority or stale roots | Existing Check refusal and CURRENT-byte assertions useful; complete stable-history/root witness absent, 04. |

An open or mkdir whose identity cannot be established does not authorize deleting an unknown occupant. An operation can also retain intentional forensic residue. Those qualifications do not excuse losing a known successful common-record creation or a meaningful compound error.

## Actual durable caller/cut obligations

The 063 correction distinguishes eleven public/cut rows; no-JOURNAL Recover and Open are separate entries. The following table retains those rows and their object kinds. Every applicable row needs 7481-04's independent expected controls, roots and stable private history; a generic post-observed helper cannot discharge it.

| Actual caller/cut | Phase and permitted control transition | Object/proof retained; specific gap |
| --- | --- | --- |
| Explicit Prune | Current selected, no JOURNAL; no selection change needed | Distinct empty/nonempty regular lease-file B; generation root and A/B required. Existing actual refusal useful; strengthen expected state. |
| No-JOURNAL Recover | Recover selects current and prunes stale | Same regular lease-file variants; preserve selected control/history at cut. Not equivalent to Open's public return. |
| Open through recovery | Recovery occurs before returning handle | Both lease-file variants; assert refusal/no invalid handle and retained expected state. Preserve separate returned-A tests. |
| Publish initial recovery | Stale generation already unheld at entry; new staging not yet reached | Both lease-file variants and actual initial-recovery nested fixture. No claim about final prune. |
| Held-reader prune | Directory lifetime lock remains held despite replaced lease pathname | Held old root/A and lease B survive; prune can defer successfully. Decoded controls alone in PT:101 do not prove raw expected history. |
| Prepared JOURNAL, new selected, Recover | Prepared old/new journal and new CURRENT already present | Both regular lease B variants; refusal retains the actual prepared cut until allowed recovery advances it. |
| Committed JOURNAL, new selected, Recover | New CURRENT and committed journal present | Both regular lease B variants; old retained on refusal; raw identities/history must be bound before fault. |
| Successful Publish final prune | New CURRENT selected/validated and committed journal written before old cleanup | Actual DM lease rows reach this phase. New control inodes are legitimate; compare to expected new cut, not pre-Publish inode. Nested final-prune proof is optional. |
| Prepared old-selected / rejected-new fresh Recover | Recovery rejects new and selects/restores old or logical absence before rejected cleanup, as fixture dictates | Existing empty/nonempty rejected-generation lease variants retained. Compare post-restoration control authority plus rejected root at actual cut. |
| Immediate rollback | Second active validation fails; old restoration may precede rejected-generation cleanup | Existing empty/nonempty lease variants; retain validation and cleanup causes. Legitimate restored CURRENT identity differs from pre-Publish. |
| Owned stale-stage Recover | Stage ownership marker and root govern removal; no generation-lease contract | Existing early/late root replacement and mandatory nested stage proof retained. Raw controls/stable private history and exact actor B require stronger witnesses. |

**Check is an additional, non-destructive contract**, not a lease-deletion variant: retain expected controls, selected/stale roots and private history unchanged on refusal. Nested child DIRECTORY B, failed-allocation DIRECTORY B, generation lease regular-file B, generation root, and directory lifetime lock are different objects. The report does not exchange one for another.

## Old seven-entry ledger and preserved invariants

| Prior entry | Current disposition |
| --- | --- |
| F-03-A | Partial. Later completed-record cleanup is repaired; earliest creation and marker/phase siblings remain 7481-01. Independent expected graph/history proof remains 7481-04. |
| F-03-B | Partial. Many joins and real completion controls are valid; compound retry remains 7481-02 and stipulated missing compound proof is 7481-06. |
| F-03-C | Partial. Bound cleanup repairs the demonstrated old B deletion. Actual A/B/recovery and failed-allocation caller proof remain 7481-03; public compound outcome also intersects 7481-02. |
| F-04-R | Resolved. Safe observers open descriptor-relative, no-follow/nonblocking where applicable, bind observed identity before dependent reads, and bounded FIFO/symlink/directory negative controls reject B. This does not make the separate expected-state oracle complete. |
| F-05/06-P | Partial. Named lease schedules and useful witnesses exist; independently expected raw controls/roots/private history across actual callers remain 7481-04. |
| F-02-P | Partial proof closure. Immediate descriptor repair and mandatory nested Recover-stage/Prune-generation resource proof are retained. Expected durable-state comparison and unsupported extra claims remain 7481-04/05. No remaining leak is inferred. |
| F-08-R | Resolved. Child acknowledgment follows an actual nonblocking flock contention result on the bound descriptor, and signal assertions distinguish the intended signal. An earlier pre-lock acknowledgment would not suffice; the current one does. |

B01 no-replace restoration and public A/B/C preservation, B02 strict bootstrap authority, R3-01 FIFO refusal, R3-02 bounded live descriptors, R3-06 returned validated A, R3-07 independent directory lifetime lock plus lease integrity, and F01 recorded capture opens remain preserved. R3-02's history retains metadata rather than one live descriptor per transaction; its finite-descriptor control is not a disk-history bound. The six F01 CURRENT/JOURNAL actual-open witnesses retain pre-swap A/B identities and public bytes and recover after explicitly restoring owned A.

F02's direct failed-child ownership repair has actual RED growth from 7 to 31 descriptors over 24 iterations and later bounded proof; it is not reopened merely because the nested final-Publish label is wrong. F07's accepted original signal chronology remains RED 12.383s, routing edit, GREEN 15.417s. The failing uncommitted source has no recovered full tree SHA; commit 729 is not relabeled as that RED tree. Its original RED/edit/GREEN hashes and the canonical provenance correction remain authoritative. No additional provenance replay is needed.

R3-04 lint remains an acceptance-gate obligation with historical diagnostic attribution. R3-05's safe-observation repair remains distinct from the newly confirmed preservation-oracle weakness.

## Ten-lens completion

| Lens | Substantive audit result |
| --- | --- |
| 1. Architecture and data flow | Traced main signal routing → lock-render → admitted closed stage → retained roots/locks → stage → authority/JOURNAL/CURRENT → validation/rollback → prune/output; Open and Check follow their distinct paths. All original changed production files and repair-created ownership edges are included. 01/02 are discontinuities in those paths. |
| 2. Happy, bad and edge behavior | Examined absent/present controls, valid/invalid bootstrap, partial versus complete records, exclusive collisions, symlink/FIFO/type changes, A/B/C substitution, validation failure, cancellation and real-operation-plus-injected completion. 01/02 survive; 03/06 separate missing proof from new defects. |
| 3. State machine and concurrency | Distinguished initial recovery, prepared/new-selected, committed/new-selected, final prune, old-selected/rejected-new recovery and immediate rollback. Verified capture identity, no-replace restore, held generation directory locks and actual contention acknowledgment. 04/05 identify proof/phase overclaims. |
| 4. Security and secret taint | Source names and local filesystem objects are the changed untrusted inputs. Descriptor-relative access, strict direct names, no-follow/type checks and fixed private record names remain. Unknown/replaced objects never become deletion authority. No new credential, provider, shell/HTTP/SQL tool or secret sink was introduced. |
| 5. Retry, resume and idempotency | Traced every changed joined-error consumer through retry, absence, conflict, recovery and success. 02 is a reachable false-success retry; 01 prevents bounded fresh recovery after self-created failed record preparation. No provider retry/rate-limit behavior changed. |
| 6. Output integrity | Closed physical file set/digest checks, actual returned-A payload and selection-before-output order remain. CLI writable-completion tests require an error without success output. 01/02 and incomplete expected-state evidence prevent acceptance; no transport envelope or provider-record completeness claim was added. |
| 7. Declaration reachability and closed surface | No source.lock, connector definition corpus, seven-lane declaration or runtime selector changes. Atlas publication owners/proofs and constraint/schema changes were inspected with canon/Make/script consumers. Compact identifier boundary matching was reviewed separately from its package's inherited fixture failure. |
| 8. CLI and App parity | Changed authoring lock-render/check/signal behavior and related guide/canon/website/generated docs were examined. Embedded runtime defs, PM connector hand parser, App save/load, approval/preview execution and provider routing remain outside this change. No new App/runtime publication route is claimed. |
| 9. Provider semantics | Inapplicable with evidence: changed code consumes admitted local bytes and local filesystem state; no HTTP, GraphQL, transport content type, redirects, paging, rate header or provider idempotency implementation changed. No live-provider research or execution was needed. |
| 10. Tests and evidence | Challenged assertion bodies and actual fault schedule, falsified the central helper using supplied evidence, independently proved record siblings, reconciled the full old seven and fresh five, added the omitted 058 compound proof, verified immutable bindings and separated current/historical/cached/RED/oracle results. Audit complete; acceptance incomplete. |

The original range also changes go.mod/go.sum: x/sys v0.47.0 becomes direct at the same version and x/text moves from v0.39.0 to v0.38.0. These are part of the reviewed dependency identity and must not be silently excluded from future source binding. They do not introduce a second runtime foundation or explain the demonstrated local record/collision defects. No additional actionable dependency defect was established.

## Original evidence and probe custody

The fresh review was read in full and remains unchanged:

| Artifact | SHA-256 |
| --- | --- |
| cp11-astra-7481d177-fresh-full-range-review.md, 380 lines / 52,454 bytes | 2d92ce239d19509aa1838c23d5d0b9f31f4e5784232a512ee2b0f93e70ca571c |
| cp11-e77-complete-ledger-audit.md | 6d0d5d80bd6355d2f578649f47dd36af5d9b441007a432b8a805dcac358010f2 |
| cp11-e77-audit-original-return.md, 429 lines / 40,699 bytes | 9d68c63c7d6b525dab6010997a9eb36dc43890ca9cfd657523bbee25e47e24ba |
| Supplied review-probes_test.go | 76cd08d998041871dcda4b576c6f09776b4848f7559a77caa87549e67dc3f7ef |
| Supplied review-probes-overlay.json | bf669b5c23665f21aa6d04e7341a2466e640225afabf23d9db07a99500a6a8e3 |
| Supplied review-probes-receipt.md | 70da86171316ddce5f93ae168b7dc186275cdfc4aad6e9ea8036e0600107e2ea |
| New cp11-7481-audit-record-siblings_test.go | 69b4966c393ab2e3aec6ebe05ce748e04fe66ab5805eecbce5ee8ed92bdb569b |
| New cp11-7481-audit-record-siblings-overlay.json | 5deaf8b224dadabac615989a927bb39f0199d300fa8cc801b1bd70b58f525b7c |
| New cp11-7481-audit-record-siblings-original-receipt.json | c749d83862aefc24e25618170fc2117d7864e4d442974429c1249f557a3de148 |
| cp11-current-three-group-package-receipts.json | 738193c2a30663995bb0e8904b131c244e8909f70f7a52b39b1840384ed27d6a |

All named private artifacts are under /Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover. The supplied probe filenames have the full prefix cp11-astra-7481d177-. The prior original-return provenance identifies message amsg_01a07495-52ae-7a00-a473-530109963ffb, physical record 9330, timestamp 2026-09-06T02:38:58.222Z and raw-record hash 34560dded37add4818eaac6415e15b595421ddde78071720ac9212f8fa24517f. The saved versus returned old audit has the preserved formatting/newline distinction; neither is a fresh acceptance. Complete prior current audit commit 3256d930 precedes repairs; d22f2ce4 is provenance-only.

The supplied three-probe receipt is explicitly a **transcription**, not a raw rollout export. Its actual invocation was:

    /usr/bin/time -p go test -count=1 -timeout 20m -overlay /Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-astra-7481d177-review-probes-overlay.json ./cmd/connectorgen -run '^TestCP11Astra7481' -v

It completed with exit 1, package 2.560s, wall 5.59s, user 5.18s, sys 4.91s, session 62804. Two desired production regressions failed and one oracle-rejection assertion failed. The temporary probe retains actual A/B witnesses; the oracle counterexample records expected root A inode 0x2657d7a8 and unrelated B inode 0x2657d822 on device 0x1000010. The overlay added only the previously nonexistent cmd/connectorgen/cp11_astra_7481d177_review_probes_test.go. No existing production or test file was replaced. I did not rerun it for a more attractive receipt.

My additional invocation was:

    /usr/bin/time -p go test -count=1 -timeout 20m -overlay /Users/karthiksivadas/pm-cli-agent-workspace/data/cli-batch1-pi-takeover/cp11-7481-audit-record-siblings-overlay.json ./cmd/connectorgen -run '^TestCP11CompleteAuditRecordSiblings$' -v

Only the checked-nonexistent cmd/connectorgen/cp11_7481_audit_record_siblings_test.go was added virtually. The preserved JSON contains the original tool command/result capture, including initial session 2826 and its final exit 1 result; it is not a rollout export. Package time was 1.079s; wall 5.26s; user 5.16s; sys 5.27s. The four test failures were:

| Case | Proven retained record | Public/Check/Recover observation |
| --- | --- | --- |
| marker / parent Close after create | Device 0x1000010, inode 0x26580024, regular, 0 bytes | Publish returns injected parent completion; Check and Recover decode marker EOF |
| marker / partial Write | Same device, inode 0x2658002f, regular, 1 byte | Publish returns actual partial-write completion error; Check and Recover decode marker EOF |
| phase / parent Close after create | Same device, inode 0x26580036, regular, 0 bytes | Publish returns injected parent completion; Check and Recover decode phase EOF |
| phase / partial Write | Same device, inode 0x2658003d, regular, 1 byte | Publish returns actual partial-write completion error; Check and Recover decode phase EOF |

For the parent-Close cases, the hook verifies the actual new record through the still-open parent before real Close and injected completion; pre-creation reads cannot trigger it. For partial Write, the hook obtains the writable descriptor identity and writes the first byte before returning the injected error. Post-return observation finds that same created object and exact bytes. The fixture checks absent CURRENT/JOURNAL and nonmutating Check before the desired fresh-Recover assertion fails. No arbitrary malformed external object or production-source overlay was substituted.

These observations are not GREEN, naturally occurring filesystem failures, process-crash tests or power-loss certification.

Current source hashes independently rechecked:

| Path abbreviation | SHA-256 |
| --- | --- |
| P | 9bf803f98aa4186ef072ac0e8e7451bd1778b690da0f8dbea60d00027cb08d0b |
| D | 1a90f8b8c82f9d498a1298778d0be495f29ef850ba3af7cbbd9657bb6ae9da81 |
| R | 3e41f07670261fede3f44a21bc9bc725cf447bb8f62942d8af08d15fb159fb32 |
| DM | 29fae715a98c1071e41ef6e297e59b49a12571452c1cbf4160c191c36cb07531 |
| RE | 83c4733639776b0eaff90f2f0df31963b94eb3b12f4832564b3355b07e6fb65d |
| G2 | 56b6a14286cc430b201c725fcda3de6ceab8b7e9e0bbd82fc09f6ea2ad0504bf |

## Gates, terminal states and limits

| Evidence/command | Terminal state and exact meaning |
| --- | --- |
| Owner final normal: go test -count=1 -timeout 20m ./cmd/connectorgen | Exit 0; package 319.928s; wall 322.297918541s. Original physical record 13267, raw hash f7f751f375e4c08a16006d97b428972c568a2d76d9d4ce01924283e7c6c90795. Intended final precommit source, bound to 7481. |
| Owner final race: go test -race -count=1 -timeout 20m ./cmd/connectorgen | Exit 0; package 783.834s; wall 786.420858334s. Physical 13428, raw hash 5c3261ad57cef5fd94f4359759bc608a37f597182a85e5e1c90eb22b45996c67. Same intended final source. |
| Owner current focused Group3 | Normal 32.789s and race 45.628s pass; these are the later intended final-source controls. |
| Earlier all-group and focused receipts | 315.166/763.610 and 30.270/42.177 are different earlier source states. Group3 earlier fixture FAIL 5.011/4.935, corrected PASS 17.146 and intermediate controls are retained as their actual chronology, not production RED or current-source equivalence. |
| Current eight supplementary gates | Successful current results retained: source-lock check 206.821s; definitions 553/0; Atlas 1.326s; canon/docs/release and boundary checks; runtime preflight explicitly cached. Literal commands/receipts remain in the 064 evidence and command index. These pass their actual gates, not the semantic audit. |
| Current boundary scanner, recovered 068 receipt | Original exit 0, wall 254.579597042s; 284 files, 553 connectors, six exceptions, zero findings/warnings. Physical 13951, timestamp 2026-09-06T06:11:27.554Z, raw hash ec88ff8c01c57e8647206c78ff003698932b66be9420a737536073ddd75c07b7. No rerun needed. |
| Boundary PACKAGE and original-source reproduction | Package failure at invalid_cli_surface, 249.388s; original-base package source under current dependencies reproduces in 0.153s. CP29 debt, not a current package PASS and not a whole-base checkout run. |
| Current original-range lint | **Not established.** Group3 precommit new-only 0 is not a base-to-7481 lint receipt. Required current acceptance gate remains open. |
| Independent original-range git diff --check | Exit 2, diagnostics in the preserved PREMATURE-F01-EDIT-SNAPSHOT.patch archive. Excluding only that exact artifact gives 0. Archive remains unchanged. |
| Independent candidate/worktree identity and ancestor checks | Completed successfully; candidate tree/base/HEAD and published ancestor verified; code/test/dependency paths match 7481. |
| Supplied probe command | Completed exit 1, 2.560s / 5.59s; historical supplied regression/oracle failures, not my rerun. |
| Additional sibling probe command | Completed exit 1, 1.079s / 5.26s; four desired production regression failures against unchanged 7481. |
| Auditor activity at finish | All inspection and artifact-write commands completed. No active command session, child process or delegated agent remains. No project edit, full/race rerun, baseline-debt rerun or publication occurred. |

The original 064 report hash quoted in the fresh review predates the 068 appendix. Current private 064 SHA-256 is a59da5a5752e5c2149ca474cf5b9ac5761fa494fe7ae470ee04084a5680ba779. That is factual chronology; the original review must remain unchanged. The 068 receipt file SHA-256 is 744d0d129e15c8eccbaf723dda0ca96487c8babc25ce3c2d53d6feea25ed5849.

The required lint acceptance gate is a current original-base-range introduced-diagnostic check with exact source attribution over the changed code surface. It must not silently use HEAD-relative proof-only changes. Preserve the established 20 global / 15 package old diagnostics separately from the 31 historical introduced group; no current surviving count is inferred here. These are not summed into a fabricated new total. Current findings need their coordinated response before final acceptance; this audit did not rerun lint or unrelated known debt.

No Linux runtime, power loss, live provider, credential, customer database, shared receiver, new service/runtime foundation, whole-repository suite, release, main merge, full-programme or no-mistakes certification is supplied. Source-lock authority, closed runtime, Atlas authoring-only status and CLI/App/provider boundaries remain intact.

## Bounded coordinated closure

Firstmate should freeze and authorize one response to the complete six-entry set, with the tables above retained. This report does not itself authorize an implementation.

1. **Mandatory production repair:** close the common-record ownership frontier and prepared/marker/phase siblings in 7481-01, and the compound allocation-to-retry edge in 7481-02. Preserve strict unknown-state refusal, all no-replace A/B/C invariants, registered anchors, valid bootstrap classes, actual selection cuts, one Close owner and typed causes.
2. **Mandatory existing proof:** make expected-state helpers independent and falsifiable; complete 7481-03's explicit failed-allocation public/direct/stage/generation rows and exact DIRECTORY A/B witnesses; complete 7481-04's controls/root/history comparison across all caller/cut rows; complete 7481-06's stipulated compound classes. Bind repaired 01/02 behavior with actual desired RED/GREEN and fresh recovery at the proper cut. Later proof additions are not retroactive pre-edit REDs.
3. **Unsupported extra claims:** correct 7481-05's nested final-Publish label and derivatives. Preserve mandatory nested Recover-stage/Prune-generation resource proof and actual final-prune lease proof. Narrow post-Stat retained-child identity wording to its observed bytes/root, or add that exact witness. No extra runtime guarantee is required.
4. **Acceptance evidence:** update canonical dispositions and Atlas proof references to the assertions actually present; run proportional focused normal/race plus the required final candidate gates, including current original-range lint attribution. Freeze the resulting source/test/dependency SHA and obtain fresh complete-original-range review of the coordinated changes and their new edges. Preserve current receipts, old seven, historical F07, CP29 debt and all custody distinctions.

There is no operational blocker or unresolved product/architecture question preventing a coordinated response. Remaining uncertainty is concrete: the specified proof cells and current original-range lint have no sufficient acceptance evidence, and no future repair has been implemented or reviewed by this auditor. The withdrawn 90% estimate is not reinstated.
<!-- END VERBATIM CP11-7481 COMPLETE AUDIT bc109e85 -->

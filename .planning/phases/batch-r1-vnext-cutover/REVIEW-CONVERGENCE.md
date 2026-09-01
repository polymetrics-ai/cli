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

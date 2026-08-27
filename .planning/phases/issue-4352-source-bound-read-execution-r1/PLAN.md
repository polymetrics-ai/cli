# Source-bound read execution foundation

## F1 source/local request-input closure — 2026-08-27

### Task delivery header

- **Issue:** Refs #4352; repair the existing #4356 branch after independent
  current-main audit finding F1 at `5d6cc316c17f63368f072171dd4d2fe80a7f75fb`.
- **Base / delivery:** `main` at `1324c52bab0b224ed8958858af7676b8b8e191b4`;
  normal fast-forward commits only to
  `fm/cli-source-bound-read-execution-r1-continuation`. PR #4356 remains open;
  no force push or merge is authorized.
- **Problem:** A retained source-bound GET can retain a declaration-owned
  `path`/`query` parameter absent from the exact locked source contract. The
  audited Asana `get_access_requests` `query.rogue` probe passes both
  `source-import --check` and `validate`.
- **Scope:** Close source-to-local *and* local-to-source request inputs for
  admitted source-bound reads. Optional source filters may stay omitted, but a
  local operation parameter or CLI `maps_to` input must be source-declared.
  Provider paging remains a separate closed pagination contract; headers/body
  are admitted only when the retained source contract admits that class.
- **Out of scope:** No source locks/captures, provider bytes, generic HTTP
  executor, credentials, live I/O, other worktrees, or #4351-owned generic
  admission rules.

### Inline GSD / TDD plan

The compatible isolated Pi worker runtime is unavailable and the canonical
delivery contract forbids role spawning, so `discuss-phase`,
`plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review` are
resolved and executed inline for this bounded repair. Required skills used:
`golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`,
`golang-security`, `golang-safety`, `golang-design-patterns`,
`golang-structs-interfaces`, and `golang-lint`.

| Slice | Red evidence | Green acceptance |
| --- | --- | --- |
| Operation-only extra | Add `query.rogue` to retained Asana `get_access_requests`; prove source-import check and validate currently accept it. | Projection removes it from the derived parameter block; source-import check detects that drift and validate reports an exact source-projection finding. |
| Operation + CLI extra | Add the same operation input and `--rogue -> query.rogue`; prove both commands currently accept the altered bundle. | Check and validation reject the operation and CLI declaration pair; valid retained optional filters remain untouched. |
| Input classes | Use source-bound fixtures with path/query plus unwanted header/body mappings. | Closure permits only names/classes retained by the descriptor, preserves config-bound path fields and closed pagination placement, and cannot make a header/body executor claim. |
| Regression / delivery | Run affected generator/runtime/CLI packages, generated source and documentation gates, then fresh credential-free Asana census. | Publish one normal reviewed repair and record its exact SHA for a separate audit. |

## Repair r4 task delivery header

- Issue: Refs #4352 — source-bound read execution foundation; repair existing PR #4356.
- Base branch: `main` at `b33983927d863032dac8220949990506e812937d`.
- Merges into: `main`.
- Delivery: PR #4356 remains open; all six frozen findings are red/green repaired and fast-forward pushed to `fm/cli-source-bound-read-execution-r1-continuation`, then a fresh independent audit reviews the final exact code SHA. No merge is authorized.
- Working branch: `fm/cli-source-bound-read-execution-r1-continuation`, immutable start `19b2bd8dc470d6fa92da1a500173c8c8c30ba59c`.
- Task: Repair closed paging, pre-credential origin refusal, direct-API ETL binding, 21 existing-lane Asana mutations, fixed-100 isolation, and generated Asana docs/help/website parity; retain provider bytes and only genuine named foundations.
- Verification: Every `AUDIT-001`–`AUDIT-006` red/green regression, targeted packages, full `cmd/connectorgen`, generator/admission/runtime-preflight/canon/doc gates, credential-free binary census, and independent exact-SHA re-audit.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Closed paging surface | live | Generated source-bound Asana reads have no raw `offset` or command `limit`; derived `--page`/`--page-cursor` navigation remains. |
| Pre-credential source origin refusal | live | Credential-store, auth-cohort, and requester observers are all zero for an invalid origin overlay. |
| Direct ETL binding | live | Direct `Read` rejects altered source ID/method/path/records/pagination/origin before auth or I/O. |
| Existing Asana write lane | live | All 19 DELETE and two no-body POST actions reach `missing --credential`; actual remaining gaps stay blocked. |
| Fixed-100 self-containment | live | The isolated test workspace carries every referenced connector source and passes without repository-tree input. |
| Truthful generated docs | live | Source-derived actual counts/pagination/lane claims match manual/help and website data; historical blockers are absent. |

Repair skills loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-documentation`, `vercel-react-best-practices`, and `vercel-composition-patterns`. Required `frontend-design` and `web-design-guidelines` are not installed; only generated website data is in scope and no React UI/component changes are planned.

## Task Delivery Header

- Issue: Closes #4352 — add source-bound read execution foundation.
- Base branch: main.
- Merges into: main.
- Delivery: Pull request open against `main` with the shared foundation committed, pushed, focused verification green, and the GitHub API-reported base verified as `main`.
- Working branch: `fm/cli-source-bound-read-execution-r1`.
- Task: Add the smallest declarative generator/runtime extension that makes a source-locked, non-mutating provider operation callable through an exact, closed operation binding. A command must reach the normal credential boundary without provider I/O, live certification, a source hash requirement, or a connector-local shim. It must distinguish bounded direct reads from genuinely stream-capable reads and refuse missing shared foundation before I/O.
- Verification: Focused red/green generator and command-runner tests; source/generator checks; built `pm` preflight for a representative read in an isolated, credential-free project; targeted regressions for direct writes, reverse ETL, binary, and delete behaviour; `go vet ./...`, `go build ./cmd/pm`, relevant individual `make verify` gates, and `git diff --check`.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Source projection considers non-mutating source operations | live | A generated-source fixture preserves a GET operation in the source-bound declaration path while an existing mutation control retains its write behavior. |
| Generated reads remain source-bound and closed | live | Tests assert the command's locked operation identity and exact method/path; a substituted route is refused before a requester can observe I/O. |
| Read semantics are honest | live | Tests classify a source-backed singleton/direct response separately from a record/pagination-backed collection and retain a named missing foundation where the source contract cannot establish a stream. |
| Missing inputs or foundation fail before provider I/O | live | Requester spies observe zero calls while the stable actionable preflight error names the missing source-bound contract/foundation. |
| Valid generated read reaches credentials | live | A credential-free built-binary invocation of the selected command stops at `missing --credential`, rather than unknown command, unsupported preflight, or provider I/O. |
| Existing safety behaviour is unchanged | live | Focused existing direct-write, reverse-ETL, binary, and delete controls continue to pass. |

## GSD discussion record

The launch brief resolves the relevant design choices: this is a shared foundation issue rather than Batch-1 connector work; its command surface is operation-identity-bound and has no arbitrary request escape hatch; direct reads are one bounded response; ETL promotion is allowed only with source-backed pagination and record semantics; materialization and runtime certification remain independent. The canonical isolated-worker runtime is unavailable and the repository delivery contract forbids role spawning, so the resolved GSD prompts are executed inline/manual.

## Required skills and CLI parity

Loaded: `golang-how-to`, `golang-cli`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, and `golang-structs-interfaces`.

This changes generated connector command execution, not a hand-authored top-level `pm` help tree. Before handoff, record the applicable generated command/help behavior and inspect `docs/cli/**`, `website/**`, and generated manual/help artifacts; explicitly mark any unchanged surface as not applicable with the reason.

## TDD slices

1. **Projection and declaration admission.** Red: add a source-projection fixture with a non-mutating operation and assert the current generator leaves it declaration-pending or otherwise lacks a closed execution binding. Green: generate only source-locked GET/read bindings with exact identity/method/path and typed inputs; leave non-supported contracts in a named declaration/foundation state. Preserve mutation/write controls.
2. **Runtime preflight and dispatch.** Red: prove a generated read has no credential-bound execution path and that a route substitution/missing contract can reach no requester. Green: resolve the command only through the embedded source-bound operation ledger, validate identity/method/path and typed inputs, then use the existing credential/auth preflight. Do not add arbitrary method/path/header/body input.
3. **Honest read class.** Red: exercise a singleton/direct-response operation, a genuinely source-backed paginated/record collection, and a path-parameter one-object operation. Green: dispatch a bounded direct read or existing stream only where its source contract proves semantics; otherwise return stable `missing_foundation` before I/O.
4. **Real locked controls.** Red: identify the Asana `getAccessRequests` source row plus available collection/pagination and path-parameter source rows, then add tests that fail on the current foundation. Green: prove the materialized commands reach the credential boundary without a provider request.
5. **Regression and verification.** Run focused cross-package tests plus existing direct-write, reverse-ETL, binary, and delete controls. Review the diff for input-boundary escape hatches, error leakage, and behavior regression.

## Audit repair gap plan (F1/F2)

The independent audit of `bfd57d3b8ddc1623b8d514b20d5822924f89b060` found two gaps. This plan is executed inline/manual because the canonical contract forbids spawning roles and no compatible isolated GSD runtime is available.

1. **F1 — generated help artifacts.** Red: reproduce the exact Verify failure at the audited head and identify each stale tracked artifact. Green: regenerate and review only the tracked Asana skill and root help golden artifacts required by the generator, then assert the repository check is clean.
2. **F2 — hermetic source evidence.** Red: add an assertion driven by a retained, pinned Asana provider artifact/lock and demonstrate that the current fixture-only descriptor construction cannot reject divergent source identity, method/path, typed inputs, or workspace pagination. Green: retain the artifact and exact lock, make the generator-to-bundle assertion consume it, and reject every divergence before any provider I/O. The read-only projection may seal an identity only for a pre-existing `implemented` direct read or stream; it must leave planned commands and every write/delete artifact byte-equivalent.
3. **Closure.** Re-run the exact Verify checks that failed at the audited head plus focused generator/runtime tests. Do not copy interfaces or changes from #4350/#4351; if a missing landed interface is required, record it as a bounded dependency.

## Captain-authorized mutation mapping repair

The full retained-source validation exposed source-cited mutation coverage that
cannot be repaired by read execution or by changing any provider behavior. Under
the narrow `012.msg` authorization, add a connector-owned, operation-granular
partial-coverage disposition for an already implemented but source-incomplete
action. It must carry the exact locked source identity/method/path and one
recognized missing-foundation category; it must reject an absent or fully
source-covered action. Keep the existing source-cited non-executable
disposition for genuinely absent actions, preserve every existing executable
command, and limit surface synchronization to the mutation mapping lane so it
does not rewrite planned Asana reads. Record red/green evidence, run
source-import/validate/surface-sync, then serialize the broad generator test
after the source-lock audit has exited.

## Expected change boundaries

- `cmd/connectorgen/sourceprojection.go` and its tests
- `internal/connectors/{engine,commandrunner}/` only where a source-bound operation/read contract requires it, with focused tests
- Generated, source-backed Asana artifacts only where the locked provider contract concretely supports a read binding; remaining Batch-1 rows stay explicitly partial, planned, or foundation-bound rather than being claimed wholesale
- This phase's `PLAN.md`, `TDD-LEDGER.md`, `VERIFICATION.md`, `RUN-STATE.md`, `SUMMARY.md`, and `REVIEW.md`

No certification policy, live provider calls, generic request interface, or edits to another worker's branch belong in this PR. The retained Asana artifact is solely a hermetic shared-foundation control, not a claim of certification or a substitute for Batch-1 materialization review.

## R6 merge-blocker repair — 2026-08-27

## Task Delivery Header

- Issue: Refs #4352 — source-bound read execution foundation, R6 repair.
- Base branch: `main` at `b9b2478b3b2451d632d28b9aa138a170ad835110`.
- Merges into: `fm/cli-source-bound-read-execution-r1-continuation` → `main`.
- Delivery: normal non-force advance of existing PR #4356, exact remote head/base read-back, then a fresh independent audit; no PR merge.
- Working branch: `fm/cli-source-bound-read-execution-r1-continuation`.
- Task: preserve every Asana row while returning its structured declared unavailable disposition at runtime, and move source-bound ETL structural preflight before adapter auth admission.
- Verification: red/green commandrunner and adapter tests; built binary unavailable examples; changed packages; generator/surface/runtime/boundary gates; exact credential-free census.

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every unavailable Asana row reports its declared reason | live | Schema, encoding, OpenAPI-sibling, and batch command calls yield a typed blocked error whose `Reason` matches declaration-owned structured fields. |
| Structural ETL drift precedes auth | live | Adapter `Read` and `ReadWithOutcome` each return structural drift with zero auth-admission and requester calls. |
| No surface or executor is fabricated | live | Generator/surface/preflight checks and 249-command census preserve 212 implemented / 37 unavailable. |

**Red:** R6 built-binary census observed 0/37 unavailable rows with their true
reason, and an adapter counter reached authentication once for a source-bound
stream path mismatch. **Green:** use structured command fields (never
untrusted notes) and preflight before `executeWithAuthCohort`, retaining the
inner preflight as defense in depth. The canonical GSD prompts were executed
inline because compatible isolated workers are unavailable and the contract
forbids role spawning. Skills: golang-how-to, golang-cli, golang-testing,
golang-error-handling, golang-security, golang-safety, golang-design-patterns,
golang-structs-interfaces, golang-context, golang-documentation, golang-lint.

## R2 independent-audit gap closure

> Superseded in part by captain instruction `021.msg`: historical planned-read
> counts are not a delivery goal. See `RECONCILIATION.md`; complete
> source-backed reads must materialize in their correct executor lane.

The independent R2 report path named by the dispatch was unavailable in both the
preserved worktree and the declared repository path. Its durable numbered audit
directive is the controlling evidence for this gap slice. It is executed
inline/manual: the canonical contract forbids role spawning and no compatible
isolated runtime is available.

1. **Classify by current capability, not merge-base status.** Captain `021.msg`
   supersedes the historical planned partition. Red: a complete locked GET is
   left planned solely because of its old declaration status. Green: every
   source-complete JSON GET with a bounded declared REST operation materializes
   as a source-bound direct read; a list becomes ETL only when its exact stream
   proves records/schema/pagination; and a genuine gap remains named before
   provider I/O. No source hash, capture byte count, certification result, or
   runtime proof participates in that admission decision.
2. **Separate admission from capture integrity.** Red: mutate capture SHA,
   byte count, capture metadata, or adapter metadata without changing the
   provider operation contract and demonstrate that mapping/declaration status
   changes. Green: source identity plus source URL/location, method/path, and
   typed provider contract decide source-bound declaration status, while the
   retained-artifact integrity verifier continues to reject changed raw bytes.
3. **Close the source-bound authenticated origin escape.** Red: configure an
   arbitrary `base_url` for a source-bound operation and observe that it reaches
   credential/auth/request construction. Green: reject the override before
   credential/auth/request construction; preserve only an explicitly bounded
   source/test origin mechanism. Tests use a requester/auth spy and assert zero
   I/O.
4. **Make operation-evidence coherent.** Red: reproduce the Verify report that
   the Asana source lock declares no operations or provider-evidenced absence.
   Green: use coherent retained source data and source-operation dispositions so
   the existing check observes the locked operation inventory without weakening
   its policy.
5. **Closure.** Run focused red/green generator and runtime tests, the retained
   source import/projection checks, `cmd/connectorgen` package tests, and the
   relevant generated-artifact checks in a clean clone. The unrelated
   blog-annotation website E2E failure is recorded as external CI state and is
   not changed without a demonstrated connection. The final capability
   reconciliation is recorded in `RECONCILIATION.md` and `TDD-LEDGER.md`.

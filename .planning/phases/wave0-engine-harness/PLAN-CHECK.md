# PLAN-CHECK — wave0-engine-harness

Goal-backward verification of PLAN.md (+ SPEC/TEST-PLAN/API-CONTRACT/DATA-MODEL/THREAT-MODEL/
RUNBOOK/EVAL-PLAN/DECISIONS) against the wave0 ROADMAP.md acceptance criteria and ground-truth
code. Run at 2026-07-02T07:46Z. Branch: connector-architecture-v2.

## Verdict: APPROVED-WITH-FLAGS

The plan is executable and will very likely achieve the wave0 acceptance criteria as literally
stated in ROADMAP.md. Ground-truth line/behavior citations in SPEC/PLAN/API-CONTRACT were checked
against the actual files and are accurate. No blocking structural defects (no circular deps, no
missing requirement coverage, no forward references). However there are several FLAG-level issues
— two real gaps between THREAT-MODEL claims and planned tests, one stale cross-reference, one
un-executed coordinator decision, and two tasks that are oversized for a single agent run — that
should be fixed before/while dispatching, not blocking the start of Wave A.

---

## 1. Task completeness vs ROADMAP.md acceptance criteria

**PASS.** All 6 acceptance bullets map to ≥1 task with a runnable verify command:

| Acceptance bullet (ROADMAP.md) | Covering task(s) | Verify command |
|---|---|---|
| Engine unit tests green (interpolation, auth, pagination, read/write, error mapping) | T-01..T-09 | `go test ./internal/connectors/engine -run Test{Schema,Interpolate,Bundle,Error,Auth,Paginat,Hooks,Read,Write} -v` |
| 3 goldens migrated w/ parity (stripe, searxng, postgres) | T-15, T-16, T-17 | `go test ./internal/connectors/engine -run 'TestParityStripe\|TestParitySearxng' -v`, `go test ./internal/connectors/native/postgres -v` |
| `connectorgen validate` rejects seeded-invalid, accepts goldens | T-11, reused by T-15/16/17 | `go test ./cmd/connectorgen -v && go run ./cmd/connectorgen validate internal/connectors/defs` |
| Conformance v2 passes for 3 goldens | T-13, reused by T-15 (`TestConformance/stripe`) | `go test ./internal/connectors/conformance -v` |
| Certify source stages pass vs `sample` | T-14 | `go test ./internal/connectors/certify -run TestSourceStages -v` |
| `go build/test/lint` green | V-21 (Wave H gate) | full command list in PLAN.md Wave H |

Cross-checked against PROJECT-level docs (`docs/plans/universal-programming-loop-prd.md`,
`docs/migration/orchestration-plan.md` wave0 row) via SPEC.md §0 references — no PRD-level
requirement relevant to wave0 is silently dropped; migration-program artifacts (inventory.json,
conventions.md, result/review schemas) are covered by T-19/D-20.

## 2. TDD pairing

**PASS.** Every behavior task B-01..B-19 has a paired T-N test task ordered before it in the same
task block, explicitly marked "(RED first)"/"(RED)". D-20 is correctly tagged `docs-only` (no pair
needed). V-21 is the verification-only gate task, correctly unpaired. `PLAN.md` Rules section
mandates RED evidence recorded in `TDD-LEDGER.md` before each B-task's implementation commit, and
EVAL-PLAN §6 makes this a hard exit gate ("every B-task... has RED evidence recorded"). T-18's RED
artifact is unusual (a recorded failing `make lint` run rather than a unit test) but this is
explicitly justified as a Makefile-target smoke assertion, consistent with `workflow.tdd_mode`'s
validation-artifact allowance — not a defect.

## 3. Dependency ordering across the 8 dispatch waves

**PASS, with one inconsistency (FLAG).**

- No task depends on a later wave's output. Verified against ground truth: Wave B (auth/pagination/
  hooks) only needs Wave A's interpolator/schema/bundle/errors; Wave C (read/write) needs Wave B's
  auth+pagination+hooks; Wave D (connector assembly/connectorgen/certify-core) needs Wave C's
  read/write; Wave E (conformance/certify-stages) needs Wave D; Wave F (goldens) needs Wave
  E-complete tooling (connectorgen validate, engine.New, conformance harness); Wave G (lint/
  inventory/docs) needs Wave D and finalizes after F; Wave H is a serial closer. This is
  conservative (Wave F strictly only needs Wave D's engine.New/connectorgen; waiting on E adds a
  wave of latency but does not create a defect) — noted as an efficiency opportunity, not a bug.
- Parallel tasks within each wave touch disjoint files (verified by reading each task's `Files:`
  list — Wave A's 4 tasks each own a distinct `.go`+`_test.go` pair; Wave B/C/D/E/F/G tasks likewise
  do not share a primary file, aside from the shared `internal/connectors/defs/` bundle format
  used independently by T-15/T-16/T-17 in different subdirectories).
- **FLAG — DECISIONS.md not fully reflected in PLAN.md**: `DECISIONS.md` #2 explicitly states "B-12
  (certify report/cliharness): **APPROVED to float to Wave B** — zero engine deps, shortens the
  critical path." `PLAN.md` still lists T-12/B-12 under the `## Wave D` header and in the Wave-D row
  of the dependency graph (`Wave D: T/B-10  T/B-11  T/B-12`), only hedging in prose ("may be
  dispatched as early as Wave B if capacity allows" / "B-12 has no engine dep — may float earlier").
  This is not a correctness bug (B-12 has no engine dependency either way, so running it in Wave D
  is safe) but it means the coordinator's explicit approval to shorten the critical path is not
  actually encoded as executable wave assignment — a coordinator/executor reading only the wave
  labels and dependency graph will dispatch B-12 in Wave D, not Wave B, missing the intended
  schedule optimization. **Fix**: move T-12/B-12's `wave:D` tag to `wave:B` and move it out of the
  Wave D graph row into the Wave B row (`Wave B: T/B-05  T/B-06  T/B-07  T/B-12`), or explicitly
  document in PLAN.md why the coordinator's decision was not applied.

## 4. Coexistence mechanism safety

**PASS.** Verified directly against `cmd/registrygen/main.go`:
- The `skip` map is at line 30 exactly as SPEC.md/PLAN.md claim (`cmd/registrygen/main.go:30`),
  currently `{connsdk, httpsource, registryset, _quarantine}`. B-16 adds `defs, engine, hooks,
  native, conformance, certify`.
- `connectorPackages()` (ground truth, `cmd/registrygen/main.go`) scans **only the immediate
  children** of `internal/connectors/` via `os.ReadDir` — it is non-recursive. This means:
  - `internal/connectors/native/postgres/` is safe from accidental registration because `native`
    itself (the top-level skip-map entry) contains no `.go` file with a package clause directly in
    it (only in the `postgres` subdirectory) — the skip entry is technically redundant
    defense-in-depth here, not strictly load-bearing, but harmless and good practice against a
    future stray file.
  - `internal/connectors/defs/defs.go` (package `defs`, top-level file) WOULD be picked up as a
    connector package without the `defs` skip entry — the skip entry is load-bearing here.
  - `internal/connectors/hooks/hookset/hookset_gen.go` and
    `internal/connectors/conformance/conformance.go` and `internal/connectors/certify/certify.go`
    are analogous: `hooks`/`conformance`/`certify` skip entries are load-bearing for `hooks`, not
    for the others' own top-level files respectively (conformance.go and certify.go declare
    packages directly under their own top-level dirs, so those skip entries are load-bearing too).
  - `engine` skip entry is load-bearing (`bundle.go` etc. declare `package engine` directly under
    `internal/connectors/engine/`).
  All six additions are therefore each justified (four load-bearing, two defense-in-depth), and the
  claim "the skip dirs contain no connector packages yet at scan level" (T-16 comment) is verified
  true given the non-recursive scan.
- `RegisterFactory` overwrite semantics (verified: `connectors.go:60`, "re-registering an existing
  name overwrites its factory while preserving its original position") mean a stray `RegisterFactory
  ("stripe", ...)` call anywhere would silently flip the golden — SPEC/RUNBOOK correctly call this
  out and PLAN.md's T-17 has a "grep-guard test asserts the package does not call `RegisterFactory`"
  requirement; T-16's parity test also asserts legacy searxng stays registered via
  `RegisterNativeLive`. Good.
- Byte-identical registrygen regen guard (`git diff --exit-code internal/connectors/registryset/`)
  is present in T-16's verify command and RUNBOOK's failure-triage table. Sound.

## 5. PaginationSpec gap fix (stripe `starting_after`/`has_more`)

**PASS.** `connsdk.Paginator` (ground truth, `paginate.go`) is `Start()/Next(resp, recordCount)
*NextPage`, driven by `connsdk.Harvest`. There is currently **no** `CursorPaginator` variant that
reads the next token from the **last emitted record** (only `CursorPaginator.TokenPath` reads a
body-level token path). Legacy `stripe.go:147` (`harvest`) hand-rolls this exact loop today
(`startingAfter` = last record's `id`; stop when `has_more != "true" || lastID == ""`). PLAN.md's
B-06 correctly identifies this as a genuine engine gap and specs two **new**, engine-local
`connsdk.Paginator` implementations (`lastRecordCursor`, `nextURL`) that satisfy the existing
interface without modifying `connsdk` itself — consistent with the "connsdk is NOT modified" rule
and with DATA-MODEL §2's `PaginationSpec` shape (`last_record_field`, `stop_path`, mutually
exclusive with `token_path` — constraint explicitly tested per TEST-PLAN §1.4 "cursor with both
token sources" malformed-spec case). TEST-PLAN §1.4 covers termination edge cases matching legacy
behavior: "empty page with `has_more=true`" defensive stop and "missing id field stop" — both
translate legacy's `lastID == ""` stop condition correctly. Covered.

Minor note: legacy stripe's `has_more` check is a **string** comparison (`hasMore != "true"`, via
`connsdk.StringAt` stringifying a JSON bool) — B-06/T-06 should use the same stringify-then-compare
semantics (or an equivalent JSON-bool-aware read) for the `stop_path` value to stay parity-safe;
this level of detail is not spelled out in PLAN.md/DATA-MODEL but is implied by "matching legacy
behavior" language. Not a blocker — flagging so the implementing agent checks it against the T-15
parity test rather than assuming a bare `bool` unmarshal.

## 6. Certify-core scope

**PASS.** Verified against `docs/architecture/connector-certification-design.md` "Implementation
order": step 1 = report.go+cliharness.go (T-12/B-12), step 2 = source stages 0-11 against
sample/github (T-14/B-14). SPEC §1.6 and PLAN.md both correctly exclude write/flow/schedule stages,
batch mode, `--all`, creds.yaml, record/replay httpx seam, `cmd/certstatus`, and CLI wiring
(`pm connectors certify` is design step 5) — matching the design doc's step ordering exactly. No
scope creep detected in T-12/T-14's task bodies.

One resolved ambiguity: the design's step-2 prerequisite note says "Prerequisite fix: `--credential`
on `pm etl check/read`". Ground truth check of `internal/cli/cli.go` confirms **no `--credential`
flag exists anywhere in cli.go today**, and `runETL`'s `check`/`read` cases resolve the connector via
`directConnector()`, which takes `--connector <name>` + `--config k=v` directly with no credential
store lookup. Since `sample` needs no secrets, T-14 can drive `etl check/read/run` against `sample`
without that flag — SPEC's claim "(not needed for sample; documented as wave1 prerequisite)" is
verified correct, not a gap.

## 7. No-new-dependency rule / validator keyword sufficiency

**PASS.** Cross-checked the minimal draft-07 dialect (`type` incl. arrays, `required`, `properties`,
`items`, `enum`, `pattern`, `minProperties`, `additionalProperties`, plus annotation-only `format`/
`default`/`title`/`description`/`$schema`) against every JSON Schema example needed by the three
goldens (DATA-MODEL §3) and the design doc's worked examples (§A `spec.json`/`schemas/issues.json`/
`writes.json`). No golden or design example uses `oneOf`, `anyOf`, `$ref`, `minimum`/`maximum`,
`multipleOf`, or `uniqueItems` — the keyword subset is sufficient for wave0's 3 bundles. The one
place where draft-07's minimal subset is insufficient (stripe's `create_customer` "email OR name"
legacy rule, which needs `anyOf`) is correctly identified and handled as a **documented deviation**
(approximated by `minProperties: 1`) rather than silently dropped or requiring a keyword addition —
exactly the intended escape valve (EVAL-PLAN §2 budget ≤2, currently 1 planned).

## 8. Parity test design provability

**PASS.** Ground truth confirms both `internal/connectors/stripe/stripe_test.go` and
`internal/connectors/searxng/searxng_test.go` already inject `httptest.Server` + `Client:
srv.Client()` and a `base_url` config override — the exact seam PLAN.md's T-15/T-16 assume ("legacy
`stripe.Connector{Client: srv.Client()}`... vs `engine.New(bundle, nil)` against the same server").
This makes "identical records for identical fixture input" directly and mechanically provable: one
`httptest.Server` instance serves both connectors in the same test, and their emitted record slices
are diffed. For postgres (no HTTP, SQL-based), the plan correctly falls back to `mode=fixture`
comparison — ground truth confirms legacy `postgres.go` already has a `fixtureMode`/`readFixture`/
`fixtureStreams` path with canned deterministic rows, so Tier-3 native/postgres can reuse the same
mechanism for provable Check/Catalog/Read parity without a live DB. Both mechanisms are sound and
match how the legacy code is actually testable today.

## 9. Task size / estimate sanity

**FLAG.** Most tasks (T-01..T-10, T-12..T-15, T-18, T-19) are appropriately scoped: 1-2 new files
plus tests, targeting a single behavior area. Two tasks are meaningfully larger than the "single
Sonnet agent run" budget implied by the rest of the plan:

- **T-11/B-11 (`cmd/connectorgen`)**: 4 Go source files (`main.go`, `validate.go`, `gen.go`,
  `new.go`) plus `main_test.go` plus **≥10 seeded-invalid testdata bundles** (EVAL-PLAN §3, "≥10
  seeded, ≥8 distinct classes"), each a multi-file directory (metadata.json/spec.json/streams.json/
  schemas/api_surface.json/docs.md). `validate` alone must implement 8 rule categories (structural
  load, schema compile, interpolation resolution, PK/cursor existence, write path_fields subset,
  api_surface rules 1-5, naming regex, docs.md heading check, fixture/secret scan) plus `gen`
  (deterministic codegen for 2 generated files) plus `new` (template scaffolding). This is
  realistically 500-800+ lines of new Go across 4 files plus ~10 small testdata trees — larger than
  any other single task in the plan and a plausible split candidate (e.g. split `validate` from
  `gen`+`new`).
- **T-17/B-17 (postgres Tier-3 native)**: ports all of legacy `postgres.go` (567 lines, ground-truth
  verified) into a 5-file component split (`connector.go`, `connection.go`, `reader.go`,
  `cataloger.go`, `cdc.go`) **plus** authors a full `defs/postgres/` bundle (metadata.json with
  `dynamic_schema: true`, spec.json with `password` x-secret, api_surface.json, docs.md) **plus** a
  fixture-mode parity test suite (Check/Catalog/Read equality, config-validation error-table parity,
  `Definition()` bundle-serving check, no-registration grep-guard) **plus** a documented CDC stub.
  Likely 600-900+ new lines across many files for a single task/agent run.

Recommendation: split T-11 into "validate" (B-11a) and "gen + new" (B-11b), and/or split T-17 into
"native/postgres component port" (B-17a) and "defs/postgres bundle + parity test" (B-17b). Neither
split is required for correctness — the dependency graph does not forbid it — but both tasks carry
higher repair-retry risk than the rest of the plan; EVAL-PLAN §7 already has a "repair-rate signal"
process that would catch this at execution time, but pre-splitting is cheaper than a mid-wave
repair loop.

## 10. THREAT-MODEL coverage vs planned tasks

**Two gaps found (FLAG).** Most THREAT-MODEL controls are explicitly test-covered: §1 secret
handling (T-04/T-09/T-12 secret-redaction/scan assertions), §2 urlencode-by-default (T-02's explicit
"path-segment injection"/"query metachars"/"double-encode guard" test rows, verified present in
TEST-PLAN §1.1), §4 fixture integrity (T-13's `pagination_terminates`/unmatched-request-fails
design), §5 write-path safety (no live writes; T-09's dry-run/validate-only tests), §6 supply chain
(no new deps enforced throughout; T-11's byte-stable gen output), §7 certify harness abuse (T-12's
ephemeral-root + no-shell-out design). However:

- **CR/LF header injection (THREAT-MODEL §2)**: "Header values: engine rejects interpolated header
  values containing CR/LF" is stated as a control but **no task in PLAN.md or TEST-PLAN.md tests
  it**. T-02 (interpolator) and T-08 (read path, which builds headers) list no CRLF/header-injection
  case. This is a real gap between the threat model and the plan — recommend adding a CRLF-rejection
  test case to T-02 or T-08 (whichever layer actually constructs headers) before or during
  execution.
- **SSRF same-host enforcement for `next_url`/Link-header follow (THREAT-MODEL §3)**: "engine
  requires same-host as the resolved base URL unless the bundle sets an explicit `allow_cross_host:
  true` escape... (T-06)" — T-06's actual test list (PLAN.md Wave B) only covers the loop-guard case
  ("same URL twice → error"), not the same-host enforcement itself. Ground-truth grep across
  DATA-MODEL.md, API-CONTRACT.md, and SPEC.md confirms **`allow_cross_host` does not appear
  anywhere** outside THREAT-MODEL.md — it is not a field in `PaginationSpec`'s JSON shape
  (DATA-MODEL §2), not in the Go `PaginationSpec` type (API-CONTRACT §2), and not in any task's file
  list or test case. This is either (a) a control that was designed but never wired into the data
  model/task list, or (b) stale threat-model prose describing an aspirational control that isn't
  actually planned for wave0. Either way it needs reconciling: either drop the `allow_cross_host`
  claim from THREAT-MODEL (if same-host enforcement isn't in scope for wave0's 3 goldens, none of
  which use `next_url`) or add the field + a same-host assertion to T-06/T-08.
  - Mitigating factor: none of the 3 wave0 goldens use `next_url` pagination (stripe uses
    `cursor`+`last_record_field`, searxng uses `page_number`, postgres has no HTTP pagination), so
    this gap does not block wave0's own acceptance criteria — but it is still a documented threat
    control with zero corresponding implementation/test commitment, which will surface as a false
    sense of security if not reconciled before a future connector actually uses `next_url`.
- Base-URL scheme/host validation ("mirrors `stripeBaseURL`... enforced once at requester build")
  has no task explicitly naming it as a test case (unlike the CRLF/SSRF items above, this is milder
  since it's implied general engine hygiene rather than a named escape-hatch field) — worth a line
  in T-03 (bundle/HTTPBase) or T-08 (read path/requester build)'s test list for completeness, but
  not treated as a standalone flag given it's a natural extension of "build a Requester" work
  already in scope.

## Additional finding: stale cross-reference (minor, not one of the 10 requested checks but caught
## during dependency-graph review)

T-11's note ("seeded-invalid corpus is shared with conformance self-tests (B-12) — B-11 owns
`cmd/connectorgen/testdata/invalid/`, B-12 symlinks/copies nothing: it has its own corpus.") cites
**B-12**, but B-12 is "certify report + cliharness" (Wave D), which has nothing to do with
conformance self-tests or a seeded-invalid corpus. The task that actually owns its own
`testdata/invalid/**` conformance corpus is **B-13** ("conformance v2", Wave E, PLAN.md line 183:
`testdata/invalid/**`). This is a copy-paste/reference error in the plan text — recommend correcting
"(B-12)" to "(B-13)" in T-11's note so a coordinator/agent doesn't go looking for a corpus reference
in the wrong task.

---

## Summary of required revisions before/during dispatch

1. **(should-fix, not blocking)** Reconcile T-12/B-12's wave placement: either move it into Wave B
   in the dependency graph and wave tag (honoring DECISIONS.md #2), or explicitly note in PLAN.md
   why the coordinator's approval to float it early was not applied.
2. **(should-fix, not blocking)** Correct T-11's note: "(B-12)" → "(B-13)" for the shared
   seeded-invalid corpus cross-reference.
3. **(should-fix, not blocking)** Add a CRLF/header-injection test case to T-02 or T-08 to close the
   THREAT-MODEL §2 gap, or downgrade THREAT-MODEL's claim if intentionally deferred.
4. **(should-fix, not blocking)** Reconcile `allow_cross_host`/same-host enforcement: either add the
   field to `PaginationSpec` (DATA-MODEL §2, API-CONTRACT §2) and a same-host test to T-06, or remove
   the claim from THREAT-MODEL §3 if genuinely out of scope for wave0 (no golden uses `next_url`).
5. **(recommended, not blocking)** Consider splitting T-11 (`cmd/connectorgen`) and T-17 (postgres
   Tier-3) into two sub-tasks each, given their size relative to the rest of the plan's tasks.

None of these are blockers to starting Wave A — items 1-2 are documentation/scheduling hygiene,
items 3-4 are security-test completeness gaps with no impact on the 3 wave0 goldens (none use the
affected code paths), and item 5 is a proactive scope-risk mitigation. Recommend the planner apply
items 1-4 as a fast documentation pass before Wave D/Wave B dispatch (whichever comes first for
B-12) and before Wave B dispatch for the CRLF/SSRF test additions, and consider item 5 when
assigning agents to Wave D/F.

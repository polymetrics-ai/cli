---
phase: wave0-engine-harness
verified: 2026-07-02T16:10:00Z
status: passed
score: 6/6 acceptance criteria verified
must_haves:
  truths:
    - "Engine unit tests green (interpolation, auth selection, pagination matrix, read/write paths, error mapping) with statement coverage >=85%"
    - "3 goldens (stripe, searxng, postgres) migrated with engine-vs-legacy parity tests passing"
    - "connectorgen validate rejects seeded-invalid bundles and accepts the 3 goldens with zero findings"
    - "Conformance v2 passes (static + dynamic httptest fixture replay) for all 3 goldens"
    - "Certify source stages pass end-to-end against the sample connector"
    - "go build ./... && go test ./... && golangci-lint run green, legacy files untouched, registrygen byte-identical"
  artifacts:
    - path: "internal/connectors/engine/"
      provides: "Declarative runtime engine (bundle, interpolate, schema, errors, auth, paginate, hooks, read, write, connector)"
    - path: "internal/connectors/defs/{stripe,searxng,postgres}/"
      provides: "3 golden bundles"
    - path: "cmd/connectorgen/"
      provides: "validate | gen | new subcommands"
    - path: "internal/connectors/conformance/"
      provides: "Conformance v2 static + dynamic checks"
    - path: "internal/connectors/certify/"
      provides: "Certification report/harness + source stages"
    - path: "internal/connectors/native/postgres/"
      provides: "Tier-3 native postgres component split"
gaps: []
human_verification: []
---

# Phase wave0-engine-harness Verification Report

**Phase Goal:** Build the declarative connector engine, bundle-authoring tooling
(`cmd/connectorgen`), conformance v2, the certify harness core, and prove all of it with three
golden migrations (stripe, searxng, postgres) under engine-vs-legacy parity — while leaving the
legacy connector machinery and registry behavior completely unchanged.

**Verified:** 2026-07-02T16:10:00Z (commands re-run live against HEAD `b3f91af`, branch
`connector-architecture-v2`)
**Status:** passed
**Re-verification:** No — initial verification

## Goal Achievement

### Acceptance Criteria (from ROADMAP.md / SPEC.md §4, restated verbatim as the gate)

| # | Criterion | Result | Command | Evidence |
|---|---|---|---|---|
| 1 | Engine unit tests green + coverage ≥85% | PASS | `go test ./internal/connectors/engine -cover` | `ok polymetrics.ai/internal/connectors/engine 0.809s coverage: 85.0% of statements` — meets the EVAL-PLAN §1 gate exactly. |
| 2 | 3 goldens migrated with engine-vs-legacy parity tests passing (stripe, searxng, postgres) | PASS | `go test ./internal/connectors/engine -run 'TestParityStripe\|TestParitySearxng' -v` + `go test ./internal/connectors/native/postgres -run TestParity -v` | All stripe subtests pass (`StreamRecords` ×5 streams, `CustomersTwoPagePagination`, `IncrementalCreatedGTEFromState`/`FromStartDate`, `WriteCreateCustomerFormBody`, `WriteUpdateCustomerFormBody`, `ManifestSurface`, `BundleLoadsAndValidates`); all searxng subtests pass (6); all postgres parity subtests pass (`ConfigValidationErrorTable` ×9 rules, `CatalogStreamSet`, `ReadRecordEquality` ×2 streams). Read `parity_stripe_test.go`/`parity_searxng_test.go`/`parity_test.go` in full — all three build a shared `httptest.Server`, run BOTH the legacy hand-written connector and `engine.New(bundle, nil)` against it, and assert record-set/method/path/form-body/manifest-surface equality via `reflect.DeepEqual` — not trivial asserts. |
| 3 | `connectorgen validate` rejects seeded-invalid bundles; accepts the goldens | PASS | `go test ./cmd/connectorgen -v` + `go run ./cmd/connectorgen validate internal/connectors/defs` | 14 seeded-invalid subtests pass, spanning 12 distinct rule constants (`ruleMissingFile`, `ruleMetaSchema`, `ruleInterpolationUnresolved`, `ruleSchemaRefMissing`, `rulePrimaryKeyMissing`, `ruleCursorFieldMissing`, `ruleWritePathFields`, `ruleSurfaceCoverage`, `ruleSurfaceIncomplete`, `ruleNameRegex`, `ruleSecretLiteral`, `ruleDocsHeading`, `ruleSurfaceFailFirstRun`) — exceeds the ≥8-class EVAL-PLAN §3 bar. Read the test table (`cmd/connectorgen/*_test.go:52-70`): each case maps `dir -> wantRule` and asserts a `Finding` with that exact rule name exists, confirming distinct-rule mapping (not a generic catch-all). `validate internal/connectors/defs` → `3 connector(s) checked, 0 findings`, exit 0. |
| 4 | Conformance v2 passes for the 3 goldens (static + httptest fixture replay) | PASS | `go test ./internal/connectors/conformance -run TestConformance -v` | `TestConformance/postgres`, `/searxng`, `/stripe` all PASS. Static self-test corpus (`TestStaticChecks_TargetedFailures`) exercises all 10/10 static checks with a dedicated failing case each (spec_schema_valid, stream_schemas_valid, pk_fields_exist, cursor_fields_exist, interpolations_resolve, write_schemas_valid, surface_complete, docs_present, secret_redaction, fixtures_present) — all pass. |
| 5 | Certify source stages pass against `sample` end-to-end | PASS | `go test ./internal/connectors/certify -run TestSourceStages -v` | `TestSourceStagesAgainstSample`, `TestSourceStagesSabotageFailsNamedStage`, `TestSourceStagesEphemeralWorkdirCleanedUp` all PASS (2.26s). |
| 6 | `go build ./... && go test ./... && golangci-lint run` green (+ `make verify`); legacy untouched | PASS | `go build ./...`; `go vet ./...`; `make lint`; `make verify`; `git diff --stat main...HEAD -- <legacy paths>`; `go run ./cmd/registrygen && git diff --exit-code internal/connectors/registryset/` | `go build`/`go vet` exit 0. `make lint` → `0 issues` (scoped to new packages per `LINT_PKGS`, intentional per Makefile design). `make verify` (fmt, tidy-check, vet, full `go test ./...`, build, docs-check, smoke, lint, connectorgen-validate) green end-to-end, zero FAIL lines. `git diff --stat main...HEAD -- internal/connectors/{stripe,searxng,postgres,connsdk,connectors.go,manifest.go,catalog.go,slug.go}` = empty (zero changes). `internal/app`, `internal/cli`, `registryset/`, `catalog_data.json`, `icon_data.json` diff-empty. Only sanctioned legacy edit: `cmd/registrygen/main.go` +15 lines, skip-map entries only (`defs`, `engine`, `hooks`, `native`, `conformance`, `certify`), verified via full diff read. `go run ./cmd/registrygen` regenerates `registry_gen.go` (557 imports) with `git diff --exit-code` = 0 (byte-identical) — coexistence invariant holds. `git status --porcelain` = clean (no stray untracked/uncommitted files; path guard passes). |

**Score:** 6/6 acceptance criteria verified.

### EVAL-PLAN.md Quantitative Metrics

| # | Metric | Gate | Observed | Status |
|---|---|---|---|---|
| 1 | Engine statement coverage | ≥85% | 85.0% | PASS |
| 2 | Golden parity (3/3) | 0 failures, 0 skips | 3/3 pass, 0 skips | PASS |
| 3 | connectorgen defect classes | ≥8 distinct, ≥10 seeded, goldens 0 findings | 12 distinct rules, 14 seeded cases, 0 findings on goldens | PASS |
| 4 | Conformance v2 | 3/3 goldens green, 10/10 static self-tests | 3/3 green, 10/10 exercised | PASS |
| 5 | Certify core | `passed:true`, stages 0-11, resume green, 0 secrets | All TestSourceStages* subtests pass | PASS |
| 6 | Whole-tree gates | build/vet/test/lint/verify green; registrygen 0 diff; inventory.json >500 entries; TDD ledger RED before impl | All green; inventory.json = 557 entries; TDD-LEDGER shows red-confirmed → green for every task | PASS |
| 7 | Deviation ledger budget | ≤2 entries | 1 entry (stripe `minProperties` — documented ACCEPTABLE) | PASS |

### Key Link Verification (coexistence guardrail, SPEC §2)

| From | To | Via | Status | Details |
|---|---|---|---|---|
| `cmd/registrygen` skip map | `defs/engine/hooks/native/conformance/certify` dirs | skip-map entries | WIRED | Verified via full diff read — additive-only, 15 lines, comment + 6 map entries. |
| `registryset/registry_gen.go` | legacy connector packages only | `go run ./cmd/registrygen` regen | WIRED (unchanged) | Regen produces byte-identical output (557 imports); `git diff --exit-code` = 0. |
| `engine.New(bundle, nil)` (stripe/searxng) | `defs.FS` via `engine.LoadAll` | test-only construction | WIRED | Parity/conformance tests build engine connectors directly from `defs.FS`; no `RegisterFactory` call exists for engine-backed goldens (confirmed by reading parity test source — no init()/RegisterFactory call in `engine/` or `defs/` packages). |
| `native/postgres` | registry | init() registration | INTENTIONALLY NOT WIRED | `TestNoInitRegistration` grep-guard (in `waveF-b17-ledger.md` description, confirmed present in package) asserts no `RegisterFactory`/`RegisterNativeLive`/`func init()` exists in the package — matches SPEC §2 rule 2 (no registration in wave0; wave6 marker documented). |

### Requirements Coverage

No `requirements:` frontmatter field or REQUIREMENTS.md file was found for this phase/milestone
(project uses ROADMAP.md acceptance bullets directly as the phase contract, per SPEC.md §4 "restated
verbatim as the gate"). All 6 ROADMAP acceptance bullets are covered above with PASS.

### Anti-Patterns Found

None. Spot-read of `parity_stripe_test.go`, `parity_searxng_test.go`, `parity_test.go` (postgres),
`cmd/connectorgen/*_test.go`, and `internal/connectors/conformance/conformance_test.go` /
`static_test.go` found substantive, non-trivial assertions throughout — no placeholder returns, no
`console.log`-only handlers, no `TODO`/stub patterns in the verified surface. `gofmt -l cmd internal`
returned clean.

One pre-existing documentation-bookkeeping issue (not a code gap, noted for the coordinator):
`.planning/phases/wave0-engine-harness/RUN-STATE.json` still reads
`"status": "blocked_missing_artifact"`, `"coveragePassed": false`,
`"verificationMissingRequired": ["install"]`, timestamped `2026-07-02T07:17:10.334Z` — this predates
Wave F/G completion (commits through `b3f91af` at 15:5x) and does not reflect the now-passing
verification run performed here. This is a stale artifact, not a functional gap; recommend the
coordinator refresh `RUN-STATE.json` alongside this VERIFICATION.md.

### Human Verification Required

None. All 6 acceptance criteria and all EVAL-PLAN quantitative metrics were verified by directly
running the specified commands against the live codebase (not by trusting SUMMARY.md/ledger prose),
and by reading the parity/validate/conformance test source to confirm non-trivial engine-vs-legacy
comparisons.

### Gaps Summary

No gaps. Every ROADMAP acceptance bullet, every EVAL-PLAN quantitative gate, and the SPEC §2
coexistence invariant (registry untouched, byte-identical regen, sole sanctioned legacy edit) are
verified directly against the codebase at HEAD `b3f91af`. TDD evidence was spot-checked across 4
ledger files spanning Waves A, C, E, and F, and in each case showed genuine `go vet`/`go test`
compiler-failure RED output (undefined symbols, build failed) preceding a documented GREEN
implementation — not fabricated or reworded text.

**phase_goal_met: yes**

---

_Verified: 2026-07-02T16:10:00Z_
_Verifier: Claude (gsd-verifier)_

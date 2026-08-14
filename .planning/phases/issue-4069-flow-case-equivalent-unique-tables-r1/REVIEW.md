# #4069 Manual Code Review

**Mode:** Standard inline review via the resolved code-review GSD prompt under
the documented manual-GSD fallback.

**Result:** PASS — no unresolved candidate finding.

## Review boundary

The review covers the new canonical-equivalent query-view policy and its
real-Parquet/DuckDB flow regression. This issue is not an active numbered
ROADMAP phase and the parent contract forbids role spawning, so the generated
GSD review was executed inline. That fallback changes the execution mechanism,
not the TDD, verification, or review standard.

## Reviewed risks

- The policy is built once from the existing immutable resolver snapshot and
  only reads its maps after construction; it neither rescans warehouse state
  nor parses, filters, or rewrites caller SQL.
- Explicit connection-scoped queries keep the zero policy and bind only the
  selected owner's table path.
- Exact-name ambiguity remains on the existing #4066 path. A canonical group
  is introduced only when each exact name is independently unique, preventing
  the new logic from weakening the three-table case-variant matrix.
- Unscoped generic queries suppress only colliding bare views and retain
  non-colliding generated aliases, so unrelated SQL such as SELECT 1 does not
  fail while views are registered.
- Unscoped flow replacement scans create a fresh warehouse ambiguity error
  with a defensive connection-list copy. The existing flow boundary therefore
  preserves errors.As and applies its truthful connection remedy rather than
  exposing DuckDB's catalog error.
- The test fixture writes only test-owned local Parquet state and verifies
  selected rows, generic SQL, quoted/unquoted omitted-flow forms, no result
  leakage, and no successful checkpoint.

## Verification inspected

- Full affected packages and the focused #3897/#4066/#4069 race selector pass.
- Existing generated-owner-alias, case-variant, exact ambiguity, reverse/read,
  action-source, and schedule selectors remain green.
- Repository vet/configured lint and direct app/CLI lint are clean.
- Candidate diff formatting is clean against the immutable audited start.

## Findings

No Critical, Warning, or Info finding was identified in the #4069 range. The
seven broader git diff --check whitespace reports belong solely to older #3897
planning files and are recorded as inherited baseline, not suppressed or
modified by this child.

## Correction 1 / 5 inline code review

**Mode:** standard inline GSD fallback after resolving
`scripts/gsd sources code-review` and `scripts/gsd prompt code-review`. The
issue is not a numbered ROADMAP phase and the canonical contract forbids role
spawning, so the worker performed the required cross-file review itself.

**Result:** PASS — no Critical, Warning, or Info candidate finding.

### Reviewed correction boundary

- `CreateConnection` resolves destination capability, applies stream defaults,
  and validates the whole local inventory before `prefixedID` or `save`.
  Exact duplicate spellings and non-local destinations remain controls.
- `RunETL` evaluates the persisted local-warehouse inventory after the
  connection/stream lookup but before a run ID and `beginRun`, so legacy state
  is not migrated yet cannot mutate state or warehouse output through another
  sync.
- `QuerySQL` owns an immutable configuration collision snapshot beside the
  resolver snapshot. The DuckDB policy suppresses only collision base and
  invented alias bindings; it returns the new typed same-owner error through
  replacement scans, preserving generic `SELECT 1` and avoiding DuckDB catalog
  text or a false one-owner `AmbiguousTableError` remedy.
- A resolver-visible physical alias is checked before invented alias
  suppression. The direct query/action/reverse paths were deliberately left
  resolver-exact, with regression tests for a missing case variant on both
  case-sensitive and case-insensitive host behavior.
- The helper folds ASCII only, sorts collision data deterministically, copies
  error slices, keeps the policy private to `App.QuerySQL`, and neither opens a
  credential nor introduces a provider, transport, SQL-rewrite, or alias
  reservation surface.
- Generated CLI manuals/golden transcripts and website docs accurately state
  the creation refusal and legacy recovery without inventing an edit/delete
  command or an unusable connection selector.

### Evidence inspected

- Strict committed RED `6e0a4a7cb`, minimum GREEN `06414fa44`, concrete flow
  type assertion `b182d559e`, and real-alias control `f2eacc769` all precede
  this review record.
- Affected app/CLI/flow/warehouse/schedule package tests, focused race,
  formatting/diff/vet/lint/tidy/build/smoke, canonical generators and release
  checks, docs/help/website checks, issue guard, and `verify-gsd-workflow`
  passed; exact commands are in `traces/correction-1-broad-verification.txt`.
- Website lint's 13 warnings and the local default-auth build diagnostic are
  outside this correction's changed paths and did not fail typecheck or build.

### Finding disposition

No candidate finding remains. The seven inherited #3897 planning-file
trailing-whitespace findings remain outside the child range, and are neither
changed nor used to weaken this correction's gates. The active no-mistakes
monitor remains untouched; push, PR metadata, exact-head CI, and Sol audit are
explicitly outside this local review.

## Authorized Website aggregate gap review

**Mode:** standard inline GSD code-review fallback after resolving
`scripts/gsd prompt code-review issue-4069-flow-case-equivalent-unique-tables-r1`.
No compatible isolated Pi reviewer is available and the canonical contract
forbids role delegation; this is a review of one deterministic generated file
and its explicit GSD evidence, not a waiver of review.

**Result:** PASS — no Critical, Warning, or Info candidate finding.

### Reviewed boundary and evidence

- The only non-evidence change is `website/lib/docs.generated.ts`, produced by
  the checked-in package generator from pre-existing #4069 source pages. Its
  SHA-256 exactly matches the independently audited expected output.
- A second no-install generator run is byte-identical and leaves the same sole
  changed path; `git diff --check` is clean. No React source, dependency,
  credential, runtime, connector, workflow, or PR topology path was changed.
- Website lint exits 0 with the known 13 warnings entirely outside this diff;
  typecheck, 80 unit tests, 27 script tests, and the production build pass.
  Browser E2E appropriately remains for GitHub CI because the local recovery
  does not authorize Playwright installation.
- The exact #4071 failure is causally reproduced by generator ownership, not
  suppressed: the committed RED trace names the failing run IDs and the GREEN
  trace proves the scoped correction.

## Correction 2 / 5 inline code review — flow manual generator parity

**Mode:** standard inline GSD `code-review` fallback after resolving the
project prompt. The existing #4069 phase is not a numbered ROADMAP phase and
the delivery contract forbids role delegation; the fallback preserves the
review requirement.

**Result:** PASS — no Critical, Warning, or Info candidate finding.

### Reviewed boundary

- `internal/cli/docs.go` remains the manual authority. The exact three
  fail-closed owner-selection lines are inserted immediately after the existing
  omitted-owner refusal in `flowHelp`; no checked-in CLI markdown is used as a
  source of truth.
- The one-word `connectionsHelp` repair is causally required by the same golden
  map iteration and accurately describes `RunETL`'s configured local-warehouse
  inventory validation before a run begins. It is help-text parity only, not a
  runtime behavior change.
- The golden transcript generator changes only the expected flow/connections
  help, bare-manual, and JSON-manual values. `pm docs generate` reproduces the
  tracked manuals without a markdown diff, and the website generator has no
  output dependency on this manual change.
- The candidate range contains no provider credentials, SQL or warehouse-policy
  logic, connector metadata, transport wiring, dependency, workflow, PR, or
  website-source change. The existing #4069 typed ambiguity, generic SQL, and
  direct-read behavior therefore remain untouched.

### Evidence inspected

- The committed causal RED records both the map-order companion mismatch and
  the exact `flow.md` CI failure before the authoritative source edit.
- Focused manual/golden/docs selectors, full `internal/cli`, vet, configured
  lint, docs/contract/connectorgen checks, no-install website lint/typecheck/
  unit/script/build, and `scripts/verify-gsd-workflow` pass.
- The candidate diff from `678e294568a8a010a460ecb05fe11a42e1eb40f2` is
  clean under `git diff --check`; the 13 Website lint warnings are inherited
  and outside touched paths.

### Finding disposition

No candidate finding remains. The remaining no-mistakes custody and natural
exact-head CI are delivery gates, not local review findings; #4071 topology is
explicitly untouched.

## Correction 3 / 5 inline code review — destination-scoped admission

**Mode:** standard inline GSD `code-review` fallback after resolving the
project prompt. The issue is not a numbered ROADMAP phase and its contract
forbids role delegation; this retains the required cross-file review.

**Result:** PASS — no Critical, Warning, or Info candidate finding.

### Reviewed boundary

- `RunETL` retains the existing whole configured local-inventory check and its
  pre-`prefixedID`/pre-`beginRun` position, but now invokes it only when the
  selected connection reports local materialization through the existing
  credential-free abstraction.
- `connectionMaterializesLocalWarehouse` remains the sole destination
  classifier. The correction neither compares a connector name nor introduces
  a warehouse literal or a connector-specific branch in shared code.
- The restored regression proves the false positive is removed at the runtime
  boundary. The existing legacy local test still verifies the true positive's
  `*warehouse.SameOwnerCaseEquivalentTableError`, typed chain, and zero-run/
  zero-warehouse-mutation behavior.
- The changed production branch is small, keeps errors returned rather than
  logged, and introduces no new input parsing, credential, I/O, concurrency,
  dependency, or SQL behavior.

### Evidence inspected

- RED on exact source `3b75f4a62fd8d743ec883a5b824164374f661857`, GREEN on
  verified production commit `a49ae952d52a6edf9786cc90e02825e2414f5b44`, and
  GSD evidence checkpoint `5a4f3d23645d3046d02055273fedd9a8b9dd67d1` are
  recorded in the paired correction-3 traces.
- Focused false-positive/true-positive tests and full `internal/app` passed;
  `go vet`, direct app lint, build, tidy/configured lint, contract,
  surface-sync, release parity, candidate diff check, and
  `verify-gsd-workflow` passed.

### Finding disposition

No production-code finding remains. Child-PR delivery is pending its owning
executor; #4071 is not body-patched, retargeted, merged, closed, or
readiness-mutated.

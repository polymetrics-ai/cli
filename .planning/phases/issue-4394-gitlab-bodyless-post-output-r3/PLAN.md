# GitLab R3 source-output policy conformance plan

## Task Delivery Header

- Issue: Refs #4394 — GitLab complete connector-definition artifact projection; related evidence for #4384 and #4413.
- Base branch: `fm/cli-top100-declaration-batch-r1` at frozen commit `ceaae873aef0dd19aa23c036b9cb598f9b3eacc8`.
- Merges into: `fm/cli-top100-declaration-batch-r1 → main`.
- Delivery: Candidate branch committed and pushed for an independent exact-SHA review; no merge or parent integration in this task.
- Working branch: `codex/4394-gitlab-bodyless-post-output-r3`.
- Task: Correct the R2 GitLab bodyless-POST direct-read projection so four source-status responses use `none` and four source-JSON Conan responses retain `json_redacted`; make the closed source-projection admission enforce that distinction; preserve exact R2 source, matrix, enabled-contract, and runtime boundaries.
- Verification: Scoped source-projection, GitLab definition/materialization, engine wire/output, and public credential-boundary tests; focused vet and race gates when capacity permits; JSON parsing, surface synchronization, Atlas/schema checks, source reconciliation, and diff inspection.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Status-only source rows use `none` end-to-end | fake | A local `httptest` provider is necessary because this task has no credentialed GitLab sandbox. It asserts the actual emitted POST has zero request bytes/no inferred content type and that an empty status response yields no decoded body; source facts supply the expected status/media. |
| JSON Conan source rows keep `json_redacted` | fake | The same bounded fixture is necessary for non-credentialed provider response observation. It asserts the exact JSON payload is decoded under the declared policy, not just that the operation parsed. |
| Source projection rejects policy drift | live | The real `connectorgen` source-projection validator loads the GitLab source/matrix/artifacts and rejects a status/JSON policy mismatch or operation/CLI mismatch; it passes only for the two source-backed pairs. |
| Public command reaches the credential boundary without provider I/O | live | The real CLI/commandrunner resolves each generated command with complete typed flags, reports missing credential, and a spy transport observes zero requests. |
| Existing direct-read runtime behavior is retained | fake | An isolated engine fixture is required to observe the wire/body contract without credentials. It covers normal, malformed, and edge status/JSON cases; no production engine code changes. |
| Atlas accurately describes the reused form | live | Catalog JSON/schema and a focused content assertion show the existing Atlas record names the status/JSON policy split and only the existing direct-read foundation. |

## Scope boundary

### In scope

1. Cherry-pick the already-reviewed GitLab R2 candidate
   `a87908418b4bf69fa7b49bd64ae9ac8fa6a574bd` onto this frozen branch as the
   R3 starting point, preserving its source-backed artifact set.
2. Change the four exact status-source operation and CLI policy declarations to
   `none`; retain the four exact Conan policy declarations as `json_redacted`.
3. Add a closed source-output/policy conformance check in the existing
   source-projection admission path, plus source-derived red/green/edge tests.
4. Correct only the existing `runtime.source-bound-bodyless-post-read.v1`
   Foundation Atlas proof/consumer text to describe the two source response
   cohorts accurately.
5. Record planning, TDD, verification, and final review evidence under this
   phase directory.

### Explicitly out of scope

- Source locks, retained source bytes, descriptors, exact source IDs, matrix
  rows/counts, enabled-contract IDs, legacy write views, and all other
  connectors.
- `internal/connectors/engine/*.go`, commandrunner production code, receivers,
  generic runtime foundations, certification behavior, importer parsing, and
  credentialed provider I/O.
- Any rewrite, rebase, merge, force-push, parent integration, PR action, or
  direct `main` change.

## Locked source and reconciliation facts

- GitLab R2 retains 1,754 source rows: 1,752 primary plus two supplemental
  binary rows, represented by 12,278 seven-lane cells.
- The exact bodyless POST cohort has eight source IDs. Four source facts record
  no success media (status-only); four record `application/json` success media
  (Conan upload-URL lookups).
- Source IDs, lane dispositions, matrix counts, and enabled-contract identity
  sets are immutable for R3. The correction changes only projection output
  policy and the conformance evidence which prevents a recurrence.

## Foundation check

| Required contract | Atlas result | Classification | R3 action |
| --- | --- | --- | --- |
| Exact source-bound, no-request-body REST POST direct read | `runtime.source-bound-bodyless-post-read.v1`; owner `engine/bundle.go` and `engine/direct_read.go`; existing engine proof tests | reuse | Preserve engine/runtime. Correct source-output policy declaration and evidence only. |
| Status-only response policy | Existing `output_policy: none` runtime behavior and bounded status-response tests | reuse | Require `none` when retained source response class is status/no media. |
| JSON response policy | Existing `json_redacted` direct-read behavior | reuse | Require `json_redacted` when retained source says JSON media. |

No genuine shared runtime foundation is missing. The only production Go change
allowed by this plan is declaration-validation conformance in
`cmd/connectorgen/sourceprojection.go`; it has no provider I/O and no runtime
execution effect.

## Red–Green–Refactor execution

### 1. Establish the exact R2 baseline

- Verify this branch is still the frozen base and clean.
- Cherry-pick R2 with `-x`; do not rebase or reconstruct its source artifacts.
- Re-run the R2 targeted baseline to prove the correction is being made against
  the exact reviewed cohort.

### 2. Red: source-policy mismatch must fail

- Add a table-driven GitLab cohort test declaring expected `{source ID,
  method,path,source response class/media,operation policy,CLI policy}`.
- Initially require `none` for the four status-only rows while R2 still exposes
  `json_redacted`; record the resulting focused test failure.
- Add validator tests that reject: status + `json_redacted`, JSON + `none`,
  operation/CLI policy drift, undeclared source binding, and malformed output
  policy. Retain valid status/`none` and JSON/`json_redacted` cases.

### 3. Green: minimal artifact and declaration validation correction

- Update only the four matching `operations.json` entries and their
  `cli_surface.json` commands to `output_policy: "none"`.
- Preserve all four Conan rows as `json_redacted`; make the cohort test assert
  this explicitly.
- In the existing closed bodyless-POST source-projection predicate, bind source
  response fact to exactly one allowed output policy and require the operation
  and CLI policies to match. Do not infer from method alone and do not loosen
  any source binding, request-body, or endpoint constraint.
- Update the existing Atlas entry only enough to state the four status/none and
  four JSON/json-redacted consumers and the revised test evidence.

### 4. Refactor and proof

- Use named table cases; no new generic abstraction or connector-name branch.
- Extend the engine fixture to exercise all eight exact source IDs, with empty
  status success bodies and JSON Conan success bodies. It must also prove a
  nonempty status response remains rejected and caller-supplied body/raw-body
  values fail before I/O.
- Extend the public CLI no-I/O test across the eight commands and exact required
  path flags. A missing credential is the expected reachability result; any
  request attempt is a failure.

## Planned validation

1. `go test -count=1` for the focused `cmd/connectorgen` source-projection and
   GitLab enabled-contract tests.
2. `go test -count=1` for `internal/connectors/defs/gitlab`, including exact
   source/materialization/matrix/missing-foundation tests.
3. `go test -count=1` for the named engine direct-read POST body/output tests.
4. `go test -count=1` for named `internal/cli` GitLab credential-boundary and
   preflight tests.
5. `go vet` only for changed Go package directories; `go test -race` only for
   the changed engine/package test scope if the shared capacity floor permits.
6. `go run ./cmd/connectorgen validate internal/connectors/defs/gitlab` and
   `go run ./cmd/connectorgen surface-sync --check internal/connectors/defs/gitlab`.
7. JSON parse every changed JSON file, validate Atlas JSON/schema, run source
   reconciliation/counted cohort tests, `git diff --check`, and inspect the
   exact changed-file list.

## GSD and skills record

- Passed: `scripts/gsd doctor`; source resolution for canonical
  `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`,
  `gsd-verify-work`, and `gsd-code-review`; and
  `GOCACHE=/private/tmp/gocache-gitlab-r3 go run ./cmd/agentcontractgen check`.
- Rendered canonical GSD prompts and executed them inline because the current
  single-worker environment forbids GSD role-agent spawning. This is an
  explicit manual fallback, not a quality-gate waiver.
- Loaded: `golang-how-to`, `go-engineering` (including fundamentals and
  agentic ETL references), `golang-cli`, `golang-testing`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-code-style`, `golang-design-patterns`,
  `golang-structs-interfaces`, `golang-documentation`,
  `connector-lane-build-order`, `github-issue-first-delivery`, and
  `firstmate-exhaustive-review`.
- CLI help/manual/website parity: no user-visible command path, flag,
  help-text, manual, website, or generated docs change is planned. The public
  command reachability/no-I/O test is applicable; help/manual/website edits are
  explicitly not applicable.

## Commit/push checkpoints

1. R2 baseline cherry-pick is retained locally only; no push until the R3
   correction is green.
2. One scoped green commit contains the correction, tests, Atlas proof text,
   and phase evidence.
3. Push normally to `origin/codex/4394-gitlab-bodyless-post-output-r3`, verify
   the exact remote SHA, then await independent review. Do not open a PR,
   integrate, or merge unless directed by the parent/captain.

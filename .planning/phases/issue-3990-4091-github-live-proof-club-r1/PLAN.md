---
phase: issue-3990-4091-github-live-proof-club-r1
plan: "01"
type: tdd
wave: 1
depends_on: []
files_modified:
  - .planning/phases/issue-3990-4091-github-live-proof-club-r1/
  - scripts/github-live-*.mjs
  - scripts/tests/github-live-*.test.mjs
autonomous: true
requirements:
  - ISSUE-3990-LIVE
  - ISSUE-4091-LIVE
---

# Live GitHub certification proof plan

## Task Delivery Header

- Issue: Refs #3990 — GitHub Certification: enforce REST and GraphQL budgets across the whole run; Refs #4091 — authorize set-replace and keyed issue-label destination modes under explicit per-connection consent.
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1 → main`
- Delivery: Pull request open from `fm/cli-github-live-proof-club-r1` against `integration/4015-mvp-flat-r1`, with committed sanitized live evidence, edge-case accounting, and local gates green.
- Working branch: `fm/cli-github-live-proof-club-r1`
- Task: Run real-GitHub proofs for the landed #3990 and #4091 behavior through the built production binary; commit reproducible sanitized evidence, fix only defects exposed in this lane, comment #4091 with its required safe command/output evidence, and state live versus simulated coverage.
- Verification: Built-binary command/read-back assertions, credential and artifact scans, focused Node/Go tests, connector drift checks, generated-artifact one-pass verification, automated code review, and API read-back of the PR base.

## Evidence table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| #3990 production whole-run admission is live | live | Built `pm` uses the configured shared coordinator; safe report events account for requests sent/not sent and touched provider resource families; provider read-back/rate evidence shows no 429/abuse exhaustion. |
| #4091 approved non-additive write is live | live | A real run-owned GitHub target is changed through plan/preview/approval/execute and an independent PM read returns the exact label set. |
| #4091 replay/refusal is live | live | Replayed token or disabled/changed authorization yields the typed refusal; target labels and checkpoint remain byte-for-byte/identity-equivalent to their pre-attempt state. |
| Production composition is proven | live | The executed command begins at the built `pm` binary; evidence records binary identity and the PR names the entry → registration → dispatch → component chain. |
| Unavailable boundary cases | fake | Each unavailable real-provider condition gets its own PR-table reason and the closest existing production-binary or deterministic test; none is labeled live. |

## Threat model

| Threat | Control | Verification |
| --- | --- | --- |
| Credential disclosure | Secrets enter only from env/stdin; no shell tracing; evidence guard scans before commit/comment | Scan every changed/evidence file and captured output for secret-like material; stop immediately if found. |
| Wrong repository mutation | Default-deny immutable ID-bound boundary; PM-only commands; independent read-back and cleanup | Boundary validator passes before credential use; final residue equals declared baseline. |
| Approval replay or scope substitution | Single-use stdin token and durable shape-bound authorization | Replay and scope-drift refusals assert unchanged provider/checkpoint state. |
| Rate exhaustion | Require-shared admission, bounded deadlines, safe wait/reset/not-sent events | Before/after touched-resource accounting and no 429/abuse response. |
| Misclassified evidence | Live/simulated field per case; no success inferred from exit code | Validator rejects missing observable assertions and unsafe fields. |

## Tasks

<task type="tdd">
  <name>1. Freeze the proof contract and observe RED</name>
  <read_first>CONTEXT.md, both issue bodies, prior VERIFICATION files, scripts/github-live-lab.mjs, scripts/github-live-proof-sweep.mjs, internal/cli/github_transport_binary_test.go</read_first>
  <action>Inventory the production commands and existing harness gaps. If a reusable live-proof driver is missing, add a failing Node test for a sanitized two-issue proof record that requires binary identity, live/simulated classification, typed-refusal/negative-side-effect fields, edge-case coverage, exact safe commands/output, cleanup, and no secret/rate-scope material. If no source change is required, record RED as the two committed verification documents' explicit missing-live assertions and the current harness execution-model refusal.</action>
  <acceptance_criteria>The TDD ledger contains an observed failing command or an explicit evidence-only RED with the exact pre-existing gap; no provider write occurs during RED.</acceptance_criteria>
</task>

<task type="execute">
  <name>2. Prepare production binary, disposable project, boundary, credential, and coordinator</name>
  <read_first>AGENTS.md, runtime integration reference if Dragonfly is needed, pm credential/connectors help, GITHUB-LIVE-LAB-BOUNDARY.json</read_first>
  <action>Build one fresh `pm`. Validate its command/help path. Initialize a run-owned project. Configure the GitHub credential only through env/stdin and inspect metadata without secret values. Validate the immutable lab boundary before any provider call. Start or prove the real shared coordinator required by certification without touching unrelated runtime state.</action>
  <acceptance_criteria>The project and binary are inside this worktree, the credential is named but never rendered, boundary ID checks pass, coordinator availability is proven, and no provider mutation has happened.</acceptance_criteria>
</task>

<task type="execute">
  <name>3. Run #3990 real-binary whole-run proof</name>
  <read_first>issue-3990 verification, certify CLI implementation/help, GitHub rate policies, live harness artifact guards</read_first>
  <action>Invoke the production certification route with real GitHub credentials and require-shared coordination. Capture only sanitized attempts, waits/resets, not-sent outcomes, terminal results, and touched REST/GraphQL resource accounting. Exercise cancellation/deadline, coordinator loss, concurrent workers, process interruption/resume, empty/single/large command cohorts, duplicate/reordered deliveries where the harness supports them, and auth/permission refusal without crossing the lab boundary.</action>
  <acceptance_criteria>All live assertions name observable provider/coordinator state; every refused request records zero send; no 429/403-abuse exhaustion occurs; untestable cases are individually classified.</acceptance_criteria>
</task>

<task type="execute">
  <name>4. Run #4091 real-binary durable consent proof</name>
  <read_first>issue-4091 verification, etl transport CLI, issue-label transport binary test, GitHub live-lab planned-write/read helpers</read_first>
  <action>Create only run-owned issue/label fixtures. Execute set-replace and keyed modes through connection creation, plan, preview, stdin approval, ETL dispatch, independent read-back, identical-scope unattended rerun, token replay, and cleanup/restore. Assert the captain edge matrix: cancellation/process death, empty/single/large worksets, duplicate/out-of-order delivery, schema/scope drift, auth/permission refusal, concurrent same-target runs, resume, and acknowledged replay. Use typed refusal and unchanged label/checkpoint assertions for every refusal.</action>
  <acceptance_criteria>The legitimate live write changes exactly the intended lab object; independent read-back proves exact final state; token replay and all executable refusals make no additional change; cleanup leaves zero unexpected residue.</acceptance_criteria>
</task>

<task type="tdd">
  <name>5. Fix only live-exposed proof-path defects</name>
  <read_first>The failing production path and its closest existing tests</read_first>
  <action>For each genuine in-scope defect, first add a regression that fails through production composition and asserts the missing observable effect, then make the smallest production change, rerun RED→GREEN, and update this plan/TDD ledger. For unrelated files, append `needs-decision:` to firstmate status and stop.</action>
  <acceptance_criteria>Every production edit has preceding RED and passing GREEN; #4125 and #4158 are untouched.</acceptance_criteria>
</task>

<task type="execute">
  <name>6. Commit sanitized evidence, regenerate once, verify, review, and deliver</name>
  <read_first>Issue evidence rules, artifact guard, generation targets, automated review routing and Claude review workflows</read_first>
  <action>Commit sanitized evidence and updated lifecycle artifacts. Post #4091's exact safe command/output evidence to its issue. Regenerate all derived connector/website/transcript/skills/manual artifacts in one pass, run drift gates and focused tests, confirm clean status, execute verify-work and code-review prompts inline, push, open the stacked PR with an edge-case table and call chains, and read its base from the API.</action>
  <acceptance_criteria>No secret or rendered rate scope appears in diff/artifacts/comments; all focused/drift gates pass; PR base reads exactly integration/4015-mvp-flat-r1.</acceptance_criteria>
</task>

## Edge-case coverage contract

| Edge case | #3990 observation | #4091 observation | Refusal/side-effect assertion |
| --- | --- | --- | --- |
| Cancellation mid-operation | Bounded admission cancellation before send and run deadline behavior | Cancel before execute or during a safely controllable pre-send stage | Typed cancellation; zero sends/writes; checkpoint unchanged |
| Connection/process dies partway | Coordinator/process loss and resumed run accounting | Kill before provider send; resume same closed plan where supported | Terminal record plus no hidden send/checkpoint |
| Empty / single / large | Empty cohort, one command, current whole cohort | Empty refused, singleton succeeds, >1 closed workset refused or explicitly unreachable from CLI | Exact counts, never exit-only |
| Duplicate / out-of-order | Shared budget sees duplicate/reordered reservations without oversend | Repeated identical scope is idempotent/exact; stale/replayed token refused | Exact target state and send count |
| Schema drift | Rate policy/resource schema mismatch refuses before send | Changed bound mapping/config/scope requires reapproval | Typed mismatch and unchanged provider/checkpoint |
| Permission / authentication | Real provider refusal classified, no retry storm | Refusal before write or provider 401/403 with unchanged target | Typed/safe provider class and unchanged state |
| Concurrent same target | Multiple processes share one budget | Two production processes race the same target | Budget never exceeded; exact label set/checkpoint semantics |
| Resume after interruption | Resumed certification accounts for prior terminal state | Resume closed plan/authorization without token if exact scope | No duplicated unintended side effect |
| Replay acknowledged item | Already observed reservation/response does not double-charge/send | Approval-token replay rejected; identical authorized scope is a new permitted run | Replay typed error and no additional provider write |

## Verification

- Focused Node harness tests and artifact self-tests.
- Targeted `internal/connectors/engine`, `internal/connectors/certify`, `internal/coordination`,
  `internal/app`, and `internal/cli` tests with `-timeout 20m`.
- Built-binary live commands and independent PM/provider read-backs.
- `go vet` on changed packages and required non-full-suite `make verify` component gates.
- One-pass generator checks and `git diff --check` / clean-status audit.
- Inline/manual `verify-work` and `code-review` artifacts.

## Success criteria

- Real provider proof closes the explicitly documented live gaps without secret disclosure.
- Production entry/registration/dispatch/component paths are explicit and exercised.
- Every captain edge case is live, simulated with its own reason, or explicitly untestable with a
  concrete reason in the PR table.
- All run-owned mutations are independently read back and cleaned/restored.
- The branch is committed, pushed, and represented by a PR against the stated integration base.

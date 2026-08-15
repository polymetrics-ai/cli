# Issue 4166 Plan — Prove Certification Coverage Gaps

**Mode:** TDD, validation-only

## Objective

Close or honestly classify the three audited proof gaps without adding product capability. Every accepted slice must have a negative control that fails when its claimed production path is broken.

## Lifecycle and Skills

- Inline GSD lifecycle: `scripts/gsd prompt discuss-phase issue-4166-validation-proof-gaps-r1`; `scripts/gsd prompt plan-phase issue-4166-validation-proof-gaps-r1 --tdd`; `scripts/gsd prompt execute-phase issue-4166-validation-proof-gaps-r1`; `scripts/gsd prompt verify-work issue-4166-validation-proof-gaps-r1`; `scripts/gsd prompt code-review issue-4166-validation-proof-gaps-r1`.
- Inline/manual fallback reason: the repository's canonical issue-worker contract forbids spawning planner, verifier, reviewer, or orchestrator roles.
- Loaded skills: `golang-how-to`, `golang-testing`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, and `golang-safety`.
- CLI parity: no command, flag, help, output, manual, generated docs, or website contract is planned to change. If that assumption becomes false, stop production edits and add the parity checklist before continuing.

## Slice 1 — Gap 1: write coverage must fail closed

**Red**

1. Add `TestFullWriteSweepFailsForDeliberatelyBrokenDeclaredAction` against an action already represented by a safe pairing.
2. Add `TestFullWriteSweepFailsForDeliberatelyBrokenPreviouslyBlockedAction` against `update_issue` or `merge_pull_request`.
3. Both tests run the certification sweep, assert the report/stage fails with the exact action name, and include an intact control. The baseline code is expected to stay green while merely recording `pass + blocked/untested`; that is the RED.

**Green**

1. Make the full write sweep exercise each declared action's real declarative validation/preparation contract with no provider send.
2. Preserve the existing live create/verify/cleanup protocol only for curated safe pairings.
3. Record per-action results honestly: `pass` when declaration preparation is exercised, `fail` with the action-specific cause when it is broken, and an explicit non-live reason when a provider mutation remains unsafe.
4. Use a test-scoped sabotage/definition seam to break an action without editing or shipping the GitHub bundle.
5. Compute coverage from the resulting report: total GitHub operations, preflighted write actions, live provider operations, and remaining non-live categories/reasons.

**Refactor / checkpoints**

- Keep the seam Runner-scoped and immutable during a run; do not introduce mutable package-global definition overrides.
- Commit the RED and GREEN as coherent checkpoints when practical.

## Slice 2 — Gap 2: declared transport registration and execution

**Red**

1. Add `TestCertificationDeclaredTransportPairResolvesAndExecutes` using the production definition-owned GitHub connector, production transport composition, durable warehouse stage, and a bounded local HTTP provider fixture.
2. Assert source read, durable stage/reopen, destination plan/apply, provider read-back, and checkpoint acknowledgement occurred. A passing RunETL result without those counters/state transitions fails.
3. Add `TestCertificationDeclaredTransportPairFailsWhenDeclarationOrRegistrationIsMissing`. Remove or unregister exactly one transport role through a test-scoped registry/factory seam and assert the certification probe fails before claiming success.
4. Baseline is RED because current GitHub-to-warehouse certification never dispatches the declared pair and the existing construction test proves only preflight.

**Green**

1. Add the smallest certification probe/harness wiring that routes the proof connection through the same `App` transport dispatch used by production.
2. Keep provider behavior in the test fixture; do not add a new runtime route or special-case connector capability.
3. Ensure the inverse test fails specifically for absent/unregistered declaration and observes zero source/destination execution.

**Refactor / checkpoints**

- Reuse existing issue-label transport fixtures and production call paths rather than duplicating transport logic in certification.
- Call out GitHub-specific factories/dispatch guards in the final report; do not generalize or fix them in this validation PR.

## Slice 3 — Gap 3: built-binary flow round trip against disposable GitHub

**Red**

1. Add credential-gated `TestGitHubCertificationFlowRoundTripFreshBinary` using `PM_CERT_GITHUB_TOKEN`, falling back to `GITHUB_TOKEN` without ever printing the value.
2. Build `pm` into a test temp directory. Every product operation after setup runs through that one binary and a fresh project root.
3. Create a dedicated private GitHub repository and one uniquely titled issue via a bounded test harness. Register the credential from the environment, never argv or a file.
4. Create a real GitHub→warehouse ETL connection/job for `issues`; run and assert Parquet row count plus advanced checkpoint. Reopen from a separate binary process and independently query the table.
5. Create/preview/approve/execute a real reverse-ETL `create_milestone` job mapping the extracted issue `title` to milestone `title`; use its durable authorization reference for the flow action. Delete the authorization-seeding milestone before the flow.
6. Write a flow containing the real sync job followed by a connector-backed action step using `create_milestone`; run the flow through the built binary and assert both steps complete.
7. Read GitHub independently and assert the exact milestone title created by the flow. Delete it, delete/close setup resources, delete the disposable repository, and assert provider 404/absence as zero residue.
8. The test must fail if the flow action is removed, if the independent provider read-back is removed, or if the binary uses a stale/local in-memory value.

**Refusal matrix**

1. Replay an acknowledged authorization/approved item and assert the typed replay refusal plus unchanged provider milestone set.
2. Expire or withhold the flow authorization/plan and assert the typed expiry/unapproved refusal plus unchanged provider state.
3. Run with an intentionally invalid authentication value sourced only from a child-only environment variable; assert the typed authentication refusal plus unchanged provider state.
4. Preview unsafe `merge_pull_request` and repository-file delete paths only; assert no provider mutation.

**Green / evidence**

- The live command passes only when credentials are present and all provider/warehouse assertions pass. A skip is recorded as Gap 3 open.
- Commit only a sanitized evidence summary: binary digest, safe test/repository lifecycle identifiers if permitted, counts, typed error classes, action choice/rationale, and zero-residue result. No provider response body, credential, approval token, or rate scope is committed.
- Explicitly list GitHub-specific production hops such as `issueLabelTransportDefinitionFactories`, `isIssueLabelTransportConnector`, issue-label contract/config keys, and GitHub hooks. Distinguish the generic flow/action/engine path from non-generalizable adapters.

## Verification Matrix

| Area | Command | Required result |
| --- | --- | --- |
| Gap 1 focused | `go test -timeout 20m ./internal/connectors/certify -run 'TestFullWriteSweepFailsForDeliberatelyBroken' -count=1` | Both negative controls fail the sabotaged report and intact control passes. |
| Gap 2 focused | `go test -timeout 20m ./internal/app -run 'TestCertificationDeclaredTransportPair' -count=1` | Execution counters/state prove all pair stages; missing registration fails. |
| Gap 3 deterministic harness | `go test -timeout 20m ./internal/cli -run 'TestGitHubCertificationFlowRoundTrip' -count=1` | Non-live harness behavior and refusal assertions pass. |
| Gap 3 live | `PM_CERT_GITHUB_TOKEN=<environment> go test -timeout 20m ./internal/cli -run '^TestGitHubCertificationFlowRoundTripFreshBinary$' -count=1 -v` | Fresh binary completes flow, provider read-back, refusals, and zero-residue cleanup. A skip is not green evidence. |
| Changed packages | focused `go test -timeout 20m` commands for each changed package, separately | Pass. |
| Static checks | `gofmt -w` on changed Go files; `go vet` on changed packages; `git diff --check` | Pass. |
| Derived drift | one-pass repo generators required by changed files, then relevant `make verify` gates individually | Generated tree clean; drift checks pass. |
| Workflow | `go run ./cmd/agentcontractgen check`; `scripts/verify-gsd-workflow` or repository equivalent | Pass with Red/Green evidence. |
| Secret scan | bounded scan of changed/tracked artifacts for known environment value performed without printing matches | Zero matches; any match is a terminal blocker. |
| PR base | `gh-axi` API read-back of `/repos/polymetrics-ai/cli/pulls/<n>` field `.base.ref` | Exactly `integration/4015-mvp-flat-r1`. |

## Scope Guards

- Do not edit or fix #4125 or #4158.
- Do not fix a product defect exposed by live validation; write `needs-decision:` to the supervisor status file and stop that slice.
- Do not run `go test ./...`, `make verify` as one command, `/no-mistakes`, or the shared daemon.
- Do not merge the PR or push any branch other than `fm/cli-validation-proof-gaps-r1`.
- Provider repository deletion is authorized only for the uniquely named test repository created by this test run; validate owner/name and creation identity before deleting.

## Commit / Push Checkpoints

1. Planning artifacts.
2. Gap 1 RED/GREEN proof.
3. Gap 2 RED/GREEN proof.
4. Gap 3 harness/live evidence or explicit open-gap evidence.
5. Verification/review artifacts and fixes.


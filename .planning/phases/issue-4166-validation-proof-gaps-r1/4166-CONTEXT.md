# Issue 4166 Context — Certification Coverage Proof Gaps

## Task Delivery Header

- Issue: Refs #4166 — test(certification): prove the three unexercised-coverage gaps are closed
- Base branch: `integration/4015-mvp-flat-r1`
- Merges into: `integration/4015-mvp-flat-r1` → `main`
- Delivery: Pull request open against `integration/4015-mvp-flat-r1`, with focused validation green and per-gap evidence posted to issue #4166.
- Working branch: `fm/cli-validation-proof-gaps-r1`
- Task: Add validation harness coverage, negative controls, and live proof for the three audited certification gaps. Gap 1 must make a deliberately broken write action fail certification, including an action formerly reported only as blocked. Gap 2 must prove the declared transport pair is resolved and executed and that missing registration fails certification. Gap 3 must drive real GitHub → durable warehouse → real GitHub as two jobs composed into a flow through one freshly built binary, using a disposable repository and provider read-back with zero residue. Product behavior is out of scope.
- Verification: Focused Go tests for certification, app transport, CLI flow, and the credential-gated live round-trip; fresh-binary execution; provider read-back and cleanup assertions; generated-artifact drift gates; GSD verification and code review; GitHub API read-back of the opened PR base.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| A deliberately broken declared write action fails full certification | fake | A test-only connector-definition or runner seam is necessary to break an action without shipping the break or making a provider mutation. The test asserts a failed terminal report and the exact action name; the intact fixture must pass. |
| An action previously categorized as `blocked` cannot mask a broken definition | fake | The same test-only seam breaks an unpaired action such as `update_issue` or `merge_pull_request`; certification must fail rather than recording a passing blocked entry. Live execution is unsafe because these actions lack a certified create/cleanup lifecycle. |
| The declared source/destination transport is resolved and executed | fake | A deterministic local provider fixture is required to count source, stage, plan, apply, and read-back calls without modifying GitHub. Removing/unregistering the declaration must make certification fail. |
| A built binary completes GitHub → Parquet → GitHub as a composed flow | live | The test creates a disposable GitHub repository with an environment-supplied certification credential, creates real ETL and reverse-ETL jobs, composes them into a flow, runs the flow through the newly built binary, independently reopens Parquet, reads the mutation from GitHub, cleans all provider objects or deletes the repository, and asserts zero residue. |
| Replay, expired/unapproved plan, and authentication refusals make no provider write | live | Each refusal records provider state before and after, asserts the typed refusal, and independently reads GitHub to prove the mutation count or target state did not change. |
| The round trip uses connector-definition behavior and identifies exceptions | live | The proof records the definition-owned operation/action and flow job references. The report separately lists every GitHub-specific code hop observed on the production call chain so non-generalizable behavior is explicit. |

## Assertion Rule

Every acceptance test contains a negative control. A command exit status, a skipped credential-gated test, or a passing report without an observable action/transport/provider state change is not proof.

## Fixed Constraints

- Validation only: tests, fixtures, harness wiring, and proof artifacts; no new connector/product capability.
- A product defect discovered by a proof is reported as `needs-decision:` and is not fixed in this branch.
- No credential value, provider rate scope, or secret-derived material may enter stdout, stderr, argv, committed files, or issue/PR text.
- Only focused package tests run locally; CI owns the full suite.
- The live target is a dedicated disposable GitHub repository. Production repositories are out of scope.
- The primary live action is `comment_issue`: it maps the extracted issue number/title back onto that same issue, is independently readable, and is fully contained by deleting the dedicated repository. This proves a genuine loop rather than unrelated read/write halves. Label/milestone pairings remain valid alternatives but would not bind as directly to the extracted issue.
- `merge_pull_request` and destructive repository-file deletes are refusal/preview candidates only, never the primary live mutation.
- Issue #4169 owns the product correction for provider-verified 401/adjacent HTTP error classification. This validation branch asserts the currently observed typed `internal/internal_error`, zero provider writes, and no checkpoint advancement without changing product behavior.
- The live flow must contain a real sync/ETL step and a real connector-backed action step. Hand-invoked halves do not satisfy Gap 3.
- At planning time neither `PM_CERT_GITHUB_TOKEN` nor `GITHUB_TOKEN` is present. The live proof therefore remains unproven until the credential-gated command can run; a skip will be reported as an open gap, never as success.

## Required Skills and Workflow

- GSD commands: `discuss-phase` → `plan-phase --tdd` → `execute-phase` → `verify-work` → `code-review`, generated by `scripts/gsd prompt` and executed inline because the canonical single-worker contract forbids workflow-role spawning.
- Go skills: `golang-how-to`, `golang-testing`, `golang-cli`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-error-handling`, `golang-security`, and `golang-safety`.
- No CLI command, flag, help topic, manual, or website behavior is intended to change. CLI help/docs/website parity is therefore expected to be not applicable unless implementation proves otherwise.

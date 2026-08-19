# GitHub live read coverage r1

## Task Delivery Header

- Issue: Refs #4015 — Production MVP; GitHub read-side certification coverage.
- Base branch: `integration/4015-mvp-flat-r1`.
- Merges into: `integration/4015-mvp-flat-r1` → `main`.
- Delivery: A PR is open against the stated base, its base is read back from the GitHub API, and the committed changes have local verification evidence.
- Working branch: `fm/cli-github-read-coverage-r1`.
- Task: Classify GitHub's complete read surface, add a substantially larger set of safe, live direct-read certification candidates, prove a post-schema read break fails certification, and state the truthful boundary to users. The existing binary candidate is separately classified as not applicable for a live-pass result in this wave.
- Verification: Exhaustive scope accounting; targeted certification tests; compiled `pm` live run using only the disposable certification identity; intentional scratch failure followed by restore; generated-file and repository verification gates; PR base API read-back.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| Every declared GitHub read command and stream is classified | live | The committed SCOPE ledger has separately reconciled 1,571-command and 37-stream totals; no record is unbucketed. |
| Safe read candidates execute against GitHub | live | Certification invokes the compiled command with the disposable credential and validates produced values, result kind, and secret-scan outcome. |
| Non-executed reads cannot inflate pass coverage | live | Every exclusion has `blocked`/`not_applicable`/other non-pass status and a concrete fixture or identity reason in the scope record. |
| Read certification fails when a compiled candidate is broken | live | A scratch post-schema mismatch produces a failing certification result; the manifest is restored before commit. |
| User-facing boundary is truthful | live | The surfaced summary reports certified candidate coverage and excluded boundary; it never claims all 1,571 commands or 37 streams executed. |

## Lifecycle and scope

- GSD route: `scripts/gsd doctor`; `sources` and generated prompts for `discuss-phase`, `plan-phase --tdd`, `execute-phase`, `verify-work`, and `code-review`; `go run ./cmd/agentcontractgen check` all succeeded before production edits.
- Inline/manual fallback: the task supplies a complete autonomous brief and this runner has no compatible isolated GSD-agent worktree facility. The single worker will perform the generated workflow stages inline and retain their PLAN, TDD ledger, verification, and review artifacts.
- Required skills loaded: `golang-how-to`, `golang-testing`, `golang-error-handling`, `golang-security`, `golang-safety`, `golang-design-patterns`, `golang-structs-interfaces`, `golang-context`, `golang-concurrency`, `golang-documentation`, and `golang-cli`.
- Decision: Firstmate authorized one direct PR against `integration/4015-mvp-flat-r1` combining GitHub candidates with the provider-neutral #4191 output-assertion foundation. It is not stacked. `allStagesPassed` is explicitly owned by the PostgreSQL certification lane and must not be modified; non-executed GitHub rows use a concrete `blocked:` reason rather than a `skipped:` roll-up escape hatch.
- Ownership guard: GitHub connector data, its certification evidence, and the provider-neutral #4191 schema/engine/direct-read assertion contract only. PostgreSQL work, write actions, broker, MCP, UI, and any `allStagesPassed` change are out of scope.
- CLI documentation parity: this adds `connectors certify --direct-read-only` and a text-report boundary. The runtime manual source in `internal/cli/docs.go`, CLI manual source `docs/cli/connectors.md`, and website CLI reference `website/content/docs/cli-reference.mdx` are updated; `make docs-check` and runtime help are required.

## Ordered delivery slices

0. Fix the credential-rendering defect found during task setup: persisted/report text must redact sensitive assignments and reject any residual unsafe value (fail closed). Prove the red case with synthetic values only, then keep real credential retrieval restricted to command substitution with no logging.
1. Inspect the current certification profile, candidate evaluator, generated CLI surface, and write-wave SCOPE model. Build exhaustive command and stream classifications before changing candidates. Escalate if safe live coverage is unexpectedly small.
2. Record the classification and create candidate data for each safely repeatable fixture read. Include expected produced values and no-secret assertions; serial execution remains the existing runner's responsibility. If the declaration-owned output assertion is absent, open/escalate the required foundation split before changing candidate data.
3. Establish red evidence by making one post-schema candidate expectation impossible in an uncommitted scratch edit and demonstrate that certification fails; restore it. Add the smallest deterministic test coverage needed for the data contract.
4. Build the CLI and execute live certification using the documented disposable identity only. Confirm rate-limit handling/resume behavior, review the result ledger, and ensure non-passes do not appear in pass roll-ups.
5. Run repository gates, review changed paths, prepare the direct-PR body with happy/bad/edge cases and GSD evidence, push, create the PR, and API-read its base.

## Commit checkpoints

- Plan/classification evidence (after scope totals reconcile).
- Green candidate/configuration slice after targeted tests and live proof.
- Verification/review fixes, if any.

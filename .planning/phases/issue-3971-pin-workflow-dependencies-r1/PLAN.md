# Plan: Pin build dependencies for the GitHub parity unblock

## Delivery identity

- Parent issue supplied by the task: `#3971` (GitHub parity).
- Security alert in scope: code-scanning alert `#135`, `PinnedDependenciesID` / OpenSSF Scorecard `Pinned-Dependencies`.
- Branch: `fm/cli-pin-workflow-dependencies-r1`.
- Child issue: [#3986](https://github.com/polymetrics-ai/cli/issues/3986), created after local work and attached to parent #3971 after duplicate searches returned no equivalent issue.

## GSD / TDD path

- `scripts/gsd doctor` passed.
- Resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review` with `scripts/gsd sources`; `go run ./cmd/agentcontractgen check` passed.
- Generated the required adapter prompts for `discuss-phase 3971 --auto`, `plan-phase 3971 --tdd --auto`, `execute-phase 3971 --interactive`, `verify-work 3971 --auto`, and explicit-file `code-review`.
- Inline/manual GSD fallback: `gsd-sdk query init.phase-op 3971` reported `phase_found: false`. The supplied GitHub parent is not a registered roadmap phase, and the task requires local work before creating the child issue while the GitHub API quota is exhausted. This phase therefore records the required discuss → plan → execute → verify → review evidence inline without changing `.planning/ROADMAP.md` or `.planning/STATE.md`.
- Required skills loaded: `github-issue-first-delivery`, `no-mistakes`, `gsd-discuss-phase`, `gsd-plan-phase`, `gsd-execute-phase`, `gsd-verify-work`, `gsd-code-review`, `golang-how-to`, `golang-continuous-integration`, `golang-security`, `golang-safety`, `golang-lint`, and `golang-testing`.

## Research and threat model

Alert #135 is a medium-severity Scorecard finding at `.github/workflows/verify.yml:86`; its remediation requires full Git commit IDs for workflow actions and hashes for Docker build inputs. The attacker-controlled boundary is an upstream mutable tag or image tag. Moving one after review could inject different build code without a repository diff.

The local audit found mutable action refs in these workflow files:

- `.github/workflows/scorecard.yml`
- `.github/workflows/security.yml`
- `.github/workflows/connector-boundary.yml`
- `.github/workflows/pr-issue-guard.yml`
- `.github/workflows/website-data.yml`
- `.github/workflows/website.yml`
- `.github/workflows/claude-review.yml`
- `.github/workflows/gsd-workflow.yml`
- `.github/workflows/verify.yml`
- `.github/workflows/release.yml`

The only Dockerfile is `website/Dockerfile` and the only literal workflow image is the PostgreSQL service in `website.yml`. Both use mutable image tags. The detailed current-tag-to-immutable-resolution record is in `TDD-LEDGER.md`.

## Scope and exclusions

In scope:

1. Replace every external GitHub Actions ref in the audited workflows with its currently resolved 40-character commit SHA and preserve the original version as a trailing comment.
2. Pin the Node Dockerfile base image and the literal PostgreSQL CI service image to their current multi-architecture manifest-list digests.
3. Add a local regression gate that rejects mutable workflow action refs or literal build images and verifies workflow YAML can still be parsed.
4. Run targeted static, YAML, and repository verification gates; record alert and inventory status honestly.

Out of scope:

- Changing action or image versions, permissions, triggers, jobs, release behavior, or application code.
- Fixing the other code-scanning alert classes.
- Dismissing security alerts, changing repository security settings, merging, or accessing secrets.

## TDD slices

1. **Red:** add a focused build-dependency static gate and run it against the current mutable workflow/image references. It must fail by naming mutable refs.
2. **Green:** substitute only the live tag/manifest resolutions captured before editing; rerun the focused gate until it passes.
3. **Refactor/hardening:** wire the focused gate into `release-workflow-check`, parse every workflow file, and review the diff for version drift or accidental behavior changes.
4. **Verification:** run the focused gate, the existing release-workflow check, YAML parsing, `git diff --check`, GSD contract validation, and the relevant non-test `make verify` gates individually.

## Acceptance criteria

- Every audited external action uses a 40-character SHA with a trailing version comment.
- Every audited literal build image uses a `sha256` manifest digest while retaining its readable tag.
- No pinned SHA/digest differs from the value resolved from its original mutable ref at research time.
- The regression gate fails on the recorded mutable baseline and passes on the production workflows.
- Alert #135 and the remaining four open alerts are checked and reported after the GitHub API quota permits a final REST read; no closure is claimed without evidence.

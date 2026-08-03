# Issue 3674 — Issueguard Unvalidated-Cloud-Checkpoint Issue Links

## GSD setup

- Issue: https://github.com/polymetrics-ai/cli/issues/3674
- Branch: `fix/3674-issueguard-checkpoint-issue-links`
- GSD preflight: `scripts/gsd doctor` passed on 2026-08-03 (node v24.13.1, 69 commands registered).
- GSD prompt path: `scripts/gsd prompt programming-loop init --phase issue-3674-issueguard-checkpoint-links --dry-run`
  was attempted first and the repo-local command registry returned
  `scripts/gsd: unknown GSD command: programming-loop`. `scripts/gsd` was therefore not used
  interactively to drive this slice, and this phase is an explicit manual-GSD fallback recorded per
  `.agents/agentic-delivery/references/gsd-pi-adapter.md`.
- Planning prompt fallback: `scripts/gsd prompt plan-phase issue-3674-issueguard-checkpoint-links --skip-research`
  generated a 142-line prompt that was applied inline as the planning fallback.
- Orchestration decision, plan cycle: `local_critical_path` — this is one focused shared-code slice in an
  already isolated worktree, so no mutating subagents were spawned.

## Required skills loaded

- `gsd-core`
- `golang-how-to`
- `golang-testing`
- `golang-error-handling`
- `golang-design-patterns`
- `golang-safety`
- `golang-security`
- `golang-documentation`

## Scope boundaries

### In scope — code

1. `internal/coordination/issueguard/guard.go` — recognize the
   `## Unvalidated cloud checkpoint — do not merge yet` /
   `## Canonical issue links preserved from the task record` PR-body pattern and extract the bare
   `https://github.com/<owner>/<repo>/issues/<n>` links listed under that section as non-closing refs.
2. `internal/coordination/issueguard/guard_test.go` — positive coverage (LF, CRLF, GitHub-indented
   bodies) plus negative coverage pinning both gates.

Code scope is exactly these two files. No other production or behavior-carrying file changes.

### In scope — GSD planning evidence

3. `.planning/phases/issue-3674-issueguard-checkpoint-links/PLAN.md`, `TDD-LEDGER.md`,
   `VERIFICATION.md`, and `RUN-STATE.json` — added alongside the code per repo convention because
   `internal/**` changed. `scripts/verify-gsd-workflow` fails the `gsd-workflow-evidence` gate
   without changed `.planning/**` evidence. These are documentation artifacts: they carry no
   behavior and are not part of the code scope above.

### Out of scope

- No files change beyond the two code files and this phase's four planning-evidence files; the
  shipped diff is exactly those six paths.
- No loosening of the general require-linked-issue rule for non-checkpoint PR bodies.
- No connector definition, CLI surface, docs, or website changes; issueguard is a shared PR-body
  validator consumed by `cmd/prissueguard` and `.github/workflows/pr-issue-guard.yml`, not a `pm`
  user surface, so CLI/help/docs/website parity is not applicable.
- No shared agent-contract prose changes.
  `.agents/agentic-delivery/contracts/issue-agent-contract.md` owns the list of PR bodies that
  satisfy the guard, and its required-workflow step 14 is knowingly left out of sync by this slice:
  it still enumerates only the `Closes #N` / `Refs #N` route and the no-mistakes delivery-record
  route, with no entry for the checkpoint route this change adds.
  Updating it is a deliberate follow-up left to firstmate to file as a separate issue, per the
  explicit decision that widening this PR into the shared contract would repeat the scope creep this
  foundation PR was extracted to avoid. Precedent for that follow-up shape: commit `235c7b22f`
  updated the same step 14 (and its `Output requirements` restatement) when the no-mistakes
  delivery-record route was added.
- No change to `completedTaskPattern`'s trailing-noun tightness. One known wording variant remains
  blocked by design: PR #3578 (marketo, wave05) carries the identical checkpoint headings but
  phrases its sentence as `...an unvalidated cloud checkpoint for the completed committed
  cli-marketo-parity-wave05-r1 connector parity implementation`, ending in `implementation` rather
  than `task`, so `completedTaskPattern`'s required `\btask\b` does not match. This is deliberately
  recorded as a follow-up rather than broadened now, because loosening the shared
  require-linked-issue gate needs separate, evidenced justification; the control sweep over 63
  non-checkpoint PRs showed zero verdict changes under the current tight pattern, and that margin is
  what a broader pattern would spend.

## Downstream impact

Measured by building `cmd/prissueguard` at base `5d61794f7` and at this branch head, then running both
binaries against every open PR body fetched live from GitHub. Of 99 open PRs, 36 flip
`blocked` -> `ok` and 63 are unchanged.

| PR | Base | Head | Note |
|---|---|---|---|
| #3540 airtable | blocked | ok (8 linked issues) | Unblocked by this change. |
| #3543 aws-cloudtrail | blocked | ok (8 linked issues) | Unblocked by this change. |
| #3544–#3577 (34 wave PRs) | blocked | ok | Same checkpoint pattern; contiguous `stage unvalidated parity checkpoint` range. |
| #3541 amazon-sqs | ok | ok | Never blocked for this reason; body carries an explicit `Refs #3134`. Verdict unchanged by this patch. |
| #3542 ashby | ok | ok | Never blocked for this reason; body carries an explicit `Refs #3207`. Verdict unchanged by this patch. |
| #3578 marketo | blocked | blocked | Deliberately still blocked; see the wording variant recorded under "Out of scope". |
| #3676 nativeset foundation | blocked | blocked | **Not fixed by this diff.** Its body carries no issue-reference keyword at all — no checkpoint headings, no `Closes`/`Refs`/parent wording — so it stays blocked at both base and head. It needs a separate PR-body fix. |

Earlier planning wording claimed this PR unblocks four PRs (#3540, #3543, #3542, #3676). That was an
unverified inference and is corrected above: only #3540 and #3543 of the named set flip, alongside the
34 other wave PRs in the #3544–#3577 range.

## Implementation plan

### Slice 1 — Red tests for the checkpoint body pattern

- Add `TestValidatePRAcceptsUnvalidatedCheckpointCanonicalIssueLinks` with LF, CRLF, and
  GitHub-indented body variants built from a real wave03 checkpoint body.
- Confirm it fails against the pre-change validator, which reports
  `PR body must reference an issue`.

### Slice 2 — Gated recognition in the validator

- Add `unvalidatedCheckpointHeadingPattern`, `canonicalIssueSectionPattern`,
  `markdownIssueURLPattern`, `completedTaskPattern`, and `markdownH2StartPattern`.
- Add `addCheckpointIssueRefs` to `ExtractIssueRefs`, gated on the checkpoint heading appearing
  before the canonical-issue-links heading and on non-negated
  `unvalidated cloud checkpoint for the completed ... task` wording between the two headings.
- Bound extraction to the canonical-issue-links section by stopping at the next H2 heading.
- Record extracted refs with keyword `checkpoint` and non-closing semantics so they can never
  downgrade an existing closing ref.

### Slice 3 — Negative coverage for both gates

- `TestValidatePRRejectsCanonicalIssueSectionWithoutCompletedTaskWording` pins the wording gate.
- `TestValidatePRRejectsCanonicalIssueSectionWithoutCheckpointHeading` pins the heading gate, so
  deleting either gate turns the suite red.

### Slice 4 — Shared negation filter and verification

- Extract `hasNonNegatedMatch(re *regexp.Regexp, text string) bool` and route both
  `hasCompletedTaskWording` and `hasExplicitIssueWording` through it; behavior-neutral.
- Run the focused verification commands and record actual outcomes in `VERIFICATION.md`.

## CLI/help/docs parity checklist

- Not applicable. `internal/coordination/issueguard` backs the `pr-issue-guard` GitHub Actions check
  via `cmd/prissueguard`; no `pm` command, flag, help topic, `docs/cli/**` page, or website surface
  changes in this slice.

## Verification checklist

- `gofmt -l internal/coordination/issueguard`
- `go vet ./internal/coordination/issueguard/...`
- `go test ./internal/coordination/issueguard/... -count=1`
- `go build ./cmd/pm`

## Commit checkpoint plan

One implementation commit for the gated recognition plus tests, then one review-fix commit carrying
the negative heading-gate test, the shared negation helper, and this planning evidence. Push and PR
state stay draft; merge is human-gated for firstmate.

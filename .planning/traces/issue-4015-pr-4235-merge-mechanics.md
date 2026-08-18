# PR #4235 Merge Mechanics

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `main`
- Merges into: `fm/cli-mvp-merge-mechanics-r1 → main`
- Delivery: Make PR #4235 protection-compliant without merging it, or leave a precise, evidence-backed blocker and a lossless clean-history candidate on the assigned branch for the captain.
- Working branch: `fm/cli-mvp-merge-mechanics-r1`
- Task: Evaluate rebasing the integration branch, squash merging while strict status checks are enabled, and repository-permitted alternatives; preserve the integration tree, 998 certification evidence files, and the GitHub command surface at 1571 declared / 1546 implemented; identify the required follow-up for `fm/cli-sync-e2e-pgpg-r1`.
- Verification: Compare Git tree IDs and diffs before and after history reconstruction; count evidence files from Git; derive GitHub declared/implemented counts from `internal/connectors/defs/github/cli_surface.json`; inspect PR checks and merge state with `gh-axi`; verify ancestry against `origin/main`; run repository checks appropriate to a planning-only mechanics report.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| No integration work is lost | live | Git tree comparison plus the 998 evidence-file and 1571/1546 command-surface invariants stay equal before and after reconstruction. |
| The selected route satisfies branch protection | live | GitHub reports PR #4235 up to date and all required checks pass on its current head, or the report names the exact permission/dispatch boundary preventing that state. |
| Squash behavior under strict checks is decided from evidence | live | GitHub's current protected-branch documentation and the live PR merge state establish whether squash can bypass a behind head. |
| The dependent branch follow-up is explicit | live | Git ancestry proves `fm/cli-sync-e2e-pgpg-r1` descends from the pre-rewrite integration head and the report gives its rebase command shape. |

## Assertion Rule

Each live check asserts an observable tree, count, ancestry, or GitHub state; command success alone is not treated as proof.

## Mechanics Record

### Live baseline

- `origin/main`: `b3df1c1c78e9057ac3129ee7aacfa4dcdc8d62e8`; its parent is the other missing commit, `2cd8be53cdece153cfb94270571fd82df3a82a63`.
- `origin/integration/4015-mvp-flat-r1`: `8bab2aa293275cd3d1e1798ea2e39b0158291c08`.
- Integration tree: `aa02f9f4e8ee639c758daa83e100b986f5f2c7f8`.
- Integration-only history from the common ancestor contains 132 non-merge commits and 5 merge commits.
- Certification evidence: 998 files under `internal/connectors/certifications/evidence`.
- GitHub surface, derived directly from `internal/connectors/defs/github/cli_surface.json`: 1571 declared, 1546 implemented.
- `git merge-tree --write-tree origin/main origin/integration/4015-mvp-flat-r1` returned the same `aa02f9...` tree. The integration tree already contains the two main commits' content; only ancestry is behind.
- The live PR was draft when inspected. `gh-axi pr checks 4235` reported 33 passed, 1 non-required Snyk failure, 10 skipped, and 2 pending `verify` entries; the required checks named by the branch rule were otherwise passing.

### Route evaluation

1. **Replay rebase: rejected after local proof.** A default replay stopped immediately on an add/add conflict. A second controlled replay using `-X theirs` completed after resolving the one expected generated-matrix deletion, but its final tree was `f2d9dad846f65a25a258c2c5cc780b26f7f5f15b`, not the integration tree. Seven files drifted, with 118 net lines lost or changed. The evidence and surface counts happened to remain 998 and 1571/1546, but the tree mismatch fails the stronger no-loss invariant. Commit `f702b9c17c95c9a2b226060302fe504e185019c8` was abandoned and was not pushed.
2. **Squash-merge PR #4235 as-is: blocked while behind.** Squash would satisfy the linear-history rule, but it does not bypass strict required status checks. GitHub defines strict checks as requiring the topic branch to be up to date with its base before merging, independently of the selected squash/rebase merge method. See [About protected branches](https://docs.github.com/en/repositories/configuring-branches-and-merges-in-your-repository/managing-protected-branches/about-protected-branches) and [Troubleshooting required status checks](https://docs.github.com/en/pull-requests/how-tos/merge-and-close-pull-requests/troubleshooting-required-status-checks). Therefore `BEHIND` blocks a squash merge definitively.
3. **Exact-tree squash reconstruction: selected.** Starting at current `origin/main`, `git merge --squash origin/integration/4015-mvp-flat-r1` produced commit `556edb49d260bb854d868657cf610a3ffa59a4e2`. It has exactly one parent (`b3df1c1...`), zero merge commits, and tree `aa02f9...`, byte-for-byte identical to the old integration head. This is both up to date and linear without conflict-resolution judgment.
4. **Replacement main PR: technically possible but rejected as the primary route.** The clean candidate could open a second main-targeted PR, but that would not make #4235 mergeable and would split its canonical checks/review record. A PR into integration also cannot move the base ref to the candidate's ancestry with a squash/rebase merge; a merge commit is the already-refused path. An admin bypass is expressly forbidden.

### No-loss proof

| Invariant | Before: `8bab2aa...` | Candidate: `556edb49...` |
| --- | ---: | ---: |
| Tree | `aa02f9f4e8ee639c758daa83e100b986f5f2c7f8` | `aa02f9f4e8ee639c758daa83e100b986f5f2c7f8` |
| `git diff --exit-code` | n/a | exit 0 against before |
| Evidence files | 998 | 998 |
| GitHub declared | 1571 | 1571 |
| GitHub implemented | 1546 | 1546 |
| Merge commits after `origin/main` | n/a | 0 |

The mechanics report is a second commit on `fm/cli-mvp-merge-mechanics-r1`; the exact integration replacement is intentionally its parent, `556edb49...`.

### Exact remaining blocker and captain decision

The task dispatch permits pushes only to `fm/cli-mvp-merge-mechanics-r1`, so this worker cannot update the head ref of #4235. The one required captain action is to perform or expressly authorize this lease-guarded ref update after fetching the published candidate:

```sh
git push --force-with-lease=refs/heads/integration/4015-mvp-flat-r1:8bab2aa293275cd3d1e1798ea2e39b0158291c08 \
  origin 556edb49d260bb854d868657cf610a3ffa59a4e2:refs/heads/integration/4015-mvp-flat-r1
```

Never substitute a bare `--force`. If the lease fails, fetch and re-run all no-loss proofs against the new integration head instead of overwriting it. After the ref update, required checks must rerun on the new head; then mark #4235 ready, confirm the required checks are current and passing, and leave the actual merge to firstmate/captain.

### Dependent branch follow-up

`fm/cli-sync-e2e-pgpg-r1` is a live local branch in another worktree. The old integration head is its ancestor and it is one commit ahead. After integration moves, its owner must rebase that one commit from the old base onto the rewritten integration branch:

```sh
git fetch origin integration/4015-mvp-flat-r1
git rebase --onto origin/integration/4015-mvp-flat-r1 \
  8bab2aa293275cd3d1e1798ea2e39b0158291c08 fm/cli-sync-e2e-pgpg-r1
```

If that branch is later published, its owner must use its own force-with-lease after verifying its tree; it was not present under that name on `origin` during this inspection.

### Verification run

- `go test -timeout 20m ./cmd/connectorgen -run '^TestCertificationSweepForGitHubIsSurfaceDerivedAndExhaustive$' -count=1` — pass.
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` — pass; exercises all 1546 implemented commands through runtime preflight.
- `go run ./cmd/connectorgen surface-sync --check` — pass; 552 connectors scanned, zero drift.
- `go run ./cmd/agentcontractgen check` — pass; canonical contract and projections current.
- Full `go test ./...` / `make verify` was not rerun locally because this task changes no product tree and the repository explicitly directs per-command agents not to run the 550+ connector suite as one command. The candidate tree is byte-identical to the already-checked integration head; GitHub must rerun all required checks on the rewritten SHA before #4235 is ready.

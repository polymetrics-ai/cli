# PR #4235 Merge Mechanics

## Task Delivery Header

- Issue: Refs #4015 — Production MVP certification
- Base branch: `main`
- Merges into: `release/0.2.0-mvp → main`
- Delivery: PR #4250 open from the lossless linear release branch to `main`, with branch protection left intact and the merge reserved for firstmate/captain.
- Working branch: `fm/cli-mvp-merge-mechanics-r1`
- Task: Evaluate rebasing the integration branch, squash merging while strict status checks are enabled, and repository-permitted alternatives; preserve the integration tree, 998 certification evidence files, and the GitHub command surface at 1571 declared / 1546 implemented; identify the required follow-up for `fm/cli-sync-e2e-pgpg-r1`.
- Verification: Compare Git tree IDs and diffs before and after history reconstruction; count evidence files from Git; derive GitHub declared/implemented counts from `internal/connectors/defs/github/cli_surface.json`; inspect PR checks and merge state with `gh-axi`; verify ancestry against `origin/main`; run repository checks appropriate to a planning-only mechanics report.

## Evidence Table

| Acceptance criterion | Evidence | Observable assertion or fake reason |
| --- | --- | --- |
| No integration work is lost | live | Git tree comparison plus the 998 evidence-file and 1571/1546 command-surface invariants stay equal before and after reconstruction. |
| The selected route satisfies branch protection | live | PR #4250 targets `main` from a non-draft release branch that contains `main`, has zero merge commits, and preserves the certified integration content. |
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
4. **Replacement release-branch PR: selected after integration protection closed the rewrite route.** The captain authorized the exact-tree ref rewrite, but GitHub rejected the force-with-lease because `integration/4015-mvp-flat-r1` itself disallows force pushes. Publishing the already-verified candidate as a new release branch therefore became the only route that preserved protection, history linearity, and every certified byte without an admin bypass.

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

### Final resolution: release branch and PR #4250

The authorized force-with-lease was rejected by GitHub's protected-branch hook with `Cannot force-push to this branch`. Under the current repository protection, rewriting `integration/4015-mvp-flat-r1` is permanently unavailable; changing that policy or using an admin bypass was outside scope and unnecessary.

Firstmate published the verified branch head `0cf70c2f8043546a68ea41fcf726220ccc11a4b5` as `release/0.2.0-mvp` and opened [PR #4250](https://github.com/polymetrics-ai/cli/pull/4250) to `main`. Independent verification and this worker's read-back agree:

- `origin/main` is an ancestor of the release head.
- The release range contains zero merge commits.
- It contains 998 certification evidence files.
- The GitHub surface remains 1571 declared / 1546 implemented.
- Compared with `integration/4015-mvp-flat-r1`, the release branch only adds this trace document; it removes or changes nothing.
- `gh-axi pr list --state open --base main --head release/0.2.0-mvp --fields url,body` returns exactly PR #4250, proving the intended head/base pairing through the wrapper's filtered API result.
- At the latest read-back, the non-draft PR had 7 passing, 1 non-required Snyk failure, 4 skipped, and 8 pending checks. Required checks must finish on this head before the captain merges it.

No policy change, force push, admin bypass, product edit, version edit, or PR merge was performed by this worker.

### Runbook for future protected integration releases

When a protected integration branch is both behind `main` and non-linear, do not plan on rewriting it. Instead:

1. Create a new release branch at current `main`.
2. Squash-apply the integration tree.
3. Require an exact tree comparison against the integration head, plus project-specific evidence and parity invariants.
4. Add only the delivery trace needed to explain the release mechanics.
5. Publish the new release branch and open a main-targeted PR.
6. Let required checks rerun on that exact head and keep the final merge human-gated.

This route works with existing protection and produces linear, up-to-date history without risking partial replay conflict resolution.

### Dependent branch follow-up

`fm/cli-sync-e2e-pgpg-r1` remains based on the unchanged integration head. Because integration was not rewritten, this release operation did **not** move its base and it requires no rebase as a consequence of PR #4250. If its work must join the 0.2.0 release line later, its owner should replay its one additional commit onto the current release/main lineage in a separate, verified branch rather than attempting to rewrite protected integration.

### Verification run

- `go test -timeout 20m ./cmd/connectorgen -run '^TestCertificationSweepForGitHubIsSurfaceDerivedAndExhaustive$' -count=1` — pass.
- `go test -timeout 20m ./internal/connectors/commandrunner -run '^TestEveryImplementedCommandPassesRuntimePreflight$' -count=1` — pass; exercises all 1546 implemented commands through runtime preflight.
- `go run ./cmd/connectorgen surface-sync --check` — pass; 552 connectors scanned, zero drift.
- `go run ./cmd/agentcontractgen check` — pass; canonical contract and projections current.
- `gh-axi pr list --state open --base main --head release/0.2.0-mvp --fields url,body` — pass; exactly PR #4250 returned.
- Release head `0cf70c2f8`: `main` is an ancestor, merge-commit count is 0, evidence count is 998, GitHub counts are 1571/1546, and `git diff --name-status origin/integration/4015-mvp-flat-r1 0cf70c2f8` reports only this added trace.
- Full `go test ./...` / `make verify` was not rerun locally because this task changes no product tree and the repository explicitly directs per-command agents not to run the 550+ connector suite as one command. The candidate's product tree is byte-identical to the already-checked integration head; GitHub is rerunning all required checks on the release SHA before PR #4250 can merge.

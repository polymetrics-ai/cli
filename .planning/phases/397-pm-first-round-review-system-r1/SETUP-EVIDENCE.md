# Issue #397 PM First-Round Review System Setup Evidence

Captured before production edits on 2026-07-23.

## Isolation and scratch inventory

```text
pwd
/Users/karthiksivadas/.treehouse/cli-83d592/2/cli

git rev-parse --git-common-dir
/Users/karthiksivadas/karthik-agent-workspace/projects/cli/.git

git status --short --branch --untracked-files=all
## HEAD (no branch)

git log --oneline --decorate -3
873cd7b25 (HEAD, origin/main, origin/HEAD, main) feat(gong): complete CLI parity surface (#232)
74ab381eb chore(project): restore PM brand and define license boundaries (#447)
c3e624486 feat(website): ship blog annotations and reader discussions (#346)

git stash list --format='%gd %H %s'
<empty>

git diff --stat
git diff --cached --stat
<empty>
```

The working directory is the recorded disposable worktree, not
`/Users/karthiksivadas/karthik-agent-workspace/projects/cli`. No browser/research artifact or prior
scratch change was carried.

## Parent fetch and exact base

```text
git fetch origin feat/cli-architecture-v2
remote head: 0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20
git merge-base --is-ancestor 0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20 origin/feat/cli-architecture-v2
exit 0

git show -s --format='%H %P %s' origin/feat/cli-architecture-v2
0f8c964ba9cfbe1b1eec8e7998eacf4158ef0e20 21d195aff0c7bd60b3bf54f14b1ce165cec9e03f chore(orchestration): sync CLI v2 parent and canonicalize PM route (#495)
```

The worktree returned to detached `origin/feat/cli-architecture-v2`, was clean, then created
`chore/pm-first-round-review-system-r1` at that exact commit. No reset, rebase, amend, force push,
or parent-branch mutation occurred.

## Parent and ownership state

- Parent issue #397: open.
- Parent PR #438: open draft from `feat/cli-architecture-v2` to `main`; human-only merge.
- PR #495 squash is the current remote parent tip.
- Full PR #493 ownership diff was derived from merge base `21d195aff...` to inspected head
  `e21e5633...`; all 18 paths are in the phase forbidden-path set.

## GSD discovery

```text
scripts/gsd doctor
# passed
scripts/gsd list
# 69 commands
scripts/gsd prompt programming-loop init --phase 397-pm-first-round-review-system-r1 --dry-run
# error: unknown GSD command or prompt: programming-loop
scripts/gsd prompt plan-phase 397-pm-first-round-review-system-r1 --skip-research
# generated official plan-phase prompt; executed through Pi tools
```

The current canonical PM route says missing `programming-loop` activates `/pm-orchestrate` as the
lifecycle owner. This phase follows that owner, not a generic/manual fallback.

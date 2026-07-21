# Issue 476 Verification

Status: `pass`

Correction 2 verification covers implementation/refactor checkpoint
`6a22aa789095da67c5b10f51476de41d3f5643ca` on ready stacked PR
https://github.com/polymetrics-ai/cli/pull/484. The reviewed predecessor
`d5181cd25d108e7748309216b14d91313f112fcd` is superseded.

| Gate | Result | Evidence |
|---|---|---|
| focused issue tests | pass | 29 tests passed, 0 failed; genuine temporary local repositories and bare remotes |
| full Shepherd tests | pass | 166 tests passed, 0 failed in 95.9s using `--test-concurrency=1` |
| strict TypeScript / Pi 0.80.6 | pass | both production adapters and imports passed strict no-emit TypeScript using the cached Pi 0.80.6 Node type surface |
| Pi extension discovery | pass | documented offline Pi 0.80.6 RPC `get_commands` returned `true` for `pm-shepherd` from `extension` |
| diff hygiene | pass | `git diff --check` and exact-range `git diff --check e659d6f1...HEAD` |
| scope hygiene | pass | exact base range contains only issue-owned adapters, matching tests/fixture, and phase 476 artifacts |
| pushed implementation ref | pass | local and origin branch equaled `6a22aa789095da67c5b10f51476de41d3f5643ca` before this evidence update; PR #484 tracks that branch |
| stacked ready PR | pass | PR #484 is open against `feat/471-pi-agent-session-shepherd` |

## Correction contract evidence

- `GitAdapter.inspect` now emits exactly the v1 repository/worktree identities used by
  `resolveCanonicalGitWorktree`. Tests cover coordinator and linked worktrees plus local, HTTPS,
  SSH, SCP, file, and absent-origin normalization.
- Immutable mode-0600 claim and worktree-binding records persist the PR base, exact base SHA,
  canonical allowed scopes, repository identity, remote identity, and worktree identity. Handoff
  rereads those records and rejects both live-object and persisted-record tampering.
- Each writable workspace owns an append-only fenced `FileStateStore` lease. Same-owner concurrent
  retries yield one winner, release is idempotent at the workspace API, stale `start` fails with
  resume guidance, and explicit same-request `resume` fences the dead owner.
- A WorkspaceAdapter-private issuer and GitAdapter WeakMap capability fence every fetch, worktree
  creation, add, commit, and push mutation. Release rejects later admission and drains accepted
  operations before releasing the underlying lease; released/recovered contexts cannot mutate.
- Handoff diffs are unfiltered and rename-disabled before immutable-scope checks. The returned
  scope is the complete sorted canonical union with dirty evidence; literal backslashes are
  rejected rather than rewritten into a permitted directory.
- Effective fetch and push endpoint identities are persisted in schema-v3 claims and revalidated.
  Late `pushurl`, divergent fetch/push endpoints, and chained `insteadOf`/`pushInsteadOf` rules are
  rejected before push; local alternate remotes remain ref- and object-free.

## Timing warning disposition

The full-suite warning came from running real Git subprocess fixtures concurrently with
`sdk-runner.test.ts`, which intentionally contains narrow wall-clock deadline assertions. The
authoritative full Shepherd command serializes test files with `--test-concurrency=1`; all 166 tests
pass and no production or test timeout was widened.

## Verification policy

Per parent direction, issue #476 runs only focused tests, the full Shepherd suite, strict
TypeScript, offline Pi RPC, and diff/scope checks. No Go, connector, certification,
runtime-service, or `make verify` gate was run during this correction cycle.

# Issue 476 Verification

Status: `in_progress`

Correction 4 is active from reviewed head `1fe994a68ec3286ee69f1be4fadf71416d601257`.
Production remained unchanged through a genuine test-only RED for the missing pre-transfer union
of committed, staged, tracked-dirty, and untracked paths; the GREEN correction is pending.
Correction-3 results below remain historical evidence for their exact head but do not satisfy
correction 4.

The genuine correction-4 RED is `1ed10ad6f9e7893cc4a921bc5f1f6fbb848c61f1`: the focused
serialized run reported 42 tests, 36 passes, and six expected failures. All five dirty-state
variants created the remote issue ref; production remained unchanged from reviewed head
`1fe994a68ec3286ee69f1be4fadf71416d601257`.

Correction 3 supersedes reviewed head `9728f9ed12e8e545eabd8b9b1b8028af80150427` with GREEN
checkpoint `db6bdd675aaced17f0d709b08a647258dfb87f15` and refactor checkpoint
`f7cb0cab0d2fb0c2ef01edc516bd3cdf950b5113` on ready stacked PR
https://github.com/polymetrics-ai/cli/pull/484.

The genuine correction-3 RED is `fa607d31`: focused 36 tests yielded 26 passes and ten expected
failures, with no production-adapter changes relative to the correction-2 implementation.

| Correction 3 gate | Result | Evidence |
|---|---|---|
| focused adapter tests | pass | 36 passed, 0 failed in 58.2s with serialized test files |
| complete Shepherd tests | pass | 173 passed, 0 failed in 107.5s with `--test-concurrency=1` |
| strict TypeScript / Pi 0.80.6 | pass | adapters and matching tests passed strict no-emit TypeScript against cached Pi 0.80.6 Node types |
| offline Pi discovery | pass | pinned Pi 0.80.6 RPC `get_commands` returned `true` for `pm-shepherd` |
| immutable-base diff | pass | `git diff --check e659d6f1...HEAD` exited 0 |
| owned path scope | pass | only issue-owned adapters, matching tests/fixture, and phase 476 artifacts changed |
| pushed exact-head equality | pass | evidence checkpoint `0a3cdfa4ce1ac46f87fc31ed14e295d17a4bb62c` matched local, tracking, and remote refs |
| forbidden gates | not run | no Go, connector, certification, runtime-backed, or `make` command ran |

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

- WorkspaceAdapter no longer passes an issuer through caller-overridable `acquireMutationLease`;
  GitAdapter privately registers a one-way acquisition closure and exposes no public lease issuer.
- Commit and push revalidate the immutable base as an ancestor. Handoff and push audit the union of
  every touched path across commit history, so add-then-remove paths cannot disappear from evidence.
- Mutating Git subprocesses receive deterministic safe config/environment overrides and reject
  executable hook, filter, credential-helper, SSH-command, include, and transport configuration.
- Schema-v4 workspace claims bind the inspected origin default branch. Push revalidates live remote
  symbolic HEAD, audits immediately before transfer, and sends the exact expected SHA refspec.

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
runtime-service, or `make verify` gate was run during correction 3.

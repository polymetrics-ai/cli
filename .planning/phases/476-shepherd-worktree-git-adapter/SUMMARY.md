# Issue 476 Summary

Status: `completed`

Issue #476 now provides a typed Git adapter and isolated-worktree policy with identities compatible
with existing Shepherd state, immutable complete-scope handoff claims, endpoint-bound remote
operations, and a nonforgeable exclusive writable capability. The Git boundary still exposes only
bounded status/fetch/branch/commit/push/diff/worktree operations; it does not expose destructive
cleanup, force push, default-branch mutation, arbitrary refspec, or unrestricted path capabilities.

The correction cycle addressed every finding from independent xhigh review of `906a45c5`:

- repository/worktree identities exactly match `target-evidence.ts` v1 identities;
- handoff evidence is derived from atomically published mode-0600 claim and worktree-binding
  records rather than caller-mutable fields;
- same-owner/same-request overlap is fenced by the existing append-only `FileStateStore` lease,
  with idempotent release and explicit dead-owner resume;
- full Shepherd verification is serialized so real Git fixture load cannot perturb SDK wall-clock
  tests, without widening any timeout;
- PR and reviewed-head evidence is concrete in `WORKER-HANDOFF.md` and `VERIFICATION.md`.

The second correction cycle addresses independent xhigh re-review of `d5181cd2`:

- every Git mutation requires the exact active WorkspaceAdapter-issued capability; release stops
  admission and waits for accepted mutations before releasing the underlying fenced lease;
- handoff audits the complete unfiltered `baseHead..head` path set before immutable-scope checks;
- effective fetch and push endpoints are bound, revalidated, rewrite-stability checked, and the
  exact validated endpoint is used for push and exact-head verification;
- alternate lease-root issuance, separator aliasing, and chained URL rewrites are covered by
  deterministic adversarial regressions.

The third correction cycle addresses independent xhigh review of `9728f9ed`:

- lease acquisition is a private registered closure, so caller overrides never receive reusable
  issuer authority;
- commit/push revalidate immutable-base ancestry, and history scope includes add-then-remove paths;
- worktree/add/push run behind deterministic safe Git config/environment fencing;
- the origin default branch is bound in schema-v4 claims and revalidated from live symbolic HEAD;
- push transfers the exact audited SHA instead of resolving a mutable local branch during transfer.

Focused tests pass 36/36 and the serialized full Shepherd suite passes 173/173. Strict TypeScript,
offline Pi 0.80.6 RPC command discovery, and exact diff/scope checks pass.

## Files delivered

- `.pi/extensions/shepherd/git-adapter.ts`: typed, sanitized, bounded Git process port.
- `.pi/extensions/shepherd/workspace-adapter.ts`: canonical issue/worktree policy, immutable claim
  binding, and fenced lease capability.
- Matching adapter tests and bounded local-Git fixture.
- `.planning/phases/476-shepherd-worktree-git-adapter/**`: GSD, TDD, verification, PR, and handoff
  evidence.

## Checkpoints

- `c4c4b8cf`: initial plan/GSD checkpoint.
- `7d44b472`: initial genuine RED checkpoint.
- `9e6e8753`: initial implementation checkpoint.
- `906a45c5`: independently reviewed predecessor.
- `36860ec5`: correction RED checkpoint.
- `e3669fc4`: correction GREEN checkpoint.
- `d91b41a8`: correction refactor/test-hardening checkpoint.
- `e8d1a3d7`: correction 2 genuine RED checkpoint.
- `6a22aa78`: correction 2 GREEN/refactor checkpoint.
- `2e255372`: correction 3 plan checkpoint.
- `fa607d31`: correction 3 genuine test-only RED checkpoint.
- `db6bdd67`: correction 3 GREEN checkpoint.
- `f7cb0cab`: correction 3 safe-config refactor checkpoint.

## Deviations

- `gitops-workflow` was not applied because the available skill is Kubernetes Argo/Flux-oriented.
- The repo GSD adapter lacks `programming-loop`; the complete lifecycle ran as
  `manual_gsd_fallback`.
- Pi 0.80.6 has no `--list-extensions`; the repository-documented offline RPC discovery passed.
- Parent policy limits this worker to Shepherd/TypeScript/Pi/diff gates; no Go, connector,
  certification, runtime-service, or `make verify` gate ran during correction.

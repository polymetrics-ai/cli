# Issue 476 Summary

Status: `completed`

Issue #476 now provides a typed Git adapter and isolated-worktree policy with identities compatible
with existing Shepherd state, immutable handoff claims, and a real exclusive writable lease. The
Git boundary still exposes only bounded status/fetch/branch/commit/push/diff/worktree operations;
it does not expose destructive cleanup, force push, default-branch mutation, arbitrary refspec, or
unrestricted path capabilities.

The correction cycle addressed every finding from independent xhigh review of `906a45c5`:

- repository/worktree identities exactly match `target-evidence.ts` v1 identities;
- handoff evidence is derived from atomically published mode-0600 claim and worktree-binding
  records rather than caller-mutable fields;
- same-owner/same-request overlap is fenced by the existing append-only `FileStateStore` lease,
  with idempotent release and explicit dead-owner resume;
- full Shepherd verification is serialized so real Git fixture load cannot perturb SDK wall-clock
  tests, without widening any timeout;
- PR and reviewed-head evidence is concrete in `WORKER-HANDOFF.md` and `VERIFICATION.md`.

Focused tests pass 21/21 and the serialized full Shepherd suite passes 158/158. Strict TypeScript,
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

## Deviations

- `gitops-workflow` was not applied because the available skill is Kubernetes Argo/Flux-oriented.
- The repo GSD adapter lacks `programming-loop`; the complete lifecycle ran as
  `manual_gsd_fallback`.
- Pi 0.80.6 has no `--list-extensions`; the repository-documented offline RPC discovery passed.
- Parent policy limits this worker to Shepherd/TypeScript/Pi/diff gates; no Go/connectors gate ran
  during correction.

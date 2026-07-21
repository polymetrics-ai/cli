Refs #476
Refs #471

## Summary

- add a typed Git process adapter with sanitized environment, bounded output/deadlines, exact head
  evidence, and repository/worktree identities compatible with existing Shepherd v1 state
- derive the canonical issue branch and trusted-root worktree path, persist immutable exact claim
  and worktree bindings, and reject mutable handoff evidence
- fence the one writable mutator with Shepherd's append-only file lease, idempotent workspace
  release, explicit stale-start guidance, and dead-owner same-request resume
- require a WorkspaceAdapter-held nonforgeable capability for every Git mutation and serialize
  release behind every accepted in-flight operation
- audit the complete canonical committed/dirty handoff path set before immutable-scope validation
- bind effective fetch/push endpoints, reject late pushurl or rewrite drift, and push/verify against
  the exact validated endpoint
- preserve dirty, untracked, conflicted, stale, and unique state without exposing worktree removal,
  reset, clean, prune, force push, default-branch push, or arbitrary refspec capability

## GSD / TDD

- GSD mode: `manual_gsd_fallback`; the repo-local adapter has no `programming-loop` command
- initial RED/GREEN: `7d44b472` / `9e6e8753`
- correction RED: `36860ec5` — 21 tests, 16 passed and 5 failed on the exact review contracts
- correction GREEN: `e3669fc4` — focused 21/21 and strict TypeScript passed
- correction refactor: `d91b41a8` — expanded identity, tamper, release, and crash-recovery evidence
- correction 2 RED: `e8d1a3d7` — 25 tests, 21 passed and 4 failed on the re-review contracts
- correction 2 GREEN/refactor: `6a22aa78` — capability, scope, endpoint, and adversarial cases pass
- skills: `gsd-programming-loop`, `gsd-workstreams`, `gsd-plan-phase`,
  `github-issue-first-delivery`, `architecture-patterns`, `javascript-testing-patterns`

## Verification

- focused issue tests — 29 passed, 0 failed
- full Shepherd suite with `--test-concurrency=1` — 166 passed, 0 failed
- strict no-emit TypeScript against cached Pi 0.80.6 Node types — pass
- documented offline Pi 0.80.6 RPC `get_commands` discovery — `true` for `pm-shepherd`
- exact-range diff, scope, and pushed implementation-ref equality — pass
- no Go, connector, certification, runtime-service, or `make verify` gate run during correction,
  per parent policy

The serialized full-suite command prevents real Git subprocess load from perturbing the SDK
runner's intentional wall-clock assertions; no timeout was widened.

## Safety and scope

- no dependency, credential, model/network call, GitHub mutation adapter, controller wiring,
  destructive cleanup, default-branch action, or merge is included
- CLI help/docs/website parity is not applicable because no `pm` CLI surface changes
- production changes remain limited to the two issue-owned adapter files

## Review route

Independent xhigh reviews of `906a45c5` and `d5181cd2` produced the corrections addressed here. A
fresh independent xhigh review must bind to the newly pushed exact candidate head. Claude and
Copilot are not requested, and this worker must not merge the sub-PR or parent PR.

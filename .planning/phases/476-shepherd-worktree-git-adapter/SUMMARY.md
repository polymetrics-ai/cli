# Issue 476 Summary

Status: `completed`

Implemented a typed Git process adapter and isolated-worktree policy adapter for Shepherd. The Git
port binds repository/worktree/origin identity, exposes bounded status/fetch/branch/commit/push/
diff/worktree operations, and verifies exact local and remote heads. The workspace policy derives
the only valid issue branch and path, persists hashed ownership claims, rejects aliases/collisions,
and reconciles an exact retry without removing dirty or unique state.

Focused tests pass 16/16 and the full Shepherd suite passes 153/153. Strict production TypeScript,
offline Pi 0.80.6 command discovery, and diff hygiene pass. The exact unsupported
`pi --list-extensions` result and the documented offline RPC fallback are recorded in
`VERIFICATION.md`.

No dependency, controller, domain, runner, extension, GitHub integration, destructive cleanup,
force/reset, default-branch push, arbitrary refspec, or unrestricted path capability was added.

## Files delivered

- `.pi/extensions/shepherd/git-adapter.ts`: typed, sanitized, bounded Git process port.
- `.pi/extensions/shepherd/workspace-adapter.ts`: canonical issue/worktree policy and ownership.
- `.pi/extensions/shepherd/git-adapter.test.ts`: typed Git safety and exact-head cases.
- `.pi/extensions/shepherd/workspace-adapter.test.ts`: ownership, containment, retry, collision,
  dirty preservation, and handoff cases.
- `.pi/extensions/shepherd/issue-476-git-fixture.ts`: bounded temporary local Git repositories.
- `.planning/phases/476-shepherd-worktree-git-adapter/**`: plan, TDD, verification, prompt, state,
  PR, and handoff evidence.

## Checkpoints

- `c4c4b8cf`: plan/GSD checkpoint.
- `7d44b472`: genuine RED test-contract checkpoint.
- `9e6e8753`: GREEN/refactor implementation checkpoint.
- Final evidence-only checkpoint follows this summary and is reported in the worker handoff.

## Deviations

- `gitops-workflow` was not applied because its available implementation is Kubernetes
  Argo/Flux-oriented; issue #476 and repository Git safety rules supplied the relevant contract.
- The repo GSD adapter lacks `programming-loop`; the complete lifecycle ran as
  `manual_gsd_fallback`.
- Pi 0.80.6 has no `--list-extensions`; the repository-documented offline RPC discovery passed.
- Parent policy superseded standalone full-repository gates. One exact `go test ./...` attempt hit
  the certify package's 10-minute timeout under shared CPU contention and its retry was stopped at
  parent direction. A full `make verify` had already passed before the policy changed.

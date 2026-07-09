# TDD Ledger: Bitbucket CLI Parity Parent

## Red evidence

- Pending for implementation lanes. Parent orchestration setup is planning-only.
- #90 will start with a focused red test for Bitbucket CLI surface/bundle validation before production Bitbucket defs are added.

## Green evidence

- Planning artifacts created before production edits.
- Required GSD/Pi health commands passed:
  - `scripts/gsd doctor`
  - `scripts/gsd verify-pi`
  - `scripts/gsd list --json`
- `scripts/gsd prompt plan-phase issue-79-bitbucket-cli-parity --skip-research --tdd` generated an official GSD planning prompt.

## Manual GSD fallback

`programming-loop` is not exposed by the current `scripts/gsd` registry. Command attempted:

```bash
scripts/gsd prompt programming-loop init --phase issue-79-bitbucket-cli-parity --dry-run
```

Result:

```text
scripts/gsd: unknown GSD command: programming-loop
```

Fallback in use: manual GSD universal runtime loop with `.pi/prompts/pm-gsd-loop.md`; maintain plan, red test/validation, green implementation, refactor, verification, commit/push, and review evidence.

## Refactor evidence

- Not started.

## Lanes

| Issue | Red | Green | Refactor | Notes |
|---:|---|---|---|---|
| #90 | pending | pending | pending | first local critical path lane |
| #91 | blocked | blocked | blocked | waits for #90 metadata |
| #92 | blocked | blocked | blocked | waits for #90 and stream definitions |
| #93 | blocked | blocked | blocked | avoid `api_surface.json` collision with #90 seed |
| #94 | blocked | blocked | blocked | waits for operation ledger |
| #95 | blocked | blocked | blocked | waits for operation ledger need classification |
| #96 | blocked | blocked | blocked | waits for operation ledger write risk classification |

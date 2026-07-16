# TDD Ledger: Issue 446

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
| --- | --- | --- | --- | --- |
| Shared PM logo | `npm run test:unit -- tests/brand-license-contract.test.ts` fails because `@/components/brand/pm-logo-mark` was deleted | Focused `-t "PM brand mark"` run passes: 2 tests | Removed three local cursor implementations and the now-unused global cursor animation | Green |
| Mixed license boundary | Contract assertions committed in the same focused test; collection was initially blocked by the missing shared logo module | Pending | Pending | Red queued |

## Rules

- Record the exact focused command and failure before production edits.
- Do not weaken an assertion to make the implementation pass.
- Update this ledger at each red, green, and refactor checkpoint.
- The license change remains human-gated even when automated tests pass.

## Environment Note

`npm ci` cannot run on current `origin/main` because PR #346 changed auth
dependencies without synchronizing `website/package-lock.json`. The focused red
run reused the existing ignored local `node_modules` tree; no dependency or
lockfile changes are part of this issue.

# TDD Ledger: Issue 446

| Slice | Red evidence | Green evidence | Refactor evidence | Status |
| --- | --- | --- | --- | --- |
| Shared PM logo | `npm run test:unit -- tests/brand-license-contract.test.ts` fails because `@/components/brand/pm-logo-mark` was deleted | Focused `-t "PM brand mark"` run passes: 2 tests | Removed three local cursor implementations and the now-unused global cursor animation | Green |
| Mixed license boundary | Full focused run fails on missing `internal/connectors/defs/LICENSE` and stale Elastic/public-source copy | Full focused run passes: 4 tests; root AGPL, nested MIT, map, and maintained copy agree | Consolidated policy in `LICENSING.md`; public surfaces link to the canonical files | Green |

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

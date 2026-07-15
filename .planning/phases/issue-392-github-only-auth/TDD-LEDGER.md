# TDD Ledger

Phase: issue-392-github-only-auth

## Red: Single Supported Provider

- Status: expected failure captured
- Command: `./node_modules/.bin/vitest run tests/auth-config.test.ts`
- Result: 1 failed, 2 passed.
- Evidence: `SOCIAL_PROVIDER_IDS` returned `['github', 'google', 'linkedin']` while the launch contract expected `['github']`.

## Green: Single Supported Provider

- Status: passed
- Targeted command: `./node_modules/.bin/vitest run tests/auth-config.test.ts`
- Targeted result: 3 tests passed.
- Full command: `npx -y pnpm@11.7.0 run test:unit`
- Full result: 9 files and 64 tests passed.
- Browser evidence: the signed-out Playwright flow found one GitHub button and zero Google or LinkedIn buttons.

## Refactor

- Status: complete
- Scope: replaced the multi-provider sign-in data loop and dynamic social-provider assembly with an explicit GitHub action while preserving the existing Better Auth callback and account-linking behavior.

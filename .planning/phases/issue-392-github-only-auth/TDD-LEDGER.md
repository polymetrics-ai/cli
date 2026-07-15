# TDD Ledger

Phase: issue-392-github-only-auth

## Red: Single Supported Provider

- Status: expected failure captured
- Command: `./node_modules/.bin/vitest run tests/auth-config.test.ts`
- Result: 1 failed, 2 passed.
- Evidence: `SOCIAL_PROVIDER_IDS` returned `['github', 'google', 'linkedin']` while the launch contract expected `['github']`.

## Green: Single Supported Provider

- Status: pending
- Command: targeted auth-config test, followed by full website unit suite.
- Expected: pass after GitHub becomes the only provider in config and UI.
- Evidence: to be recorded after implementation.

## Refactor

- Status: pending
- Scope: replace the multi-provider sign-in data loop and dynamic social-provider assembly with explicit GitHub configuration.

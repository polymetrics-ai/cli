# TDD Ledger

Phase: issue-392-github-only-auth

## Red: Single Supported Provider

- Status: pending
- Command: targeted `website/tests/auth-config.test.ts` Vitest run.
- Expected: fail because `SOCIAL_PROVIDER_IDS` still contains Google and LinkedIn.
- Evidence: to be recorded before production edits.

## Green: Single Supported Provider

- Status: pending
- Command: targeted auth-config test, followed by full website unit suite.
- Expected: pass after GitHub becomes the only provider in config and UI.
- Evidence: to be recorded after implementation.

## Refactor

- Status: pending
- Scope: replace the multi-provider sign-in data loop and dynamic social-provider assembly with explicit GitHub configuration.

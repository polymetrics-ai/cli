# PLAN: Issue 392 GitHub-Only Authentication

## Objective

Make GitHub the single launch OAuth provider for blog discussions so every visible sign-in action is backed by configured server behavior.

## Slice Boundaries

- Keep the existing Better Auth session, callback, account-linking, and test-auth behavior.
- Narrow the public provider contract, sign-in dialog, example environment, and deployment runbook to GitHub.
- Reuse the existing visual system; do not redesign unrelated discussion surfaces.
- Do not read or change local secret values, deploy production, merge the parent PR to `main`, or install Chatwoot.

## Required Skills And Context

- Loaded: `vercel-react-best-practices` and `vercel-composition-patterns`.
- Missing from the checkout: `frontend-design` and `web-design-guidelines`; preserve established components and tokens as the documented fallback.
- Read: issue agent contract and GSD universal runtime loop.
- Manual GSD fallback: `scripts/gsd prompt programming-loop ...` returned `unknown GSD command`.

## Tasks

1. Add a failing unit assertion that GitHub is the only supported social provider.
2. Narrow the provider type and Better Auth configuration to GitHub.
3. Simplify the sign-in dialog to a single GitHub action.
4. Remove Google and LinkedIn from example environment and production deployment instructions.
5. Run focused and full website checks.
6. Restart the updated website on port 3100 and verify the sign-in surface in a browser.
7. Commit and push the green slice, open a stacked PR into `feat/293-blog-annotations`, and record the review blocker.

## TDD Sequence

- Red: provider-contract test expects only `github` while production code still lists three providers.
- Green: provider config, UI, and docs expose GitHub only.
- Refactor: remove generic provider data and conditional object assembly that no longer serve multiple providers.

## Verification Commands

- `scripts/gsd doctor`
- targeted Vitest auth-config test
- website typecheck
- full website unit suite
- website production build
- focused Playwright/browser smoke checks on port 3100

## Commit And Push Checkpoints

1. Planning and red-test checkpoint.
2. Green implementation and documentation checkpoint.
3. Review-fix checkpoint only if findings require code changes.

## Risks

- A local GitHub OAuth app must use the exact port-3100 callback URL.
- A real OAuth round trip cannot be automated without operating the user's GitHub session.
- The parent PR currently requires human review fallback because automated reviewers are unavailable.

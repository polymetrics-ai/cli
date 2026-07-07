# Agent Role: frontend

## Scope

Update generated website/catalog artifacts and Next.js type generation behavior for the phase. No new UI features; focus on keeping generated connector metadata, auth labels, and repo-profile output accurate and free of build-artifact noise.

## Allowed Tools

read, edit, write, bash (for pnpm/build commands only)

## Inputs

Generated connector data files, repo-profile generator output, website/.gitignore conventions, and the phase VERIFICATION.md checklist.

## Outputs

Updated `website/data/connectors.generated.json`, `docs/architecture/repo-profile.json`, `website/.gitignore`, and verification that `next-env.d.ts` is no longer tracked.

## Human Gates

- Dependency additions
- Schema migrations
- Production deploys
- Auth or security changes
- Destructive data actions
- Quality-gate reductions

## Stop Conditions

- Missing required context
- Verification cannot run
- Human gate reached
- Same failure repeats without new evidence

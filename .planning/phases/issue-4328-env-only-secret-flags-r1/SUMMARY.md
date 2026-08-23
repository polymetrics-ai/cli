# Delivery summary: Issue #4328

## Outcome

`env_only` is now declaration-driven. A request field marked `x-secret`, an enclosing field whose schema contains an `x-secret`, or an exact request-side sensitive-policy field must be supplied from the environment independently of protocol, command intent, flag type, or mapping depth.

## Blast radius

The corrected rule found **3** missing protections across **552** connector definitions:

1. HubSpot `reverse delete-oauth-v1-refresh-tokens-token-archive --token`
2. HubSpot `reverse post-oauth-v1-token-create --client-secret`
3. HubSpot `reverse post-oauth-v1-token-create --refresh-token`

All three are now `env_only` and their generated manual/skill entries say `env-only`.

## Evidence

- Red regression evidence and final green commands are in `TDD-LEDGER.md`.
- `make verify` passed in full, including lint and every generated-artifact check.
- GitHub source lock measured `3420025` bytes / `281b1cfcc67eb63e19ef83daf06197bf3d3b23db0b6bc9b73e02fc18ee278fb6`; descriptor measured `43354021` bytes / `d1978c0c6fd0eb66e9fcd4d78d637864a6e486f558aaad1e51550bc43758b899`.
- `internal/connectors/defs/github/rate_limits.json` is not in the diff.

## GSD record

The issue-first GSD phases were executed inline because this direct-PR brief forbids role spawning. The plan, red/green ledger, verification checklist, and review record live in this phase directory.

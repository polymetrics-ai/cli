# Agent Trace: security

> DRAFT — populated from phase artifacts during the review-fix slice. This trace should be
> cross-checked against actual run logs before the phase is marked complete.

## Rendered Prompt Or Prompt Reference

.planning/phases/github-projects-discussions/THREAT-MODEL.md, AGENTS.md security rules

## Files Inspected

internal/connectors/defs/github/fixtures/**/*.json, internal/connectors/defs/github/spec.json, website/data/connectors.generated.json

## Actions Taken

Verified no secret-shaped literals in fixtures; clarified auth scope metadata for PAT vs GitHub App.

## Commands Run

grep -R "token\|secret\|password" internal/connectors/defs/github/fixtures/ (reviewed output)

## Findings

No secrets committed; auth notes now distinguish token types.

## Handoff Summary

Security review complete.

## Verification Evidence

THREAT-MODEL.md and fixture review.

## Unresolved Risks

Low — no auth-scope refresh or destructive action introduced.

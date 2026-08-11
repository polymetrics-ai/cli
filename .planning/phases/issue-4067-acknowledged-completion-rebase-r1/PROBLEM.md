# #4067 — acknowledged transport completion rebase

**Primary issue:** [#4067](https://github.com/polymetrics-ai/cli/issues/4067)
**Parent chain:** #3864 → #3862 → #4015
**Existing stacked PR:** [#4059](https://github.com/polymetrics-ai/cli/pull/4059)
**Rejected immutable candidate:** `883a86cf0040d559edcd4777413d1c2de20cd94a`

## Problem

After a transport ETL run durably acknowledges a checkpoint, another writer can change an unrelated part of the project state before the original run reaches `completeRun`. The original App still holds the old whole-state revision. Its normal final-completion save therefore returns `errStateRevisionConflict`, returns a zero `Run`, and leaves the acknowledged run durably `running` after reopen. The Sol witness reproduced this through the real JSON persisted-state path in every canonical mode.

## Exact interleaving

1. A transport run creates its `running` run and persists its own acknowledged stream checkpoint.
2. A distinct App instance persists an unrelated stream/checkpoint/run or other unrelated project field.
3. The first run reaches ordinary successful final completion with a stale whole-state revision.
4. The existing whole-state guard rejects completion, so the acknowledged target run remains durably `running` and the returned run is zero.

## Why this is not #4046 R9

#4046 deliberately permits a latest-state terminalization exception only after the typed `errTransportStreamStateConflict` from a rejected stale checkpoint writer. This issue concerns an acknowledged checkpoint whose terminal **completion** is blocked by an ordinary revision conflict caused by an unrelated later write. Expanding #4046 would weaken its typed-conflict-only boundary. #4067 therefore owns a separate, narrower acknowledged-completion rule.

## Required outcome

Rebase only final completion onto latest locked state when both conditions hold:

- the target generated run is still `running`; and
- the target stream exactly equals the checkpoint this run already acknowledged.

The rebase may mutate only that run's terminal fields and that run's final stream/run metadata. It must preserve all winner and unrelated state, never replay destination work, never overwrite a checkpoint, never become a generic last-writer-wins refresh, preserve committed/indeterminate/definite-not-committed outcome truth, and retain the detectable error chain.

## Scope

- `internal/app` final-completion logic and focused package-local tests.
- Canonical regeneration only of `website/lib/docs.generated.ts` and `internal/connectors/certifications/flow-matrix.json` for candidate-owned drift.
- This GSD/TDD evidence directory and #4059 update.

## Non-scope

- #4046 typed stale-writer failure finalization, R7/R8 stream-entry CAS, source identity, checkpoint writes, destination replay, provider/credential/network/warehouse/container/external-service work, and merge authority.

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

## R2 post-acknowledgement error and missing-run defects

The r2 Sol audit rejected candidate `3f84693bfbc128523a66e22653db7227fb9c0869` for two further manifestations at the same acknowledged boundary.

1. `runTransportETL` receives a non-zero orchestrator result plus cancellation or a source error after the checkpoint callback has persisted. It returns `etlExecutionResult{}` with that error, so `RunETL` calls ordinary `failRun`. If a second App has already committed unrelated state, ordinary whole-state revision protection rejects the stale failure write: the returned run is zero and the durable run remains `running`.
2. In an acknowledged latest-state completion rebase, an exact target run removed by the second App falls through to plain `run not found`. That error is not detectable as `errStateRevisionConflict` even though the operation must fail closed with no mutation.

The repair must retain a witness only after a durable acknowledgement, terminalize the exact still-running target run only after matching its acknowledged stream, preserve the original post-ack error, and make the missing target typed. It must not make ordinary `failRun` stale-state tolerant.

## R3 two-page typed-conflict defect

The r3 Sol audit then exercised two pages rather than a post-ack cancellation or ordinary source error. Page one durably acknowledged. A second App advanced the same stream to a winner checkpoint. The original App still applied page two, and its checkpoint CAS correctly returned `errTransportStreamStateConflict`. `runTransportETL` retained the page-one result, but `failAcknowledgedTransportRun` consulted the stale page-one witness first; its exact-stream guard rejected the winner before the existing #4046 typed-conflict `failRun` terminalizer could record the losing run. The observable result was a typed conflict plus `Run{}` and a durably reopened loser still `running`.

The required repair is ordering only: send that typed error directly to `failRun` before any acknowledged-witness lookup. This leaves r2's ordinary post-ack error guard and #4046's typed-conflict-only boundary intact. It never overwrites the winner, retries the checkpoint, replays either page, or turns #4067 into production transport wiring.

## Scope

- `internal/app` final-completion logic and focused package-local tests.
- Canonical regeneration only of `website/lib/docs.generated.ts` and `internal/connectors/certifications/flow-matrix.json` for candidate-owned drift.
- This GSD/TDD evidence directory and #4059 update.

## Non-scope

- #4046 typed stale-writer failure finalization, R7/R8 stream-entry CAS, source identity, checkpoint writes, destination replay, provider/credential/network/warehouse/container/external-service work, and merge authority.

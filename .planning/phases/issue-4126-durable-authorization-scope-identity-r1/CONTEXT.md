# Context — Issue 4126 durable authorization scope identity

## Settled decisions

- The shared authorization is an App-owned durable record, not a reusable token
  and not a replacement for `ReversePlan.PlanHash`.
- A scope hashes only its declared bound shape. Payload records, record counts,
  timestamps, cursor positions, and run identifiers have no field in the scope
  model and therefore cannot affect its identity.
- Destination configuration and credential changes are represented by the
  existing derived configuration digest and credential revision. Raw config,
  credentials, and secrets are never copied to the scope record.
- The existing plan/seal expiry is the authorization expiry for this slice. It
  is trusted authentication metadata, is part of the scope identity, and needs
  no new policy duration.
- The existing single-use reverse-plan approval is consumed atomically with
  creation of the authorization record. A later run re-derives the scope,
  checks the record before dispatch, and proceeds without a token only on an
  exact identity match.
- Reverse-plan execution is the narrow shared-app proof seam. No flow runner,
  schedule firing/rendering, or GitHub destination mode is added here.

## Manual-inline GSD fallback

The installed adapter commands were resolved through `scripts/gsd sources` and
their generated prompts were executed inline. This issue is an issue-first
foundation rather than a numbered roadmap phase, and the canonical single-worker
contract for this disposable worktree forbids lifecycle role spawning.

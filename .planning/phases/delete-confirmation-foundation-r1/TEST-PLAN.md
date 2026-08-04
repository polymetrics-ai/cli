# Test plan — destructive-write confirmation foundation

All tests use in-memory records, temporary project state, `httptest`, or bundle fixtures. No live
provider calls or credential values are permitted.

## Strict TDD slices

1. **Plan/declaration** — red proves the typed confirmation object is unsupported and destructive
   intent is not normalized for future operations; green adds the closed schema and normalized
   target policy.
2. **Preview** — red proves a destructive stored plan can execute without a prior preview; green
   persists the engine preview digest and refuses execution before preview.
3. **Explicit approval** — red proves caller-minted evidence and copied evidence can execute and
   replay; green requires an external-key authenticated, target-bound, expiring one-shot grant.
4. **Execute seam** — red proves no generic pre-dispatch wrapper exists for a future `rest_write`
   executor; green demonstrates the same wrapper blocks the callback until evidence is complete and
   then invokes it exactly once.
5. **Bypass resistance** — red proves the generic reverse path/direct engine path can bypass the
   full lifecycle; green covers connector-command, bulk reverse ETL, state tampering, digest
   mismatch, and `batchable: false` preservation.
6. **Hardening review** — red proves token-hash state tamper, stale-reader double dispatch, expired
   generic re-preview, incomplete request digests, and native SQS bypass. Green shares atomic grant
   consumption and canonical prepared-write execution across command, bulk, declarative, and SQS
   paths; Asana and Zendesk fixtures use real app preview/approval without provider calls.
7. **Trusted-input review** — red proves caller-key authority, stale whole-state resurrection, and
   mutable lifetime control. Green adds a vault-root plan seal, authority-derived grant expiry,
   configuration/batchability target binding, revision CAS, and an authenticated monotonic
   consumption marker outside state JSON.

## Verification

The review fix round executes one focused command across the changed connectors, vault, engine,
hook, conformance, native SQS, app, Asana fixture, and Zendesk Support fixture packages. The outer
pipeline owns subsequent test, lint, and CI phases, including the full 550+ connector suite.

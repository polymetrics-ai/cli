# SUMMARY — #3856 immutable polling-watermark conformance corpus

## Delivered

- An embedded, versioned `v1` polling-watermark corpus with SHA-256 evidence
  and defensive copies, independent from #3810's generic corpus.
- A reusable registered-lane runner with no fixture, filter, or skip input;
  registration fails closed on unsafe descriptor contracts and persisted
  checkpoints fail closed on incompatible provenance.
- Eleven mandatory fixtures covering equal-watermark page replay, empty and
  non-advancing pages, raw NULL/precision/coercion policy, unstable keysets,
  bounded overlap/commit lag, source/schema incompatibility, acknowledgement
  replay, tombstone/history behavior, hard-delete invisibility, and missing
  executor/evidence admission.
- Deterministic reference-lane tests that assert source requests, raw cursor
  values, durable checkpoint envelopes and positions, typed recovery, replay,
  tombstone/history transitions, and admission behavior.

## Source commits

- `00891ab91` — initial corpus and runner (#3856)
- `3fc992f65` — #4074 bounded-overlap request derivation
- `aa4d8c8a9` — staticcheck cleanup
- `cd92037fe` — admission and checkpoint-provenance hardening

## Scope and exclusions

No #3857 descriptor/preflight, #3858 provider execution, #3859 apply strategy,
provider implementation, PostgreSQL adapter, public CDC promotion, live
credential, or provider/database mutation was added. Shepherd #3995 / PR #4062
remains excluded. Credential-backed Transport is deferred to the first
provider-boundary child and final #4019 gate.

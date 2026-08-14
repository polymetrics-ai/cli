# TDD ledger — issue #3754

Manual-GSD fallback: the requested phase is absent from the numerical roadmap, so generated
GSD prompts are executed inline and this ledger is the durable red/green record.

| ID | Requirement | RED evidence to retain | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| R1 | Default process-local protection is real and labelled honestly | Existing declaration has no explicit coordination mode/provenance | Local limiter shares one opaque scope in-process; visible safe status says `process-local` and never `shared` | Planned |
| R2 | `require_shared` is explicit and never inherited | Absent/invalid coordination declaration is accepted or endpoint configuration selects shared by itself | Schema/semantic table tests prove local default and only declared `require_shared` selects shared | Planned |
| R3 | Shared coordinator grants atomically from server-time TTL state | Two clients can both consume a one-unit shared scope or no shared registry type exists | Atomic grant/block, reset expiry, context cancellation, and typed unavailable reason pass | Planned |
| R4 | Require-shared fails closed | Resolver uses local limiter after shared open/ping/admission failure | `errors.As` reaches typed reason naming missing coordinator; no request is sent | Planned |
| R5 | Scope identity remains secret-free | No cross-registry test proves raw subject/key absence | Test canary scope/binding material is absent from public status/error/storage key, files, logs, receipts, and delivery evidence; type path accepts only #3863 opaque scope key | Planned |
| R6 | Two processes obey one budget when shared is engaged | No cross-process real coordinator test exists | Opt-in Dragonfly test launches two helper processes under one opaque key; exactly one grant succeeds per window | Planned |

## Red command log

Pending. The exact failing commands and verbatim non-secret output will be appended before their
corresponding production implementation changes.

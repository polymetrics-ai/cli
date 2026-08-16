---
coverage:
  - id: D1
    description: PostgreSQL history transport is admitted by the outer descriptor and registry.
    verification:
      - kind: unit
        ref: internal/app/transport_composition_test.go TestOpenPostgresHistoryModeResolvesRegisteredExecutors
        status: pass
    human_judgment: false
  - id: D2
    description: The shipped binary writes, supersedes, and safely replays PostgreSQL history.
    verification:
      - kind: integration
        ref: internal/cli/postgres_transport_binary_integration_test.go TestPMBinaryExecutesPostgresIncrementalDedupeHistory
        status: pass
    human_judgment: false
---

# Summary — PostgreSQL incremental dedupe history repair

The outer PostgreSQL transport now declares `incremental_dedupe_history` on
both legs and maps it to `dedupe_history` / `managed_incremental_dedupe_history`.
The native destination seals the typed PostgreSQL history route before adapter
I/O, preserves the declared primary key, and carries the durable page position
to the existing conditional history fence. A fresh built binary proves target
history state independently of its own report: initial version, closed-and-new
version after update, and unchanged replay.

See `VERIFICATION.md` for commands, results, and the GSD inline fallback.

# PostgreSQL connector parity issue tree — captain correction r2

- **Status:** current execution scope as of 2026-08-10.
- **Supersedes:** r1 only for issue-tree scope and dependency order. r1 remains source-pinned
  evidence of the original audit and is retained intact.
- **Live parent:** [#3972 — Postgres Parity](https://github.com/polymetrics-ai/cli/issues/3972).

## Why r2 exists

The r1 tree correctly decomposed typed database foundations, PostgreSQL polling reads, target
writes, CDC, worksets, bootstrap, and certification. It did not assign an implementation owner for
the complete warehouse-mediated flow/mode matrix. Its parent and some child text also used short
forms such as “API→PostgreSQL,” which could be read as a direct connector-to-connector route.

The captain correction requires database read/write and API read/write to be considered only through
the warehouse, with every advertised `internal/synccontract.Mode` outcome stated honestly. The
correction makes that contract executable without reopening the active typed-foundation work.

## Binding route and mode truth

A connector implements only one side of a persisted warehouse flow. A source writes to its
connection-owned warehouse state; a target consumes a sealed warehouse table or immutable workset.
Neither receives a live counterpart connector, counterpart credential, or arbitrary relation. The
four canonical routes are:

1. API read → warehouse → API write.
2. API read → warehouse → PostgreSQL write.
3. PostgreSQL read → warehouse → API write.
4. PostgreSQL read → warehouse → PostgreSQL write.

PostgreSQL `change_capture` is a database-source extension of routes 3 and 4, not a direct target
feed: PostgreSQL 14+ streamed `pgoutput` v2 first produces a bounded transaction-preserving
warehouse workset and receives an acknowledgement only after the durable warehouse receipt.

| Mode | Current required parity result |
| --- | --- |
| `full_overwrite` | Supported PostgreSQL target mode after typed ownership, write-session, and driver foundations; consumes a warehouse workset. |
| `full_append` | Supported PostgreSQL target mode after the same foundations; consumes a warehouse workset. |
| `incremental_append` | Supported PostgreSQL target mode after the same foundations; consumes a warehouse workset. |
| `incremental_upsert` | Supported PostgreSQL target mode after the same foundations; key required. Derived change worksets use it first. |
| `incremental_dedupe` | Supported PostgreSQL target mode after the same foundations; deterministic winning row required. |
| `incremental_dedupe_history` | Recognized vocabulary but explicitly non-executable in phase one; must produce a typed rejection, never an invented history claim. |
| `change_capture` | PostgreSQL-source-only CDC with v2 streaming, bounded stage, durable receipt before LSN acknowledgement, and no cursor fallback. It is never a target write mode. |

API sides are admitted only where their own typed connector contract exists. Naming an API leg does
not claim generic API CDC, generic API writes, or a shortcut around plan → preview → approval →
execute.

## Corrected dependency graph

| Key | Issue | Scope | Hard dependencies |
| --- | --- | --- | --- |
| F1 | [#3974](https://github.com/polymetrics-ai/cli/issues/3974) | Typed database framework | — |
| F2 | [#3981](https://github.com/polymetrics-ai/cli/issues/3981) | Managed target ownership | F1 |
| F3 | [#3973](https://github.com/polymetrics-ai/cli/issues/3973) | Warehouse-input write session | F1, F2, #3859 |
| F4 | [#3975](https://github.com/polymetrics-ai/cli/issues/3975) | Committed transaction stage and warehouse receipt | F1 |
| F5 | [#3980](https://github.com/polymetrics-ai/cli/issues/3980) | Immutable warehouse delivery worksets | F1, F2 |
| P1 | [#3976](https://github.com/polymetrics-ai/cli/issues/3976) | PostgreSQL exact read into warehouse | F1, #3858 |
| P2 | [#3982](https://github.com/polymetrics-ai/cli/issues/3982) | PostgreSQL managed target driver | F1, F2, F3, #3859 |
| P3 | [#3977](https://github.com/polymetrics-ai/cli/issues/3977) | PostgreSQL CDC-to-warehouse producer | F1, F4, P1, PR #3967 |
| P4 | [#3979](https://github.com/polymetrics-ai/cli/issues/3979) | Gap-free warehouse bootstrap | F4, F5, P1, P3 |
| P5 | [#3983](https://github.com/polymetrics-ai/cli/issues/3983) | Warehouse workset to PostgreSQL target | F5, P2 |
| P6 | [#3987](https://github.com/polymetrics-ai/cli/issues/3987) | **New:** four-flow/seven-mode warehouse conformance | P1–P5, #3864 |
| C1 | [#3978](https://github.com/polymetrics-ai/cli/issues/3978) | Final live certification and capability promotion | P1–P6 |

### Dependency change

#3987 is a new hard gate before #3978. The former graph allowed final certification after P1–P5 even
though no issue proved every warehouse route or classified every advertised sync mode. The new gate
adds an integration wave after P1–P5 and before certification. It does **not** block or alter active
Wave A #3974: #3987 starts only after the existing read, write, CDC, bootstrap, and workset slices
are ready.

The corrected sequence is: Wave A F1; Wave B F2/F4 and P1 when #3858 lands; Wave C F3/F5/P3; Wave D
P2; Wave E P4/P5; Wave F P6; Wave G C1. P6 scaffolding may be reviewed earlier, but it cannot close
before P1–P5.

## Live REST audit and corrections

The audit read [#3972](https://github.com/polymetrics-ai/cli/issues/3972) and all eleven original
children via REST, then verified twelve attached sub-issues after correction. No GraphQL request was
used.

| Issue | Change | Why |
| --- | --- | --- |
| [#3972](https://github.com/polymetrics-ai/cli/issues/3972) | Body rewritten | Names all four warehouse routes, all seven mode outcomes, #3987, and the explicit extra certification gate. |
| [#3974](https://github.com/polymetrics-ai/cli/issues/3974) | **No change** | The active typed-framework foundation is correct and deliberately does not own routing, target DDL, write sessions, or CDC. |
| [#3981](https://github.com/polymetrics-ai/cli/issues/3981) | Binding scope-amendment comment | Makes target provisioning consume a sealed warehouse reference rather than a live source. |
| [#3973](https://github.com/polymetrics-ai/cli/issues/3973) | Binding scope-amendment comment | Makes the database write session warehouse-input only and preserves typed non-support for the two non-target modes. |
| [#3975](https://github.com/polymetrics-ai/cli/issues/3975) | Binding scope-amendment comment | Defines the transaction receipt as warehouse materialization before acknowledgement. |
| [#3980](https://github.com/polymetrics-ai/cli/issues/3980) | Binding scope-amendment comment | Confirms it is the sole outbound warehouse producer, never direct transport. |
| [#3976](https://github.com/polymetrics-ai/cli/issues/3976) | Binding scope-amendment comment | Requires exact PostgreSQL reads to materialize into the warehouse before a target leg. |
| [#3982](https://github.com/polymetrics-ai/cli/issues/3982) | Binding scope-amendment comment | Replaces ambiguous API→PostgreSQL wording with API → warehouse → PostgreSQL and names typed non-support. |
| [#3977](https://github.com/polymetrics-ai/cli/issues/3977) | Binding scope-amendment comment | Requires CDC-to-warehouse before separately approved target consumption. |
| [#3979](https://github.com/polymetrics-ai/cli/issues/3979) | Binding scope-amendment comment | Names the connection-owned warehouse/WAL/Parquet boundary and forbids a direct drain. |
| [#3983](https://github.com/polymetrics-ai/cli/issues/3983) | Binding scope-amendment comment | Defines PostgreSQL → warehouse → PostgreSQL workset delivery. |
| [#3987](https://github.com/polymetrics-ai/cli/issues/3987) | Created and attached as P6 | Owns the otherwise missing flow/mode conformance layer and gates C1. |
| [#3978](https://github.com/polymetrics-ai/cli/issues/3978) | Body rewritten | Requires P6 plus live proof of all four routes and all seven mode outcomes before capability promotion. |

Issue comments are deliberate durable amendments rather than silent deletion of the original scope.
They control where an original short-form route could be mistaken for a direct hop.

## Completion rule

Neither the tree nor its files certify PostgreSQL. `read`, `write`, and `cdc` stay at their truthful
runtime status until #3978 has accepted live evidence. The repository-wide Foundation Check and
`make connector-runtime-preflight` remain mandatory for every claimed executable command.

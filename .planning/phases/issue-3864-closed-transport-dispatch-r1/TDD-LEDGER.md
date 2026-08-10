# #3864 TDD ledger

| ID | Requirement | RED evidence | GREEN evidence | Status |
| --- | --- | --- | --- | --- |
| T1 | Canonical `full_append` reaches registered transports rather than the pre-resolution hard stop. | Pending test creation; current code hard-stops in `App.RunETL`. | Pending | Planned |
| T2 | Registry preflight rejects missing executor, type/family mismatch, unsupported mode, unsafe acknowledgement, missing strategy, and unavailable conformance verifier before source read. | Pending | Pending | Planned |
| T3 | Destination plan receives the descriptor-resolved strategy, not `upsert`. | Pending | Pending | Planned |
| T4 | One orchestrator handles fake API→API, API→database, database→API, database→database through warehouse-stage mediation without pair branches. | Pending | Pending | Planned |
| T5 | Candidate checkpoint commits only after durable destination acknowledgement; failed/unsafe acknowledgement leaves it uncommitted. | Pending | Pending | Planned |
| T6 | Source/destination context cancellation stops without an acknowledgement or checkpoint commit. | Pending | Pending | Planned |
| T7 | `connectors inspect` JSON/manual explicitly projects descriptor eligibility or unsupported roles without credential resolution. | Pending | Pending | Planned |

No production changes were made before this ledger and plan were created.

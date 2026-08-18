# UAT — Full-overwrite rerun correctness

Status: all automated acceptance deliverables passed.

| Deliverable | Direct evidence | Verdict |
| --- | --- | --- |
| Full-refresh boundary | The six-mode test withholds prior checkpoints from `full_append` and `full_overwrite`, and preserves them for all four incremental modes. | Pass |
| Full-overwrite rerun | Live first run was `3/3`; the changed-source rerun was `3/3`. | Pass |
| Exact replacement | Independent PostgreSQL queries changed target IDs from `[1 2 3]` to `[2 3 4]`; named ID 2 changed to `replacement-two`; deleted ID 1 count is zero. | Pass |
| Incremental no-regression | A dedicated live unchanged `incremental_upsert` rerun was `0/0`, retained three rows, and preserved `id=2 label="event-two"`. | Pass |
| All execution paths | One selector owns generic, run-scoped overwrite, serial Arrow, and pipelined Arrow source requests. | Pass |
| Safety and parity | Resume identity, generation, final candidate commit, publication order, destination plan, CLI/docs surface, and dependencies are unchanged. | Pass |

No human-judgment acceptance item remains. Detailed red/green counts are in `traces/red-full-refresh-checkpoint.md` and `traces/green-full-refresh-checkpoint.md`.

# UAT — PostgreSQL transport surface truthfulness

Automated inline verification is sufficient for this metadata/preflight lane; no human judgment is required.

| Deliverable | Result | Evidence |
| --- | --- | --- |
| Full overwrite is executable from production construction | Pass | `app.Open` resolves the registered PostgreSQL pair and its fixture full-overwrite source produces exactly three records. |
| Every advertised PostgreSQL destination mode preflights | Pass | Production registry table test resolves the five exact modes and asserts both executor references. |
| Unpaired history is refused before execution | Pass | Production registry returns `source transport does not support sync mode "incremental_dedupe_history"`; no executor is invoked. |
| Inspection and generated artifacts are truthful | Pass | Scoped certification generation/check plus built `pm connectors inspect postgres --json` and docs validation passed. |

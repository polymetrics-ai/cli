# Discussion log — Issue #3976

Manual inline `discuss-phase` record, 2026-08-11.

## Resolved from the issue contract and research

| Question | Decision | Source |
| --- | --- | --- |
| Does a new child issue own dynamic catalog discovery? | No. #3976 expressly owns exact catalog/type mapping and is the smallest existing source-boundary owner. | #3972, #3976, audit §8 |
| What may discovery inspect? | The configured live database, configured schema, and base relations only; no system namespace, view, arbitrary query, or destination target scope. | #3976, audit §§1–2 |
| How is variation represented? | Use #4034's typed `database` catalog identities, ordered relation/column/key values, native identity/modifiers, logical types, and deterministic fingerprint. | #4034, audit §2.2 |
| How should unknown or unsafe native shapes behave? | Return a named, typed, secret-safe failure. Do not degrade them to coarse text/object values. | #3976, audit §6 |
| What proves non-static behavior? | A real PostgreSQL test creates two materially different configured schemas and independently queries `pg_catalog` as its oracle; their discovered catalogs must differ without code/config schema edits. | user objective, audit §7 |
| Which downstream work stays out? | Parquet typing (#3980), target DDL/write/evolution (#3982), outbound apply (#3983), CDC execution (#3977), and flow/mode orchestration (#3987). | #3972 and child scopes |
| Can this child integrate now? | No. It may be built/reviewed, but parent integration is held until corrected #4058 is green and merged. | user ordering requirement, #4058 |

## Manual fallback rationale

The generated official GSD flow normally expects a numbered roadmap phase and
runtime roles. The canonical issue-first contract fixes one inline canonical
worker for this stacked child. This discussion therefore records decisions
already fixed by issue and repository contracts rather than reopening them.

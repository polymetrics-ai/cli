# Discussion Log

## 2026-08-18 — autonomous locked-context pass

The supervisor brief supplies all material decisions and explicitly requires autonomous execution. `--auto` therefore resolves the discussion phase without a human checkpoint.

- Priority: PostgreSQL → PostgreSQL control first.
- Proof standard: independently read the destination; never infer success from process status.
- Incremental standard: run the same admitted mode twice and state whether rows duplicate, skip, or update.
- Safety: do not restart local container runtimes; keep secrets out of argv, files, logs, and evidence; use only the disposable GitHub identity/repository if GitHub is reached.
- Scope: evidence and reporting only; no product fixes.
- Delivery: direct PR to `integration/4015-mvp-flat-r1` with `Refs #4015`.
- Time box: ship a decisive control result even when the three GitHub routes remain unattempted.

# Issue #4031 — Discussion Log

Mode: `scripts/gsd prompt discuss-phase 4031 --auto`.

The task brief already locked every material decision, so auto mode selected
the following bounded defaults:

1. Keep the architecture's composite keyset requirement; do not reduce it to the legacy scalar reader behavior.
2. Use executable PostgreSQL placeholders `$1` and `$2`, with `$1` reused for both cursor comparisons.
3. Add one focused static documentation assertion in the existing PostgreSQL test package so symbolic `$cursor` and `$primary_key` cannot reappear.
4. Do not alter PostgreSQL implementation, connector claims, certification, GitHub, or the #3855 branch.

No unresolved gray areas or deferred ideas remain.

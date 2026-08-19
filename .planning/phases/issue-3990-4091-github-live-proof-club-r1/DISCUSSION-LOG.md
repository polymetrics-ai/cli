# Issues #3990 and #4091 — discussion log

**Date:** 2026-08-15
**Mode:** `--auto`, inline/manual fallback

The captain supplied a complete non-interactive launch brief and a binding edge-case addendum. The
official GSD adapter was healthy, but the combined issue slug is not a ROADMAP phase, so
`init.phase-op` returned `phase_found: false`. The worker therefore captured the supplied decisions
directly in `CONTEXT.md`; no defaults broaden the target or the authorized provider-write scope.

## Auto-selected implementation decisions

| Area | Decision | Why |
| --- | --- | --- |
| Live target | Reuse the immutable, run-owned GitHub lab boundary | It is the existing default-deny provider safety boundary. |
| Execution path | Built `pm` production entry point only | Hand-building a component would reproduce the audited reachability defect. |
| Evidence | Sanitized structured records plus exact safe commands/output | The issues require reproducibility without credential or scope disclosure. |
| Failures | Typed refusal plus unchanged provider/checkpoint state | Exit status alone cannot prove a refusal happened before side effects. |
| Edge cases | Enumerate every captain-named case | Missing cases must be explicitly marked untestable/inapplicable, never silently dropped. |
| Scope | Fix only live-exposed defects in touched proof paths | Unrelated findings remain firstmate decisions. |

## Deferred

- #4125 and #4158 are explicitly excluded.
- Any permission class for which no separately scoped real credential exists will be recorded as an
  unavailable live observation and backed only by the existing deterministic refusal evidence; it
  will not be mislabeled live.

# Discussion log — issue-3755-rate-limit-operator-output-r1

## GSD discussion execution

`scripts/gsd doctor` passed. `scripts/gsd sources` resolved `discuss-phase`, `plan-phase`, `execute-phase`, `verify-work`, and `code-review`. The generated `scripts/gsd prompt discuss-phase 3755` was applied inline.

Pi's interactive runtime is unavailable in this worker, and the repository's canonical single-worker contract forbids role spawning. This is the documented manual-GSD fallback from `.agents/agentic-delivery/references/gsd-pi-adapter.md`; the issue supplies all product decisions, so no interactive questionnaire is needed.

## Decisions fixed by the issue and existing foundations

| Area | Decision | Source |
| --- | --- | --- |
| Output carrier | One bounded summary on a completed or failed ETL run, rendered by the existing human and `--json` paths | issue #3755 |
| Declaration absence | absent `rate_limits.json` is `undeclared`, not unlimited | issue #3755 |
| Selected policy identity | policy ID + subject kind only; report structural selector reasons but no runtime selector values | issue #3755; #3875 |
| Provider facts | use typed observation/reset/remaining/429 facts only; retain no header, URL, body, token, subject, binding, or opaque key | #3874; issue #3755 |
| Slowness split | aggregate local pacing, provider 429 retry waits, and request latency independently | issue #3755 |
| Boundedness | summaries coalesce by policy and scalar counters/durations; no per-request event list | issue #3755 |
| Enforcement | reporting is observational and must not change selector resolution, admission, retry, or legacy pacing | issue #3755; #3877 |
| Test data | only `internal/connectors/engine/testdata/` may declare a rate policy | issue #3755 |
| Parity | no new CLI surface; update ETL result documentation and prove human/JSON output | CLI parity contract |

No product ambiguity remains. Tests use only synthetic bundle metadata, an `httptest` provider, and fixture-safe values.

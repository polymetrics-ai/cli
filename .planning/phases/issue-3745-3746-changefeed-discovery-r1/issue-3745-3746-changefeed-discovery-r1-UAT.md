# UAT — issues #3745 and #3746 truthful changefeed discovery

Mode: inline GSD `verify-work` fallback. The coverage block in `SUMMARY.md` makes all three
deliverables automated, deterministic checks; no human judgment is needed for this foundation.

| Deliverable | Evidence | Result |
| --- | --- | --- |
| D1: closed, fail-closed descriptor | descriptor/loader unit suites | passed |
| D2: PostgreSQL no longer appears as CDC and inspect explains why | CLI regression suite plus executed local catalog/inspect commands | passed |
| D3: bounded foundation scope | changed-path audit, `git diff --check`, connector-boundary gate | passed |

Verdict: passed (automated). No UAT fix plan is required.

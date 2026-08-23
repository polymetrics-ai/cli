# Issue 4325 — TDD Ledger

| Slice | Red evidence to record before production edit | Green assertion | Refactor/quality gate | Status |
| --- | --- | --- | --- | --- |
| Baseline source drift | `go run ./cmd/connectorgen source-import asana --check` and the CircleCI equivalent each exited 1 because current main lacks `defs/<connector>/sources/` | Re-pinned import and descriptor check pass with provider-derived exact set | `connectorgen validate` and `surface-sync --check` | red reproduced |
| CircleCI/Sentry/Vercel surfaces | A built baseline binary returned exit 2 / `unknown command` for all three `operations list` probes | The same command returns `missing --credential` and preflight accepts it | Generated command/help checks | red reproduced |
| Jira reachability | Enabled row lacks typed operation or disabled write row contradicts implemented command | Every enabled row has the exact typed operation and direct writes report their actual state | Real preflight sweep | pending |
| Docker Hub/Notion truth | Contradictory status/citation or wrong covered-by action is detectable | Metadata points to existing source/action and current refusal line | Bundle validation | pending |
| Stripe semantics | Four JSON response routes classify as binary | The four routes classify as JSON direct reads and command output policy matches | Semantic regression check | pending |
| Evidence reasons | Report scan finds forbidden scope reason or an uncited foundation gap | Scan finds none; all remaining citations resolve to the stated runtime refusal | Independent Gate B rerun | pending |
| Final gate | Independent report returns NO-GO | Independent report returns GO | Full `make verify` and review | pending |

Every red/green command and its result is appended when it executes. No test is
weakened or skipped to advance a row.

# Issue 4325 — TDD Ledger

| Slice | Red evidence to record before production edit | Green assertion | Refactor/quality gate | Status |
| --- | --- | --- | --- | --- |
| Baseline source drift | `source-import --check` rejects the report’s drifted locks or a fresh identity comparison finds the stated stale/missing rows | Re-pinned import and descriptor check pass with provider-derived exact set | `connectorgen validate` and `surface-sync --check` | pending |
| CircleCI/Sentry/Vercel surfaces | Built binary returns `unknown command` for an intended provider operation | The same command returns `missing --credential` and preflight accepts it | Generated command/help checks | pending |
| Jira reachability | Enabled row lacks typed operation or disabled write row contradicts implemented command | Every enabled row has the exact typed operation and direct writes report their actual state | Real preflight sweep | pending |
| Docker Hub/Notion truth | Contradictory status/citation or wrong covered-by action is detectable | Metadata points to existing source/action and current refusal line | Bundle validation | pending |
| Stripe semantics | Four JSON response routes classify as binary | The four routes classify as JSON direct reads and command output policy matches | Semantic regression check | pending |
| Evidence reasons | Report scan finds forbidden scope reason or an uncited foundation gap | Scan finds none; all remaining citations resolve to the stated runtime refusal | Independent Gate B rerun | pending |
| Final gate | Independent report returns NO-GO | Independent report returns GO | Full `make verify` and review | pending |

Every red/green command and its result is appended when it executes. No test is
weakened or skipped to advance a row.

# Top daily-use declarations — increment 001

Batch 1 has pinned and cross-checked 4,378 publicly documented operations
across ten daily-use connectors. All 4,378 are declared; 476 are currently
command-backed and enabled. Live credentialed certification is pending for all
ten and was not attempted.

| Connector | Rank | Documented / declared | Disabled | Runnable commands | Write actions (delete) | ETL source | Reverse-ETL eligibility |
| --- | ---: | ---: | ---: | ---: | ---: | --- | --- |
| Docker Hub | 3 | 54 / 54 | 13 | 41 | 20 (6) | declared | foundation-gap |
| GitLab | 5 | 1,755 / 1,755 | 1,751 | 4 | 0 (0) | declared | foundation-gap |
| Jira | 7 | 617 / 617 | 322 | 295 | 292 (84) | declared | foundation-gap |
| Vercel | 8 | 400 / 400 | 400 | 0 | 18 (8) | declared | foundation-gap |
| Notion | 9 | 49 / 49 | 6 | 43 | 24 (4) | declared | foundation-gap |
| Stripe | 10 | 589 / 589 | 581 | 8 | 3 (1) | declared | foundation-gap |
| Bitbucket | 11 | 331 / 331 | 328 | 3 | 54 (53) | declared | foundation-gap |
| CircleCI | 12 | 111 / 111 | 111 | 0 | 7 (3) | declared | foundation-gap |
| Sentry | 13 | 223 / 223 | 223 | 0 | 0 (0) | declared | foundation-gap |
| Asana | 18 | 249 / 249 | 167 | 82 | 73 (4) | declared | foundation-gap |
| **Total** | — | **4,378 / 4,378 (100%)** | **3,902** | **476** | **491 (163)** | **10 / 10** | **0 eligible / 10 gap** |

All ETL source descriptors use the definition-owned
`declarative_stream_source` contract from PR #4286, each with its actual
stream allowlist and connector-owned evidence reference. A provider mutation is
a direct-write endpoint, even when it has a typed write action: it does not
imply reverse ETL. The cohort has 2,370 direct-write endpoints and 118 enabled
direct-write bindings. Reverse-ETL eligibility is an attribute on those rows,
and has one real, recoverable foundation gap:
`generic-typed-destination-executor`.
`internal/app/issue_label_warehouse_transport.go:85-95` still registers only
the closed issue-label destination, and no bundle has a matching typed action
binding. `TRANSPORT-GAP.md` and `FOUNDATION-GAP-REASONS.json` carry the exact
evidence and minimal change.

No provider credentials, provider calls, generic HTTP writer, or destination
`transport_binding` was added. `connector-boundary` remains a required CI gate
because its detached local capture produced no exit record.

# Top daily-use declarations — increment 001

## Complete six-class map checkpoint

The batch is now mapped before certification. See
`COMPLETE-PARITY-MAP.md` for the required per-connector source operation
classes, command-derived `ENABLED%`, delete coverage, and #4286
definition-owned transport gaps. Certification remains deliberately deferred.

This checkpoint pins public API descriptions and records declaration parity
only. It performs no live credentialed certification; every connector retains
`live_certification: pending`.

| Connector | Rank | Documented / declared | Disabled | Runnable commands | Write actions (delete) |
| --- | ---: | ---: | ---: | ---: | ---: |
| Docker Hub | 3 | 54 / 54 | 13 | 41 | 20 (6) |
| GitLab | 5 | 1,755 / 1,755 | 1,751 | 4 | 0 (0) |
| Jira | 7 | 617 / 617 | 27 | 590 | 292 (84) |
| Vercel | 8 | 400 / 400 | 378 | 0 | 18 (8) |
| Notion | 9 | 49 / 49 | 5 | 49 | 24 (4) |
| Stripe | 10 | 589 / 589 | 581 | 8 | 3 (1) |
| Bitbucket | 11 | 331 / 331 | 134 | 5 | 54 (53) |
| CircleCI | 12 | 111 / 111 | 95 | 0 | 7 (3) |
| Sentry | 13 | 223 / 223 | 220 | 0 | 0 (0) |
| Asana | 18 | 249 / 249 | 164 | 249 | 73 (4) |
| **Total** | — | **4,378 / 4,378 (100%)** | **3,376** | **938** | **487 (161)** |

All ten were selected from the ranked daily-use shortlist because they expose
a retrievable primary public OpenAPI description and have an existing
declarative API-surface inventory. Docker Hub is the rank-3 proof job. Native
MySQL and the absent Gitea bundle are deferred: Gitea must be created before it
can be a declaration candidate, not pinned as though it already had a bundle.

## Docker Hub runnable-parity retrofit

Docker Hub's 54 pinned method/path rows have exact source crosswalk and
disposition records. Its 50 typed source contracts comprise 23 `rest_read` and
27 `rest_write` contracts, including six documented delete contracts.

Forty-one source operations are runnable: four existing ETL commands, 17
operation-bound direct reads, and 20 source-backed reverse-ETL actions. The
write actions include six typed destructive deletes. `ENABLED%` is **75.93%**
(41 runnable operations / 54 documented operations); declared coverage is
100%. Elevated token scope is runtime authorization metadata, not a disabled
disposition.

The 13 disabled rows are ten documented foundation gaps and three schema/media
incompatibilities. Token list/detail/update/delete endpoints are enabled
because their pinned responses expose metadata, not the secret value. The two
token-create responses and the login/2FA/auth-token routes are declared with
secret fields and `sensitive_policy`, then held for the named secret-response
and live secret-write foundation work. `unsafe-to-exercise` is zero. All gaps
carry exact evidence and `recoverable: true` in `REJECTION-LIST.json`.
`certification-sweep.json` has 43 non-live rows (41 CLI commands plus
read/write capability rows); fixture or live proof remains pending and no live
provider call was made.

The engine still has no registered generic declarative API source/destination
transport. No `sync_transport.json` was fabricated; the recoverable #4093
transport gap is recorded for all ten connectors in `TRANSPORT-GAP.md`.

# Increment 001 summary

Ten ranked daily-use API connectors now have reproducible public-source locks and complete source-to-`api_surface` declaration parity: Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana.

- Public operations pinned and declared: 4,378 / 4,378 (100%).
- Explicitly disabled with exact evidence/recovery: 3,405.
- Typed write actions retained: 471, including 157 `delete` actions.
- Generated non-live certification commands: 909; all connector live certifications remain pending.
- No `sync_transport.json` was invented: all ten source/destination transport gaps are explicitly `recoverable: true`, with runtime evidence and the smallest safe recovery in `TRANSPORT-GAP.md`; #4093 tracks the foundation.

Verification and review evidence is in `VERIFICATION.md`, `SOURCE-LOCK-VERIFICATION.json`, and `REVIEW.md`. This is a checkpoint summary, not the final batch delivery: the captain's full-parity order requires the existing cohort to be completed before any new daily-use connector is selected.

## Captain full-parity correction — Docker Hub complete

The later full-parity order suspends new-connector selection until this first
cohort is declared exhaustively. Docker Hub is now the completed proof slice:
54 / 54 pinned operations have an exact source-to-surface crosswalk and an
itemized declaration/disposition. Its 49 source-contract inventory entries are
23 `rest_read` plus 26 `rest_write`, including six typed delete contracts.
Four existing streams remain the only executable routes; 50 terminal routes
remain explicitly disabled (46 elevated-scope, three response-less HEAD
foundation gaps, one deprecated login). No `writes.json` action,
`sync_transport.json`, provider credential, or live certification claim was
added. The next full-parity connector is Notion, not Increment 2.

# Increment 001 summary

Ten ranked daily-use API connectors now have reproducible public-source locks and complete source-to-`api_surface` declaration parity: Docker Hub, GitLab, Jira, Vercel, Notion, Stripe, Bitbucket, CircleCI, Sentry, and Asana.

- Public operations pinned and declared: 4,378 / 4,378 (100%).
- Explicitly disabled with exact evidence/recovery: 3,405.
- Typed write actions retained: 471, including 157 `delete` actions.
- Generated non-live certification commands: 909; all connector live certifications remain pending.
- No `sync_transport.json` was invented: shared generic declarative source/destination registration is tracked by #4093.

Verification and review evidence is in `VERIFICATION.md`, `SOURCE-LOCK-VERIFICATION.json`, and `REVIEW.md`. This is a checkpoint summary, not the final batch delivery: the next increment must be selected from the ranked daily-use pool after the #4283 checkpoint commit.

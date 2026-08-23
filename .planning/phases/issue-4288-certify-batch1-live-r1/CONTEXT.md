# Issue #4288 context — batch-1 live certification

## Decisions fixed by the issue and launch brief

- Scope is Jira, Asana, then Notion only. Docker Hub, CircleCI, Sentry, Vercel,
  Stripe, Bitbucket, and GitLab are owned by the concurrent declaration lane and
  are excluded.
- Certify only capabilities already declared and implemented. Do not alter a
  connector definition or shared engine behavior to make a cell pass.
- A certification-scope allowlist entry is necessary before the existing
  generator can create or validate any evidence for these three connectors.
  It is not a connector-definition or engine change.
- Credentials are live, free/self-serve, secret-safe, and connector-isolated.
  Live provider reads come before any scratch-owned mutation. Captcha, device
  2FA, a payment/card requirement, an unapproved paid decision, or a page that
  defeats all prescribed interaction techniques is a stop.
- Evidence contains counts, status, response classification, and salted
  fingerprints only. It contains no token, account identifier, personal data,
  raw response scalar, or provider verification value.

## Manual GSD discussion fallback

This direct-PR task is governed by a concrete issue and launch brief, and the
canonical contract forbids spawning GSD role agents. No unresolved product
choice remained after the issue, AGENTS.md, and provisioning runbook review.
The inline discussion decision is to preserve those constraints verbatim;
`DISCUSSION-LOG.md` records the audit trail.

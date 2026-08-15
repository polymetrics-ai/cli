# Discussion Log: live-path defects r1

## 2026-08-16 — non-interactive issue-authorized discussion

- **#4119:** canonicalize and decide admission at the requester send boundary;
  do not add a redirect-only exception. Keep local-only connectors redirectable.
- **#4125:** choose a typed pre-I/O rejection for invalid windows. A bound is
  explicit rather than silently clamped, and it protects both duration and TTL
  derivations.
- **#4169:** map provider-verified authentication rejection to the existing
  caller/credential category without recording or displaying credential data.
  Do not reclassify genuine internal errors.
- **Scope:** #4158 remains untouched. The PR will reference all three assigned
  issues because the launch brief explicitly directs one PR, even though its
  fixes remain separate commits.

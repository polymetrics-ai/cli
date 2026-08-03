# Issue tree — engine-refresh-token-auth-r1

Created **before** any code was written, per the captain's standing order.

| Issue | Title | Capability |
| --- | --- | --- |
| [#3702](https://github.com/polymetrics-ai/cli/issues/3702) | feat(engine): OAuth2 refresh-token grant as a shared engine auth mode | parent |
| [#3703](https://github.com/polymetrics-ai/cli/issues/3703) | feat(engine): add an `oauth2_refresh_token` auth mode to the engine dispatch | 1 — config shape + first exchange |
| [#3704](https://github.com/polymetrics-ai/cli/issues/3704) | feat(connsdk): reuse a refresh-token access token until near expiry | 2 — expiry-aware caching |
| [#3705](https://github.com/polymetrics-ai/cli/issues/3705) | feat(engine): persist provider-rotated refresh tokens to the encrypted local vault | 3 — rotation |
| [#3706](https://github.com/polymetrics-ai/cli/issues/3706) | feat(connsdk): refresh on 401 at most once per request | 4 — bounded 401 refresh |
| [#3707](https://github.com/polymetrics-ai/cli/issues/3707) | feat(connsdk): collapse concurrent refresh-token exchanges to exactly one | 5 — concurrency |

Sub-issues are linked to the parent through GitHub's native sub-issue relationship
(`gh-axi issue subissue add 3702 3703 3704 3705 3706 3707`), so #3702 renders them as a checklist.

## What each one unblocks

- **#3703** — every connector whose provider issues user-context tokens. Immediately Reddit, whose
  `spec.json` records that "OAuth token acquisition/refresh is out of scope; the caller supplies a
  valid token" against tokens that expire one hour after issuance.
- **#3704** — unattended scheduled syncs of any length. Without reuse the mode exchanges per
  request; without a conservative default TTL a provider that omits `expires_in` produces a
  connector that dies an hour in, which is the exact failure being removed.
- **#3705** — every rotating provider. Without it, adoption is a one-run time bomb: the connector
  works once and then fails with `invalid_grant` long after the run that caused it.
- **#3706** — long-running syncs against providers that revoke out of band (password change, scope
  change, revoked app). Also converts a wall of 401s into one clean terminal error.
- **#3707** — multi-stream syncs, i.e. essentially all of them. Several streams share one
  authenticator; without this they exchange N times, and some providers invalidate a refresh token
  on concurrent use.

## Identity note

Issues were created under `karthik-sivadas` rather than `alfred-polymetrics-ai`: no Alfred
credential exists in this delivery environment, and GitHub authorship cannot be reattributed after
creation. The captain authorised this as a **bounded exception covering issue creation only**. It
does not extend to approvals, merges, or branch-protection changes, and none of those were
performed.

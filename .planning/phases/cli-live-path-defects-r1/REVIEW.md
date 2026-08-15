# Code Review: live-path defects r1

## Method

Inline/manual `code-review` fallback, per the generated GSD prompt: this task
forbids role spawning and no compatible isolated GSD worker is available.
Reviewed the changed source and tests, error-chain behavior, arithmetic bounds,
request side effects, and focused test evidence.

## Findings

No actionable findings.

- `CredentialRejectedError` exposes no raw URL, response body, or credential;
  the formatter's `As` method preserves only that safe identity for 401.
- `classifyError` maps that identity and raw terminal 401 fallback to one
  explicit auth outcome, while generic errors still use the internal category.
- Window validation runs before millisecond/duration conversion and before
  `EnsureAvailable`; the maximum includes the one-second TTL slack and all
  rejected cases are typed.
- Redirect tests assert final provider counts and typed reasons instead of only
  checking non-nil errors. Local-only redirects remain exercised.

## Disposition

The isolated legacy full-flow failure is not changed or re-baselined. It occurs
before the auth assertion and is recorded in `VERIFICATION.md` for its owner.

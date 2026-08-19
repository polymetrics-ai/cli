# Summary — issue-3753-rate-limit-enforcement-r1

Manual GSD execution completed for #3753 and #3754 because Pi interactive roles were unavailable and the canonical contract forbids replacement role spawning.

- Declared HTTP policies now resolve per connector request using endpoint, tier, and auth selectors.
- A process-local, context-cancellable registry enforces every matching fixed/sliding/token/leaky budget, including request and point costs.
- Scope keys come only from `CoordinationIdentity.RateScopeKey`; the implementation does not read secrets or credential revisions and retains only opaque scope projections in the registry.
- Engine request entry points now acquire a resolved requester before direct, declarative, paginated, multipart, and binary sends. Legacy `base.rate_limit` remains separate.
- Only testdata declares rate limits. Production `defs.go` stays unchanged, and #3755 operator output is excluded.

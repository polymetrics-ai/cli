# Summary — issue-3752-rate-limit-admission-r1

Status: #3751 and #3752 implementation slices are committed locally; final scoped gates and review remain.

The branch now loads and validates an optional, provider-cited `rate_limits.json` declaration with
typed policy, selector, scope, budget, citation, and honest-state values. No legacy bundle is made
mandatory and no existing rate-limit metadata is activated. `connsdk.Requester` now has a
context-aware pre-send admission seam, parsed typed observations, exact `Retry-After` reset timing,
full jitter only for fallback retries, and a terminal `RateLimitError` that unwraps to `HTTPError`.
Resolver integration (#3753), registry/coordinator scope keys (#3754), and operator output (#3755)
remain intentionally deferred.

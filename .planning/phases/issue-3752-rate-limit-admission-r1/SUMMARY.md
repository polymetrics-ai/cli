# Summary — issue-3752-rate-limit-admission-r1

Status: #3751 is implemented and green; #3752 requester work is next.

The branch now loads and validates an optional, provider-cited `rate_limits.json` declaration with
typed policy, selector, scope, budget, citation, and honest-state values. No legacy bundle is made
mandatory and no existing rate-limit metadata is activated. The remaining requester work adds
admission, observations, exact provider-reset handling, bounded fallback jitter, and a typed 429
error. Resolver integration (#3753), registry/coordinator scope keys (#3754), and operator output
(#3755) remain intentionally deferred.

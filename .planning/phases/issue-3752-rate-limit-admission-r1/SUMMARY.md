# Summary — issue-3752-rate-limit-admission-r1

Status: planning complete; implementation has not begun.

The phase will first add an optional provider-cited `rate_limits.json` contract, then add requester
admission, observations, exact provider-reset handling, bounded fallback jitter, and a typed 429
error. Resolver integration (#3753), registry/coordinator scope keys (#3754), and operator output
(#3755) remain intentionally deferred.

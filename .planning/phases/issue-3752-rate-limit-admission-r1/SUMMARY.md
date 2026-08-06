# Summary — issue-3752-rate-limit-admission-r1

Status: #3751 and #3752 are committed and locally verified; no-mistakes/PR/CI are firstmate-gated.

The branch now loads and validates an optional, provider-cited `rate_limits.json` declaration with
typed policy, selector, scope, budget, citation, and honest-state values. No legacy bundle is made
mandatory and no existing rate-limit metadata is activated. `connsdk.Requester` now has a
context-aware logical-send admission seam, parsed typed observations, exact `Retry-After` reset timing,
full jitter only for fallback retries, and a terminal `RateLimitError` that unwraps to `HTTPError`.
Resolver integration (#3753), registry/coordinator scope keys (#3754), and operator output (#3755)
remain intentionally deferred.

The local manual-GSD fallback recorded `UAT.md` and `REVIEW.md`. Scoped package, schema, build,
lint, docs, smoke, contract, boundary, and release checks passed. A full connsdk race run reports
an unchanged multipart test-data race outside this slice; focused new rate-limit and loader race
tests pass. The branch was then rebased onto `origin/main` at `d215d9636` and those scoped gates
were re-run; the independently landed `Changefeed` bundle field and this slice's `RateLimits`
loader are both retained.

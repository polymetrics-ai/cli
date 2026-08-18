# Summary — Issue 4126 durable authorization scope identity

Implemented the App-owned durable authorization record and its canonical,
content-free scope identity. The initial approved reverse-plan proceed now
atomically consumes its token and creates one safe record; a later run with no
token re-derives the scope and refuses revocation, expiry, or any bound-scope
drift before reaching the provider.

The scope persists only opaque references and derived credential/configuration
digests. It intentionally excludes record content/count, timestamps, cursors,
and run identity; `reversePlanHash` remains the original payload-bound first-run
identity.

The TDD suite proves changed-payload execution and zero outbound sends for
scope change, revocation, expiry, and replay. The loopback GitHub provider is
the individually documented hermetic fake for the existing plan/preview/write
path; no flow runner, schedule firing path, or GitHub destination mode was
added.

Local verification is recorded in `VERIFICATION.md`. Delivery pushes this
stacked branch and explicitly opens its PR against
`integration/4015-mvp-flat-r1`; the API-reported base must be verified.

CI additionally exposed that the new typed token-replay refusal reached the
generic CLI error fallback. It is now explicitly classified as caller
validation (`validation_error`, exit 3), retaining the existing no-state-write
replay proof rather than accepting `internal_error`.

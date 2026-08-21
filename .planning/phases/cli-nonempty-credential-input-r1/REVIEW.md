# Inline code review — provider-neutral non-empty credentials

## Scope

Manual standard-depth review of the new credential contract and each changed
input, persistence, engine, connsdk, test, CLI manual, generated manual, and
website-documentation path. This is the documented inline fallback: the
canonical worker contract forbids spawning a separate reviewer in this
delivery context.

## Findings

No remaining Critical, Warning, or Info findings.

One review observation was corrected before this report: an existing
secret-store test comment became detached while inserting the new App test.
The comment now documents its original invalid-key test and no behavior was
changed by that correction.

## Security disposition

- The shared typed error contains only the non-secret field name.
- Stdin removes at most one documented delimiter; persistence remains
  byte-exact after that boundary.
- Checks occur before App/vault mutation and again before request/header/query
  or OAuth token-form mutation.
- Existing optional no-auth selection and optional refresh-token client-secret
  behavior are preserved.
- No provider names, Twenty branches, secret values, plaintext files, logs, or
  process-list transports were added.

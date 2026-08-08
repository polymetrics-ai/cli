# TDD Ledger — Zoom Chatbot documented-operation parity, R1

## Planned RED contract

Before any production engine or Zoom bundle change, the RED checkpoint will contain only tests,
synthetic fixtures, and planning evidence. It must fail against the current branch because:

- Zoom remains at `23` executable / `1,819` local implementable rows, with one direct write;
  the target requires `27` / `1,815` and five direct writes.
- `chatbot messages send`, `chatbot messages edit`, `chatbot messages delete`, and
  `chatbot link-unfurls create` do not resolve through the real commandrunner preflight.
- The declared `json_object` flag type needed for the provider's named `content` object is
  unsupported; the existing generic `json` type remains deliberately unsupported.
- A raw OAuth client-credentials declaration using `client_auth: basic` is rejected by the bundle
  schema before the test can prove the required Basic token exchange.

The failure output and the red commit hash will be appended here immediately after the test run.

## Planned GREEN contracts

- `client_auth: basic` makes the token request use HTTP Basic for the client ID/secret and keeps
  `grant_type=client_credentials` in the form body without copying credentials there. Existing
  default client-credentials behavior remains form-post compatible.
- `json_object` parses one object only and reaches the operation's declared body schema as an
  object. Scalars, arrays, malformed documents, extra documents, and the unscoped generic `json`
  type are rejected.
- All four Chatbot direct writes pass real runtime preflight and binary plan lifecycle checks;
  no-body Link Unfurls asserts a `204` status, and DELETE requires destructive confirmation.
- Zoom's endpoint ledger changes only its four Chatbot rows. No `unsafe_or_disallowed` row is
  introduced.

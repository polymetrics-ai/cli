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

The test ran before any production bundle or engine change:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine ./internal/connectors/connsdk ./internal/connectors/commandrunner ./internal/connectors/defs/zoom/...
--- FAIL: TestSelectAuthOAuth2ClientCredentialsBasicClientAuth (0.00s)
    auth_test.go:203: Load bundle with declared basic client auth: load bundle acme: streams.json: /base/auth/0/client_auth: additional property not allowed
FAIL
FAIL	polymetrics.ai/internal/connectors/engine
--- FAIL: TestCoerceFlagValueAcceptsJSONObject (0.00s)
    runner_test.go:1265: coerce declared json_object: connector command "unknown" is blocked: flag --content has unsupported type "json_object"
FAIL
FAIL	polymetrics.ai/internal/connectors/commandrunner
--- FAIL: TestProviderInventoryLedgerIsComplete (0.03s)
    command_surface_test.go:155: executable rows = 23, want 27
    command_surface_test.go:158: operations awaiting Zoom-local contracts = 1819, want 1815
--- FAIL: TestCoveredStreamsHaveReachableCommands (0.03s)
    command_surface_test.go:254: reachable direct_write operation commands = 1, want 5
--- FAIL: TestChatbotDirectWriteCommandsAreReachable (0.03s)
    command_surface_test.go:279: Preflight("chatbot messages send") = connector command "chatbot messages send" is blocked: unknown command, want declared executable Chatbot action
    command_surface_test.go:279: Preflight("chatbot messages edit") = connector command "chatbot messages edit" is blocked: unknown command, want declared executable Chatbot action
    command_surface_test.go:279: Preflight("chatbot messages delete") = connector command "chatbot messages delete" is blocked: unknown command, want declared executable Chatbot action
    command_surface_test.go:279: Preflight("chatbot link-unfurls create") = connector command "chatbot link-unfurls create" is blocked: unknown command, want declared executable Chatbot action
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/zoom
FAIL
```

The captured RED state is test-only, will be committed and pushed as its own checkpoint, and
contains no provider credential or token value.

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

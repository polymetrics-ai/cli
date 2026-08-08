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

## GREEN foundation — OAuth client credentials with HTTP Basic client auth

Commit `c3038e29c` adds a narrow reusable `client_auth: basic` option to the existing
`oauth2_client_credentials` contract. Empty or `form` retains the established form-post behavior;
`basic` moves only client ID/secret into HTTP Basic and leaves `grant_type`, scope, and declared
extra parameters in the form. Unsupported styles, and `client_auth` on non-client-credentials
auth modes, fail static validation.

The RED raw-bundle test now loads the declared field and proves the exact token-wire contract:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/connsdk ./internal/connectors/engine
ok  polymetrics.ai/internal/connectors/connsdk
ok  polymetrics.ai/internal/connectors/engine
```

The test asserts that client credentials are absent from the token request form and present only
in Basic, then confirms the received access token becomes the API request bearer header. All test
values are synthetic and no value is emitted. This foundation is independent of and precedes the
Zoom Chatbot declaration; it unblocks any connector whose documented OAuth client credentials use
the standard Basic client-auth style.

## GREEN foundation — closed typed JSON object command input

Commit `68dc984fe` adds `json_object` as a deliberately closed command flag type. It accepts one
object only, uses a number-preserving decoder, and rejects a scalar, array, malformed input, or a
second JSON document. The existing generic `json` type remains an explicit runtime rejection.

The commandrunner additionally refuses this type for path/query bindings, while connectorgen
requires its command declaration to map to an operation body and checks compatibility with the
linked object-typed body-schema field. The operation executor then validates the assembled closed
body schema before issuing a request; this is not a generic raw-body escape hatch.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/commandrunner ./cmd/connectorgen
ok  polymetrics.ai/internal/connectors/commandrunner
ok  polymetrics.ai/cmd/connectorgen
```

The foundation is separate from Zoom authoring and unblocks any declared provider operation whose
named, schema-owned member is an object rather than a scalar or string array.

## RED contract — executable Chatbot fixture lifecycle

Before declaring any Zoom Chatbot operation, four synthetic response fixtures and the
`TestChatbotCommandsExecuteWithFixture` contract were added. The test builds each real command
before creating a fixture credential, then exercises the complete no-network plan/preview and
approved execution lifecycle against separate token/API loopback servers. It requires HTTP Basic
client credentials at the token endpoint, Bearer auth at the API endpoint, exact request
method/path/body, DELETE typed confirmation, response redaction, and Link Unfurls' successful
`204` with no invented response body.

The focused Zoom suite failed against the still-undeclared Chatbot slice as expected:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
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
--- FAIL: TestChatbotCommandsExecuteWithFixture (0.03s)
    command_surface_test.go:1502: BuildWriteCommand(send) = connector command "chatbot messages send" is blocked: unknown command, want declared Chatbot command
FAIL
FAIL	polymetrics.ai/internal/connectors/defs/zoom	2.783s
FAIL
```

This RED checkpoint contains only test/fixture/evidence changes. It does not declare a Zoom
operation, add a credential field, or alter generated output.

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

## RED/GREEN foundation — true empty bodies for declared status actions

The Chatbot fixture found a general executor defect after its declaration was exercised: a typed
nil `map[string]any` passed through `Requester.DoLimited` as a non-nil interface and serialized as
the JSON literal `null`. That violates a documented no-body operation even when the response is
correctly status-only. Before changing the executor, the existing operation-scoped-origin test was
strengthened to assert both an empty payload and absent `Content-Type`:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth$'
--- FAIL: TestOperationDirectWriteUsesDeclaredOperationOriginAndAuth (0.00s)
    direct_write_test.go:241: no-body direct-write payload = "null", want no payload
FAIL
FAIL    polymetrics.ai/internal/connectors/engine
FAIL
```

The red assertion was committed and pushed as `b81cefb78`. Commit `acbf7405c` then split the
direct-write dispatch: JSON operations retain their declared body, while `format=none` passes an
untyped `nil` to the requester. This preserves the ordinary requester's no-body behavior and does
not add a generic raw transport.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok      polymetrics.ai/internal/connectors/engine
```

This is a reusable foundation for every declared POST/PUT/PATCH/DELETE action whose contract has
no request body; it directly unblocks Zoom Chatbot Link Unfurls and future status-only actions.

## GREEN connector — four declared Chatbot operations

The four provider operations now have `rest_write` contracts, dedicated client-credential fields,
closed typed body schemas, command paths, generated command metadata, and synthetic fixtures.
`connectorgen surface-sync` generated every derivable `api_surface`, `output_policy`, path mapping,
and response-cap field; `surface-reconcile --notes-contains provider_module=chatbot` changed only
the four Chatbot ledger rows to `covered_by.direct_write`.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/defs/zoom/...
ok      polymetrics.ai/internal/connectors/defs/zoom

$ go run ./cmd/connectorgen validate internal/connectors/defs/zoom
connectorgen validate: 1 connector(s) checked, 0 findings

$ go run ./cmd/connectorgen surface-sync --check
connectorgen surface-sync: 551 connector(s) scanned, 0 field(s) filled and 0 field(s) corrected across 0 connector(s)

$ go run ./cmd/connectorgen surface-reconcile --check --notes-contains provider_module=chatbot
connectorgen surface-reconcile: 551 connector(s) scanned; covered=0 blocked=0 unchanged=0 refused=0
```

The isolated lifecycle fixture performs one real token exchange and one action request for each
command against loopback servers. It proves HTTP Basic client authentication with the client ID and
secret absent from the form, Bearer use at the action endpoint, exact method/path/body, no
pagination input, plan/preview no-network behavior, DELETE typed confirmation, redaction, and
Link Unfurls' `204` status-only result. All fixture identifiers are synthetic; no secret or token
value is emitted.

## RED/GREEN foundation — redact typed path inputs in direct-write errors

Manual review found that `json_redacted` correctly removed declared response/body values but a
transport error could still preserve a declared path parameter inside its request URL. The existing
error-redaction test was extended with a typed `message_id` path binding before production code
changed:

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine -run '^TestOperationDirectWriteJSONRedactedErrorsHideDeclaredRequestAndResponseFields$'
--- FAIL: TestOperationDirectWriteJSONRedactedErrorsHideDeclaredRequestAndResponseFields (0.00s)
    direct_write_test.go:416: json_redacted direct-write error exposed sensitive request or response content
FAIL
FAIL    polymetrics.ai/internal/connectors/engine
FAIL
```

The red checkpoint is `c9c89c707`. Commit `070432f40` collects typed path values into the
already-established direct-write literal-redaction set. `json_redacted` still applies its strict
JSON error policy; other output policies retain their established complete provider diagnostics
except for those declared path/body literals. This avoids changing general error semantics merely
to protect an identifier in a URL.

```text
$ go test -count=1 -timeout 20m ./internal/connectors/engine
ok      polymetrics.ai/internal/connectors/engine
```

The foundation protects Chatbot message, user, and trigger path values as well as future typed
direct-write operations. It does not create a raw HTTP capability or expose credential material.

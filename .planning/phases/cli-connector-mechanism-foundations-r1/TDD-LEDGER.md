# TDD Ledger: Dual-Mechanism Connector Foundations (P0)

Branch: `fm/cli-connector-mechanism-foundations-r1`

## Manual GSD Fallback

The project-local GSD adapter did not expose `programming-loop`; see `PLAN.md`.
This ledger records the manual plan → red → green → verification cycle.

## Recovery Boundary

The branch contained previous commits when it was recovered. No red result is
invented for that earlier work. This ledger covers only the new recovery-audit
hardening slice below.

## Cycle 1: Browser-session credential outcome

Red test to add before implementation:

```bash
go test ./internal/browserauth/driver -run 'TestFlowLogin|TestNewFlowRejects'
```

Expected failure: `driver.Flow` / `NewFlow` is absent, so a real controlled
browser session cannot satisfy the top-level `browserauth.Flow` contract and
return `browserauth.Credential{Session: ...}`.

Red result: failed to compile with `undefined: Flow`, `undefined: NewFlow`,
and `undefined: FlowConfig`.

Green criteria:

- It returns only `RequiredCookies`.
- It applies the declared origin and optional CSRF-cookie mapping.
- It records a browser-resolution fingerprint.
- It closes the controlled browser session on success and error.
- Its public browser-session interface remains navigation/wait/cookie/close
  only; no password-entry capability is introduced.

Green result:

```bash
go test ./internal/browserauth/driver -run 'TestFlowLogin|TestNewFlowRejects|TestNoTypingSurface|TestSessionInterfaceIsMinimal'
```

Passed. The opt-in real-browser flow test also passed with
`POLYMETRICS_BROWSER_INTEGRATION=1`.

## Cycle 2: Loopback redirect confinement

Red test to add before implementation:

```bash
go test ./internal/browserauth/loopback -run TestNewRejectsNonLoopbackRedirectHost
```

Expected failure: a configured non-loopback host is currently accepted even
though the package promises a loopback authorization-code callback.

Red result: `0.0.0.0`, `192.0.2.25`, `example.invalid`, and `localhost` were
all accepted.

Green criteria: only literal loopback IP addresses are accepted; the default
`127.0.0.1` and IPv6 `::1` remain valid.

Green result:

```bash
go test ./internal/browserauth/loopback -run 'TestNewRejectsNonLoopbackRedirectHost|TestNewAllowsLiteralLoopbackRedirectHost'
```

Passed.

## Cycle 3: Credential-shape and expiry boundary

Red test to add before implementation:

```bash
go test ./internal/browserauth -run 'TestLoginRejectsInvalidCredentialShape|TestCredentialNeedsReauthentication'
```

Expected failure: the documented top-level `browserauth.Login` function
forwards a flow result without enforcing its "exactly one outcome" contract,
and session expiry has no common re-authentication decision helper.

Red result: failed to compile because `Credential.NeedsReauthentication` did
not exist.

Green criteria: no-flow-result and dual-result credentials are rejected;
official token expiry and a session expiry hint can tell callers to run the
same browser-auth flow again, while unknown expiry remains non-expired rather
than spuriously forcing a login.

Green result:

```bash
go test ./internal/browserauth -run 'TestLoginRejectsInvalidCredentialShape|TestCredentialNeedsReauthentication'
```

Passed.

## Cycle 4: Public mechanism metadata completeness

Red test to add before implementation:

```bash
go test ./internal/connectors/engine -run TestMechanismSynthesisPreservesWebGovernanceMetadata
```

Expected failure: the engine parses the web mechanism's upstream pin,
breakage-review cadence, and disabled reason but its connector-facing
metadata/definition projection silently omits them. That contradicts the
documented promise that inspect surfaces the full declared mechanism block.

Red result: failed to compile because connector-facing `MechanismSpec` had no
`UpstreamPin`, `BreakageReviewCadenceDays`, or `DisabledReason` fields.

Green criteria: public metadata, definition/catalog JSON, and the connector
manual retain the same declared governance values without renderer-specific
string copies.

Green result:

```bash
go test ./internal/connectors/engine -run TestMechanismSynthesisPreservesWebGovernanceMetadata
go test ./internal/connectors -run TestMechanismSectionRendersWebGovernanceMetadata
```

Passed.

## Final Results

The focused red/green cycles above passed. The broader local verification is
recorded in `VERIFICATION.md`; surface-sync remains explicitly unperformed
until the required no-mistakes rebase onto current `main`.

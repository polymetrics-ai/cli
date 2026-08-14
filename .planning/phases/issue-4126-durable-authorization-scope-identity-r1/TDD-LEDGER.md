# TDD ledger — Issue 4126 durable authorization scope identity

| ID | Guarantee | Red assertion | Green proof |
| --- | --- | --- | --- |
| R1 | Content-free identity | Replacing a plan's source rows with different content, count, and timestamp can change a standing identity. | Real App scope derivation before and after the row replacement returns one equal identity because none of those payload fields is in `AuthorizationScope`. |
| R2 | Exact bound scope | A changed scope property can silently retain standing authorization. | Each named bound property changes the canonical identity. |
| R3 | Unattended repeat | Every reverse execution still requires a raw approval token. | The loopback GitHub provider receives two new writes after an empty-token run following the first approved dispatch; no external destination is needed. |
| R4 | Fail closed before send | A changed scope can return an error after a write has been sent. | Changed mappings yield `AuthorizationScopeChangedError` and the individual loopback provider's request counter does not increase. |
| R5 | Revocable/expiring authority | Revoked or expired records can dispatch or collapse into a text-only reason. | Each durable-state case returns its distinct typed reason and its own loopback provider counter remains unchanged. |
| R6 | Exactly-once bootstrap | The original token can create more than one standing authorization or send. | Replay yields `AuthorizationTokenReplayError`; the loopback send count and record count remain unchanged. |
| R7 | No material leak | Record serialization can include a token, secret, raw credential, or raw destination config. | State and marshaled record assertions exclude the actual fixture secret/token while the safe opaque reference and digest remain. |

## Red command

```sh
go test -timeout 20m ./internal/app -run 'Test(AuthorizationScope|RunReverseETL.*Authorization)' -count=1
```

The initial compile failure is retained at `traces/red-run.txt`; no production
authorization implementation existed at red time.

## Green command

```sh
go test -timeout 20m ./internal/app -run 'Test(AuthorizationScope|RunReverseETL.*Authorization)' -count=1
```

Result: passed after the green implementation, including the observable
loopback provider-send assertions for every refusal path.

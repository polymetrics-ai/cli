# API contract — destructive write gate

## Bundle declaration

Both `writes.json` actions and `operations.json` operations may declare:

```json
{
  "confirmation": {
    "kind": "destructive"
  }
}
```

The object is closed (`additionalProperties: false`) and `kind` is a closed enum containing only
`destructive`. Existing `confirm: "destructive"` declarations remain accepted and normalize to the
same runtime requirement; new code consumes the normalized type rather than raw strings.

Destructive intent is also inferred from typed metadata, so an omitted declaration cannot make a
DELETE safe. Inference lets existing undispositioned operation ledgers remain loadable while the
runtime refuses unsafe execution.

## Runtime evidence

The persisted authenticated grant contains:

- plan ID;
- immutable plan hash;
- engine-produced canonical prepared-request digest;
- concrete target digest and secret-safe credential revision;
- issue and expiry timestamps plus a one-shot nonce;
- approval-token hash and typed confirmation kind;
- HMAC-SHA256 authentication from a vault-derived key outside state JSON.

`internal/app` authenticates and atomically consumes the grant and approval token through
`JSONStore.Update`. It then creates opaque, execution-scoped evidence whose shared one-shot state
cannot be copied into a replay. The raw token is never copied into engine evidence, persisted JSON,
logs, fixtures, or errors.

## Gate seam

The engine exposes `PreparedWrite`: normalized target policy, exact canonical outbound requests,
complete definition/hook identity, and credential revision. Preview and execution use the same
prepared-write digest, and the wrapper consumes authenticated evidence before invoking the supplied
executor callback. Declarative `writes.json` and native Amazon SQS use the wrapper. The future
`rest_write` executor supplies its prepared requests and executor closure without changing the gate.

## Compatibility

- Safe writes retain current plan/approval behavior.
- Destructive plans add mandatory preview state before execution.
- Canonical public command names and operation IDs are unchanged.
- `batchable: false` is checked independently before plan storage and again before bulk execution.

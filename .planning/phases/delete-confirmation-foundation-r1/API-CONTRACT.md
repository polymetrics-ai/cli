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

Before preview, destructive plans persist a vault-authenticated plan seal containing plan identity,
mode, connector/action identity, credential and effective-configuration revisions, batchability,
typed confirmation, and an authority-generated lifetime.

The persisted authenticated grant contains:

- plan ID;
- immutable plan hash;
- engine-produced canonical prepared-request digest;
- concrete target digest, secret-safe credential/configuration revisions, and batchability;
- issue and expiry timestamps plus a one-shot nonce;
- approval-token hash and typed confirmation kind;
- HMAC-SHA256 authentication from a vault-derived key outside state JSON.

Grant expiry is derived from trusted process time and capped by the signed plan deadline; mutable
state timestamps do not extend it. `internal/app` authenticates and consumes the grant under a
locked revisioned state update. Before state replacement or dispatch, it creates an authenticated
create-exclusive marker under the project vault. It then creates opaque, execution-scoped evidence
whose shared one-shot state cannot be copied into a replay. The raw token is never copied into
engine evidence, persisted JSON, logs, fixtures, or errors.

## Gate seam

The engine exposes `PreparedWrite`: normalized target policy, exact canonical outbound requests,
complete definition/hook identity, and credential revision. Preview and execution use the same
prepared-write digest, and the wrapper consumes authenticated evidence before invoking the supplied
executor callback. Declarative `writes.json` and native Amazon SQS use the wrapper. The future
`rest_write` executor supplies its prepared requests and executor closure without changing the gate.

Production evidence is rooted only in the opened project vault. Caller-key authorities produce
explicitly untrusted evidence. Fixture evidence has a distinct MAC-bound scope and can authorize
only loopback prepared requests.

## Compatibility

- Safe writes retain current plan/approval behavior.
- Destructive plans add mandatory preview state before execution.
- Canonical public command names and operation IDs are unchanged.
- `batchable: false` is checked independently before plan storage and again before bulk execution.

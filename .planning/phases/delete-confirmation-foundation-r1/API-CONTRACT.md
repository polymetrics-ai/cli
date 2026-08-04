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

The shared approval evidence contains:

- plan ID;
- immutable plan hash;
- engine-produced preview digest;
- approval timestamp;
- typed confirmation kind.

The approval token is validated and consumed by `internal/app` and is never copied into engine
evidence, JSON output, logs, fixtures, or errors.

## Gate seam

The engine exposes a target/policy constructor and an execute wrapper. The wrapper validates all
evidence before invoking the supplied executor callback. Declarative `writes.json` execution uses
the wrapper. The future `rest_write` executor supplies its operation target, preview digest,
approval evidence, and HTTP executor closure without changing the gate.

## Compatibility

- Safe writes retain current plan/approval behavior.
- Destructive plans add mandatory preview state before execution.
- Canonical public command names and operation IDs are unchanged.
- `batchable: false` is checked independently before plan storage and again before bulk execution.

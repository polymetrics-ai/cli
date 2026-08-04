# Data model — destructive approval evidence

## Confirmation

`WriteConfirmation` is a value type with a closed `Kind`. The zero value means no typed
confirmation; `destructive` is the only valid non-zero kind.

## Persisted approval grant

`WriteApprovalGrant` is HMAC-authenticated with a vault-derived key that is never stored in state:

| Field | Purpose |
| --- | --- |
| `PlanID` | proves the write belongs to a stored plan |
| `PlanHash` | binds the approved destination, action, config, mappings, records, and payload files |
| `PreviewDigest` | binds approval to every canonical request, body, query, definition, and hook identity |
| `Target` | binds connector, operation, method, mutation class, concrete target digest, and credential revision |
| `ApprovalTokenHash` | binds the single token without persisting the raw value |
| `Nonce`, `IssuedAt`, `ExpiresAt` | makes the grant unique and time-bounded |
| `Confirmation` | satisfies the target's closed confirmation requirement |
| `MAC` | detects modification or substitution of every preceding field |

`WriteApprovalEvidence` is an opaque in-memory capability produced only after grant authentication
and atomic state consumption. Copies share one atomic use marker, so only one executor callback can
consume it.

## Reverse plan state

Destructive plans persist preview time, digest, and the authenticated grant. None contains a secret
value. The raw approval token is returned only to the human-facing lifecycle step and only its hash
is stored. The transition to `executing` clears the token hash and grant in one locked reload/update.

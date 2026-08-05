# Data model — destructive approval evidence

## Confirmation

`WriteConfirmation` is a value type with a closed `Kind`. The zero value means no typed
confirmation; `destructive` is the only valid non-zero kind.

## Persisted approval grant

`WritePlanSeal` authenticates the planned lifetime and every pre-preview semantic input: plan ID
and hash, mode, connector/action, credential revision, effective-configuration digest,
batchability, scope, and typed confirmation.

`WriteApprovalGrant` is HMAC-authenticated with a vault-derived key that is never stored in state:

| Field | Purpose |
| --- | --- |
| `PlanID` | proves the write belongs to a stored plan |
| `PlanHash` | binds the approved destination, action, config, mappings, records, and payload files |
| `PreviewDigest` | binds approval to every canonical request, body, query, definition, and hook identity |
| `Target` | binds connector, operation, method, mutation class, concrete target digest, credential/configuration revisions, batchability, scope, and confirmation |
| `ApprovalTokenHash` | binds the single token without persisting the raw value |
| `Nonce`, `IssuedAt`, `ExpiresAt` | makes the grant unique and time-bounded |
| `Confirmation` | satisfies the target's closed confirmation requirement |
| `MAC` | detects modification or substitution of every preceding field |

`WriteApprovalEvidence` is an opaque in-memory capability produced only after grant authentication,
revisioned state consumption, and creation of the vault consumption marker. Copies share one atomic
use marker, so only one executor callback can consume it.

## Reverse plan state

Destructive plans persist the plan seal, preview time, digest, and authenticated grant. None contains
a secret value. `state.Revision` makes every whole-state replacement compare-and-swap. The raw
approval token is returned only to the human-facing lifecycle step and only its hash is stored. The
transition to `executing` records the nonce under the vault and clears the token hash and grant in
one locked reload/update.

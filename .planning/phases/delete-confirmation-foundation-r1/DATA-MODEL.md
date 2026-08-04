# Data model — destructive approval evidence

## Confirmation

`WriteConfirmation` is a value type with a closed `Kind`. The zero value means no typed
confirmation; `destructive` is the only valid non-zero kind.

## Approval evidence

`WriteApprovalEvidence` binds execution to:

| Field | Purpose |
| --- | --- |
| `PlanID` | proves the write belongs to a stored plan |
| `PlanHash` | binds the approved destination, action, config, mappings, records, and payload files |
| `PreviewDigest` | binds approval to the engine dry-run preview of the same action and records |
| `ApprovedAt` | distinguishes explicit approved execution from planned/previewed state |
| `Confirmation` | satisfies the target's closed confirmation requirement |

## Reverse plan state

Destructive plans persist preview time and digest. The digest is not secret. The raw approval token
is returned only to the human-facing lifecycle step and only its hash is stored.

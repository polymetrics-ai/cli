# Refs #4303 — Canonical repair R2 TDD ledger

Every production change begins with a focused production-shaped test added in the same group. `Red:` and `Green:` entries are filled with the exact command and result as work proceeds.

| IDs | Red proof | Green proof | Status |
| --- | --- | --- | --- |
| AB-B01 | Preview/JSON/digest omit a paired tombstone DELETE but apply sends it. | Exact ordinary and delete actions with destructive semantics are shown and digest-bound before token issuance. | planned |
| AB-B02 | A valid multi-action declaration mutates through action A then fails descriptor-wide read-back fields for action B. | Each action owns compatible read-back; incompatible identity/fields fail before I/O. | planned |
| AB-B03 | Missing-ok delete reports unchanged then fails to read absence/checkpoint. | Complete unchanged+written accounting independently reads absence and checkpoints once without replay. | planned |
| AB-B04 | Binding/read-back maxima mismatch performs writes before receipt/read-back refusal. | The sealed unit respects all action/read-back bounds before provider I/O. | planned |
| AB-B05 | Legal escaped locators overflow encoded receipt after provider writes. | Pre-I/O sizing splits/refuses and safe escaped receipts complete. | planned |
| AB-B06 | Two individually legal receipt parts overflow their composite envelope post-write. | Combined budget reserves exact envelope overhead before either write. | planned |
| AB-B07 | A helper-accepted composite exceeds the acknowledgement’s private limit post-write. | One canonical bound applies to construction and attachment before I/O. | planned |
| AB-B08 | Provider success plus local post-success failure loses destination results. | Ordered sanitized provider results survive the failure chain and no checkpoint is claimed. | planned |
| AB-B09 | Equal payload/index from distinct durable worksets sends the same key. | Same-workset retry is stable; separate worksets send distinct keys. | planned |
| AB-W01 | Omitted legacy batch serializes as invalid zero batch. | Absence is omitted; supplied generic batch semantically validates. | planned |
| AB-W02 | Structural schema accepts neither/both record and tombstone mappings. | Schema admits exactly one mapping mode. | planned |
| AB-W03 | Eligible ghost action loads/advertises but has no strategy. | Exact action/strategy closure fails before selection or I/O. | planned |
| AB-W04 | Orphan binding loads despite unreachable strategy/write action. | Every binding belongs to selected reachable strategy and write-set. | planned |
| AB-W05 | Help/manual/website omit clamp and tombstone wording. | Exact semantics are asserted across all operator surfaces. | planned |

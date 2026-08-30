# Vercel Track A context

Issue #4421 is the Vercel child of Batch R1 parent #4325. This Track A slice
records source-accounted mapping evidence only; it neither promotes a lane to
runtime execution nor changes a connector definition's executable surface.

The immutable provider-fact authority is the schema-v2 Vercel source lock at
`internal/connectors/defs/vercel/sources/vercel-operation-source-lock.json`.
The existing crosswalk, API surface, streams, writes, and sync transport files
are projection backlinks only. They cannot add, remove, or certify source rows.

The runner has no compatible isolated GSD role runtime and the active task
contract forbids spawning roles. The documented GSD inline/manual fallback is
therefore used for discussion, planning, execution, verification, and review.

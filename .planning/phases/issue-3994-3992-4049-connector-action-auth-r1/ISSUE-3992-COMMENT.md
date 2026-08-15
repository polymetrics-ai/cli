Captain scope correction recorded for this delivery:

We are not adding a durable schedule-authorization object or per-firing derived approval token.
Approval belongs once to the existing ETL/reverse-ETL job at connection + schema + preview
granularity. A flow may compose only positively resolved existing jobs (an action job must already
have standing approval), and a schedule may target only a positively resolved existing flow. Both
boundaries fail with typed errors before writing a flow, schedule, crontab entry, or sentinel.

The installed payload remains exactly `pm --root <root> flow run <name> --json`; it carries no
approval token or authority. On each unattended firing the binary reloads the referenced jobs and
revalidates credential revision/config digest, manifest digest, source scope, mappings,
destination/action, confirmation policy, expiry, and revocation. Drift/refusal/ambiguous delivery
parks the existing durable schedule fire state and never auto-replays a potentially non-idempotent
write.

Reason: the schedule inherits a flow composed from already-approved jobs, so a second authority
model would duplicate approval and create another bearer artifact to store, expire, and protect.
Keeping authority on the job provides the approve-once workflow while per-firing drift validation
keeps unattended execution fail-closed.

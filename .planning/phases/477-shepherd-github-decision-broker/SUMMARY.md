# Summary: #477

Status: implementation complete; declared child verification equivalent green; ready stacked PR.

The owned slice adds a pure durable human-decision aggregate and a typed GitHub comment broker
without controller wiring. The broker routes issue-vs-PR gates deterministically, binds every
request to the exact repository/target/generation/head, owns one idempotency marker, accepts one
exact allowlisted unedited human command, persists minimal accepted evidence, and consumes it once.

The file repository uses strict schemas, private files, no-follow reads, atomic replacement+fsync,
exclusive transactions, and dead-process lock recovery. Retry/restart recovers an externally
created marker comment before making another write. The GitHub adapter is request-specific, uses
ambient host auth only, bounds argv execution/output/pagination, and never accepts a token field or
generic comment body from its caller.

Bots (including `name[bot]`), edited comments, non-allowlisted actors, malformed/multiline/unknown
commands, duplicate markers, multiple decisions, stale bindings, expired/future responses, emoji,
review text, CI state, and silence do not authorize anything. Parent merge is a separate
`parent_merge` exact-head gate with its own immutable marker.

Live comment mutation remains skipped without a designated sandbox. Parent-merge approval remains a
separate exact-head human gate and is never inferred from automated review, CI, reactions, silence,
or generic review text. Final evidence: focused 26 pass + 1 sandbox skip; full Shepherd 163 pass +
1 sandbox skip; strict TypeScript, Pi RPC discovery, and diff check pass. The parent intentionally
superseded and cancelled the child `make verify` run; supplemental standalone Go gates passed.

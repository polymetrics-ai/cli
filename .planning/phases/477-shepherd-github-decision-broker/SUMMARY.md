# Summary: #477

Status: the final lock-snapshot review correction is implemented and locally verified; fresh
parent-owned xhigh review is pending on the new exact head.

The owned slice adds a pure durable human-decision aggregate and a typed GitHub comment broker
without controller wiring. The broker routes issue-vs-PR gates deterministically, binds every
request to the exact repository/target/generation/head, owns one idempotency marker, accepts one
exact allowlisted unedited human command, persists minimal accepted evidence, and consumes it once.

The file repository uses strict schemas, private files, no-follow reads, atomic replacement+fsync,
and token-named atomically published locks with deterministic acquisition and token-fenced reclaim/
release. Retry/restart recovers an externally created marker comment before making another write.
The GitHub adapter is request-specific, uses ambient host auth only, bounds argv execution/output/
pagination, classifies redacted transient/permanent failures, and retries only transient failures
within a separate capped backoff policy.

Bots (including `name[bot]`), edited comments, non-allowlisted actors, malformed/multiline/unknown
commands, duplicate markers, multiple decisions, stale bindings, expired/future responses, emoji,
review text, CI state, and silence do not authorize anything. Parent merge is a discriminated
`parent_merge` exact-head gate whose only affirmative command is `approve-merge`. Questions reject
credential/control/bidi/mention spoofing and render as escaped quoted Markdown; only configured,
validated humans are mentioned on the dedicated allowlist line.

Independent review of head `87eb80f561d416da245e753a5dbc887a3384a05d` identified timestamp/ID
validation, numeric identity, lock fencing, parent-merge typing, display/secret safety, allowlist
mentioning, and transport retry gaps. The correction was test-first: 23 pass/15 fail/1 skip at RED,
then final focused 41 pass/0 fail/1 skip. The complete Shepherd suite is 178 pass/0 fail/1 skip;
strict TypeScript over all 11 production modules, offline Pi 0.80.6 RPC discovery, immutable-base,
owned-scope, and diff checks pass.

Live comment mutation remains skipped without a designated sandbox. Parent-merge approval remains a
separate exact-head human gate and is never inferred from automated review, CI, reactions, silence,
or generic review text. Per coordinator policy, no Go, connector, `make verify`, Claude/Copilot,
live GitHub, or merge action was run in the correction cycle.

Fresh review of `f5a4dc68a7b76f708858542a7190ca3d1f375044` found that benign lock-file
rename/release between `readdir`, `lstat`, and owner read can surface as a transaction failure under
contention. The follow-up preserves the token-fenced design while discriminating valid, missing,
and invalid owner reads. Vanished/transitioned snapshots rescan only through the existing slept,
attempt-bounded acquisition loop; stable malformed locks still fail closed.

The test-first correction recorded 18 pass/1 expected failure at RED, exposing both `ENOENT` and
`invalid-owner`, then reached focused GREEN at 43 pass/0 fail/1 skip. The complete Shepherd suite is
180 pass/0 fail/1 skip; strict owned and all-production TypeScript, offline Pi 0.80.6 RPC, and
full-range diff/base/scope checks pass. No broad, live, review-bot, or merge gate was run.

# Summary: #477

Status: planned; implementation pending.

The owned slice will add a pure durable human-decision aggregate and a typed GitHub comment broker
without controller wiring. The broker will route issue-vs-PR gates deterministically, bind every
request to the exact repository/target/generation/head, own one idempotency marker, accept one exact
allowlisted unedited human command, persist the minimal accepted evidence, and consume it once.

Live comment mutation remains skipped without a designated sandbox. Parent-merge approval remains a
separate exact-head human gate and is never inferred from automated review, CI, reactions, silence,
or generic review text.

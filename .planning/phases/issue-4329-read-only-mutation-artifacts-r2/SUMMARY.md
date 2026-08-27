# Summary — issue #4329, r2

The shared source-import path now turns a retained, cited provider mutation
into the existing non-executable mutation artifact only when the connector
explicitly declares no write capability and no complete declaration-owned
action exists. It leaves real executable delete/reverse-ETL actions intact.

The tests load byte-identical full Sentry/Vercel source locks, retain each
classified mutation, validate real read/mutation pairs, and—after a clean
merge of current main—prove their exact source-locked delete operations stay
implemented reverse-ETL actions when their complete declared foundation exists.
The code and tests change no connector bundle or user command; source-bound
materialization of those connectors remains the named downstream gap to a
binary credential probe.

Local verification is recorded in `VERIFICATION.md`; remote PR/audit/CI gates
remain pending.

Audit M1 additionally closes the omitted-member loophole: automatic artifacts
now require a JSON-explicit `capabilities.write: false`, rather than treating
an absent field as an opt-out. The internal presence marker is not a public
capability and therefore does not alter certification output.

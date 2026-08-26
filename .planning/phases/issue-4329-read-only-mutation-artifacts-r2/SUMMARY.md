# Summary — issue #4329, r2

The shared source-import path now turns a retained, cited provider mutation
into the existing non-executable mutation artifact only when the connector
explicitly declares no write capability and no complete declaration-owned
action exists. It leaves real executable delete/reverse-ETL actions intact.

The tests load byte-identical full Sentry/Vercel source locks, retain each
classified mutation, and validate real read/mutation pairs. The code and tests
change no connector bundle or user command; source-bound materialization of
those connectors remains the named downstream gap to a binary credential probe.

Local verification is recorded in `VERIFICATION.md`; remote PR/audit/CI gates
remain pending.

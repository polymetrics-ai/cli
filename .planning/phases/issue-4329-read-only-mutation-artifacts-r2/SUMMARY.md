# Summary — issue #4329, r2

The shared source-import path now turns a retained, cited provider mutation
into the existing non-executable mutation artifact only when the connector
explicitly declares no write capability and no complete declaration-owned
action exists. It leaves real executable delete/reverse-ETL actions intact.

Sentry and Vercel source-lock citations are regression vectors. The code and
tests change no connector bundle or user command; source-bound materialization
of those connectors remains owned downstream work.

Local verification is recorded in `VERIFICATION.md`; remote PR/audit/CI gates
remain pending.

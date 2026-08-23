# Summary — Issue 4316 operation-route foundation

Implemented a closed, named route declaration under `streams.base.routes`, selected only by a
stream, write action, or operation declaration. Selected version routes resolve a declared origin
from the established connector base configuration and retain the source-locked version in the
operation path; callers receive no route or URL input.

The shared resolver is used by declarative ETL, reverse ETL writes (including binary uploads),
operation direct read/write, and binary downloads. Unknown, conflicting, invalid, and
version-mismatched routes fail before provider I/O with `MissingOperationRouteError`, including the
operation source URL when present.

Help Scout now declares `mailbox_v3`, enables its five v3 reads, and preserves its stored `/v2`
default. Generated command, endpoint-ledger, and website projections are current.

Manual GSD fallback: the project Pi adapter emits official prompts but cannot provide its expected
isolated runtime agents; this phase was executed and reviewed inline with the recorded red/green
and local verification evidence.

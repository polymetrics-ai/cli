# Google Calendar parity resume — focused verification complete

Recovered PR #3554 is now active through the declarative engine rather than shadowed by the legacy native connector. Google Calendar API v3 Discovery revision `20260731` documents 38 operations exactly: 11 GET, 15 POST, 4 PATCH, 4 PUT, and 4 DELETE.

Twelve operations are genuinely reachable: 11 fixture-backed GET streams and the bounded typed `freeBusy.query` direct read. The other 26 are represented as blocked operation-ledger rows, not CLI commands and not `planned`, because `rest_write` remains schema-only with no command-runner dispatch. The connector must not be called complete for executable-operation parity until that foundation exists.

Citation coverage is 25/25 declared request-field uses and 38/38 operation-level sources. Focused validation, conformance, command preflight, CLI, build, generated docs/website, lint, smoke, boundary, and release-workflow gates passed. The branch is ready for the required no-mistakes handoff.

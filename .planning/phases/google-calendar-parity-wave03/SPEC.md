# Google Calendar parity resume specification

Expose every documented Google Calendar v3 operation in the connector ledger. The 11 GET operations and one bounded POST free/busy read must be executable through the current declarative runtime. The remaining 26 documented mutations must be recorded as blocked, not planned or implemented, because the only compatible executor kind is `rest_write` and commandrunner does not dispatch it.

Every declared request input must have provider-owned evidence in the phase research matrix. No live credentials or provider calls are required.

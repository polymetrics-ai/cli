# Google Calendar parity resume specification

Expose every documented Google Calendar v3 operation through the current declarative runtime. The 11 GET operations and one bounded POST free/busy read are executable reads. Each of the other 26 documented mutations is a typed `writes.json` reverse-ETL action with a record schema, fixture, risk/confirmation metadata, source citation, and reachable CLI plan/preview/approval/execute surface. No operation uses the unavailable `rest_write` executor.

Every declared request input must have provider-owned evidence in the phase research matrix. No live credentials or provider calls are required.

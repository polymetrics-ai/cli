# Summary — Issue #151 Chatwoot stream runner

Status: planned.

Planned slice: add fixture-backed stream runner coverage for all seven Chatwoot ETL streams, correct message pagination to official `after` cursor semantics, and add a narrow fan-out parent-pagination override to the declarative engine so parent conversation sweeps can remain page-number based while child message reads use cursor pagination.

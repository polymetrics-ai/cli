# CONTEXT — issues #3745 and #3746 truthful changefeed discovery

## Phase mapping

This foundation implements contract slice #3745 and discovery slice #3746 for parent #2986,
in that dependency order. It is one shared-core foundation branch, with PostgreSQL as the only
connector-specific evidence target. The branch is
`fm/cli-found-changefeed-contract-r1`.

## Locked decisions

- A changefeed is an evidence-backed executable contract, never a `CDCReader` type assertion or a
  hand-maintained boolean. The public `cdc` catalog filter remains a compatibility spelling but
  returns only executable, implemented changefeeds.
- The closed mechanism taxonomy is: `logical_replication`, `incremental_cursor`, `webhook`,
  `event_stream`, and `polling_watermark`. Status is separately closed as `implemented`,
  `planned`, `unsupported`, or `unknown`.
- The descriptor lives in an optional dedicated `changefeed.json`, adjacent to `metadata.json`.
  It carries source artifact URL/version/retrieval date, status, mechanism, executor identity
  when implemented, checkpoint/recovery details, and delivery guarantees. Unsupported entries
  retain an evidence source and concise reason but do not claim an executor.
- The new capability interface is distinct from the legacy `CDCReader`. Existing CDC types stay
  available only for migration compatibility; they cannot make a connector visible as CDC.
- A catalog entry is implemented CDC only when its descriptor says `implemented` and a registered
  executor reports a matching descriptor. Missing either factor is fail-closed.
- PostgreSQL is explicitly `unsupported` in this slice. Its decoder and its gated
  `ReadCDC` stub are not a replication executor. No dependency, slot lifecycle, provider call,
  or local PostgreSQL integration run is authorized.

## Scope fences

- Do not classify or edit the remaining connector fleet. An absent descriptor remains safely
  unknown/non-capable until the provider-artifact survey supplies measured evidence.
- Do not implement a PostgreSQL CDC executor, add `pglogrepl`, call a provider, or use credentials.
- Do not add #3747 positive conformance fixtures, #3748 manual/docs/website work, or #3749
  connectorgen/generation enforcement. The CLI surface in #3746 is limited to catalog/inspect
  JSON projection required to explain this negative result.
- Do not edit `internal/connectors/commandrunner/runner.go`, shared write/redaction regions, or
  unrelated native connectors.
- Do not introduce any masking/redaction path, `redact_fields`, or redacting output policy.

## Evidence read before design

- Parent #2986 and sub-issues #3745–#3749 via `gh-axi issue view <n> --full`.
- `data/cli-foundations-research-cdc-and-gaps-r1/report.md`, §1 (lines 115–357): taxonomy,
  descriptor shape, executor matching, recovery and guarantee requirements.
- `data/cli-database-query-and-changefeed-parity-r1/report.md` (all lines 1–483): PostgreSQL's
  current status is unsupported; its logical decoder is not executable CDC.
- `docs/migration/HANDOFF-CODEX.md`, `docs/migration/conventions.md`, and
  `docs/architecture/connector-architecture-v2-design.md` in full.

## GSD execution note

The adapter was healthy. Commands are executed inline because the assigned firstmate worker brief
requires an autonomous single worker and does not authorize role spawning. This is the documented
Pi-adapter fallback and does not weaken TDD, verification, or review.

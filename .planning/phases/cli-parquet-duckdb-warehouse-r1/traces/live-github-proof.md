# Live GitHub proof — Parquet warehouse and reverse ETL round trip

Run against the **live GitHub REST API**, not fixtures. Repository: `golang/example` (public).
Token obtained from `gh auth token`; never printed, never logged, redacted from every captured
stream.

## Read-only discipline

The token carries `repo` scope, which grants write, so the exercise was constrained to reads:

- Every `pm` call against GitHub was a **source read** (`pm etl run`), which issues GETs only.
- The reverse-ETL **destination was the local `outbox` connector**, writing
  `.polymetrics/outbox/prs_to_outbox.jsonl`. The *source* is the live-derived Parquet table; the
  *write target* is a local file. No GitHub object was created, modified or deleted.
- Confirmed after the fact by re-reading the repository: labels 11 → 11, pull requests 74 → 74,
  open issues 11 → 11. Unchanged.

Because the round trip was provable end to end with a local destination, no
`untested-because-write-scoped` gap remains. Writing *back into GitHub* was neither performed nor
needed: the reverse path under test is the read of a Parquet table and the mapping of its rows,
which is identical whatever the destination connector is.

## Ground truth, taken from the API independently of pm

```
gh api repos/golang/example/labels --paginate            -> 11 labels
gh api repos/golang/example/pulls?state=all&per_page=100 -> 74 pull requests
```

## ETL: live GitHub → Parquet

| connection | stream | sync mode | read | loaded | failed |
|---|---|---|---:|---:|---:|
| `gh_labels_sync` | `labels` | `full_refresh_overwrite` | 11 | 11 | 0 |
| `gh_prs_sync` | `pull_requests` | `incremental_append_deduped` | 74 | 74 | 0 |
| `gh_issues_sync` | `issues` | `incremental_append_deduped` | 0 | 0 | 0 |

The empty `issues` sync is **correct, not a defect**: GitHub's issues endpoint returns pull requests
too, and the connector filters them out. Of the 11 open "issues" on this repository,
`select(.pull_request == null) | length` is **0** — every one is a pull request. It also exercised
the zero-row path against live data: `tables/gh_issues.parquet` was still created and still listed.

On-disk layout after three live connections — each connection in its own directory, Parquet table
beside a JSONL write-ahead log and an ownership record:

```
warehouse/ws_.../github/conn_295cbea18bc87659/{owner.json, tables/gh_issues.parquet,        wal/issues.jsonl}
warehouse/ws_.../github/conn_529fd8195055a738/{owner.json, tables/gh_labels.parquet,        wal/labels.jsonl}
warehouse/ws_.../github/conn_.../             {owner.json, tables/gh_pull_requests.parquet, wal/pull_requests.jsonl}
```

## Assertion on the records that landed, not on exit status

Counts, from the Parquet tables through DuckDB: `gh_pull_requests` = **74**, `gh_labels` = **11** —
matching the API exactly.

Field values compared row by row against the API:

```
pull_requests: IDENTICAL — all 74 records match on node_id, number, updated_at, state
labels:        IDENTICAL — all 11 records match on name and color
```

Real values now living in Parquet:

```
MDExOlB1bGxSZXF1ZXN0Mjk5MDk2MTc3   13   2019-08-30T19:55:19Z   closed
MDExOlB1bGxSZXF1ZXN0MjU5OTM5MTgw   12   2019-03-11T12:17:48Z   closed
bug           fc2929
cla: yes      0e8a16
```

## Reverse ETL: Parquet → outbox, round trip asserted

```
plan    prs_to_outbox   source_table=gh_pull_requests  source_connection=gh_prs_sync  record_count=74
preview first sample    {"changed_at":"2017-07-10T20:28:47Z","external_id":"MDExOlB1bGxSZXF1ZXN0MTI5NzYzMDgz",
                         "pr_number":8,"pr_state":"closed"}
run     completed       staged=74  succeeded=74  failed=0
```

Every outbox record was then compared back against the live API:

```
outbox records: 74
records whose values disagree with the live API: 0
live API records absent from the outbox: 0
ROUND TRIP: IDENTICAL
github API -> pm ETL -> Parquet -> pm reverse ETL -> outbox, 74 records, values preserved
```

## The rebuild-from-WAL change, exercised against live data

Append modes used to stream into the table `O_APPEND`; they now replay the write-ahead log. The
risk that introduces is duplication or loss, so both modes were run repeatedly against live data.

```
append run 1: read/loaded=11 11   table rows=11   distinct names=11
append run 2: read/loaded=11 11   table rows=22   distinct names=11
append run 3: read/loaded=11 11   table rows=33   distinct names=11

deduped run 1: table rows=74   duplicated node_ids=0
deduped run 2: table rows=74   duplicated node_ids=0
```

Append grows by exactly 11 rows per run — append semantics preserved, nothing duplicated beyond
what append means and nothing lost. Deduped holds at 74 with zero duplicate primary keys across
repeated live syncs.

## One limitation observed, pre-existing and preserved

A table that synced **zero rows** is a zero-byte file, so `pm query run --sql "... from gh_issues"`
reports `Table with name gh_issues does not exist` rather than returning zero rows. `pm query run
--table gh_issues` does return zero rows.

This is not introduced here: the JSONL engine skipped zero-byte tables the same way, and an empty
Parquet file has no schema to register a view from. Inventing one would make the view claim columns
that were never observed. Recorded rather than silently "fixed" into something misleading.

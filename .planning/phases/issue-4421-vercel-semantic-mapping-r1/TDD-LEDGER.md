# TDD ledger — Issue #4421 Vercel semantic mapping repair

## Red

Planned failing assertion: the frozen matrix must report these source-cited lane cells as applicable and `mapped_unproven`, but it currently reports them `not_applicable`:

| Source operation | Lane | Source evidence |
| --- | --- | --- |
| `vercel.rest.artifactExists` | `direct_read` | HEAD description says it is equivalent to GET without a body. |
| `vercel.rest.artifactQuery` | `direct_read` | POST summary/description says Query and retained 200 JSON response. |
| `vercel.rest.readSessionFile` | `direct_read` | POST summary says Read and retained 200 response. |
| `vercel.rest.readSessionFile` | `binary_download` | Retained 200 response `application/octet-stream`, schema format `binary`. |
| `vercel.rest.writeSessionFiles` | `binary_upload` | Retained description requires `Content-Type: application/gzip` for gzipped tarball upload. |

## Green acceptance criteria

- The above cells are source-backed `mapped_unproven` with exact source-ID backlinks.
- `artifactQuery` and `readSessionFile` are not falsely direct-write/reverse-ETL candidates.
- A normal mutation POST remains a direct-write/reverse-ETL candidate and is not a semantic direct read.
- Missing 2xx response, absent/contradictory binary media, or state-changing operation wording rejects a semantic promotion.
- Matrix summary counts reconcile the 400-row lock and all seven cells per row.

## Execution evidence

The focused red run against the frozen target failed exactly as intended before the implementation:

```text
semantic lane vercel.rest.artifactExists direct_read = (not_applicable, not_applicable), want (applicable, mapped_unproven)
semantic lane vercel.rest.artifactQuery direct_read = (not_applicable, not_applicable), want (applicable, mapped_unproven)
semantic lane vercel.rest.readSessionFile direct_read = (not_applicable, not_applicable), want (applicable, mapped_unproven)
semantic lane vercel.rest.readSessionFile binary_download = (not_applicable, not_applicable), want (applicable, mapped_unproven)
semantic lane vercel.rest.writeSessionFiles binary_upload = (not_applicable, not_applicable), want (applicable, mapped_unproven)
```

The green test suite adds happy coverage for the three non-GET reads and both binary media lanes, a normal mutation POST negative, and edge negatives for absent 2xx response, contradictory mutation wording, JSON media carrying a binary schema, octet-stream without a binary schema, and gzip header text without binary payload documentation. Matrix validation checks each mapping backlink and recalculates the 400-row / 2,800-lane summary.

## Refactor constraints

- No fixed operation-ID decision table.
- No HTTP-method allow-list that admits arbitrary POST/HEAD reads.
- Reuse retained source summary, description, request/response media, schema, and 2xx response facts only.
- Do not add dependencies or modify shared code.

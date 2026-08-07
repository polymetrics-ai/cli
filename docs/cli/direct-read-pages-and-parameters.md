# Direct reads: pages and parameters

A direct read returns **one page** of a collection and tells you where that page
sits. It is page-wise exploration, not bulk extraction — `pm etl run` is the
path that stores what it reads; a direct read does not.

## Is this all of it?

Every direct read answers that question explicitly. In `--json`, the envelope
carries a `page` object:

```json
{
  "kind": "ConnectorCommandDirectRead",
  "response": [ ... ],
  "page": {
    "strategy": "page_number",
    "records": 100,
    "size": 100,
    "number": 1,
    "has_more": true,
    "next_number": 2,
    "complete": false,
    "reason": "more_pages"
  }
}
```

`complete` is the field to read. It is `true` only when the page is provably the
whole collection. When it is `false`, `reason` says why:

| reason | meaning |
| --- | --- |
| `more_pages` | the provider has at least one further page |
| `pagination_not_declared` | the connector declares no pagination strategy, so completeness cannot be proved from one request |
| `ambiguous_collection_shape` | the response holds more than one array and the paged one cannot be identified |

Without `--json`, the body is printed as before and a note goes to **stderr**,
so piping the body stays lossless:

```
note: page 1 of a paged result (100 records); more remain — rerun with --page 2
```

## Getting the rest

How you ask for the next page depends on what the API supports. The connector's
declared pagination strategy decides, and the result tells you which you have.

**Addressable page numbers** — `page_number` and `offset_limit` strategies:

```sh
pm github pulls files view --pull-number 3894 --json          # page 1, next_number: 2
pm github pulls files view --pull-number 3894 --page 2 --json # page 2
```

**Forward cursors only** — `cursor`, `next_url` and `link_header` strategies
have no page number to name. They hand back an opaque token instead:

```sh
pm gong logs list --logType Info --json                    # next_cursor: "30"
pm gong logs list --logType Info --page-cursor 30 --json    # the next page
```

Asking a strategy for navigation it cannot honour is **refused**, rather than
quietly answered with page one:

```
error: direct read pagination strategy "cursor" has no addressable page number;
       pass the previous page's next_cursor instead
```

`--page` and `--page-cursor` are mutually exclusive, and only direct reads
accept them.

## Where command flags come from

A direct-read command's flags are **derived from the connector's own provider
specification**, not hand-written. `pm <connector> <command> --help` lists every
parameter the endpoint accepts, with its type, its allowed values when the
specification declares an enum, and whether it is required:

```
FLAGS
  --direction (enum): The direction to sort the results by. values=asc|desc maps_to=query.direction
  --sort (enum): The property by which to sort the results. values=created maps_to=query.sort
  --tool-name (string): The name of a code scanning tool. maps_to=query.tool_name
```

Invalid values and missing required flags are rejected **before any network
call**:

```sh
$ pm github code-scanning analyses view --direction sideways
error: invalid --direction "sideways", want one of asc|desc
```

Paging parameters are deliberately **not** among these flags. `page` and
`per_page` are answered by `--page` / `--page-cursor` from the connector's
declared pagination spec, so there is exactly one way to page and it cannot be
bypassed by setting the raw parameter.

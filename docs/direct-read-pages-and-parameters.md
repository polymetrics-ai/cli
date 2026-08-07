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
| `pagination_declared_none` | the connector declares pagination type `none`; a declaration is not proof the provider agrees |
| `pagination_not_addressable_for_request` | a strategy is declared, but it cannot page this request (a POST read carries its selection in a body) |
| `pagination_spec_invalid` | the declared strategy's spec is unusable, so paging degraded to a single page for that connector |
| `page_size_not_requested` | the declared strategy stops on a short page, but names no page-size parameter — the provider chose the size, so a short page proves nothing |
| `ambiguous_collection_shape` | the response holds more than one array and the paged one cannot be identified |

`size` is reported only when a page-size parameter actually reached the wire. A
connector whose pagination spec names none gets the provider's own default, and
the result says so rather than reporting a size it never asked for.

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
accept them. `--page 0` names no page and is refused rather than treated as
unset.

### When a command has its own paging flag

A few connectors declared a paging flag by hand before the derivation existed —
notion's `--page-size` and `--start-cursor`, bahmni's `--start-index`, gong's
`--cursor`. Where one exists, **your value wins**: it is sent as written and the
declared default never overwrites it. Because you chose the window yourself, the
result reports no `number`/`next_number` — the engine has no page number it can
honestly name for it.

Combining such a flag with `--page`/`--page-cursor` selects two different pages
in one request, so it is refused before anything is sent:

```
error: direct read received --page and the paging parameter --start-index;
       they select different pages, so pass one of them
```

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

Paging parameters are deliberately **not** derived. They are answered by
`--page` / `--page-cursor` from the connector's declared pagination spec. The
exclusion is by meaning rather than by name: a parameter whose own specification
calls it a cursor is dropped even when the connector's pagination spec never
mentions it, so a derived flag can never become a second way to page. A
parameter that merely shares a name with a paging one is kept — github's
`before` on `/repos/{owner}/{repo}/notifications` is an ISO 8601 timestamp
filter, and it survives.

A path variable the connection already supplies — github's `owner`/`repo`,
interpolated into the endpoint's own path template — is not derived either.
That is the only config-driven exclusion: a config key nothing templates into
the request, such as github's ETL-only `since`, still becomes a flag, because
nothing else would supply it.

# Official Surface Capture: Front CLI Surface Metadata

Captured during #189 from public Front documentation. No credentials were used.

## Source commands

```bash
python3 - <<'PY'
from urllib.request import urlopen, Request
import re, json
from collections import Counter
UA = {'User-Agent': 'polymetrics-agent/1.0'}
llms = urlopen(Request('https://dev.frontapp.com/llms.txt', headers=UA), timeout=30).read().decode('utf-8', 'replace')
# Extract API Reference markdown links, fetch pages with OpenAPI definition blocks, and count method/path rows.
PY
```

## Observed results

- `llms.txt` contained 397 markdown links: 51 guide links and 346 API Reference links.
- 255 API Reference pages exposed per-page `# OpenAPI definition` JSON blocks before rate limiting.
- Parsed OpenAPI operations from those pages: 255 total, 254 unique method/path pairs.
- Parsed method split: `GET=123`, `POST=76`, `PATCH=26`, `PUT=3`, `DELETE=27`.
- One duplicate parsed method/path pair: `PATCH /channels/{channel_id}` from duplicate `Update Channel` pages.
- 91 API Reference links had no per-page OpenAPI block; these were primarily category pages, plugin SDK methods/data models, channel API overview pages, and voice-call category pages.

## Baseline mismatch

Parent issue #188 records an official baseline of 342 operations with method split
`GET=216`, `POST=69`, `PATCH=31`, `DELETE=26`. The per-page Markdown OpenAPI capture above did not
reproduce that baseline. While investigating the ReadMe-rendered HTML, the page exposed registry
metadata for `core-api.json` and `channel-api.json`, but direct unauthenticated fetches of guessed
registry URLs returned 404. Subsequent requests to `dev.frontapp.com` returned HTTP 429 rate-limit
responses.

## Disposition

- #189 adds safe `cli_surface.json` metadata for implemented current streams and representative
  planned intents without overclaiming full API coverage.
- #192 remains responsible for the complete 342-operation ledger and exact method/path
  classification once the full ReadMe OpenAPI registry can be fetched or the official source can be
  re-captured without rate limiting.
- Do not classify unparsed operations as `out_of_scope` merely because they are not implemented.

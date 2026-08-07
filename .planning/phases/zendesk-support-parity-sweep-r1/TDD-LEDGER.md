# TDD ledger — zendesk-support parity sweep

Strict red-first. The test is written against the REAL shipped bundle, run, and its failure recorded
verbatim before any production edit.

## Cycle 1 — the surface is inventoried but not reachable

**Red:** `cmd/connectorgen/zendesk_support_documented_surface_test.go` written and run against the
bundle as shipped on `main`, 2026-08-07.

```
$ go test ./cmd/connectorgen/ -run TestZendeskSupportDocumentedSurfaceIsComplete
--- FAIL: TestZendeskSupportDocumentedSurfaceIsComplete (0.00s)
    zendesk_support_documented_surface_test.go:199: GET /api/v2/{target_type}/{target_id}/relationship_fields/{field_id}/{source_type}: blocked row must carry a 'Named dependency:' marker
    zendesk_support_documented_surface_test.go:199: GET /api/v2/account/email_settings: blocked row must carry a 'Named dependency:' marker
    zendesk_support_documented_surface_test.go:199: GET /api/v2/account/settings: blocked row must carry a 'Named dependency:' marker
    ... 509 rows in total carry no 'Named dependency:' marker ...
    zendesk_support_documented_surface_test.go:248: covered rows = 122, want at least 625 — a blocked row is a gap, not a disposition
    zendesk_support_documented_surface_test.go:268: POST /api/v2/any_channel/validate_token: read-shaped POST is not covered by a command
    zendesk_support_documented_surface_test.go:268: POST /api/v2/autocomplete/tags: read-shaped POST is not covered by a command
    zendesk_support_documented_surface_test.go:268: POST /api/v2/custom_objects/{custom_object_key}/records/search: read-shaped POST is not covered by a command
    zendesk_support_documented_surface_test.go:268: POST /api/v2/it_asset_management/assets/search: read-shaped POST is not covered by a command
    zendesk_support_documented_surface_test.go:268: POST /api/v2/problems/autocomplete: read-shaped POST is not covered by a command
    zendesk_support_documented_surface_test.go:268: POST /api/v2/users/autocomplete: read-shaped POST is not covered by a command
    zendesk_support_documented_surface_test.go:268: POST /api/v2/views/preview: read-shaped POST is not covered by a command
    zendesk_support_documented_surface_test.go:268: POST /api/v2/views/preview/count: read-shaped POST is not covered by a command
FAIL
FAIL	polymetrics.ai/cmd/connectorgen	0.533s
```

`509` is the exact count of `Named dependency:` failures, measured:
`go test ... 2>&1 | grep -c "Named dependency"` → `509`.

**What red proves.** The 631-row inventory is already complete and correctly counted — every
assertion about totals, method split, duplicates, malformed paths, carried rows and dispositions
passes on the shipped bundle. What fails is exclusively **reachability**: 509 rows disposed by one
blanket shared-foundation sentence, 122 covered against a floor of 625, and all eight read-shaped
POSTs uncovered. That is the honest statement of this connector's gap and it is why the count did
not move.

**Green:** pending — slices 2–4.

## Cycle 0 — derivation checked against the test's own rules before adoption

Finding 34's lesson is that a derivation can pass a count and still violate the rules the test
enforces. `tools/derive_zendesk_support.py` therefore asserts, and exits non-zero on, any row
containing `?`, `*` or a space and any repeated (method, path) pair, **before** writing
`DERIVED-OPERATIONS.json`. It ran clean at 625.

It also refused its first run: `refusing to author paging flag 'start_time' on
GET /api/v2/incremental/organizations`. That refusal was correct behaviour on an incorrect
blocklist — see PLAN.md, "Paging". The blocklist was narrowed with the reason recorded, not widened
to make the run pass.

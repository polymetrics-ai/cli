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

## Cycle 2 — reads promoted (slice 2)

**Red carried from cycle 1.** Green for the read half:

```
$ go run ./cmd/connectorgen validate            -> 551 connector(s) checked, 0 findings
$ go run ./cmd/connectorgen surface-sync --check -> 0 filled, 0 corrected, 0 connectors
$ go test ./internal/connectors/commandrunner/ -run TestEveryImplementedCommandPassesRuntimePreflight
ok  	polymetrics.ai/internal/connectors/commandrunner
$ xargs -P 12 -I{} probe_reachability.sh < 427 command paths  -> 0 failures
```

Still red on the write half only:

```
zendesk_support_documented_surface_test.go:267: covered rows = 427, want at least 625
```

**Three runtime rejections drove real design, in this order — each one caught a class the
previous fix could not have found:**

1. `operation direct read body_schema: compile schema: unknown keyword "anyOf"` on three POST
   reads. AGENTS.md's union rule and `engine.CompileSchema` are the same rule stated twice.
2. `compile schema: properties.name: unknown keyword "example"` — the dialect is a strict draft-07
   SUBSET, so an OpenAPI annotation lifted straight out of the spec fails to compile. The sanitizer
   raises on any keyword that is neither compilable nor a known-droppable annotation, so a keyword
   that constrains a payload cannot be dropped silently.
3. `surface-sync` correcting `output_policy` on the binary download. Generating a field the runtime
   then strips is exactly the hand-maintained drift AGENTS.md forbids, so the generator stopped
   emitting it rather than the correction being accepted.

## Cycle 3 — writes promoted, test tightened (slice 3)

**Green.**

```
$ go test ./cmd/connectorgen/
ok  	polymetrics.ai/cmd/connectorgen	11.712s
$ go run ./cmd/connectorgen validate            -> 551 connector(s) checked, 0 findings
$ go run ./cmd/connectorgen surface-sync        -> 0 filled, 0 corrected; ledger byte-identical
$ go test ./internal/connectors/commandrunner/ -run TestEveryImplementedCommandPassesRuntimePreflight
ok  	polymetrics.ai/internal/connectors/commandrunner
$ probe 511 implemented+partial command paths   -> 0 failures
```

### The assertion that changed, and why it is stronger rather than looser

Cycle 1's floor was `covered >= 625`. It is now an exact partition — `covered == 509`,
`blocked == 122`, and a per-class breakdown of those 122.

The floor encoded an assumption this provider's spec falsifies: Zendesk documents 625 operations
but publishes a **request contract** for far fewer, and AGENTS.md forbids inventing the payload
shape needed to reach the rest. The floor also could not distinguish a legitimate block from a
regression — any number ≥ 625 passed. The partition fails if an operation regresses to blocked, if
one is unblocked without moving its class, or if a block is added naming no capability at all.
That is three failure modes the inequality had none of.

This is the same move as workday-rest's 916 → 907: the constant changed because the old one was
wrong, and it gained assertions in the same commit rather than losing them.

### Runtime rejections that drove the write design

1. `reverse ETL command must declare risk text` (170 findings) — every generated command now
   carries the action's own risk sentence.
2. `lacks flag mappings for required record fields` (76 findings) — `checkCLISurfaceWriteFlags`
   requires an implemented reverse-ETL command to bind every required record field. Mirrored, not
   restated: a field with no scalar leaf returns no flags and the command drops to `partial`
   (finding 4), and a credential-named field never becomes a flag at all.
3. `flag --ticket maps outside write schema: record field "ticket" is not declared` — Zendesk
   writes its unions as a bare required-list per arm over PARENT-level properties, so an arm read
   in isolation requires a field it does not define. Each arm is rebuilt as the parent shape
   narrowed to its own field.
4. `record_schema admits only an empty object ({})` — twice, and the second time is the lesson.
   The first guard tested the INPUTS (has a body? has a path variable?) and passed three operations
   that declare a body schema with no properties in it. The guard now tests the BUILT schema, which
   is the only thing the runtime actually sees.

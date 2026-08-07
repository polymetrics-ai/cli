# Workday Rest documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`workday-rest`, landing order 30). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://community.workday.com/sites/default/files/file-hosting/restapi/services2026.30.json`
- **Kind**: other
- **Retrieved**: 2026-08-07, 617538 bytes
- **Documented operations: 920**
- **By method**: n/a
- **Read / write split**: 655 read, 265 write
- **Deprecated (still counted)**: unknown

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | unknown |
| Re-derived | 920 |
| Delta | unknown |

**Undetermined** — a defensible count could not be established from a current artifact.

**How it was counted.** The given index.html (https://community.workday.com/.../restapi/index.html) is a JS SPA shell ('Workday REST Services Directory', React app, no server-rendered content - confirmed by fetching the raw HTML: 1226 bytes, just a <div id=root> and a bundle.js reference). Rather than rendering it in a browser, I downloaded and read its bundle.74683fb3f0cf33d26e31.js (1,734,934 bytes) directly, which revealed the exact runtime call `fetch("services2026.30.json")` used to populate the UI. Fetching that manifest directly (200 OK, 617538 bytes) yielded Workday's own machine-generated directory: a JSON object with 'productionConfidenceLevel' (52 current, actively-documented service modules, each already resolved to its latest version), 'previewConfidenceLevel' (0, empty - no services currently in preview), and 'archivedServices' (38 entries = superseded older versions of 17 of those same 52 service names, e.g. Absence Management v1-v4 all archived now that v5 is current; plus 2 service names - 'Performance Management' and 'Payroll Public' - that exist ONLY in archivedServices with no current replacement at all, i.e. fully retired). Each of the 52 current entries carries its own literal 'operations' array of {path, method} pairs generated from that service's real OpenAPI 2.0 spec. Summing len(operations) raw across all 52 gives 1143, but the per-entry 'method' field is not always a real HTTP verb: cross-referencing the manifest against the bundle's own OPEN_API_CUSTOM_TAGS config (found the exact string 'x-workday-confidence-level' defined there as CONFIDENCE_LEVEL) proved that 2 of the 7 distinct 'method' values found - 'parameters' (177 occurrences) and 'x-workday-confidence-level' (46 occurrences) - are OpenAPI path-item-level metadata keys (Swagger 2.0's per-path shared 'parameters' array, and a Workday vendor extension), not HTTP methods; this was confirmed directly in the data by observing that every 'parameters'-tagged entry shares the exact same path as a sibling real-method entry on the same service (e.g. '/values/leave/status/' appears once as method:get and once as method:parameters), i.e. it decorates an existing operation rather than naming a new one. Excluding both non-method keys leaves exactly 5 real HTTP methods - get=655, post=154, patch=58, delete=33, put=20 - summing to operations_total=920 (reported here as operations_read=655 [get] and operations_write=265 [post+patch+delete+put=154+58+33+20]). No within-service duplicate (path,method) pairs were found. I attempted to cross-validate this manifest against the raw per-service OpenAPI files themselves (e.g. openApiFiles/wql_v1_20260727_oas2.json) but every direct fetch attempt 404'd; response headers showed CloudFront-Key-Pair-Id / CloudFront-Policy / CloudFront-Signature cookie-setting on the 404, indicating those individual spec files sit behind a signed-cookie CloudFront flow that a plain HTTP fetch cannot satisfy - I did not pursue this further since the manifest itself is the exact data source the live docs UI renders from (not a lesser derivative), so it stands as authoritative on its own.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 0** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 0** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm workday-rest <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#3231** (new generation); children are expected at **#3232–#3238** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/workday-rest_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^workday-rest_'`) — gong carried two, and a targeted
   `-run` missed the second.
2. **GREEN** — author the bundle to satisfy it.
3. **REFACTOR** — docs, catalogs, operation endpoint ledger resync.
4. Gates, then no-mistakes.

`check_red_observed.py` refuses to let this connector proceed to implementation until the red
failure is real observed output.

## Safety notes

- Do not loosen `connectorgen validate`, the connector boundary gate, `certify`, or
  `TestEveryImplementedCommandPassesRuntimePreflight` to make this pass.
- Nothing is marked `implemented` unless its command genuinely runs; every block names a dependency.
- Run the WHOLE `cmd/connectorgen` package plus `internal/cli`, never just a targeted `-run`.
- Regenerating docs rewrites ~1,034 files of pre-existing `main` drift (finding F4) — revert every
  non-workday-rest path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.

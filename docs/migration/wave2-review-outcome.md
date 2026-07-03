# Wave2 adversarial review outcome (wf_374c6e45-4d6, 2026-07-03)

Sample: 80 reviewed; 14 were out-of-scope sample-selection errors (native_go/L-bucket = wave3/4
scope). In-scope: 66. FAILS: 18 (27% — below 30% halt line, but SYSTEMATIC template defects).

## Defect classes (population-wide sweep needed, not just the 18)
1. **Missing `projection: passthrough`** where legacy emits raw records verbatim (todoist, toggl,
   wordpress, workable, younium, typeform, printify + everhour id-coercion variant). Detectable
   mechanically: legacy read path does emit(record) without mapRecord field-building.
2. **Bogus/misdeclared incremental blocks** (cursor_field w/o request_param + legacy never filters,
   or missing bare block where legacy publishes CursorFields): pylon, reply-io, wordpress,
   convertkit(-inverse). Rule: bare incremental.cursor_field ONLY iff legacy publishes CursorFields;
   request_param ONLY iff legacy sends a server-side filter.
3. **Fixture page_size leaked into live config** (`page_size: 2`): aviationstack, awin-advertiser.
   Sweep: any pagination.page_size < 10 in streams.json.
4. **Fixture schema-echoes** (not real wire shapes): workable, younium, everhour.
5. Singles: hibob (last_path_segment misuse), nasa (unbounded pagination default vs legacy cap),
   mux (dead spec keys), zenefits (check makes network call legacy never did).

## Repair plan (next steps, in order)
1. Mechanical sweep script over ALL 411 bundles for classes 1-3 signatures → full defect roster.
2. Repair fan-out (Sonnet agents, 1-3 bundles each) for the 18 + sweep hits; conventions.md patch
   adding: passthrough-decision rule (grep legacy for verbatim emit), incremental-declaration
   truth table, fixture-value/live-config separation rule.
3. Re-run conformance + a fresh 10% review sample of REPAIRED classes; then wave2 truly closes.

Full findings: docs/migration/wave2-review-raw.json (result.fails[]).

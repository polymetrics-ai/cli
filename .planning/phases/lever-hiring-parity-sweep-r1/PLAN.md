# Lever Hiring documented-operation parity — plan

> **Generated mechanically** from `data/cli-top50-fixed-schema-sweep-r1/MASTER-PLAN.json`
> (`lever-hiring`, landing order 2). The operation count below was derived in the sweep-wide
> artifact pass, not re-reasoned here. **Per-connector findings are NOT pre-planned** — if this
> connector surprises you during implementation, STOP and record it rather than forcing it into
> this shape.

Part of `cli-top50-fixed-schema-sweep-r1`. One connector, one PR.

## Operation surface, derived before authoring

- **Artifact**: `https://hire.lever.co/developer/documentation`
- **Kind**: html_reference
- **Retrieved**: 2026-08-07, 1203890 bytes
- **Documented operations: 106**
- **By method**: DELETE 11, GET 55, POST 26, PUT 14
- **Read / write split**: 55 read, 51 write
- **Deprecated (still counted)**: 2

### Against the provider-artifact ledger

| | |
| --- | ---: |
| Ledger recorded | 67 |
| Re-derived | 106 |
| Delta | 39 |

**Finding: the ledger is stale.** The live artifact disagrees; see the note below.

**How it was counted.** No machine-readable spec exists for Lever's Hiring API (searched GitHub for segmentio-style official repos; only lever/postings-api exists and that is the separate, unauthenticated Postings/job-board API, a different product from the Hiring API at hire.lever.co). Counted directly from https://hire.lever.co/developer/documentation (HTTP 200, 1203890 bytes decompressed HTML, fetched with curl -sS -L --compressed). This is one single long HTML page: an h1#lever-api-reference marks the start of the reference section; each h2 is a resource group (31 groups: applications, archive-reasons, audit-events, candidates, contacts, disposition-stages, eeo, feedback, feedback-templates, files, form-fields, interviews, notes, offers, opportunities, panels, postings, posting-forms, profile-forms, profile-form-templates, referrals, requisitions, requisition-fields, resumes, sources, stages, surveys, tags, uploads, users, webhooks-via-the-api); each h3 under it is either a real endpoint or a sub-type/attribute description. For each h3, I regex-searched (scoped to that h3's HTML slice, up to the next heading) for `<pre[^>]*><code[^>]*>METHOD /path</code></pre>` blocks (two different pre/code tag-attribute styles are used inconsistently in the page: 42 use `<pre class="hljs"><code class="hljs bash">`, 64 use bare `<pre><code>` -- matched both). 101 of 125 h3 sections yielded 106 unique (METHOD,path) pairs (4 h3 sections document 2-3 distinct endpoints each, e.g. 'Update contact links by opportunity' = addLinks+removeLinks); the other 24 h3 sections are field-type/object-type documentation (Code, Currency, Date, Dropdown, ApprovalStep, Approver, Diversity Surveys category header, etc.) or cross-reference notes with no endpoint of their own (verified each by hand, e.g. 'Create an application' just points readers to 'Apply to a posting' + 'Upload a file'). The h2 'Candidates' section (the pre-2020 predecessor to 'Opportunities') is now only a 1-sentence deprecation redirect with zero remaining endpoint declarations of its own, so it contributes 0 (not double-counted with Opportunities). Verified 0 duplicate (METHOD,path) pairs and 0 param-name-normalized collisions across all 106. Webhook EVENT names (10, excluded from total per policy) were read from the h3#event-payloads list under the pre-reference 'Webhooks' guide section (event: applicationCreated, candidateHired, etc.) -- kept fully distinct from the h2 'Audit Events' API resource (GET /audit_events) and its 'Tracked actions' subsection, whose entries (account:deactivated, user:removed, etc.) are audit-log action-type values, not webhook events, and are correctly NOT counted as webhook events.

## Hazards

- None recorded during derivation beyond the standard bar.

## Webhook inventory (input for `cli-webhook-surface-sweep-r1`)

- **Webhook events: 10** — excluded from the operation total per the counting policy.
- **Webhook management endpoints: 4** — these stay **in scope** and are counted in the
  total. Create/list/update/delete of a webhook *subscription* is an ordinary REST operation; only
  webhook *events* are deferred.

## Required scopes — all five must be covered

Every documented operation must land as an ETL stream, a reverse-ETL write, a direct read, a direct
write, or a binary transfer, **and** be individually reachable as its own `pm lever-hiring <command>`.
Every `api_surface.json` row must carry exactly one of `executable`,
`blocked-with-named-dependency`, or `unsupported-with-source-citation` — **never blank, and never
the legacy `excluded` category**.

## Issues

Parent **#3223** (new generation); children are expected at **#3224–#3230** (new-generation pattern) — **CONFIRM with `gh-axi` before using them in a PR body**, this is inferred from the pattern, not verified.

Use `Closes` only for what this PR genuinely completes and `Refs` for the rest.

## TDD sequence — the red test is NOT generated, by design

1. **RED** — write `cmd/connectorgen/lever-hiring_api_surface_test.go` **against this connector's real
   bundle** and **watch it fail**. Paste the verbatim failure into `TDD-LEDGER.md` and set
   `tdd.red_confirmed` / `tdd.red_failure` in `RUN-STATE.json`. Check first for a pre-existing
   surface test (`ls cmd/connectorgen/ | grep '^lever-hiring_'`) — gong carried two, and a targeted
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
  non-lever-hiring path.
- Inspect the website catalog diff **by object, not by line**.
- No credential or token-derived value is ever emitted.

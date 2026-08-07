---
coverage:
  - id: D1
    description: Help Scout Mailbox's documented operation surface is recorded at 144, deduped on the templated path the provider publishes rather than on example URLs.
    verification:
      - kind: unit
        ref: cmd/connectorgen/help_scout_api_surface_test.go (TestHelpScoutAPISurfaceOperationLedger — row count, method split, uniqueness, no "?" rows)
        status: pass
    human_judgment: false
  - id: D2
    description: The wildcard row standing for 33 report endpoints is gone; a wildcard is a family of operations, not an operation.
    verification:
      - kind: unit
        ref: cmd/connectorgen/help_scout_api_surface_test.go (rejects "*" in any row)
        status: pass
    human_judgment: false
  - id: D3
    description: Exactly one binary download is modelled, and its base64-in-JSON sibling is not.
    verification:
      - kind: unit
        ref: cmd/connectorgen/help_scout_api_surface_test.go (both endpoints pinned explicitly)
        status: pass
      - kind: other
        ref: ./pm help-scout attachments download --help
        status: pass
    human_judgment: false
  - id: D4
    description: Collapsing the sync/async customer delete into one operation loses no documented behaviour.
    verification:
      - kind: other
        ref: ./pm help-scout delete-customer plan --help exposes --customer-id and --async (enum true|false, omit_when_absent)
        status: pass
    human_judgment: false
  - id: D5
    description: All 139 covered operations are reachable as pm commands; before this phase the bundle had 4 streams and no command surface.
    verification:
      - kind: other
        ref: built ./pm invoked over all 139 command paths from cli_surface.json — 0 unreachable
        status: pass
      - kind: integration
        ref: internal/connectors/commandrunner TestEveryImplementedCommandPassesRuntimePreflight
        status: pass
    human_judgment: false
  - id: D6
    description: The change is confined to help-scout — no shared generated artifact drifted.
    verification:
      - kind: other
        ref: connectorgen surface-sync --check (551 scanned, 0 filled / 0 corrected); operation_endpoint_ledger.json unchanged; changed-path audit
        status: pass
    human_judgment: false
  - id: D7
    description: Stream schemas describe the records Help Scout actually returns.
    verification:
      - kind: other
        ref: connectorgen validate — 0 findings after it rejected an assumed x-primary-key ["id"] on user statuses, which key on userId
        status: pass
    human_judgment: false
  - id: D8
    description: Docs, website catalogs and golden transcripts reflect the new help-scout command surface.
    verification:
      - kind: other
        ref: deferred by design — shared generated artifacts regenerate ONCE at the end of the consolidated sweep
        status: unknown
    human_judgment: true
    rationale: >-
      Not done yet, and sweep-level rather than help-scout-level. A human must confirm the
      end-of-sweep regeneration ran before the PR merges; the CLI help/docs/website parity overlay is
      unmet until then. Recorded as unknown deliberately.
---

# SUMMARY — help-scout documented-operation parity

## Delivered

Help Scout goes from **8 rows standing in for 144 operations, 4 of them read-only streams** to a full
declarative surface.

| | Before | After |
| --- | ---: | ---: |
| `api_surface.json` rows | 8 | **144** |
| Covered | 4 | **139** |
| Blocked (named dependency) | 0 | **5** |
| Legacy `excluded` | 4 (one a **wildcard** covering 33 endpoints) | **0** |
| Streams · direct reads · binary · writes | 4 · 0 · 0 · 0 | **24 · 49 · 1 · 65** |
| Reachable commands | **0** | **139** |
| `capabilities.write` | false | **true** |

## Why 144, against a ledger of 146 and a derivation of 145

Help Scout publishes no machine-readable spec, and the ledger's recorded artifact URL **404s**. The
surface is the 146 endpoint pages in the shared left-nav of `developer.helpscout.com/mailbox-api/`,
each fetched and parsed individually — each renders exactly one `METHOD path` line.

**Both deltas are dedup, not missing endpoints:**

- **146 → 145** — `…/threads/{threadId}/original-source` is documented once per `Accept` header.
  Their literal request lines match, so the sweep derivation already caught it.
- **145 → 144** — `DELETE /v2/customers/{customerId}` is documented as *Delete Customer* and *Delete
  Customer Asynchronously*. Their **literal** lines differ (`?async=true`), so deduping on the
  example path misses it — but both pages publish the **same templated path**.

The second is **exactly the defect the captain flagged on lever-hiring: double-counting query-string
variants**. Counting is therefore done on the templated path Help Scout itself declares, and the test
now rejects any row containing `?`.

## The trap inside that collapse

Whichever page survives the dedup also *names* the write action — and the async page won. That would
have shipped a command called `delete-customer-asynchronously` whose path performs the
**synchronous** delete. Renamed to `delete_customer` (named for the endpoint), with `--async` exposed
as an optional `omit_when_absent` query flag. Collapsing the count loses no documented behaviour.

## Judgements made, not defaulted

- **Binary: exactly one.** `…/attachments/{id}/file` streams bytes (`Content-Disposition`,
  `image/png`). Its sibling `…/data` returns `{"data":"<base64>"}` as `application/hal+json` and is an
  ordinary read. One path segment apart; both pinned by the test.
- **Stream vs direct read.** Streams are the `List …` pages plus the two organization sub-collections
  whose responses are `_embedded` arrays despite "Get" wording. Everything else is a direct read
  **including all 33 reports** — a report is an aggregate over a time window, not a record
  collection. `Get Conversation` embeds `_embedded.threads` and would be misclassified by an
  `_embedded`-only rule, so the title is the signal, not the response shape.
- **Read vs write.** All 79 GETs read; all 65 mutations write. No POST is a disguised read — Help
  Scout puts every filter and search on GET query parameters.
- **v3 → blocked.** `base_url` is `…/v2` and `normalizeDirectReadPathForBaseURL` only strips a prefix
  the path starts with, so `/v3/…` builds `…/v2/v3/…`. The named remedy — rebase to host root —
  changes a **shipped `spec.json` default** and needs a config migration, which does not belong in a
  parity commit.

## GSD / TDD

Red test authored, then **run and captured** before any production edit. It was committed in an
explicit `red_confirmed: false` state while a corrupted shared Go build cache prevented the run, and
the observed failure was committed separately once the cache recovered — rather than claiming
evidence that did not exist. Lifecycle and skill routing in
`.planning/traces/gsd-top50-sweep-continue-r1.md`.

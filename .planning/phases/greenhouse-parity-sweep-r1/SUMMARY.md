---
coverage:
  - id: D1
    description: Greenhouse Harvest's documented operation surface is recorded at its true size of 138, re-derived from the provider artifact rather than carried from the ledger.
    verification:
      - kind: unit
        ref: cmd/connectorgen/greenhouse_api_surface_test.go (TestGreenhouseAPISurfaceOperationLedger — row count, method split, uniqueness)
        status: pass
    human_judgment: false
  - id: D2
    description: Every one of the 138 rows carries exactly one disposition, and every blocked row names a checkable dependency.
    verification:
      - kind: unit
        ref: cmd/connectorgen/greenhouse_api_surface_test.go (exactly-one-disposition, zero legacy excluded, "Named dependency:" marker)
        status: pass
    human_judgment: false
  - id: D3
    description: The markup-damaged DELETE /tags/candidate/{tag_id} declaration the shipped bundle had dropped is restored and executable.
    verification:
      - kind: unit
        ref: cmd/connectorgen/greenhouse_api_surface_test.go (explicit pin on the row)
        status: pass
      - kind: other
        ref: ./pm greenhouse destroy-candidate-tag plan --help
        status: pass
    human_judgment: false
  - id: D4
    description: All 127 covered operations are reachable as pm commands; before this phase `pm greenhouse` did not exist.
    verification:
      - kind: other
        ref: built ./pm invoked over all 127 command paths from cli_surface.json — 0 unreachable
        status: pass
      - kind: integration
        ref: internal/connectors/commandrunner TestEveryImplementedCommandPassesRuntimePreflight
        status: pass
    human_judgment: false
  - id: D5
    description: The change is confined to greenhouse — no shared generated artifact drifted.
    verification:
      - kind: other
        ref: connectorgen surface-sync --check (551 scanned, 0 filled / 0 corrected); operation_endpoint_ledger.json unchanged; changed-path audit
        status: pass
    human_judgment: false
  - id: D6
    description: Docs, website catalogs and golden transcripts reflect the new greenhouse command surface.
    verification:
      - kind: other
        ref: deferred by design — shared generated artifacts regenerate ONCE at the end of the consolidated sweep, not per connector
        status: unknown
    human_judgment: true
    rationale: >-
      Regeneration is a sweep-level step, not a greenhouse-level one, and it is not done yet. A human
      must confirm it happened before the sweep PR merges. Recording it as unknown rather than pass
      is the point: the CLI help/docs/website parity overlay is unmet until that regeneration runs.
---

# SUMMARY — greenhouse documented-operation parity

## Delivered

Greenhouse Harvest goes from **129 recorded rows, none reachable** to **138 dispositioned rows and
127 working commands**.

| | Before | After |
| --- | ---: | ---: |
| `api_surface.json` rows | 129 | **138** |
| Covered | 126 | **127** |
| Blocked (named dependency) | 0 | **11** |
| Legacy `excluded` | 3 | **0** |
| `operation_ledger_version` | unset | **1** |
| Reachable commands | **0** (`pm greenhouse` was `unknown command`) | **127** |

## How 138 was established

Re-fetched `developers.greenhouse.io/harvest.html` — **1,636,662 bytes, byte-identical to the sweep
derivation**, which proves it is the same artifact and lets the derivation be *reproduced* rather
than trusted. Greenhouse publishes no OpenAPI for Harvest, so each operation is the canonical
`<h3>HTTP Request</h3>` declaration following every endpoint section: 138 declarations, zero
duplicates, `GET 69 / POST 28 / PUT 8 / PATCH 19 / DELETE 14`, reconciling exactly with the ledger.

## The nine rows the bundle was missing, and why

1. **One markup-damaged declaration.** `DELETE /v1/tags/candidate/{tag id}` carries a stray
   unescaped `&#39;` and a placeholder with a literal space in Greenhouse's own docs. A regex ending
   at `[^<\s]+` truncates it; ending at `[^<]+` recovers it.
2. **Eight Harvest v2 operations.** Three share an `<h2>` with a deprecated v1 sibling, so counting
   headings (135) instead of declarations (138) loses them.

## Judgements made, not defaulted

- **Out-of-base v2 → blocked.** One HTTPBase (`…/v1`); the runtime endpoint ledger needs
  `operation.rest.path` to equal an `api_surface` row verbatim, so correct URL construction and
  ledger membership cannot both hold. Settled on chatwoot; not re-litigated.
- **Deprecated is a disposition, not an exclusion.** The 3 deprecated v1 mutations are blocked with
  `model: deprecated`, each naming its **v2 successor**.
- **Stream vs direct read.** 43 singleton detail GETs stay streams; converting them would delete
  shipped, fixture-backed functionality inside a parity commit. Each instead gained a required named
  flag bound to its config key. Direct-read modelling recorded as a follow-up (PROGRESS finding 21).
- **Binary: none.** Harvest returns JSON everywhere; attachments are referenced by URL.
- **Read vs write.** Conventional — no POST is a disguised read.

## GSD / TDD

Red test written and **run against the real bundle** before any production edit; verbatim failure in
`TDD-LEDGER.md` and `RUN-STATE.json`. Lifecycle and skill routing recorded in
`.planning/traces/gsd-top50-sweep-continue-r1.md`.

---
coverage:
  - id: D1
    description: Production connector declarations are embedded in defs.FS.
    verification:
      - kind: unit
        ref: internal/connectors/engine TestProductionDefinitionsEmbedEveryRateLimitDeclaration
        status: pass
    human_judgment: false
  - id: D2
    description: The first 25 researched connectors have a valid cited policy or explicit unknown declaration.
    verification:
      - kind: integration
        ref: jq plus internal/connectors/engine tests and connectorgen validate
        status: pass
    human_judgment: false
---

# Provider-cited rate-limit declarations R1 summary

The first production rate-limit batch adds 25 declarations and the optional `defs.FS` embed
pattern that ships them. Harvest and CallRail carry cited, scoped policies. The other 23
providers carry explicit `unknown` states because their published policy varies by unrepresented
plan, auth, token, tenant, endpoint, or subject scope. All 25 were later rechecked against the
authoritative provider-artifact sweep ledger and are `done` records with live enumerable provider
artifacts; no declaration is based on an unresearched, skipped, or retired surface.

The gitignored resumability ledger contains every connector verdict, provider URL, retrieval date,
and one-line rationale. `PROGRESS.md` resumes at `sendgrid` for the next batch.

No `streams.json` throttle, credential, CLI surface, or documentation page changed.

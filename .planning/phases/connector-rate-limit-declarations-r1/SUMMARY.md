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
    description: The retained 24 first-batch connectors have a valid cited policy or explicit unknown declaration.
    verification:
      - kind: integration
        ref: jq plus internal/connectors/engine tests and connectorgen validate
        status: pass
    human_judgment: false
---

# Provider-cited rate-limit declarations R1 summary

The first production rate-limit batch originally added 25 declarations and the optional `defs.FS`
embed pattern that ships them. Population recheck removed `vercel`, because it is absent from the
authoritative sweep rather than a `done` record; the retained 24 remain valid. Harvest and CallRail
carry cited, scoped policies. The other 22 providers carry explicit `unknown` states because their
published policy varies by unrepresented plan, auth, token, tenant, endpoint, or subject scope.

The gitignored resumability ledger contains every connector verdict, provider URL, retrieval date,
and one-line rationale. `PROGRESS.md` contains the exact next connector for the next batch.

The second batch adds 25 more declarations: Aha! carries two cited account-scoped fixed-window
budgets (20 requests/second and 300 requests/minute), while 24 providers carry evidence-backed
`unknown` results. The current population-gated total is 49 declarations: 3 `declared` and 46
`unknown`. Every current declaration joins to a `done` sweep record; the separate population audit
lists seven deprecation candidates, none of which has a declaration.

No `streams.json` throttle, credential, CLI surface, or documentation page changed.

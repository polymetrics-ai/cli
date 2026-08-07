# Captain's decision — Marketo scope includes Adobe

**Date:** 2026-08-07
**Decision key:** `marketo-adobe-v2-asset-scope`

## The captain's ruling, verbatim

> "Marketo will include Adobe also if it was acquired and it's in the new API."

## What this settles

**Marketo's documented surface is the FULL surface, including Adobe's v2 Asset API.** The connector
covers what Adobe currently publishes as Marketo's API, not a historical core-REST-only subset.

The conditional in the ruling is the operative test and must be applied literally:

- **"if it was acquired"** — the asset surface must genuinely belong to the Marketo product under
  Adobe, not merely sit adjacent to it in Adobe's wider developer documentation. Adobe publishes many
  APIs; only the ones that are Marketo's are in scope.
- **"and it's in the new API"** — it must appear in the **current** published specification. A
  deprecated or retired endpoint is out of scope regardless of history.

**Both conditions must hold.** An endpoint that fails either is excluded and must be recorded with
the reason, not silently dropped.

## Consequence for the count

The three derivations were **320** (HTML reference), **327** (earlier lane's specification) and
**367** (fresh recount). The ledger's own **322** matches none of them.

**This ruling points at the larger surface**, so ~367 is the expected direction — **but the number is
still to be established, not assumed.** The captain resolved the *scope* question; he did not pick a
number, and neither should we.

**Re-derive under the stated rule** and let the count fall out of it. Then explain the delta against
each of the three prior derivations: for each, say whether it was missing asset operations, counting
retired endpoints, or double-counting query-string variants — the last being the exact error found in
lever-hiring.

## What this does not authorise

- **It does not authorise picking 367 because it is the biggest.** Derive it.
- **It does not authorise including every Adobe API.** Apply both conditions.
- **It does not change the sweep's rules.** Webhook events are still not operations. Every blocked row
  still carries a machine-checkable `Named dependency:` marker. The red test still comes first and is
  still run and captured before authoring.

## Routing

Unblocks `cli-marketo-parity-wave05-r1`, to be authored on the consolidated sweep branch
`fm/cli-top50-sweep-consolidated` in the normal work order.

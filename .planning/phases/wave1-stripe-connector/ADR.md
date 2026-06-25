# ADR — Declarative-HTTP connector template (Stripe) on connsdk

## Status
Accepted (Wave 1 batch 1).

## Context
~500 SaaS connectors are declarative HTTP APIs. They must be real Go packages (per the program's
"hand-write all in Go on a shared SDK" decision) but must not reinvent auth/pagination/retry. GitHub
(Wave 0) kept bespoke HTTP code; we need a connsdk-based exemplar.

## Decision
1. Stripe is the **canonical declarative-HTTP template**: a thin connector package that composes
   `connsdk` (Requester + Bearer + RecordsAt + cursor state) with Stripe-specific stream defs,
   endpoints, and write actions. Other HTTP connectors copy this shape.
2. Same per-system conventions as github: `package <name>`, `New() connectors.Connector`, `init()`
   self-registration, blank import in `registryset`, fixture mode for credential-free conformance.
3. Stripe's id-cursor pagination (`has_more` + `starting_after=<last id>`) is implemented in-package
   for now; if the pattern recurs (it will), extract a reusable `connsdk.IdCursorPaginator`.

## Consequences
- (+) Establishes the repeatable HTTP template; future connectors are mostly stream/endpoint data.
- (+) connsdk gets exercised by a real connector, surfacing any gaps to fix once centrally.
- (−) Per-connector stream mapping is still hand-written (intended — quality + real schemas).

## Alternatives rejected
- Manifest-interpreter engine (the program chose hand-written Go connectors).
- Bespoke HTTP per connector like github (defeats the connsdk investment).

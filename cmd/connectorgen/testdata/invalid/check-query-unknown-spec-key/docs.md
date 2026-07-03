# Overview

Check Query Unknown Spec Key is a deliberately invalid `connectorgen validate` corpus case: its
`streams.json` `base.check.query`'s `limit` entry templates an undeclared spec key
(`config.nope_limit`), which `engine.ResolveCheck` must reject via `checkInterpolations`'s
`base.check.query` validation (checkquery-ledger.md).

## Auth setup

None; this connector has no auth block.

## Streams notes

`widgets` is a single stream, no pagination.

## Write actions & risks

None; this connector has no write actions.

## Known limits

None; this is test fixture data.

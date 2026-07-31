Part of #277 (S7). Write scope: `fixtures/**`, `docs.md`. Deps: #279, #280, #281, #282, #283.

- `fixtures/**` for streams + writes; `docs.md`; `pm connectors certify twenty`; parity-deviation ledger entries for any documented gap (target: none).

Acceptance: `make connectorgen-validate`, `make verify`, focused tests, `pm connectors certify twenty` all green.
Refs #277
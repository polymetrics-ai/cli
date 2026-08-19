# Discussion log — connector batch pipeline r1

## Fixed decisions from the task

1. A batch is one branch and one PR; merge serialization applies to the batch,
   not each connector.
2. Provider-operation counts are ledger evidence. The pipeline validates their
   presence but never estimates or rewrites them.
3. Every operation is accounted for as executable, provider-blocked with a
   reason, or justified-excluded. Omission is never an outcome.
4. `availability: implemented` is legal only when the real runtime preflight
   accepts its command and derived `api_surface` matches its operation.
5. Runtime output is preserved. The pipeline must not add a redacting output
   policy or masking path.
6. Public artifact fetches are allowed only for the authoring stage and must
   record the retrieval date. No provider API call or credential is allowed.
7. A failed connector is removed from the candidate manifest/branch and kept in
   the batch report with its actual reason. It must not be silently skipped.

## Resolved local questions

- The ledger is not in the checkout. Its authoritative source is the absolute
  firstmate-workspace path recorded in `CONTEXT.md`; ignored-path behavior was
  the cause of the earlier false negative.
- The live ledger has grown beyond the original survey snapshot. The committed
  manifest records the snapshot metadata and each selected row, so a later
  ledger growth cannot rewrite history.
- A new `connectorgen batch` command is genuinely needed. Existing commands
  validate/generate a bundle but cannot consume a ledger, retain candidate/drop
  decisions, or aggregate a per-connector gate.

## Originally deferred by an external contract

- The generator that emits v2 `api_surface.json` provenance waited for #3869
  to merge so it could consume the provider-artifact table and endpoint
  citations.
- Default executable output policy waited for #3870; the then-current
  `surface-sync` default was redacting and could not be used for a new bundle.

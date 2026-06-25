# RUNBOOK / Rollback — Stripe connector

## Nature
Additive: a new pure-Go connector package + one registryset import + one catalog entry flip. No
dependency, no schema, no build-mode change.

## Verify
- `make verify` green. `pm connectors inspect stripe --json` → kind Connector, read+write.
- Fixture mode: `pm credentials add stripe-local --connector stripe --config mode=fixture` →
  catalog/read work without live creds.

## Rollback
- Remove `internal/connectors/stripe/`, drop its blank import from registryset/registry_gen.go,
  revert the `source-stripe` catalog entry to planned_native_port. No other connector affected.

## Operational notes
- Live use needs a Stripe secret key in the vault (`client_secret`); reverse-ETL writes are
  approval-gated. No background jobs added.

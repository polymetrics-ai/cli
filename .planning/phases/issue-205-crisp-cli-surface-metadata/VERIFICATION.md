# Verification — Issue #205 Crisp CLI surface metadata

## Planned commands

```bash
go run ./cmd/connectorgen validate internal/connectors/defs/crisp
```

Run once before adding the scaffold for red evidence, then again after implementation for green evidence.

```bash
go run ./cmd/connectorgen validate internal/connectors/defs
```

Run after targeted validation to ensure the new bundle does not break fleet validation.

## CLI help/docs/website parity

#205 adds docs/help metadata only, not runtime command dispatch. Runtime help/manual/website generation is marked pending for #206.

Checks deferred to #206 unless #205 unexpectedly changes runtime help behavior:

```bash
pm help connectors
pm connectors
pm connectors inspect crisp --json
rg -n "crisp|Crisp" docs/cli website
```

## Current result

Pending. No production files changed yet.

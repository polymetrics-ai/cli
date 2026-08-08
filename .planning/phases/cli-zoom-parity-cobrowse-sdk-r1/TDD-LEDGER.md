# TDD Ledger — Zoom Cobrowse SDK documented-operation parity, R1

## RED — pending before production declarations

The red checkpoint will change only `internal/connectors/defs/zoom/command_surface_test.go` and
synthetic Cobrowse fixtures. It will assert the documented four-operation target before any Zoom
production declaration changes:

- covered operations: `18 → 22`
- locally blocked: `1824 → 1820`
- direct reads: `13 → 17`
- writes: unchanged at `2`
- two exact monthly report GET routes with only explicit `from`/`to` input; two exact session-ID
  routes; response redaction; and proof that response-only pagination fields are never sent.

Run the test against the current pre-Cobrowse bundle, capture the literal failure here, commit and
push the red-only state before editing `operations.json`, `cli_surface.json`, `api_surface.json`,
metadata, generated ledgers, or docs.

## GREEN — pending

Record focused test, conformance, surface/validator, binary, docs/website, and inline review
evidence after the declarations exist and the red test becomes green.

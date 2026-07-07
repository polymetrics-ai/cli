# Phase Summary

Phase: github-graphql-engine

Status: completed_with_warnings.

This phase adds safe fixed GraphQL request support to the declarative connector engine for GitHub CLI
parity and future GraphQL connectors. It intentionally does not expose a raw GraphQL command.

Implemented:

- `StreamSpec.GraphQL` for POST query streams with fixed bundle documents and declared variables.
- `WriteAction.GraphQL` with `body_type: graphql` for fixed mutation bodies.
- GraphQL top-level `errors[]` fail-closed handling for reads and writes.
- Bundle/schema validation that rejects templated GraphQL documents, wrong query/mutation class,
  missing GraphQL blocks, invalid operation names, and invalid variable names.
- Explicit typed GraphQL variables for JSON integers, numbers, and booleans without guessing from
  string templates.
- Tests for fixed payload construction, record-query override prevention, error handling, and load
  validation.

# Discussion log — issue 4371

## Fixed decisions from issue #4371

1. Keep every provider-documented operation visible. An unavailable contract
   remains a source-cited named foundation gap; it is never erased to make
   validation pass.
2. The cited-only source-reference projection is closed. Extra mutation gaps
   and disposition payloads cannot be silently layered onto it.
3. Reject incompatible cited-only mutation dispositions before output. This is
   narrower and safer than weakening the descriptor validator or fabricating a
   runtime representation.
4. Preserve normal disposition behavior for byte-backed/contract-complete
   OpenAPI and Swagger operations byte-for-byte.
5. This work does not make provider mutations executable, does not use
   credentials or provider I/O, and does not run reverse ETL.

## Required skills and reading completed

- `connector-lane-build-order`; `firstmate-exhaustive-review`.
- `golang-how-to`, `golang-design-patterns`, `golang-structs-interfaces`,
  `golang-error-handling`, `golang-security`, `golang-safety`,
  `golang-testing`, and `golang-cli`.
- Connector canon, implementation procedure, migration conventions, source
  projection/importer predecessor evidence, issue #4371, and GSD adapter.

## Delivery header location

The completed task-delivery header and acceptance-evidence table are in
`PLAN.md` below. The direct PR must target `main`; after opening, its base is
read back from GitHub's pull-request API.

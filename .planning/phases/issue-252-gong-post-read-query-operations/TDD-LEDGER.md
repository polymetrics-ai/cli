# TDD Ledger — issue #252 Gong typed POST read-query operation execution

## Red (planned before production code)

- Engine operation direct-read sends a POST JSON body from fixed/default body plus typed overrides.
- Engine rejects POST read-query without schema/content type/max bytes/output policy.
- Engine validates body schema before network send.
- Commandrunner allows only typed `path.*`, `query.*`, and `body.*` mappings; raw body mappings fail.
- Validator rejects implemented operation commands with unsafe operation shapes.

## Green

- Added `connectors.OperationDirectReadRequest` / `OperationDirectReader` and engine `OperationDirectRead` for schema-gated `rest_read` GET/POST execution.
- POST operation reads require connector-relative paths, `application/json`, positive `max_bytes`, declared `body_schema`, and supported output policy.
- Commandrunner maps only connector-authored `path.*`, `query.*`, and `body.*` flags; unknown/raw `--body` remains blocked.
- Gong now implements typed POST reads for `meetings integration-status`, `flows steps`, and `flows prospects`; broader arbitrary-filter POST reads remain planned until safe typed filters are authored.
- Validator/conformance allow opt-in GET/POST direct-read coverage and reject unsafe implemented operation shapes.

Evidence:
- `go test ./internal/connectors/engine -run 'OperationDirectRead|WriteJSONArray|WriteMultipart|DirectRead|Write' -count=1`
- `go test ./internal/connectors/commandrunner -run 'OperationDirectRead|DirectRead|RedactRecord' -count=1`
- `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`

## Refactor

- Kept operation direct-read opt-in through explicit output policies and operation metadata, so existing operation ledger rows without output policies stay blocked.
- Updated remaining planned Gong POST read-query blocker text from “executor missing” to “safe typed filter/body flags not yet authored.”

## Skills

gsd-core, golang-how-to, golang-cli, golang-testing, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-context, golang-documentation.

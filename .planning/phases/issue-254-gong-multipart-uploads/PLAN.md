# Plan: Gong bounded typed multipart uploads

Parent: #133
Issue: #254
Branch: `feat/133-gong-cli-parity`
Parent PR: https://github.com/polymetrics-ai/cli/pull/232 (draft)

## GSD command path

- `scripts/gsd doctor`: passed.
- `scripts/gsd prompt plan-phase 254 --skip-research --tdd`: generated `/tmp/gsd-plan-phase-254.prompt.md`.
- `scripts/gsd prompt programming-loop init --phase issue-133-gong-engine-shapes --dry-run`: unavailable (`unknown GSD command: programming-loop`); manual-GSD fallback active.

## Required skills loaded

gsd-core, golang-how-to, golang-cli, golang-testing, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-context, golang-documentation.

## Objective

Add bounded, typed multipart upload support for connector reverse-ETL actions while preserving safety and approval gates.

## Scope

- Add stdlib-only `connsdk` multipart request primitive.
- Rebuild multipart bodies safely per retry.
- Validate file paths and sizes before network send.
- Add typed multipart write body mode with explicit parts.
- Redact file/path/content-like fields in previews, plan samples, errors, and docs.
- Keep Gong multipart operations blocked until typed parts and bounds are safe.

## Non-goals

- No generic `--upload`, arbitrary multipart fields, arbitrary URL/method, or raw HTTP write.
- No credentialed Gong upload tests.
- No new dependencies.

## Implementation slices

1. Red connsdk multipart tests for request shape and path/size safety.
2. Red engine multipart write tests for declared parts only and preflight failures.
3. Extend write action schema/loader/model for multipart parts.
4. Implement streaming/reopenable multipart requester.
5. Add validator and preview redaction rules.

## Verification

- `go test ./internal/connectors/connsdk -run Multipart -count=1`
- `go test ./internal/connectors/engine -run 'WriteMultipart|Write' -count=1`
- `go test ./internal/connectors/commandrunner -run Multipart -count=1`
- `go test ./cmd/connectorgen -run 'Operation|Gong' -count=1`
- `go run ./cmd/connectorgen validate internal/connectors/defs`
- `go test ./internal/connectors/conformance -run 'TestConformance/gong|Static' -count=1`

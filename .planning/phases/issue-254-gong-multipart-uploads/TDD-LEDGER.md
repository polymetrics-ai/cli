# TDD Ledger — issue #254 Gong bounded typed multipart uploads

## Red (planned before production code)

- connsdk multipart test proves boundary/content type, file part, auth/default headers, and response handling.
- connsdk safety tests fail for too-large, missing, traversal, and unsafe paths before network send.
- Engine multipart write sends only declared parts and enforces max bytes.
- Preview/redaction test for multipart fields.
- Validator rejects unsafe implemented multipart commands.

## Green

Pending implementation.

## Refactor

Pending implementation.

## Skills

gsd-core, golang-how-to, golang-cli, golang-testing, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-context, golang-documentation.

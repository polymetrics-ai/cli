# TDD Ledger — issue #252 Gong typed POST read-query operation execution

## Red (planned before production code)

- Engine operation direct-read sends a POST JSON body from fixed/default body plus typed overrides.
- Engine rejects POST read-query without schema/content type/max bytes/output policy.
- Engine validates body schema before network send.
- Commandrunner allows only typed `path.*`, `query.*`, and `body.*` mappings; raw body mappings fail.
- Validator rejects implemented operation commands with unsafe operation shapes.

## Green

Pending implementation.

## Refactor

Pending implementation.

## Skills

gsd-core, golang-how-to, golang-cli, golang-testing, golang-design-patterns, golang-structs-interfaces, golang-error-handling, golang-security, golang-safety, golang-context, golang-documentation.

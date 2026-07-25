# Phase 505 — WhatsApp connector (connectors-as-data)

Parent issue: #505 (sub-issues #506–#515, #527). Branch: `fm/cli-whatsapp-connector-r1`.

## Objective

Deliver a full-parity, config-driven WhatsApp connector under
`internal/connectors/defs/whatsapp/`, at the same completeness bar as `github`/`gong`,
covering the union of two access modes:

1. **Cloud API + Business Management API** (Meta Graph `v25.0`) — the executable business surface.
2. **WhatsApp Web multidevice (whatsmeow)** — modeled as documented, config-scoped ops from
   `vicentereig/whatsapp-cli` (not executed by the declarative HTTP engine; requires a whatsmeow
   session, which is a human gate — see #527).

Framed for HMS/healthcare patient messaging alongside the Bahmni EMR connector: recipient numbers
and message bodies are patient PHI (redacted by default; sends need consent + template pre-approval).

## Architecture decision — pure `defs/` bundle, no native adapter

The connector engine is HTTP-based. The Cloud API mode is fully modeled declaratively. The
whatsmeow (WhatsApp Web) mode is a websocket protocol and cannot be executed by the HTTP engine;
adding `go.mau.fi/whatsmeow` is disallowed (no new dependencies without approval). Following the
`github` precedent (which models local `git` operations as `operations.json` `local_git` entries +
`local_workflow`/`unsupported_local` cli commands + `operation`-blocked api_surface rows), the
whatsmeow mode is modeled as data: `operation` ledger rows (`model: local_workflow`, blocked),
cli commands (`availability: unsupported_local`), and documented `operations.json` entries. This
keeps the deliverable a single coherent `defs/` bundle with zero new Go and zero new dependencies.

## Surface (Cloud API, executable)

- Streams (ETL): `phone_numbers`, `message_templates`, `subscribed_apps`, `waba`.
- Direct reads (GET): phone-number / template / business-profile / media-URL detail.
- Typed POST read-queries (#513): messaging / conversation / pricing / template analytics.
- Reverse-ETL writes (#512): sends (all 11 message types), mark-read, typing, template CRUD,
  business-profile update, phone-number admin (register/deregister/request-code/verify/two-step),
  app subscribe/unsubscribe, QR create/delete, bounded multipart media upload (#515), media delete.
- JSON-array bodies (#514): contacts array, interactive rows, template components arrays.
- Binary (#511): media upload (multipart) + media download (bounded).

## Required skills

Per `.agents/agentic-delivery/references/required-skills-routing.md`: `golang-how-to`,
`golang-cli`, `golang-security`, `golang-safety`, plus the connector authoring recipe in
`docs/migration/conventions.md` and the parity gate in
`.agents/agentic-delivery/references/cli-help-docs-website-parity.md`.

## GSD runtime note

Repo-local GSD adapter (`scripts/gsd`) requires the Pi runtime; this crewmate runs under Claude
Code, so the manual GSD/TDD loop was used and is recorded here as the sanctioned fallback
(research → author → red gate → green → parity → verify).

# Plan — Bahmni-docker connector (issue #516)

## Objective

Deliver `internal/connectors/defs/bahmni-docker/` as a config-driven connectors-as-data bundle with
complete API parity at the same bar as the existing `github`/`gong` bundles, covering parent roadmap
issue #516 and sub-issues #517-#526 in one coherent connector definition.

The connector targets a local Bahmni deployment (`https://github.com/Bahmni/bahmni-docker`) through
the OpenMRS REST v1, Bahmni-core REST, and OpenMRS FHIR2 R4 surfaces it exposes. It must be usable
locally via config (`base_url` + basic auth) and must never read or print secret values at
inspection time.

## GSD runtime

The repo-local GSD adapter (`scripts/gsd`) was not available in this disposable worktree, so this
phase ran the **manual-GSD fallback**: plan before production edits, red/green TDD ledger, and a
verification checklist, all recorded here per `AGENTS.md`.

## Required skills loaded

- `golang-how-to` (orchestrator), `golang-testing`, `golang-security`, `golang-safety`
- `.agents/agentic-delivery/references/required-skills-routing.md`
- `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`
- `docs/migration/conventions.md` (connector authoring recipe)

## Approach

Tier-1 declarative bundle, zero Go. Verified precedent first: all 547 pre-existing connectors are
pure `defs/` with no `native/` adapter, so adding Go here would break the model. The oracle for
correctness is `cmd/connectorgen/validate.go` plus the fixture-replay conformance harness in
`internal/connectors/conformance/`.

## Scope

- 12 ETL streams: patients, encounters, observations, visits, concepts, locations, providers,
  drug_orders, lab_orders, lab_results, appointments, diagnoses.
- 12 bounded direct reads: 6 OpenMRS by-UUID, 4 FHIR R4 by-id, Bahmni-core patient profile detail,
  and one schema-gated typed POST patient search (#524).
- 9 typed reverse-ETL write actions, including a top-level JSON array observation upload (#525) and
  a bounded multipart patient-document upload (#526).
- `api_surface.json` in `operation_ledger_version: 1` mode: every endpoint is a stream, direct read,
  write, or an explicitly blocked operation row with recorded evidence (#520).

## Deliberate decisions

1. **Full-refresh streams, no `incremental` block.** OpenMRS/Bahmni REST exposes no universal
   modified-since cursor across these list endpoints. Recorded under `docs.md` "Known limits"; this
   is why conformance `cursor_advances` legitimately skips rather than being wired to a fake cursor.
2. **Reverse-ETL CLI commands use `availability: partial`**, mirroring gong's existing convention.
3. **Binary download is a blocked `binary_read` operation row**, not an executable command: the
   engine's direct-read executor only supports `rest_read` GET/POST. Binary *upload* is executable
   as a typed bounded multipart write. This is the "bounded binary metadata" #522 asks for.
4. **No raw escape hatch anywhere** — no generic HTTP write, raw JSON body, arbitrary OpenMRS
   method/path/body, shell write, or SQL write.
5. **No write fixtures**, matching gong exactly; `write_request_shape`/`delete_semantics` skip.

## Out of scope

- Live/credentialed checks against any real Bahmni deployment (human-gated).
- Engine-level PHI redaction policy (see VERIFICATION.md outstanding findings).

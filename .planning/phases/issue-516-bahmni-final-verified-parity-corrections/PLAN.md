# Plan — Issue #516 Bahmni final verified parity corrections

## Objective

Continue PR #533 from rejected head `4f5822533d880fbe294a4d33969c11b12ed2675e` until the Bahmni connector reaches truthful, version-pinned, independently reproduced parity with supported Bahmni/OpenMRS contracts. Do not merge and do not claim full parity until a fresh exact-head audit passes.

## GSD runtime

- `scripts/gsd doctor` passed on 2026-07-26.
- Generated the official adapter prompt with `scripts/gsd prompt plan-phase issue-516-bahmni-final-verified-parity-corrections --skip-research`.
- The repo does not expose the previously documented shell `programming-loop` command, so this phase follows the GSD plan/TDD/verification artifact requirements directly under the healthy `plan-phase` adapter path.

## Inputs

- Final audit report: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-bahmni-final-audit-r1/report.md`.
- Captain decision: `/Users/karthiksivadas/karthik-agent-workspace/data/cli-bahmni-final-audit-r1/decision-fix-to-verified-full-parity.md`.
- PR #533 current branch/head synced locally: `feat/bahmni-docker-connector` at `4f5822533d880fbe294a4d33969c11b12ed2675e`.
- no-mistakes run `01KYDK7BF4TMSYKJHA8GA8BF6P` remains as a post-check CI monitor; `no-mistakes axi sync --check` reported local/remote/pipeline heads equal and safety `already_synchronized`.

## Required corrections

1. Appointments: remove false `patient_uuid` server-side scoping; use the version-matched `/appointments` `forDate` contract exactly or block unsupported patient-scoped behavior. Add negative/counterfactual tests.
2. Reads: correct every audited failed read contract:
   - patients list requires `patient_query` (`q`) unless a proven enumeration route is added;
   - lab results require the official `concept` query list;
   - diagnoses use the official `/bahmnicore/diagnosis/search` route;
   - remove/block unsupported `bahmnicore patient-detail` GET;
   - correct patient search to version-matched `GET /bahmni/search/patient` from bahmni-commons.
3. Writes: remove or block unsupported diagnosis POST and bulk-observation POST immediately; retain only version-matched supported writes with fixtures/tests/proof. If disposable-write proof is not available for every action, make write capability truthful by removing/unadvertising unproven actions.
4. PHI policy: enforce connector-sensitive field redaction in runtime tests, not only docs. Normal help/inspection/logs/errors/previews/direct output must not expose raw patient identifiers or clinical fields by default.
5. Diagnosis identity: remove nullable `existingObs` `x-primary-key` unless a stable non-null replacement is proven.
6. Coverage: add Bahmni-specific regression tests for valid/invalid requests, auth, root arrays, two-page pagination, required filters, unsupported routes, PHI output handling, and retained writes.
7. Validation: run Codex-only/local validation. Do not use unrelated 446-doc churn; do not use live credentials unless explicitly provided; do not execute writes against real clinical data.
8. PR metadata: retitle/update body away from stale `Bahmni-docker` and premature “full parity” until final independent audit passes.

## Boundaries

- No merge.
- No new dependencies without approval.
- No credentialed connector checks or live writes unless explicitly requested/provided.
- No generic raw HTTP/write escape hatch.
- Reverse ETL remains plan → preview → approval → execute.
- Keep the 446 unrelated generated connector docs out of this PR.

## Local preview checkpoint addendum - 2026-07-26

- Captain requested a local-only preview binary named `pm-bahmni` after the next coherent buildable correction checkpoint.
- Build path: authoritative Go build from this branch into worktree-root `./pm-bahmni`; expose `/Users/karthiksivadas/.local/bin/pm-bahmni` as a symlink without touching normal `pm`.
- Record source commit, dirty-tree warning/tree hash when applicable, byte size, human size, SHA-256, command path, and credential-free/safe synthetic-lab smoke results.
- Preview is not a merge/parity certification; no write command should be run except disposable synthetic-lab actions explicitly verified at that checkpoint.

## GSD adapter fallback - 2026-07-26

Attempted `scripts/gsd prompt programming-loop ...`; adapter returned `unknown GSD command: programming-loop`. Proceeding with manual GSD programming loop for this correction slice: write focused failing tests, implement smallest production changes, run focused validation, then build local preview.

## Typed write completion directive - 2026-07-26

Captain clarified that scalar-only live-write checks are a useful first pass but not completion. Every retained Bahmni write must be safely executable through the typed CLI surface and receive live synthetic proof. Structured string-valued inputs such as `person`, `identifiers`, `notes`, and other nested objects/arrays must be converted to explicit named, schema-bound flags/builders. If a write cannot be represented or proven against the pinned STANDARD alpha stack, remove it or mark it unavailable with source/live evidence rather than advertising an unusable mutation.

## PHI production-readiness directive - 2026-07-26

Captain clarified that truthful metadata cannot normalize an unprotected-PHI disclaimer into the production candidate. Readiness requires enforced default protection: every normal Bahmni read/help/error/preview path must emit opaque references or apply explicit command-level `redact_fields` for clinical identifiers/values. Raw clinical content may flow only through a trusted typed execution path with deliberate authorization. Add tests covering patient, encounter, observation, diagnosis, lab, appointment, note, and error outputs. If this cannot be completed in PR #533, keep the connector alpha and explicitly report PHI protection as a production blocker; do not call the connector ready.

## Live synthetic write verification authorization - 2026-07-26

Captain authorized parallel live verification of every retained Bahmni write against the existing loopback-only Podman lab after typed/schema tests pass and `pm-bahmni` is rebuilt. Use unique `SYN-CONN-*` disposable identifiers. Every mutation must use the typed CLI plan -> preview -> explicit approval -> execute path; do not use raw JSON/method/path write escape hatches. Do not print or persist credentials/PHI, touch Karthik/Rohit records, reseed/restart containers, or collide across lanes. Capture safe opaque evidence/status only. Stop any failing operation, then either fix it or mark it unavailable with source/live evidence and rerun before claiming readiness.

## Review-fix slice - 2026-07-26

Manual-GSD continuation for prior review findings at head `4e89af9ea5436088f5cef8e9f14e6eee0696b290`. Findings verified as legitimate before production edits:

- Generated catalog and website data still advertise removed Bahmni write actions from stale outputs.
- Source still contains the old mixed-stack support claim despite the frozen STANDARD-only alpha scope.
- Write preview request-line warnings can leak clinical UUID path fields before command-level redaction.
- Appointment command help still claims unsupported `patient_uuid` scoping.
- Patient-search approval/help text still describes the old POST/body route rather than GET query params.

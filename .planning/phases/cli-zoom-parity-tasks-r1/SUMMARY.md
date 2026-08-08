# Summary — Zoom Tasks documented-operation parity, R1

## Delivered

Zoom's live Tasks artifact was re-fetched from
`https://developers.zoom.us/docs/api/tasks.md` on `2026-08-08T15:22:01Z` (HTTP `200`, `68,605`
bytes, SHA-256 `53920b1c473e4d8ccdad03475d6d14f13b6b0b54ce036127dd644e51850f09be`). Its 17
documented methods are all implemented as exact typed operations and provider-native CLI paths:

- six direct reads: assignees, collaborators, comments, import status, task list, and task detail;
- eleven approval-gated direct writes: assignee/collaborator/comment changes, file upload, import
  submission, task create/delete/update; and
- four DELETEs plus PATCH with documented `204 No Content` status-only semantics.

No source pagination parameters were invented; no Tasks endpoint is classified
`unsafe_or_disallowed`; no documented operation was excluded as a duplicate.

## Foundations delivered separately

- `235d6a322` — narrow declared multipart `307`/`308` redirect support for a fixed provider URL,
  bounded Zoom-owned host suffix, snapshot rebuild, and declared bearer reapplication. This can
  unblock similarly declared provider multipart uploads without becoming a generic redirect tool.
- `122b8d8d1` — narrow declared JSON-file validation and file-extension constraints for multipart
  parts. This can unblock providers with a documented JSON-only file requirement without allowing
  caller-selected validators or file policies.

## Evidence and handoff

RED was committed before production declarations. GREEN fixture tests cover all 17 commands, and a
freshly built binary accepted the base, namespace, group, and each exact Tasks help route. Scoped
tests, vet, lint, docs/smoke/contract/generator/boundary/release checks passed; manual inline
`verify-work` and `code-review` found no actionable issue. The runtime cannot register this
provider category and the parent contract forbids role spawning, so the phase records the approved
manual-GSD fallback.

Issue #3939 is the completed provider-owned Tasks slice under parent #3915. A future worker should
select the next provider-owned category from the parent issue tree, re-fetch that category's live
artifact, and start its own plan/RED checkpoint before declaration work.

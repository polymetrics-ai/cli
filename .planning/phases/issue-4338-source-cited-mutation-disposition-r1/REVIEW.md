# Inline code review — issue 4338

## Method

`scripts/gsd prompt code-review 4338` was resolved on 2026-08-24. The
requested GSD reviewer role cannot run in this direct-PR adapter session, so a
standard-depth inline review covered the three production files and the
behavioural tests:

- `cmd/connectorgen/sourceimport.go`
- `cmd/connectorgen/sourceprojection.go`
- `cmd/connectorgen/sourceprojection_test.go`

The review traced each disposition from source-import document loading through
runtime-gap emission, source projection, and executable-coverage validation.
It also checked the action/command completeness predicates and the mutation vs.
read-only boundary.

## Findings

No critical, warning, or actionable informational findings.

## Evidence

- Input validation rejects blank, oversized, multiline, duplicate,
  non-mutating, and non-absolute citations before they can affect projection.
- Application requires an exact locked source operation, its provider URL/hash/
  byte/location provenance, and exact method/path citation.
- Application, projection, and executable coverage independently reject a
  complete action or matching `implemented` command claim.
- The tests exercise four distinct provider shapes plus the Vercel 159-row
  scale, not a connector-specific or declaration-count-only shortcut.
- No definitions or generated connector artifacts are changed, and the GitHub
  byte-stability test remains green.

## External review route

On PR creation the selected route is `claude_auto`: a non-draft PR opened by a
trusted repository author should trigger the repository's Claude review Action.
No manual Claude or Copilot request is appropriate unless that automatic route
fails or is skipped. The PR body will record the head SHA and route status.

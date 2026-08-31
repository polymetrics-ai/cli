# Inline code review — issue 4364

Reviewed against exact parent `687eb1ded6b42cc456f8cc3c1e97f0a84fd042a8`.

## Findings

No blocking finding in the scoped change.

## Reviewed invariants

1. **Source identity is fail-closed.** A row resolves by exact lock ID or by
   its explicitly retained provider operation ID plus exact method/path. Zero
   and multiple candidate matches fail; no prefix, schema-name, or HTTP-method
   fallback exists.
2. **Seven-lane accounting is complete.** Every decoded row must carry all
   seven fixed lanes. The report reconciles the frozen primary denominator,
   connector-local supplements, and exactly seven cells per row.
3. **Deferred visibility is non-executable.** M-U and M-F entries carry typed
   reasons/capabilities only. The new command has no output or runtime-artifact
   flag, no runtime dependency, and zero executable declarations.
4. **Semantic evidence remains source-led.** The mutation safeguard checks
   retained source semantics on the row or lane evidence, never HTTP method.
5. **Foundation evidence is truthful.** Existing M-F records resolve through
   the connector matrix/ledger and current Atlas; historical records without a
   source-specific seven-lane target are not misrepresented as a gap for an
   unrelated cell.

## Review limits

- No source lock, matrix, descriptor, runtime definition, engine, transport,
  certification, or Atlas file changed, so no runtime behavior was reviewed or
  claimed.
- A full `./cmd/connectorgen` suite was not used as the gate because this
  shared package has known unrelated Batch R1 baseline failures; focused normal
  and race tests exercise the new command and actual cohort directly.

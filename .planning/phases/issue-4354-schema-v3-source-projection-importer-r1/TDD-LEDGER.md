# Schema-v3 source projection/importer foundation — TDD ledger

## Planned red/green/refactor slices

| Slice | Red assertion before production code | Green assertion | Refactor guard |
| --- | --- | --- | --- |
| Schema-v3 admission | The immutable retained Outreach schema-v3 input fails to import/project with its existing reader path. | The exact lock bytes/digest, origin and every source operation survive canonical import. | No network/cache fallback, re-pin, or descriptor identity rewrite. |
| Shared reader proof | A second independent schema-v3 fixture fails in the same reader/importer boundary. | It produces a distinct cited operation descriptor through the same common code. | Unsupported/malformed kinds still fail with contextual errors and size/path bounds. |
| Six-lane projection | A recognized unsupported operation is omitted, incorrectly blocked as provenance, or promoted to `implemented`. | It has six visible lane classifications and a specific `missing_foundation` reason. | Existing executable operations keep their canonical mapping and preflight remains the implementation authority. |

## Red

Pending. Each command, failing output, and first failing test name will be
recorded before the corresponding production edit.

## Green

Pending. Each passing command will name the assertion that changed from red.

## Refactor / review

Pending. Record error-wrapping, defensive-copy, ordering, provenance, and
generic-escape review outcomes after the final implementation.

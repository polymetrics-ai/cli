# TDD ledger — request-contract execution envelopes

| ID | Guarantee | Required red assertion | Green proof |
| --- | --- | --- | --- |
| B1 | Valid missing common bounds do not abort source import | Gong-shaped `workspaceId` fails with `unbounded request schema string has no maxLength`. | The operation imports without a schema gap. |
| B2 | Source schema and PM policy remain distinct | New descriptor assertion cannot find an envelope and observes the old import abort. | Schema remains exactly source-owned; separate envelope carries policy version, origin, unit, default, hard, and effective cap. |
| B3 | Generation fails closed without a finite envelope | Sabotaged/removal path passes today because no envelope exists. | Descriptor/projection validation rejects missing, zero, inconsistent, or over-ceiling envelopes for executable common inputs. |
| B4 | Runtime enforces the exact encoded boundary before I/O | PM-labelled cap assertion fails under the old generic error contract. | Exact cap reaches local HTTP once; cap+1 and UTF-8 expansion reject with zero additional hits. |
| B5 | Numeric parsing is exact and byte-bounded | An integer beyond `int64` and a high-precision decimal are rejected or narrowed by `ParseInt64`/`ParseFloat64`. | `big.Int`/`big.Rat` validation accepts finite lexemes, sends identical bytes, and rejects invalid syntax before I/O. |
| B6 | Semantic ambiguity remains visible | A broad bypass would incorrectly erase dynamic/composition/untyped gaps. | Focused importer cases remain source-traced and merge-blocked. |
| B7 | Existing command limits do not silently drift | Projection snapshot detects any change to current 8 KiB/256/1 MiB compatibility behavior. | Existing bounded fixture remains semantically identical aside from additive descriptor provenance. |

## Red evidence

Pending. The first production-behavior change begins only after B1/B2 tests are
added and run failing with `GOFLAGS=-p=3`.

## Green evidence

Pending implementation.

## Deliberate sabotage evidence

Pending after green. The new envelope/enforcement path will be intentionally
broken, the focused new tests must fail, and the production code will then be
restored before final verification.

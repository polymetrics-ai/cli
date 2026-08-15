# #4077 — Transport record-isolation discussion log

> **Audit trail only.** `CONTEXT.md` records the operative decisions.

**Date:** 2026-08-13
**Mode:** `discuss-phase --auto` inline/manual named-phase fallback
**Reason:** The issue brief and exact-head reproduction already fix every material product and safety
decision; the official phase lookup has no numeric roadmap entry.

## Accepted value boundary

| Option | Description | Selected |
|---|---|---|
| Broad reflection clone | Copy arbitrary maps/slices/pointers dynamically. | |
| Explicit closed value contract | Clone only enumerated supported values; reject everything else. | ✓ |
| Preserve unknown values | Return the original `any` value by default. | |

**Decision:** explicit closed value contract. It preserves accepted Transport semantics without
silently passing an unknown mutable reference value.

## Omitted concrete values

| Option | Description | Selected |
|---|---|---|
| Leave as historical `[]byte` only | Retain the #4047 value set. | |
| Add explicit `json.RawMessage` and `map[string]string` copy cases | Preserve both proven omitted values and their nested use. | ✓ |
| Normalize provider values elsewhere | Alter provider/warehouse payload behavior. | |

**Decision:** add only the two demonstrated value forms at the Transport clone boundary.

## Rejection behavior

| Option | Description | Selected |
|---|---|---|
| Panic/drop unsupported values | Make the data disappear or terminate unexpectedly. | |
| Fail with boundary context before stage/apply | Keep the source intact and prevent the unsafe crossing. | ✓ |
| Forward unknown value untouched | Preserve the current aliasing hazard. | |

**Decision:** contextual error before the relevant downstream boundary.

## Evidence standard

| Option | Description | Selected |
|---|---|---|
| Direct helper test only | Prove dynamic-type omission without orchestration. | |
| Direct plus stage/destination mutation regression | Prove the visible source-record symptom at both real copy boundaries. | ✓ |
| Credentialed provider/database run | Exercise unrelated external systems. | |

**Decision:** focused deterministic unit/fake evidence; no credentialed claim.

## Deferred ideas

None.

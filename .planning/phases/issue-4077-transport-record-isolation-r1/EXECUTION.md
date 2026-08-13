# #4077 — execution record

## Status

The RED regression is failing as expected. No production source has changed.

## Next action

Commit the focused RED regression and its evidence, then implement the minimal closed clone correction.

## RED outcome

`json.RawMessage` and `map[string]string` were still returned by the `cloneRecordValue` default path
and therefore shared source-owned storage. A `map[string]int` also crossed each boundary without
rejection. The failure is behavioral, not a compile/setup failure.

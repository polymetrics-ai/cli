# DISCUSSION LOG — issue #3852

Mode: `scripts/gsd prompt discuss-phase issue-3852-output-policy-declaration-r1 --auto`.

The issue, repository migration rules, and parallel-lane ownership settle every material choice:

- The non-redacting declaration values are the already implemented direct-write policies `none`
  and `json`; no runtime behavior needs changing.
- The enum remains closed and retains all existing narrow values for compatibility. The direct-read
  and direct-write lists are reconciled by their union, with `binary_file_bounded` retained only
  for its existing binary-download declaration compatibility.
- The red test is an in-memory bundle load with a `direct_write` CLI command declaring `json`; it
  asserts the runtime validates and returns the decoded body before asserting the schema load.
- No connector is a target of this foundation, so the bundle fleet is deliberately untouched.
- The authoring documentation must favor `json` for complete write results and `none` when the
  caller requires no response body. It must not tell new authors to choose redacting policies by
  default.

There are no unresolved product or safety decisions. The task forbids live provider calls and
credentials; fixtures and unit tests are sufficient.

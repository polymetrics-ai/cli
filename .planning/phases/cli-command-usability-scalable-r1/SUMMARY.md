# Issue #4193 summary — scalable command help

## Delivered

- Replaced the `legacyLeafManualTopic` two-command switch with one shared
  wrapper help path. It derives valid invocations from the same manual text
  used for runtime help and generated `docs/cli` pages.
- Added complete manuals for previously undocumented legacy wrappers: `init`,
  `help`, `man`, `extract`, and `worker`.
- Kept unknown top-level and leaf commands as usage errors even when a help
  flag is present; malformed approval-token carriers remain validated before
  any help resolver.
- Kept the pre-existing closed ETL transport manual as a declared resolver on
  its wrapper, rather than an `if command == ...` router branch. It is the
  one residual contextual manual because its selected transport name and
  security contract are dynamic.
- Replaced the two-row changed-command test with a registry/manual-derived
  legacy sweep, plus all 8,900 generated connector command leaves.

## Proof

- Built binary legacy sweep: 63 paths × two help spellings × empty/initialized
  project = 252 successful requests.
- Built binary dynamic-root sweep: 36 roots × two help spellings ×
  empty/initialized project = 144 successful requests.
- Generated connector-leaf resolver sweep: 8,900 paths × `--help`/`-h` =
  17,800 successful `NAME` manual renders before dispatch.
- Complete `internal/cli` and `cmd/connectorgen` package suites passed; all
  required individual repository gates are recorded in `VERIFICATION.md`.

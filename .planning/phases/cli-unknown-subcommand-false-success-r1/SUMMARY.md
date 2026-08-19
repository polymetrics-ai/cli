# Summary — cli-unknown-subcommand-false-success-r1

## Delivered

- Resolved connector help paths are now validated before a manual is rendered. An unresolved
  command beneath a known connector returns the complete-path `usage_error` at exit 2 instead of
  silently returning the connector root manual at exit 0.
- Existing connector root/bare help, one-segment group help, and declared deep-command help retain
  their successful output and exit code.
- Updated the canonical website CLI and agent documentation so a `CommandManual` is promised only
  for resolved help paths, then regenerated the website docs data.
- Added a red/green regression covering a real deep command, invalid deep command, and the JSON
  envelope. Added the corresponding reviewed golden transcript entry.

## Manual binary probe

Using the freshly built `./pm` binary:

```text
pm gong calls definitely-not-real --help -> exit 2
pm gong calls transcript --help          -> exit 0
pm copper people list --help             -> exit 2
pm gong --help                           -> exit 0
pm gong                                  -> exit 0
```

Exactly **one** previously-passing bogus invocation was exercised and now correctly fails:
`pm gong calls definitely-not-real --help`.

## Deferred delivery stage

The task contract assigns the subsequent no-mistakes validation and PR opening to firstmate after
this committed handoff. No PR or push was attempted from this worker.

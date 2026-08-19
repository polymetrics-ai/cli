# CLI root-manual snapshot RED/GREEN

## RED

`go test -timeout 20m ./internal/cli -count=1` initially exited 1 after the
generated command surfaces added `avni`, `oura`, `perigon`, and `pingdom` to
the root connector manual. Only the nine root manual transcript cases failed:
`root_bare_manual`, `root_long_help`, `root_short_help`, `root_help_command`,
`root_man_command`, `root_json_help`, `root_late_json_help`, `root_equals_form`,
and `root_space_form`.

## GREEN

The standard targeted generator update passed:

```text
POLYMETRICS_UPDATE_GOLDEN_TRANSCRIPTS=1 \
POLYMETRICS_GOLDEN_TRANSCRIPT_NAMES='root_bare_manual,root_long_help,root_short_help,root_help_command,root_man_command,root_json_help,root_late_json_help,root_equals_form,root_space_form' \
go test -timeout 20m ./internal/cli -run '^TestGoldenTranscripts$' -count=1
# ok polymetrics.ai/internal/cli 7.515s
```

The scoped golden test passed again without update mode (9.380s), and the
complete package rerun passed in 534.660s. The full `make verify` gate also
passed after the snapshot refresh.

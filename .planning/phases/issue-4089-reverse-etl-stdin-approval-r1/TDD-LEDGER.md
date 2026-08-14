# #4089 — TDD ledger

**Status:** planned; production files remain untouched.

| Checkpoint | Evidence | Result |
| --- | --- | --- |
| Plan | GSD prompts resolved inline; issue decisions, scope, parity, and safety checks recorded before production edits. | pending commit |
| Red: bounded stdin carrier | New real-binary CLI test is added before implementation and is expected to reject the current argv-only behavior. | pending |
| Green: generic wiring | Both request construction paths use the existing reader; existing approval and replay tests pass through stdin. | pending |
| Refactor: docs and safety | Generated manuals, website data, transcripts, and stale-syntax scan are green. | pending |

## Required red/green record

- Red: run the new focused test against the unchanged production implementation and record its non-zero exit plus the expected failure reason here.
- Green: rerun the same selector after the minimal production change and record its zero exit, stdout-safe process argv observation, surface assertions, and replay result here.

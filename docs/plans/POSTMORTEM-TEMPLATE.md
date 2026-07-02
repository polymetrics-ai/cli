# Postmortem — <title>

- Date of incident/failure: <YYYY-MM-DD>
- Phase / wave: <e.g. wave0-engine-harness / Wave F>
- Authors: <roles/agents involved>
- Status: draft | reviewed | closed

## Summary

One paragraph: what broke, blast radius, how it was detected, how it was resolved.

## Impact

- Gates affected (build/test/lint/conformance/certify/parity):
- Connectors/bundles affected:
- Data or resource leaks (certify ledger reference, if any):
- Time lost / repair-agent runs consumed:

## Timeline

| Time (UTC) | Event |
|---|---|
| | first bad commit / dispatch |
| | detection (which gate/test caught it — or failed to) |
| | mitigation / revert |
| | resolution verified |

## Root cause

Technical root cause, plus the process cause (e.g. prompt-template defect replicated by fan-out
agents, missing seeded-invalid class, gate ordering).

## What went well / what went poorly

## Action items

| # | Action | Type (gate/test/prompt/docs/engine) | Owner | Done |
|---|---|---|---|---|
| 1 | | | | |

## Ledger updates

- conventions.md patched? (ref)
- seeded-invalid corpus extended? (class)
- quarantine.json entries? (names)

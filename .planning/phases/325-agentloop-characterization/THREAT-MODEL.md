# Threat Model: Phase 0

## Trust boundaries

- Fixture path and bytes are untrusted local input.
- Driver arguments and environment are untrusted configuration.
- Existing prompt/state/session artifacts are untrusted and must not be read while the fuse is
  closed.

## Primary threats and controls

| Threat | Control |
| --- | --- |
| Fuse bypass through env/argument/state | hard-coded closed status; no enable parser or command |
| Wrapper bypass | canonical tracked inventory plus marker/inventory equality test |
| Persistence before denial | driver guard before repo root state setup and prompt read/write |
| Raw transcript ingestion | `.jsonl` rejected before open; strict bounded `.json` fixture shape |
| Secret or private-path leakage | structural string scanner and no arbitrary event payload |
| Expectation echo masquerading as validation | semantic event matcher derives violation first |
| Resource exhaustion | maximum fixture size, bounded fixture count, bounded events and strings |
| Ambiguous automation result | stable JSON fields and distinct 64/65/78 process exits |
| Symlink/path surprise | regular-file check; directory loader rejects symlink entries |

The blanket fuse intentionally sacrifices availability for safety. Phase 1 may replace it only
with brokered capabilities proven by dependent tests; Phase 0 has no waiver.

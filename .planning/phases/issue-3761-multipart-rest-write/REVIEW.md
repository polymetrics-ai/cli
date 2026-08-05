---
status: clean
reviewer: inline-manual-fallback
depth: standard
files_reviewed: 12
findings:
  critical: 0
  warning: 0
  info: 0
  total: 0
---

# Code review — issue #3761 multipart `rest_write`

The GSD reviewer role was not spawned because the worker brief forbids role
spawning. The changed engine, requester, app lifecycle, schema, docs, and
loopback tests were reviewed inline against the parent and child contracts.

## Findings resolved during review

1. The app previously discovered upload identities only from field names that
   contained `file_path`. A valid declared multipart source named `upload`
   therefore reached preview without an approved digest. The retained red test
   failed with that error; the fix exports the closed multipart file-field list
   through direct-write metadata and hashes only those exact declared fields.
2. `OperationDirectWriteMetadata` accepted a fixture with no endpoint
   declaration, allowing the real `commandrunner.Preflight` path to accept an
   implemented command. The retained red mutation set `bundle.Surface = nil`;
   the engine now rejects it before preflight can pass. The shipped-registry
   regression keeps the operations-derived endpoint projection covered.
3. Lint found one ignored test-file close result. The loopback fixture now
   explicitly discards that cleanup error; `make lint` is clean.

## Review result

No unresolved correctness, safety, retry, redirect, capability, or
no-redaction findings remain. The changed-package tests and non-aggregate gates
listed in `VERIFICATION.md` passed after the fixes.

# TDD ledger — issue #4342 binary upload CLI and certification foundation

## Red

- [ ] `commandrunner`: a new `binary_upload` intent cannot resolve unless it references an implemented write action with `body_type` `binary_upload`, `base64_upload`, or a multipart file part; arbitrary write paths and JSON/form actions fail before I/O.
- [ ] `internal/cli`/`app`: a GitHub-style binary upload command creates a plan and does not directly dispatch; preview/approval binds source bytes and later file mutation prevents provider I/O.
- [ ] `certify`: refusal produces `blocked`/`not_live`, never an upload pass; transfer test requires exact byte count/digest, provider response, independent read-back, and cleanup.
- [ ] `connectorgen`: a `binary_upload` command projects to a dedicated binary-upload sweep/evidence class; `file_upload` stays rejected as executable.

## Green

- [ ] Production implementation for every red assertion.

## Refactor

- [ ] Reuse the existing write-command plan and engine binary payload preparation; no parallel raw upload/requester or generic CLI surface.
- [ ] Keep report data typed and capability names explicit rather than inferring upload from command strings or a single `binary` cell.

## Deliberate break proof

- [ ] Temporarily remove `binary_upload` planner admission, run the commandrunner/CLI behavior test, and record its failure.
- [ ] Restore the admission, rerun it, and record success.

## Commands/results

Pending implementation.

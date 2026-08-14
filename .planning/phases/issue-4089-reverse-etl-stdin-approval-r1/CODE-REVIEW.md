# #4089 — inline code review

## Scope review

- The app approval model was not changed: plan status, persisted preview,
  destructive confirmation, token hashing, and exact-once consumption remain in
  `internal/app.RunReverseETL`.
- Both CLI request builders obtain their value through the same existing
  `readApprovalTokenFromStdin` carrier. There is no connector-specific branch
  or GitHub literal in the CLI/app mechanism.
- Certification's in-process harness only injects OS stdin for its existing
  public-CLI tests; it does not parse or persist an approval token.

## Security review

- `--approve` is refused before execution and its supplied value is never
  formatted into an error.
- `--approval-token-stdin` accepts no value, and the reused 4 KiB one-line
  reader rejects empty, oversized, and multiline input
  before either reverse plan lookup path reaches write dispatch.
- The real-binary regression records the observed `ps` command line and checks
  argv, environment, project files, captured logs, receipt, and evidence
  independently for token absence.

## Outcome

No unresolved correctness, security, portability, or parity finding.

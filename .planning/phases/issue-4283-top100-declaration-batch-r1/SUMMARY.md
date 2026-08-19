# Increment 001 summary

Ten ranked daily-use API connectors have reproducible public-source locks and
exact source-to-`api_surface` declaration records. This remains a
declaration-only, non-live-certification checkpoint.

- Public operations pinned and declared: 4,378 / 4,378 (100%).
- Docker Hub runnable retrofit: 41 commands, 20 typed write actions, and six
  typed delete actions across 54 pinned operations.
- Docker Hub disabled dispositions: 13 (ten named foundation gaps and three
  schema/media incompatibilities; `unsafe-to-exercise` is zero).
- All ten connectors retain the recoverable #4093 `sync_transport` gap; no
  generic transport descriptor was invented.
- No credential was requested or used; all live certification remains pending.

## Docker Hub corrected deliverable

Docker Hub is the first runnable-parity proof slice. Its 50 operation contracts
(23 `rest_read`, 27 `rest_write`) are backed by four existing ETL commands, 17
operation-bound direct reads, and 20 reverse-ETL commands linked to typed
writes. The reverse-ETL surface preserves plan, preview, approval, and execute;
it does not expose a generic HTTP mutation path.

The public OpenAPI is the only schema source. Required path-item parameters are
derived exactly from the pinned document. The `params-import` limitation that
does not merge those inherited parameters is recorded as a recoverable
foundation/generator gap rather than concealed by manual flag invention. The
token creates and the three credential exchanges remain source-declared
foundation gaps because the current engine cannot safely execute
`sensitive_policy` writes or return a secret response.

Verification evidence is recorded in `VERIFICATION.md`, source pin evidence in
`SOURCE-LOCK-VERIFICATION.json`, and exact operation rejections in
`REJECTION-LIST.json`. Complete the current local gates and commit this Docker
Hub slice before selecting another connector. Gitea is not a definition bundle
on `main` and must be created or substituted before a later increment.

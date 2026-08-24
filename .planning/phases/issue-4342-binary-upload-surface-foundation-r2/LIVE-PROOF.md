# GitHub binary-upload live proof — 2026-08-24

This is observed provider evidence for the public `binary_upload` path. It is
deliberately not a generated matrix-evidence record: the generic certification
runner has no transfer/read-back/cleanup input contract for one-file uploads,
so its upload stage remains `not_live`, never `pass`.

## Historical fixture disposition

The first disposable fixture, draft tag `v0.0.1-upload-test` (release ID
`375528670`), was deleted during its cleanup after the earlier read-back. That
is why it no longer appeared when Firstmate audited the repository. Its removed
state is not used as current proof; the fresh, empty draft below is retained so
the completed revalidation remains independently inspectable.

## Fresh real-provider revalidation

- Repository: `karthik-sivadas/pm-binary-upload-testbed`.
- Retained disposable draft: tag `v0.0.2-binary-upload-proof-4343`, release ID
  `375616321`.
- Command: `pm github releases assets upload`, using the encrypted local
  `github-live-upload-proof` credential. No credential value, approval token,
  or Authorization header was printed or written.
- The first plan printed `Preview required before an approval token is issued.`
  Preview re-supplied the withheld `--file-path`, persisted the binary digest,
  and issued its one-time token only in human output. That token was passed to
  `pm reverse run --approval-token-stdin` through a shell variable and was not
  printed. The successful plan/run IDs are `rplan_56a3733e5826f19d` and
  `rrun_d63170008bde387e`.

### Happy path and independent read-back

| Fact | Actual result |
| --- | --- |
| Source / transmitted bytes | 32 |
| Source SHA-256 | `66687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f2925` |
| Provider result | completed run, `records_succeeded=1`, one retained HTTP `201` response |
| Provider asset | ID `527459889`, `state=uploaded`, `size=32`, `digest=sha256:66687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f2925` |
| Independent read-back | `gh-axi release download` followed by `cmp`: 32 bytes and the same SHA-256 |

The retained provider response names `uploads.github.com`,
`application/octet-stream`, and the exact source digest. The independent
download therefore proves byte equality, not merely a `201` status.

### Safe-refusal attempts

Each attempt used the real public command against release `375616321`, then
queried GitHub's release asset ledger immediately afterward. The command stops
during planning before it can produce a provider receipt; the one actual asset
remained the happy-path 32-byte asset until cleanup.

| Attempt | Actual command result | Provider ledger after refusal |
| --- | --- | --- |
| 64 MiB + 1 byte file | `error: payload identity for file_path: payload file exceeds declared byte cap 67108864` (exit 1) | only asset `527459889`, 32 bytes |
| Arbitrary `--content-type image/png` | `error: unknown flag --content-type for command "releases assets upload"` (exit 1) | only asset `527459889`, 32 bytes |
| Missing file | `error: payload identity for file_path: lstat …/missing-live-4343.bin: no such file or directory` (exit 1) | only asset `527459889`, 32 bytes |

The media refusal is intentionally parser-level: the public binary-upload
surface accepts no caller-controlled media type at all. The runtime additionally
enforces that a raw/base64 declaration admits only the executor's fixed
`application/octet-stream`; the engine behavioral test arms a transport double
and proves that declaration-level refusal makes zero calls. The real attempts
above establish that no arbitrary media value can enter the public provider
path; the provider ledger establishes no partial asset appeared. They are not
packet capture, so this record does not claim a separate raw-socket trace.

### Cleanup

`gh-axi api DELETE /repos/karthik-sivadas/pm-binary-upload-testbed/releases/assets/527459889`
returned an empty successful response. A subsequent release query reported the
same draft release with `assets: []`. The draft is intentionally left empty for
audit; no proof asset or partial asset remains.

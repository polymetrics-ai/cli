# GitHub binary-upload live proof — 2026-08-24

This is an observed, bounded provider proof for the new `binary_upload` public
path. It is deliberately not an accepted matrix-evidence record: the generic
certification runner has no transfer/read-back/cleanup input contract for this
one-file upload yet, so it continues to report `not_live` rather than treating a
plan or this external proof as a generated certification pass.

## Scope and non-secret identity

- Provider target: the dedicated disposable draft release in
  `karthik-sivadas/pm-binary-upload-testbed`, tag `v0.0.1-upload-test`.
- Public command: `pm github releases assets upload` using the encrypted local
  credential `github-live-upload-proof`; the credential value, approval token,
  Authorization header, and response body were neither printed nor persisted.
- PM binary SHA-256:
  `542b3eb808ce4050dabdd1af26b14633d79c5ae2015c90953bbba3300b8e0f5f`.
- Provider run: `rrun_1216b926793cf619` completed after plan
  `rplan_211d62e73f9d1da9` was previewed and its human-only token was passed on
  stdin.

## Observed transfer and independent read-back

| Fact | Observed value |
| --- | --- |
| Upload host | `uploads.github.com` |
| Asset name | `binary-upload-live-4343.bin` |
| Transmitted bytes | 32 |
| Source SHA-256 | `66687aadf862bd776c8fc18b8e9f8e20089714856ee233b3902a591d0d5f2925` |
| Provider receipt | one retained response, HTTP `201`; `records_written=1` |
| Independent read-back | `gh-axi release download` fetched the named asset; `cmp` succeeded; 32 bytes and the same SHA-256 |

The plan was first created with a project-root-relative source path and had no
approval token. Preview then persisted the digest and issued the one-time human
token; execute sent the exact approved bytes. This proves the real CLI/App
lifecycle and public upload-host authentication boundary, while the focused
tests cover unsafe/missing/changed/oversize pre-I/O refusals.

## Cleanup ledger

After read-back, `gh-axi release delete v0.0.1-upload-test --repo
karthik-sivadas/pm-binary-upload-testbed --yes` removed the dedicated draft
release and its proof asset. `gh-axi release view` for that exact tag then
returned not found. No provider resource remains from this proof.

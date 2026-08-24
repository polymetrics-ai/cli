# Batch-1 retained-source re-pin report — 2026-08-24

Authority: Firstmate inbox `009.msg` explicitly authorizes a re-pin only where
the response is a real provider document. The pre-retain lock copies are in
this directory and were proven byte-identical to their original locks before
any source fetch or lock edit.

The source locks do not carry reviewed license or terms text. Retention uses
the command's permitted provenance values `license: undetermined` and
`terms: undetermined`; this records that no provider-specific redistribution
claim was inferred from a v2 lock while retaining the source identity.

| Connector | Source kind / direct URL result | Pre-retain lock backup | Old identity | Observed real document | Re-pin disposition |
| --- | --- | --- | --- | --- | --- |
| GitLab | OpenAPI YAML; direct HTTP/2 200, no redirect, `Content-Type: text/plain; charset=utf-8`; `openapi: 3.0.0` | `gitlab-operation-source-lock.pre-retain.json` (SHA-256 `dcd25881e036792d6fc75485ae00d310a1cf416696aa36fd6d1c52bafb7331da`) | 3,576,860 bytes; `6b6ad591ff1b54ab429d0502812a2b2955501f1f6bebdae1888ba0bea086cf82` | 3,611,238 bytes; `b33e80759af3a529d71a0bb58c8c76d65c9f0b20774a042196c6d3c0c57310bd` | Authorized real-document drift; replace only `rest.bytes` and `rest.sha256`, then retain and hermetically import. |
| Jira | OpenAPI JSON; direct HTTP/2 200, no redirect, `Content-Type: text/plain; charset=utf-8`; `openapi: 3.0.1` | `jira-operation-source-lock.pre-retain.json` (SHA-256 `4eb8aaf19b59b7581138dc8bd7e353c61ad819010dfcae605a53c5fe18a5f661`) | 2,456,011 bytes; `511d0b97390cc47aa0e1367189210a41f32088d9c869e7bb01f43698bdf7e5e8` | 2,456,011 bytes; `e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf` | Authorized real-document drift; replace only `rest.sha256`, then retain and hermetically import. |

Probe commands were credential-free and did not follow redirects:

```text
curl --silent --show-error --fail --max-redirs 0 --connect-timeout 15 --max-time 90 \
  --dump-header <headers> --output <artifact> \
  --write-out 'status=%{http_code} content_type=%{content_type} url_effective=%{url_effective} redirect_url=%{redirect_url}' \
  <lock rest.source_url>
```

Neither re-pin is an error page, login wall, or redirect. The raw probe files
are transient local verification material; the only retained source inputs are
the content-addressed connector-owned artifacts written by `source-retain`.

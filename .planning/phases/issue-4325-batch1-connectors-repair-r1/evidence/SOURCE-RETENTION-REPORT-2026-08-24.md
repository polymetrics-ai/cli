# Batch-1 retained source artifact report — 2026-08-24

This report follows the landed `source-retain` contract from #4348. Before any
fetch, every v2 source lock was copied to a same-named
`*.pre-retain.json` file in this evidence directory and byte-compared. The
copies' SHA-256 values are in the companion re-pin report where a lock changed.

All maintenance commands were credential-free, sequential, and used the
connector-owned URL already present in the lock:

```text
GOMAXPROCS=2 go run ./cmd/connectorgen source-retain <connector> \
  --retrieved-at <recorded-RFC3339> --license undetermined --terms undetermined
GOMAXPROCS=2 go run ./cmd/connectorgen source-import <connector> --check
```

`undetermined` is intentional provenance, not an inferred license: the v2
locks have no reviewed provider license/terms field and this lane did not make
a redistribution claim from a public documentation URL alone.

| Connector | Retained source / identity | Retained file | Retention disposition | Hermetic import result |
| --- | --- | --- | --- | --- |
| Docker Hub | `https://docs.docker.com/reference/api/hub/latest.yaml`; 148,322 bytes / `99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756` | `sources/artifacts/99d9d53c2d93656a3c66d604885abd153dc5df285abc0ecb13802a3bc53d0756.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | fails: `components.schemas["team_repo"]` points to unresolved `#/components/responses/team_repo`. |
| Notion | `https://developers.notion.com/openapi.json`; 1,304,814 bytes / `dee5763763b0b9fbad2aa8d5adb173ca350ec26dda557e658c5dbe9d2ea2f258` | `sources/artifacts/dee5763763b0b9fbad2aa8d5adb173ca350ec26dda557e658c5dbe9d2ea2f258.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | fails: `PATCH /v1/blocks/{block_id}/children` exceeds retained request-media descriptor bytes. |
| Stripe | `https://raw.githubusercontent.com/stripe/openapi/master/openapi/spec3.json`; 7,967,776 bytes / `3653ad45bbec54fcbe461c541c908355b715018bdf455a0e11b27bedb2cbdee5` | `sources/artifacts/3653ad45bbec54fcbe461c541c908355b715018bdf455a0e11b27bedb2cbdee5.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | fails: `GET /v1/account` response `200` exceeds schema depth. |
| Bitbucket | `https://developer.atlassian.com/cloud/bitbucket/swagger.v3.json`; 1,359,673 bytes / `3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3` | `sources/artifacts/3dbfe6a80143511a287e58c21a193d3551ab5d41e8b60e65c1ae121b7000dec3.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | verified after canonical regeneration: 297 operations, 0 inbound events. |
| GitLab | `https://gitlab.com/gitlab-org/gitlab/-/raw/master/doc/api/openapi/openapi_v3.yaml`; 3,611,238 bytes / `b33e80759af3a529d71a0bb58c8c76d65c9f0b20774a042196c6d3c0c57310bd` | `sources/artifacts/b33e80759af3a529d71a0bb58c8c76d65c9f0b20774a042196c6d3c0c57310bd.artifact` | re-pinned under inbox 009 authority; retained at `2026-08-24T12:01:06Z`; old identity is in `REPIN-REPORT-2026-08-24.md`. | verified after canonical regeneration: 1,752 operations, 0 inbound events. |
| CircleCI | `https://circleci.com/api/v2/openapi.json`; 621,321 bytes / `61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07` | `sources/artifacts/61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | verified after canonical regeneration: 111 operations, 0 inbound events. |
| Sentry | `https://raw.githubusercontent.com/getsentry/sentry-api-schema/main/openapi-derefed.json`; 3,868,570 bytes / `b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435` | `sources/artifacts/b71216654e44cc18f5e262fbb5075df67f1504a123d4bcb51cc8e8cc74ebd435.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | verified after canonical regeneration: 223 operations, 0 inbound events. |
| Vercel | `https://openapi.vercel.sh/`; 10,463,249 bytes / `74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28` | `sources/artifacts/74cb7ff3dc0b89cc344b13ac9c6d5f1d9b7d7a9356cfd6b5a779da51fd43da28.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | verified after canonical regeneration: 400 operations, 0 inbound events. |
| Asana | `https://raw.githubusercontent.com/Asana/openapi/56796a67a3c093eedf55fd9682357957a2ebfd85/defs/asana_oas.yaml`; 3,066,750 bytes / `cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56` | `sources/artifacts/cb3b90f4e0af56035eab0c648974f625b942a28a7144aa6c2326e38ca0bb3d56.artifact` | preserved; retained at `2026-08-24T11:55:22Z` | fails: 25 source operations have no complete executable action. |
| Jira | `https://dac-static.atlassian.com/cloud/jira/platform/swagger-v3.v3.json`; 2,456,011 bytes / `e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf` | `sources/artifacts/e7136af43bf72cd4ea5ada91ec665b318b60008814122461d4436a43b6c732bf.artifact` | re-pinned under inbox 009 authority; retained at `2026-08-24T12:01:06Z`; old identity is in `REPIN-REPORT-2026-08-24.md`. | fails: 14 source operations have no complete executable action. |

Every retained source received its required connector-owned provenance manifest
at `sources/<connector>-retained-artifacts.json`. Eight identities were
preserved; GitLab and Jira are the only re-pins. The re-pins were both direct
HTTP/2 200, no-redirect OpenAPI documents, not error/login bodies; their old
and new identities, OpenAPI markers, and pre-change backup hashes are recorded
in [REPIN-REPORT-2026-08-24.md](REPIN-REPORT-2026-08-24.md).

The final hermetic split is 5/10 verified imports (Bitbucket, GitLab,
CircleCI, Sentry, Vercel), with five genuine remaining source/action blockers
(Docker Hub, Notion, Stripe, Asana, Jira). Retention itself is complete for all
ten and no live fetch is needed for normal source-import verification.

## Built-binary reachability after retained regeneration

The main-merged runtime-valid generated-command-path foundation was proved with
the built `pm` binary in a separate initialized project per connector, with no
credential configured. Every command below stopped at exactly
`error: missing --credential`; none returned `unknown command` or a path-segment
error.

| Connector | Implemented commands exercised | Credential-bound | Other result |
| --- | ---: | ---: | ---: |
| Bitbucket | 50 | 50 | 0 |
| GitLab | 4 | 4 | 0 |
| CircleCI | 43 | 43 | 0 |

The Bitbucket proof specifically covers all 28 paths that were previously
rejected because source-derived command segments contained `{parameter}` text.
Canonical source import regenerated those paths (`cli=28`); no connector-local
renaming was used.

# Overview

Reply.io is a declarative HTTP bundle for the legacy Reply.io v1 API used by the existing Go connector and the current Reply.io v3 API published at docs.reply.io. The legacy Go package still emits four v1 streams (people, campaigns, tasks, and email accounts) through `X-Api-Key`; this bundle preserves those streams under `/v1/...` paths. The current v3 OpenAPI file (`https://docs.reply.io/api-reference/bundled.yaml`, last fetched for this pass on 2026-07-04) adds 81 documented GET streams and 189 write actions.

The legacy package remains registered and unchanged until the wave6 registry flip.

## Auth setup

Provide `api_key` as a secret for legacy v1 reads. It is sent as `X-Api-Key`, matching `internal/connectors/reply-io`.

Provide `bearer_token` as a secret for current v3 streams and writes. The v3 docs require `Authorization: Bearer <api key>`; when `bearer_token` is present, the engine selects bearer auth for the whole runtime. Both secrets are marked `x-secret: true`.

`base_url` defaults to `https://api.reply.io`. Legacy streams include `/v1` in their paths; v3 streams and writes include `/v3` in their paths. Path parameters such as `id`, `sequence_id`, and `contact_id` are declared as config fields for v3 read streams and as record fields for write actions.

## Streams notes

The four legacy streams remain first in `streams.json` and retain their legacy root-array projection plus the static `stream` marker computed field. Their page-number pagination remains `page`/`limit` with page size 100, matching legacy defaults and existing two-page people fixtures.

Every current documented v3 GET operation that is not marked coming-soon in the OpenAPI becomes a stream. Paginated v3 list endpoints use `offset_limit` with `top`/`skip` when both parameters are documented. Endpoints returning an `items` envelope read `records.path: items`; root arrays read `records.path: ""`; detail-style object responses use root-object extraction. New v3 streams use `projection: passthrough` with permissive schemas so fields from the vendor response are preserved.

Optional v3 query filters beyond pagination are intentionally not declared. The conformance harness synthesizes every declared spec property, so broad optional filters would be sent with synthetic values in fixture replays; this bundle models the endpoint and pagination contract without pretending to validate every optional query combination.

## Write actions & risks

`writes.json` contains 189 v3 actions for JSON-object or empty-body POST/PUT/PATCH/DELETE operations that the engine can express. Delete actions are marked destructive and idempotent for 404. Each action has a record schema for path/body fields, a fixture under `fixtures/writes/`, and a per-action risk string naming the external mutation.

Excluded write surfaces are listed in `api_surface.json`: multipart/binary uploads, root-array body operations, required-query mutations, credential-bearing connectivity tests, coming-soon endpoints, and semantic POST-body reads/reports.

## Known limits

- **Auth is selected bundle-wide.** Reply v1 uses `X-Api-Key`; Reply v3 uses Bearer auth. The engine cannot choose auth per stream or per action. Supplying `bearer_token` selects v3 auth for the runtime; omitting it keeps legacy v1 auth.
- **V3 write schemas are permissive.** The OpenAPI contains large nested request schemas. This pass models path fields and top-level object body fields, but does not recursively validate every nested property. The write plan/preview/approval flow remains required before execution.
- **Root-array, multipart, and required-query writes are blocked by dialect gaps.** The write engine builds object bodies and has no per-action query or multipart/body-array mode, so those operations are excluded and tracked in quarantine.
- **Semantic POST-body reads are excluded.** The read engine currently ignores `stream.body`, so filter/report endpoints that are reads expressed as POST bodies cannot be modeled as streams without a hook or engine support.
- **Legacy defensive fallback mapping is narrowed.** Legacy can synthesize missing ids from alternate fields and try multiple response envelopes. The declarative streams use fixed record paths and schemas; v3 streams use passthrough projection to avoid dropping vendor fields.

# Phase 505 — verification checklist

Local gates (all green):

- [x] `gofmt -w cmd internal` — clean.
- [x] `go vet ./internal/... ./cmd/...` — clean.
- [x] `go build ./cmd/pm` — builds.
- [x] `go run ./cmd/connectorgen validate internal/connectors/defs` — `548 connector(s), 0 findings`.
- [x] `go test ./internal/connectors/conformance/ -run TestConformance/whatsapp` — `ok`.
- [x] `go test ./internal/connectors/bundleregistry/ ./internal/connectors/defs/ ./cmd/connectorgen/` — `ok`.
- [x] `go test ./internal/cli/` — `ok` (golden transcripts + catalog count regenerated).
- [x] `make docs-check` — `Validated connector docs in docs/connectors`.

CLI help / manual / docs / website parity
(per `.agents/agentic-delivery/references/cli-help-docs-website-parity.md`):

- [x] `pm connectors inspect whatsapp --json` — renders; no secret values; icon falls back to
      `icons/pm-sample.svg`.
- [x] `pm whatsapp` (bare namespace) — renders the connector manual and exits 0.
- [x] Runtime help catalog count updated (`internal/cli/docs.go`: 552 / 548).
- [x] `docs/cli/connectors.md` count mirror updated.
- [x] `docs/connectors/whatsapp/{MANUAL.md,SKILL.md}` generated; catalog `all-connectors.json`
      (552), `all-connectors.md`, and `README.md` given a scoped whatsapp entry only (the
      pre-existing drift of the other 450 connector docs was deliberately NOT regenerated — out of
      scope for this issue).
- [x] Website generated connector data regenerated to 548 (`npm run gen:bundles/gen:catalog/
      gen:connectors`): `website/data/connectors.generated.json`,
      `website/lib/connectors.catalog.{data.generated.json,generated.ts}`,
      `website/lib/connectors.generated.ts`.

Safety / policy:

- [x] No secret values requested, printed, or stored; `access_token` is `x-secret`.
- [x] No generic raw HTTP write, raw Graph method/path/body, or raw whatsmeow escape hatch.
- [x] Message sends, media upload, destructive deletes/unsubscribe, and phone-number
      register/deregister/two-step-PIN are typed reverse-ETL with `confirm: destructive`, risk +
      approval text, and PHI-redaction language. Read receipts, typing indicators, template
      create/edit, profile update, code request/verify, app subscribe, and QR create are
      approval-gated only; metadata/docs wording matches that split.
- [x] Messaging/conversation/pricing analytics are declared `unsupported_api` (Graph exposes them
      only as field expansions, which the declarative engine cannot author without a raw `fields`
      flag); `analytics template` executes over the flat-query `template_analytics` edge.
- [x] Access mode scoped by `spec.mode` (cloud|web); the two modes' credentials are never conflated.

Not run locally (environment/human gates):

- `make verify`'s full `go test -timeout 20m ./...` and `lint` — deferred to the no-mistakes
  pipeline / CI (some runtime tests need local Podman/Postgres/Dragonfly services).
- Live Cloud API / whatsmeow credentials, live sends, template submission, media payloads beyond
  fixtures, and healthcare consent sign-off — human gates per #505.

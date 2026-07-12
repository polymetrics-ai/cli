# TDD Ledger — Issue #186

## GSD note

`scripts/gsd prompt programming-loop ...` is unavailable (`unknown GSD command: programming-loop`), so this slice uses the manual GSD/TDD fallback.

## Red target

- Freshchat embedded bundle must expose typed `file_upload` operations for `freshchat.files.upload` and `freshchat.images.upload`.
- `pm freshchat file upload` / `pm freshchat image upload` must be operation-backed and fail closed before credential resolution.
- Connectorgen should validate that Freshchat binary upload commands use typed operation metadata instead of a raw upload escape hatch.

Expected initial failure: Freshchat has no `operations.json`, and upload commands are merely excluded metadata entries with no typed operation IDs.

## Green target

Add operation metadata and command-surface wiring while preserving blocked-by-default upload behavior.

## Verification ledger

Red evidence:

```bash
gofmt -w internal/connectors/engine/bundle_test.go internal/cli/cli_test.go cmd/connectorgen/freshchat_api_surface_test.go
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatFileUploadOperations
```

Initial failure: Freshchat operations missing `freshchat.files.upload` / `freshchat.images.upload` because no `operations.json` existed.

Green focused gates:

```bash
go test ./internal/connectors/engine -run TestBundleLoadEmbeddedFreshchatFileUploadOperations
go test ./internal/cli -run TestFreshchatUploadCommandsBlockTypedOperationsBeforeCredentialResolution
go test ./cmd/connectorgen -run 'TestValidate_CLISurface|TestFreshchatAPISurfaceLedger|TestFreshchatBinaryUploadCommandsUseTypedOperations'
go run ./cmd/connectorgen validate internal/connectors/defs
```

Results: pass; connectorgen reported `547 connector(s) checked, 0 findings`.

Full gates pass:

```bash
cd website && pnpm run gen:website-data
gofmt -w cmd internal
go vet ./...
go test ./...
go build ./cmd/pm
make verify
go run ./cmd/connectorgen validate internal/connectors/defs
```

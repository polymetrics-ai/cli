# Test Plan

## Unit Tests

- Sync mode parser accepts all canonical modes and rejects invalid modes.
- Validation rejects missing cursor for incremental modes.
- Validation rejects missing primary key for deduped modes.
- Dedup ordering uses cursor, extracted timestamp, then raw ID.
- Delete/tombstone records are omitted when newest.

## App Tests

- `full_refresh_append` intentionally duplicates rows across runs.
- `full_refresh_overwrite` preserves prior final output when a refresh fails.
- `incremental_append` advances cursor only after success.
- `incremental_append_deduped` keeps one final row per primary key.
- `full_refresh_overwrite_deduped` removes records missing from the latest source snapshot.

## CLI Tests

- Help and generated skills list all five sync modes.
- GitHub manifest advertises pull request primary key and cursor defaults.

## Verification

- `go test ./...`
- `go vet ./...`
- `go build ./cmd/pm`
- `make smoke`
- `./pm docs generate --dir docs/cli`
- `./pm skills generate --dir docs/skills`


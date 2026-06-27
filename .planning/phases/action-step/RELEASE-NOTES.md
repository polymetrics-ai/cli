# RELEASE-NOTES — Action Step (Phase 1)

## New capability: `action` step in flow manifests

Flow manifests now support a fourth step kind: `action`. An action step is a reverse-ETL
write gated behind a flow-level approval token.

### Example manifest addition

```json
{
  "id": "send-emails",
  "kind": "action",
  "source_table": "lead_outreach",
  "destination_connector": "sendgrid",
  "destination_credential": "sg-prod",
  "action": "create",
  "mappings": {"to": "email", "subject": "subject_line", "body": "body"},
  "max_retries": 3,
  "batch_size": 50,
  "in": ["lead_outreach"],
  "out": []
}
```

### New CLI flags

- `pm flow plan <manifest>` — now includes `approval_token` in output when action steps present.
- `pm flow run <manifest> --token <tok>` — executes action steps with token.
- `pm flow run <manifest> --per-action` — prompts for individual token per action step.

### Safety features

- Idempotent writes (deterministic record ID, identity map).
- Deduplicated on email/domain/external-id before write.
- 429-aware exponential backoff with jitter.
- Dead-letter queue for permanently failing records.
- Schema-drift detection — pauses step on breaking schema change.
- Audit receipts in ledger for every action batch.

### Breaking changes

None. Existing `sync` and `query` step flows are unaffected.

### Known limitations

- DLQ replay (`pm flow dlq retry`) deferred to a future phase.
- Runtime-backed (Dragonfly queue) action batching deferred to Phase 3+.

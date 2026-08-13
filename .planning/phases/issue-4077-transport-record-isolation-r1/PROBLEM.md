# Exact-head diagnosis — #4077

**Baseline:** `c67f40a5ff67a131950f3123e70527027dca8493`

## Reproduction

An untracked test was added only in a disposable detached worktree at the exact baseline and removed
after capture. It ran:

```text
go test -count=1 -run 'TestReproduction' ./internal/synctransport
```

It failed as expected:

```text
json.RawMessage clone mutated source storage: "[\"source\":true}"
map[string]string clone mutated source storage: map[string]string{"owner":"clone"}
stage/destination mutated source json.RawMessage: "[\"source\":true}"
stage/destination mutated source map[string]string: map[string]string{"owner":"destination"}
```

## Causal separation

| Category | Evidence |
|---|---|
| Initiating trigger | Source record contains `json.RawMessage` or `map[string]string`. Those dynamic concrete types miss `cloneRecordValue`'s `[]byte` and `map[string]any` cases. |
| Masking condition | #4047 tested `[]byte` and `map[string]any`, which do match its switch; no test used named byte-slice or string-map values. |
| Visible symptom | A stage changes the raw JSON byte and a destination changes the string map after the source has emitted the record. The source-owned values observe both mutations. |
| Disconfirming evidence | The same exact-head test mutated clone `[]byte` and `map[string]any` values and the source stayed unchanged. The fault is the missing type cases, not generic orchestrator copying. |

## Scope conclusion

The reproduction is a closed core record-isolation failure. It needs no provider, credential,
warehouse-format, polling, rate/auth, or certification action.

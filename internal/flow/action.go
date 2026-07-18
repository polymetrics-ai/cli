package flow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/telemetry"
)

// ---------------------------------------------------------------------------
// ActionResult — returned by HTTPActionRunner.Execute
// ---------------------------------------------------------------------------

// ActionResult summarises a single action step execution.
type ActionResult struct {
	RecordsAttempted int
	RecordsSucceeded int
	RecordsFailed    int
	DLQPath          string
	ReceiptIDs       []string
}

// ---------------------------------------------------------------------------
// StepActionRunner — interface used by Engine for action dispatch
// ---------------------------------------------------------------------------

// StepActionRunner executes a single action step.
// Implemented by HTTPActionRunner; stubbed in tests via stubActionRunner.
type StepActionRunner interface {
	ExecuteStep(ctx context.Context, step FlowStep, records []map[string]any, token, runID string) (ActionResult, error)
}

// ---------------------------------------------------------------------------
// SchemaSnapshot — persisted field→type map
// ---------------------------------------------------------------------------

// SchemaSnapshot holds the persisted schema for a single action step.
type SchemaSnapshot struct {
	Flow   string            `json:"flow"`
	Step   string            `json:"step"`
	Fields map[string]string `json:"fields"`
}

// ---------------------------------------------------------------------------
// HTTPActionRunner — concrete implementation of all 7 safety features
// ---------------------------------------------------------------------------

// HTTPActionRunner sends records to an HTTP destination with all safety invariants.
// The DestURL endpoint must accept a JSON array POST at /write and return a JSON
// object with "accepted" and optionally "external_ids" fields.
type HTTPActionRunner struct {
	FlowName   string
	StepID     string
	StateDir   string
	DLQDir     string
	Ledger     LedgerAdapter
	MaxRetries int
	BatchSize  int

	// DestURL is the HTTP endpoint for writing records (used in tests via httptest.Server).
	DestURL    string
	HTTPClient *http.Client

	// Sleep is injectable for tests (set to instant sleep).
	Sleep func(ctx context.Context, d time.Duration) error

	// LiveSchema, when non-nil, is used instead of fetching from the connector.
	// Populated directly in tests; in production it would come from Catalog().
	LiveSchema map[string]string

	// mu guards identityMap and sentIDs.
	mu          sync.Mutex
	identityMap map[string]string // pm_id → external_id (loaded from disk)
	sentIDs     map[string]bool   // in-run guard
}

func (r *HTTPActionRunner) maxRetries() int {
	if r.MaxRetries > 0 {
		return r.MaxRetries
	}
	return 3
}

func (r *HTTPActionRunner) batchSize() int {
	if r.BatchSize > 0 {
		return r.BatchSize
	}
	return 100
}

func (r *HTTPActionRunner) sleep(ctx context.Context, d time.Duration) error {
	if r.Sleep != nil {
		return r.Sleep(ctx, d)
	}
	return ctxSleepAction(ctx, d)
}

func ctxSleepAction(ctx context.Context, d time.Duration) error {
	if d <= 0 {
		return ctx.Err()
	}
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}

func (r *HTTPActionRunner) client() *http.Client {
	if r.HTTPClient != nil {
		return r.HTTPClient
	}
	return &http.Client{Timeout: 30 * time.Second}
}

// ---------------------------------------------------------------------------
// Identity map (file-backed)
// ---------------------------------------------------------------------------

func (r *HTTPActionRunner) identityMapPath() string {
	return filepath.Join(r.StateDir, "identity_map.json")
}

func (r *HTTPActionRunner) loadIdentityMap() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.identityMap != nil {
		return nil
	}
	r.identityMap = map[string]string{}
	data, err := os.ReadFile(r.identityMapPath())
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(data, &r.identityMap)
}

func (r *HTTPActionRunner) saveIdentityMap() error {
	r.mu.Lock()
	defer r.mu.Unlock()
	if err := os.MkdirAll(filepath.Dir(r.identityMapPath()), 0o700); err != nil {
		return err
	}
	data, err := json.Marshal(r.identityMap)
	if err != nil {
		return err
	}
	return os.WriteFile(r.identityMapPath(), data, 0o600)
}

func (r *HTTPActionRunner) isMapped(pmID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, ok := r.identityMap[pmID]
	return ok
}

func (r *HTTPActionRunner) storeMapping(pmID, extID string) {
	r.mu.Lock()
	r.identityMap[pmID] = extID
	r.mu.Unlock()
}

// ---------------------------------------------------------------------------
// Schema-drift detection
// ---------------------------------------------------------------------------

func (r *HTTPActionRunner) schemaSnapPath() string {
	return filepath.Join(r.StateDir, fmt.Sprintf("schema_snap_%s_%s.json", r.FlowName, r.StepID))
}

func (r *HTTPActionRunner) checkSchemaDrift() error {
	if r.LiveSchema == nil {
		// No live schema provided; skip drift detection (first run).
		return nil
	}
	data, err := os.ReadFile(r.schemaSnapPath())
	if os.IsNotExist(err) {
		// First run: save snapshot and proceed.
		return r.saveSchemaSnapshot(r.LiveSchema)
	}
	if err != nil {
		return err
	}
	var snap SchemaSnapshot
	if err := json.Unmarshal(data, &snap); err != nil {
		return err
	}
	// Check for breaking changes: removed fields or type changes.
	for field, snapType := range snap.Fields {
		liveType, ok := r.LiveSchema[field]
		if !ok {
			return fmt.Errorf("%w: field %q was removed", ErrSchemaDrift, field)
		}
		if liveType != snapType {
			return fmt.Errorf("%w: field %q type changed %s→%s", ErrSchemaDrift, field, snapType, liveType)
		}
	}
	// Additive fields are non-breaking.
	return nil
}

func (r *HTTPActionRunner) saveSchemaSnapshot(fields map[string]string) error {
	snap := SchemaSnapshot{
		Flow:   r.FlowName,
		Step:   r.StepID,
		Fields: fields,
	}
	data, err := json.Marshal(snap)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(r.schemaSnapPath()), 0o700); err != nil {
		return err
	}
	return os.WriteFile(r.schemaSnapPath(), data, 0o600)
}

// ---------------------------------------------------------------------------
// Deduplication
// ---------------------------------------------------------------------------

var defaultDedupeKeys = []string{"email", "domain", "external_id"}

func deduplicateRecords(records []map[string]any) []map[string]any {
	seen := map[string]int{} // dedup key → index in out
	var out []map[string]any
	for _, rec := range records {
		key := dedupeKey(rec)
		if key == "" {
			out = append(out, rec)
			continue
		}
		if idx, ok := seen[key]; ok {
			out[idx] = rec // last-write-wins
		} else {
			seen[key] = len(out)
			out = append(out, rec)
		}
	}
	return out
}

func dedupeKey(rec map[string]any) string {
	var parts []string
	for _, k := range defaultDedupeKeys {
		if v, ok := rec[k]; ok && v != nil {
			parts = append(parts, fmt.Sprintf("%s=%v", k, v))
		}
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "|")
}

// ---------------------------------------------------------------------------
// Deterministic record ID
// ---------------------------------------------------------------------------

// deterministicRecordID returns a SHA-256 hex of the record's content, scoped
// to the flow + step to prevent cross-flow collisions.
func deterministicRecordID(flowName, stepID string, record map[string]any) string {
	// Sort keys for determinism.
	keys := make([]string, 0, len(record))
	for k := range record {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	h := sha256.New()
	fmt.Fprintf(h, "%s/%s:", flowName, stepID)
	for _, k := range keys {
		v, _ := json.Marshal(record[k])
		fmt.Fprintf(h, "%s=%s;", k, v)
	}
	return hex.EncodeToString(h.Sum(nil))
}

// ---------------------------------------------------------------------------
// DLQ write
// ---------------------------------------------------------------------------

type dlqEntry struct {
	PMID     string         `json:"pm_id"`
	Error    string         `json:"error"`
	Attempts int            `json:"attempts"`
	Record   map[string]any `json:"record"` // field names only, values redacted
}

func (r *HTTPActionRunner) writeDLQ(runID string, entries []dlqEntry) (string, error) {
	if len(entries) == 0 {
		return "", nil
	}
	dir := filepath.Join(r.DLQDir, r.FlowName, r.StepID)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	path := filepath.Join(dir, runID+".ndjson")
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return "", err
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	for _, e := range entries {
		if err := enc.Encode(e); err != nil {
			return "", err
		}
	}
	return path, nil
}

func redactRecord(rec map[string]any) map[string]any {
	out := make(map[string]any, len(rec))
	for k := range rec {
		out[k] = "[REDACTED]"
	}
	return out
}

// ---------------------------------------------------------------------------
// backoff helpers
// ---------------------------------------------------------------------------

const (
	backoffBase = 500 * time.Millisecond
	backoffMax  = 30 * time.Second
)

func backoffDuration(attempt int) time.Duration {
	d := backoffBase
	for i := 0; i < attempt; i++ {
		d *= 2
		if d > backoffMax {
			d = backoffMax
			break
		}
	}
	// ±25% jitter
	jitter := time.Duration(rand.Int63n(int64(d/2))) - d/4
	d += jitter
	if d < 0 {
		d = 0
	}
	return d
}

// ---------------------------------------------------------------------------
// HTTP write with retry
// ---------------------------------------------------------------------------

type writeResponse struct {
	Accepted    []string          `json:"accepted"`
	Failed      []string          `json:"failed"`
	ExternalIDs map[string]string `json:"external_ids"`
}

// sendBatch POSTs records to DestURL with retry on 429/5xx.
// Returns (response, retryAfter, statusCode, error).
func (r *HTTPActionRunner) sendBatch(ctx context.Context, records []map[string]any, metrics *telemetry.RunCounters) (writeResponse, error) {
	payload, err := json.Marshal(records)
	if err != nil {
		return writeResponse{}, fmt.Errorf("marshal batch: %w", err)
	}

	maxAttempts := r.maxRetries() + 1
	started := time.Now()
	responseBytes := 0
	defer func() {
		telemetry.RecordConnectorOperation(ctx, http.MethodPost, time.Since(started), responseBytes)
	}()
	var lastErr error
	for attempt := 0; attempt < maxAttempts; attempt++ {
		if attempt > 0 {
			metrics.RecordBatchRetried()
			metrics.Flush(ctx)
			telemetry.RecordAPIRetry(ctx, http.MethodPost)
			wait := backoffDuration(attempt - 1)
			if err := r.sleep(ctx, wait); err != nil {
				return writeResponse{}, err
			}
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, r.DestURL, bytes.NewReader(payload))
		if err != nil {
			return writeResponse{}, err
		}
		req.Header.Set("Content-Type", "application/json")

		telemetry.RecordAPICall(ctx, http.MethodPost, len(payload))
		resp, err := r.client().Do(req)
		if err != nil {
			lastErr = err
			continue
		}

		body, _ := io.ReadAll(io.LimitReader(resp.Body, 64<<10))
		resp.Body.Close()
		responseBytes += len(body)

		if resp.StatusCode == http.StatusTooManyRequests {
			// Honor Retry-After if present.
			if ra := resp.Header.Get("Retry-After"); ra != "" {
				if secs, err := strconv.Atoi(ra); err == nil && secs > 0 {
					wait := time.Duration(secs) * time.Second
					telemetry.RecordRateLimitWait(ctx, http.MethodPost, wait)
					if sleepErr := r.sleep(ctx, wait); sleepErr != nil {
						return writeResponse{}, sleepErr
					}
				}
			}
			lastErr = fmt.Errorf("http 429 rate limited")
			continue
		}

		if resp.StatusCode >= 500 {
			lastErr = fmt.Errorf("http %d server error", resp.StatusCode)
			continue
		}

		if resp.StatusCode >= 400 {
			// Permanent client error — do not retry.
			return writeResponse{}, fmt.Errorf("http %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
		}

		var wr writeResponse
		if err := json.Unmarshal(body, &wr); err != nil {
			// Non-JSON success — treat all as accepted.
			return writeResponse{Accepted: extractPMIDs(records)}, nil
		}
		return wr, nil
	}
	return writeResponse{}, fmt.Errorf("max retries exceeded: %w", lastErr)
}

func extractPMIDs(records []map[string]any) []string {
	ids := make([]string, 0, len(records))
	for _, r := range records {
		if id, ok := r["_pm_id"].(string); ok {
			ids = append(ids, id)
		}
	}
	return ids
}

// ---------------------------------------------------------------------------
// Execute — main entry point
// ---------------------------------------------------------------------------

// Execute sends records to the destination with all 7 safety invariants applied.
// runID is a caller-supplied unique string for DLQ file naming.
func (r *HTTPActionRunner) Execute(ctx context.Context, records []map[string]any, runID string) (ActionResult, error) {
	// 1. Schema-drift detection (before any write).
	if err := r.checkSchemaDrift(); err != nil {
		return ActionResult{}, err
	}

	// 2. Load identity map.
	if err := r.loadIdentityMap(); err != nil {
		return ActionResult{}, fmt.Errorf("load identity map: %w", err)
	}

	// Init in-run sent set.
	r.mu.Lock()
	if r.sentIDs == nil {
		r.sentIDs = map[string]bool{}
	}
	r.mu.Unlock()

	// 3. Deduplicate.
	records = deduplicateRecords(records)

	// 4. Stamp _pm_id + filter already-sent/already-mapped.
	var toSend []map[string]any
	for _, rec := range records {
		// Compute pm_id from a copy that excludes _pm_id to ensure determinism.
		bare := make(map[string]any, len(rec))
		for k, v := range rec {
			if k != "_pm_id" {
				bare[k] = v
			}
		}
		pmID := deterministicRecordID(r.FlowName, r.StepID, bare)
		// Clone the record and stamp _pm_id so we don't mutate the caller's slice.
		stamped := make(map[string]any, len(rec)+1)
		for k, v := range rec {
			stamped[k] = v
		}
		stamped["_pm_id"] = pmID
		rec = stamped

		r.mu.Lock()
		alreadySent := r.sentIDs[pmID]
		r.mu.Unlock()
		if alreadySent {
			continue
		}
		if r.isMapped(pmID) {
			continue
		}
		toSend = append(toSend, rec)
	}

	result := ActionResult{RecordsAttempted: len(records)}
	var dlqEntries []dlqEntry
	metrics := telemetry.NewRunCounters(ctx)
	if len(records) > 0 && len(toSend) == 0 {
		metrics.RecordBatchSkipped()
		metrics.Flush(ctx)
	}

	// 5. Send in batches.
	bsz := r.batchSize()
	for i := 0; i < len(toSend); i += bsz {
		end := i + bsz
		if end > len(toSend) {
			end = len(toSend)
		}
		batch := toSend[i:end]
		metrics.RecordBatchCreated()
		metrics.Flush(ctx)

		wr, err := r.sendBatch(ctx, batch, metrics)
		if err != nil {
			// Batch-level failure (permanent error or context cancelled).
			// Quarantine the whole batch.
			for _, rec := range batch {
				pmID, _ := rec["_pm_id"].(string)
				dlqEntries = append(dlqEntries, dlqEntry{
					PMID:     pmID,
					Error:    err.Error(),
					Attempts: r.maxRetries() + 1,
					Record:   redactRecord(rec),
				})
				result.RecordsFailed++
			}
			continue
		}
		metrics.RecordBatchFlushed()
		metrics.Flush(ctx)

		// Mark accepted records.
		acceptedSet := map[string]bool{}
		for _, id := range wr.Accepted {
			acceptedSet[id] = true
		}

		for _, rec := range batch {
			pmID, _ := rec["_pm_id"].(string)
			if wr.Failed != nil {
				failed := false
				for _, fid := range wr.Failed {
					if fid == pmID {
						failed = true
						break
					}
				}
				if failed {
					dlqEntries = append(dlqEntries, dlqEntry{
						PMID:     pmID,
						Error:    "destination rejected record",
						Attempts: 1,
						Record:   redactRecord(rec),
					})
					result.RecordsFailed++
					continue
				}
			}

			// Mark as sent in-run.
			r.mu.Lock()
			r.sentIDs[pmID] = true
			r.mu.Unlock()

			// Store external ID if returned.
			if extID, ok := wr.ExternalIDs[pmID]; ok {
				r.storeMapping(pmID, extID)
			} else {
				// Store the pm_id itself so re-runs know it was sent.
				r.storeMapping(pmID, pmID)
			}

			result.RecordsSucceeded++
		}
	}

	// 6. Persist identity map.
	if err := r.saveIdentityMap(); err != nil {
		return result, fmt.Errorf("save identity map: %w", err)
	}

	// 7. Write DLQ.
	if len(dlqEntries) > 0 {
		dlqPath, err := r.writeDLQ(runID, dlqEntries)
		if err != nil {
			return result, fmt.Errorf("write dlq: %w", err)
		}
		result.DLQPath = dlqPath
	}

	// 8. Ledger receipt.
	if r.Ledger != nil && result.RecordsSucceeded > 0 {
		_ = r.Ledger.Append(ctx, LedgerRecord{
			Mode:      "action",
			Operation: r.FlowName + "/" + r.StepID,
			Status:    "receipt",
		})
		result.ReceiptIDs = append(result.ReceiptIDs, runID)
	}

	return result, nil
}

// ExecuteStep implements StepActionRunner for use by Engine.
func (r *HTTPActionRunner) ExecuteStep(ctx context.Context, step FlowStep, records []map[string]any, token, runID string) (ActionResult, error) {
	// token is validated by the Engine before calling here.
	return r.Execute(ctx, records, runID)
}

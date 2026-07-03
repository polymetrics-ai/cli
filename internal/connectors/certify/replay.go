package certify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

// ReplayTransport serves recorded cassette entries (written by
// RecordingTransport.Flush) back to the caller deterministically, with zero
// network access. Requests are matched against the CURRENT stage's cassette
// entries strictly in recorded sequence order; an unmatched request (wrong
// stage, wrong method/path, or the stage's sequence already exhausted) is a
// hard failure, per certification design §E: "unmatched request in replay =>
// failure (catches drift)".
type ReplayTransport struct {
	dir string

	mu    sync.Mutex
	stage string
	// next is the 0-based index of the next cassette entry to be served for
	// the current stage.
	next int
}

// NewReplayTransport builds a ReplayTransport rooted at dir (the same
// directory RecordingTransport was given). dir must already exist; a
// missing cassette root is treated as a hard configuration error rather
// than "everything is unmatched" (which would be a confusing way to learn
// the flag/path was wrong).
func NewReplayTransport(dir string) (*ReplayTransport, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("certify: cassette directory %s: %w", dir, err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("certify: cassette path %s is not a directory", dir)
	}
	return &ReplayTransport{dir: dir}, nil
}

// SetStage selects the stage whose cassette entries subsequent RoundTrip
// calls should be matched against, resetting the per-stage sequence cursor
// to the start. Certify calls this once per stage, mirroring
// RecordingTransport's one-transport-per-stage usage.
func (rt *ReplayTransport) SetStage(stage string) {
	rt.mu.Lock()
	defer rt.mu.Unlock()
	rt.stage = stage
	rt.next = 0
}

// RoundTrip returns the next cassette entry for the current stage if (and
// only if) its recorded method+path match req; otherwise it returns an
// error naming what was expected vs. what arrived, so a stage failure can
// point directly at the drift.
func (rt *ReplayTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	rt.mu.Lock()
	stage := rt.stage
	idx := rt.next
	rt.mu.Unlock()

	if stage == "" {
		return nil, fmt.Errorf("certify: replay: no stage selected (call SetStage first)")
	}

	entries, err := rt.loadStage(stage)
	if err != nil {
		return nil, err
	}
	if idx >= len(entries) {
		return nil, fmt.Errorf("certify: replay: stage %q cassette sequence exhausted (have %d entries, request %d: %s %s)",
			stage, len(entries), idx+1, req.Method, req.URL.Path)
	}

	entry := entries[idx]
	if entry.Request.Method != req.Method || entry.Request.Path != req.URL.Path {
		return nil, fmt.Errorf("certify: replay: unmatched request in stage %q at sequence %d: got %s %s, cassette expects %s %s",
			stage, idx+1, req.Method, req.URL.Path, entry.Request.Method, entry.Request.Path)
	}

	rt.mu.Lock()
	rt.next++
	rt.mu.Unlock()

	header := make(http.Header, len(entry.Response.Header))
	for k, v := range entry.Response.Header {
		header[k] = append([]string(nil), v...)
	}

	body := io.NopCloser(bytes.NewReader([]byte(entry.Response.Body)))
	return &http.Response{
		StatusCode: entry.Response.Status,
		Status:     http.StatusText(entry.Response.Status),
		Header:     header,
		Body:       body,
		Request:    req,
	}, nil
}

// loadStage reads and parses every cassette entry for stage, in sequence
// order, from <rt.dir>/<stage>/NNNN.json. An unknown/missing stage
// directory is a hard error rather than an empty (always-unmatched)
// sequence, so the failure message is unambiguous.
func (rt *ReplayTransport) loadStage(stage string) ([]cassetteEntry, error) {
	stageDir := filepath.Join(rt.dir, stage)
	dirEntries, err := os.ReadDir(stageDir)
	if err != nil {
		return nil, fmt.Errorf("certify: replay: no cassette recorded for stage %q (%s): %w", stage, stageDir, err)
	}

	names := make([]string, 0, len(dirEntries))
	for _, de := range dirEntries {
		if de.IsDir() {
			continue
		}
		names = append(names, de.Name())
	}
	names = sortedCassetteSeqs(names)

	entries := make([]cassetteEntry, 0, len(names))
	for _, name := range names {
		raw, err := os.ReadFile(filepath.Join(stageDir, name))
		if err != nil {
			return nil, fmt.Errorf("certify: replay: read cassette %s: %w", name, err)
		}
		var entry cassetteEntry
		if err := json.Unmarshal(raw, &entry); err != nil {
			return nil, fmt.Errorf("certify: replay: parse cassette %s: %w", name, err)
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

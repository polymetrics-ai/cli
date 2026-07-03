// Package certify: ledger.go implements the write-ahead leak ledger
// (design docs/architecture/connector-certification-design.md §C):
// "before any live write, append {action, tag, connector, entity_hint,
// planned_at} to certify-ledger.jsonl; after verified cleanup append {tag,
// cleaned_at}". The ledger is append-only JSONL so a crash mid-run still
// leaves a durable, greppable trail of every tag certify ever planned to
// create.
package certify

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// ledgerFileName is the fixed on-disk name for the write-ahead ledger,
// living at the root of the ephemeral certify workdir (design §C
// "certify-ledger.jsonl").
const ledgerFileName = "certify-ledger.jsonl"

// LedgerEntry is one line of certify-ledger.jsonl. A "planned" entry has
// PlannedAt set and CleanedAt zero; a "cleaned" entry (recorded by
// RecordCleaned) carries only Tag and CleanedAt — LoadLedger folds the two
// together per-tag via LedgerEntries.StatusFor.
type LedgerEntry struct {
	Action     string    `json:"action,omitempty"`
	Tag        string    `json:"tag"`
	Connector  string    `json:"connector,omitempty"`
	EntityHint string    `json:"entity_hint,omitempty"`
	// PlannedAt/CleanedAt use "omitzero" (not "omitempty": time.Time is a
	// struct, so encoding/json's omitempty never treats it as empty) so a
	// planned-only entry never serializes a spurious zero-value
	// "cleaned_at" field a naive reader could mistake for an actual
	// timestamp (TestLedgerRecordPlannedWritesBeforeCreate).
	PlannedAt time.Time `json:"planned_at,omitzero"`
	CleanedAt time.Time `json:"cleaned_at,omitzero"`
}

// Ledger is a write-ahead JSONL ledger rooted at a certify ephemeral workdir
// (or, for --sweep, a batch/creds root). All writes are append-only and
// fsync'd via os.File.Sync so a process crash immediately after a write
// still leaves the entry durable on disk.
type Ledger struct {
	mu   sync.Mutex
	path string
}

// NewLedger opens (creating if necessary) the ledger file at
// <root>/certify-ledger.jsonl.
func NewLedger(root string) (*Ledger, error) {
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("certify: create ledger root %s: %w", root, err)
	}
	path := filepath.Join(root, ledgerFileName)
	// Touch the file so callers can always os.Stat/os.ReadFile it even
	// before the first entry is appended.
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, fmt.Errorf("certify: open ledger %s: %w", path, err)
	}
	_ = f.Close()
	return &Ledger{path: path}, nil
}

// Path returns the ledger's on-disk file path.
func (l *Ledger) Path() string { return l.path }

// RecordPlanned appends a planned-write entry BEFORE any live write is
// attempted (design §C "write-ahead"). entry.PlannedAt is set to time.Now()
// when the caller leaves it zero, so production call sites never need to
// remember to stamp it themselves; tests may pre-set it (e.g. to simulate
// an aged entry for the sweeper).
func (l *Ledger) RecordPlanned(entry LedgerEntry) error {
	if entry.Tag == "" {
		return fmt.Errorf("certify: ledger entry requires a Tag")
	}
	if entry.PlannedAt.IsZero() {
		entry.PlannedAt = time.Now().UTC()
	}
	return l.append(entry)
}

// RecordCleaned appends a {tag, cleaned_at} entry after a cleanup has been
// verified (design §C "after verified cleanup append {tag, cleaned_at}").
func (l *Ledger) RecordCleaned(tag string) error {
	if tag == "" {
		return fmt.Errorf("certify: RecordCleaned requires a tag")
	}
	return l.append(LedgerEntry{Tag: tag, CleanedAt: time.Now().UTC()})
}

func (l *Ledger) append(entry LedgerEntry) error {
	l.mu.Lock()
	defer l.mu.Unlock()

	f, err := os.OpenFile(l.path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("certify: open ledger %s: %w", l.path, err)
	}
	defer func() { _ = f.Close() }()

	raw, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("certify: marshal ledger entry: %w", err)
	}
	if _, err := f.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("certify: write ledger entry: %w", err)
	}
	// Fsync so a crash immediately after this call still leaves the entry
	// durable — this is the entire point of a *write-ahead* ledger.
	if err := f.Sync(); err != nil {
		return fmt.Errorf("certify: sync ledger: %w", err)
	}
	return nil
}

// CopyTo copies the ledger file to
// <reportDir>/certifications/ledger/<connector>.jsonl (design §C "Ledger
// copied into .polymetrics/certifications/ledger/ even on crash"). Safe to
// call even when the ledger has zero entries (an empty/touched file is
// still copied).
func (l *Ledger) CopyTo(reportDir, connector string) error {
	destDir := filepath.Join(reportDir, certificationsDirName, "ledger")
	if err := os.MkdirAll(destDir, 0o755); err != nil {
		return fmt.Errorf("certify: create ledger copy dir: %w", err)
	}
	src, err := os.Open(l.path)
	if err != nil {
		return fmt.Errorf("certify: open ledger for copy %s: %w", l.path, err)
	}
	defer func() { _ = src.Close() }()

	destPath := filepath.Join(destDir, connector+".jsonl")
	dst, err := os.OpenFile(destPath, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644)
	if err != nil {
		return fmt.Errorf("certify: create ledger copy %s: %w", destPath, err)
	}
	defer func() { _ = dst.Close() }()

	if _, err := io.Copy(dst, src); err != nil {
		return fmt.Errorf("certify: copy ledger to %s: %w", destPath, err)
	}
	return nil
}

// LedgerStatus is the folded planned/cleaned state for one tag.
type LedgerStatus struct {
	Tag        string
	Connector  string
	Action     string
	EntityHint string
	PlannedAt  time.Time
	Cleaned    bool
	CleanedAt  time.Time
}

// LedgerEntries is the fully-loaded, per-tag-folded view of a ledger file
// produced by LoadLedger.
type LedgerEntries struct {
	byTag map[string]*LedgerStatus
	order []string
}

// StatusFor returns the folded status for tag, if any entry mentions it.
func (e LedgerEntries) StatusFor(tag string) (LedgerStatus, bool) {
	s, ok := e.byTag[tag]
	if !ok {
		return LedgerStatus{}, false
	}
	return *s, true
}

// All returns every known tag's status, in first-seen order.
func (e LedgerEntries) All() []LedgerStatus {
	out := make([]LedgerStatus, 0, len(e.order))
	for _, tag := range e.order {
		out = append(out, *e.byTag[tag])
	}
	return out
}

// Uncleaned returns every tag with a planned_at but no cleaned_at — the
// sweeper's core query (design §C "ledger entries without cleaned_at").
func (e LedgerEntries) Uncleaned() []LedgerStatus {
	var out []LedgerStatus
	for _, tag := range e.order {
		s := e.byTag[tag]
		if !s.PlannedAt.IsZero() && !s.Cleaned {
			out = append(out, *s)
		}
	}
	return out
}

// LoadLedger reads <root>/certify-ledger.jsonl and folds planned/cleaned
// entries together by tag. A missing ledger file is treated as empty (a
// certify run that never attempted a write leaves no ledger at all).
func LoadLedger(root string) (LedgerEntries, error) {
	path := filepath.Join(root, ledgerFileName)
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return LedgerEntries{byTag: map[string]*LedgerStatus{}}, nil
		}
		return LedgerEntries{}, fmt.Errorf("certify: open ledger %s: %w", path, err)
	}
	defer func() { _ = f.Close() }()

	entries := LedgerEntries{byTag: map[string]*LedgerStatus{}}
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 64*1024), 4*1024*1024)
	for scanner.Scan() {
		line := scanner.Text()
		if line == "" {
			continue
		}
		var entry LedgerEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return LedgerEntries{}, fmt.Errorf("certify: parse ledger line %q: %w", line, err)
		}
		status, ok := entries.byTag[entry.Tag]
		if !ok {
			status = &LedgerStatus{Tag: entry.Tag}
			entries.byTag[entry.Tag] = status
			entries.order = append(entries.order, entry.Tag)
		}
		if !entry.PlannedAt.IsZero() {
			status.PlannedAt = entry.PlannedAt
			status.Connector = entry.Connector
			status.Action = entry.Action
			status.EntityHint = entry.EntityHint
		}
		if !entry.CleanedAt.IsZero() {
			status.Cleaned = true
			status.CleanedAt = entry.CleanedAt
		}
	}
	if err := scanner.Err(); err != nil {
		return LedgerEntries{}, fmt.Errorf("certify: scan ledger %s: %w", path, err)
	}
	return entries, nil
}

package decision

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/sensitive"
)

type Actor string

const (
	ActorHuman    Actor = "human"
	ActorShepherd Actor = "shepherd"
	ActorContract Actor = "contract"
)

type Record struct {
	ID          string    `json:"decision_id"`
	DeliveryID  string    `json:"delivery_id"`
	ExecutionID string    `json:"execution_id"`
	UnitID      string    `json:"unit_id"`
	QuestionID  string    `json:"question_id"`
	Question    string    `json:"question"`
	Answer      string    `json:"answer"`
	Actor       Actor     `json:"actor"`
	Basis       string    `json:"basis"`
	At          time.Time `json:"at"`
}

const maxLedgerBytes = 8 * 1024 * 1024

type Store struct {
	mu   sync.Mutex
	path string
	file *os.File
	seen map[string][]byte
}

func Open(directory string) (*Store, error) {
	if directory == "" {
		return nil, errors.New("decision directory is required")
	}
	if err := os.MkdirAll(directory, 0o700); err != nil {
		return nil, fmt.Errorf("create decision directory: %w", err)
	}
	if err := os.Chmod(directory, 0o700); err != nil {
		return nil, fmt.Errorf("secure decision directory: %w", err)
	}
	path := filepath.Join(directory, "decisions.jsonl")
	_, statErr := os.Lstat(path)
	created := os.IsNotExist(statErr)
	if statErr != nil && !created {
		return nil, fmt.Errorf("inspect decision ledger: %w", statErr)
	}
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open decision ledger: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return nil, err
	}
	if created {
		directoryFile, err := os.Open(directory)
		if err != nil {
			_ = file.Close()
			return nil, fmt.Errorf("open decision directory for sync: %w", err)
		}
		syncErr := directoryFile.Sync()
		closeErr := directoryFile.Close()
		if syncErr != nil || closeErr != nil {
			_ = file.Close()
			return nil, errors.Join(syncErr, closeErr)
		}
	}
	seen, err := recoverAndIndex(file)
	if err != nil {
		_ = file.Close()
		return nil, err
	}
	return &Store{path: path, file: file, seen: seen}, nil
}

func (s *Store) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.file.Close()
}

func (s *Store) Append(_ context.Context, record Record) error {
	if err := validate(record); err != nil {
		return err
	}
	raw, err := json.Marshal(record)
	if err != nil {
		return err
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if existing, ok := s.seen[record.ID]; ok {
		if bytes.Equal(existing, raw) {
			return nil
		}
		return errors.New("decision id collides with different immutable content")
	}
	if info, err := s.file.Stat(); err != nil || info.Size()+int64(len(raw)+1) > maxLedgerBytes {
		if err != nil {
			return err
		}
		return errors.New("decision ledger exceeds the bounded size")
	}
	if _, err := s.file.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("append decision: %w", err)
	}
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("sync decision: %w", err)
	}
	s.seen[record.ID] = append([]byte(nil), raw...)
	return nil
}

func (s *Store) Records() ([]Record, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err := s.file.Sync(); err != nil {
		return nil, err
	}
	return Read(filepath.Dir(s.path))
}

type Snapshot struct {
	DeliveryID string
	Revision   int64
	LedgerHash string
	Summary    string
}

func (s *Store) Snapshot(deliveryID string) (Snapshot, error) {
	if deliveryID == "" {
		return Snapshot{}, errors.New("delivery identity is required")
	}
	records, err := s.Records()
	if err != nil {
		return Snapshot{}, err
	}
	return SnapshotRecords(records, deliveryID)
}

func SnapshotRecords(records []Record, deliveryID string) (Snapshot, error) {
	if deliveryID == "" {
		return Snapshot{}, errors.New("delivery identity is required")
	}
	filtered := make([]Record, 0, len(records))
	for _, record := range records {
		if record.DeliveryID == deliveryID {
			filtered = append(filtered, record)
		}
	}
	sort.SliceStable(filtered, func(i, j int) bool {
		if filtered[i].At.Equal(filtered[j].At) {
			return filtered[i].ID < filtered[j].ID
		}
		return filtered[i].At.Before(filtered[j].At)
	})
	summary := Markdown(filtered)
	hash := sha256.Sum256([]byte(summary))
	return Snapshot{DeliveryID: deliveryID, Revision: int64(len(filtered)),
		LedgerHash: "sha256:" + hex.EncodeToString(hash[:]), Summary: summary}, nil
}

func recoverAndIndex(file *os.File) (map[string][]byte, error) {
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	raw, err := io.ReadAll(io.LimitReader(file, maxLedgerBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxLedgerBytes {
		return nil, errors.New("decision ledger exceeds the bounded size")
	}
	if len(raw) > 0 && raw[len(raw)-1] != '\n' {
		start := bytes.LastIndexByte(raw, '\n') + 1
		tail := raw[start:]
		var record Record
		if err := json.Unmarshal(tail, &record); err == nil && validate(record) == nil {
			if len(raw)+1 > maxLedgerBytes {
				return nil, errors.New("repaired decision ledger would exceed the bounded size")
			}
			if _, err := file.Seek(0, io.SeekEnd); err != nil {
				return nil, err
			}
			if _, err := file.Write([]byte{'\n'}); err != nil {
				return nil, fmt.Errorf("complete decision ledger tail: %w", err)
			}
			raw = append(raw, '\n')
		} else {
			if err := file.Truncate(int64(start)); err != nil {
				return nil, fmt.Errorf("truncate torn decision ledger tail: %w", err)
			}
			raw = raw[:start]
		}
		if err := file.Sync(); err != nil {
			return nil, fmt.Errorf("sync repaired decision ledger: %w", err)
		}
	}
	_, _, seen, err := scanLedger(raw)
	if err != nil {
		return nil, err
	}
	_, err = file.Seek(0, io.SeekEnd)
	return seen, err
}

func scanLedger(raw []byte) ([]byte, []Record, map[string][]byte, error) {
	seen := make(map[string][]byte)
	var records []Record
	var cleaned bytes.Buffer
	scanner := bufio.NewScanner(bytes.NewReader(raw))
	scanner.Buffer(make([]byte, 4096), 64*1024)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		var record Record
		if err := json.Unmarshal(line, &record); err != nil {
			return nil, nil, nil, fmt.Errorf("decode decision: %w", err)
		}
		if err := validate(record); err != nil {
			return nil, nil, nil, fmt.Errorf("invalid decision: %w", err)
		}
		canonical, err := json.Marshal(record)
		if err != nil {
			return nil, nil, nil, err
		}
		if existing, duplicate := seen[record.ID]; duplicate {
			if !bytes.Equal(existing, canonical) {
				return nil, nil, nil, errors.New("decision id collides with different immutable content")
			}
			continue
		}
		seen[record.ID] = append([]byte(nil), canonical...)
		records = append(records, record)
		cleaned.Write(line)
		cleaned.WriteByte('\n')
	}
	if err := scanner.Err(); err != nil {
		return nil, nil, nil, err
	}
	return cleaned.Bytes(), records, seen, nil
}

func Read(directory string) ([]Record, error) {
	file, err := os.Open(filepath.Join(directory, "decisions.jsonl"))
	if err != nil {
		return nil, err
	}
	defer func() { _ = file.Close() }()
	raw, err := io.ReadAll(io.LimitReader(file, maxLedgerBytes+1))
	if err != nil {
		return nil, err
	}
	if len(raw) > maxLedgerBytes || (len(raw) > 0 && raw[len(raw)-1] != '\n') {
		return nil, errors.New("decision ledger is oversized or has an unrecovered tail")
	}
	_, records, _, err := scanLedger(raw)
	return records, err
}

func Markdown(records []Record) string {
	sort.SliceStable(records, func(i, j int) bool { return records[i].At.Before(records[j].At) })
	var out strings.Builder
	out.WriteString("## Shepherd decisions\n\n")
	out.WriteString("| Time (UTC) | Unit | Decision | Answer | Actor | Basis |\n")
	out.WriteString("|---|---|---|---|---|---|\n")
	for _, record := range records {
		fmt.Fprintf(&out, "| %s | %s | %s | %s | %s | %s |\n",
			record.At.UTC().Format(time.RFC3339), escape(record.UnitID), escape(record.Question),
			escape(record.Answer), escape(string(record.Actor)), escape(record.Basis))
	}
	return out.String()
}

func validate(record Record) error {
	if record.ID == "" || record.DeliveryID == "" || record.ExecutionID == "" || record.UnitID == "" ||
		record.QuestionID == "" || record.Question == "" || record.Answer == "" || record.Basis == "" || record.At.IsZero() {
		return errors.New("complete decision provenance is required")
	}
	if record.Actor != ActorHuman && record.Actor != ActorShepherd && record.Actor != ActorContract {
		return errors.New("decision actor must be human, shepherd, or contract")
	}
	if err := sensitive.ValidatePublicIdentifier(record.UnitID); err != nil {
		return fmt.Errorf("decision unit identity: %w", err)
	}
	for _, value := range []string{record.Question, record.Answer, record.Basis} {
		if err := sensitive.ValidatePublicText(value); err != nil {
			return fmt.Errorf("decision public text: %w", err)
		}
	}
	for _, value := range []string{record.ID, record.DeliveryID, record.ExecutionID, record.UnitID, record.QuestionID, record.Question, record.Answer, record.Basis} {
		if len(value) > 512 || strings.ContainsAny(value, "\r\n\x00") {
			return errors.New("decision field is unsafe")
		}
	}
	return nil
}

func escape(value string) string { return strings.ReplaceAll(value, "|", "\\|") }

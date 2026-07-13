package decision

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
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

type Store struct {
	mu   sync.Mutex
	path string
	file *os.File
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
	file, err := os.OpenFile(path, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o600)
	if err != nil {
		return nil, fmt.Errorf("open decision ledger: %w", err)
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return nil, err
	}
	return &Store{path: path, file: file}, nil
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
	if _, err := s.file.Write(append(raw, '\n')); err != nil {
		return fmt.Errorf("append decision: %w", err)
	}
	if err := s.file.Sync(); err != nil {
		return fmt.Errorf("sync decision: %w", err)
	}
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

func Read(directory string) ([]Record, error) {
	file, err := os.Open(filepath.Join(directory, "decisions.jsonl"))
	if err != nil {
		return nil, err
	}
	defer file.Close()
	var records []Record
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 64*1024)
	for scanner.Scan() {
		var record Record
		if err := json.Unmarshal(scanner.Bytes(), &record); err != nil {
			return nil, fmt.Errorf("decode decision: %w", err)
		}
		if err := validate(record); err != nil {
			return nil, fmt.Errorf("invalid decision: %w", err)
		}
		records = append(records, record)
	}
	return records, scanner.Err()
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
	for _, value := range []string{record.ID, record.DeliveryID, record.ExecutionID, record.UnitID, record.QuestionID, record.Question, record.Answer, record.Basis} {
		if len(value) > 512 || strings.ContainsAny(value, "\r\n\x00") {
			return errors.New("decision field is unsafe")
		}
	}
	return nil
}

func escape(value string) string { return strings.ReplaceAll(value, "|", "\\|") }

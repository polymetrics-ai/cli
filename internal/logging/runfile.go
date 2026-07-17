package logging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"polymetrics.ai/internal/safety"
)

const defaultMaxLogFiles = 25
const retentionRecentWindow = 2 * time.Second

// RunFileOptions configures per-run JSONL routing.
type RunFileOptions struct {
	MaxFiles int
}

// RunFileHandler writes records to .polymetrics/logs/<run-id>.jsonl based on context.
type RunFileHandler struct {
	state  *runFileState
	attrs  []groupedAttrs
	groups []string
}

type runFileState struct {
	mu         sync.Mutex
	projectDir string
	maxFiles   int
	logsRoot   *os.Root
	files      map[string]*os.File
	leases     map[string]struct{}
	projectKey string
	closed     bool
}

// NewRunFileHandler returns a context-routed JSONL slog handler.
func NewRunFileHandler(projectDir string, opts RunFileOptions) *RunFileHandler {
	if projectDir == "" {
		projectDir = filepath.Join(".", ".polymetrics")
	}
	maxFiles := opts.MaxFiles
	if maxFiles <= 0 {
		maxFiles = defaultMaxLogFiles
	}
	return &RunFileHandler{state: &runFileState{projectDir: projectDir, maxFiles: maxFiles, files: map[string]*os.File{}, leases: map[string]struct{}{}}}
}

func (h *RunFileHandler) Enabled(context.Context, slog.Level) bool { return h != nil }

func (h *RunFileHandler) Handle(ctx context.Context, record slog.Record) error {
	if h == nil || h.state == nil {
		return nil
	}
	runID := RunIDFromContext(ctx)
	if !validRunID(runID) {
		return nil
	}
	fields := h.recordFields(record)
	line, err := json.Marshal(fields)
	if err != nil {
		return err
	}
	line = append(line, '\n')

	h.state.mu.Lock()
	defer h.state.mu.Unlock()
	file, err := h.state.fileForLocked(runID)
	if err != nil {
		return err
	}
	if _, err := file.Write(line); err != nil {
		return err
	}
	return nil
}

func (h *RunFileHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	clone := *h
	clone.attrs = append(cloneGroupedAttrs(h.attrs), groupedAttrs{groups: cloneStrings(h.groups), attrs: cloneAttrs(attrs)})
	clone.groups = cloneStrings(h.groups)
	return &clone
}

func (h *RunFileHandler) WithGroup(name string) slog.Handler {
	clone := *h
	clone.attrs = cloneGroupedAttrs(h.attrs)
	if name == "" {
		clone.groups = cloneStrings(h.groups)
	} else {
		clone.groups = append(cloneStrings(h.groups), name)
	}
	return &clone
}

// Close closes all open run log files.
func (h *RunFileHandler) Close() error {
	if h == nil || h.state == nil {
		return nil
	}
	return h.state.close()
}

func (h *RunFileHandler) recordFields(record slog.Record) map[string]any {
	fields := map[string]any{
		"time":  record.Time.UTC().Format(time.RFC3339Nano),
		"level": record.Level.String(),
		"msg":   record.Message,
	}
	entries := append(cloneGroupedAttrs(h.attrs), groupedAttrs{groups: cloneStrings(h.groups)})
	record.Attrs(func(attr slog.Attr) bool {
		last := len(entries) - 1
		if len(entries[last].attrs) < maxDynamicAttrs {
			entries[last].attrs = append(entries[last].attrs, attr)
		}
		return len(entries[last].attrs) < maxDynamicAttrs
	})
	for _, attr := range mergeGroupAttrs(wrapRunFileEntries(entries)) {
		addAttr(fields, nil, attr)
	}
	return fields
}

func (s *runFileState) fileForLocked(runID string) (*os.File, error) {
	if s.closed {
		return nil, fmt.Errorf("log handler closed")
	}
	if file := s.files[runID]; file != nil {
		return file, nil
	}
	if err := s.ensureLogsRootLocked(); err != nil {
		return nil, err
	}
	name := runID + ".jsonl"
	file, err := s.logsRoot.OpenFile(name, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o600)
	if err != nil {
		return nil, err
	}
	if err := file.Chmod(0o600); err != nil {
		_ = file.Close()
		return nil, err
	}
	s.files[runID] = file
	s.acquireLeaseLocked(runID)
	if err := s.pruneLocked(runID); err != nil {
		return nil, err
	}
	return file, nil
}

func (s *runFileState) ensureLogsRootLocked() error {
	if s.logsRoot != nil {
		return nil
	}
	absProjectDir, err := filepath.Abs(s.projectDir)
	if err == nil {
		s.projectKey = canonicalProjectPath(absProjectDir)
	}
	if err != nil {
		return err
	}
	parentDir := filepath.Dir(absProjectDir)
	projectBase := filepath.Base(absProjectDir)
	parentRoot, err := os.OpenRoot(parentDir)
	if err != nil {
		return err
	}
	defer parentRoot.Close()
	if err := ensureRootDir(parentRoot, projectBase, 0o700); err != nil {
		return err
	}
	projectRoot, err := parentRoot.OpenRoot(projectBase)
	if err != nil {
		return err
	}
	defer projectRoot.Close()
	if err := ensureRootDir(projectRoot, "logs", 0o700); err != nil {
		return err
	}
	logsRoot, err := projectRoot.OpenRoot("logs")
	if err != nil {
		return err
	}
	s.logsRoot = logsRoot
	return nil
}

func ensureRootDir(root *os.Root, name string, mode os.FileMode) error {
	info, err := root.Lstat(name)
	if errors.Is(err, os.ErrNotExist) {
		if err := root.Mkdir(name, mode); err != nil {
			return err
		}
		info, err = root.Lstat(name)
	}
	if err != nil {
		return err
	}
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%s must not be a symlink", name)
	}
	if !info.IsDir() {
		return fmt.Errorf("%s is not a directory", name)
	}
	if err := root.Chmod(name, mode); err != nil {
		return err
	}
	return nil
}

// pruneLocked deletes only inactive, handler-owned logs. It intentionally skips active,
// leased, recent, or uncertain files because portable stdlib-only cross-process
// file locking is not available here; the safe failure mode is extra retention.
func (s *runFileState) pruneLocked(currentRunID string) error {
	if s.maxFiles <= 0 || s.logsRoot == nil {
		return nil
	}
	entries, err := fs.ReadDir(s.logsRoot.FS(), ".")
	if err != nil {
		return err
	}
	type logEntry struct {
		name    string
		runID   string
		modTime time.Time
		active  bool
		recent  bool
	}
	logs := make([]logEntry, 0, len(entries))
	for _, entry := range entries {
		name := entry.Name()
		runID, ok := ownedRunLogName(name)
		if entry.IsDir() || !ok {
			continue
		}
		info, err := s.logsRoot.Stat(name)
		if err != nil || !info.Mode().IsRegular() {
			continue
		}
		active := s.files[runID] != nil || isRunLogActive(s.projectKey, runID)
		recent := time.Since(info.ModTime()) < retentionRecentWindow
		logs = append(logs, logEntry{name: name, runID: runID, modTime: info.ModTime(), active: active, recent: recent})
	}
	if len(logs) <= s.maxFiles {
		return nil
	}
	sort.Slice(logs, func(i, j int) bool {
		if logs[i].modTime.Equal(logs[j].modTime) {
			return logs[i].name < logs[j].name
		}
		return logs[i].modTime.Before(logs[j].modTime)
	})
	remove := len(logs) - s.maxFiles
	for _, entry := range logs {
		if remove <= 0 {
			break
		}
		if entry.runID == currentRunID || entry.active || entry.recent {
			continue
		}
		err := s.logsRoot.Remove(entry.name)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			continue
		}
		remove--
	}
	return nil
}

func (s *runFileState) close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.closed {
		return nil
	}
	s.closed = true
	var errs []error
	for runID, file := range s.files {
		if err := file.Close(); err != nil {
			errs = append(errs, fmt.Errorf("close %s: %w", runID, err))
			continue
		}
		s.releaseLeaseLocked(runID)
		delete(s.files, runID)
	}
	if s.logsRoot != nil {
		if err := s.logsRoot.Close(); err != nil {
			errs = append(errs, err)
		}
		s.logsRoot = nil
	}
	return errorsJoin(errs)
}

func wrapRunFileEntries(entries []groupedAttrs) []slog.Attr {
	out := make([]slog.Attr, 0)
	for _, entry := range entries {
		out = append(out, wrapGroups(entry.groups, entry.attrs)...)
	}
	return out
}

var runLogLeases = struct {
	mu       sync.Mutex
	projects map[string]map[string]int
}{projects: map[string]map[string]int{}}

func (s *runFileState) acquireLeaseLocked(runID string) {
	if s.projectKey == "" {
		s.projectKey = canonicalProjectPath(s.projectDir)
	}
	runLogLeases.mu.Lock()
	defer runLogLeases.mu.Unlock()
	runs := runLogLeases.projects[s.projectKey]
	if runs == nil {
		runs = map[string]int{}
		runLogLeases.projects[s.projectKey] = runs
	}
	runs[runID]++
	s.leases[runID] = struct{}{}
}

func (s *runFileState) releaseLeaseLocked(runID string) {
	if _, ok := s.leases[runID]; !ok {
		return
	}
	delete(s.leases, runID)
	runLogLeases.mu.Lock()
	defer runLogLeases.mu.Unlock()
	runs := runLogLeases.projects[s.projectKey]
	if runs == nil {
		return
	}
	if runs[runID] <= 1 {
		delete(runs, runID)
	} else {
		runs[runID]--
	}
	if len(runs) == 0 {
		delete(runLogLeases.projects, s.projectKey)
	}
}

func isRunLogActive(projectKey, runID string) bool {
	runLogLeases.mu.Lock()
	defer runLogLeases.mu.Unlock()
	return runLogLeases.projects[projectKey][runID] > 0
}

func canonicalProjectPath(path string) string {
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	if real, err := filepath.EvalSymlinks(abs); err == nil {
		return filepath.Clean(real)
	}
	return filepath.Clean(abs)
}

func ownedRunLogName(name string) (string, bool) {
	if !strings.HasSuffix(name, ".jsonl") {
		return "", false
	}
	runID := strings.TrimSuffix(name, ".jsonl")
	if !strings.HasPrefix(runID, "run_") || !validRunID(runID) {
		return "", false
	}
	return runID, true
}

func validRunID(runID string) bool {
	if runID == "" || len(runID) > 128 || runID == "." || runID == ".." || strings.Contains(runID, "..") {
		return false
	}
	if strings.ContainsAny(runID, `/\\`) {
		return false
	}
	for _, r := range runID {
		switch {
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= '0' && r <= '9':
		case r == '_' || r == '-' || r == '.':
		default:
			return false
		}
		if r < 0x20 || r == 0x7f || (r >= 0x80 && r <= 0x9f) || safety.IsDangerousUnicode(r) {
			return false
		}
	}
	return true
}

func addAttr(fields map[string]any, groups []string, attr slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if attr.Key == "" && attr.Value.Kind() != slog.KindGroup {
		return
	}
	target := fields
	for _, group := range groups {
		if group == "" {
			continue
		}
		next, _ := target[group].(map[string]any)
		if next == nil {
			next = map[string]any{}
			target[group] = next
		}
		target = next
	}
	if attr.Value.Kind() == slog.KindGroup {
		groupTarget := target
		if attr.Key != "" {
			existing, _ := target[attr.Key].(map[string]any)
			if existing == nil {
				existing = map[string]any{}
				target[attr.Key] = existing
			}
			groupTarget = existing
		}
		for _, child := range attr.Value.Group() {
			addAttr(groupTarget, nil, child)
		}
		return
	}
	target[attr.Key] = slogValueAny(attr.Value)
}

func slogValueAny(value slog.Value) any {
	value = value.Resolve()
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindBool:
		return value.Bool()
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindFloat64:
		return value.Float64()
	case slog.KindInt64:
		return value.Int64()
	case slog.KindTime:
		return value.Time().UTC().Format(time.RFC3339Nano)
	case slog.KindUint64:
		return value.Uint64()
	case slog.KindGroup:
		out := map[string]any{}
		for _, attr := range value.Group() {
			addAttr(out, nil, attr)
		}
		return out
	case slog.KindAny:
		return typeMarker(value.Any())
	default:
		return value.String()
	}
}

func errorsJoin(errs []error) error {
	return errors.Join(errs...)
}

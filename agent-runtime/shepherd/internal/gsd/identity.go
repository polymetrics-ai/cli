package gsd

import (
	"bufio"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var sessionIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type sessionIdentityFingerprint struct {
	ID      string
	Size    int64
	ModTime time.Time
	Info    os.FileInfo
}

type SessionIdentityBaseline struct {
	files map[string]sessionIdentityFingerprint
}

type SessionIdentityEvidence struct {
	SessionID   string
	Fingerprint string
	Model       string
	Thinking    string
}

type strictSessionHeader struct {
	Type          string `json:"type"`
	Version       int    `json:"version"`
	ID            string `json:"id"`
	Timestamp     string `json:"timestamp"`
	CWD           string `json:"cwd"`
	ParentSession string `json:"parentSession,omitempty"`
}

// LatestSessionID returns only the ID of the newest Pi session whose header is
// bound to the exact requested worktree. It never reads message or tool rows.
func LatestSessionID(root, workDir string) (string, error) {
	_, latestID, err := latestSession(root, workDir)
	return latestID, err
}

func latestSession(root, workDir string) (string, string, error) {
	return latestSessionSince(root, workDir, time.Time{})
}

func latestSessionSince(root, workDir string, notBefore time.Time) (string, string, error) {
	if err := validateSessionRoot(root); err != nil {
		return "", "", err
	}
	var latestPath, latestID string
	var latestAt time.Time
	files := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		files++
		if files > 10_000 {
			return errors.New("session identity file count exceeds its bound")
		}
		header, info, err := readStrictSessionHeader(path)
		if err != nil {
			return err
		}
		if header.CWD != workDir || header.ParentSession != "" {
			return nil
		}
		if !notBefore.IsZero() && info.ModTime().Before(notBefore) {
			return nil
		}
		if latestID == "" || info.ModTime().After(latestAt) {
			latestPath = path
			latestID = header.ID
			latestAt = info.ModTime()
		}
		return nil
	})
	if err != nil {
		return "", "", err
	}
	if latestID == "" {
		return "", "", errors.New("no session is bound to the requested worktree")
	}
	return latestPath, latestID, nil
}

func readStrictSessionHeader(path string) (strictSessionHeader, os.FileInfo, error) {
	var header strictSessionHeader
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return header, nil, errors.New("session identity evidence must be a regular non-symlink file")
	}
	file, err := os.Open(path)
	if err != nil {
		return header, nil, err
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 64*1024)
	if !scanner.Scan() {
		_ = file.Close()
		if scanner.Err() != nil {
			return header, nil, scanner.Err()
		}
		return header, nil, errors.New("session identity header is missing")
	}
	raw := append([]byte(nil), scanner.Bytes()...)
	after, statErr := file.Stat()
	pathAfter, pathErr := os.Lstat(path)
	closeErr := file.Close()
	if err := rejectDuplicateJSONFields(raw); err != nil || jsonStrictDecode(raw, &header) != nil {
		return header, nil, errors.New("session identity header is malformed, duplicate, unknown, or trailing")
	}
	if statErr != nil || pathErr != nil || !os.SameFile(info, after) || !os.SameFile(after, pathAfter) || info.Size() != after.Size() || !info.ModTime().Equal(after.ModTime()) {
		return header, nil, errors.New("session identity file changed while reading its header")
	}
	if closeErr != nil {
		return header, nil, closeErr
	}
	if header.Type != "session" || header.Version != 3 || !sessionIDPattern.MatchString(header.ID) || !filepath.IsAbs(header.CWD) {
		return header, nil, errors.New("session identity header fields are invalid")
	}
	if _, err := time.Parse(time.RFC3339Nano, header.Timestamp); err != nil {
		return header, nil, errors.New("session identity timestamp is invalid")
	}
	return header, after, nil
}

func validateSessionRoot(root string) error {
	if !filepath.IsAbs(root) {
		return errors.New("session identity root must be absolute")
	}
	info, err := os.Lstat(root)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.IsDir() || !runtimePathOwnedByCurrentUser(info) {
		return errors.New("session identity root must be an owned non-symlink directory")
	}
	return nil
}

// CaptureSessionIdentityBaseline records only strict top-level session file
// identities before a governed launch. No message or tool content is retained.
func CaptureSessionIdentityBaseline(root, workDir string) (SessionIdentityBaseline, error) {
	if err := validateSessionRoot(root); err != nil {
		return SessionIdentityBaseline{}, err
	}
	files := make(map[string]sessionIdentityFingerprint)
	totalFiles := 0
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		totalFiles++
		if totalFiles > 10_000 {
			return errors.New("session identity file count exceeds its bound")
		}
		header, info, err := readStrictSessionHeader(path)
		if err != nil {
			return err
		}
		if header.CWD == workDir && header.ParentSession == "" {
			files[path] = sessionIdentityFingerprint{ID: header.ID, Size: info.Size(), ModTime: info.ModTime(), Info: info}
		}
		return nil
	})
	if errors.Is(err, os.ErrNotExist) {
		return SessionIdentityBaseline{files: files}, nil
	}
	if err != nil {
		return SessionIdentityBaseline{}, err
	}
	return SessionIdentityBaseline{files: files}, nil
}

// ReadSessionIdentityForRun requires exactly one new or changed top-level
// session after the captured baseline and launch timestamp.
func ReadSessionIdentityForRun(root, workDir string, baseline SessionIdentityBaseline, started time.Time, expectedModel, expectedThinking string) (string, string, error) {
	evidence, err := ReadSessionIdentityEvidenceForRun(root, workDir, baseline, started, expectedModel, expectedThinking)
	return evidence.Model, evidence.Thinking, err
}

func ReadSessionIdentityEvidenceForRun(root, workDir string, baseline SessionIdentityBaseline, started time.Time, expectedModel, expectedThinking string) (SessionIdentityEvidence, error) {
	current, err := CaptureSessionIdentityBaseline(root, workDir)
	if err != nil {
		return SessionIdentityEvidence{}, err
	}
	var candidate string
	var candidateOffset int64
	for path, fingerprint := range current.files {
		before, existed := baseline.files[path]
		changed := !existed || before.ID != fingerprint.ID || before.Size != fingerprint.Size || !before.ModTime.Equal(fingerprint.ModTime)
		if !changed {
			continue
		}
		if existed && (!os.SameFile(before.Info, fingerprint.Info) || fingerprint.Size < before.Size) {
			return SessionIdentityEvidence{}, errors.New("top-level session file was replaced or truncated during the governed run")
		}
		if fingerprint.ModTime.Before(started) {
			return SessionIdentityEvidence{}, errors.New("changed top-level session evidence predates the current run")
		}
		if candidate != "" {
			return SessionIdentityEvidence{}, errors.New("multiple top-level sessions changed during one governed run")
		}
		candidate = path
		if existed {
			candidateOffset = before.Size
		}
	}
	if candidate == "" {
		return SessionIdentityEvidence{}, errors.New("current governed run did not create or update one top-level session")
	}
	model, thinking, fingerprint, err := readSessionIdentityDelta(candidate, candidateOffset, expectedModel, expectedThinking, current.files[candidate].Info)
	if err != nil {
		return SessionIdentityEvidence{}, err
	}
	return SessionIdentityEvidence{SessionID: current.files[candidate].ID, Fingerprint: fingerprint, Model: model, Thinking: thinking}, nil
}

// ValidateSessionSuccessfulCompletion requires the final assistant row in one
// exact durable session to record a successful stop.
func ValidateSessionSuccessfulCompletion(root, sessionID string) error {
	if err := validateSessionRoot(root); err != nil || !sessionIDPattern.MatchString(sessionID) {
		return errors.New("valid durable session completion identity is required")
	}
	var selectedPath string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		header, _, err := readStrictSessionHeader(path)
		if err != nil {
			return err
		}
		if header.ID != sessionID {
			return nil
		}
		if selectedPath != "" {
			return errors.New("duplicate durable session completion identity")
		}
		selectedPath = path
		return nil
	})
	if err != nil || selectedPath == "" {
		return errors.Join(errors.New("durable session completion is missing"), err)
	}
	file, info, err := openRuntimeFileNoFollow(selectedPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if info.Size() < 0 || info.Size() > 16*1024*1024 {
		return errors.New("durable session completion is oversized")
	}
	scanner := bufio.NewScanner(io.LimitReader(file, 16*1024*1024+1))
	scanner.Buffer(make([]byte, 4096), 8*1024*1024)
	lastAssistantStop := ""
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		var event struct {
			Type    string `json:"type"`
			Message struct {
				Role       string `json:"role"`
				StopReason string `json:"stopReason"`
			} `json:"message"`
		}
		if err := rejectDuplicateJSONFields(line); err != nil || json.Unmarshal(line, &event) != nil {
			return errors.New("durable session completion is malformed")
		}
		if event.Message.Role != "" && event.Type != "message" {
			return errors.New("assistant-shaped payload appears outside a durable message row")
		}
		if event.Type == "message" && event.Message.Role == "assistant" {
			lastAssistantStop = event.Message.StopReason
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if lastAssistantStop != "stop" {
		return errors.New("durable session final assistant response was not successful")
	}
	return nil
}

// ValidateSessionAssistantProof binds one strict terminal assistant proof to the
// exact durable top-level session selected for the governed run.
func ValidateSessionAssistantProof(root, sessionID, expectedHash string) error {
	if err := validateSessionRoot(root); err != nil || !sessionIDPattern.MatchString(sessionID) ||
		!strings.HasPrefix(expectedHash, "sha256:") || len(expectedHash) != len("sha256:")+64 {
		return errors.New("valid durable assistant proof identity is required")
	}
	var selectedPath string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		header, _, err := readStrictSessionHeader(path)
		if err != nil {
			return err
		}
		if header.ID != sessionID {
			return errors.New("unexpected additional session in validator invocation root")
		}
		if selectedPath != "" {
			return errors.New("duplicate durable validator session identity")
		}
		selectedPath = path
		return nil
	})
	if err != nil || selectedPath == "" {
		return errors.Join(errors.New("durable validator session proof is missing"), err)
	}
	file, info, err := openRuntimeFileNoFollow(selectedPath)
	if err != nil {
		return err
	}
	defer func() { _ = file.Close() }()
	if info.Size() < 0 || info.Size() > 16*1024*1024 {
		return errors.New("durable validator session proof is oversized")
	}
	scanner := bufio.NewScanner(io.LimitReader(file, 16*1024*1024+1))
	scanner.Buffer(make([]byte, 4096), 8*1024*1024)
	assistantRows, matchedRow := 0, 0
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		var event struct {
			Type    string `json:"type"`
			Message struct {
				Role       string `json:"role"`
				Content    any    `json:"content"`
				StopReason string `json:"stopReason"`
			} `json:"message"`
		}
		if err := rejectDuplicateJSONFields(line); err != nil || json.Unmarshal(line, &event) != nil {
			return errors.New("durable validator session proof is malformed")
		}
		if event.Message.Role != "" && event.Type != "message" {
			return errors.New("assistant-shaped payload appears outside a durable message row")
		}
		if event.Type != "message" || event.Message.Role != "assistant" {
			continue
		}
		assistantRows++
		text := strings.TrimSpace(sessionAssistantText(event.Message.Content))
		if text == "" {
			continue
		}
		digest := sha256.Sum256([]byte(text))
		observedHash := "sha256:" + hex.EncodeToString(digest[:])
		if observedHash == expectedHash {
			if event.Message.StopReason != "stop" || matchedRow != 0 {
				return errors.New("durable validator proof is not one successful terminal response")
			}
			matchedRow = assistantRows
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	if matchedRow == 0 || matchedRow != assistantRows {
		return errors.New("stream proof does not match the final durable assistant response")
	}
	return nil
}

func sessionAssistantText(value any) string {
	var text strings.Builder
	var collect func(any)
	collect = func(current any) {
		switch typed := current.(type) {
		case map[string]any:
			if value, ok := typed["text"].(string); ok {
				text.WriteString(value)
			}
			for key, nested := range typed {
				if key != "text" {
					collect(nested)
				}
			}
		case []any:
			for _, nested := range typed {
				collect(nested)
			}
		}
	}
	collect(value)
	return text.String()
}

// ValidateSessionHasNoToolUse proves that one bounded durable session contains no tool events or tool content.
func ValidateSessionHasNoToolUse(root, sessionID string) error {
	if err := validateSessionRoot(root); err != nil || !sessionIDPattern.MatchString(sessionID) {
		return errors.New("valid no-tool session identity is required")
	}
	var selectedPath string
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		header, _, err := readStrictSessionHeader(path)
		if err != nil {
			return err
		}
		if header.ID != sessionID {
			return errors.New("unexpected additional session in no-tool invocation root")
		}
		if selectedPath != "" {
			return errors.New("duplicate durable session identity")
		}
		selectedPath = path
		return nil
	})
	if err != nil {
		return err
	}
	if selectedPath == "" {
		return errors.New("durable no-tool session is missing")
	}
	file, before, err := openRuntimeFileNoFollow(selectedPath)
	if err != nil || before.Size() > 16*1024*1024 {
		return errors.New("open bounded durable no-tool session")
	}
	defer func() { _ = file.Close() }()
	scanner := bufio.NewScanner(io.LimitReader(file, 16*1024*1024+1))
	scanner.Buffer(make([]byte, 4096), 8*1024*1024)
	rows := 0
	for scanner.Scan() {
		rows++
		if rows > 100_000 {
			return errors.New("durable no-tool session exceeds its row bound")
		}
		var event struct {
			Type       string `json:"type"`
			Role       string `json:"role"`
			ToolName   string `json:"toolName"`
			ToolCallID string `json:"toolCallId"`
			Message    struct {
				Role       string `json:"role"`
				ToolName   string `json:"toolName"`
				ToolCallID string `json:"toolCallId"`
				Content    []struct {
					Type string `json:"type"`
				} `json:"content"`
			} `json:"message"`
		}
		line := append([]byte(nil), scanner.Bytes()...)
		var generic any
		if err := rejectDuplicateJSONFields(line); err != nil || json.Unmarshal(line, &generic) != nil ||
			sessionContainsToolEvidence(generic) || json.Unmarshal(line, &event) != nil {
			return errors.New("durable no-tool session is malformed, duplicate, or contains tool evidence")
		}
		if strings.Contains(strings.ToLower(event.Type), "tool") || strings.Contains(strings.ToLower(event.Role), "tool") ||
			strings.Contains(strings.ToLower(event.Message.Role), "tool") || event.ToolName != "" || event.ToolCallID != "" ||
			event.Message.ToolName != "" || event.Message.ToolCallID != "" {
			return errors.New("durable recovery session contains forbidden tool use")
		}
		for _, content := range event.Message.Content {
			if strings.Contains(strings.ToLower(content.Type), "tool") {
				return errors.New("durable recovery session contains forbidden tool content")
			}
		}
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	after, statErr := file.Stat()
	pathAfter, pathErr := os.Lstat(selectedPath)
	if statErr != nil || pathErr != nil || !os.SameFile(before, after) || !os.SameFile(after, pathAfter) ||
		before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return errors.New("durable no-tool session changed during validation")
	}
	return nil
}

func sessionContainsToolEvidence(value any) bool {
	switch typed := value.(type) {
	case map[string]any:
		for key, child := range typed {
			lowerKey := strings.ToLower(key)
			if strings.Contains(lowerKey, "tool") && sessionToolValuePresent(child) {
				return true
			}
			if lowerKey == "role" || lowerKey == "type" {
				if text, ok := child.(string); ok && strings.Contains(strings.ToLower(text), "tool") {
					return true
				}
			}
			if sessionContainsToolEvidence(child) {
				return true
			}
		}
	case []any:
		for _, child := range typed {
			if sessionContainsToolEvidence(child) {
				return true
			}
		}
	}
	return false
}

func sessionToolValuePresent(value any) bool {
	switch typed := value.(type) {
	case nil:
		return false
	case string:
		return strings.TrimSpace(typed) != ""
	case []any:
		return len(typed) != 0
	case map[string]any:
		return len(typed) != 0
	case bool:
		return typed
	case float64:
		return typed != 0
	default:
		return true
	}
}

func readSessionIdentityDelta(path string, offset int64, expectedModel, expectedThinking string, selected os.FileInfo) (string, string, string, error) {
	file, before, err := openRuntimeFileNoFollow(path)
	if err != nil || selected == nil || !os.SameFile(selected, before) {
		return "", "", "", errors.New("open selected current session identity delta")
	}
	defer func() { _ = file.Close() }()
	if offset < 0 || before.Size() < offset || before.Size()-offset > 16*1024*1024 {
		return "", "", "", errors.New("current session identity delta is invalid or oversized")
	}
	if _, err := file.Seek(offset, io.SeekStart); err != nil {
		return "", "", "", err
	}
	var model, thinking string
	var assistantObserved bool
	deltaDigest := sha256.New()
	scanner := bufio.NewScanner(io.LimitReader(file, 16*1024*1024+1))
	scanner.Buffer(make([]byte, 4096), 8*1024*1024)
	rows := 0
	for scanner.Scan() {
		rows++
		if rows > 100_000 {
			return "", "", "", errors.New("current session identity delta exceeds its row bound")
		}
		var event struct {
			Type          string `json:"type"`
			Provider      string `json:"provider"`
			ModelID       string `json:"modelId"`
			ThinkingLevel string `json:"thinkingLevel"`
			Message       struct {
				Role     string `json:"role"`
				Provider string `json:"provider"`
				Model    string `json:"model"`
			} `json:"message"`
		}
		line := append([]byte(nil), scanner.Bytes()...)
		_, _ = deltaDigest.Write(line)
		_, _ = deltaDigest.Write([]byte{'\n'})
		if err := rejectDuplicateJSONFields(line); err != nil || json.Unmarshal(line, &event) != nil {
			return "", "", "", errors.New("current session identity delta is malformed or duplicate")
		}
		if event.Message.Role != "" && event.Type != "message" {
			return "", "", "", errors.New("assistant-shaped payload appears outside a current session message row")
		}
		transitions := make([]string, 0, 2)
		if event.Type == "model_change" {
			if event.Provider == "" || event.ModelID == "" {
				return "", "", "", errors.New("current session contains a partial model transition")
			}
			transitions = append(transitions, event.Provider+"/"+event.ModelID)
		}
		if event.Type == "message" && event.Message.Role == "assistant" {
			if event.Message.Provider == "" || event.Message.Model == "" {
				return "", "", "", errors.New("current session contains a partial assistant identity")
			}
			assistantObserved = true
			transitions = append(transitions, event.Message.Provider+"/"+event.Message.Model)
		}
		for _, transitionModel := range transitions {
			if transitionModel != expectedModel {
				return "", "", "", fmt.Errorf("unexpected current-session model transition to %s", transitionModel)
			}
			model = transitionModel
		}
		if event.Type == "thinking_level_change" {
			if event.ThinkingLevel == "" {
				return "", "", "", errors.New("current session contains a partial thinking transition")
			}
			if event.ThinkingLevel != expectedThinking {
				return "", "", "", fmt.Errorf("unexpected current-session thinking transition to %s", event.ThinkingLevel)
			}
			thinking = event.ThinkingLevel
		}
	}
	if err := scanner.Err(); err != nil {
		return "", "", "", err
	}
	after, err := file.Stat()
	pathAfter, pathErr := os.Lstat(path)
	if err != nil || pathErr != nil || !os.SameFile(before, after) || !os.SameFile(after, pathAfter) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return "", "", "", errors.New("current session identity delta changed while reading")
	}
	if model == "" || thinking == "" || !assistantObserved {
		return "", "", "", errors.New("current session delta lacks exact assistant model and thinking evidence")
	}
	return model, thinking, "sha256:" + hex.EncodeToString(deltaDigest.Sum(nil)), nil
}

// ReadSessionIdentity projects only provider/model/thinking metadata from durable Pi sessions.
// Message content and tool payloads are intentionally not retained or returned.
func ReadSessionIdentity(root, workDir string) (string, string, error) {
	return ReadSessionIdentitySince(root, workDir, time.Time{})
}

// ReadSessionIdentitySince accepts only a top-level session touched during the
// current launched run. It prevents stale or delegated sessions from filling
// missing live identity events.
func ReadSessionIdentitySince(root, workDir string, notBefore time.Time) (string, string, error) {
	path, _, err := latestSessionSince(root, workDir, notBefore)
	if err != nil {
		return "", "", err
	}
	return readSessionIdentityPath(path)
}

func readSessionIdentityPath(path string) (string, string, error) {
	var model, thinking string
	info, err := os.Lstat(path)
	if err != nil || info.Mode()&os.ModeSymlink != 0 || !info.Mode().IsRegular() {
		return "", "", errors.New("session identity evidence must be a regular non-symlink file")
	}
	file, err := os.Open(path)
	if err != nil {
		return "", "", err
	}
	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 4096), 8*1024*1024)
	for scanner.Scan() {
		var event struct {
			Type          string `json:"type"`
			Provider      string `json:"provider"`
			ModelID       string `json:"modelId"`
			ThinkingLevel string `json:"thinkingLevel"`
			Message       struct {
				Role     string `json:"role"`
				Provider string `json:"provider"`
				Model    string `json:"model"`
			} `json:"message"`
		}
		line := append([]byte(nil), scanner.Bytes()...)
		if err := rejectDuplicateJSONFields(line); err != nil || json.Unmarshal(line, &event) != nil {
			_ = file.Close()
			return "", "", errors.New("session identity event is malformed or duplicate")
		}
		if event.Message.Role != "" && event.Type != "message" {
			_ = file.Close()
			return "", "", errors.New("assistant-shaped payload appears outside a session message row")
		}
		if event.Type == "model_change" && event.Provider != "" && event.ModelID != "" {
			model = event.Provider + "/" + event.ModelID
		}
		if event.Type == "thinking_level_change" && event.ThinkingLevel != "" {
			thinking = event.ThinkingLevel
		}
		if event.Type == "message" && event.Message.Role == "assistant" && event.Message.Provider != "" && event.Message.Model != "" {
			model = event.Message.Provider + "/" + event.Message.Model
		}
	}
	scanErr := scanner.Err()
	after, statErr := file.Stat()
	closeErr := file.Close()
	if scanErr != nil {
		return "", "", scanErr
	}
	if statErr != nil || !os.SameFile(info, after) || info.Size() != after.Size() || !info.ModTime().Equal(after.ModTime()) {
		return "", "", errors.New("session identity evidence changed while reading")
	}
	if closeErr != nil {
		return "", "", closeErr
	}
	if model == "" || thinking == "" {
		return "", "", errors.New("session identity metadata is incomplete")
	}
	return model, thinking, nil
}

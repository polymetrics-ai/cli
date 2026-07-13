package gsd

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

var sessionIDPattern = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// LatestSessionID returns only the ID of the newest Pi session whose header is
// bound to the exact requested worktree. It never reads message or tool rows.
func LatestSessionID(root, workDir string) (string, error) {
	_, latestID, err := latestSession(root, workDir)
	return latestID, err
}

func latestSession(root, workDir string) (string, string, error) {
	var latestPath, latestID string
	var latestAt time.Time
	err := filepath.WalkDir(root, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".jsonl") {
			return nil
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		scanner := bufio.NewScanner(file)
		scanner.Buffer(make([]byte, 4096), 64*1024)
		var header struct {
			Type string `json:"type"`
			ID   string `json:"id"`
			CWD  string `json:"cwd"`
		}
		if scanner.Scan() {
			_ = json.Unmarshal(scanner.Bytes(), &header)
		}
		scanErr := scanner.Err()
		closeErr := file.Close()
		if scanErr != nil {
			return scanErr
		}
		if closeErr != nil {
			return closeErr
		}
		if header.Type != "session" || header.CWD != workDir || !sessionIDPattern.MatchString(header.ID) {
			return nil
		}
		info, err := entry.Info()
		if err != nil {
			return err
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

// ReadSessionIdentity projects only provider/model/thinking metadata from durable Pi sessions.
// Message content and tool payloads are intentionally not retained or returned.
func ReadSessionIdentity(root, workDir string) (string, string, error) {
	path, _, err := latestSession(root, workDir)
	if err != nil {
		return "", "", err
	}
	var model, thinking string
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
		if json.Unmarshal(scanner.Bytes(), &event) != nil {
			continue
		}
		if event.Type == "model_change" && event.Provider != "" && event.ModelID != "" {
			model = event.Provider + "/" + event.ModelID
		}
		if event.Type == "thinking_level_change" && event.ThinkingLevel != "" {
			thinking = event.ThinkingLevel
		}
		if event.Message.Role == "assistant" && event.Message.Provider != "" && event.Message.Model != "" {
			model = event.Message.Provider + "/" + event.Message.Model
		}
	}
	scanErr := scanner.Err()
	closeErr := file.Close()
	if scanErr != nil {
		return "", "", scanErr
	}
	if closeErr != nil {
		return "", "", closeErr
	}
	if model == "" || thinking == "" {
		return "", "", errors.New("session identity metadata is incomplete")
	}
	return model, thinking, nil
}

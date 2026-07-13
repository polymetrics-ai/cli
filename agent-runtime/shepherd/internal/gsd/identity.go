package gsd

import (
	"bufio"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// ReadSessionIdentity projects only provider/model/thinking metadata from durable Pi sessions.
// Message content and tool payloads are intentionally not retained or returned.
func ReadSessionIdentity(root string) (string, string, error) {
	var model, thinking string
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
		scanner.Buffer(make([]byte, 4096), 1024*1024)
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
			return scanErr
		}
		return closeErr
	})
	if err != nil {
		return "", "", err
	}
	if model == "" || thinking == "" {
		return "", "", errors.New("session identity metadata is incomplete")
	}
	return model, thinking, nil
}

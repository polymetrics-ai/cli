package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os/exec"
	"strings"
)

const maxCommentBytes = 60 * 1024

type runner interface {
	Run(context.Context, []string, []byte) ([]byte, error)
}

type Target struct {
	Repository  string
	PullRequest int
	DeliveryID  string
}

type comment struct {
	ID   int64  `json:"id"`
	Body string `json:"body"`
	User struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	} `json:"user"`
}

func decodeComments(raw []byte) ([]comment, error) {
	var pages [][]comment
	if err := json.Unmarshal(raw, &pages); err == nil {
		var comments []comment
		for _, page := range pages {
			comments = append(comments, page...)
		}
		return comments, nil
	}
	var comments []comment
	if err := json.Unmarshal(raw, &comments); err != nil {
		return nil, errors.New("decode bounded PR comment listing")
	}
	return comments, nil
}

func ValidateTarget(target Target) error {
	parts := strings.Split(target.Repository, "/")
	if len(parts) != 2 || !safeSlug(parts[0]) || !safeSlug(parts[1]) || target.PullRequest <= 0 || !safeDeliveryID(target.DeliveryID) {
		return errors.New("valid repository, pull request, and delivery identity are required")
	}
	return nil
}

func safeSlug(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || strings.ContainsRune("._-", r) {
			continue
		}
		return false
	}
	return true
}

func safeDeliveryID(value string) bool {
	return strings.HasPrefix(value, "issue-") && safeSlug(value)
}

type cliRunner struct{}

func (cliRunner) Run(ctx context.Context, args []string, stdin []byte) ([]byte, error) {
	command := exec.CommandContext(ctx, "gh", args...)
	command.Stdin = bytes.NewReader(stdin)
	stdout := boundedBuffer{limit: 4 * 1024 * 1024}
	command.Stdout = &stdout
	if err := command.Run(); err != nil {
		return nil, errors.New("gh api command failed")
	}
	if stdout.exceeded {
		return nil, errors.New("GitHub response exceeded the bounded size")
	}
	return stdout.buffer.Bytes(), nil
}

type boundedBuffer struct {
	buffer   bytes.Buffer
	limit    int
	exceeded bool
}

func (b *boundedBuffer) Write(raw []byte) (int, error) {
	remaining := b.limit - b.buffer.Len()
	if remaining <= 0 {
		b.exceeded = true
		return len(raw), nil
	}
	if len(raw) > remaining {
		b.exceeded = true
		_, _ = b.buffer.Write(raw[:remaining])
		return len(raw), nil
	}
	return b.buffer.Write(raw)
}

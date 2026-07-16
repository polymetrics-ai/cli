package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/outbox"
)

// NewCLIOutboxDispatcher is the only production constructor for a write-capable
// GitHub transport. The raw comment port is created inside this boundary and is
// handed directly to a durable, authority-fenced outbox dispatcher.
func NewCLIOutboxDispatcher(store *outbox.Store, fence outbox.ExecutionFence) *outbox.Dispatcher {
	return outbox.NewGitHubDispatcher(store, cliCommentPort{runner: cliRunner{}}, fence)
}

type cliCommentPort struct {
	runner runner
}

func (p cliCommentPort) Actor(ctx context.Context) (string, error) {
	raw, err := p.runner.Run(ctx, []string{"api", "user", "--method", "GET"}, nil)
	if err != nil {
		return "", fmt.Errorf("read bounded GitHub writer identity: %w", err)
	}
	var actor struct {
		Login string `json:"login"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(raw, &actor); err != nil || !safeSlug(actor.Login) ||
		(actor.Type != "User" && actor.Type != "Bot") {
		return "", errors.New("GitHub writer identity is missing or unsafe")
	}
	return actor.Login, nil
}

func (p cliCommentPort) ListComments(ctx context.Context, target outbox.Target) ([]outbox.Comment, error) {
	if err := validateOutboxTarget(target); err != nil {
		return nil, err
	}
	endpoint := "repos/" + target.Repository + "/issues/" + strconv.Itoa(target.PullRequest) + "/comments"
	raw, err := p.runner.Run(ctx, []string{"api", endpoint, "--paginate", "--slurp", "--method", "GET"}, nil)
	if err != nil {
		return nil, fmt.Errorf("list bounded PR comments: %w", err)
	}
	comments, err := decodeComments(raw)
	if err != nil {
		return nil, err
	}
	result := make([]outbox.Comment, 0, len(comments))
	for _, comment := range comments {
		result = append(result, outbox.Comment{ID: comment.ID, Body: comment.Body,
			Author: comment.User.Login, AuthorType: comment.User.Type})
	}
	return result, nil
}

func (p cliCommentPort) CreateComment(ctx context.Context, target outbox.Target, body string) (outbox.Comment, error) {
	if err := validateOutboxTarget(target); err != nil {
		return outbox.Comment{}, err
	}
	payload, err := marshalCommentBody(body)
	if err != nil {
		return outbox.Comment{}, err
	}
	endpoint := "repos/" + target.Repository + "/issues/" + strconv.Itoa(target.PullRequest) + "/comments"
	raw, err := p.runner.Run(ctx, []string{"api", endpoint, "--method", "POST", "--input", "-"}, payload)
	if err != nil {
		return outbox.Comment{}, outbox.NewAmbiguousWriteError(err)
	}
	var created comment
	if err := json.Unmarshal(raw, &created); err != nil || created.ID <= 0 {
		return outbox.Comment{}, outbox.NewAmbiguousWriteError(errors.New("GitHub create response has no bounded comment identity"))
	}
	return outbox.Comment{ID: created.ID, Body: created.Body, Author: created.User.Login,
		AuthorType: created.User.Type}, nil
}

func (p cliCommentPort) UpdateComment(ctx context.Context, target outbox.Target, commentID int64, body string) (outbox.Comment, error) {
	if err := validateOutboxTarget(target); err != nil {
		return outbox.Comment{}, err
	}
	if commentID <= 0 {
		return outbox.Comment{}, errors.New("positive GitHub comment identity is required")
	}
	payload, err := marshalCommentBody(body)
	if err != nil {
		return outbox.Comment{}, err
	}
	endpoint := "repos/" + target.Repository + "/issues/comments/" + strconv.FormatInt(commentID, 10)
	raw, err := p.runner.Run(ctx, []string{"api", endpoint, "--method", "PATCH", "--input", "-"}, payload)
	if err != nil {
		return outbox.Comment{}, outbox.NewAmbiguousWriteError(err)
	}
	var updated comment
	if err := json.Unmarshal(raw, &updated); err != nil || updated.ID <= 0 {
		return outbox.Comment{}, outbox.NewAmbiguousWriteError(errors.New("GitHub update response has no bounded comment identity"))
	}
	return outbox.Comment{ID: updated.ID, Body: updated.Body, Author: updated.User.Login,
		AuthorType: updated.User.Type}, nil
}

func validateOutboxTarget(target outbox.Target) error {
	parts := strings.Split(target.Repository, "/")
	if len(parts) != 2 || !safeSlug(parts[0]) || !safeSlug(parts[1]) || target.Issue <= 0 || target.PullRequest <= 0 {
		return errors.New("complete bounded GitHub outbox target is required")
	}
	return nil
}

func marshalCommentBody(body string) ([]byte, error) {
	if strings.TrimSpace(body) == "" || len(body) > maxCommentBytes {
		return nil, errors.New("bounded GitHub comment body is required")
	}
	return json.Marshal(struct {
		Body string `json:"body"`
	}{Body: body})
}

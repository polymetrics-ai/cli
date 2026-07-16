package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

type QuestionRequest struct {
	RequestID         string
	Repository        string
	Issue             int
	PullRequest       int
	DeliveryID        string
	UnitID            string
	Generation        int64
	HeadSHA           string
	Evidence          string
	Options           []string
	RecommendedOption string
	SafeDefault       string
	ExpiresAt         time.Time
	Mention           string
	QuestionCommentID int64
}

const allowlistedDecisionActorID int64 = 6113982

type Reply struct {
	RequestID string
	Option    string
	Author    string
	AuthorID  int64
	CommentID int64
	CreatedAt time.Time
}

type replyCommentReader interface {
	ListComments(context.Context, QuestionRequest) ([]richComment, error)
}

type ReplyPoller struct {
	reader replyCommentReader
}

func newReplyPoller(reader replyCommentReader) *ReplyPoller { return &ReplyPoller{reader: reader} }

func NewCLIReplyPoller() *ReplyPoller {
	return newReplyPoller(cliReplyReader{runner: cliRunner{}})
}

type cliReplyReader struct {
	runner runner
}

func (r cliReplyReader) ListComments(ctx context.Context, request QuestionRequest) ([]richComment, error) {
	endpoint := "repos/" + request.Repository + "/issues/" + strconv.Itoa(request.PullRequest) + "/comments"
	raw, err := r.runner.Run(ctx, []string{"api", endpoint, "--paginate", "--slurp", "--method", "GET"}, nil)
	if err != nil {
		return nil, fmt.Errorf("list PR comments: %w", err)
	}
	return decodeRichComments(raw)
}

func (p *ReplyPoller) PollDecisionReplies(ctx context.Context, request QuestionRequest, allowlistedHuman string) ([]Reply, error) {
	if p == nil || p.reader == nil {
		return nil, errors.New("GitHub runner is required")
	}
	if err := validateQuestionRequest(request); err != nil {
		return nil, err
	}
	if err := ValidateTarget(Target{Repository: request.Repository, PullRequest: request.PullRequest, DeliveryID: request.DeliveryID}); err != nil {
		return nil, err
	}
	if !safeSlug(allowlistedHuman) {
		return nil, errors.New("allowlisted human login is required")
	}
	comments, err := p.reader.ListComments(ctx, request)
	if err != nil {
		return nil, err
	}
	var replies []Reply
	for _, comment := range comments {
		if comment.ID <= request.QuestionCommentID {
			continue
		}
		reply, ok := parseDecisionReply(comment, request, allowlistedHuman)
		if ok && reply.CreatedAt.Before(request.ExpiresAt) {
			replies = append(replies, reply)
		}
	}
	if len(replies) > 1 {
		return nil, errors.New("multiple valid decision replies are ambiguous")
	}
	return replies, nil
}

func parseDecisionReply(comment richComment, request QuestionRequest, allowlistedHuman string) (Reply, bool) {
	createdAt, createdErr := time.Parse(time.RFC3339Nano, comment.CreatedAt)
	updatedAt, updatedErr := time.Parse(time.RFC3339Nano, comment.UpdatedAt)
	if comment.User.Login != allowlistedHuman || comment.User.ID != allowlistedDecisionActorID ||
		comment.User.Type != "User" || createdErr != nil || updatedErr != nil || createdAt.IsZero() ||
		updatedAt.IsZero() || !createdAt.Equal(updatedAt) {
		return Reply{}, false
	}
	fields := strings.Fields(strings.TrimSpace(comment.Body))
	if len(fields) != 4 || fields[0] != "/shepherd" || fields[1] != "decide" || fields[2] != request.RequestID {
		return Reply{}, false
	}
	option := fields[3]
	for _, allowed := range request.Options {
		if option == allowed {
			return Reply{RequestID: request.RequestID, Option: option, Author: comment.User.Login,
				AuthorID: comment.User.ID, CommentID: comment.ID, CreatedAt: createdAt}, true
		}
	}
	return Reply{}, false
}

func validateQuestionRequest(request QuestionRequest) error {
	if request.RequestID == "" || request.Repository == "" || request.Issue <= 0 || request.PullRequest <= 0 ||
		request.DeliveryID == "" || request.UnitID == "" || request.Generation <= 0 || len(request.HeadSHA) != 40 ||
		strings.TrimSpace(request.Evidence) == "" || len(request.Options) == 0 || request.ExpiresAt.IsZero() ||
		request.Mention == "" || request.QuestionCommentID <= 0 {
		return errors.New("complete question request identity is required")
	}
	if !safeSlug(request.RequestID) || !safeDeliveryID(request.DeliveryID) || !safeSlug(strings.TrimPrefix(request.Mention, "@")) || strings.ContainsAny(request.UnitID, "\r\n\x00") {
		return errors.New("complete question request identity is required")
	}
	for _, option := range request.Options {
		if option == "" || strings.ContainsAny(option, "\r\n\x00 ") {
			return errors.New("question options must be bounded single tokens")
		}
	}
	return nil
}

type richComment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
		ID    int64  `json:"id"`
		Type  string `json:"type"`
	} `json:"user"`
}

func decodeRichComments(raw []byte) ([]richComment, error) {
	var pages [][]richComment
	if err := json.Unmarshal(raw, &pages); err == nil {
		var comments []richComment
		for _, page := range pages {
			comments = append(comments, page...)
		}
		return comments, nil
	}
	var comments []richComment
	if err := json.Unmarshal(raw, &comments); err != nil {
		return nil, errors.New("decode bounded PR comment listing")
	}
	return comments, nil
}

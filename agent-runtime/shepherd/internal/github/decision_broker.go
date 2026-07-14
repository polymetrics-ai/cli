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

const maxQuestionBodyBytes = 32 * 1024

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
}

type Reply struct {
	RequestID string
	Option    string
	Author    string
	CommentID int64
}

func (c *Client) SyncQuestionComment(ctx context.Context, request QuestionRequest) (int64, error) {
	if c == nil || c.runner == nil {
		return 0, errors.New("GitHub runner is required")
	}
	if err := validateQuestionRequest(request); err != nil {
		return 0, err
	}
	marker := "<!-- shepherd-question:" + request.DeliveryID + ":" + request.RequestID + " -->"
	body := marker + "\n\n" + renderQuestionBody(request) + "\n"
	if len(body) > maxQuestionBodyBytes {
		return 0, errors.New("question comment exceeds the bounded size")
	}
	target := Target{Repository: request.Repository, PullRequest: request.PullRequest, DeliveryID: request.DeliveryID}
	if err := ValidateTarget(target); err != nil {
		return 0, err
	}
	endpoint := "repos/" + request.Repository + "/issues/" + strconv.Itoa(request.PullRequest) + "/comments"
	raw, err := c.runner.Run(ctx, []string{"api", endpoint, "--paginate", "--slurp", "--method", "GET"}, nil)
	if err != nil {
		return 0, fmt.Errorf("list PR comments: %w", err)
	}
	comments, err := decodeRichComments(raw)
	if err != nil {
		return 0, err
	}
	var owned []int64
	for _, comment := range comments {
		if strings.HasPrefix(comment.Body, marker) {
			owned = append(owned, comment.ID)
		}
	}
	if len(owned) > 1 {
		return 0, errors.New("multiple Shepherd question comments claim this request")
	}
	method := "POST"
	commentID := int64(0)
	if len(owned) == 1 {
		method = "PATCH"
		commentID = owned[0]
		endpoint = "repos/" + request.Repository + "/issues/comments/" + strconv.FormatInt(commentID, 10)
	}
	payload, err := json.Marshal(struct {
		Body string `json:"body"`
	}{Body: body})
	if err != nil {
		return 0, err
	}
	written, err := c.runner.Run(ctx, []string{"api", endpoint, "--method", method, "--input", "-"}, payload)
	if err != nil {
		return 0, fmt.Errorf("publish Shepherd question: %w", err)
	}
	if commentID != 0 {
		return commentID, nil
	}
	var created richComment
	if err := json.Unmarshal(written, &created); err == nil && created.ID > 0 {
		return created.ID, nil
	}
	return 0, errors.New("question publisher did not return a comment id")
}

func (c *Client) PollDecisionReplies(ctx context.Context, request QuestionRequest, allowlistedHuman string) ([]Reply, error) {
	if c == nil || c.runner == nil {
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
	endpoint := "repos/" + request.Repository + "/issues/" + strconv.Itoa(request.PullRequest) + "/comments"
	raw, err := c.runner.Run(ctx, []string{"api", endpoint, "--paginate", "--slurp", "--method", "GET"}, nil)
	if err != nil {
		return nil, fmt.Errorf("list PR comments: %w", err)
	}
	comments, err := decodeRichComments(raw)
	if err != nil {
		return nil, err
	}
	var replies []Reply
	for _, comment := range comments {
		reply, ok := parseDecisionReply(comment, request, allowlistedHuman)
		if ok {
			replies = append(replies, reply)
		}
	}
	return replies, nil
}

func renderQuestionBody(request QuestionRequest) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "@%s Shepherd needs a decision.\n\n", strings.TrimPrefix(request.Mention, "@"))
	fmt.Fprintf(&builder, "- Request: `%s`\n", request.RequestID)
	fmt.Fprintf(&builder, "- Issue: #%d\n", request.Issue)
	fmt.Fprintf(&builder, "- PR: #%d\n", request.PullRequest)
	fmt.Fprintf(&builder, "- Unit: `%s`\n", request.UnitID)
	fmt.Fprintf(&builder, "- Generation: `%d`\n", request.Generation)
	fmt.Fprintf(&builder, "- Head: `%s`\n", request.HeadSHA)
	fmt.Fprintf(&builder, "- Evidence: %s\n", boundedLine(request.Evidence, 1000))
	if request.RecommendedOption != "" {
		fmt.Fprintf(&builder, "- Recommended option: `%s`\n", request.RecommendedOption)
	}
	if request.SafeDefault != "" {
		fmt.Fprintf(&builder, "- Safe default at expiry: `%s`\n", request.SafeDefault)
	}
	fmt.Fprintf(&builder, "- Expires: `%s`\n\n", request.ExpiresAt.UTC().Format(time.RFC3339))
	builder.WriteString("Options:\n")
	for _, option := range request.Options {
		fmt.Fprintf(&builder, "- `%s`\n", option)
	}
	fmt.Fprintf(&builder, "\nReply exactly with: `/shepherd decide %s <option>`\n", request.RequestID)
	return builder.String()
}

func parseDecisionReply(comment richComment, request QuestionRequest, allowlistedHuman string) (Reply, bool) {
	if comment.User.Login != allowlistedHuman || comment.User.Type == "Bot" || comment.CreatedAt != comment.UpdatedAt {
		return Reply{}, false
	}
	fields := strings.Fields(strings.TrimSpace(comment.Body))
	if len(fields) != 4 || fields[0] != "/shepherd" || fields[1] != "decide" || fields[2] != request.RequestID {
		return Reply{}, false
	}
	option := fields[3]
	for _, allowed := range request.Options {
		if option == allowed {
			return Reply{RequestID: request.RequestID, Option: option, Author: comment.User.Login, CommentID: comment.ID}, true
		}
	}
	return Reply{}, false
}

func validateQuestionRequest(request QuestionRequest) error {
	if request.RequestID == "" || request.Repository == "" || request.Issue <= 0 || request.PullRequest <= 0 ||
		request.DeliveryID == "" || request.UnitID == "" || request.Generation <= 0 || len(request.HeadSHA) != 40 ||
		strings.TrimSpace(request.Evidence) == "" || len(request.Options) == 0 || request.ExpiresAt.IsZero() || request.Mention == "" {
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

func boundedLine(value string, limit int) string {
	value = strings.TrimSpace(strings.Map(func(r rune) rune {
		if r < 0x20 || r == 0x7f {
			return -1
		}
		return r
	}, value))
	if len(value) > limit {
		return value[:limit]
	}
	return value
}

type richComment struct {
	ID        int64  `json:"id"`
	Body      string `json:"body"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
	User      struct {
		Login string `json:"login"`
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

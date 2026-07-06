package issueguard

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type IssueRef struct {
	Number  int
	Keyword string
	Closing bool
}

type Result struct {
	OK         bool
	Issues     []IssueRef
	Violations []string
}

var conventionalTitlePattern = regexp.MustCompile(`^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9._/-]+\))?!?: .+`)
var issueRefPattern = regexp.MustCompile(`(?i)\b(close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved|ref|refs|reference|references|relates to|related to|issue):?\s+#([1-9][0-9]*)\b`)

var closingKeywords = map[string]bool{
	"close":    true,
	"closes":   true,
	"closed":   true,
	"fix":      true,
	"fixes":    true,
	"fixed":    true,
	"resolve":  true,
	"resolves": true,
	"resolved": true,
}

func ValidatePRBody(title, body string) Result {
	var violations []string
	if !conventionalTitlePattern.MatchString(strings.TrimSpace(title)) {
		violations = append(violations, "PR title must use Conventional Commits, for example feat(github): add cli surface metadata")
	}

	issues := ExtractIssueRefs(body)
	if len(issues) == 0 {
		violations = append(violations, "PR body must reference an issue with Closes #123 for completed work or Refs #123 for stacked/incremental work")
	}

	return Result{
		OK:         len(violations) == 0,
		Issues:     issues,
		Violations: violations,
	}
}

func ExtractIssueRefs(text string) []IssueRef {
	matches := issueRefPattern.FindAllStringSubmatch(text, -1)
	if len(matches) == 0 {
		return nil
	}

	seen := map[int]IssueRef{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		number, err := strconv.Atoi(match[2])
		if err != nil {
			continue
		}
		keyword := strings.ToLower(match[1])
		ref := IssueRef{
			Number:  number,
			Keyword: keyword,
			Closing: closingKeywords[keyword],
		}
		if existing, ok := seen[number]; ok && existing.Closing {
			continue
		}
		seen[number] = ref
	}

	issues := make([]IssueRef, 0, len(seen))
	for _, ref := range seen {
		issues = append(issues, ref)
	}
	sort.Slice(issues, func(i, j int) bool {
		return issues[i].Number < issues[j].Number
	})
	return issues
}

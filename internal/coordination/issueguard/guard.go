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
	OK                   bool
	Issues               []IssueRef
	DeliveryRecord       bool
	ExplicitIssueWording bool
	Violations           []string
}

var conventionalTitlePattern = regexp.MustCompile(`^(build|chore|ci|docs|feat|fix|perf|refactor|revert|style|test)(\([a-z0-9][a-z0-9._-]*\))?!?: .+`)
var issueRefPattern = regexp.MustCompile(`(?i)\b(close|closes|closed|fix|fixes|fixed|resolve|resolves|resolved|ref|refs):?\s+(?:[a-z0-9_.-]+/[a-z0-9_.-]+)?#([1-9][0-9]*)\b`)
var issueTokenPattern = regexp.MustCompile(`(?:[A-Za-z0-9_.-]+/[A-Za-z0-9_.-]+)?#([1-9][0-9]*)\b`)
var issueWordNumberPattern = regexp.MustCompile(`(?i)\bissues?\s+(?:[a-z0-9_.-]+/[a-z0-9_.-]+)?#?([1-9][0-9]*)\b`)
var deliveryIssuePhrasePattern = regexp.MustCompile(`(?i)\b(?:deliver(?:s|ed|ing)?|implement(?:s|ed|ing)?|complete(?:s|d|ing)?|ship(?:s|ped|ping)?)\b(?:\.[0-9]|[^.\n\r]){0,80}\bissues?\b(?:\.[0-9]|[^.\n\r]){0,160}`)
var letteredDeliveryIssuePhrasePattern = regexp.MustCompile(`\b(?:[Dd]eliver(?:s|ed|ing)?|[Ii]mplement(?:s|ed|ing)?|[Cc]omplete(?:s|d|ing)?|[Ss]hip(?:s|ped|ping)?)\b(?:\.[0-9]|[^.\n\r]){0,80}\b[Ii]ssue\s+[A-Z]\b(?:\.[0-9]|[^.\n\r]){0,60}\b(?i:migration|slice|phase|workstream|delivery|plan|scope|implementation|contract|guard)\b`)
var parentIssuePattern = regexp.MustCompile(`(?i)\bparent\s+(?:issue\s+)?(?:[a-z0-9_.-]+/[a-z0-9_.-]+)?#([1-9][0-9]*)\b`)
var markdownH2Pattern = regexp.MustCompile(`(?m)^##\s+([A-Za-z][A-Za-z ]*)\s*$`)
var unvalidatedCheckpointHeadingPattern = regexp.MustCompile(`(?im)^##[ \t]+unvalidated cloud checkpoint[ \t]+—[ \t]+do not merge yet[ \t]*\r?$`)
var canonicalIssueSectionPattern = regexp.MustCompile(`(?im)^[ \t]{0,4}##[ \t]+canonical issue links preserved from the task record[ \t]*\r?$`)
var markdownIssueURLPattern = regexp.MustCompile(`(?im)^[ \t]*-[ \t]+https://git` + `hub\.com/[a-z0-9_.-]+/[a-z0-9_.-]+/issues/([1-9][0-9]*)[ \t]*\r?$`)
var completedTaskPattern = regexp.MustCompile(`(?i)\bunvalidated cloud checkpoint for the completed\b[^.\n\r]{0,160}\btask\b`)
var markdownH2StartPattern = regexp.MustCompile(`(?m)^[ \t]{0,4}##[ \t]+`)

const noMistakesDeliveryMarker = "Updates from [git push no-mistakes](https://" + "git" + "hub.com/kunchenguid/no-mistakes)"

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

func ValidatePR(title, body string) Result {
	var violations []string
	if !conventionalTitlePattern.MatchString(strings.TrimSpace(title)) {
		violations = append(violations, "PR title must use Conventional Commits, for example feat(connector): add cli surface metadata")
	}

	issues := ExtractIssueRefs(body)
	deliveryRecord := hasNoMistakesDeliveryRecord(body)
	explicitIssueWording := hasExplicitIssueWording(body)
	if len(issues) == 0 && !deliveryRecord && !explicitIssueWording {
		violations = append(violations, "PR body must reference an issue with Closes #123 for completed work, Refs #123 for stacked/incremental work, or explicit parent/delivery issue wording")
	}

	return Result{
		OK:                   len(violations) == 0,
		Issues:               issues,
		DeliveryRecord:       deliveryRecord,
		ExplicitIssueWording: explicitIssueWording,
		Violations:           violations,
	}
}

func ExtractIssueRefs(text string) []IssueRef {
	matches := issueRefPattern.FindAllStringSubmatch(text, -1)
	seen := map[int]IssueRef{}
	for _, match := range matches {
		if len(match) < 3 {
			continue
		}
		keyword := strings.ToLower(match[1])
		addIssueRef(seen, match[2], keyword)
	}

	for _, match := range parentIssuePattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		addIssueRef(seen, match[1], "parent")
	}

	addCheckpointIssueRefs(seen, text)

	for _, loc := range deliveryIssuePhrasePattern.FindAllStringIndex(text, -1) {
		if hasNegationPrefix(text, loc[0]) {
			continue
		}
		addDeliveryIssueTokens(seen, text[loc[0]:loc[1]], "issues")
	}

	if len(seen) == 0 {
		return nil
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

func addCheckpointIssueRefs(seen map[int]IssueRef, text string) {
	checkpointHeading := unvalidatedCheckpointHeadingPattern.FindStringIndex(text)
	if checkpointHeading == nil {
		return
	}

	heading := canonicalIssueSectionPattern.FindStringIndex(text)
	if heading == nil || heading[0] <= checkpointHeading[1] || !hasCompletedTaskWording(text[checkpointHeading[1]:heading[0]]) {
		return
	}
	sectionEnd := len(text)
	if nextHeading := markdownH2StartPattern.FindStringIndex(text[heading[1]:]); nextHeading != nil {
		sectionEnd = heading[1] + nextHeading[0]
	}
	section := text[heading[1]:sectionEnd]
	for _, match := range markdownIssueURLPattern.FindAllStringSubmatch(section, -1) {
		if len(match) < 2 {
			continue
		}
		addIssueRef(seen, match[1], "checkpoint")
	}
}

func hasCompletedTaskWording(text string) bool {
	return hasNonNegatedMatch(completedTaskPattern, text)
}

func hasNoMistakesDeliveryRecord(text string) bool {
	if !strings.Contains(text, noMistakesDeliveryMarker) {
		return false
	}

	headings := map[string]bool{}
	for _, match := range markdownH2Pattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		headings[strings.ToLower(strings.TrimSpace(match[1]))] = true
	}

	for _, required := range []string{"intent", "what changed", "testing", "pipeline"} {
		if !headings[required] {
			return false
		}
	}
	return true
}

func addDeliveryIssueTokens(seen map[int]IssueRef, text, keyword string) {
	addIssueTokens(seen, text, keyword)
	for _, match := range issueWordNumberPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		addIssueRef(seen, match[1], keyword)
	}
}

func addIssueTokens(seen map[int]IssueRef, text, keyword string) {
	for _, match := range issueTokenPattern.FindAllStringSubmatch(text, -1) {
		if len(match) < 2 {
			continue
		}
		addIssueRef(seen, match[1], keyword)
	}
}

func addIssueRef(seen map[int]IssueRef, rawNumber, keyword string) {
	number, err := strconv.Atoi(rawNumber)
	if err != nil {
		return
	}
	ref := IssueRef{
		Number:  number,
		Keyword: keyword,
		Closing: closingKeywords[keyword],
	}
	if existing, ok := seen[number]; ok && existing.Closing {
		return
	}
	seen[number] = ref
}

func hasExplicitIssueWording(text string) bool {
	return hasNonNegatedMatch(letteredDeliveryIssuePhrasePattern, text)
}

func hasNonNegatedMatch(re *regexp.Regexp, text string) bool {
	for _, loc := range re.FindAllStringIndex(text, -1) {
		if hasNegationPrefix(text, loc[0]) {
			continue
		}
		return true
	}
	return false
}

func hasNegationPrefix(text string, start int) bool {
	prefixStart := start - 16
	if prefixStart < 0 {
		prefixStart = 0
	}
	prefix := strings.ToLower(text[prefixStart:start])
	return strings.Contains(prefix, "do not ") || strings.Contains(prefix, "don't ") || strings.Contains(prefix, "not ")
}

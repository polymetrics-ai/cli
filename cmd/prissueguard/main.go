package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"polymetrics.ai/internal/coordination/issueguard"
)

var featureManagerBranchPattern = regexp.MustCompile(`^fm/([a-z0-9][a-z0-9._-]*)$`)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}

func run(args []string, stdout io.Writer, stderr io.Writer, getenv func(string) string) int {
	var title string
	var body string
	var bodyFile string
	var headRef string
	var planningRoot string

	flags := flag.NewFlagSet("prissueguard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&title, "title", getenv("PR_TITLE"), "pull request title")
	flags.StringVar(&body, "body", getenv("PR_BODY"), "pull request body")
	flags.StringVar(&bodyFile, "body-file", "", "file containing pull request body")
	flags.StringVar(&headRef, "head-ref", firstNonEmpty(getenv("HEAD_REF"), getenv("GITHUB_HEAD_REF")), "pull request head branch")
	flags.StringVar(&planningRoot, "planning-root", ".planning/phases", "GSD planning phases directory")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			writef(stderr, "read PR body file: %v\n", err)
			return 2
		}
		body = string(data)
	}

	result := issueguard.ValidatePR(title, body)
	result = applyPlanningIssueFallback(result, headRef, planningRoot)
	if result.OK {
		writef(stdout, "issueguard: ok (%d linked issue%s)\n", len(result.Issues), plural(len(result.Issues)))
		return 0
	}

	writeln(stderr, "issueguard: blocked")
	for _, violation := range result.Violations {
		writef(stderr, "- %s\n", violation)
	}
	return 1
}

func applyPlanningIssueFallback(result issueguard.Result, headRef, planningRoot string) issueguard.Result {
	if len(result.Issues) > 0 || !hasViolation(result.Violations, issueguard.LinkedIssueViolation) {
		return result
	}

	refs := planningIssueRefs(headRef, planningRoot)
	if len(refs) == 0 {
		return result
	}

	result.Issues = refs
	result.Violations = removeViolation(result.Violations, issueguard.LinkedIssueViolation)
	result.OK = len(result.Violations) == 0
	return result
}

func planningIssueRefs(headRef, planningRoot string) []issueguard.IssueRef {
	match := featureManagerBranchPattern.FindStringSubmatch(headRef)
	if len(match) != 2 || planningRoot == "" {
		return nil
	}

	planPath := filepath.Join(planningRoot, match[1], "PLAN.md")
	data, err := os.ReadFile(planPath)
	if err != nil {
		return nil
	}
	return issueguard.ExtractIssueRefs(string(data))
}

func removeViolation(violations []string, target string) []string {
	filtered := violations[:0]
	for _, violation := range violations {
		if violation != target {
			filtered = append(filtered, violation)
		}
	}
	return filtered
}

func hasViolation(violations []string, target string) bool {
	for _, violation := range violations {
		if violation == target {
			return true
		}
	}
	return false
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value != "" {
			return value
		}
	}
	return ""
}

func writef(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

func writeln(w io.Writer, text string) {
	_, _ = fmt.Fprintln(w, text)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

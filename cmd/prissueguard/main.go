package main

import (
	"flag"
	"fmt"
	"io"
	"os"

	"polymetrics.ai/internal/coordination/issueguard"
)

func main() {
	os.Exit(run(os.Args[1:], os.Stdout, os.Stderr, os.Getenv))
}

func run(args []string, stdout io.Writer, stderr io.Writer, getenv func(string) string) int {
	var title string
	var body string
	var bodyFile string

	flags := flag.NewFlagSet("prissueguard", flag.ContinueOnError)
	flags.SetOutput(stderr)
	flags.StringVar(&title, "title", getenv("PR_TITLE"), "pull request title")
	flags.StringVar(&body, "body", getenv("PR_BODY"), "pull request body")
	flags.StringVar(&bodyFile, "body-file", "", "file containing pull request body")
	if err := flags.Parse(args); err != nil {
		return 2
	}

	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			_, _ = fmt.Fprintf(stderr, "read PR body file: %v\n", err)
			return 2
		}
		body = string(data)
	}

	result := issueguard.ValidatePR(title, body)
	if result.OK {
		_, _ = fmt.Fprintf(stdout, "issueguard: ok (%d linked issue%s)\n", len(result.Issues), plural(len(result.Issues)))
		return 0
	}

	_, _ = fmt.Fprintln(stderr, "issueguard: blocked")
	for _, violation := range result.Violations {
		_, _ = fmt.Fprintf(stderr, "- %s\n", violation)
	}
	return 1
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

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
			if !writef(stderr, "read PR body file: %v\n", err) {
				return 2
			}
			return 2
		}
		body = string(data)
	}

	result := issueguard.ValidatePR(title, body)
	if result.OK {
		if !writef(stdout, "issueguard: ok (%d linked issue%s)\n", len(result.Issues), plural(len(result.Issues))) {
			return 1
		}
		return 0
	}

	if !writeln(stderr, "issueguard: blocked") {
		return 1
	}
	for _, violation := range result.Violations {
		if !writef(stderr, "- %s\n", violation) {
			return 1
		}
	}
	return 1
}

func writef(w io.Writer, format string, args ...any) bool {
	_, err := fmt.Fprintf(w, format, args...)
	return err == nil
}

func writeln(w io.Writer, args ...any) bool {
	_, err := fmt.Fprintln(w, args...)
	return err == nil
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

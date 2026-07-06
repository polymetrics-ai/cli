package main

import (
	"flag"
	"fmt"
	"os"

	"polymetrics.ai/internal/coordination/issueguard"
)

func main() {
	var title string
	var body string
	var bodyFile string
	flag.StringVar(&title, "title", os.Getenv("PR_TITLE"), "pull request title")
	flag.StringVar(&body, "body", os.Getenv("PR_BODY"), "pull request body")
	flag.StringVar(&bodyFile, "body-file", "", "file containing pull request body")
	flag.Parse()

	if bodyFile != "" {
		data, err := os.ReadFile(bodyFile)
		if err != nil {
			fmt.Fprintf(os.Stderr, "read PR body file: %v\n", err)
			os.Exit(2)
		}
		body = string(data)
	}

	result := issueguard.ValidatePRBody(title, body)
	if result.OK {
		fmt.Printf("issueguard: ok (%d linked issue%s)\n", len(result.Issues), plural(len(result.Issues)))
		return
	}

	fmt.Fprintln(os.Stderr, "issueguard: blocked")
	for _, violation := range result.Violations {
		fmt.Fprintf(os.Stderr, "- %s\n", violation)
	}
	os.Exit(1)
}

func plural(n int) string {
	if n == 1 {
		return ""
	}
	return "s"
}

func init() {
	flag.CommandLine.SetOutput(os.Stderr)
}

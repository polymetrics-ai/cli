package main

import "bytes"

type sourceProofTestResult struct {
	runs, passes int
	invalid      bool
}
type sourceProofResult struct {
	pkg   string
	tests map[string]sourceProofTestResult
	valid bool
}

func parseSourceProofResult(raw []byte) sourceProofResult {
	result := sourceProofResult{tests: map[string]sourceProofTestResult{}, valid: true}
	packagePass := 0
	for _, line := range bytes.Split(bytes.TrimSpace(raw), []byte{10}) {
		var event struct {
			Time        string
			Action      string
			Package     string
			Test        string
			Output      string
			Elapsed     float64
			FailedBuild string
		}
		if decodeStrictJSON(line, &event) != nil || event.Package == "" {
			result.valid = false
			break
		}
		if result.pkg == "" {
			result.pkg = event.Package
		}
		if result.pkg != event.Package {
			result.valid = false
			break
		}
		switch event.Action {
		case "start", "run", "pause", "cont", "output", "pass":
		default:
			result.valid = false
		}
		if event.Test != "" {
			entry := result.tests[event.Test]
			if event.Action == "run" {
				entry.runs++
				entry.invalid = entry.invalid || entry.passes != 0 || packagePass != 0
			}
			if event.Action == "pass" {
				entry.passes++
				entry.invalid = entry.invalid || entry.runs != 1 || packagePass != 0
			}
			result.tests[event.Test] = entry
		}
		if event.Test == "" && event.Action == "pass" {
			packagePass++
		}
	}
	result.valid = result.valid && packagePass == 1
	return result
}

// successfulSelection checks the exact observed result, without granting any
// assertion, source, lane or target authority to its caller.
func (result sourceProofResult) successfulSelection(pkg, selected string) bool {
	entry := result.tests[selected]
	return result.valid && result.pkg == pkg && entry.runs == 1 && entry.passes == 1 && !entry.invalid
}

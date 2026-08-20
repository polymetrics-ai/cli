package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
)

var evidenceSHARE = regexp.MustCompile(`^[0-9a-f]{40}$`)

type foundationEvidenceManifest struct {
	SchemaVersion      int                       `json:"schema_version"`
	Integration        string                    `json:"integration"`
	Checkpoint         string                    `json:"checkpoint"`
	CodeSHA            string                    `json:"code_sha"`
	ReviewedSHA        string                    `json:"reviewed_sha"`
	BaseSHA            string                    `json:"base_sha"`
	ComponentInputs    json.RawMessage           `json:"component_inputs"`
	InputMethod        string                    `json:"input_method"`
	FocusedChecks      []foundationEvidenceCheck `json:"focused_checks"`
	GateLedger         []foundationEvidenceGate  `json:"gate_ledger"`
	DeferredGates      []string                  `json:"deferred_gates"`
	CredentialHandling string                    `json:"credential_handling"`
	TemporaryMaterial  json.RawMessage           `json:"temporary_material"`
}

type foundationEvidenceCheck struct {
	Command   string   `json:"command"`
	Status    string   `json:"status"`
	Assertion string   `json:"assertion"`
	Tests     []string `json:"tests"`
	Modes     []string `json:"modes"`
}

type foundationEvidenceGate struct {
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	CommandIndexes []int    `json:"command_indexes,omitempty"`
	Tests          []string `json:"tests,omitempty"`
	Modes          []string `json:"modes,omitempty"`
	Reason         string   `json:"reason,omitempty"`
}

func runEvidenceGate(args []string, stdout, stderr io.Writer) int {
	if len(args) != 4 {
		logln(stderr, "usage: connectorgen evidence-gate <evidence-manifest.json> <TDD-LEDGER.md> <REVIEW.md>")
		return 2
	}
	manifest, err := os.ReadFile(args[1])
	if err != nil {
		logln(stderr, "connectorgen evidence-gate:", err)
		return 1
	}
	tdd, err := os.ReadFile(args[2])
	if err != nil {
		logln(stderr, "connectorgen evidence-gate:", err)
		return 1
	}
	review, err := os.ReadFile(args[3])
	if err != nil {
		logln(stderr, "connectorgen evidence-gate:", err)
		return 1
	}
	if err := validateFoundationEvidence(manifest, tdd, review); err != nil {
		logln(stderr, "connectorgen evidence-gate:", err)
		return 1
	}
	logln(stdout, "connectorgen evidence-gate: evidence claims match the command ledger and reviewed SHA")
	return 0
}

func validateFoundationEvidence(manifestRaw, tddRaw, reviewRaw []byte) error {
	decoder := json.NewDecoder(bytes.NewReader(manifestRaw))
	decoder.DisallowUnknownFields()
	var manifest foundationEvidenceManifest
	if err := decoder.Decode(&manifest); err != nil {
		return fmt.Errorf("parse evidence manifest: %w", err)
	}
	if manifest.SchemaVersion != 2 {
		return fmt.Errorf("evidence schema_version = %d, want 2", manifest.SchemaVersion)
	}
	if !evidenceSHARE.MatchString(manifest.CodeSHA) || !evidenceSHARE.MatchString(manifest.ReviewedSHA) {
		return fmt.Errorf("code_sha and reviewed_sha must be lowercase 40-character SHAs")
	}
	reviewedSHA := ""
	for _, line := range strings.Split(string(reviewRaw), "\n") {
		if value, found := strings.CutPrefix(strings.TrimSpace(line), "source_sha:"); found {
			reviewedSHA = strings.TrimSpace(value)
			break
		}
	}
	if reviewedSHA == "" || manifest.ReviewedSHA != reviewedSHA {
		return fmt.Errorf("reviewed_sha %q does not match review source_sha %q", manifest.ReviewedSHA, reviewedSHA)
	}
	for index, check := range manifest.FocusedChecks {
		if strings.TrimSpace(check.Command) == "" || strings.TrimSpace(check.Assertion) == "" {
			return fmt.Errorf("focused_checks[%d] has no command or assertion", index)
		}
		switch check.Status {
		case "passed", "cached":
		default:
			return fmt.Errorf("focused_checks[%d] has unsupported result status %q", index, check.Status)
		}
		if len(check.Tests) == 0 || len(check.Modes) == 0 {
			return fmt.Errorf("focused_checks[%d] must bind named tests and modes", index)
		}
	}
	tddStatuses := foundationTDDStatuses(tddRaw)
	seen := map[string]bool{}
	for _, gate := range manifest.GateLedger {
		if seen[gate.ID] || tddStatuses[gate.ID] == "" {
			return fmt.Errorf("gate %q is duplicate or absent from TDD ledger", gate.ID)
		}
		seen[gate.ID] = true
		if gate.Status != tddStatuses[gate.ID] {
			return fmt.Errorf("gate %s status %q disagrees with TDD ledger %q", gate.ID, gate.Status, tddStatuses[gate.ID])
		}
		switch gate.Status {
		case "passed":
			if len(gate.CommandIndexes) == 0 || len(gate.Tests) == 0 || len(gate.Modes) == 0 {
				return fmt.Errorf("passed gate %s has no named command, test, or mode evidence", gate.ID)
			}
			for _, commandIndex := range gate.CommandIndexes {
				if commandIndex < 0 || commandIndex >= len(manifest.FocusedChecks) {
					return fmt.Errorf("passed gate %s references command index %d outside the ledger", gate.ID, commandIndex)
				}
				if manifest.FocusedChecks[commandIndex].Status != "passed" {
					return fmt.Errorf("passed gate %s relies on %s command evidence", gate.ID, manifest.FocusedChecks[commandIndex].Status)
				}
			}
		case "deferred", "provisional":
			if strings.TrimSpace(gate.Reason) == "" {
				return fmt.Errorf("%s gate %s has no reason", gate.Status, gate.ID)
			}
		default:
			return fmt.Errorf("gate %s has unsupported status %q", gate.ID, gate.Status)
		}
	}
	if len(seen) != len(tddStatuses) {
		return fmt.Errorf("gate ledger covers %d of %d TDD gates", len(seen), len(tddStatuses))
	}
	if len(manifest.DeferredGates) == 0 {
		return fmt.Errorf("deferred gates must remain explicit")
	}
	return nil
}

func foundationTDDStatuses(raw []byte) map[string]string {
	statuses := map[string]string{}
	for _, line := range strings.Split(string(raw), "\n") {
		fields := strings.Split(line, "|")
		if len(fields) < 7 {
			continue
		}
		id := strings.TrimSpace(fields[1])
		if !strings.HasPrefix(id, "INT-") {
			continue
		}
		rawStatus := strings.ToLower(strings.TrimSpace(fields[len(fields)-2]))
		switch {
		case rawStatus == "green", rawStatus == "passed":
			statuses[id] = "passed"
		case strings.HasPrefix(rawStatus, "ready"), rawStatus == "provisional":
			statuses[id] = "provisional"
		default:
			statuses[id] = "deferred"
		}
	}
	return statuses
}

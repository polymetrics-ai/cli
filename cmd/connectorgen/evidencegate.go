package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	evidenceSHARE  = regexp.MustCompile(`^[0-9a-f]{40}$`)
	evidenceSHA256 = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

const foundationEvidenceSchemaVersion = 3

// foundationEvidenceManifest binds a final evidence-only commit to the exact
// implementation commit it certifies. CodeSHA is deliberately the parent of
// the evidence closure rather than the closure commit itself: a committed file
// cannot truthfully name the SHA that includes its own bytes.
type foundationEvidenceManifest struct {
	SchemaVersion      int                       `json:"schema_version"`
	Integration        string                    `json:"integration"`
	Checkpoint         string                    `json:"checkpoint"`
	CodeSHA            string                    `json:"code_sha"`
	ReviewedSHA        string                    `json:"reviewed_sha"`
	BaseSHA            string                    `json:"base_sha"`
	ComponentInputs    []foundationEvidenceInput `json:"component_inputs"`
	InputMethod        string                    `json:"input_method"`
	FocusedChecks      []foundationEvidenceCheck `json:"focused_checks"`
	GateLedger         []foundationEvidenceGate  `json:"gate_ledger"`
	DeferredGates      []string                  `json:"deferred_gates"`
	CredentialHandling string                    `json:"credential_handling"`
	TemporaryMaterial  json.RawMessage           `json:"temporary_material"`
	ArtifactClosure    foundationArtifactClosure `json:"artifact_closure"`
}

// foundationEvidenceInput is the minimal immutable component identity. The
// provenance-rich input manifest remains authoritative; this repeated form
// makes every component and its preserving merge visible at the evidence
// boundary without allowing an opaque component map.
type foundationEvidenceInput struct {
	Issue           int    `json:"issue"`
	SHA             string `json:"sha"`
	PreservingMerge string `json:"preserving_merge"`
}

type foundationInputManifest struct {
	Inputs []struct {
		Issue                  int    `json:"issue"`
		SHA                    string `json:"sha"`
		IntegrationMergeCommit string `json:"integration_merge_commit"`
	} `json:"inputs"`
}

type foundationArtifactClosure struct {
	ImplementationSHA  string                       `json:"implementation_sha"`
	SubjectFingerprint string                       `json:"subject_fingerprint"`
	Artifacts          []foundationEvidenceArtifact `json:"artifacts"`
}

type foundationEvidenceArtifact struct {
	Category string `json:"category"`
	Path     string `json:"path"`
	SHA256   string `json:"sha256"`
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

// foundationEvidenceValidationContext separates the graph and filesystem
// boundary from pure manifest validation. Unit tests use a closed fixture;
// the command path constructs this only from the current repository.
type foundationEvidenceValidationContext struct {
	HeadSHA       string
	ParentSHA     string
	InputManifest []byte
	ChangedPaths  []string
	WorktreeClean bool
	IsAncestor    func(ancestor, descendant string) bool
	ReadArtifact  func(path string) ([]byte, error)
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
	root, err := repoRoot()
	if err != nil {
		logln(stderr, "connectorgen evidence-gate:", err)
		return 1
	}
	context, err := foundationEvidenceContext(root)
	if err != nil {
		logln(stderr, "connectorgen evidence-gate:", err)
		return 1
	}
	if err := validateFoundationEvidenceWithContext(manifest, tdd, review, context); err != nil {
		logln(stderr, "connectorgen evidence-gate:", err)
		return 1
	}
	logln(stdout, "connectorgen evidence-gate: exact implementation graph, artifact closure, command ledger, and review identity agree")
	return 0
}

func foundationEvidenceContext(root string) (foundationEvidenceValidationContext, error) {
	head, err := gitOutput(root, "rev-parse", "HEAD")
	if err != nil {
		return foundationEvidenceValidationContext{}, fmt.Errorf("resolve evidence closure HEAD: %w", err)
	}
	parent, err := gitOutput(root, "rev-parse", "HEAD^")
	if err != nil {
		return foundationEvidenceValidationContext{}, fmt.Errorf("resolve evidence closure parent: %w", err)
	}
	status, err := gitOutput(root, "status", "--porcelain")
	if err != nil {
		return foundationEvidenceValidationContext{}, fmt.Errorf("inspect worktree: %w", err)
	}
	changed, err := gitLines(root, "diff", "--name-only", parent, head)
	if err != nil {
		return foundationEvidenceValidationContext{}, fmt.Errorf("inspect evidence closure diff: %w", err)
	}
	input, err := os.ReadFile(filepath.Join(root, "data", "cli-current-foundations-main-integration-r1", "input-manifest.json"))
	if err != nil {
		return foundationEvidenceValidationContext{}, fmt.Errorf("read foundation input manifest: %w", err)
	}
	return foundationEvidenceValidationContext{
		HeadSHA:       head,
		ParentSHA:     parent,
		InputManifest: input,
		ChangedPaths:  changed,
		WorktreeClean: status == "",
		IsAncestor: func(ancestor, descendant string) bool {
			command := exec.Command("git", "merge-base", "--is-ancestor", ancestor, descendant)
			command.Dir = root
			return command.Run() == nil
		},
		ReadArtifact: func(path string) ([]byte, error) {
			return os.ReadFile(filepath.Join(root, filepath.FromSlash(path)))
		},
	}, nil
}

func gitOutput(root string, args ...string) (string, error) {
	command := exec.Command("git", args...)
	command.Dir = root
	output, err := command.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(output)), nil
}

func gitLines(root string, args ...string) ([]string, error) {
	output, err := gitOutput(root, args...)
	if err != nil {
		return nil, err
	}
	if output == "" {
		return []string{}, nil
	}
	return strings.Split(output, "\n"), nil
}

// validateFoundationEvidence remains available to tests that validate a
// standalone record, but never certifies it: Foundation evidence is meaningful
// only with its clean repository graph and artifact reader.
func validateFoundationEvidence(manifestRaw, tddRaw, reviewRaw []byte) error {
	return fmt.Errorf("foundation evidence requires an exact repository graph")
}

func validateFoundationEvidenceWithContext(manifestRaw, tddRaw, reviewRaw []byte, context foundationEvidenceValidationContext) error {
	manifest, err := decodeFoundationEvidenceManifest(manifestRaw)
	if err != nil {
		return err
	}
	if err := validateFoundationEvidenceGraph(manifest, context); err != nil {
		return err
	}
	if err := validateFoundationArtifactClosure(manifest.ArtifactClosure, manifest.CodeSHA, context.ReadArtifact); err != nil {
		return err
	}
	if err := validateFoundationReviewIdentity(manifest, reviewRaw); err != nil {
		return err
	}
	return validateFoundationCommandLedger(manifest, tddRaw)
}

func decodeFoundationEvidenceManifest(raw []byte) (foundationEvidenceManifest, error) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var manifest foundationEvidenceManifest
	if err := decoder.Decode(&manifest); err != nil {
		return foundationEvidenceManifest{}, fmt.Errorf("parse evidence manifest: %w", err)
	}
	var trailing any
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return foundationEvidenceManifest{}, fmt.Errorf("parse evidence manifest: trailing JSON value")
		}
		return foundationEvidenceManifest{}, fmt.Errorf("parse evidence manifest: trailing JSON value")
	}
	if manifest.SchemaVersion != foundationEvidenceSchemaVersion {
		return foundationEvidenceManifest{}, fmt.Errorf("evidence schema_version = %d, want %d", manifest.SchemaVersion, foundationEvidenceSchemaVersion)
	}
	if !evidenceSHARE.MatchString(manifest.CodeSHA) || !evidenceSHARE.MatchString(manifest.ReviewedSHA) || !evidenceSHARE.MatchString(manifest.BaseSHA) {
		return foundationEvidenceManifest{}, fmt.Errorf("code_sha, reviewed_sha, and base_sha must be lowercase 40-character SHAs")
	}
	return manifest, nil
}

func validateFoundationEvidenceGraph(manifest foundationEvidenceManifest, context foundationEvidenceValidationContext) error {
	if !context.WorktreeClean {
		return fmt.Errorf("evidence closure requires a clean worktree")
	}
	if !evidenceSHARE.MatchString(context.HeadSHA) || !evidenceSHARE.MatchString(context.ParentSHA) {
		return fmt.Errorf("evidence closure repository graph is incomplete")
	}
	if manifest.CodeSHA == context.HeadSHA {
		return fmt.Errorf("code_sha self-reference is forbidden; evidence must name its implementation parent")
	}
	if manifest.CodeSHA != context.ParentSHA {
		return fmt.Errorf("implementation SHA %q does not match evidence closure parent %q", manifest.CodeSHA, context.ParentSHA)
	}
	if context.IsAncestor == nil || !context.IsAncestor(manifest.BaseSHA, manifest.CodeSHA) {
		return fmt.Errorf("diff base %q is not an ancestor of implementation SHA %q", manifest.BaseSHA, manifest.CodeSHA)
	}
	for _, path := range context.ChangedPaths {
		if !foundationEvidenceClosurePath(path) {
			return fmt.Errorf("evidence closure changed non-evidence path %q", path)
		}
	}
	return validateFoundationComponentInputs(manifest.ComponentInputs, context.InputManifest, manifest.CodeSHA, context.IsAncestor)
}

func foundationEvidenceClosurePath(path string) bool {
	return strings.HasPrefix(path, "data/cli-current-foundations-main-integration-r1/") ||
		strings.HasPrefix(path, ".planning/phases/cli-current-foundations-main-integration-r1/")
}

func validateFoundationComponentInputs(inputs []foundationEvidenceInput, inputManifestRaw []byte, implementationSHA string, isAncestor func(string, string) bool) error {
	if isAncestor == nil {
		return fmt.Errorf("component graph reader is required")
	}
	var inputManifest foundationInputManifest
	if err := json.Unmarshal(inputManifestRaw, &inputManifest); err != nil {
		return fmt.Errorf("parse foundation input manifest: %w", err)
	}
	expected := make(map[int]foundationEvidenceInput, len(inputManifest.Inputs))
	for _, input := range inputManifest.Inputs {
		if input.Issue == 0 || !evidenceSHARE.MatchString(input.SHA) || !evidenceSHARE.MatchString(input.IntegrationMergeCommit) {
			return fmt.Errorf("foundation input manifest has an incomplete component identity")
		}
		if _, duplicate := expected[input.Issue]; duplicate {
			return fmt.Errorf("foundation input manifest duplicates issue %d", input.Issue)
		}
		expected[input.Issue] = foundationEvidenceInput{Issue: input.Issue, SHA: input.SHA, PreservingMerge: input.IntegrationMergeCommit}
	}
	if len(expected) != 5 || len(inputs) != len(expected) {
		return fmt.Errorf("component input count = %d, want the five immutable Foundation components", len(inputs))
	}
	seen := make(map[int]bool, len(inputs))
	for _, input := range inputs {
		if seen[input.Issue] {
			return fmt.Errorf("component input issue %d is duplicated", input.Issue)
		}
		seen[input.Issue] = true
		expectedInput, found := expected[input.Issue]
		if !found || input.SHA != expectedInput.SHA {
			return fmt.Errorf("component head for issue %d does not match the immutable input manifest", input.Issue)
		}
		if input.PreservingMerge != expectedInput.PreservingMerge {
			return fmt.Errorf("preserving merge for issue %d does not match the immutable input manifest", input.Issue)
		}
		if !isAncestor(input.SHA, input.PreservingMerge) {
			return fmt.Errorf("component head for issue %d is not preserved by merge %q", input.Issue, input.PreservingMerge)
		}
		if !isAncestor(input.PreservingMerge, implementationSHA) {
			return fmt.Errorf("preserving merge for issue %d is not an ancestor of implementation SHA", input.Issue)
		}
	}
	return nil
}

var foundationArtifactCategories = map[string]bool{
	"source": true, "cli": true, "docs": true, "website": true, "skills": true,
	"ledger": true, "matrix": true, "candidate": true, "sweep": true, "foundation_evidence": true,
	"certification_subject": true,
}

func validateFoundationArtifactClosure(closure foundationArtifactClosure, implementationSHA string, readArtifact func(path string) ([]byte, error)) error {
	if closure.ImplementationSHA != implementationSHA {
		return fmt.Errorf("artifact closure implementation SHA does not match code_sha")
	}
	if !evidenceSHA256.MatchString(closure.SubjectFingerprint) {
		return fmt.Errorf("artifact closure subject_fingerprint must be a lowercase SHA-256 digest")
	}
	if readArtifact == nil {
		return fmt.Errorf("artifact closure reader is required")
	}
	subjectRaw, err := readArtifact(certificationSubjectArtifactPath)
	if err != nil {
		return fmt.Errorf("read current certification subject: %w", err)
	}
	var subjectArtifact currentCertificationSubjectArtifact
	if err := decodeStrictJSON(subjectRaw, &subjectArtifact); err != nil {
		return fmt.Errorf("parse current certification subject: %w", err)
	}
	if subjectArtifact.SchemaVersion != certificationSubjectSchemaVersion {
		return fmt.Errorf("current certification subject schema_version %d is unsupported", subjectArtifact.SchemaVersion)
	}
	if err := validateCertificationSubject(subjectArtifact.Subject); err != nil {
		return fmt.Errorf("current certification subject: %w", err)
	}
	if closure.SubjectFingerprint != subjectArtifact.Subject.Fingerprint {
		return fmt.Errorf("artifact closure subject_fingerprint does not match the current certification subject")
	}
	seenPaths := make(map[string]bool, len(closure.Artifacts))
	seenCategories := make(map[string]bool, len(closure.Artifacts))
	for _, artifact := range closure.Artifacts {
		if !foundationArtifactCategories[artifact.Category] {
			return fmt.Errorf("artifact closure category %q is unsupported", artifact.Category)
		}
		if seenCategories[artifact.Category] {
			return fmt.Errorf("artifact closure category %q is duplicated", artifact.Category)
		}
		seenCategories[artifact.Category] = true
		if !foundationEvidenceArtifactPath(artifact.Path) {
			return fmt.Errorf("artifact closure path %q is unsafe", artifact.Path)
		}
		if seenPaths[artifact.Path] {
			return fmt.Errorf("artifact closure path %q is duplicated", artifact.Path)
		}
		seenPaths[artifact.Path] = true
		if !evidenceSHA256.MatchString(artifact.SHA256) {
			return fmt.Errorf("artifact closure digest for %q is invalid", artifact.Path)
		}
		contents, err := readArtifact(artifact.Path)
		if err != nil {
			return fmt.Errorf("read closure artifact %q: %w", artifact.Path, err)
		}
		digest := sha256.Sum256(contents)
		if hex.EncodeToString(digest[:]) != artifact.SHA256 {
			return fmt.Errorf("artifact digest for %q does not match the closure", artifact.Path)
		}
	}
	for category := range foundationArtifactCategories {
		if !seenCategories[category] {
			return fmt.Errorf("required artifact category %q is absent from the closure", category)
		}
	}
	return nil
}

func foundationEvidenceArtifactPath(path string) bool {
	if strings.TrimSpace(path) == "" || filepath.IsAbs(path) || strings.Contains(path, "\\") {
		return false
	}
	clean := filepath.ToSlash(filepath.Clean(filepath.FromSlash(path)))
	return clean == path && !strings.HasPrefix(path, "../") && path != "."
}

func validateFoundationReviewIdentity(manifest foundationEvidenceManifest, reviewRaw []byte) error {
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
	return nil
}

func validateFoundationCommandLedger(manifest foundationEvidenceManifest, tddRaw []byte) error {
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

package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// enabledContractFinalLaneStatus is the machine-readable final-build outcome
// for one of the fixed source-lock lanes. It reports declaration completeness;
// it does not admit an operation or change a provider-backed source outcome.
type enabledContractFinalLaneStatus string

const (
	enabledContractFinalLanePresent  enabledContractFinalLaneStatus = "PRESENT"
	enabledContractFinalLanePartial  enabledContractFinalLaneStatus = "PARTIAL"
	enabledContractFinalLaneComplete enabledContractFinalLaneStatus = "COMPLETE"
	enabledContractFinalLaneMissing  enabledContractFinalLaneStatus = "MISSING"
)

type enabledContractFinalLaneResult struct {
	Name            string                               `json:"name"`
	Status          enabledContractFinalLaneStatus       `json:"status"`
	Reason          string                               `json:"reason"`
	Citations       []connectors.EnabledContractCitation `json:"citations"`
	UnmappedMapping int                                  `json:"unmapped_mapping"`
}

// checkEnabledConnectorContract is the authoring-side source-lock bridge for
// the optional enabled connector contract. The engine validates the closed
// JSON and the actual runtime binding; this check owns immutable source
// identity reconciliation and retained supplemental-document evidence.
func checkEnabledConnectorContract(fsys fs.FS, bundle engine.Bundle) []Finding {
	contract := bundle.EnabledContract
	if contract == nil {
		return nil
	}

	if err := contract.Validate(); err != nil {
		return []Finding{enabledContractFinding(bundle.Name, "enabled_connector_contract.json", err)}
	}

	findings := make([]Finding, 0)
	for _, result := range enabledContractFinalLaneResults(fsys, bundle) {
		if result.Status != enabledContractFinalLaneMissing {
			continue
		}
		findings = append(findings, enabledContractFinding(bundle.Name, "enabled_connector_contract.json", fmt.Errorf("lane %q: %s", result.Name, result.Reason)))
	}
	primary, err := loadEnabledContractSourceLock(fsys, bundle.Name, contract.SourceLock.Path)
	if err != nil {
		return append(findings, enabledContractFinding(bundle.Name, contract.SourceLock.Path, err))
	}
	if err := checkEnabledContractPrimarySourceEvidence(fsys, bundle.Name, primary, contract.SourceLock); err != nil {
		findings = append(findings, enabledContractFinding(bundle.Name, contract.SourceLock.Path, err))
	}
	if err := contract.ReconcileSourceOperations(enabledContractSourceOperations(primary)); err != nil {
		findings = append(findings, enabledContractFinding(bundle.Name, contract.SourceLock.Path, err))
	}

	for _, supplemental := range contract.SupplementalSourceLocks {
		lock, err := loadEnabledContractSourceLock(fsys, bundle.Name, supplemental.Path)
		if err != nil {
			findings = append(findings, enabledContractFinding(bundle.Name, supplemental.Path, err))
			continue
		}
		if err := checkEnabledContractRetainedDocuments(fsys, bundle.Name, lock); err != nil {
			findings = append(findings, enabledContractFinding(bundle.Name, supplemental.Path, err))
			continue
		}
		if err := contract.ReconcileSupplementalSourceOperations(supplemental.Path, enabledContractSourceOperations(lock)); err != nil {
			findings = append(findings, enabledContractFinding(bundle.Name, supplemental.Path, err))
		}
	}
	return findings
}

func checkEnabledContractPrimarySourceEvidence(fsys fs.FS, connector string, lock sourceImportLock, evidence connectors.EnabledContractSourceLock) error {
	if lock.SchemaVersion < 3 {
		if lock.Rest.SHA256 != evidence.SHA256 || lock.Rest.Bytes != evidence.Bytes {
			return fmt.Errorf("primary source lock identity does not match enabled_connector_contract.json")
		}
		return nil
	}
	if err := checkEnabledContractRetainedDocuments(fsys, connector, lock); err != nil {
		return err
	}
	for _, document := range lock.Rest.SourceDocuments {
		if document.isUnavailable() || document.isSourceReference() {
			continue
		}
		artifact := document.Artifact
		if artifact.SHA256 == evidence.SHA256 && artifact.Bytes == evidence.Bytes {
			return nil
		}
	}
	return fmt.Errorf("primary source lock identity does not match enabled_connector_contract.json")
}

// enabledContractFinalLaneResults projects the already-validated contract into
// final-build outcomes. Every exact lane remains visible: a zero-source lane
// is PRESENT, partial source accounting is PARTIAL, all-accounted source
// coverage is COMPLETE, and an unavailable declared artifact is MISSING.
// Contract validation owns the fixed lane vocabulary, cited reasons, and
// source coverage arithmetic; this view deliberately adds no new policy.
func enabledContractFinalLaneResults(fsys fs.FS, bundle engine.Bundle) []enabledContractFinalLaneResult {
	contract := bundle.EnabledContract
	if contract == nil {
		return nil
	}
	lanes := append([]connectors.EnabledConnectorLane(nil), contract.Lanes...)
	sort.Slice(lanes, func(left, right int) bool { return lanes[left].Name < lanes[right].Name })
	results := make([]enabledContractFinalLaneResult, 0, len(lanes))
	for _, lane := range lanes {
		result := enabledContractFinalLaneResult{
			Name:            lane.Name,
			Reason:          lane.Reason,
			Citations:       append([]connectors.EnabledContractCitation(nil), lane.Citations...),
			UnmappedMapping: lane.Source.UnmappedMapping,
		}
		missing := make([]string, 0)
		for _, artifact := range lane.Artifacts {
			if _, err := fs.Stat(fsys, path.Join(bundle.Name, artifact)); err != nil {
				missing = append(missing, fmt.Sprintf("artifact is unavailable: %q: %v", artifact, err))
			}
		}
		if len(missing) > 0 {
			result.Status = enabledContractFinalLaneMissing
			result.Reason = strings.Join(missing, "; ")
			results = append(results, result)
			continue
		}
		switch lane.Source.Coverage {
		case connectors.EnabledCoverageComplete:
			result.Status = enabledContractFinalLaneComplete
		case connectors.EnabledCoveragePartial:
			result.Status = enabledContractFinalLanePartial
		default:
			result.Status = enabledContractFinalLanePresent
		}
		results = append(results, result)
	}
	return results
}

func enabledContractFinding(connector, file string, err error) Finding {
	return Finding{
		Connector: connector,
		File:      file,
		Rule:      ruleEnabledConnectorContract,
		Message:   err.Error(),
	}
}

func loadEnabledContractSourceLock(fsys fs.FS, connector, artifact string) (sourceImportLock, error) {
	raw, err := fs.ReadFile(fsys, path.Join(connector, artifact))
	if err != nil {
		return sourceImportLock{}, fmt.Errorf("read source lock: %w", err)
	}
	lock, err := parseSourceImportLock(raw, connector)
	if err != nil {
		return sourceImportLock{}, err
	}
	return lock, nil
}

func enabledContractSourceOperations(lock sourceImportLock) []connectors.EnabledContractSourceOperation {
	operations := make([]connectors.EnabledContractSourceOperation, 0, lock.Counts.REST)
	if len(lock.Rest.SourceDocuments) == 0 {
		for _, operation := range lock.Rest.Operations {
			operations = append(operations, connectors.EnabledContractSourceOperation{ID: operation.ID, Method: operation.Method})
		}
		return operations
	}
	for _, document := range lock.Rest.SourceDocuments {
		for _, operation := range document.Operations {
			operations = append(operations, connectors.EnabledContractSourceOperation{ID: operation.ID, Method: operation.Method})
		}
	}
	return operations
}

func checkEnabledContractRetainedDocuments(fsys fs.FS, connector string, lock sourceImportLock) error {
	if lock.SchemaVersion < 3 {
		return nil
	}
	for _, document := range lock.Rest.SourceDocuments {
		if document.isUnavailable() || document.isSourceReference() {
			continue
		}
		artifact := document.Artifact
		artifactPath := path.Join(connector, "sources", "artifacts", strings.ToLower(artifact.SHA256)+sourceImportRetainedArtifactExtension)
		raw, err := fs.ReadFile(fsys, artifactPath)
		if err != nil {
			return fmt.Errorf("read retained source document %q: %w", document.ID, err)
		}
		if len(raw) != int(artifact.Bytes) || fmt.Sprintf("%x", sha256.Sum256(raw)) != strings.ToLower(artifact.SHA256) {
			return fmt.Errorf("retained source document %q does not match its source lock identity", document.ID)
		}
	}
	return nil
}

package main

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

// checkEnabledConnectorContract is the authoring-side source-lock bridge for
// the optional enabled connector contract. The engine validates the closed
// JSON and the actual runtime binding; this check owns immutable source
// identity reconciliation and retained supplemental-document evidence.
func checkEnabledConnectorContract(fsys fs.FS, bundle engine.Bundle) []Finding {
	contract := bundle.EnabledContract
	if contract == nil {
		return nil
	}

	findings := make([]Finding, 0)
	for _, lane := range contract.Lanes {
		for _, artifact := range lane.Artifacts {
			if _, err := fs.Stat(fsys, path.Join(bundle.Name, artifact)); err != nil {
				findings = append(findings, enabledContractFinding(bundle.Name, artifact, fmt.Errorf("lane %q artifact is unavailable: %w", lane.Name, err)))
			}
		}
	}
	primary, err := loadEnabledContractSourceLock(fsys, bundle.Name, contract.SourceLock.Path)
	if err != nil {
		return append(findings, enabledContractFinding(bundle.Name, contract.SourceLock.Path, err))
	}
	if primary.SchemaVersion < 3 && (primary.Rest.SHA256 != contract.SourceLock.SHA256 || primary.Rest.Bytes != contract.SourceLock.Bytes) {
		return append(findings, enabledContractFinding(bundle.Name, contract.SourceLock.Path, fmt.Errorf("primary source lock identity does not match enabled_connector_contract.json")))
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

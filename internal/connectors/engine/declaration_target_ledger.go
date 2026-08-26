package engine

import (
	"fmt"
	"io/fs"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

// DeclarationAdmissionSourcesFile is the required repository-level source
// denominator and the compact target ledger embedded into the production CLI.
const DeclarationAdmissionSourcesFile = "declaration_admission_sources.json"

type declarationAdmissionSourceCatalog struct {
	SchemaVersion      int                                   `json:"schema_version"`
	Cohort             string                                `json:"cohort"`
	ExpectedConnectors int                                   `json:"expected_connectors"`
	ExpectedOperations int                                   `json:"expected_source_operations"`
	SourceOperations   []declarationAdmissionSourceOperation `json:"source_operations"`
}

type declarationAdmissionSourceOperation struct {
	Connector           string                 `json:"connector"`
	ID                  string                 `json:"id"`
	Protocol            string                 `json:"protocol"`
	SourceURL           string                 `json:"source_url"`
	Location            string                 `json:"location"`
	ProviderOperationID string                 `json:"operation_id"`
	Method              string                 `json:"method"`
	BasePath            string                 `json:"base_path,omitempty"`
	Path                string                 `json:"path"`
	Binding             CommandBindingIdentity `json:"binding"`
	DestructiveKind     string                 `json:"destructive_kind"`
}

type declarationTargetLedger struct {
	entries map[string]connectors.CommandFoundationTarget
}

func loadDeclarationTargetLedgers(fsys fs.FS) (map[string]*declarationTargetLedger, error) {
	if !fileExists(fsys, DeclarationAdmissionSourcesFile) {
		return nil, nil
	}
	raw, err := readFile(fsys, DeclarationAdmissionSourcesFile)
	if err != nil {
		return nil, err
	}
	if err := ValidateDeclarationAdmissionSources(raw); err != nil {
		return nil, err
	}
	var catalog declarationAdmissionSourceCatalog
	if err := strictDecode(raw, &catalog); err != nil {
		return nil, fmt.Errorf("%s: %w", DeclarationAdmissionSourcesFile, err)
	}
	if catalog.SchemaVersion != 1 || strings.TrimSpace(catalog.Cohort) == "" {
		return nil, fmt.Errorf("%s: invalid schema version or empty cohort", DeclarationAdmissionSourcesFile)
	}
	if catalog.ExpectedOperations != len(catalog.SourceOperations) || catalog.ExpectedOperations <= 0 {
		return nil, fmt.Errorf("%s: expected_source_operations = %d, found %d", DeclarationAdmissionSourcesFile, catalog.ExpectedOperations, len(catalog.SourceOperations))
	}

	ledgers := make(map[string]*declarationTargetLedger)
	seenSources := make(map[string]struct{}, len(catalog.SourceOperations))
	seenIdentities := make(map[string]string, len(catalog.SourceOperations))
	seenBindings := make(map[string]string, len(catalog.SourceOperations))
	for index, source := range catalog.SourceOperations {
		canonicalSourceURL, err := validateDeclarationAdmissionSource(source)
		if err != nil {
			return nil, fmt.Errorf("%s: source operation %d: %w", DeclarationAdmissionSourcesFile, index+1, err)
		}
		if _, duplicate := seenSources[source.ID]; duplicate {
			return nil, fmt.Errorf("%s: duplicate source identity %q", DeclarationAdmissionSourcesFile, source.ID)
		}
		seenSources[source.ID] = struct{}{}
		method, effectivePath, err := declarationAdmissionSourceEndpoint(source)
		if err != nil {
			return nil, fmt.Errorf("%s: source operation %q: %w", DeclarationAdmissionSourcesFile, source.ID, err)
		}
		identity := strings.Join([]string{canonicalSourceURL, source.Location, source.Protocol, source.ProviderOperationID, method, effectivePath}, "\x00")
		if previous, duplicate := seenIdentities[identity]; duplicate {
			return nil, fmt.Errorf("%s: source operations %q and %q duplicate one exact provider operation", DeclarationAdmissionSourcesFile, previous, source.ID)
		}
		seenIdentities[identity] = source.ID
		bindingKey := strings.Join([]string{source.Connector, source.Binding.Kind, source.Binding.ID}, "\x00")
		if previous, duplicate := seenBindings[bindingKey]; duplicate {
			return nil, fmt.Errorf("%s: source operations %q and %q claim one canonical binding", DeclarationAdmissionSourcesFile, previous, source.ID)
		}
		seenBindings[bindingKey] = source.ID

		ledger := ledgers[source.Connector]
		if ledger == nil {
			ledger = &declarationTargetLedger{entries: make(map[string]connectors.CommandFoundationTarget)}
			ledgers[source.Connector] = ledger
		}
		ledger.entries[source.ID] = connectors.CommandFoundationTarget{
			SourceID: source.ID, ProviderOperationID: source.ProviderOperationID,
			Binding:         connectors.CommandBindingIdentity{Kind: source.Binding.Kind, ID: source.Binding.ID},
			DestructiveKind: source.DestructiveKind, Method: method, Path: effectivePath,
		}
	}
	if catalog.ExpectedConnectors != len(ledgers) || catalog.ExpectedConnectors <= 0 {
		return nil, fmt.Errorf("%s: expected_connectors = %d, found %d", DeclarationAdmissionSourcesFile, catalog.ExpectedConnectors, len(ledgers))
	}
	return ledgers, nil
}

func validateDeclarationAdmissionSource(source declarationAdmissionSourceOperation) (string, error) {
	if !namePattern.MatchString(source.Connector) {
		return "", fmt.Errorf("invalid connector %q", source.Connector)
	}
	values := []struct{ label, value string }{
		{label: "source identity", value: source.ID},
		{label: "document location", value: source.Location},
		{label: "binding identity", value: source.Binding.ID},
	}
	for _, item := range values {
		if item.value == "" || item.value != strings.TrimSpace(item.value) {
			return "", fmt.Errorf("%s must be nonempty and canonical", item.label)
		}
		if err := safety.RejectDangerousChars(item.value, item.label); err != nil {
			return "", err
		}
	}
	canonicalSourceURL, err := safety.CanonicalProviderCitationURL(source.SourceURL)
	if err != nil {
		return "", fmt.Errorf("invalid provider citation URL: %w", err)
	}
	if canonicalSourceURL != source.SourceURL {
		return "", fmt.Errorf("source URL must use the canonical provider citation URL")
	}
	if source.ProviderOperationID != strings.TrimSpace(source.ProviderOperationID) {
		return "", fmt.Errorf("provider operation identity must be canonical when present")
	}
	if source.Protocol != "rest" && source.Protocol != "graphql" {
		return "", fmt.Errorf("unsupported protocol %q", source.Protocol)
	}
	if !validCommandBinding(source.Binding.Kind, source.Binding.ID) {
		return "", fmt.Errorf("invalid binding %q/%q", source.Binding.Kind, source.Binding.ID)
	}
	switch source.DestructiveKind {
	case "none", "delete", "destructive":
	default:
		return "", fmt.Errorf("invalid destructive kind %q", source.DestructiveKind)
	}
	_, _, err = declarationAdmissionSourceEndpoint(source)
	return canonicalSourceURL, err
}

func declarationAdmissionSourceEndpoint(source declarationAdmissionSourceOperation) (string, string, error) {
	method := strings.ToUpper(strings.TrimSpace(source.Method))
	if source.Protocol == "graphql" && method == "GRAPHQL" {
		if source.BasePath != "" && source.BasePath != "/" {
			return "", "", fmt.Errorf("GraphQL operation identity must not declare an HTTP base path")
		}
		if err := ValidateCommandEndpoint(method, source.Path); err != nil {
			return "", "", err
		}
		return method, source.Path, nil
	}
	if source.Protocol == "rest" && method == "GRAPHQL" {
		return "", "", fmt.Errorf("REST source operation cannot declare a GraphQL operation identity")
	}
	path, err := declarationAdmissionSourcePath(source.BasePath, source.Path)
	if err != nil {
		return "", "", err
	}
	if err := ValidateCommandEndpoint(method, path); err != nil {
		return "", "", err
	}
	return method, path, nil
}

func declarationAdmissionSourcePath(basePath, operationPath string) (string, error) {
	if basePath == "" || basePath == "/" {
		return operationPath, nil
	}
	if err := ValidateCommandEndpoint("GET", basePath); err != nil {
		return "", fmt.Errorf("invalid base path: %w", err)
	}
	if err := ValidateCommandEndpoint("GET", operationPath); err != nil {
		return "", fmt.Errorf("invalid operation path: %w", err)
	}
	return strings.TrimRight(basePath, "/") + operationPath, nil
}

func validCommandBinding(kind, id string) bool {
	if id == "" || id != strings.TrimSpace(id) {
		return false
	}
	switch kind {
	case connectors.CommandBindingCommand, connectors.CommandBindingStream, connectors.CommandBindingWrite, connectors.CommandBindingOperation:
		return true
	default:
		return false
	}
}

func (ledger *declarationTargetLedger) target(sourceID string) (connectors.CommandFoundationTarget, bool) {
	if ledger == nil {
		return connectors.CommandFoundationTarget{}, false
	}
	target, ok := ledger.entries[sourceID]
	return target, ok
}

// DeclarationAdmissionTargets returns the compact source identities embedded
// for shipped deferred preflight. The returned slice is detached from Bundle.
func DeclarationAdmissionTargets(b Bundle) []connectors.CommandFoundationTarget {
	if b.declarationTargets == nil {
		return nil
	}
	out := make([]connectors.CommandFoundationTarget, 0, len(b.declarationTargets.entries))
	for _, target := range b.declarationTargets.entries {
		out = append(out, target)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].SourceID < out[j].SourceID })
	return out
}

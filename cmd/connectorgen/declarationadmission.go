package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/commandrunner"
	"polymetrics.ai/internal/connectors/engine"
	"polymetrics.ai/internal/failures"
)

var declarationAdmissionActionTemplateRE = regexp.MustCompile(`\{\{\s*(?:config|record)\.([-A-Za-z0-9_]+)\s*\}\}`)

const (
	declarationAdmissionSchemaVersion = 1

	declarationAdmissionStateImplemented = "implemented"
	declarationAdmissionStateDeferred    = "deferred"

	declarationAdmissionLaneETL            = "etl"
	declarationAdmissionLaneReverseETL     = "reverse_etl"
	declarationAdmissionLaneDirectRead     = "direct_read"
	declarationAdmissionLaneDirectWrite    = "direct_write"
	declarationAdmissionLaneBinaryDownload = "binary_download"
	declarationAdmissionLaneBinaryUpload   = "binary_upload"
)

type declarationAdmissionDocument struct {
	SchemaVersion    int                                   `json:"schema_version"`
	Connector        string                                `json:"connector"`
	SourceOperations []declarationAdmissionSourceOperation `json:"source_operations"`
	Declarations     []declarationAdmissionDeclaration     `json:"declarations"`
}

type declarationAdmissionSourceCatalog struct {
	SchemaVersion      int                                   `json:"schema_version"`
	Cohort             string                                `json:"cohort"`
	ExpectedConnectors int                                   `json:"expected_connectors"`
	ExpectedOperations int                                   `json:"expected_source_operations"`
	SourceOperations   []declarationAdmissionSourceOperation `json:"source_operations"`
}

type declarationAdmissionCatalog struct {
	SchemaVersion        int                               `json:"schema_version"`
	Cohort               string                            `json:"cohort"`
	ExpectedDeclarations int                               `json:"expected_declarations"`
	Declarations         []declarationAdmissionDeclaration `json:"declarations"`
}

type declarationAdmissionSourceOperation struct {
	Connector           string                      `json:"connector,omitempty"`
	ID                  string                      `json:"id"`
	Protocol            string                      `json:"protocol"`
	SourceURL           string                      `json:"source_url"`
	Location            string                      `json:"location"`
	ProviderOperationID string                      `json:"operation_id"`
	Method              string                      `json:"method"`
	BasePath            string                      `json:"base_path,omitempty"`
	Path                string                      `json:"path"`
	Binding             declarationAdmissionBinding `json:"binding"`
	DestructiveKind     string                      `json:"destructive_kind"`
}

type declarationAdmissionDeclaration struct {
	Connector   string                           `json:"connector,omitempty"`
	SourceID    string                           `json:"source_id"`
	Lane        string                           `json:"lane"`
	Command     string                           `json:"command"`
	State       string                           `json:"state"`
	Canonical   declarationAdmissionEndpoint     `json:"canonical"`
	Binding     declarationAdmissionBinding      `json:"binding"`
	Foundation  *declarationAdmissionFoundation  `json:"foundation_gap,omitempty"`
	Destructive *declarationAdmissionDestructive `json:"destructive,omitempty"`
}

type declarationAdmissionBinding struct {
	Kind string `json:"kind"`
	ID   string `json:"id"`
}

type declarationAdmissionEndpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type declarationAdmissionFoundation struct {
	ID        string                       `json:"id"`
	Reason    string                       `json:"reason"`
	Component string                       `json:"component"`
	Evidence  string                       `json:"evidence"`
	Target    declarationAdmissionEndpoint `json:"target"`
}

type declarationAdmissionDestructive struct {
	Kind   string `json:"kind"`
	Reason string `json:"reason"`
}

type declarationAdmissionReport struct {
	Findings          []Finding `json:"findings"`
	ConnectorsChecked int       `json:"connectors_checked"`
	SourceOperations  int       `json:"source_operations"`
}

type declarationAdmissionOptions struct {
	dir    string
	asJSON bool
}

// runDeclarationAdmission checks the required independent source cohort and
// its separate declaration catalog.
// It does not fetch provider material or run a provider operation. It reuses
// commandrunner's no-I/O preflight so executable and deferred declarations are
// checked by the same resolver as the CLI without becoming runtime or live
// certification.
func runDeclarationAdmission(args []string, stdout, stderr io.Writer) int {
	options, err := parseDeclarationAdmissionOptions(args)
	if err != nil {
		logf(stderr, "connectorgen declaration-admission: %v\n", err)
		return 2
	}
	report, err := declarationAdmissionPathCheck(options.dir)
	if err != nil {
		logf(stderr, "connectorgen declaration-admission: %v\n", err)
		return 1
	}
	if options.asJSON {
		encoder := json.NewEncoder(stdout)
		encoder.SetIndent("", "  ")
		if err := encoder.Encode(report); err != nil {
			logf(stderr, "connectorgen declaration-admission: encode report: %v\n", err)
			return 1
		}
	} else {
		for _, finding := range report.Findings {
			logf(stdout, "%s: %s: %s\n", finding.Connector, finding.File, finding.Message)
		}
		logf(stdout, "connectorgen declaration-admission: %d connector(s), %d source operation(s), %d finding(s)\n", report.ConnectorsChecked, report.SourceOperations, len(report.Findings))
	}
	if len(report.Findings) > 0 {
		return 1
	}
	return 0
}

func parseDeclarationAdmissionOptions(args []string) (declarationAdmissionOptions, error) {
	options := declarationAdmissionOptions{}
	for _, argument := range args[1:] {
		switch argument {
		case "--json":
			options.asJSON = true
		default:
			if strings.HasPrefix(argument, "-") || options.dir != "" {
				return declarationAdmissionOptions{}, fmt.Errorf("unexpected argument %q", argument)
			}
			options.dir = argument
		}
	}
	if options.dir != "" {
		return options, nil
	}
	root, err := repoRoot()
	if err != nil {
		return declarationAdmissionOptions{}, fmt.Errorf("resolve repository root: %w", err)
	}
	options.dir = filepath.Join(root, "internal", "connectors", "defs")
	return options, nil
}

func declarationAdmissionPathCheck(dir string) (declarationAdmissionReport, error) {
	sourceFile := filepath.Join(dir, "declaration_admission_sources.json")
	declarationFile := filepath.Join(dir, "declaration_admissions.json")
	sourceRaw, err := os.ReadFile(sourceFile)
	if err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("read required independent source cohort %s: %w", sourceFile, err)
	}
	if err := engine.ValidateDeclarationAdmissionSources(sourceRaw); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("validate independent source cohort: %w", err)
	}
	declarationRaw, err := os.ReadFile(declarationFile)
	if err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("read required declaration catalog %s: %w", declarationFile, err)
	}
	if err := engine.ValidateDeclarationAdmission(declarationRaw); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("validate declaration catalog: %w", err)
	}

	var sources declarationAdmissionSourceCatalog
	if err := decodeSourceStrictJSON(sourceRaw, &sources); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("parse independent source cohort: %w", err)
	}
	var declarations declarationAdmissionCatalog
	if err := decodeSourceStrictJSON(declarationRaw, &declarations); err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("parse declaration catalog: %w", err)
	}

	report := declarationAdmissionReport{Findings: []Finding{}, SourceOperations: len(sources.SourceOperations)}
	addCatalogFinding := func(file, message string) {
		report.Findings = append(report.Findings, Finding{File: file, Rule: "declaration_admission", Message: message})
	}
	if sources.SchemaVersion != declarationAdmissionSchemaVersion || declarations.SchemaVersion != declarationAdmissionSchemaVersion {
		addCatalogFinding("declaration_admission_sources.json", "catalog schema_version must be 1")
	}
	if strings.TrimSpace(sources.Cohort) == "" || sources.Cohort != declarations.Cohort {
		addCatalogFinding("declaration_admissions.json", "declaration cohort does not match the independent source cohort")
	}
	if sources.ExpectedConnectors <= 0 || sources.ExpectedOperations <= 0 || declarations.ExpectedDeclarations <= 0 {
		addCatalogFinding("declaration_admission_sources.json", "repository admission cohort expected counts must be nonzero")
	}
	if sources.ExpectedOperations != len(sources.SourceOperations) {
		addCatalogFinding("declaration_admission_sources.json", fmt.Sprintf("expected_source_operations = %d, found %d", sources.ExpectedOperations, len(sources.SourceOperations)))
	}
	if declarations.ExpectedDeclarations != len(declarations.Declarations) {
		addCatalogFinding("declaration_admissions.json", fmt.Sprintf("expected_declarations = %d, found %d", declarations.ExpectedDeclarations, len(declarations.Declarations)))
	}

	documents := map[string]*declarationAdmissionDocument{}
	for _, source := range sources.SourceOperations {
		document := documents[source.Connector]
		if document == nil {
			document = &declarationAdmissionDocument{SchemaVersion: declarationAdmissionSchemaVersion, Connector: source.Connector}
			documents[source.Connector] = document
		}
		document.SourceOperations = append(document.SourceOperations, source)
	}
	for _, declaration := range declarations.Declarations {
		document := documents[declaration.Connector]
		if document == nil {
			document = &declarationAdmissionDocument{SchemaVersion: declarationAdmissionSchemaVersion, Connector: declaration.Connector}
			documents[declaration.Connector] = document
		}
		document.Declarations = append(document.Declarations, declaration)
	}
	if sources.ExpectedConnectors != len(documents) {
		addCatalogFinding("declaration_admission_sources.json", fmt.Sprintf("expected_connectors = %d, found %d", sources.ExpectedConnectors, len(documents)))
	}

	connectorNames := make([]string, 0, len(documents))
	for connector := range documents {
		connectorNames = append(connectorNames, connector)
	}
	sort.Strings(connectorNames)
	report.ConnectorsChecked = len(connectorNames)
	for _, connector := range connectorNames {
		bundle, err := engine.Load(os.DirFS(dir), connector)
		if err != nil {
			report.Findings = append(report.Findings, declarationAdmissionFinding(connector, declarationFile, "load connector bundle: "+err.Error()))
			continue
		}
		report.Findings = append(report.Findings, declarationAdmissionFindings(bundle, *documents[connector])...)
	}
	sort.Slice(report.Findings, func(i, j int) bool {
		if report.Findings[i].Connector != report.Findings[j].Connector {
			return report.Findings[i].Connector < report.Findings[j].Connector
		}
		if report.Findings[i].File != report.Findings[j].File {
			return report.Findings[i].File < report.Findings[j].File
		}
		return report.Findings[i].Message < report.Findings[j].Message
	})
	return report, nil
}

func declarationAdmissionFindings(bundle engine.Bundle, document declarationAdmissionDocument) []Finding {
	findings := []Finding{}
	file := "declaration_admissions.json"
	add := func(message string) {
		findings = append(findings, Finding{Connector: bundle.Name, File: file, Rule: "declaration_admission", Message: message})
	}
	if document.SchemaVersion != declarationAdmissionSchemaVersion {
		add(fmt.Sprintf("schema_version = %d, want %d", document.SchemaVersion, declarationAdmissionSchemaVersion))
	}
	if document.Connector != bundle.Name {
		add(fmt.Sprintf("connector %q does not match bundle %q", document.Connector, bundle.Name))
	}
	if len(document.SourceOperations) == 0 {
		add("source declaration has no operations")
	}
	sources := make(map[string]declarationAdmissionSourceOperation, len(document.SourceOperations))
	identities := make(map[string]string, len(document.SourceOperations))
	for _, source := range document.SourceOperations {
		if source.Connector != "" && source.Connector != bundle.Name {
			add(fmt.Sprintf("source operation %q belongs to connector %q", source.ID, source.Connector))
			continue
		}
		if source.ID == "" {
			add("source operation has no identity")
			continue
		}
		if _, duplicate := sources[source.ID]; duplicate {
			add("duplicate source operation " + source.ID)
			continue
		}
		sources[source.ID] = source
		if !declarationAdmissionProviderURL(source.SourceURL) {
			add("source operation " + source.ID + " has no valid provider source URL")
		}
		if strings.TrimSpace(source.Location) == "" {
			add("source operation " + source.ID + " has no exact provider operation citation")
		}
		if source.ProviderOperationID != strings.TrimSpace(source.ProviderOperationID) {
			add("source operation " + source.ID + " has a noncanonical provider operation identity")
		}
		if source.Protocol != "rest" && source.Protocol != "graphql" {
			add("source operation " + source.ID + " has an invalid protocol")
		}
		if !declarationAdmissionBindingValid(source.Binding) {
			add("source operation " + source.ID + " has no exact canonical binding")
		}
		switch source.DestructiveKind {
		case "none", "delete", "destructive":
		default:
			add("source operation " + source.ID + " has an invalid destructive semantic")
		}
		effectivePath, err := declarationAdmissionEffectivePath(source.BasePath, source.Path)
		if err != nil {
			add("source operation " + source.ID + ": " + err.Error())
			continue
		}
		if err := engine.ValidateCommandEndpoint(strings.ToUpper(strings.TrimSpace(source.Method)), effectivePath); err != nil {
			add("source operation " + source.ID + ": " + err.Error())
			continue
		}
		identity := declarationAdmissionSourceIdentity(source)
		if previous, duplicate := identities[identity]; duplicate {
			add(fmt.Sprintf("duplicate exact provider operation identity for source operations %s and %s", previous, source.ID))
		} else {
			identities[identity] = source.ID
		}
	}
	declared := make(map[string]bool, len(document.Declarations))
	for _, declaration := range document.Declarations {
		if declaration.Connector != "" && declaration.Connector != bundle.Name {
			add(fmt.Sprintf("declaration for source operation %q belongs to connector %q", declaration.SourceID, declaration.Connector))
			continue
		}
		source, exists := sources[declaration.SourceID]
		if declaration.SourceID == "" || !exists {
			add("declaration references unknown source operation " + declaration.SourceID)
			continue
		}
		if declared[declaration.SourceID] {
			add("duplicate declaration for source operation " + declaration.SourceID)
			continue
		}
		declared[declaration.SourceID] = true
		declarationAdmissionCheckRow(&findings, bundle, file, source, declaration)
	}
	for sourceID := range sources {
		if !declared[sourceID] {
			add("source operation " + sourceID + " has no declaration")
		}
	}
	sort.Slice(findings, func(i, j int) bool { return findings[i].Message < findings[j].Message })
	return findings
}

func declarationAdmissionProviderURL(raw string) bool {
	return validateSourceImportPublishedURL(raw) == nil
}

// declarationAdmissionSourceIdentity is the source-row completeness key. It
// intentionally does not consult a source lock or importer: a URL plus its
// exact documented operation is enough to establish the independent admission
// denominator, but two aliases for that same documented operation are not.
func declarationAdmissionSourceIdentity(source declarationAdmissionSourceOperation) string {
	path, err := declarationAdmissionEffectivePath(source.BasePath, source.Path)
	if err != nil {
		return ""
	}
	return strings.Join([]string{source.SourceURL, source.Location, strings.ToUpper(strings.TrimSpace(source.Method)), path, source.Binding.Kind, source.Binding.ID}, "\x00")
}

func declarationAdmissionCheckRow(findings *[]Finding, bundle engine.Bundle, file string, source declarationAdmissionSourceOperation, declaration declarationAdmissionDeclaration) {
	add := func(message string) {
		*findings = append(*findings, Finding{Connector: bundle.Name, File: file, Rule: "declaration_admission", Message: "source operation " + source.ID + ": " + message})
	}
	if !declarationAdmissionLaneValid(declaration.Lane) {
		add("lane is not one of the admission lanes")
		return
	}
	effectivePath, err := declarationAdmissionEffectivePath(source.BasePath, source.Path)
	if err != nil {
		add(err.Error())
		return
	}
	if !strings.EqualFold(declaration.Canonical.Method, source.Method) || declaration.Canonical.Path != effectivePath {
		add("base-path mismatch or stale canonical endpoint")
	}
	if !declarationAdmissionBindingValid(declaration.Binding) || !declarationAdmissionBindingsMatch(declaration.Binding, source.Binding) {
		add("canonical binding does not match the independent source operation")
	}
	switch source.DestructiveKind {
	case "none":
		if declaration.Destructive != nil {
			if declaration.Destructive.Kind == "delete" {
				add("delete semantics do not match the independent source semantic")
			} else {
				add("destructive metadata does not match the independent source semantic")
			}
		}
	case "delete", "destructive":
		if declaration.Destructive == nil {
			add(source.DestructiveKind + " operation lacks destructive metadata")
		} else if declaration.Destructive.Kind != source.DestructiveKind {
			add("delete semantics do not match the independent source semantic")
		} else if strings.TrimSpace(declaration.Destructive.Reason) == "" {
			add("destructive metadata lacks a reason")
		}
	}
	commandPath, commandPathErr := commandrunner.CommandPathSegments(declaration.Command)
	if declaration.Command == "" {
		add("has no discoverable command mapping")
		return
	}
	if commandPathErr != nil {
		add("has no canonical round-trippable command path: " + commandPathErr.Error())
		return
	}
	command, matches := declarationAdmissionCommand(bundle, declaration.Command)
	if matches != 1 {
		add("has no unique discoverable command mapping")
		return
	}
	if declarationAdmissionLaneForIntent(command.Intent) != declaration.Lane {
		add("lane does not match command intent")
	}
	if !declarationAdmissionCommandCitesEndpoint(command, declaration.Canonical) || !declarationAdmissionSurfaceHasEndpoint(bundle.Surface, declaration.Canonical) {
		add("command does not map to the canonical API surface endpoint")
	}
	switch declaration.State {
	case declarationAdmissionStateImplemented:
		if declaration.Foundation != nil || command.Foundation != nil {
			add("implemented declaration must not retain a foundation gap")
		}
		resolved, resolveErr := engine.ResolveImplementedCommandPath(bundle, declaration.Command)
		if command.Availability != declarationAdmissionStateImplemented || resolveErr != nil ||
			!strings.EqualFold(resolved.Method, declaration.Canonical.Method) || resolved.Path != declaration.Canonical.Path ||
			!declarationAdmissionRuntimeBindingMatches(resolved.Binding, source.Binding) {
			add("implemented declaration has no valid runtime binding")
		} else if err := commandrunner.Preflight(engine.New(bundle, nil), commandPath); err != nil {
			add("implemented declaration fails runtime preflight: " + err.Error())
		}
		if !declarationAdmissionResolvedDestructiveMatches(resolved, source.DestructiveKind) {
			add("implemented destructive declaration does not retain destructive runtime metadata")
		}
	case declarationAdmissionStateDeferred:
		if !declarationAdmissionFoundationSpecific(declaration.Foundation) {
			add("deferred declaration requires a specific missing implementation component with evidence")
		} else if message := declarationAdmissionFoundationEvidenceFinding(bundle, declaration, declaration.Foundation); message != "" {
			add(message)
		}
		if !declarationAdmissionFoundationMatches(declaration.Foundation, command.Foundation) {
			add("deferred declaration requires the same named foundation gap on its command")
		}
		if !declarationAdmissionFoundationTargetMatchesSource(command.Foundation, source) {
			add("deferred command foundation target does not retain the exact source identity and binding")
		}
		if declaration.Foundation != nil && !declarationAdmissionEndpointsMatch(declaration.Foundation.Target, declaration.Canonical) {
			add("deferred declaration foundation target does not match the canonical API surface endpoint")
		}
		if command.Availability != declarationAdmissionStateDeferred {
			add("deferred declaration command is not deferred")
		}
		if message := declarationAdmissionDeferredPreflight(bundle, commandPath); message != "" {
			add(message)
		}
	default:
		add("state must be implemented or deferred")
	}
}

func declarationAdmissionFinding(connector, filename, message string) Finding {
	return Finding{Connector: connector, File: filepath.ToSlash(filepath.Join("sources", filepath.Base(filename))), Rule: "declaration_admission", Message: message}
}

func declarationAdmissionLaneValid(lane string) bool {
	switch lane {
	case declarationAdmissionLaneETL, declarationAdmissionLaneReverseETL, declarationAdmissionLaneDirectRead, declarationAdmissionLaneDirectWrite, declarationAdmissionLaneBinaryDownload, declarationAdmissionLaneBinaryUpload:
		return true
	default:
		return false
	}
}

func declarationAdmissionLaneForIntent(intent string) string {
	switch intent {
	case declarationAdmissionLaneETL, declarationAdmissionLaneReverseETL, declarationAdmissionLaneDirectRead, declarationAdmissionLaneDirectWrite, declarationAdmissionLaneBinaryDownload, declarationAdmissionLaneBinaryUpload:
		return intent
	default:
		return ""
	}
}

func declarationAdmissionEffectivePath(basePath, operationPath string) (string, error) {
	if !strings.HasPrefix(operationPath, "/") {
		return "", errors.New("source path must be a connector-relative absolute path")
	}
	effective := operationPath
	if basePath == "" || basePath == "/" {
		if err := engine.ValidateCommandEndpoint("GET", effective); err != nil {
			return "", errors.New("source path must be a canonical connector-relative absolute path")
		}
		return effective, nil
	}
	if err := engine.ValidateCommandEndpoint("GET", basePath); err != nil {
		return "", errors.New("base path must be a canonical connector-relative absolute path")
	}
	effective = strings.TrimRight(basePath, "/") + operationPath
	if err := engine.ValidateCommandEndpoint("GET", effective); err != nil {
		return "", errors.New("source path must be a canonical connector-relative absolute path")
	}
	return effective, nil
}

func declarationAdmissionCommand(bundle engine.Bundle, path string) (engine.CLICommand, int) {
	if bundle.CLISurface == nil {
		return engine.CLICommand{}, 0
	}
	var command engine.CLICommand
	count := 0
	for _, candidate := range bundle.CLISurface.Commands {
		if candidate.Path == path {
			command = candidate
			count++
		}
	}
	return command, count
}

func declarationAdmissionCommandCitesEndpoint(command engine.CLICommand, endpoint declarationAdmissionEndpoint) bool {
	return len(command.APISurface) == 1 && strings.EqualFold(command.APISurface[0].Method, endpoint.Method) && command.APISurface[0].Path == endpoint.Path
}

func declarationAdmissionSurfaceHasEndpoint(surface *engine.APISurface, endpoint declarationAdmissionEndpoint) bool {
	if surface == nil {
		return false
	}
	matches := 0
	for _, candidate := range surface.Endpoints {
		if strings.EqualFold(candidate.Method, endpoint.Method) && candidate.Path == endpoint.Path {
			matches++
			if candidate.Excluded != nil || (candidate.Operation != nil && candidate.Operation.Model == "disallowed") {
				return false
			}
		}
	}
	return matches == 1
}

func declarationAdmissionFoundationMatches(declaration *declarationAdmissionFoundation, command *engine.CommandFoundation) bool {
	return declarationAdmissionFoundationSpecific(declaration) && command != nil && declaration.ID == command.ID && declaration.Reason == command.Reason &&
		declaration.Component == command.Component && declaration.Evidence == command.Evidence &&
		declarationAdmissionEndpointsMatch(declaration.Target, declarationAdmissionEndpoint{Method: command.Target.Method, Path: command.Target.Path})
}

func declarationAdmissionFoundationSpecific(foundation *declarationAdmissionFoundation) bool {
	return foundation != nil && strings.TrimSpace(foundation.ID) != "" && strings.TrimSpace(foundation.Reason) != "" &&
		declarationAdmissionFoundationComponentValid(foundation.Component) && declarationAdmissionFoundationEvidenceValid(foundation.Component, foundation.Evidence) &&
		strings.TrimSpace(foundation.Target.Method) != "" && strings.TrimSpace(foundation.Target.Path) != ""
}

// declarationAdmissionFoundationComponentValid accepts missing runtime pieces,
// never a provider policy, operation kind, source-retention choice, or a live
// certification state. Those facts can explain risk, but cannot remove a
// cited source operation from the declaration and command surface.
func declarationAdmissionFoundationComponentValid(component string) bool {
	return connectors.ValidCommandFoundationComponent(component)
}

func declarationAdmissionFoundationEvidenceValid(component, evidence string) bool {
	return connectors.ValidCommandFoundationEvidence(component, evidence)
}

func declarationAdmissionFoundationEvidenceFinding(bundle engine.Bundle, declaration declarationAdmissionDeclaration, foundation *declarationAdmissionFoundation) string {
	if foundation.Component != "typed_write_action" {
		return ""
	}
	if declaration.Lane != declarationAdmissionLaneReverseETL && declaration.Lane != declarationAdmissionLaneBinaryUpload {
		return "typed_write_action foundation does not apply to this lane"
	}
	for _, action := range bundle.Writes {
		if strings.EqualFold(action.Method, declaration.Canonical.Method) && declarationAdmissionActionPath(action.Path) == declaration.Canonical.Path {
			return "typed_write_action foundation is stale: a declared write action already maps to the canonical endpoint"
		}
	}
	return ""
}

func declarationAdmissionActionPath(path string) string {
	return declarationAdmissionActionTemplateRE.ReplaceAllString(path, "{$1}")
}

func declarationAdmissionEndpointsMatch(left, right declarationAdmissionEndpoint) bool {
	return strings.EqualFold(left.Method, right.Method) && left.Path == right.Path
}

func declarationAdmissionBindingValid(binding declarationAdmissionBinding) bool {
	if binding.ID == "" || binding.ID != strings.TrimSpace(binding.ID) {
		return false
	}
	switch binding.Kind {
	case connectors.CommandBindingCommand, connectors.CommandBindingStream, connectors.CommandBindingWrite, connectors.CommandBindingOperation:
		return true
	default:
		return false
	}
}

func declarationAdmissionBindingsMatch(left, right declarationAdmissionBinding) bool {
	return left.Kind == right.Kind && left.ID == right.ID
}

func declarationAdmissionRuntimeBindingMatches(runtime connectors.CommandBindingIdentity, source declarationAdmissionBinding) bool {
	return runtime.Kind == source.Kind && runtime.ID == source.ID
}

func declarationAdmissionFoundationTargetMatchesSource(foundation *engine.CommandFoundation, source declarationAdmissionSourceOperation) bool {
	if foundation == nil {
		return false
	}
	effectivePath, err := declarationAdmissionEffectivePath(source.BasePath, source.Path)
	if err != nil {
		return false
	}
	target := foundation.Target
	return target.SourceID == source.ID && target.ProviderOperationID == source.ProviderOperationID &&
		target.Binding.Kind == source.Binding.Kind && target.Binding.ID == source.Binding.ID &&
		target.DestructiveKind == source.DestructiveKind && strings.EqualFold(target.Method, source.Method) && target.Path == effectivePath
}

func declarationAdmissionResolvedDestructiveMatches(resolved engine.ResolvedCommandBinding, semantic string) bool {
	switch semantic {
	case "none":
		return !resolved.Destructive.RequiresApproval()
	case "delete":
		return strings.EqualFold(resolved.Destructive.MutationClass, "delete") && resolved.Destructive.RequiresApproval()
	case "destructive":
		return resolved.Destructive.RequiresApproval()
	default:
		return false
	}
}

func declarationAdmissionDeferredPreflight(bundle engine.Bundle, path []string) string {
	err := commandrunner.Preflight(engine.New(bundle, nil), path)
	var blocked *commandrunner.BlockedCommandError
	if !errors.As(err, &blocked) {
		return fmt.Sprintf("deferred target does not pass runtime preflight: %v", err)
	}
	if blocked.Failure == nil || blocked.Failure.Domain() != failures.DomainSystem || blocked.Failure.Code() != "missing_foundation" {
		return fmt.Sprintf("deferred target does not resolve to system/missing_foundation: %v", blocked)
	}
	return ""
}

package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
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

type declarationAdmissionSourceOperation struct {
	ID                  string `json:"id"`
	SourceURL           string `json:"source_url"`
	Location            string `json:"location"`
	ProviderOperationID string `json:"operation_id"`
	Method              string `json:"method"`
	BasePath            string `json:"base_path,omitempty"`
	Path                string `json:"path"`
}

type declarationAdmissionDeclaration struct {
	SourceID    string                           `json:"source_id"`
	Lane        string                           `json:"lane"`
	Command     string                           `json:"command"`
	State       string                           `json:"state"`
	Canonical   declarationAdmissionEndpoint     `json:"canonical"`
	Foundation  *declarationAdmissionFoundation  `json:"foundation_gap,omitempty"`
	Destructive *declarationAdmissionDestructive `json:"destructive,omitempty"`
}

type declarationAdmissionEndpoint struct {
	Method string `json:"method"`
	Path   string `json:"path"`
}

type declarationAdmissionFoundation struct {
	ID        string `json:"id"`
	Reason    string `json:"reason"`
	Component string `json:"component"`
	Evidence  string `json:"evidence"`
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

// runDeclarationAdmission checks optional, source-cited admission sidecars.
// It does not fetch provider material or invoke commandrunner: those belong to
// source-lock retention and runtime certification, respectively.
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
	entries, err := os.ReadDir(dir)
	if err != nil {
		return declarationAdmissionReport{}, fmt.Errorf("read definitions: %w", err)
	}
	report := declarationAdmissionReport{Findings: []Finding{}}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		connector := entry.Name()
		admissionFile := filepath.Join(dir, connector, "sources", connector+"-declaration-admission.json")
		raw, err := os.ReadFile(admissionFile)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		report.ConnectorsChecked++
		if err != nil {
			report.Findings = append(report.Findings, declarationAdmissionFinding(connector, admissionFile, "read admission declaration: "+err.Error()))
			continue
		}
		if err := engine.ValidateDeclarationAdmission(raw); err != nil {
			report.Findings = append(report.Findings, declarationAdmissionFinding(connector, admissionFile, "schema: "+err.Error()))
			continue
		}
		var document declarationAdmissionDocument
		if err := decodeSourceStrictJSON(raw, &document); err != nil {
			report.Findings = append(report.Findings, declarationAdmissionFinding(connector, admissionFile, "parse admission declaration: "+err.Error()))
			continue
		}
		bundle, err := engine.Load(os.DirFS(dir), connector)
		if err != nil {
			report.Findings = append(report.Findings, declarationAdmissionFinding(connector, admissionFile, "load connector bundle: "+err.Error()))
			continue
		}
		report.SourceOperations += len(document.SourceOperations)
		report.Findings = append(report.Findings, declarationAdmissionFindings(bundle, document)...)
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
	file := filepath.ToSlash(filepath.Join("sources", bundle.Name+"-declaration-admission.json"))
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
	for _, source := range document.SourceOperations {
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
		if strings.TrimSpace(source.Location) == "" || strings.TrimSpace(source.ProviderOperationID) == "" {
			add("source operation " + source.ID + " has no exact provider operation citation")
		}
		if _, err := declarationAdmissionEffectivePath(source.BasePath, source.Path); err != nil {
			add("source operation " + source.ID + ": " + err.Error())
		}
	}
	declared := make(map[string]bool, len(document.Declarations))
	for _, declaration := range document.Declarations {
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
	if raw == "" || raw != strings.TrimSpace(raw) || strings.ContainsAny(raw, "\r\n") {
		return false
	}
	parsed, err := url.Parse(raw)
	if err != nil || !parsed.IsAbs() || parsed.Host == "" || parsed.User != nil {
		return false
	}
	return parsed.Scheme == "https" || parsed.Scheme == "http"
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
	if strings.EqualFold(source.Method, "DELETE") {
		if declaration.Destructive == nil || declaration.Destructive.Kind != "delete" || strings.TrimSpace(declaration.Destructive.Reason) == "" {
			add("delete operation lacks destructive metadata")
		}
	} else if declaration.Destructive != nil && strings.TrimSpace(declaration.Destructive.Reason) == "" {
		add("destructive metadata lacks a reason")
	}
	switch declaration.State {
	case declarationAdmissionStateImplemented:
		if declaration.Foundation != nil || command.Foundation != nil {
			add("implemented declaration must not retain a foundation gap")
		}
		if command.Availability != declarationAdmissionStateImplemented || !declarationAdmissionImplementedBinding(bundle, command, declaration.Lane, declaration.Canonical) {
			add("implemented declaration has no valid runtime binding")
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
		if command.Availability != declarationAdmissionStateDeferred {
			add("deferred declaration command is not deferred")
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
	if !strings.HasPrefix(operationPath, "/") || strings.Contains(operationPath, "?") || strings.Contains(operationPath, "#") || strings.Contains(operationPath, "..") {
		return "", errors.New("source path must be a connector-relative absolute path")
	}
	if basePath == "" || basePath == "/" {
		return operationPath, nil
	}
	if !strings.HasPrefix(basePath, "/") || strings.Contains(basePath, "?") || strings.Contains(basePath, "#") || strings.Contains(basePath, "..") {
		return "", errors.New("base path must be a connector-relative absolute path")
	}
	return strings.TrimRight(basePath, "/") + operationPath, nil
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
	for _, reference := range command.APISurface {
		if strings.EqualFold(reference.Method, endpoint.Method) && reference.Path == endpoint.Path {
			return true
		}
	}
	return false
}

func declarationAdmissionSurfaceHasEndpoint(surface *engine.APISurface, endpoint declarationAdmissionEndpoint) bool {
	if surface == nil {
		return false
	}
	for _, candidate := range surface.Endpoints {
		if strings.EqualFold(candidate.Method, endpoint.Method) && candidate.Path == endpoint.Path {
			return true
		}
	}
	return false
}

func declarationAdmissionFoundationMatches(declaration *declarationAdmissionFoundation, command *engine.CommandFoundation) bool {
	return declarationAdmissionFoundationSpecific(declaration) && command != nil && declaration.ID == command.ID && declaration.Reason == command.Reason
}

func declarationAdmissionFoundationSpecific(foundation *declarationAdmissionFoundation) bool {
	return foundation != nil && strings.TrimSpace(foundation.ID) != "" && strings.TrimSpace(foundation.Reason) != "" &&
		declarationAdmissionFoundationComponentValid(foundation.Component) && strings.TrimSpace(foundation.Evidence) != ""
}

// declarationAdmissionFoundationComponentValid accepts missing runtime pieces,
// never a provider policy, operation kind, source-retention choice, or a live
// certification state. Those facts can explain risk, but cannot remove a
// cited source operation from the declaration and command surface.
func declarationAdmissionFoundationComponentValid(component string) bool {
	switch component {
	case "typed_write_action", "typed_record_schema", "typed_request_body", "typed_response_descriptor",
		"binary_transfer_binding", "source_importer", "runtime_executor", "idempotency_contract", "conformance_fixture":
		return true
	default:
		return false
	}
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

func declarationAdmissionImplementedBinding(bundle engine.Bundle, command engine.CLICommand, lane string, endpoint declarationAdmissionEndpoint) bool {
	switch lane {
	case declarationAdmissionLaneETL:
		for _, stream := range bundle.Streams {
			method := stream.Method
			if method == "" {
				method = "GET"
			}
			if command.Stream == stream.Name && strings.EqualFold(method, endpoint.Method) && stream.Path == endpoint.Path {
				return true
			}
		}
	case declarationAdmissionLaneReverseETL, declarationAdmissionLaneBinaryUpload:
		for _, action := range bundle.Writes {
			if command.Write == action.Name && strings.EqualFold(action.Method, endpoint.Method) && declarationAdmissionActionPath(action.Path) == endpoint.Path {
				return true
			}
		}
	case declarationAdmissionLaneDirectRead, declarationAdmissionLaneDirectWrite, declarationAdmissionLaneBinaryDownload:
		for _, operation := range bundle.Operations {
			if command.Operation != operation.ID {
				continue
			}
			if operation.REST != nil && strings.EqualFold(operation.REST.Method, endpoint.Method) && operation.REST.Path == endpoint.Path {
				return true
			}
			if operation.Binary != nil && strings.EqualFold(operation.Binary.Method, endpoint.Method) && operation.Binary.Path == endpoint.Path {
				return true
			}
		}
	}
	return false
}

func declarationAdmissionActionPath(path string) string {
	return declarationAdmissionActionTemplateRE.ReplaceAllString(path, "{$1}")
}

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

// sourceMaterializeUsage intentionally promises only the closed v4 bridge.
// It is not a generic HTTP/OpenAPI-to-runtime compiler: request and response
// shapes remain executable only when the v4 plan selects one of the typed
// bindings below.
const sourceMaterializeUsage = `usage:
  connectorgen source-materialize <connector> [--defs <dir>] [--check]

Materializes one schema-v4, connector-owned source lock into its complete
declarative bundle. The command is offline: retained source bytes and the
locked materialization plan are the only inputs. It stages every owned output,
validates the staged bundle, and publishes only after all checks pass.

  --check  verifies byte-identical owned outputs; does not write
  --defs    connector definitions root (default internal/connectors/defs)`

type sourceMaterializeOptions struct {
	Connector string
	DefsDir   string
	Check     bool
}

// sourceMaterialization is the strict v4 top-level plan. Its fields are
// deliberately declarative and narrow: it names a connector shape, safe
// configuration/auth/server defaults, one fixed check, and an accounting row
// for every operation in the retained provider inventory.
type sourceMaterialization struct {
	Metadata   sourceMaterializationMetadata    `json:"metadata"`
	Config     sourceMaterializationConfig      `json:"config"`
	Auth       sourceMaterializationAuth        `json:"auth"`
	Server     sourceMaterializationServer      `json:"server"`
	Check      sourceMaterializationCheck       `json:"check"`
	Operations []sourceMaterializationOperation `json:"operations"`
}

type sourceMaterializationMetadata struct {
	DisplayName     string `json:"display_name"`
	Description     string `json:"description"`
	IntegrationType string `json:"integration_type"`
	ReleaseStage    string `json:"release_stage"`
	DocsURL         string `json:"docs_url"`
}

type sourceMaterializationConfig struct {
	Properties []sourceMaterializationConfigProperty `json:"properties"`
}

type sourceMaterializationConfigProperty struct {
	Name        string  `json:"name"`
	Type        string  `json:"type"`
	Format      string  `json:"format,omitempty"`
	Description string  `json:"description"`
	Default     *string `json:"default,omitempty"`
	Required    bool    `json:"required,omitempty"`
	Secret      bool    `json:"secret,omitempty"`
}

type sourceMaterializationAuth struct {
	Mode     string `json:"mode"`
	Token    string `json:"token,omitempty"`
	Header   string `json:"header,omitempty"`
	Prefix   string `json:"prefix,omitempty"`
	Param    string `json:"param,omitempty"`
	Value    string `json:"value,omitempty"`
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type sourceMaterializationServer struct {
	URL       string `json:"url"`
	UserAgent string `json:"user_agent"`
}

type sourceMaterializationCheck struct {
	SourceID        string   `json:"source_id"`
	Method          string   `json:"method"`
	Path            string   `json:"path"`
	SuccessStatuses []string `json:"success_statuses"`
}

// sourceMaterializationOperation records the disposition for exactly one
// source operation. A materialized row must select a closed binding. Blocked
// and unsupported rows must retain their citation and an actionable reason;
// neither creates a transport escape hatch.
type sourceMaterializationOperation struct {
	SourceID string                              `json:"source_id"`
	State    string                              `json:"state"`
	Citation sourceMaterializationCitation       `json:"citation"`
	Binding  *sourceMaterializationOperationBind `json:"binding,omitempty"`
	Reason   string                              `json:"reason,omitempty"`
	Inputs   []sourceMaterializationInputBinding `json:"inputs,omitempty"`
}

type sourceMaterializationCitation struct {
	DocumentID string `json:"document_id"`
	Location   string `json:"location"`
}

// sourceMaterializationOperationBind supports only bounded GET direct reads
// and JSON reverse-ETL writes. The fields are intentionally explicit so no
// source response status, output policy, command name, body media arm, risk,
// or approval rule is guessed by this command.
type sourceMaterializationOperationBind struct {
	Kind             string   `json:"kind"`
	ID               string   `json:"id"`
	CommandPath      string   `json:"command_path,omitempty"`
	CommandSummary   string   `json:"command_summary,omitempty"`
	OutputPolicy     string   `json:"output_policy"`
	MaxResponseBytes int      `json:"max_response_bytes"`
	SuccessStatuses  []string `json:"success_statuses"`
	RequestMedia     string   `json:"request_media,omitempty"`
	WriteKind        string   `json:"write_kind,omitempty"`
	MutationClass    string   `json:"mutation_class,omitempty"`
	Approval         string   `json:"approval,omitempty"`
	Risk             string   `json:"risk"`
}

type sourceMaterializationInputBinding struct {
	Source string `json:"source"`
	Target string `json:"target"`
}

type sourceMaterializeOutput struct {
	RelativePath string
	Bytes        []byte
}

type sourceMaterializeFoundationRow struct {
	SourceID string                        `json:"source_id"`
	State    string                        `json:"state"`
	Citation sourceMaterializationCitation `json:"citation"`
	Reason   string                        `json:"reason,omitempty"`
	Binding  string                        `json:"binding,omitempty"`
}

func runSourceMaterialize(args []string, stdout, stderr io.Writer) int {
	return runSourceMaterializeWithFetcher(args, stdout, stderr, nil)
}

func runSourceMaterializeWithFetcher(args []string, stdout, stderr io.Writer, fetcher sourceImportFetcher) int {
	for _, arg := range args[1:] {
		if arg == "--help" || arg == "-h" {
			logln(stdout, sourceMaterializeUsage)
			return 0
		}
	}
	opts, err := parseSourceMaterializeOptions(args[1:])
	if err != nil {
		logln(stderr, "connectorgen source-materialize:", err)
		logln(stderr, sourceMaterializeUsage)
		return 2
	}
	lock, err := loadConnectorSourceImportLock(opts.DefsDir, opts.Connector)
	if err != nil {
		logln(stderr, "connectorgen source-materialize:", err)
		return 1
	}
	if lock.SchemaVersion != 4 || lock.Materialization == nil {
		logln(stderr, "connectorgen source-materialize: source lock schema v4 with one materialization block is required; v1-v3 locks remain source-import only")
		return 1
	}
	if fetcher == nil && sourceImportLockRequiresRetainedArtifact(lock) {
		fetcher, err = newConnectorSourceImportRetainedArtifactFetcher(opts.DefsDir, opts.Connector, defaultSourceImportLimits())
		if err != nil {
			logln(stderr, "connectorgen source-materialize:", err)
			return 1
		}
	}
	if fetcher == nil {
		logln(stderr, "connectorgen source-materialize: a materialized v4 lock requires retained source artifacts")
		return 1
	}
	result, err := importSourceLockResult(context.Background(), lock, fetcher, defaultSourceImportLimits())
	if err != nil {
		logln(stderr, "connectorgen source-materialize: import retained source lock:", err)
		return 1
	}
	outputs, err := sourceMaterializeOutputs(lock, result)
	if err != nil {
		logln(stderr, "connectorgen source-materialize:", err)
		return 1
	}
	sourcesDir, err := sourceImportConnectorSourcesDir(opts.DefsDir, opts.Connector)
	if err != nil {
		logln(stderr, "connectorgen source-materialize:", err)
		return 1
	}
	bundleDir := filepath.Dir(sourcesDir)
	if err := sourceMaterializeValidateOutputPaths(bundleDir, outputs); err != nil {
		logln(stderr, "connectorgen source-materialize:", err)
		return 1
	}
	if opts.Check {
		if changed, err := sourceMaterializeHasDrift(bundleDir, outputs); err != nil {
			logln(stderr, "connectorgen source-materialize:", err)
			return 1
		} else if changed {
			logln(stderr, "connectorgen source-materialize: owned source-materialized outputs have drifted; rerun without --check")
			return 1
		}
		logf(stdout, "connectorgen source-materialize: %s verified (%d source operations)\n", opts.Connector, len(result.Operations))
		return 0
	}
	if err := sourceMaterializePublish(bundleDir, opts.Connector, outputs, stderr); err != nil {
		logln(stderr, "connectorgen source-materialize:", err)
		return 1
	}
	logf(stdout, "connectorgen source-materialize: %s materialized (%d source operations)\n", opts.Connector, len(result.Operations))
	return 0
}

func parseSourceMaterializeOptions(args []string) (sourceMaterializeOptions, error) {
	opts := sourceMaterializeOptions{DefsDir: filepath.Join("internal", "connectors", "defs")}
	defsSet := false
	for i := 0; i < len(args); i++ {
		switch arg := args[i]; arg {
		case "--check":
			if opts.Check {
				return sourceMaterializeOptions{}, fmt.Errorf("--check may be specified only once")
			}
			opts.Check = true
		case "--defs":
			if defsSet || i+1 >= len(args) || strings.HasPrefix(args[i+1], "-") || strings.TrimSpace(args[i+1]) == "" {
				return sourceMaterializeOptions{}, fmt.Errorf("--defs requires one non-empty value and may be specified only once")
			}
			i++
			opts.DefsDir = args[i]
			defsSet = true
		default:
			if strings.HasPrefix(arg, "--defs=") {
				if defsSet || strings.TrimSpace(strings.TrimPrefix(arg, "--defs=")) == "" {
					return sourceMaterializeOptions{}, fmt.Errorf("--defs requires one non-empty value and may be specified only once")
				}
				opts.DefsDir = strings.TrimPrefix(arg, "--defs=")
				defsSet = true
				continue
			}
			if strings.HasPrefix(arg, "-") {
				return sourceMaterializeOptions{}, fmt.Errorf("unknown flag %q", arg)
			}
			if opts.Connector != "" {
				return sourceMaterializeOptions{}, fmt.Errorf("only one connector may be materialized at a time")
			}
			opts.Connector = arg
		}
	}
	if opts.Connector == "" {
		return sourceMaterializeOptions{}, fmt.Errorf("a connector name is required")
	}
	if err := validateSourceImportConnector(opts.Connector); err != nil {
		return sourceMaterializeOptions{}, err
	}
	return opts, nil
}

// validateSourceMaterializationWire performs validation that is independent
// of imported provider bytes. It runs from parseSourceImportLock so callers
// never receive a nominally-v4 lock whose materialization fields were merely
// ignored by source-import.
func validateSourceMaterializationWire(lock sourceImportLock) error {
	plan := lock.Materialization
	if plan == nil {
		return fmt.Errorf("source lock v4 requires exactly one materialization block")
	}
	if err := validateSourceMaterializationMetadata(plan.Metadata); err != nil {
		return err
	}
	properties, err := sourceMaterializationConfigProperties(plan.Config)
	if err != nil {
		return err
	}
	if err := validateSourceMaterializationAuth(plan.Auth, properties); err != nil {
		return err
	}
	if err := validateSourceMaterializationServer(plan.Server, properties); err != nil {
		return err
	}
	if len(plan.Operations) == 0 {
		return fmt.Errorf("source lock v4 materialization has no per-operation accounting rows")
	}
	seen := map[string]bool{}
	for _, row := range plan.Operations {
		if strings.TrimSpace(row.SourceID) == "" || seen[row.SourceID] {
			return fmt.Errorf("source lock v4 materialization has a missing or duplicate source operation ID")
		}
		seen[row.SourceID] = true
		if strings.TrimSpace(row.Citation.DocumentID) == "" || strings.TrimSpace(row.Citation.Location) == "" {
			return fmt.Errorf("source lock v4 materialization operation %q requires an explicit document_id and location citation", row.SourceID)
		}
		switch row.State {
		case "materialized":
			if row.Binding == nil || strings.TrimSpace(row.Reason) != "" {
				return fmt.Errorf("source lock v4 materialization operation %q requires one binding and no reason", row.SourceID)
			}
			if err := validateSourceMaterializationBinding(*row.Binding); err != nil {
				return fmt.Errorf("source lock v4 materialization operation %q: %w", row.SourceID, err)
			}
		case "blocked", "unsupported":
			if row.Binding != nil || strings.TrimSpace(row.Reason) == "" {
				return fmt.Errorf("source lock v4 materialization operation %q must have a non-empty reason and no binding when %s", row.SourceID, row.State)
			}
			if len(row.Inputs) != 0 {
				return fmt.Errorf("source lock v4 materialization operation %q must not declare inputs when %s; blocked and unsupported rows are accounting-only", row.SourceID, row.State)
			}
		default:
			return fmt.Errorf("source lock v4 materialization operation %q has unsupported state %q", row.SourceID, row.State)
		}
		inputSeen := map[string]bool{}
		for _, input := range row.Inputs {
			if input.Source == "" || input.Target == "" || inputSeen[input.Source] {
				return fmt.Errorf("source lock v4 materialization operation %q has an invalid or duplicate input binding", row.SourceID)
			}
			inputSeen[input.Source] = true
		}
	}
	return nil
}

func validateSourceMaterializationMetadata(metadata sourceMaterializationMetadata) error {
	for name, value := range map[string]string{
		"display_name": metadata.DisplayName, "description": metadata.Description,
		"integration_type": metadata.IntegrationType, "release_stage": metadata.ReleaseStage,
		"docs_url": metadata.DocsURL,
	} {
		if strings.TrimSpace(value) == "" || strings.ContainsAny(value, "\r\n") {
			return fmt.Errorf("source lock v4 materialization metadata.%s must be a non-empty one-line value", name)
		}
	}
	if metadata.IntegrationType != "api" {
		return fmt.Errorf("source lock v4 materialization metadata.integration_type %q is not supported", metadata.IntegrationType)
	}
	return nil
}

func sourceMaterializationConfigProperties(config sourceMaterializationConfig) (map[string]sourceMaterializationConfigProperty, error) {
	if len(config.Properties) == 0 {
		return nil, fmt.Errorf("source lock v4 materialization config.properties must be non-empty")
	}
	properties := make(map[string]sourceMaterializationConfigProperty, len(config.Properties))
	for _, property := range config.Properties {
		if !sourceMaterializationIdentifier(property.Name) || property.Type != "string" || strings.TrimSpace(property.Description) == "" {
			return nil, fmt.Errorf("source lock v4 materialization config property is invalid")
		}
		if property.Secret && property.Default != nil {
			return nil, fmt.Errorf("source lock v4 materialization config property %q must not give a secret a default", property.Name)
		}
		if _, duplicate := properties[property.Name]; duplicate {
			return nil, fmt.Errorf("source lock v4 materialization config duplicates property %q", property.Name)
		}
		properties[property.Name] = property
	}
	return properties, nil
}

func validateSourceMaterializationAuth(auth sourceMaterializationAuth, properties map[string]sourceMaterializationConfigProperty) error {
	switch auth.Mode {
	case "none":
		if auth.Token != "" || auth.Header != "" || auth.Prefix != "" || auth.Param != "" || auth.Value != "" || auth.Username != "" || auth.Password != "" {
			return fmt.Errorf("source lock v4 materialization auth.mode none must not declare credentials")
		}
	case "bearer":
		if !sourceMaterializationSecretTemplate(auth.Token, properties) || auth.Header != "" || auth.Prefix != "" || auth.Param != "" || auth.Value != "" || auth.Username != "" || auth.Password != "" {
			return fmt.Errorf("source lock v4 materialization bearer auth requires token {{ secrets.<declared-secret> }} only")
		}
	case "api_key_header":
		if auth.Header == "" || !sourceMaterializationSecretTemplate(auth.Value, properties) || auth.Token != "" || auth.Prefix != "" || auth.Param != "" || auth.Username != "" || auth.Password != "" {
			return fmt.Errorf("source lock v4 materialization api_key_header auth requires a fixed header and {{ secrets.<declared-secret> }} value")
		}
	case "api_key_query":
		if auth.Param == "" || !sourceMaterializationSecretTemplate(auth.Value, properties) || auth.Token != "" || auth.Header != "" || auth.Prefix != "" || auth.Username != "" || auth.Password != "" {
			return fmt.Errorf("source lock v4 materialization api_key_query auth requires a fixed parameter and {{ secrets.<declared-secret> }} value")
		}
	default:
		return fmt.Errorf("source lock v4 materialization auth.mode %q is not supported", auth.Mode)
	}
	return nil
}

func sourceMaterializationSecretTemplate(value string, properties map[string]sourceMaterializationConfigProperty) bool {
	const prefix = "{{ secrets."
	if !strings.HasPrefix(value, prefix) || !strings.HasSuffix(value, " }}") {
		return false
	}
	name := strings.TrimSuffix(strings.TrimPrefix(value, prefix), " }}")
	property, ok := properties[name]
	return ok && property.Secret
}

func validateSourceMaterializationServer(server sourceMaterializationServer, properties map[string]sourceMaterializationConfigProperty) error {
	if strings.TrimSpace(server.UserAgent) == "" || strings.ContainsAny(server.UserAgent, "\r\n") {
		return fmt.Errorf("source lock v4 materialization server.user_agent must be explicit")
	}
	const prefix = "{{ config."
	if !strings.HasPrefix(server.URL, prefix) || !strings.HasSuffix(server.URL, " }}") {
		return fmt.Errorf("source lock v4 materialization server.url must be exactly one config template")
	}
	name := strings.TrimSuffix(strings.TrimPrefix(server.URL, prefix), " }}")
	property, ok := properties[name]
	if !ok || property.Secret || property.Default == nil || property.Type != "string" {
		return fmt.Errorf("source lock v4 materialization server.url must reference a declared non-secret defaulted config property")
	}
	return nil
}

func validateSourceMaterializationBinding(binding sourceMaterializationOperationBind) error {
	if !sourceMaterializationIdentifier(binding.ID) || len(binding.SuccessStatuses) == 0 || strings.TrimSpace(binding.Risk) == "" {
		return fmt.Errorf("binding requires a valid id, risk, and success_statuses")
	}
	switch binding.Kind {
	case "direct_read":
		if binding.CommandPath == "" || binding.CommandSummary == "" || binding.MaxResponseBytes <= 0 || binding.OutputPolicy == "" || binding.RequestMedia != "" || binding.WriteKind != "" || binding.MutationClass != "" || binding.Approval != "" {
			return fmt.Errorf("direct_read binding requires command_path, command_summary, output_policy, and positive max_response_bytes only")
		}
	case "write":
		if binding.CommandPath != "" || binding.CommandSummary != "" || binding.MaxResponseBytes != 0 || binding.OutputPolicy != "" || binding.RequestMedia == "" || binding.WriteKind == "" || binding.MutationClass != binding.WriteKind || (binding.Approval != "none" && binding.Approval != "destructive") {
			return fmt.Errorf("write binding requires request_media, matching write_kind/mutation_class, and approval (none or destructive) only")
		}
		if binding.WriteKind != "create" && binding.WriteKind != "update" && binding.WriteKind != "upsert" && binding.WriteKind != "delete" && binding.WriteKind != "custom" {
			return fmt.Errorf("write binding has unsupported write_kind %q", binding.WriteKind)
		}
	default:
		return fmt.Errorf("binding kind %q is not supported", binding.Kind)
	}
	return nil
}

// sourceMaterializationIdentifier is intentionally separate from connector
// names: config keys and action IDs use the established snake_case bundle
// vocabulary, while connector directory names remain lowercase-dash only.
func sourceMaterializationIdentifier(value string) bool {
	if value == "" || value[0] < 'a' || value[0] > 'z' {
		return false
	}
	for _, character := range value[1:] {
		if (character < 'a' || character > 'z') && (character < '0' || character > '9') && character != '_' {
			return false
		}
	}
	return true
}

func sourceMaterializeOutputs(lock sourceImportLock, result sourceImportResult) ([]sourceMaterializeOutput, error) {
	if lock.Materialization == nil {
		return nil, errors.New("source lock v4 materialization is absent")
	}
	plan := *lock.Materialization
	byID := make(map[string]sourceOperationDescriptor, len(result.Operations))
	for _, operation := range result.Operations {
		byID[operation.SourceID] = operation
	}
	if len(byID) != len(plan.Operations) {
		return nil, fmt.Errorf("source lock v4 materialization accounts for %d operations but retained source imports %d", len(plan.Operations), len(byID))
	}
	rows := make(map[string]sourceMaterializationOperation, len(plan.Operations))
	for _, row := range plan.Operations {
		rows[row.SourceID] = row
	}
	properties, err := sourceMaterializationConfigProperties(plan.Config)
	if err != nil {
		return nil, err
	}
	if err := sourceMaterializeValidateCheck(plan.Check, byID, rows); err != nil {
		return nil, err
	}
	metadata, spec, streams := sourceMaterializeBaseDocuments(lock.Connector, plan, properties)
	operations := []any{}
	writes := []any{}
	endpoints := []any{}
	commands := []any{}
	foundation := make([]sourceMaterializeFoundationRow, 0, len(plan.Operations))
	seenBindings := map[string]bool{}
	for _, row := range plan.Operations {
		source, found := byID[row.SourceID]
		if !found {
			return nil, fmt.Errorf("source lock v4 materialization references unknown source operation %q", row.SourceID)
		}
		if source.Source.DocumentID != row.Citation.DocumentID || source.Source.Location != row.Citation.Location {
			return nil, fmt.Errorf("source lock v4 materialization citation for %q does not match the retained source descriptor", row.SourceID)
		}
		foundationRow := sourceMaterializeFoundationRow{SourceID: row.SourceID, State: row.State, Citation: row.Citation, Reason: row.Reason}
		switch row.State {
		case "materialized":
			if row.Binding == nil {
				return nil, fmt.Errorf("source operation %q is materialized without a binding", row.SourceID)
			}
			if err := sourceMaterializeValidateInputBindings(source, row); err != nil {
				return nil, err
			}
			binding := *row.Binding
			if seenBindings[binding.ID] {
				return nil, fmt.Errorf("source lock v4 materialization duplicates binding id %q", binding.ID)
			}
			seenBindings[binding.ID] = true
			foundationRow.Binding = binding.Kind + ":" + binding.ID
			switch binding.Kind {
			case "direct_read":
				op, command, endpoint, err := sourceMaterializeDirectRead(source, row, binding)
				if err != nil {
					return nil, err
				}
				operations = append(operations, op)
				commands = append(commands, command)
				endpoints = append(endpoints, endpoint)
			case "write":
				action, endpoint, err := sourceMaterializeWrite(source, row, binding)
				if err != nil {
					return nil, err
				}
				writes = append(writes, action)
				endpoints = append(endpoints, endpoint)
			}
		case "blocked", "unsupported":
			endpoints = append(endpoints, sourceMaterializeBlockedEndpoint(source, row))
		default:
			return nil, fmt.Errorf("source operation %q has unsupported accounting state %q", row.SourceID, row.State)
		}
		foundation = append(foundation, foundationRow)
	}
	if len(byID) != len(foundation) {
		return nil, fmt.Errorf("source lock v4 materialization does not account for every source operation")
	}
	descriptor, err := marshalSourceImportResult(result)
	if err != nil {
		return nil, fmt.Errorf("encode source descriptor: %w", err)
	}
	outputs := []sourceMaterializeOutput{
		{RelativePath: "metadata.json", Bytes: sourceMaterializeJSON(metadata)},
		{RelativePath: "spec.json", Bytes: sourceMaterializeJSON(spec)},
		{RelativePath: "streams.json", Bytes: sourceMaterializeJSON(streams)},
		{RelativePath: "writes.json", Bytes: sourceMaterializeJSON(map[string]any{"actions": writes})},
		{RelativePath: "operations.json", Bytes: sourceMaterializeJSON(map[string]any{"operations": operations})},
		{RelativePath: "api_surface.json", Bytes: sourceMaterializeJSON(map[string]any{"operation_ledger_version": 1, "api": plan.Metadata.DisplayName + " source-lock API", "docs": plan.Metadata.DocsURL, "scope": "Generated only from source lock v4 explicit materialization bindings.", "endpoints": endpoints})},
		{RelativePath: "cli_surface.json", Bytes: sourceMaterializeJSON(map[string]any{"tagline": plan.Metadata.Description, "usage": lock.Connector + " <command>", "commands": commands})},
		{RelativePath: "docs.md", Bytes: sourceMaterializeDocs(lock.Connector, plan, foundation)},
		{RelativePath: filepath.ToSlash(filepath.Join("sources", lock.Connector+"-operation-descriptor.json")), Bytes: descriptor},
		{RelativePath: "missing-foundation.json", Bytes: sourceMaterializeJSON(map[string]any{"schema_version": 1, "connector": lock.Connector, "source_lock": filepath.ToSlash(filepath.Join("sources", lock.Connector+"-operation-source-lock.json")), "operations": foundation})},
	}
	return outputs, nil
}

func sourceMaterializeBaseDocuments(connector string, plan sourceMaterialization, properties map[string]sourceMaterializationConfigProperty) (map[string]any, map[string]any, map[string]any) {
	configProperties := map[string]any{}
	required := []any{}
	names := make([]string, 0, len(properties))
	for name := range properties {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		property := properties[name]
		schema := map[string]any{"type": property.Type, "description": property.Description}
		if property.Format != "" {
			schema["format"] = property.Format
		}
		if property.Default != nil {
			schema["default"] = *property.Default
		}
		if property.Secret {
			schema["x-secret"] = true
		}
		if property.Required {
			required = append(required, name)
		}
		configProperties[name] = schema
	}
	// engine's connection-spec dialect intentionally keeps the accepted
	// JSON-Schema subset small; it does not admit additionalProperties here.
	// The v4 plan itself remains closed by strict decoding and by emitting only
	// its declared config keys.
	spec := map[string]any{"$schema": "http://json-schema.org/draft-07/schema#", "title": plan.Metadata.DisplayName + " Connection Specification", "type": "object", "properties": configProperties}
	if len(required) > 0 {
		spec["required"] = required
	}
	capabilities := map[string]bool{"check": true, "read": false, "write": false, "query": false, "cdc": false, "dynamic_schema": false}
	for _, row := range plan.Operations {
		if row.State != "materialized" || row.Binding == nil {
			continue
		}
		if row.Binding.Kind == "direct_read" {
			capabilities["read"] = true
		}
		if row.Binding.Kind == "write" {
			capabilities["write"] = true
		}
	}
	metadata := map[string]any{"name": connector, "display_name": plan.Metadata.DisplayName, "description": plan.Metadata.Description, "integration_type": plan.Metadata.IntegrationType, "docs_url": plan.Metadata.DocsURL, "release_stage": plan.Metadata.ReleaseStage, "capabilities": capabilities}
	base := map[string]any{"url": plan.Server.URL, "user_agent": plan.Server.UserAgent, "pagination": map[string]any{"type": "none"}, "check": map[string]any{"method": plan.Check.Method, "path": plan.Check.Path, "success_statuses": plan.Check.SuccessStatuses}}
	if plan.Auth.Mode != "none" {
		base["auth"] = []any{sourceMaterializeAuthDocument(plan.Auth)}
	}
	return metadata, spec, map[string]any{"base": base, "streams": []any{}}
}

func sourceMaterializeAuthDocument(auth sourceMaterializationAuth) map[string]any {
	result := map[string]any{"mode": auth.Mode}
	if auth.Token != "" {
		result["token"] = auth.Token
	}
	if auth.Header != "" {
		result["header"] = auth.Header
	}
	if auth.Prefix != "" {
		result["prefix"] = auth.Prefix
	}
	if auth.Param != "" {
		result["param"] = auth.Param
	}
	if auth.Value != "" {
		result["value"] = auth.Value
	}
	if auth.Username != "" {
		result["username"] = auth.Username
	}
	if auth.Password != "" {
		result["password"] = auth.Password
	}
	return result
}

func sourceMaterializeValidateCheck(check sourceMaterializationCheck, sources map[string]sourceOperationDescriptor, rows map[string]sourceMaterializationOperation) error {
	source, found := sources[check.SourceID]
	row, selected := rows[check.SourceID]
	if !found || !selected || strings.ToUpper(source.Method) != "GET" || check.Method != "GET" || check.Path != source.Path || len(check.SuccessStatuses) == 0 {
		return fmt.Errorf("source lock v4 materialization check must select one exact retained GET operation with explicit success statuses")
	}
	if row.State != "materialized" || row.Binding == nil || row.Binding.Kind != "direct_read" {
		return fmt.Errorf("source lock v4 materialization check %q must select a materialized direct_read operation", check.SourceID)
	}
	if err := sourceMaterializeValidateMaterializedRuntime(source, *row.Binding); err != nil {
		return fmt.Errorf("source lock v4 materialization check %q is not runtime-admissible: %w", check.SourceID, err)
	}
	if len(source.Request.Path) != 0 || len(source.Request.Query) != 0 || len(source.Request.Header) != 0 || source.Request.Body != nil || len(source.Request.Media) != 0 {
		return fmt.Errorf("source lock v4 materialization check %q has caller-controlled inputs and cannot be an implicit default", check.SourceID)
	}
	return sourceMaterializeValidateSuccessStatuses(source, check.SuccessStatuses)
}

// sourceMaterializeValidateMaterializedRuntime accepts only a descriptor that
// the generated bundle can express without guessing a provider transport,
// scope, page walk, or byte policy. A source row with one of these contracts
// remains valid as blocked/unsupported accounting, but cannot be executable.
func sourceMaterializeValidateMaterializedRuntime(source sourceOperationDescriptor, binding sourceMaterializationOperationBind) error {
	if source.Runtime.MergeBlocked || len(source.Runtime.Gaps) != 0 || source.Runtime.NonExecutableMutation != nil || source.Runtime.PartialCoverageMutation != nil {
		return fmt.Errorf("source operation %q runtime reachability is blocked; mark it blocked or unsupported with its retained mapping reason", source.SourceID)
	}
	if source.AuthScopes.Declared {
		return fmt.Errorf("source operation %q has a declared authentication scope that source-materialize cannot faithfully bind", source.SourceID)
	}
	if source.Pagination != nil {
		return fmt.Errorf("source operation %q has declared pagination that source-materialize cannot faithfully bind", source.SourceID)
	}
	if source.Servers.Root.Declared || source.Servers.PathItem.Declared || source.Servers.Operation.Declared || source.Servers.Swagger != nil || len(source.Servers.Gaps) != 0 {
		return fmt.Errorf("source operation %q has declared server routing that source-materialize cannot faithfully bind", source.SourceID)
	}
	switch binding.Kind {
	case "direct_read":
		if source.ByteLimits.Request != 0 {
			return fmt.Errorf("source operation %q has a declared request byte limit that source-materialize cannot faithfully bind", source.SourceID)
		}
		if source.ByteLimits.Response > 0 && source.ByteLimits.Response != int64(binding.MaxResponseBytes) {
			return fmt.Errorf("source operation %q response byte limit %d does not exactly match direct_read max_response_bytes %d", source.SourceID, source.ByteLimits.Response, binding.MaxResponseBytes)
		}
	case "write":
		if source.ByteLimits.Request != 0 || source.ByteLimits.Response != 0 {
			return fmt.Errorf("source operation %q has declared byte limits that source-materialize cannot faithfully bind for a write", source.SourceID)
		}
	default:
		return fmt.Errorf("source operation %q has unsupported runtime binding kind %q", source.SourceID, binding.Kind)
	}
	return nil
}

func sourceMaterializeDirectRead(source sourceOperationDescriptor, row sourceMaterializationOperation, binding sourceMaterializationOperationBind) (map[string]any, map[string]any, map[string]any, error) {
	if strings.ToUpper(source.Method) != "GET" || source.Protocol != "rest" || source.Output.Class != sourceOutputJSON || source.Request.Body != nil || len(source.Request.Media) != 0 || binding.OutputPolicy != "json_redacted" {
		return nil, nil, nil, fmt.Errorf("source operation %q is not a bounded JSON GET direct-read contract", source.SourceID)
	}
	if err := sourceMaterializeValidateMaterializedRuntime(source, binding); err != nil {
		return nil, nil, nil, err
	}
	if err := sourceMaterializeValidateSuccessStatuses(source, binding.SuccessStatuses); err != nil {
		return nil, nil, nil, err
	}
	if len(source.Request.Query) != 0 {
		return nil, nil, nil, fmt.Errorf("source operation %q has query wire semantics that source-materialize cannot faithfully preserve; mark it blocked or unsupported with an explicit reason", source.SourceID)
	}
	if len(row.Inputs) != 0 {
		return nil, nil, nil, fmt.Errorf("source operation %q direct-read inputs require query wire semantics that source-materialize cannot faithfully preserve", source.SourceID)
	}
	if len(source.Request.Path) != 0 || len(source.Request.Header) != 0 {
		return nil, nil, nil, fmt.Errorf("source operation %q needs a path/header foundation not declared by source-materialize", source.SourceID)
	}
	rest := map[string]any{"method": "GET", "path": source.Path, "max_bytes": binding.MaxResponseBytes, "response": map[string]any{"success_statuses": binding.SuccessStatuses}}
	op := map[string]any{"id": binding.ID, "kind": "rest_read", "summary": binding.CommandSummary, "source_url": sourceMaterializeCitationURL(source), "risk": binding.Risk, "approval": "none", "output_policy": binding.OutputPolicy, "source_operation": map[string]any{"id": source.SourceID, "method": "GET", "path": source.Path}, "rest": rest}
	// A rest_read operation becomes executable through its implemented direct
	// read CLI surface; api_surface's operations ledger is reserved for fixed
	// GraphQL operations by the existing validator.
	endpoint := map[string]any{"method": "GET", "path": source.Path, "covered_by": map[string]any{"direct_read": binding.CommandPath}}
	command := map[string]any{"path": binding.CommandPath, "summary": binding.CommandSummary, "intent": "direct_read", "availability": "implemented", "operation": binding.ID, "source_operation": source.SourceID, "source_url": sourceMaterializeCitationURL(source), "api_surface": []any{map[string]any{"method": "GET", "path": source.Path}}, "output_policy": binding.OutputPolicy, "risk": binding.Risk, "approval": "none"}
	return op, command, endpoint, nil
}

func sourceMaterializeWrite(source sourceOperationDescriptor, row sourceMaterializationOperation, binding sourceMaterializationOperationBind) (map[string]any, map[string]any, error) {
	if source.Protocol != "rest" || !sourceMaterializeMutationMethod(source.Method) || source.Request.Body == nil {
		return nil, nil, fmt.Errorf("source operation %q is not the selected JSON mutation contract", source.SourceID)
	}
	if source.RequestMediaType() != "application/json" || binding.RequestMedia != "application/json" {
		return nil, nil, fmt.Errorf("source operation %q is not the selected JSON mutation contract: generated body_type json sends application/json, so source and binding media must be exactly application/json", source.SourceID)
	}
	if err := sourceMaterializeValidateMaterializedRuntime(source, binding); err != nil {
		return nil, nil, err
	}
	if err := sourceMaterializeValidateSuccessStatuses(source, binding.SuccessStatuses); err != nil {
		return nil, nil, err
	}
	if len(source.Request.Path) != 0 || len(source.Request.Query) != 0 || len(source.Request.Header) != 0 {
		return nil, nil, fmt.Errorf("source operation %q needs path/query/header write support not declared by source-materialize", source.SourceID)
	}
	schema, ok := source.Request.Body.Schema.(map[string]any)
	if !ok || schema["type"] != "object" {
		return nil, nil, fmt.Errorf("source operation %q JSON request body must be an object", source.SourceID)
	}
	if additional, exists := schema["additionalProperties"]; !exists || additional != false {
		return nil, nil, fmt.Errorf("source operation %q JSON request body does not declare additionalProperties: false; runtime cannot faithfully materialize an open provider object, so mark it blocked or unsupported with an explicit reason", source.SourceID)
	}
	if err := sourceMaterializeValidateBodyInputs(source, row); err != nil {
		return nil, nil, err
	}
	for _, input := range row.Inputs {
		field, isBody := strings.CutPrefix(input.Source, "body.")
		if !isBody || field == "" || input.Target != "record."+field {
			return nil, nil, fmt.Errorf("source operation %q write inputs support only exact body.<field> -> record.<field> bindings", source.SourceID)
		}
	}
	statuses, err := sourceMaterializeStatusCodes(binding.SuccessStatuses)
	if err != nil {
		return nil, nil, err
	}
	action := map[string]any{"name": binding.ID, "kind": binding.WriteKind, "method": strings.ToUpper(source.Method), "path": strings.TrimPrefix(source.Path, "/"), "body_type": "json", "body_required": source.Request.Body.Required, "record_schema": schema, "risk": binding.Risk, "success_statuses": statuses}
	if binding.Approval == "destructive" {
		action["confirmation"] = map[string]any{"kind": "destructive"}
	}
	endpoint := map[string]any{"method": strings.ToUpper(source.Method), "path": source.Path, "covered_by": map[string]any{"write": binding.ID}}
	return action, endpoint, nil
}

func (source sourceOperationDescriptor) RequestMediaType() string {
	if source.Request.MediaType != "" {
		return source.Request.MediaType
	}
	return ""
}

func sourceMaterializeBlockedEndpoint(source sourceOperationDescriptor, row sourceMaterializationOperation) map[string]any {
	model, risk := "direct_read", "low"
	if sourceMaterializeMutationMethod(source.Method) {
		model, risk = "admin_reverse_etl", "high"
	}
	if row.State == "unsupported" {
		model = "disallowed"
	}
	return map[string]any{"method": strings.ToUpper(source.Method), "path": source.Path, "operation": map[string]any{"model": model, "status": "blocked", "risk": risk, "blocked_by_default": true, "reason": row.Reason, "source_url": sourceMaterializeCitationURL(source), "notes": "source_operation=" + source.SourceID}}
}

func sourceMaterializeValidateInputBindings(source sourceOperationDescriptor, row sourceMaterializationOperation) error {
	bound := map[string]string{}
	declared := map[string]bool{}
	for _, group := range []struct {
		location string
		values   []sourceParameterDescriptor
	}{{"path", source.Request.Path}, {"query", source.Request.Query}, {"header", source.Request.Header}} {
		for _, parameter := range group.values {
			declared[group.location+"."+parameter.Name] = true
		}
	}
	if source.Request.Body != nil {
		schema, ok := source.Request.Body.Schema.(map[string]any)
		if !ok {
			return fmt.Errorf("source operation %q request body has no object schema for typed input bindings", source.SourceID)
		}
		properties, _ := schema["properties"].(map[string]any)
		for name := range properties {
			declared["body."+name] = true
		}
	}
	for _, input := range row.Inputs {
		if input.Source == "" || input.Source != strings.TrimSpace(input.Source) || input.Target == "" || input.Target != strings.TrimSpace(input.Target) {
			return fmt.Errorf("source operation %q has an invalid input binding", source.SourceID)
		}
		if _, duplicate := bound[input.Source]; duplicate {
			return fmt.Errorf("source operation %q duplicates input binding %q", source.SourceID, input.Source)
		}
		// A request-content union has no selected Body descriptor. Keep its
		// body inputs pending for sourceMaterializeWrite, which rejects the
		// unselected media union with the precise actionable error rather than
		// incorrectly treating a documented field as absent.
		if !declared[input.Source] && !(strings.HasPrefix(input.Source, "body.") && len(source.Request.Media) > 0) {
			return fmt.Errorf("source operation %q binds undeclared source input %q", source.SourceID, input.Source)
		}
		bound[input.Source] = input.Target
	}
	for _, group := range []struct {
		location string
		values   []sourceParameterDescriptor
	}{{"path", source.Request.Path}, {"query", source.Request.Query}, {"header", source.Request.Header}} {
		for _, parameter := range group.values {
			key := group.location + "." + parameter.Name
			if parameter.Required && bound[key] == "" {
				return fmt.Errorf("source operation %q has required input %q without an explicit binding", source.SourceID, key)
			}
		}
	}
	if source.Request.Body != nil && source.Request.Body.Required {
		if err := sourceMaterializeValidateBodyInputs(source, row); err != nil {
			return err
		}
	}
	return nil
}

func sourceMaterializeValidateBodyInputs(source sourceOperationDescriptor, row sourceMaterializationOperation) error {
	if source.Request.Body == nil {
		return nil
	}
	schema, ok := source.Request.Body.Schema.(map[string]any)
	if !ok {
		return fmt.Errorf("source operation %q request body has no object schema for typed input bindings", source.SourceID)
	}
	required, _ := schema["required"].([]any)
	bound := map[string]string{}
	for _, input := range row.Inputs {
		bound[input.Source] = input.Target
	}
	for _, raw := range required {
		name, ok := raw.(string)
		if !ok {
			return fmt.Errorf("source operation %q request body required field is not a string", source.SourceID)
		}
		if bound["body."+name] == "" {
			return fmt.Errorf("source operation %q has required input %q without an explicit binding", source.SourceID, "body."+name)
		}
	}
	return nil
}

func sourceMaterializeValidateSuccessStatuses(source sourceOperationDescriptor, statuses []string) error {
	available := map[string]bool{}
	for _, response := range source.Responses {
		available[response.Status] = true
	}
	seen := map[string]bool{}
	for _, status := range statuses {
		if seen[status] || !available[status] || len(status) != 3 || status[0] != '2' {
			return fmt.Errorf("source operation %q binding names unsupported success status %q", source.SourceID, status)
		}
		seen[status] = true
	}
	return nil
}

func sourceMaterializeStatusCodes(statuses []string) ([]any, error) {
	result := make([]any, 0, len(statuses))
	for _, status := range statuses {
		code, err := strconv.Atoi(status)
		if err != nil || code < 200 || code > 299 {
			return nil, fmt.Errorf("materialized write has invalid success status %q", status)
		}
		result = append(result, code)
	}
	return result, nil
}

func sourceMaterializeMutationMethod(method string) bool {
	switch strings.ToUpper(method) {
	case "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func sourceMaterializeCitationURL(source sourceOperationDescriptor) string {
	if source.Source.CitationURL != "" {
		return source.Source.CitationURL
	}
	if source.Source.PublishedURL != "" {
		return source.Source.PublishedURL
	}
	return source.Source.URL
}

func sourceMaterializeJSON(value any) []byte {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		panic(err)
	}
	return append(raw, '\n')
}

func sourceMaterializeDocs(connector string, plan sourceMaterialization, rows []sourceMaterializeFoundationRow) []byte {
	var out strings.Builder
	fmt.Fprintf(&out, "# Overview\n\n%s\n\n## Auth setup\n\nAuthentication mode: `%s`.\n\n## Streams notes\n\nThis source-lock v4 materialization declares no stream, transport, or event contract unless a later explicit materialization section adds one.\n\n## Write actions & risks\n\n", plan.Metadata.Description, plan.Auth.Mode)
	for _, row := range rows {
		fmt.Fprintf(&out, "- `%s`: %s.\n", row.SourceID, row.State)
	}
	fmt.Fprintf(&out, "\n## Known limits\n\nGenerated from `%s`; only explicit typed bindings are operational.\n", connector+"-operation-source-lock.json")
	return []byte(out.String())
}

func sourceMaterializeValidateOutputPaths(bundleDir string, outputs []sourceMaterializeOutput) error {
	seen := map[string]bool{}
	for _, output := range outputs {
		if output.RelativePath == "" || filepath.IsAbs(output.RelativePath) || !sourceImportPathWithin(bundleDir, filepath.Join(bundleDir, output.RelativePath)) || seen[output.RelativePath] {
			return fmt.Errorf("owned output path %q escapes or duplicates the bundle", output.RelativePath)
		}
		seen[output.RelativePath] = true
		if err := sourceMaterializeNoSymlinkPath(bundleDir, output.RelativePath); err != nil {
			return err
		}
	}
	return nil
}

func sourceMaterializeNoSymlinkPath(bundleDir, relative string) error {
	current := bundleDir
	for _, segment := range strings.Split(filepath.Clean(relative), string(filepath.Separator)) {
		current = filepath.Join(current, segment)
		info, err := os.Lstat(current)
		if os.IsNotExist(err) {
			continue
		}
		if err != nil {
			return fmt.Errorf("inspect owned output %q: %w", relative, err)
		}
		if info.Mode()&os.ModeSymlink != 0 {
			return fmt.Errorf("owned output path %q must not traverse a symlink", relative)
		}
	}
	return nil
}

func sourceMaterializeHasDrift(bundleDir string, outputs []sourceMaterializeOutput) (bool, error) {
	for _, output := range outputs {
		raw, err := os.ReadFile(filepath.Join(bundleDir, output.RelativePath))
		if err != nil {
			if os.IsNotExist(err) {
				return true, nil
			}
			return false, fmt.Errorf("read owned output %q: %w", output.RelativePath, err)
		}
		if !bytes.Equal(raw, output.Bytes) {
			return true, nil
		}
	}
	return false, nil
}

type sourceMaterializePublishOps struct {
	Rename    func(oldPath, newPath string) error
	Remove    func(path string) error
	RemoveAll func(path string) error
	Warn      func(message string)
}

func sourceMaterializeDefaultPublishOps(warnings io.Writer) sourceMaterializePublishOps {
	return sourceMaterializePublishOps{
		Rename:    os.Rename,
		Remove:    os.Remove,
		RemoveAll: os.RemoveAll,
		Warn: func(message string) {
			if warnings != nil {
				logln(warnings, "connectorgen source-materialize:", message)
			}
		},
	}
}

func sourceMaterializePublishWarning(ops sourceMaterializePublishOps, format string, args ...any) {
	if ops.Warn != nil {
		ops.Warn(fmt.Sprintf(format, args...))
	}
}

// sourceMaterializePublish installs a fully validated staged bundle through a
// same-parent two-rename transaction. Filesystems do not generally provide an
// atomic replacement for a non-empty directory, so this is deliberately a
// rollback-protected publish, not a crash-atomic directory swap: an install
// rename error restores the prior bundle. If the rollback rename itself fails,
// both recovery paths are retained and named in the error for manual repair.
func sourceMaterializePublish(bundleDir, connector string, outputs []sourceMaterializeOutput, warnings io.Writer) error {
	return sourceMaterializePublishWithOps(bundleDir, connector, outputs, sourceMaterializeDefaultPublishOps(warnings))
}

func sourceMaterializePublishWithOps(bundleDir, connector string, outputs []sourceMaterializeOutput, ops sourceMaterializePublishOps) (err error) {
	parent := filepath.Dir(bundleDir)
	stageRoot, err := os.MkdirTemp(parent, "."+connector+"-source-materialize-")
	if err != nil {
		return fmt.Errorf("create staging directory: %w", err)
	}
	preserveRecoveryArtifacts := false
	rolledBack := false
	defer func() {
		if !preserveRecoveryArtifacts {
			if cleanupErr := ops.RemoveAll(stageRoot); cleanupErr != nil {
				phase := "after successful installation"
				if err != nil {
					phase = "after failed publish"
					if rolledBack {
						phase = "after successful rollback"
					}
				}
				sourceMaterializePublishWarning(ops, "warning: staging cleanup %q %s failed: %v", stageRoot, phase, cleanupErr)
			}
		}
	}()
	stageBundle := filepath.Join(stageRoot, connector)
	if err := sourceMaterializeCopyTree(bundleDir, stageBundle); err != nil {
		return err
	}
	for _, output := range outputs {
		path := filepath.Join(stageBundle, output.RelativePath)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			return fmt.Errorf("create staged output parent: %w", err)
		}
		if err := os.WriteFile(path, output.Bytes, 0o644); err != nil {
			return fmt.Errorf("write staged output %q: %w", output.RelativePath, err)
		}
	}
	if err := sourceMaterializeValidateStaged(stageRoot, connector); err != nil {
		return err
	}
	backup, err := os.MkdirTemp(parent, "."+connector+"-source-materialize-backup-")
	if err != nil {
		return fmt.Errorf("reserve publish backup: %w", err)
	}
	if removeErr := ops.Remove(backup); removeErr != nil {
		if cleanupErr := ops.RemoveAll(backup); cleanupErr != nil {
			return fmt.Errorf("prepare publish backup: %v; cleanup reserved backup %q failed: %v", removeErr, backup, cleanupErr)
		}
		return fmt.Errorf("prepare publish backup: %w", removeErr)
	}
	if err := ops.Rename(bundleDir, backup); err != nil {
		return fmt.Errorf("prepare recoverable publish: %w", err)
	}
	if err := ops.Rename(stageBundle, bundleDir); err != nil {
		if restoreErr := ops.Rename(backup, bundleDir); restoreErr != nil {
			preserveRecoveryArtifacts = true
			return fmt.Errorf("publish staged bundle: %w; rollback failed: %v; previous bundle retained at %q and staged bundle retained at %q for manual recovery", err, restoreErr, backup, stageBundle)
		}
		rolledBack = true
		return fmt.Errorf("publish staged bundle: %w", err)
	}
	// The live bundle is installed at this point. Backup cleanup is best effort:
	// returning it as an error would falsely claim that a successful publish
	// failed, while the hidden sibling backup remains recoverable for cleanup.
	if cleanupErr := ops.RemoveAll(backup); cleanupErr != nil {
		sourceMaterializePublishWarning(ops, "warning: backup cleanup %q after successful installation failed: %v", backup, cleanupErr)
	}
	return nil
}

func sourceMaterializeCopyTree(source, destination string) error {
	return filepath.WalkDir(source, func(path string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		relative, err := filepath.Rel(source, path)
		if err != nil {
			return err
		}
		target := filepath.Join(destination, relative)
		if !sourceImportPathWithin(destination, target) {
			return fmt.Errorf("staging copy path escapes destination")
		}
		if entry.Type()&os.ModeSymlink != 0 {
			return fmt.Errorf("source bundle contains symlink %q", relative)
		}
		if entry.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("source bundle contains non-regular file %q", relative)
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, raw, 0o644)
	})
}

func sourceMaterializeValidateStaged(stageRoot, connector string) error {
	fsys := os.DirFS(stageRoot)
	bundle, err := engine.Load(fsys, connector)
	if err != nil {
		return fmt.Errorf("load staged bundle: %w", err)
	}
	findings := []Finding{}
	findings = append(findings, checkName(bundle)...)
	findings = append(findings, checkInterpolations(bundle)...)
	findings = append(findings, checkSchemaRefs(fsys, bundle)...)
	findings = append(findings, checkWritePathFields(bundle)...)
	findings = append(findings, checkAPISurface(bundle)...)
	findings = append(findings, checkCLISurface(bundle)...)
	findings = append(findings, checkDocsHeadings(bundle)...)
	if len(findings) > 0 {
		return fmt.Errorf("staged bundle checks failed: %s: %s", findings[0].File, findings[0].Message)
	}
	return nil
}

// validateOperationalContractPath is an opt-in validate gate. It deliberately
// reads a single bundle only; it neither changes registration nor upgrades a
// connector's capabilities. "declared" means every currently declared
// materialization capability (check/read/write) has its closed contract.
func validateOperationalContractPath(dir, connector, profile string) ([]Finding, error) {
	cleanDir := filepath.Clean(dir)
	if !isBundleDir(cleanDir) {
		return nil, fmt.Errorf("operational contract requires one connector bundle directory")
	}
	if connector == "" {
		connector = filepath.Base(cleanDir)
	} else if filepath.Base(cleanDir) != connector {
		return nil, sourceMaterializeBundleTargetMismatchError{Connector: connector, Target: filepath.Base(cleanDir)}
	}
	bundle, err := engine.Load(os.DirFS(filepath.Dir(cleanDir)), connector)
	if err != nil {
		return nil, fmt.Errorf("load operational-contract bundle: %w", err)
	}
	if bundle.Metadata.IntegrationType != "api" {
		return nil, fmt.Errorf("operational contract requires metadata.integration_type \"api\", got %q", bundle.Metadata.IntegrationType)
	}
	wanted, err := sourceMaterializeOperationalProfile(profile, bundle)
	if err != nil {
		return nil, err
	}
	findings := []Finding{}
	for _, capability := range wanted {
		if message := sourceMaterializeOperationalCapability(bundle, capability); message != "" {
			findings = append(findings, Finding{Connector: bundle.Name, File: "metadata.json", Rule: "operational_contract", Message: message})
		}
	}
	return findings, nil
}

// sourceMaterializeBundleTargetMismatchError makes connector selection
// fail-closed whenever a caller passes a concrete bundle directory for a
// different connector. It prevents the operational gate from silently
// inspecting a sibling bundle under the same definitions root.
type sourceMaterializeBundleTargetMismatchError struct {
	Connector string
	Target    string
}

func (err sourceMaterializeBundleTargetMismatchError) Error() string {
	return fmt.Sprintf("--connector %q does not match bundle target %q", err.Connector, err.Target)
}

func sourceMaterializeOperationalProfile(profile string, bundle engine.Bundle) ([]string, error) {
	if profile == "declared" {
		capabilities := []string{}
		if bundle.Metadata.Capabilities.Check {
			capabilities = append(capabilities, "check")
		}
		if bundle.Metadata.Capabilities.Read {
			capabilities = append(capabilities, "read")
		}
		if bundle.Metadata.Capabilities.Write {
			capabilities = append(capabilities, "write")
		}
		if bundle.Metadata.Capabilities.Query || bundle.Metadata.Capabilities.CDC || bundle.Metadata.Capabilities.DynamicSchema {
			return nil, fmt.Errorf("operational contract declared profile supports only check/read/write; connector declares an unsupported capability")
		}
		return capabilities, nil
	}
	parts := strings.Split(profile, ",")
	if len(parts) == 0 {
		return nil, fmt.Errorf("operational contract profile is empty")
	}
	seen := map[string]bool{}
	capabilities := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "check" && part != "read" && part != "write" || seen[part] {
			return nil, fmt.Errorf("operational contract profile must be check, read, write, a comma-separated unique combination, or declared")
		}
		seen[part] = true
		capabilities = append(capabilities, part)
	}
	sort.Strings(capabilities)
	return capabilities, nil
}

func sourceMaterializeOperationalCapability(bundle engine.Bundle, capability string) string {
	switch capability {
	case "check":
		if !bundle.Metadata.Capabilities.Check {
			return "operational contract requires check but metadata.capabilities.check is false"
		}
		if bundle.HTTP.Check == nil {
			return "operational contract requires a declared base.check request"
		}
	case "read":
		if !bundle.Metadata.Capabilities.Read {
			return "operational contract requires read but metadata.capabilities.read is false"
		}
		if bundle.CLISurface != nil {
			for _, command := range bundle.CLISurface.Commands {
				if command.Availability == "implemented" && command.Intent == "direct_read" && command.Operation != "" {
					return ""
				}
			}
		}
		if len(bundle.Streams) == 0 {
			return "operational contract requires an implemented direct-read operation or a stream"
		}
	case "write":
		if !bundle.Metadata.Capabilities.Write {
			return "operational contract requires write but metadata.capabilities.write is false"
		}
		if len(bundle.Writes) == 0 {
			return "operational contract requires a typed writes.json action"
		}
	}
	return ""
}

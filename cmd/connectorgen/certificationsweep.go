package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	certificationSweepSchemaVersion       = 1
	certificationSweepEligiblePendingLive = "eligible_pending_live"
	certificationSweepSchemaConformant    = "schema_conformant"
	certificationSweepFixtureRequired     = "fixture_required"
	certificationSweepStatusProductDefect = "product_defect"
	certificationSweepProviderRefused     = "provider_refused"
	certificationSweepNotApplicable       = "not_applicable"
	certificationSweepArtifactFile        = "certification-sweep.json"
	certificationObservationFile          = "certification-observations.json"

	// certificationParityKind* is the entire scheduler-facing operation
	// vocabulary. Provider-specific operation kinds (for example GraphQL) are
	// normalized through their executable CLI/transport contract; callers must
	// never invent a ninth kind in certification.json.
	certificationParityKindRESTRead       = "rest_read"
	certificationParityKindRESTWrite      = "rest_write"
	certificationParityKindETL            = "etl"
	certificationParityKindReverseETL     = "reverse_etl"
	certificationParityKindBinaryDownload = "binary_download"
	certificationParityKindFileUpload     = "file_upload"
	certificationParityKindCDC            = "cdc"
	certificationParityKindChangefeed     = "changefeed"

	certificationParityClassDirectRead  = "direct_read"
	certificationParityClassDirectWrite = "direct_write"
	certificationParityClassETL         = "etl"
	certificationParityClassReverseETL  = "reverse_etl"
	certificationParityClassBinary      = "binary"
)

type certificationParityTransportRole string

const (
	certificationParityTransportSource      certificationParityTransportRole = "source"
	certificationParityTransportDestination certificationParityTransportRole = "destination"
)

// certificationParityInput is one declaration-owned execution source. A
// classifier call deliberately accepts one source at a time so a command,
// capability, changefeed, or managed transport cannot silently borrow a
// classification from an unrelated declaration.
type certificationParityInput struct {
	Command       *engine.CLICommand
	Operation     *engine.OperationSpec
	Write         *engine.WriteAction
	Stream        *engine.StreamSpec
	Capabilities  *engine.Capabilities
	Capability    string
	Changefeed    *connectors.ChangefeedDescriptor
	Transport     *connectors.SyncTransportDescriptor
	TransportRole certificationParityTransportRole
}

// certificationParityProjection is the generated scheduler identity. Empty
// kind/class is a genuine non-applicable source (such as config/auth), never
// a spelling for an additional operation kind. A malformed executable source
// instead returns an error so the sweep records a product defect.
type certificationParityProjection struct {
	OperationKind   string
	OpClass         string
	WriteActionKind string
}

func classifyCertificationParity(input certificationParityInput) (certificationParityProjection, error) {
	sources := 0
	if input.Command != nil {
		sources++
	}
	if input.Capabilities != nil {
		sources++
	}
	if input.Changefeed != nil {
		sources++
	}
	if input.Transport != nil {
		sources++
	}
	if sources == 0 {
		return certificationParityProjection{}, errors.New("parity classification requires one declaration source")
	}
	if sources != 1 {
		return certificationParityProjection{}, errors.New("parity classification accepts exactly one declaration source")
	}
	if input.Command != nil {
		return classifyCertificationParityCommand(*input.Command, input.Operation, input.Write, input.Stream)
	}
	if input.Capabilities != nil {
		switch input.Capability {
		case "read":
			if input.Capabilities.Read {
				return certificationParityProjection{OperationKind: certificationParityKindRESTRead, OpClass: certificationParityClassDirectRead}, nil
			}
		case "write":
			if input.Capabilities.Write {
				return certificationParityProjection{OperationKind: certificationParityKindRESTWrite, OpClass: certificationParityClassDirectWrite}, nil
			}
		case "cdc", "":
			if input.Capabilities.CDC {
				return certificationParityProjection{OperationKind: certificationParityKindCDC, OpClass: certificationParityClassETL}, nil
			}
		default:
			return certificationParityProjection{}, fmt.Errorf("capability projection has unsupported capability %q", input.Capability)
		}
		return certificationParityProjection{}, nil
	}
	if input.Changefeed != nil {
		// A declared changefeed remains a changefeed projection even when its
		// availability is not implemented. The sweep's accounting status carries
		// that availability; suppressing the kind would turn a real ETL
		// subcontract into an indistinguishable N/A row.
		return certificationParityProjection{OperationKind: certificationParityKindChangefeed, OpClass: certificationParityClassETL}, nil
	}
	switch input.TransportRole {
	case certificationParityTransportSource:
		if input.Transport.Source == nil {
			return certificationParityProjection{}, errors.New("source transport projection requires declared source transport")
		}
		return certificationParityProjection{OperationKind: certificationParityKindETL, OpClass: certificationParityClassETL}, nil
	case certificationParityTransportDestination:
		if input.Transport.Destination == nil {
			return certificationParityProjection{}, errors.New("destination transport projection requires declared destination transport")
		}
		return certificationParityProjection{OperationKind: certificationParityKindReverseETL, OpClass: certificationParityClassReverseETL}, nil
	default:
		return certificationParityProjection{}, fmt.Errorf("transport projection has unsupported role %q", input.TransportRole)
	}
}

func classifyCertificationParityCommand(command engine.CLICommand, operation *engine.OperationSpec, write *engine.WriteAction, stream *engine.StreamSpec) (certificationParityProjection, error) {
	operationKind := ""
	if operation != nil {
		operationKind = operation.Kind
	}
	if command.Operation != "" && operation == nil {
		return certificationParityProjection{}, fmt.Errorf("command %q references missing operation %q", command.Path, command.Operation)
	}
	if command.Write != "" && write == nil {
		return certificationParityProjection{}, fmt.Errorf("command %q references missing write action %q", command.Path, command.Write)
	}
	switch command.Intent {
	case "direct_read":
		if operationKind != "" && operationKind != certificationParityKindRESTRead && operationKind != "graphql_query" {
			return certificationParityProjection{}, fmt.Errorf("direct_read command %q references operation kind %q", command.Path, operationKind)
		}
		return certificationParityProjection{OperationKind: certificationParityKindRESTRead, OpClass: certificationParityClassDirectRead}, nil
	case "direct_write":
		switch operationKind {
		case "", certificationParityKindRESTWrite, "graphql_mutation":
			projection := certificationParityProjection{OperationKind: certificationParityKindRESTWrite, OpClass: certificationParityClassDirectWrite}
			if write != nil {
				if !validCertificationWriteActionKind(write.Kind) {
					return certificationParityProjection{}, fmt.Errorf("direct_write command %q write action %q has unsupported kind %q", command.Path, write.Name, write.Kind)
				}
				projection.WriteActionKind = write.Kind
			} else if operation != nil {
				kind, err := certificationOperationWriteActionKind(command, *operation)
				if err != nil {
					return certificationParityProjection{}, err
				}
				projection.WriteActionKind = kind
			}
			return projection, nil
		case certificationParityKindFileUpload:
			return certificationParityProjection{OperationKind: certificationParityKindFileUpload, OpClass: certificationParityClassBinary}, nil
		default:
			return certificationParityProjection{}, fmt.Errorf("direct_write command %q references operation kind %q", command.Path, operationKind)
		}
	case "etl":
		if command.Stream == "" || stream == nil || stream.Name != command.Stream {
			return certificationParityProjection{}, fmt.Errorf("etl command %q requires declared stream %q", command.Path, command.Stream)
		}
		return certificationParityProjection{OperationKind: certificationParityKindETL, OpClass: certificationParityClassETL}, nil
	case "reverse_etl":
		if command.Write == "" || write == nil || write.Name != command.Write {
			return certificationParityProjection{}, fmt.Errorf("reverse_etl command %q requires declared write action %q", command.Path, command.Write)
		}
		if !validCertificationWriteActionKind(write.Kind) {
			return certificationParityProjection{}, fmt.Errorf("reverse_etl command %q write action %q has unsupported kind %q", command.Path, write.Name, write.Kind)
		}
		// Delete is an independently selectable mutation family inside the
		// existing direct-write parity cell. It is not a ninth operation kind
		// or sixth class, even when its command reaches the shared write safety
		// boundary through a reverse-ETL action declaration.
		if write.Kind == "delete" {
			return certificationParityProjection{OperationKind: certificationParityKindRESTWrite, OpClass: certificationParityClassDirectWrite, WriteActionKind: write.Kind}, nil
		}
		return certificationParityProjection{OperationKind: certificationParityKindReverseETL, OpClass: certificationParityClassReverseETL, WriteActionKind: write.Kind}, nil
	case "binary_download":
		if operationKind != "" && operationKind != certificationParityKindBinaryDownload {
			return certificationParityProjection{}, fmt.Errorf("binary_download command %q references operation kind %q", command.Path, operationKind)
		}
		return certificationParityProjection{OperationKind: certificationParityKindBinaryDownload, OpClass: certificationParityClassBinary}, nil
	default:
		return certificationParityProjection{}, nil
	}
}

// certificationOperationWriteActionKind derives the mutation family from the
// operation declaration when a command is not backed by writes.json. The
// writes.json branch stays authoritative because it is the more specific
// action declaration. Operation-backed actions must be explicit enough to
// classify; an uncertain POST-style mutation remains a product defect instead
// of being silently folded into custom.
func certificationOperationWriteActionKind(command engine.CLICommand, operation engine.OperationSpec) (string, error) {
	if validCertificationWriteActionKind(operation.MutationClass) {
		return operation.MutationClass, nil
	}
	switch operation.Kind {
	case "graphql_mutation":
		// GraphQL mutations have no REST verb with CRUD semantics. Their declared
		// operation kind is nevertheless sufficient to identify the custom
		// mutation family, without guessing from a provider-specific operation ID.
		return "custom", nil
	case certificationParityKindRESTWrite:
		if operation.REST == nil {
			return "", fmt.Errorf("direct_write command %q operation %q cannot determine write action kind: rest_write has no REST declaration", command.Path, operation.ID)
		}
		switch strings.ToUpper(strings.TrimSpace(operation.REST.Method)) {
		case "DELETE":
			return "delete", nil
		case "PATCH":
			return "update", nil
		case "PUT":
			return "upsert", nil
		}
	}
	return "", fmt.Errorf("direct_write command %q operation %q cannot determine write action kind from its operation declaration", command.Path, operation.ID)
}

func validCertificationWriteActionKind(kind string) bool {
	switch kind {
	case "create", "update", "upsert", "delete", "custom":
		return true
	default:
		return false
	}
}

func validCertificationParityKind(kind string) bool {
	switch kind {
	case certificationParityKindRESTRead, certificationParityKindRESTWrite, certificationParityKindETL,
		certificationParityKindReverseETL, certificationParityKindBinaryDownload, certificationParityKindFileUpload,
		certificationParityKindCDC, certificationParityKindChangefeed:
		return true
	default:
		return false
	}
}

func validCertificationParityClass(opClass string) bool {
	switch opClass {
	case certificationParityClassDirectRead, certificationParityClassDirectWrite, certificationParityClassETL,
		certificationParityClassReverseETL, certificationParityClassBinary:
		return true
	default:
		return false
	}
}

// certificationSweep is the deterministic, source-derived accounting record
// for every schedulable connector declaration: CLI commands, applicable
// capabilities, changefeeds, and closed transport roles. It deliberately holds
// no provider response or credential: a generated candidate becomes a passing
// result only when a separate live proof is accepted.
type certificationSweep struct {
	SchemaVersion    int                                 `json:"schema_version"`
	Connector        string                              `json:"connector"`
	Source           string                              `json:"source"`
	DeclaredCommands int                                 `json:"declared_commands"`
	DeclaredRows     int                                 `json:"declared_rows"`
	StatusTotal      int                                 `json:"status_total"`
	Commands         []certificationSweepCommand         `json:"commands"`
	ProductDefects   []certificationSweepProductDefect   `json:"product_defects"`
	ProviderRefusals []certificationSweepProviderRefusal `json:"provider_refusals"`
}

// certificationSweepCommand is one generated certification candidate and its
// current, non-affirmative accounting status. Its path is either a CLI command
// path or a stable declaration identity such as "capability read" or
// "transport destination". Assertion metadata remains a narrow overlay from
// certification.json and is available only for CLI-command rows.
type certificationSweepCommand struct {
	Summary             string                                `json:"summary"`
	Path                string                                `json:"path"`
	Intent              string                                `json:"intent"`
	Availability        string                                `json:"availability"`
	OperationKind       *string                               `json:"operation_kind"`
	OpClass             *string                               `json:"op_class"`
	WriteActionKind     string                                `json:"write_action_kind,omitempty"`
	Stream              string                                `json:"stream,omitempty"`
	Flags               []certificationSweepFlag              `json:"flags"`
	APISurface          []engine.CLISurfaceEndpointRef        `json:"api_surface"`
	Status              string                                `json:"status"`
	Reason              string                                `json:"reason"`
	RequiredFlags       []string                              `json:"required_flags,omitempty"`
	OutputAssertions    []engine.CertificationOutputAssertion `json:"output_assertions,omitempty"`
	AssertionSource     string                                `json:"assertion_source,omitempty"`
	CertificationCohort string                                `json:"certification_cohort,omitempty"`
}

// certificationSweepFlag retains the declaration-owned inputs a fixture or
// live runner must satisfy. It intentionally carries enum values rather than
// replacing them with hand-authored example arguments.
type certificationSweepFlag struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Values   []string `json:"values,omitempty"`
	MapsTo   string   `json:"maps_to,omitempty"`
}

// certificationSweepProductDefect records a disagreement between the declared
// CLI surface and its own executable operation contract. It is intentionally
// separate from provider refusals: this row can be corrected locally without
// treating a provider response as a product bug.
type certificationSweepProductDefect struct {
	Command       string `json:"command"`
	Flag          string `json:"flag"`
	PathParameter string `json:"path_parameter"`
	Reason        string `json:"reason"`
}

// certificationSweepProviderRefusal captures a concrete provider rejection
// from a later live sweep. Generation creates none because no provider has
// been contacted; the validator keeps the non-pass evidence shape honest.
type certificationSweepProviderRefusal struct {
	Command        string `json:"command"`
	ProviderStatus int    `json:"provider_status"`
	Reason         string `json:"reason"`
}

type certificationSweepOptions struct {
	repoRoot  string
	connector string
	check     bool
}

// runCertificationSweep generates or verifies the source-controlled candidate
// ledger for exactly one connector. It is intentionally a local, deterministic
// generator: it does not read credentials or contact a provider.
func runCertificationSweep(args []string, stdout, stderr io.Writer) int {
	options, err := parseCertificationSweepOptions(args)
	if err != nil {
		logf(stderr, "connectorgen certification-sweep: %v\n", err)
		return 2
	}
	sweep, err := buildCertificationSweep(options.repoRoot, options.connector)
	if err != nil {
		logf(stderr, "connectorgen certification-sweep: %v\n", err)
		return 1
	}
	raw, err := marshalCertificationSweep(sweep)
	if err != nil {
		logf(stderr, "connectorgen certification-sweep: %v\n", err)
		return 1
	}
	path := certificationSweepArtifactPath(options.repoRoot, options.connector)
	if options.check {
		current, err := os.ReadFile(path)
		if err != nil {
			logf(stderr, "connectorgen certification-sweep: generated artifact %q is missing; run `go run ./cmd/connectorgen certification-sweep --connector %s`\n", filepath.ToSlash(path), options.connector)
			return 1
		}
		if err := validateCertificationSweepArtifact(current); err != nil {
			logf(stderr, "connectorgen certification-sweep: generated artifact %q is invalid: %v\n", filepath.ToSlash(path), err)
			return 1
		}
		if !bytes.Equal(current, raw) {
			logf(stderr, "connectorgen certification-sweep: generated artifact %q has drift; run `go run ./cmd/connectorgen certification-sweep --connector %s`\n", filepath.ToSlash(path), options.connector)
			return 1
		}
		logf(stdout, "connectorgen certification-sweep: %s is current (%d rows; %d CLI commands)\n", filepath.ToSlash(path), sweep.DeclaredRows, sweep.DeclaredCommands)
		return 0
	}
	if err := writeGeneratedArtifact(path, raw); err != nil {
		logf(stderr, "connectorgen certification-sweep: write %q: %v\n", filepath.ToSlash(path), err)
		return 1
	}
	logf(stdout, "connectorgen certification-sweep: wrote %s (%d rows; %d CLI commands)\n", filepath.ToSlash(path), sweep.DeclaredRows, sweep.DeclaredCommands)
	return 0
}

func parseCertificationSweepOptions(args []string) (certificationSweepOptions, error) {
	options := certificationSweepOptions{}
	for index := 1; index < len(args); index++ {
		arg := args[index]
		switch arg {
		case "--connector":
			if index+1 >= len(args) {
				return certificationSweepOptions{}, errors.New("--connector requires a name")
			}
			index++
			options.connector = args[index]
		case "--check":
			options.check = true
		default:
			if strings.HasPrefix(arg, "-") || options.repoRoot != "" {
				return certificationSweepOptions{}, fmt.Errorf("unexpected argument %q", arg)
			}
			options.repoRoot = arg
		}
	}
	if !isSafeProofIdentifier(options.connector) {
		return certificationSweepOptions{}, errors.New("--connector must be a safe connector name")
	}
	if options.repoRoot == "" {
		root, err := repoRoot()
		if err != nil {
			return certificationSweepOptions{}, fmt.Errorf("resolve repository root: %w", err)
		}
		options.repoRoot = root
	}
	root, err := filepath.Abs(options.repoRoot)
	if err != nil {
		return certificationSweepOptions{}, fmt.Errorf("resolve repository root: %w", err)
	}
	options.repoRoot = root
	return options, nil
}

func certificationSweepArtifactPath(repoRoot, connector string) string {
	return filepath.Join(repoRoot, "internal", "connectors", "defs", connector, certificationSweepArtifactFile)
}

func certificationSweepObservationPath(repoRoot, connector string) string {
	return filepath.Join(repoRoot, "internal", "connectors", "defs", connector, certificationObservationFile)
}

func buildCertificationSweep(repoRoot, connector string) (certificationSweep, error) {
	definitionsRoot := filepath.Join(repoRoot, "internal", "connectors", "defs")
	bundle, err := engine.Load(os.DirFS(definitionsRoot), connector)
	if err != nil {
		return certificationSweep{}, fmt.Errorf("load connector %q: %w", connector, err)
	}
	assertions, err := certificationSweepAssertions(&bundle)
	if err != nil {
		return certificationSweep{}, err
	}
	graphql, err := certificationSweepGraphQLProfileFor(&bundle)
	if err != nil {
		return certificationSweep{}, err
	}
	providerRefusals, err := certificationSweepProviderRefusals(repoRoot, connector)
	if err != nil {
		return certificationSweep{}, err
	}
	refusalByCommand := make(map[string]certificationSweepProviderRefusal, len(providerRefusals))
	for _, refusal := range providerRefusals {
		if _, duplicate := refusalByCommand[refusal.Command]; duplicate {
			return certificationSweep{}, fmt.Errorf("provider-refusal observation duplicates command %q", refusal.Command)
		}
		refusalByCommand[refusal.Command] = refusal
	}
	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}
	writes := make(map[string]engine.WriteAction, len(bundle.Writes))
	for _, write := range bundle.Writes {
		writes[write.Name] = write
	}
	streams := make(map[string]engine.StreamSpec, len(bundle.Streams))
	for _, stream := range bundle.Streams {
		streams[stream.Name] = stream
	}
	commands := []engine.CLICommand(nil)
	if bundle.CLISurface != nil {
		commands = append(commands, bundle.CLISurface.Commands...)
	}
	sort.Slice(commands, func(i, j int) bool { return commands[i].Path < commands[j].Path })

	sweep := certificationSweep{
		SchemaVersion:    certificationSweepSchemaVersion,
		Connector:        connector,
		Source:           "connector declarations",
		DeclaredCommands: len(commands),
		Commands:         make([]certificationSweepCommand, 0, len(commands)+5),
		ProductDefects:   []certificationSweepProductDefect{},
		ProviderRefusals: providerRefusals,
	}
	seen := make(map[string]bool, len(commands))
	for _, command := range commands {
		if strings.TrimSpace(command.Path) == "" || seen[command.Path] {
			return certificationSweep{}, fmt.Errorf("connector %q cli_surface has missing or duplicate command path %q", connector, command.Path)
		}
		seen[command.Path] = true
		var operation *engine.OperationSpec
		if item, found := operations[command.Operation]; found {
			operation = &item
		}
		var write *engine.WriteAction
		if item, found := writes[command.Write]; found {
			write = &item
		}
		var stream *engine.StreamSpec
		if item, found := streams[command.Stream]; found {
			stream = &item
		}
		projection, projectionErr := classifyCertificationParity(certificationParityInput{
			Command: &command, Operation: operation, Write: write, Stream: stream,
		})
		var operationValue engine.OperationSpec
		if operation != nil {
			operationValue = *operation
		}
		row, defect := classifyCertificationSweepCommand(command, operationValue, projection, projectionErr, assertions[command.Path], graphql, refusalByCommand[command.Path])
		sweep.Commands = append(sweep.Commands, row)
		if defect != nil {
			sweep.ProductDefects = append(sweep.ProductDefects, *defect)
		}
	}
	declarations, declarationDefects := certificationSweepDeclarationRows(&bundle)
	for _, row := range declarations {
		if seen[row.Path] {
			return certificationSweep{}, fmt.Errorf("connector %q declaration path %q conflicts with a CLI command", connector, row.Path)
		}
		seen[row.Path] = true
		sweep.Commands = append(sweep.Commands, row)
	}
	sweep.ProductDefects = append(sweep.ProductDefects, declarationDefects...)
	for path := range assertions {
		if !seen[path] {
			return certificationSweep{}, fmt.Errorf("certification assertion overlay command %q is absent from cli_surface.json", path)
		}
	}
	for path := range refusalByCommand {
		if !seen[path] {
			return certificationSweep{}, fmt.Errorf("provider-refusal observation command %q is absent from cli_surface.json", path)
		}
	}
	sweep.DeclaredRows = len(sweep.Commands)
	sweep.StatusTotal = len(sweep.Commands)
	if err := validateCertificationSweep(sweep); err != nil {
		return certificationSweep{}, err
	}
	return sweep, nil
}

type certificationSweepDeclaration struct {
	Summary      string
	Path         string
	Intent       string
	Availability string
	Status       string
	Reason       string
	Input        certificationParityInput
}

// certificationSweepDeclarationRows makes non-CLI contracts schedulable
// without inventing a user command or consulting the generic Connector.Write
// capability for a managed destination. These rows are deliberately fixture
// required: G3+ owns cell execution and proof binding.
func certificationSweepDeclarationRows(bundle *engine.Bundle) ([]certificationSweepCommand, []certificationSweepProductDefect) {
	if bundle == nil {
		return nil, []certificationSweepProductDefect{{Command: "<bundle>", Flag: "<parity-projection>", PathParameter: "<parity-projection>", Reason: "parity projection requires a bundle"}}
	}
	declarations := make([]certificationSweepDeclaration, 0, 5)
	capabilities := &bundle.Metadata.Capabilities
	if capabilities.Read {
		declarations = append(declarations, certificationSweepDeclaration{
			Summary: "Declared read capability", Path: "capability read", Intent: "capability_read", Availability: "implemented",
			Status: certificationSweepFixtureRequired, Reason: "declared read capability requires connector-owned fixture and live proof",
			Input: certificationParityInput{Capabilities: capabilities, Capability: "read"},
		})
	}
	if capabilities.Write {
		declarations = append(declarations, certificationSweepDeclaration{
			Summary: "Declared write capability", Path: "capability write", Intent: "capability_write", Availability: "implemented",
			Status: certificationSweepFixtureRequired, Reason: "declared write capability requires connector-owned fixture and live proof",
			Input: certificationParityInput{Capabilities: capabilities, Capability: "write"},
		})
	}
	if capabilities.CDC {
		declarations = append(declarations, certificationSweepDeclaration{
			Summary: "Declared CDC capability", Path: "capability cdc", Intent: "capability_cdc", Availability: "implemented",
			Status: certificationSweepFixtureRequired, Reason: "declared CDC capability requires connector-owned fixture and live proof",
			Input: certificationParityInput{Capabilities: capabilities, Capability: "cdc"},
		})
	}
	if bundle.Changefeed != nil {
		status := certificationSweepFixtureRequired
		reason := "declared changefeed requires connector-owned fixture and live proof"
		availability := string(bundle.Changefeed.Status)
		if bundle.Changefeed.Status != connectors.ChangefeedStatusImplemented {
			status = certificationSweepNotApplicable
			reason = fmt.Sprintf("declared changefeed status is %q, not implemented", bundle.Changefeed.Status)
		}
		declarations = append(declarations, certificationSweepDeclaration{
			Summary: "Declared changefeed", Path: "changefeed", Intent: "changefeed", Availability: availability,
			Status: status, Reason: reason, Input: certificationParityInput{Changefeed: bundle.Changefeed},
		})
	}
	if bundle.SyncTransport != nil && bundle.SyncTransport.Source != nil {
		declarations = append(declarations, certificationSweepDeclaration{
			Summary: "Declared source transport", Path: "transport source", Intent: "transport_source", Availability: "implemented",
			Status: certificationSweepFixtureRequired, Reason: "declared source transport requires connector-owned fixture and live proof",
			Input: certificationParityInput{Transport: bundle.SyncTransport, TransportRole: certificationParityTransportSource},
		})
	}
	if bundle.SyncTransport != nil && bundle.SyncTransport.Destination != nil {
		declarations = append(declarations, certificationSweepDeclaration{
			Summary: "Declared managed destination transport", Path: "transport destination", Intent: "transport_destination", Availability: "implemented",
			Status: certificationSweepFixtureRequired, Reason: "declared managed destination transport requires connector-owned fixture and live proof",
			Input: certificationParityInput{Transport: bundle.SyncTransport, TransportRole: certificationParityTransportDestination},
		})
	}

	rows := make([]certificationSweepCommand, 0, len(declarations))
	defects := make([]certificationSweepProductDefect, 0)
	for _, declaration := range declarations {
		projection, err := classifyCertificationParity(declaration.Input)
		row := certificationSweepCommand{
			Summary: declaration.Summary, Path: declaration.Path, Intent: declaration.Intent, Availability: declaration.Availability,
			OperationKind: certificationSweepParityValue(projection.OperationKind), OpClass: certificationSweepParityValue(projection.OpClass),
			Flags: []certificationSweepFlag{}, APISurface: []engine.CLISurfaceEndpointRef{}, Status: declaration.Status, Reason: declaration.Reason,
		}
		if err == nil && (row.OperationKind == nil || row.OpClass == nil) {
			err = errors.New("declared parity source has no valid projection")
		}
		if err != nil {
			row.Status = certificationSweepStatusProductDefect
			row.Reason = "parity projection: " + err.Error()
			defects = append(defects, certificationSweepProductDefect{Command: row.Path, Flag: "<parity-projection>", PathParameter: "<parity-projection>", Reason: row.Reason})
		}
		rows = append(rows, row)
	}
	return rows, defects
}

type certificationSweepObservations struct {
	SchemaVersion    int                                 `json:"schema_version"`
	ProviderRefusals []certificationSweepProviderRefusal `json:"provider_refusals"`
}

func certificationSweepProviderRefusals(repoRoot, connector string) ([]certificationSweepProviderRefusal, error) {
	raw, err := os.ReadFile(certificationSweepObservationPath(repoRoot, connector))
	if errors.Is(err, os.ErrNotExist) {
		return []certificationSweepProviderRefusal{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("read certification observations for connector %q: %w", connector, err)
	}
	var observations certificationSweepObservations
	if err := decodeStrictJSON(raw, &observations); err != nil {
		return nil, fmt.Errorf("parse certification observations for connector %q: %w", connector, err)
	}
	if observations.SchemaVersion != certificationSweepSchemaVersion {
		return nil, fmt.Errorf("certification observations for connector %q require schema_version %d", connector, certificationSweepSchemaVersion)
	}
	return append([]certificationSweepProviderRefusal(nil), observations.ProviderRefusals...), nil
}

type certificationSweepAssertionOverlay struct {
	Assertions []engine.CertificationOutputAssertion
	Source     string
	Cohort     string
}

type certificationSweepGraphQLProfile struct {
	CommandPrefix          string
	SchemaConformantReason string
	FixtureRequiredReason  string
}

func certificationSweepAssertions(bundle *engine.Bundle) (map[string]certificationSweepAssertionOverlay, error) {
	assertions := map[string]certificationSweepAssertionOverlay{}
	if bundle.Certification == nil {
		return assertions, nil
	}
	for _, candidate := range bundle.Certification.DirectReadCandidates {
		if strings.TrimSpace(candidate.Command) == "" || len(candidate.OutputAssertions) == 0 {
			return nil, errors.New("direct-read certification assertion overlay requires command and produced-value assertion")
		}
		if _, found := assertions[candidate.Command]; found {
			return nil, fmt.Errorf("multiple certification assertion overlays for command %q", candidate.Command)
		}
		assertions[candidate.Command] = certificationSweepAssertionOverlay{
			Assertions: append([]engine.CertificationOutputAssertion(nil), candidate.OutputAssertions...),
			Source:     certificationSweepDirectReadAssertionSource(candidate),
			Cohort:     candidate.Cohort,
		}
	}
	if graphql := bundle.Certification.GraphQL; graphql != nil {
		for _, candidate := range graphql.LiveCandidates {
			if strings.TrimSpace(candidate.Command) == "" || len(candidate.OutputAssertions) == 0 {
				return nil, errors.New("GraphQL certification assertion overlay requires command and produced-value assertion")
			}
			if _, found := assertions[candidate.Command]; found {
				return nil, fmt.Errorf("multiple certification assertion overlays for command %q", candidate.Command)
			}
			assertions[candidate.Command] = certificationSweepAssertionOverlay{
				Assertions: append([]engine.CertificationOutputAssertion(nil), candidate.OutputAssertions...),
				Source:     "certification.json graphql.live_candidates",
			}
		}
	}
	return assertions, nil
}

func certificationSweepDirectReadAssertionSource(candidate engine.CertificationCommandCandidate) string {
	if candidate.Generated {
		return "certification.json direct_read_candidates generated from cli_surface.json"
	}
	return "certification.json direct_read_candidates"
}

func certificationSweepGraphQLProfileFor(bundle *engine.Bundle) (certificationSweepGraphQLProfile, error) {
	if bundle.Certification == nil || bundle.Certification.GraphQL == nil {
		return certificationSweepGraphQLProfile{}, nil
	}
	graphql := bundle.Certification.GraphQL
	if strings.TrimSpace(graphql.CommandPrefix) == "" || strings.TrimSpace(graphql.SchemaConformantReason) == "" || strings.TrimSpace(graphql.FixtureRequiredReason) == "" {
		return certificationSweepGraphQLProfile{}, errors.New("GraphQL certification profile is missing a command prefix or non-pass reason")
	}
	return certificationSweepGraphQLProfile{
		CommandPrefix:          graphql.CommandPrefix,
		SchemaConformantReason: graphql.SchemaConformantReason,
		FixtureRequiredReason:  graphql.FixtureRequiredReason,
	}, nil
}

func classifyCertificationSweepCommand(command engine.CLICommand, operation engine.OperationSpec, projection certificationParityProjection, projectionErr error, assertion certificationSweepAssertionOverlay, graphql certificationSweepGraphQLProfile, providerRefusal certificationSweepProviderRefusal) (certificationSweepCommand, *certificationSweepProductDefect) {
	row := certificationSweepCommand{
		Summary:             command.Summary,
		Path:                command.Path,
		Intent:              command.Intent,
		Availability:        command.Availability,
		OperationKind:       certificationSweepParityValue(projection.OperationKind),
		OpClass:             certificationSweepParityValue(projection.OpClass),
		WriteActionKind:     projection.WriteActionKind,
		Stream:              command.Stream,
		Flags:               certificationSweepFlags(command.Flags),
		APISurface:          append([]engine.CLISurfaceEndpointRef(nil), command.APISurface...),
		RequiredFlags:       certificationSweepRequiredFlags(command.Flags),
		CertificationCohort: assertion.Cohort,
	}
	if projectionErr != nil {
		row.Status = certificationSweepStatusProductDefect
		row.Reason = "parity projection: " + projectionErr.Error()
		return row, &certificationSweepProductDefect{Command: command.Path, Flag: "<parity-projection>", PathParameter: "<parity-projection>", Reason: row.Reason}
	}
	if command.Availability != "implemented" {
		row.Status = certificationSweepNotApplicable
		row.Reason = fmt.Sprintf("declared availability is %q, not implemented", command.Availability)
		return row, nil
	}
	if defect := requiredPathFlagDefect(command, operation); defect != nil {
		row.Status = certificationSweepStatusProductDefect
		row.Reason = defect.Reason
		if len(assertion.Assertions) != 0 {
			row.OutputAssertions = assertion.Assertions
			row.AssertionSource = assertion.Source
		}
		return row, defect
	}
	if providerRefusal.Command != "" {
		row.Status = certificationSweepProviderRefused
		row.Reason = providerRefusal.Reason
		if len(assertion.Assertions) != 0 {
			row.OutputAssertions = assertion.Assertions
			row.AssertionSource = assertion.Source
		}
		return row, nil
	}
	if graphql.CommandPrefix != "" && strings.HasPrefix(command.Path, graphql.CommandPrefix) {
		switch operation.Kind {
		case "graphql_query":
			if command.Intent != "direct_read" {
				row.Status = certificationSweepStatusProductDefect
				row.Reason = "GraphQL query command does not declare direct_read intent"
				return row, &certificationSweepProductDefect{Command: command.Path, Reason: row.Reason}
			}
			if len(assertion.Assertions) != 0 {
				row.Status = certificationSweepEligiblePendingLive
				row.Reason = "declaration-owned GraphQL produced-value assertions exist; live execution is pending"
				row.OutputAssertions = assertion.Assertions
				row.AssertionSource = assertion.Source
				return row, nil
			}
			row.Status = certificationSweepSchemaConformant
			row.Reason = graphql.SchemaConformantReason
			return row, nil
		case "graphql_mutation":
			if command.Intent != "direct_write" {
				row.Status = certificationSweepStatusProductDefect
				row.Reason = "GraphQL mutation command does not declare direct_write intent"
				return row, &certificationSweepProductDefect{Command: command.Path, Reason: row.Reason}
			}
			row.Status = certificationSweepFixtureRequired
			row.Reason = certificationSweepFixtureReason(graphql.FixtureRequiredReason, row.RequiredFlags)
			return row, nil
		default:
			row.Status = certificationSweepStatusProductDefect
			row.Reason = fmt.Sprintf("GraphQL certification command has operation kind %q", operation.Kind)
			return row, &certificationSweepProductDefect{Command: command.Path, Reason: row.Reason}
		}
	}
	if len(assertion.Assertions) != 0 {
		row.Status = certificationSweepEligiblePendingLive
		row.Reason = "declaration-owned produced-value assertions exist; live execution is pending"
		row.OutputAssertions = assertion.Assertions
		row.AssertionSource = assertion.Source
		return row, nil
	}
	switch command.Intent {
	case "direct_read":
		row.Status = certificationSweepFixtureRequired
		row.Reason = certificationSweepFixtureReason("direct-read fixture and declaration-owned produced-value assertion are required", row.RequiredFlags)
	case "binary_download":
		row.Status = certificationSweepFixtureRequired
		row.Reason = certificationSweepFixtureReason("owned binary fixture, destination, and produced-value assertion are required", row.RequiredFlags)
	case "etl":
		row.Status = certificationSweepFixtureRequired
		row.Reason = certificationSweepFixtureReason("stream fixture and produced-record assertion are required", row.RequiredFlags)
	case "direct_write", "reverse_etl":
		row.Status = certificationSweepFixtureRequired
		row.Reason = certificationSweepFixtureReason("owned mutation fixture, plan, preview, approval, execution, independent read-back, and cleanup are required", row.RequiredFlags)
	default:
		row.Status = certificationSweepNotApplicable
		row.Reason = fmt.Sprintf("intent %q has no provider certification execution contract", command.Intent)
	}
	return row, nil
}

func certificationSweepParityValue(value string) *string {
	if value == "" {
		return nil
	}
	return &value
}

func certificationSweepFlags(flags []engine.CLIFlag) []certificationSweepFlag {
	out := make([]certificationSweepFlag, 0, len(flags))
	for _, flag := range flags {
		out = append(out, certificationSweepFlag{
			Name:     flag.Name,
			Type:     flag.Type,
			Required: flag.Required,
			Values:   append([]string(nil), flag.Values...),
			MapsTo:   flag.MapsTo,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

func certificationSweepRequiredFlags(flags []engine.CLIFlag) []string {
	required := make([]string, 0, len(flags))
	for _, flag := range flags {
		if flag.Required {
			required = append(required, "--"+flag.Name)
		}
	}
	sort.Strings(required)
	return required
}

func certificationSweepFixtureReason(prefix string, requiredFlags []string) string {
	if len(requiredFlags) == 0 {
		return prefix
	}
	return prefix + "; required flags: " + strings.Join(requiredFlags, ", ")
}

func requiredPathFlagDefect(command engine.CLICommand, operation engine.OperationSpec) *certificationSweepProductDefect {
	if command.Operation == "" || operation.REST == nil {
		return nil
	}
	required := make([]string, 0, len(operation.REST.Parameters))
	for _, parameter := range operation.REST.Parameters {
		if parameter.In == "path" && parameter.Required {
			required = append(required, parameter.Name)
		}
	}
	sort.Strings(required)
	for _, pathParameter := range required {
		var mappedFlag *engine.CLIFlag
		for index := range command.Flags {
			flag := &command.Flags[index]
			mappedParameter, mapped := strings.CutPrefix(flag.MapsTo, "path.")
			if mapped && mappedParameter == pathParameter {
				mappedFlag = flag
				break
			}
		}
		if mappedFlag == nil {
			return &certificationSweepProductDefect{
				Command:       command.Path,
				Flag:          "<missing>",
				PathParameter: pathParameter,
				Reason:        fmt.Sprintf("required REST path parameter %q has no mapped CLI flag", pathParameter),
			}
		}
		if !mappedFlag.Required {
			return &certificationSweepProductDefect{
				Command:       command.Path,
				Flag:          mappedFlag.Name,
				PathParameter: pathParameter,
				Reason:        fmt.Sprintf("required REST path parameter %q maps to CLI flag --%s that is not required", pathParameter, mappedFlag.Name),
			}
		}
	}
	return nil
}

func validateCertificationSweep(sweep certificationSweep) error {
	if sweep.SchemaVersion != certificationSweepSchemaVersion || !isSafeProofIdentifier(sweep.Connector) || sweep.Source != "connector declarations" {
		return errors.New("sweep requires supported schema, safe connector, and connector declarations source")
	}
	if sweep.DeclaredCommands < 0 || sweep.DeclaredRows != len(sweep.Commands) || sweep.StatusTotal != len(sweep.Commands) || sweep.DeclaredCommands > sweep.DeclaredRows {
		return errors.New("sweep declaration totals do not reconcile")
	}
	paths := make(map[string]certificationSweepCommand, len(sweep.Commands))
	for _, command := range sweep.Commands {
		if strings.TrimSpace(command.Path) == "" || strings.TrimSpace(command.Summary) == "" || strings.TrimSpace(command.Intent) == "" || strings.TrimSpace(command.Availability) == "" || strings.TrimSpace(command.Reason) == "" {
			return errors.New("sweep command requires summary, path, intent, availability, and reason")
		}
		if _, found := paths[command.Path]; found {
			return fmt.Errorf("sweep has duplicate command path %q", command.Path)
		}
		if !validCertificationSweepStatus(command.Status) || command.Status == "pass" {
			return fmt.Errorf("sweep command %q has invalid affirmative status %q", command.Path, command.Status)
		}
		if err := validateCertificationSweepParity(command); err != nil {
			return err
		}
		if command.Status == certificationSweepEligiblePendingLive && (len(command.OutputAssertions) == 0 || command.AssertionSource == "") {
			return fmt.Errorf("eligible command %q requires produced-value assertions and their source", command.Path)
		}
		if command.Status != certificationSweepEligiblePendingLive && command.Status != certificationSweepStatusProductDefect && command.Status != certificationSweepProviderRefused && (len(command.OutputAssertions) != 0 || command.AssertionSource != "") {
			return fmt.Errorf("non-eligible command %q cannot carry assertion overlay", command.Path)
		}
		if (command.Status == certificationSweepStatusProductDefect || command.Status == certificationSweepProviderRefused) && (len(command.OutputAssertions) == 0) != (command.AssertionSource == "") {
			return fmt.Errorf("non-eligible assertion command %q must retain both assertion overlay fields or neither", command.Path)
		}
		paths[command.Path] = command
	}
	defects := make(map[string]bool, len(sweep.ProductDefects))
	for _, defect := range sweep.ProductDefects {
		if strings.TrimSpace(defect.Command) == "" || strings.TrimSpace(defect.Flag) == "" || strings.TrimSpace(defect.PathParameter) == "" || strings.TrimSpace(defect.Reason) == "" || defects[defect.Command] {
			return errors.New("product defect requires one concrete command, flag, path parameter, and reason")
		}
		if command, found := paths[defect.Command]; !found || command.Status != certificationSweepStatusProductDefect {
			return fmt.Errorf("product defect %q is not the command's classification", defect.Command)
		}
		defects[defect.Command] = true
	}
	providerRefusals := make(map[string]bool, len(sweep.ProviderRefusals))
	for _, refusal := range sweep.ProviderRefusals {
		if strings.TrimSpace(refusal.Command) == "" || refusal.ProviderStatus < 100 || refusal.ProviderStatus > 599 || strings.TrimSpace(refusal.Reason) == "" {
			return errors.New("provider refusal requires command, concrete HTTP status, and reason")
		}
		if command, found := paths[refusal.Command]; !found || command.Status != certificationSweepProviderRefused {
			return fmt.Errorf("provider refusal %q is not the command's classification", refusal.Command)
		}
		if providerRefusals[refusal.Command] {
			return fmt.Errorf("provider refusal %q is duplicated", refusal.Command)
		}
		providerRefusals[refusal.Command] = true
	}
	for path, command := range paths {
		if command.Status == certificationSweepStatusProductDefect && !defects[path] {
			return fmt.Errorf("product-defect command %q has no concrete product-defect record", path)
		}
		if command.Status == certificationSweepProviderRefused && !providerRefusals[path] {
			return fmt.Errorf("provider-refused command %q has no concrete provider refusal record", path)
		}
	}
	return nil
}

func validateCertificationSweepParity(command certificationSweepCommand) error {
	if (command.OperationKind == nil) != (command.OpClass == nil) {
		return fmt.Errorf("sweep command %q must carry operation_kind and op_class together", command.Path)
	}
	if command.OperationKind == nil {
		if certificationSweepIntentRequiresParity(command.Intent) && command.Status != certificationSweepStatusProductDefect {
			return fmt.Errorf("sweep command %q has no valid parity projection and is not a product defect", command.Path)
		}
		if command.WriteActionKind != "" {
			return fmt.Errorf("sweep command %q has write action kind without a parity projection", command.Path)
		}
		return nil
	}
	if !validCertificationParityKind(*command.OperationKind) || !validCertificationParityClass(*command.OpClass) {
		return fmt.Errorf("sweep command %q has unsupported parity projection %q/%q", command.Path, *command.OperationKind, *command.OpClass)
	}
	if command.WriteActionKind != "" {
		if !validCertificationWriteActionKind(command.WriteActionKind) {
			return fmt.Errorf("sweep command %q has unsupported write action kind %q", command.Path, command.WriteActionKind)
		}
		if *command.OperationKind != certificationParityKindRESTWrite && *command.OperationKind != certificationParityKindReverseETL {
			return fmt.Errorf("sweep command %q has write action kind outside a write parity projection", command.Path)
		}
	}
	return nil
}

func certificationSweepIntentRequiresParity(intent string) bool {
	switch intent {
	case "direct_read", "direct_write", "etl", "reverse_etl", "binary_download",
		"capability_read", "capability_write", "capability_cdc", "changefeed", "transport_source", "transport_destination":
		return true
	default:
		return false
	}
}

func validCertificationSweepStatus(status string) bool {
	switch status {
	case certificationSweepEligiblePendingLive, certificationSweepSchemaConformant, certificationSweepFixtureRequired, certificationSweepStatusProductDefect, certificationSweepProviderRefused, certificationSweepNotApplicable:
		return true
	default:
		return false
	}
}

func marshalCertificationSweep(sweep certificationSweep) ([]byte, error) {
	if err := validateCertificationSweep(sweep); err != nil {
		return nil, err
	}
	raw, err := json.MarshalIndent(sweep, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("encode certification sweep: %w", err)
	}
	return append(raw, '\n'), nil
}

func validateCertificationSweepArtifact(raw []byte) error {
	var sweep certificationSweep
	if err := decodeStrictJSON(raw, &sweep); err != nil {
		return fmt.Errorf("parse: %w", err)
	}
	return validateCertificationSweep(sweep)
}

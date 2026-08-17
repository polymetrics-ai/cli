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
)

// certificationSweep is the deterministic, source-derived accounting record
// for one connector's declared CLI surface. It deliberately holds no provider
// response or credential: a generated candidate becomes a passing result only
// when a separate live proof is accepted.
type certificationSweep struct {
	SchemaVersion    int                                 `json:"schema_version"`
	Connector        string                              `json:"connector"`
	Source           string                              `json:"source"`
	DeclaredCommands int                                 `json:"declared_commands"`
	StatusTotal      int                                 `json:"status_total"`
	Commands         []certificationSweepCommand         `json:"commands"`
	ProductDefects   []certificationSweepProductDefect   `json:"product_defects"`
	ProviderRefusals []certificationSweepProviderRefusal `json:"provider_refusals"`
}

// certificationSweepCommand is one generated certification candidate and its
// current, non-affirmative accounting status. Assertion metadata is a narrow
// overlay from certification.json; its command identity always originates in
// cli_surface.json.
type certificationSweepCommand struct {
	Summary          string                                `json:"summary"`
	Path             string                                `json:"path"`
	Intent           string                                `json:"intent"`
	Availability     string                                `json:"availability"`
	Stream           string                                `json:"stream,omitempty"`
	Flags            []certificationSweepFlag              `json:"flags"`
	APISurface       []engine.CLISurfaceEndpointRef        `json:"api_surface"`
	Status           string                                `json:"status"`
	Reason           string                                `json:"reason"`
	RequiredFlags    []string                              `json:"required_flags,omitempty"`
	OutputAssertions []engine.CertificationOutputAssertion `json:"output_assertions,omitempty"`
	AssertionSource  string                                `json:"assertion_source,omitempty"`
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
		logf(stdout, "connectorgen certification-sweep: %s is current (%d commands)\n", filepath.ToSlash(path), sweep.DeclaredCommands)
		return 0
	}
	if err := writeGeneratedArtifact(path, raw); err != nil {
		logf(stderr, "connectorgen certification-sweep: write %q: %v\n", filepath.ToSlash(path), err)
		return 1
	}
	logf(stdout, "connectorgen certification-sweep: wrote %s (%d commands)\n", filepath.ToSlash(path), sweep.DeclaredCommands)
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
	if bundle.CLISurface == nil {
		return certificationSweep{}, fmt.Errorf("connector %q has no cli_surface.json", connector)
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
	commands := append([]engine.CLICommand(nil), bundle.CLISurface.Commands...)
	sort.Slice(commands, func(i, j int) bool { return commands[i].Path < commands[j].Path })

	sweep := certificationSweep{
		SchemaVersion:    certificationSweepSchemaVersion,
		Connector:        connector,
		Source:           "cli_surface.json",
		DeclaredCommands: len(commands),
		Commands:         make([]certificationSweepCommand, 0, len(commands)),
		ProductDefects:   []certificationSweepProductDefect{},
		ProviderRefusals: providerRefusals,
	}
	seen := make(map[string]bool, len(commands))
	for _, command := range commands {
		if strings.TrimSpace(command.Path) == "" || seen[command.Path] {
			return certificationSweep{}, fmt.Errorf("connector %q cli_surface has missing or duplicate command path %q", connector, command.Path)
		}
		seen[command.Path] = true
		row, defect := classifyCertificationSweepCommand(command, operations[command.Operation], assertions[command.Path], graphql, refusalByCommand[command.Path])
		sweep.Commands = append(sweep.Commands, row)
		if defect != nil {
			sweep.ProductDefects = append(sweep.ProductDefects, *defect)
		}
	}
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
	sweep.StatusTotal = len(sweep.Commands)
	if err := validateCertificationSweep(sweep); err != nil {
		return certificationSweep{}, err
	}
	return sweep, nil
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
			Source:     "certification.json direct_read_candidates",
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

func classifyCertificationSweepCommand(command engine.CLICommand, operation engine.OperationSpec, assertion certificationSweepAssertionOverlay, graphql certificationSweepGraphQLProfile, providerRefusal certificationSweepProviderRefusal) (certificationSweepCommand, *certificationSweepProductDefect) {
	row := certificationSweepCommand{
		Summary:       command.Summary,
		Path:          command.Path,
		Intent:        command.Intent,
		Availability:  command.Availability,
		Stream:        command.Stream,
		Flags:         certificationSweepFlags(command.Flags),
		APISurface:    append([]engine.CLISurfaceEndpointRef(nil), command.APISurface...),
		RequiredFlags: certificationSweepRequiredFlags(command.Flags),
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
	if sweep.SchemaVersion != certificationSweepSchemaVersion || !isSafeProofIdentifier(sweep.Connector) || sweep.Source != "cli_surface.json" {
		return errors.New("sweep requires supported schema, safe connector, and cli_surface.json source")
	}
	if sweep.DeclaredCommands <= 0 || sweep.DeclaredCommands != len(sweep.Commands) || sweep.StatusTotal != len(sweep.Commands) {
		return errors.New("sweep command totals do not reconcile")
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

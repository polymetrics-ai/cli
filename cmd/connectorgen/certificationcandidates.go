package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

const certificationCandidatesFile = "certification.json"

// readCandidateGeneration is the narrow source-owned input used to derive
// direct-read candidates. Endpoint and command data always comes from the CLI
// surface; this shape contains only fixture values and auditable membership.
type readCandidateGeneration struct {
	RequiredFlagDefaults map[string]string
	Cohorts              []engine.CertificationReadCandidateCohort
}

// buildGeneratedReadCandidates derives serially executable direct-read
// candidates from declared command paths and required flags. It never promotes
// a candidate to a pass and deliberately excludes write, ETL, and unavailable
// commands: those require their own lifecycle contracts.
func buildGeneratedReadCandidates(connector string, commands []engine.CLICommand, generation readCandidateGeneration) ([]engine.CertificationCommandCandidate, error) {
	cohortByCommand := make(map[string]string)
	for _, cohort := range generation.Cohorts {
		if strings.TrimSpace(cohort.Name) == "" {
			return nil, fmt.Errorf("read candidate cohort has an empty name")
		}
		for _, path := range cohort.Commands {
			if prior, exists := cohortByCommand[path]; exists {
				return nil, fmt.Errorf("read candidate command %q belongs to cohorts %q and %q", path, prior, cohort.Name)
			}
			cohortByCommand[path] = cohort.Name
		}
	}

	byPath := make(map[string]engine.CLICommand, len(commands))
	for _, command := range commands {
		if strings.TrimSpace(command.Path) == "" {
			return nil, fmt.Errorf("cli surface contains a command with an empty path")
		}
		if _, exists := byPath[command.Path]; exists {
			return nil, fmt.Errorf("cli surface duplicates command path %q", command.Path)
		}
		byPath[command.Path] = command
	}
	for path := range cohortByCommand {
		if _, exists := byPath[path]; !exists {
			return nil, fmt.Errorf("read candidate cohort command %q is absent from cli_surface.json", path)
		}
	}

	paths := make([]string, 0, len(cohortByCommand))
	for path := range cohortByCommand {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	generated := make([]engine.CertificationCommandCandidate, 0, len(paths))
	for _, path := range paths {
		command := byPath[path]
		if command.Intent != "direct_read" || command.Availability != "implemented" {
			continue
		}
		args := []engine.CertificationCommandArg{{Connector: true}}
		for _, token := range strings.Fields(command.Path) {
			args = append(args, engine.CertificationCommandArg{Literal: token})
		}
		args = append(args,
			engine.CertificationCommandArg{Literal: "--credential"},
			engine.CertificationCommandArg{SourceCredential: true},
		)
		for _, flag := range command.Flags {
			if !flag.Required {
				continue
			}
			defaultValue, found := generation.RequiredFlagDefaults[flag.Name]
			if !found || strings.TrimSpace(defaultValue) == "" {
				return nil, fmt.Errorf("read candidate %q requires --%s without a connector-owned default", command.Path, flag.Name)
			}
			args = append(args,
				engine.CertificationCommandArg{Literal: "--" + flag.Name},
				engine.CertificationCommandArg{ConfigKey: flag.Name, Default: defaultValue},
			)
		}
		args = append(args, engine.CertificationCommandArg{Literal: "--json"})
		generated = append(generated, engine.CertificationCommandCandidate{
			StageName: "generated_direct_read_" + strings.NewReplacer(" ", "_", "-", "_").Replace(command.Path),
			Command:   command.Path,
			Args:      args,
			OutputAssertions: []engine.CertificationOutputAssertion{{
				JSONPointer: "/response",
				ValueType:   "object_or_array",
			}},
			Cohort:    cohortByCommand[path],
			Generated: true,
		})
	}
	return generated, nil
}

// buildGeneratedMutationCandidates derives the declared execution contract for
// mutation commands without invoking a provider, creating a fixture, or
// resolving a credential. Escape classification is connector-owned data; an
// unmatched candidate fails closed to the explicit unassessed classification.
func buildGeneratedMutationCandidates(_ string, commands []engine.CLICommand, operations []engine.OperationSpec, writes []engine.WriteAction, generation engine.CertificationMutationCandidateGeneration) ([]engine.CertificationMutationCandidate, error) {
	intents := make(map[string]struct{}, len(generation.Cohort.Intents))
	for _, intent := range generation.Cohort.Intents {
		intents[intent] = struct{}{}
	}
	operationsByID := make(map[string]engine.OperationSpec, len(operations))
	for _, operation := range operations {
		if strings.TrimSpace(operation.ID) == "" {
			return nil, fmt.Errorf("operations.json contains an operation with an empty id")
		}
		if _, duplicate := operationsByID[operation.ID]; duplicate {
			return nil, fmt.Errorf("operations.json duplicates operation %q", operation.ID)
		}
		operationsByID[operation.ID] = operation
	}
	writesByName := make(map[string]engine.WriteAction, len(writes))
	for _, write := range writes {
		if strings.TrimSpace(write.Name) == "" {
			return nil, fmt.Errorf("writes.json contains an action with an empty name")
		}
		if _, duplicate := writesByName[write.Name]; duplicate {
			return nil, fmt.Errorf("writes.json duplicates action %q", write.Name)
		}
		writesByName[write.Name] = write
	}
	selectedCommands := make(map[string]struct{})
	selectedOperations := make(map[string]struct{})
	selectedWrites := make(map[string]struct{})
	for _, command := range commands {
		if command.Availability != "implemented" {
			continue
		}
		if _, selected := intents[command.Intent]; !selected {
			continue
		}
		selectedCommands[command.Path] = struct{}{}
		switch command.Intent {
		case "direct_write":
			selectedOperations[command.Operation] = struct{}{}
		case "reverse_etl":
			selectedWrites[command.Write] = struct{}{}
		}
	}
	if err := validateMutationClassificationSelectors(generation, selectedCommands, selectedOperations, selectedWrites); err != nil {
		return nil, err
	}

	sorted := append([]engine.CLICommand(nil), commands...)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Path < sorted[j].Path })
	generated := make([]engine.CertificationMutationCandidate, 0, generation.Cohort.CommandCount)
	for _, command := range sorted {
		if command.Availability != "implemented" {
			continue
		}
		if _, selected := intents[command.Intent]; !selected {
			continue
		}
		candidate, err := generatedMutationCandidate(command, operationsByID, writesByName, generation)
		if err != nil {
			return nil, err
		}
		generated = append(generated, candidate)
	}
	if len(generated) != generation.Cohort.CommandCount {
		return nil, fmt.Errorf("mutation candidate cohort %q generated %d commands, want %d", generation.Cohort.Name, len(generated), generation.Cohort.CommandCount)
	}
	deriveMutationFixtureProvisioning(generated)
	sortMutationCandidatesForFixtureLifecycle(generated)
	return generated, nil
}

// validateMutationClassificationSelectors makes family membership auditable:
// a misspelled selector cannot silently fall through to a broad contained
// family. The connector must name only declarations selected by its cohort.
func validateMutationClassificationSelectors(generation engine.CertificationMutationCandidateGeneration, commands, operations, writes map[string]struct{}) error {
	for _, family := range generation.Families {
		for _, selector := range family.Commands {
			if _, found := commands[selector]; !found {
				return fmt.Errorf("mutation containment family %q names command %q outside its declared cohort", family.ID, selector)
			}
		}
		for _, selector := range family.ExcludeCommands {
			if _, found := commands[selector]; !found {
				return fmt.Errorf("mutation containment family %q excludes command %q outside its declared cohort", family.ID, selector)
			}
		}
		for _, selector := range family.Operations {
			if _, found := operations[selector]; !found {
				return fmt.Errorf("mutation containment family %q names operation %q outside its declared cohort", family.ID, selector)
			}
		}
		for _, selector := range family.ExcludeOperations {
			if _, found := operations[selector]; !found {
				return fmt.Errorf("mutation containment family %q excludes operation %q outside its declared cohort", family.ID, selector)
			}
		}
		for _, selector := range family.Writes {
			if _, found := writes[selector]; !found {
				return fmt.Errorf("mutation containment family %q names write action %q outside its declared cohort", family.ID, selector)
			}
		}
		for _, selector := range family.ExcludeWrites {
			if _, found := writes[selector]; !found {
				return fmt.Errorf("mutation containment family %q excludes write action %q outside its declared cohort", family.ID, selector)
			}
		}
	}
	return nil
}

func generatedMutationCandidate(command engine.CLICommand, operations map[string]engine.OperationSpec, writes map[string]engine.WriteAction, generation engine.CertificationMutationCandidateGeneration) (engine.CertificationMutationCandidate, error) {
	candidate := engine.CertificationMutationCandidate{
		Command:        command.Path,
		CommandTokens:  strings.Fields(command.Path),
		Intent:         command.Intent,
		Cohort:         generation.Cohort.Name,
		CredentialFlag: "--credential",
		JSONMode:       mutationJSONMode(),
		RequiredFlags:  mutationRequiredFlags(command.Flags),
		Generated:      true,
	}
	var schema json.RawMessage
	switch command.Intent {
	case "direct_write":
		operation, found := operations[command.Operation]
		if !found {
			return engine.CertificationMutationCandidate{}, fmt.Errorf("mutation command %q references missing operation %q", command.Path, command.Operation)
		}
		candidate.Declaration = engine.CertificationMutationDeclaration{Kind: "operation", ID: operation.ID, Executor: "direct_write"}
		address, err := mutationAddressFromCommand(command, operation, engine.WriteAction{})
		if err != nil {
			return engine.CertificationMutationCandidate{}, err
		}
		candidate.Address = address
		if operation.GraphQL != nil {
			schema = operation.GraphQL.VariablesSchema
		} else if operation.REST != nil {
			schema = operation.REST.BodySchema
		}
	case "reverse_etl":
		write, found := writes[command.Write]
		if !found {
			return engine.CertificationMutationCandidate{}, fmt.Errorf("mutation command %q references missing write action %q", command.Path, command.Write)
		}
		candidate.Declaration = engine.CertificationMutationDeclaration{Kind: "write_action", ID: write.Name, Executor: "reverse_plan"}
		address, err := mutationAddressFromCommand(command, engine.OperationSpec{}, write)
		if err != nil {
			return engine.CertificationMutationCandidate{}, err
		}
		candidate.Address = address
		schema = write.RecordSchema
	default:
		return engine.CertificationMutationCandidate{}, fmt.Errorf("mutation command %q has unsupported intent %q", command.Path, command.Intent)
	}
	slots, err := mutationInputSlots(command.Flags, schema, command.Intent)
	if err != nil {
		return engine.CertificationMutationCandidate{}, fmt.Errorf("mutation command %q input slots: %w", command.Path, err)
	}
	candidate.InputSlots = slots
	classification, err := classifyGeneratedMutationCandidate(command, candidate.Declaration, generation)
	if err != nil {
		return engine.CertificationMutationCandidate{}, err
	}
	candidate.Classification = classification
	return candidate, nil
}

func mutationJSONMode() *bool {
	value := true
	return &value
}

func mutationAddressFromCommand(command engine.CLICommand, operation engine.OperationSpec, write engine.WriteAction) (engine.CertificationMutationAddress, error) {
	if len(command.APISurface) == 1 {
		address := command.APISurface[0]
		if strings.TrimSpace(address.Method) == "" || strings.TrimSpace(address.Path) == "" {
			return engine.CertificationMutationAddress{}, fmt.Errorf("mutation command %q has an incomplete cli_surface address", command.Path)
		}
		return engine.CertificationMutationAddress{Source: "cli_surface", Transport: mutationAddressTransport(operation, command.Intent), Method: address.Method, Path: address.Path}, nil
	}
	if len(command.APISurface) > 1 {
		return engine.CertificationMutationAddress{}, fmt.Errorf("mutation command %q has %d cli_surface addresses", command.Path, len(command.APISurface))
	}
	if command.Intent == "reverse_etl" && strings.TrimSpace(write.Method) != "" && strings.TrimSpace(write.Path) != "" {
		return engine.CertificationMutationAddress{Source: "write_action", Transport: "rest", Method: write.Method, Path: write.Path}, nil
	}
	if command.Intent == "direct_write" && operation.REST != nil && strings.TrimSpace(operation.REST.Method) != "" && strings.TrimSpace(operation.REST.Path) != "" {
		return engine.CertificationMutationAddress{Source: "operation", Transport: "rest", Method: operation.REST.Method, Path: operation.REST.Path}, nil
	}
	if command.Intent == "direct_write" && operation.GraphQL != nil && strings.TrimSpace(operation.GraphQL.Path) != "" {
		return engine.CertificationMutationAddress{Source: "operation", Transport: "graphql", Method: "POST", Path: operation.GraphQL.Path}, nil
	}
	return engine.CertificationMutationAddress{}, fmt.Errorf("mutation command %q has no declared address", command.Path)
}

func mutationAddressTransport(operation engine.OperationSpec, intent string) string {
	if intent == "direct_write" && operation.GraphQL != nil {
		return "graphql"
	}
	return "rest"
}

// deriveMutationFixtureProvisioning turns the declared REST endpoint tree
// into a fixture plan. A POST or PUT on a collection is the provisioner for
// every candidate on that collection; a terminal path parameter identifies an
// item inside the same collection and is intentionally removed before matching.
//
// GraphQL's shared transport endpoint is not a resource collection, and an
// alias without cli_surface evidence must not borrow a lifecycle from a raw
// write action. Both become explicit named exceptions for later work rather
// than a permissive long-lived-fixture fallback.
func deriveMutationFixtureProvisioning(candidates []engine.CertificationMutationCandidate) {
	provisioners := make(map[string][]string)
	for _, candidate := range candidates {
		if !mutationCandidateHasRESTCollection(candidate) {
			continue
		}
		collection := mutationCollectionPath(candidate.Address.Path)
		if mutationAddressIsProvisioner(candidate.Address.Method) {
			provisioners[collection] = append(provisioners[collection], candidate.Command)
		}
	}
	for collection := range provisioners {
		sort.Strings(provisioners[collection])
	}

	for index := range candidates {
		candidate := &candidates[index]
		if candidate.Address.Transport == "graphql" {
			candidate.Fixture = engine.CertificationMutationFixtureProvisioning{
				Strategy:      "named_exception",
				ExceptionCode: "graphql_transport_not_collection",
				Evidence:      "the shared GraphQL transport endpoint is not a resource collection, so no REST CRUD fixture cycle can be inferred",
			}
			continue
		}
		if candidate.Address.Source != "cli_surface" {
			candidate.Fixture = engine.CertificationMutationFixtureProvisioning{
				Strategy:      "named_exception",
				ExceptionCode: "missing_api_surface",
				Evidence:      "the command has no cli_surface API address, so no declared collection cycle can be derived",
			}
			continue
		}
		collection := mutationCollectionPath(candidate.Address.Path)
		creators := provisioners[collection]
		if len(creators) == 0 {
			candidate.Fixture = engine.CertificationMutationFixtureProvisioning{
				Strategy:      "named_exception",
				ExceptionCode: "collection_without_creator",
				Evidence:      fmt.Sprintf("no POST or PUT declaration shares the derived REST collection %q", collection),
			}
			continue
		}
		candidate.Fixture = engine.CertificationMutationFixtureProvisioning{
			Strategy:            "derived_collection_cycle",
			Collection:          collection,
			CollectionDepth:     mutationCollectionDepth(collection),
			ProvisionerCommands: append([]string(nil), creators...),
			Evidence:            fmt.Sprintf("POST or PUT declaration(s) %s share the derived REST collection %q; URL depth %d orders parent collections first", strings.Join(creators, ", "), collection, mutationCollectionDepth(collection)),
		}
	}
}

func mutationCandidateHasRESTCollection(candidate engine.CertificationMutationCandidate) bool {
	return candidate.Address.Source == "cli_surface" && candidate.Address.Transport == "rest" && strings.TrimSpace(candidate.Address.Path) != ""
}

func mutationAddressIsProvisioner(method string) bool {
	method = strings.ToUpper(strings.TrimSpace(method))
	return method == "POST" || method == "PUT"
}

func mutationCollectionPath(path string) string {
	segments := strings.FieldsFunc(strings.TrimSpace(path), func(r rune) bool { return r == '/' })
	if len(segments) > 0 && mutationPathParameter(segments[len(segments)-1]) {
		segments = segments[:len(segments)-1]
	}
	return "/" + strings.Join(segments, "/")
}

func mutationPathParameter(segment string) bool {
	return len(segment) > 2 && strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}")
}

func mutationCollectionDepth(collection string) int {
	return len(strings.FieldsFunc(strings.TrimSpace(collection), func(r rune) bool { return r == '/' }))
}

// sortMutationCandidatesForFixtureLifecycle makes the future execution order
// explicit in the generated artifact: a parent collection always precedes a
// child collection. Named exceptions sort after derivable cycles and cannot be
// mistaken for a runnable fixture plan.
func sortMutationCandidatesForFixtureLifecycle(candidates []engine.CertificationMutationCandidate) {
	sort.Slice(candidates, func(i, j int) bool {
		left, right := candidates[i], candidates[j]
		leftDerived := left.Fixture.Strategy == "derived_collection_cycle"
		rightDerived := right.Fixture.Strategy == "derived_collection_cycle"
		if leftDerived != rightDerived {
			return leftDerived
		}
		if leftDerived {
			if left.Fixture.CollectionDepth != right.Fixture.CollectionDepth {
				return left.Fixture.CollectionDepth < right.Fixture.CollectionDepth
			}
			if left.Fixture.Collection != right.Fixture.Collection {
				return left.Fixture.Collection < right.Fixture.Collection
			}
		}
		if left.Fixture.ExceptionCode != right.Fixture.ExceptionCode {
			return left.Fixture.ExceptionCode < right.Fixture.ExceptionCode
		}
		return left.Command < right.Command
	})
}

func mutationInputSlots(flags []engine.CLIFlag, schema json.RawMessage, intent string) ([]engine.CertificationMutationInputSlot, error) {
	byPath := map[string]engine.CertificationMutationInputSlot{}
	for _, flag := range flags {
		path := flag.MapsTo
		if path == "" {
			path = "flag." + flag.Name
		}
		mergeMutationInputSlot(byPath, engine.CertificationMutationInputSlot{
			Path: path, Type: flag.Type, Required: flag.Required, Values: append([]string(nil), flag.Values...),
		})
	}
	if len(schema) != 0 {
		prefix := "body."
		if intent == "reverse_etl" {
			prefix = "record."
		}
		var contract struct {
			Required   []string                   `json:"required"`
			Properties map[string]json.RawMessage `json:"properties"`
		}
		if err := json.Unmarshal(schema, &contract); err != nil {
			return nil, fmt.Errorf("decode declared input schema: %w", err)
		}
		for _, name := range contract.Required {
			typeName := "json"
			if raw, found := contract.Properties[name]; found {
				typeName = mutationSchemaType(raw)
			}
			mergeMutationInputSlot(byPath, engine.CertificationMutationInputSlot{Path: prefix + name, Type: typeName, Required: true})
		}
	}
	paths := make([]string, 0, len(byPath))
	for path := range byPath {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	slots := make([]engine.CertificationMutationInputSlot, 0, len(paths))
	for _, path := range paths {
		slots = append(slots, byPath[path])
	}
	return slots, nil
}

func mergeMutationInputSlot(slots map[string]engine.CertificationMutationInputSlot, next engine.CertificationMutationInputSlot) {
	if current, found := slots[next.Path]; found {
		current.Required = current.Required || next.Required
		if current.Type == "" {
			current.Type = next.Type
		}
		if len(current.Values) == 0 && len(next.Values) != 0 {
			current.Values = next.Values
		}
		slots[next.Path] = current
		return
	}
	slots[next.Path] = next
}

func mutationSchemaType(raw json.RawMessage) string {
	var schema struct {
		Type any `json:"type"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil {
		return "json"
	}
	if value, ok := schema.Type.(string); ok && value != "" {
		return value
	}
	return "json"
}

func mutationRequiredFlags(flags []engine.CLIFlag) []string {
	required := make([]string, 0, len(flags))
	for _, flag := range flags {
		if flag.Required {
			required = append(required, "--"+flag.Name)
		}
	}
	sort.Strings(required)
	return required
}

func classifyGeneratedMutationCandidate(command engine.CLICommand, declaration engine.CertificationMutationDeclaration, generation engine.CertificationMutationCandidateGeneration) (engine.CertificationMutationClassification, error) {
	var matched *engine.CertificationMutationClassificationFamily
	for index := range generation.Families {
		family := &generation.Families[index]
		if !mutationFamilyMatches(*family, command, declaration) {
			continue
		}
		if matched != nil {
			return engine.CertificationMutationClassification{}, fmt.Errorf("mutation command %q matches containment families %q and %q", command.Path, matched.ID, family.ID)
		}
		matched = family
	}
	if matched == nil {
		classification := generation.Unassessed
		classification.Family = "unassessed"
		return classification, nil
	}
	classification := matched.Classification
	classification.Family = matched.ID
	return classification, nil
}

func mutationFamilyMatches(family engine.CertificationMutationClassificationFamily, command engine.CLICommand, declaration engine.CertificationMutationDeclaration) bool {
	if containsMutationSelector(family.ExcludeCommands, command.Path) || containsMutationSelector(family.ExcludeOperations, declaration.ID) || containsMutationSelector(family.ExcludeWrites, declaration.ID) {
		return false
	}
	return containsMutationSelector(family.Commands, command.Path) ||
		containsMutationSelector(family.Operations, declaration.ID) ||
		containsMutationSelector(family.Writes, declaration.ID) ||
		containsMutationSelector(family.Intents, command.Intent)
}

func containsMutationSelector(values []string, value string) bool {
	for _, candidate := range values {
		if candidate == value {
			return true
		}
	}
	return false
}

func runCertificationCandidates(args []string, stdout, stderr io.Writer) int {
	root := "."
	connector := ""
	check := false
	for index := 1; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--check":
			check = true
		case strings.HasPrefix(arg, "--connector="):
			connector = strings.TrimPrefix(arg, "--connector=")
		case arg == "--connector":
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				logln(stderr, "connectorgen certification-candidates: --connector requires a name")
				return 2
			}
			index++
			connector = args[index]
		case strings.HasPrefix(arg, "-"):
			logf(stderr, "connectorgen certification-candidates: unknown flag %q\n", arg)
			return 2
		case root == ".":
			root = arg
		default:
			logf(stderr, "connectorgen certification-candidates: unexpected extra argument %q\n", arg)
			return 2
		}
	}
	if strings.TrimSpace(connector) == "" {
		logln(stderr, "connectorgen certification-candidates: --connector requires a name")
		return 2
	}
	if code := generateCertificationCandidates(root, connector, check); code != nil {
		logf(stderr, "connectorgen certification-candidates: %v\n", code)
		return 1
	}
	if check {
		logf(stdout, "certification candidates are current: connector=%s\n", connector)
	} else {
		logf(stdout, "generated certification candidates: connector=%s\n", connector)
	}
	return 0
}

func generateCertificationCandidates(root, connector string, check bool) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	definitionsRoot := filepath.Join(absRoot, "internal", "connectors", "defs")
	bundle, err := engine.Load(os.DirFS(definitionsRoot), connector)
	if err != nil {
		return fmt.Errorf("load connector %q: %w", connector, err)
	}
	if bundle.Certification == nil {
		return fmt.Errorf("connector %q has no certification.json", connector)
	}
	if bundle.Certification.DirectReadGeneration == nil && bundle.Certification.MutationGeneration == nil {
		return fmt.Errorf("connector %q has no candidate generation declaration", connector)
	}
	if bundle.CLISurface == nil {
		return fmt.Errorf("connector %q has no cli_surface.json", connector)
	}
	if generation := bundle.Certification.DirectReadGeneration; generation != nil {
		generated, err := buildGeneratedReadCandidates(connector, bundle.CLISurface.Commands, readCandidateGeneration{
			RequiredFlagDefaults: generation.RequiredFlagDefaults,
			Cohorts:              generation.Cohorts,
		})
		if err != nil {
			return err
		}
		bundle.Certification.DirectReadCandidates, err = mergeGeneratedReadCandidates(bundle.Certification.DirectReadCandidates, generated)
		if err != nil {
			return err
		}
	}
	if generation := bundle.Certification.MutationGeneration; generation != nil {
		generated, err := buildGeneratedMutationCandidates(connector, bundle.CLISurface.Commands, bundle.Operations, bundle.Writes, *generation)
		if err != nil {
			return err
		}
		bundle.Certification.MutationCandidates, err = mergeGeneratedMutationCandidates(bundle.Certification.MutationCandidates, generated)
		if err != nil {
			return err
		}
	}
	raw, err := json.MarshalIndent(bundle.Certification, "", "  ")
	if err != nil {
		return fmt.Errorf("render certification candidates: %w", err)
	}
	raw = append(raw, '\n')
	path := filepath.Join(definitionsRoot, connector, certificationCandidatesFile)
	if check {
		committed, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read certification candidates: %w", err)
		}
		if !bytes.Equal(committed, raw) {
			return fmt.Errorf("certification candidates are stale; run `connectorgen certification-candidates --connector %s`", connector)
		}
		return nil
	}
	if err := writeGeneratedArtifact(path, raw); err != nil {
		return fmt.Errorf("write certification candidates: %w", err)
	}
	return nil
}

// mergeGeneratedReadCandidates leaves explicitly authored candidates intact.
// A manual command shadows a generated command because only its author can
// supply a more specific produced-value assertion or argument shape.
func mergeGeneratedReadCandidates(existing, generated []engine.CertificationCommandCandidate) ([]engine.CertificationCommandCandidate, error) {
	manual := make([]engine.CertificationCommandCandidate, 0, len(existing))
	manualCommands := make(map[string]struct{})
	for _, candidate := range existing {
		if candidate.Generated {
			continue
		}
		if _, duplicate := manualCommands[candidate.Command]; duplicate {
			return nil, fmt.Errorf("manual direct-read candidate command %q is duplicated", candidate.Command)
		}
		manualCommands[candidate.Command] = struct{}{}
		manual = append(manual, candidate)
	}
	filtered := generated[:0]
	for _, candidate := range generated {
		if _, manual := manualCommands[candidate.Command]; manual {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return append(manual, filtered...), nil
}

// mergeGeneratedMutationCandidates has the same bounded-override contract as
// read candidates: a manual row must name its exact command and reason, and
// can shadow only that one generated command.
func mergeGeneratedMutationCandidates(existing, generated []engine.CertificationMutationCandidate) ([]engine.CertificationMutationCandidate, error) {
	manual := make([]engine.CertificationMutationCandidate, 0, len(existing))
	manualCommands := make(map[string]struct{})
	generatedCommands := make(map[string]struct{}, len(generated))
	for _, candidate := range generated {
		if _, duplicate := generatedCommands[candidate.Command]; duplicate {
			return nil, fmt.Errorf("generated mutation candidate command %q is duplicated", candidate.Command)
		}
		generatedCommands[candidate.Command] = struct{}{}
	}
	for _, candidate := range existing {
		if candidate.Generated {
			continue
		}
		if strings.TrimSpace(candidate.OverrideReason) == "" {
			return nil, fmt.Errorf("manual mutation candidate command %q has no override_reason", candidate.Command)
		}
		if _, duplicate := manualCommands[candidate.Command]; duplicate {
			return nil, fmt.Errorf("manual mutation candidate command %q is duplicated", candidate.Command)
		}
		if _, known := generatedCommands[candidate.Command]; !known {
			return nil, fmt.Errorf("manual mutation candidate command %q does not shadow a generated command", candidate.Command)
		}
		manualCommands[candidate.Command] = struct{}{}
		manual = append(manual, candidate)
	}
	filtered := generated[:0]
	for _, candidate := range generated {
		if _, manual := manualCommands[candidate.Command]; manual {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return append(manual, filtered...), nil
}

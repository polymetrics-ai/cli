package certify

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

// GraphQLCertificationCommandResult is one declaration-owned GraphQL command
// accounting row. Schema conformance is intentionally distinct from a pass:
// it proves the fixed operation still agrees with its source-pinned schema,
// not that a provider produced the asserted value.
type GraphQLCertificationCommandResult struct {
	Command   string `json:"command"`
	Operation string `json:"operation"`
	Result    string `json:"result"`
	Reason    string `json:"reason"`
}

// GraphQLCertificationResult records the bounded GraphQL certification
// boundary. Commands that were not sent keep a concrete non-pass result so an
// aggregate cannot be mistaken for live provider evidence.
type GraphQLCertificationResult struct {
	Result           string                              `json:"result"`
	SchemaConformant int                                 `json:"schema_conformant"`
	LiveRequired     int                                 `json:"live_required"`
	FixtureBound     int                                 `json:"fixture_bound"`
	Commands         []GraphQLCertificationCommandResult `json:"commands"`
	Reason           string                              `json:"reason,omitempty"`
}

type graphQLCertificationInventory struct {
	SchemaConformant int
	LiveRequired     int
	FixtureBound     int
	commands         []GraphQLCertificationCommandResult
	live             map[string]directReadCandidate
	liveOrder        []string
}

func (inventory graphQLCertificationInventory) liveCandidate(command string) (directReadCandidate, bool) {
	candidate, found := inventory.live[command]
	return candidate, found
}

type graphQLSourceLock struct {
	GraphQL graphQLSourceSchema `json:"graphql"`
}

// graphQLSourceSchema is the connector-owned, source-pinned compact schema
// lock format. It is deliberately not a second query language or an executor:
// it supplies only the root contracts needed to compare fixed operations.
type graphQLSourceSchema struct {
	SourceURL      string                     `json:"source_url"`
	SHA256         string                     `json:"sha256"`
	Bytes          int                        `json:"bytes"`
	QueryFields    []graphQLSourceRootField   `json:"query_fields"`
	MutationFields []graphQLSourceRootField   `json:"mutation_fields"`
	TypeSystem     map[string]json.RawMessage `json:"type_system"`
}

type graphQLSourceRootField struct {
	Root       string                     `json:"root"`
	Name       string                     `json:"name"`
	Arguments  []graphQLSourceArgument    `json:"arguments"`
	ReturnType graphQLSourceTypeReference `json:"return_type"`
}

type graphQLSourceArgument struct {
	Name string                     `json:"name"`
	Type graphQLSourceTypeReference `json:"type"`
}

type graphQLSourceTypeReference struct {
	Kind    string                      `json:"kind"`
	Name    string                      `json:"name"`
	OfType  *graphQLSourceTypeReference `json:"of_type,omitempty"`
	NonNull bool                        `json:"non_null"`
}

type compiledGraphQLSourceSchema struct {
	queryRoots    map[string]graphQLSourceRootField
	mutationRoots map[string]graphQLSourceRootField
}

// graphQLCertificationInventoryFor compiles the connector's source-pinned
// root schema and compares every declared command inside its definition-owned
// command prefix. It never fetches a schema or contacts a provider.
func graphQLCertificationInventoryFor(connector string) (graphQLCertificationInventory, error) {
	bundle, err := engine.Load(defs.FS, connector)
	if err != nil {
		return graphQLCertificationInventory{}, fmt.Errorf("load connector definition: %w", err)
	}
	if bundle.Certification == nil || bundle.Certification.GraphQL == nil {
		return graphQLCertificationInventory{}, fmt.Errorf("connector has no declared GraphQL certification profile")
	}
	if bundle.CLISurface == nil {
		return graphQLCertificationInventory{}, fmt.Errorf("connector has no declared CLI surface")
	}
	profile := bundle.Certification.GraphQL
	raw, err := fs.ReadFile(defs.FS, path.Join(connector, profile.SourceLock))
	if err != nil {
		return graphQLCertificationInventory{}, fmt.Errorf("read declared GraphQL source lock: %w", err)
	}
	schema, err := compileGraphQLSourceSchema(raw)
	if err != nil {
		return graphQLCertificationInventory{}, fmt.Errorf("compile declared GraphQL source lock: %w", err)
	}

	operations := make(map[string]engine.OperationSpec, len(bundle.Operations))
	for _, operation := range bundle.Operations {
		operations[operation.ID] = operation
	}
	live := make(map[string]directReadCandidate, len(profile.LiveCandidates))
	for _, candidate := range profile.LiveCandidates {
		if _, duplicate := live[candidate.Command]; duplicate {
			return graphQLCertificationInventory{}, fmt.Errorf("declared GraphQL live candidates duplicate command %q", candidate.Command)
		}
		live[candidate.Command] = commandCandidateFor(connector, nil, candidate)
	}

	commands := append([]engine.CLICommand(nil), bundle.CLISurface.Commands...)
	sort.Slice(commands, func(i, j int) bool { return commands[i].Path < commands[j].Path })
	inventory := graphQLCertificationInventory{live: live}
	seenLive := make(map[string]bool, len(live))
	for _, command := range commands {
		if !strings.HasPrefix(command.Path, profile.CommandPrefix) {
			continue
		}
		operation, found := operations[command.Operation]
		if !found {
			return graphQLCertificationInventory{}, fmt.Errorf("GraphQL command %q references absent operation %q", command.Path, command.Operation)
		}
		if command.Availability != "implemented" {
			return graphQLCertificationInventory{}, fmt.Errorf("GraphQL command %q is not implemented", command.Path)
		}
		result, err := graphQLSchemaConformance(command, operation, schema)
		if err != nil {
			return graphQLCertificationInventory{}, err
		}
		if candidate, isLive := live[command.Path]; isLive {
			if operation.Kind != "graphql_query" || command.Intent != "direct_read" {
				return graphQLCertificationInventory{}, fmt.Errorf("GraphQL live candidate %q is not a fixed direct-read query", command.Path)
			}
			if candidate.Command != command.Path {
				return graphQLCertificationInventory{}, fmt.Errorf("GraphQL live candidate %q does not match its command", command.Path)
			}
			result.Result = "pending_live"
			result.Reason = "declaration-owned produced-value assertion requires a live read-only provider execution"
			inventory.LiveRequired++
			seenLive[command.Path] = true
		} else if operation.Kind == "graphql_query" {
			result.Result = "schema_conformant"
			result.Reason = profile.SchemaConformantReason
			inventory.SchemaConformant++
		} else {
			result.Result = "fixture_required"
			result.Reason = profile.FixtureRequiredReason
			inventory.FixtureBound++
		}
		inventory.commands = append(inventory.commands, result)
	}
	for command := range live {
		if !seenLive[command] {
			return graphQLCertificationInventory{}, fmt.Errorf("GraphQL live candidate %q is outside declared command prefix %q", command, profile.CommandPrefix)
		}
	}
	for command := range live {
		inventory.liveOrder = append(inventory.liveOrder, command)
	}
	sort.Strings(inventory.liveOrder)
	if len(inventory.commands) == 0 {
		return graphQLCertificationInventory{}, fmt.Errorf("declared GraphQL command prefix %q selects no CLI commands", profile.CommandPrefix)
	}
	return inventory, nil
}

func compileGraphQLSourceSchema(raw []byte) (compiledGraphQLSourceSchema, error) {
	var lock graphQLSourceLock
	if err := json.Unmarshal(raw, &lock); err != nil {
		return compiledGraphQLSourceSchema{}, fmt.Errorf("invalid JSON: %w", err)
	}
	source := lock.GraphQL
	if !strings.HasPrefix(source.SourceURL, "https://") || len(source.SHA256) != 64 || !lowerHex(source.SHA256) || source.Bytes <= 0 {
		return compiledGraphQLSourceSchema{}, fmt.Errorf("source provenance must declare HTTPS URL, SHA-256, and positive byte count")
	}
	typeNames, err := graphQLSourceTypeNames(source.TypeSystem)
	if err != nil {
		return compiledGraphQLSourceSchema{}, err
	}
	compiled := compiledGraphQLSourceSchema{
		queryRoots:    make(map[string]graphQLSourceRootField, len(source.QueryFields)),
		mutationRoots: make(map[string]graphQLSourceRootField, len(source.MutationFields)),
	}
	if err := addGraphQLSourceRoots(compiled.queryRoots, source.QueryFields, typeNames); err != nil {
		return compiledGraphQLSourceSchema{}, fmt.Errorf("query roots: %w", err)
	}
	if err := addGraphQLSourceRoots(compiled.mutationRoots, source.MutationFields, typeNames); err != nil {
		return compiledGraphQLSourceSchema{}, fmt.Errorf("mutation roots: %w", err)
	}
	if len(compiled.queryRoots) == 0 || len(compiled.mutationRoots) == 0 {
		return compiledGraphQLSourceSchema{}, fmt.Errorf("source lock must declare query and mutation roots")
	}
	return compiled, nil
}

func lowerHex(value string) bool {
	for _, character := range value {
		if character < '0' || (character > '9' && character < 'a') || character > 'f' {
			return false
		}
	}
	return true
}

func graphQLSourceTypeNames(typeSystem map[string]json.RawMessage) (map[string]bool, error) {
	if len(typeSystem) == 0 {
		return nil, fmt.Errorf("source lock has no type system")
	}
	names := make(map[string]bool)
	for kind, raw := range typeSystem {
		var entries []json.RawMessage
		if err := json.Unmarshal(raw, &entries); err != nil || len(entries) == 0 {
			return nil, fmt.Errorf("type system kind %q is not a non-empty array", kind)
		}
		for _, entry := range entries {
			var name string
			if err := json.Unmarshal(entry, &name); err == nil {
				if name == "" {
					return nil, fmt.Errorf("type system kind %q contains an empty name", kind)
				}
				names[name] = true
				continue
			}
			var object struct {
				Name string `json:"name"`
			}
			if err := json.Unmarshal(entry, &object); err != nil || object.Name == "" {
				return nil, fmt.Errorf("type system kind %q contains an unnamed type", kind)
			}
			names[object.Name] = true
		}
	}
	return names, nil
}

func addGraphQLSourceRoots(destination map[string]graphQLSourceRootField, fields []graphQLSourceRootField, typeNames map[string]bool) error {
	for _, field := range fields {
		if strings.TrimSpace(field.Root) == "" || strings.TrimSpace(field.Name) == "" {
			return fmt.Errorf("root field must declare root and name")
		}
		returnName := graphQLNamedType(field.ReturnType)
		if returnName == "" || (!typeNames[returnName] && !graphQLBuiltInScalar(returnName)) {
			return fmt.Errorf("root field %q returns an absent type", field.Name)
		}
		if _, duplicate := destination[field.Name]; duplicate {
			return fmt.Errorf("duplicate root field %q", field.Name)
		}
		arguments := make(map[string]bool, len(field.Arguments))
		for _, argument := range field.Arguments {
			if strings.TrimSpace(argument.Name) == "" || graphQLNamedType(argument.Type) == "" {
				return fmt.Errorf("root field %q has an invalid argument", field.Name)
			}
			if arguments[argument.Name] {
				return fmt.Errorf("root field %q has duplicate argument %q", field.Name, argument.Name)
			}
			arguments[argument.Name] = true
		}
		destination[field.Name] = field
	}
	return nil
}

func graphQLBuiltInScalar(name string) bool {
	switch name {
	case "Boolean", "Float", "ID", "Int", "String":
		return true
	default:
		return false
	}
}

func graphQLNamedType(reference graphQLSourceTypeReference) string {
	if reference.Name != "" {
		return reference.Name
	}
	if reference.OfType != nil {
		return graphQLNamedType(*reference.OfType)
	}
	return ""
}

func graphQLSchemaConformance(command engine.CLICommand, operation engine.OperationSpec, schema compiledGraphQLSourceSchema) (GraphQLCertificationCommandResult, error) {
	if operation.GraphQL == nil {
		return GraphQLCertificationCommandResult{}, fmt.Errorf("GraphQL command %q has no fixed GraphQL operation", command.Path)
	}
	var roots map[string]graphQLSourceRootField
	switch operation.Kind {
	case "graphql_query":
		if command.Intent != "direct_read" {
			return GraphQLCertificationCommandResult{}, fmt.Errorf("GraphQL query command %q has intent %q, want direct_read", command.Path, command.Intent)
		}
		roots = schema.queryRoots
	case "graphql_mutation":
		if command.Intent != "direct_write" {
			return GraphQLCertificationCommandResult{}, fmt.Errorf("GraphQL mutation command %q has intent %q, want direct_write", command.Path, command.Intent)
		}
		roots = schema.mutationRoots
	default:
		return GraphQLCertificationCommandResult{}, fmt.Errorf("command %q maps to non-GraphQL operation kind %q", command.Path, operation.Kind)
	}
	root, err := fixedGraphQLRootSelection(operation.GraphQL.Document)
	if err != nil {
		return GraphQLCertificationCommandResult{}, fmt.Errorf("GraphQL command %q fixed document: %w", command.Path, err)
	}
	field, found := roots[root]
	if !found {
		return GraphQLCertificationCommandResult{}, fmt.Errorf("GraphQL command %q root field %q is absent from its source-pinned schema", command.Path, root)
	}
	argumentVariables, err := fixedGraphQLRootArgumentVariables(operation.GraphQL.Document)
	if err != nil {
		return GraphQLCertificationCommandResult{}, fmt.Errorf("GraphQL command %q root field %q: %w", command.Path, root, err)
	}
	if err := graphQLVariablesMatchSource(operation.GraphQL.VariablesSchema, field.Arguments, argumentVariables); err != nil {
		return GraphQLCertificationCommandResult{}, fmt.Errorf("GraphQL command %q root field %q: %w", command.Path, root, err)
	}
	return GraphQLCertificationCommandResult{Command: command.Path, Operation: operation.ID}, nil
}

// fixedGraphQLRootSelection reads the first root field after a generated fixed
// operation's selection brace. Engine admission has already rejected mixed or
// caller-authored documents; this deliberately bounded reader does not grow
// into a general GraphQL parser.
func fixedGraphQLRootSelection(document string) (string, error) {
	offset, err := fixedGraphQLFirstSelectionOffset(document)
	if err != nil {
		return "", err
	}
	name, _, err := fixedGraphQLSelectionNameAt(document, offset)
	return name, err
}

func fixedGraphQLFirstSelectionOffset(document string) (int, error) {
	depth := 0
	inString := false
	escaped := false
	for index := 0; index < len(document); index++ {
		character := document[index]
		if inString {
			if escaped {
				escaped = false
				continue
			}
			switch character {
			case '\\':
				escaped = true
			case '"':
				inString = false
			}
			continue
		}
		if character == '"' {
			inString = true
			continue
		}
		switch character {
		case '(':
			depth++
		case ')':
			depth--
		case '{':
			if depth == 0 {
				return index + 1, nil
			}
		}
	}
	return 0, fmt.Errorf("has no root selection")
}

func fixedGraphQLSelectionNameAt(document string, offset int) (string, int, error) {
	for offset < len(document) && strings.ContainsRune(" \t\r\n,", rune(document[offset])) {
		offset++
	}
	start := offset
	for offset < len(document) && ((document[offset] >= 'a' && document[offset] <= 'z') || (document[offset] >= 'A' && document[offset] <= 'Z') || (document[offset] >= '0' && document[offset] <= '9') || document[offset] == '_') {
		offset++
	}
	if start == offset {
		return "", 0, fmt.Errorf("root selection has no field name")
	}
	name := document[start:offset]
	for offset < len(document) && strings.ContainsRune(" \t\r\n,", rune(document[offset])) {
		offset++
	}
	if offset < len(document) && document[offset] == ':' {
		return fixedGraphQLSelectionNameAt(document, offset+1)
	}
	return name, offset, nil
}

func fixedGraphQLRootArgumentVariables(document string) (map[string]string, error) {
	offset, err := fixedGraphQLFirstSelectionOffset(document)
	if err != nil {
		return nil, err
	}
	_, offset, err = fixedGraphQLSelectionNameAt(document, offset)
	if err != nil {
		return nil, err
	}
	if offset >= len(document) || document[offset] != '(' {
		return map[string]string{}, nil
	}
	end, err := fixedGraphQLMatchingParenthesis(document, offset)
	if err != nil {
		return nil, err
	}
	arguments := make(map[string]string)
	for offset++; offset < end; {
		for offset < end && strings.ContainsRune(" \t\r\n,", rune(document[offset])) {
			offset++
		}
		if offset == end {
			break
		}
		name, next, err := fixedGraphQLNameAt(document, offset)
		if err != nil {
			return nil, err
		}
		offset = next
		for offset < end && strings.ContainsRune(" \t\r\n", rune(document[offset])) {
			offset++
		}
		if offset >= end || document[offset] != ':' {
			return nil, fmt.Errorf("root argument %q has no value separator", name)
		}
		offset++
		for offset < end && strings.ContainsRune(" \t\r\n", rune(document[offset])) {
			offset++
		}
		if offset >= end || document[offset] != '$' {
			return nil, fmt.Errorf("root argument %q is not bound to a fixed variable", name)
		}
		variable, next, err := fixedGraphQLNameAt(document, offset+1)
		if err != nil {
			return nil, fmt.Errorf("root argument %q has an invalid variable: %w", name, err)
		}
		if _, duplicate := arguments[name]; duplicate {
			return nil, fmt.Errorf("root argument %q is duplicated", name)
		}
		arguments[name] = variable
		offset = next
	}
	return arguments, nil
}

func fixedGraphQLMatchingParenthesis(document string, offset int) (int, error) {
	depth := 0
	for index := offset; index < len(document); index++ {
		switch document[index] {
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return index, nil
			}
		}
	}
	return 0, fmt.Errorf("root argument list is unclosed")
}

func fixedGraphQLNameAt(document string, offset int) (string, int, error) {
	start := offset
	for offset < len(document) && ((document[offset] >= 'a' && document[offset] <= 'z') || (document[offset] >= 'A' && document[offset] <= 'Z') || (document[offset] >= '0' && document[offset] <= '9') || document[offset] == '_') {
		offset++
	}
	if start == offset {
		return "", 0, fmt.Errorf("expected GraphQL name")
	}
	return document[start:offset], offset, nil
}

func graphQLVariablesMatchSource(raw json.RawMessage, arguments []graphQLSourceArgument, selected map[string]string) error {
	var schema struct {
		Type       string                     `json:"type"`
		Properties map[string]json.RawMessage `json:"properties"`
		Required   []string                   `json:"required"`
	}
	if err := json.Unmarshal(raw, &schema); err != nil || schema.Type != "object" || schema.Properties == nil {
		return fmt.Errorf("fixed variables schema is not a closed object")
	}
	seenRequired := make(map[string]bool, len(schema.Required))
	for _, name := range schema.Required {
		seenRequired[name] = true
	}
	if len(schema.Properties) != len(selected) {
		return fmt.Errorf("variables schema property count does not match fixed root arguments")
	}
	for _, argument := range arguments {
		variable, selected := selected[argument.Name]
		if argument.Type.NonNull && !selected {
			return fmt.Errorf("fixed root arguments omit required source argument %q", argument.Name)
		}
		if !selected {
			continue
		}
		if _, found := schema.Properties[variable]; !found {
			return fmt.Errorf("variables schema omits fixed variable %q for source argument %q", variable, argument.Name)
		}
		if argument.Type.NonNull && !seenRequired[variable] {
			return fmt.Errorf("variables schema requiredness does not match source argument %q", argument.Name)
		}
	}
	for argument, variable := range selected {
		found := false
		for _, sourceArgument := range arguments {
			if sourceArgument.Name == argument {
				found = true
				break
			}
		}
		if !found {
			return fmt.Errorf("fixed root argument %q is absent from source schema", argument)
		}
		if _, found := schema.Properties[variable]; !found {
			return fmt.Errorf("variables schema omits fixed variable %q", variable)
		}
	}
	return nil
}

func stageGraphQLSchemaConformance(rc *runContext, rep *Report) error {
	if !rc.opts.Full {
		reason := "skipped: --full not set (GraphQL schema and live-read certification are full-certificate only)"
		rep.Capabilities.GraphQL = &GraphQLCertificationResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "graphql_schema_conformance", reason)
		return nil
	}
	profile := certificationProfileFor(rc.opts.Connector)
	if profile.spec == nil || profile.spec.GraphQL == nil {
		reason := fmt.Sprintf("skipped: connector %q has no definition-owned GraphQL certification profile", rc.opts.Connector)
		rep.Capabilities.GraphQL = &GraphQLCertificationResult{Result: "skipped", Reason: reason}
		skipStage(rc, rep, "graphql_schema_conformance", reason)
		return nil
	}
	inventoryFor := graphQLCertificationInventoryFor
	if rc.graphQLInventory != nil {
		inventoryFor = rc.graphQLInventory
	}
	inventory, err := inventoryFor(rc.opts.Connector)
	if err != nil {
		reason := "declared GraphQL certification capability cannot compile its source-pinned schema: " + err.Error()
		rep.Capabilities.GraphQL = &GraphQLCertificationResult{Result: "unexecutable", Reason: reason}
		unexecutableStage(rc, rep, "graphql_schema_conformance", reason)
		return nil
	}
	result := &GraphQLCertificationResult{
		Result:           "not_live",
		SchemaConformant: inventory.SchemaConformant,
		LiveRequired:     inventory.LiveRequired,
		FixtureBound:     inventory.FixtureBound,
		Commands:         append([]GraphQLCertificationCommandResult(nil), inventory.commands...),
	}
	rep.Capabilities.GraphQL = result
	recordStage(rc, rep, "graphql_schema_conformance", 0, func() (bool, CLIStageInfo, string) {
		return true, CLIStageInfo{ArgvRedacted: "definition-owned GraphQL source lock", ExitCode: 0, Kind: "GraphQLSchemaConformance"}, ""
	})

	allLivePassed := true
	for _, command := range inventory.liveOrder {
		candidate := inventory.live[command]
		stage := recordStage(rc, rep, candidate.StageName, 2, func() (bool, CLIStageInfo, string) {
			res := rc.run(candidate.Args...)
			passed, reason := assertKind(rc, candidate.StageName, res, "ConnectorCommandDirectRead", 0)
			if !passed {
				return false, cliInfoFrom(res), reason
			}
			passed, reason = assertDirectReadOutputAssertions(candidate.StageName, res, candidate.OutputAssertions)
			if !passed {
				return false, cliInfoFrom(res), reason
			}
			if hits := ScanForSecrets(res.Stdout, secretValuesFromEnv(rc.opts.SecretEnv)); len(hits) != 0 {
				return false, cliInfoFrom(res), fmt.Sprintf("%s: secret value leaked in output: %v", candidate.StageName, hits)
			}
			return true, cliInfoFrom(res), ""
		})
		for index := range result.Commands {
			if result.Commands[index].Command == command {
				if stage.Passed {
					result.Commands[index].Result = "pass"
					result.Commands[index].Reason = "live read-only provider execution satisfied declaration-owned produced-value assertions"
				} else {
					result.Commands[index].Result = "fail"
					result.Commands[index].Reason = stage.Error
					allLivePassed = false
				}
				break
			}
		}
	}
	if !allLivePassed {
		result.Result = "fail"
		result.Reason = "one or more declaration-owned GraphQL live produced-value assertions failed"
		return nil
	}
	result.Reason = fmt.Sprintf("%d commands are schema-conformant only and %d mutations remain fixture-bound; neither result is a live provider pass", result.SchemaConformant, result.FixtureBound)
	if rc.opts.DirectReadOnly {
		// A direct-read certificate is intentionally scoped to the freshly
		// executed provider reads above. Fixture-bound mutations remain plainly
		// classified in the inventory, but cannot make this bounded read proof
		// fail by pretending it claimed whole-connector completion.
		skipStage(rc, rep, "graphql_inventory_boundary", "skipped: direct-read-only certification does not claim fixture-bound GraphQL mutations")
		return nil
	}
	recordStage(rc, rep, "graphql_inventory_boundary", 0, func() (bool, CLIStageInfo, string) {
		return false, CLIStageInfo{ArgvRedacted: "definition-owned GraphQL inventory", ExitCode: 0, Kind: "GraphQLCertificationInventory"}, "not_live: " + result.Reason
	})
	return nil
}

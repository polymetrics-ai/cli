package engine

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"mime"
	"reflect"
	"regexp/syntax"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"polymetrics.ai/internal/safety"
)

const (
	maxStructuredRESTBodyDepth  = 16
	maxStructuredRESTBodyFields = 256
	maxStructuredRESTBodyItems  = 1024
	maxStructuredRESTBodyNodes  = maxStructuredRESTBodyFields * maxStructuredRESTBodyItems
)

// PreflightOperationStructuredJSONBodyField proves that a command's one
// structured input names a closed, bounded field of its fixed operation. It
// is deliberately a declaration-path preflight: callers never provide
// a body template, route, method, media type, or arbitrary nested path.
func PreflightOperationStructuredJSONBodyField(b Bundle, operation, field string) error {
	op, err := findOperation(b, operation)
	if err != nil {
		return err
	}
	return ValidateOperationStructuredJSONBodyField(op, field)
}

// ValidateOperationStructuredJSONBodyField is the common declaration gate for
// a json command flag on a fixed operation. GraphQL retains its existing
// closed-variable contract. REST writes and POST reads use their respective
// closed body-schema compilation path, which lets direct reads accept a named
// provider-declared object or array without creating a generic request body.
// Commandrunner and connectorgen call this same function; typed execution
// compiles the same schema at its shared body boundary.
func ValidateOperationStructuredJSONBodyField(op OperationSpec, field string) error {
	field = strings.TrimSpace(field)
	if field == "" {
		return fmt.Errorf("operation %q structured body field is required", op.ID)
	}

	switch op.Kind {
	case "graphql_query", "graphql_mutation":
		if err := safety.ValidateIdentifier(field, "structured body field"); err != nil {
			return err
		}
		return ValidateGraphQLOperationStructuredJSONVariable(op, field)
	case "rest_write":
		return validateRESTStructuredJSONBodyField(op, field)
	case "rest_read":
		return validateRESTDirectReadStructuredJSONBodyField(op, field)
	default:
		return fmt.Errorf("operation %q structured JSON body requires rest_read, rest_write, graphql_query, or graphql_mutation, got %q", op.ID, op.Kind)
	}
}

func validateRESTDirectReadStructuredJSONBodyField(op OperationSpec, field string) error {
	if op.REST == nil {
		return fmt.Errorf("operation %q rest_read has no rest declaration", op.ID)
	}
	if strings.ToUpper(strings.TrimSpace(op.REST.Method)) != "POST" {
		return fmt.Errorf("operation %q structured JSON body requires a POST rest_read", op.ID)
	}
	if operationDirectReadContentType(op) != "application/json" {
		return fmt.Errorf("operation %q structured JSON body requires application/json content_type", op.ID)
	}
	_, compiled, err := compileOperationDirectReadBodySchema(op)
	if err != nil {
		return err
	}
	resolved, err := resolveOperationDirectWriteBodySchemaPath(compiled.root, field)
	if err != nil {
		return fmt.Errorf("operation %q body_schema does not declare structured field %q: %w", op.ID, field, err)
	}
	if !isObjectType(resolved.node) && !isArrayType(resolved.node) {
		return fmt.Errorf("operation %q body_schema field %q must be an object or array", op.ID, field)
	}
	return nil
}

func validateRESTStructuredJSONBodyField(op OperationSpec, field string) error {
	if op.REST == nil {
		return fmt.Errorf("operation %q rest_write has no rest declaration", op.ID)
	}
	if op.REST.Multipart != nil {
		return fmt.Errorf("operation %q structured JSON body is unavailable for multipart writes", op.ID)
	}
	if _, _, err := operationStructuredJSONContentType(op); err != nil {
		return err
	}
	root, err := structuredRESTBodySchema(op)
	if err != nil {
		return err
	}
	resolved, err := resolveOperationDirectWriteBodySchemaPath(root, field)
	if err != nil {
		return fmt.Errorf("operation %q body_schema does not declare structured field %q: %w", op.ID, field, err)
	}
	if !isObjectType(resolved.node) && !isArrayType(resolved.node) {
		return fmt.Errorf("operation %q body_schema field %q must be an object or array", op.ID, field)
	}
	return nil
}

func ValidateOperationDirectWriteCLIFlags(op OperationSpec, flags []CLIFlag) error {
	if op.Kind != "rest_write" || op.REST == nil || !operationDirectWriteUsesJSONBody(op) || !operationHasStructuredRESTBodyField(op) {
		return nil
	}
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return err
	}
	staticBody, err := canonicalizeStructuredRESTBodyFragment(compiled, op, op.REST.Body, "rest.body")
	if err != nil {
		return err
	}
	mappings := make([]structuredRESTBodyCLIFieldMapping, 0, len(flags))
	for _, flag := range flags {
		path, ok := strings.CutPrefix(strings.TrimSpace(flag.MapsTo), "body.")
		if !ok || path == "" {
			continue
		}
		resolved, err := resolveOperationDirectWriteBodySchemaPath(compiled.root, path)
		if err != nil {
			return fmt.Errorf("operation %q CLI flag --%s body field %q: %w", op.ID, flag.Name, path, err)
		}
		if err := validateOperationDirectWriteCLIFlagType(op, flag, path, resolved.node); err != nil {
			return err
		}
		if err := validateOperationDirectWriteStaticBodyMapping(staticBody, resolved); err != nil {
			return fmt.Errorf("operation %q CLI flag --%s body field %q: %w", op.ID, flag.Name, path, err)
		}
		mappings = append(mappings, structuredRESTBodyCLIFieldMapping{flag: flag, path: resolved})
	}
	if err := validateStructuredRESTBodyCLIArrayMappings(op.ID, staticBody, mappings); err != nil {
		return err
	}
	coverage := structuredRESTBodyCLICoverageFromValue(staticBody)
	for _, mapping := range mappings {
		if !mapping.flag.Required {
			continue
		}
		structuredRESTBodyAddCLICoverage(coverage, mapping.path.steps)
	}
	return validateStructuredRESTBodyCLILowerBounds(op.ID, compiled.root, coverage, "")
}

type structuredRESTBodyCLIFieldMapping struct {
	flag CLIFlag
	path operationDirectWriteBodyPath
}

type structuredRESTBodyCLIArrayProjection struct {
	prefix  []operationDirectWriteBodyPathStep
	indices map[int]structuredRESTBodyCLIArrayIndex
}

type structuredRESTBodyCLIArrayIndex struct {
	field    string
	required bool
}

// validateStructuredRESTBodyCLIArrayMappings rejects a declared flag that
// would make the generated command synthesize a sparse array. Static rest.body
// items form the only permitted prefix; every later projected item needs a
// required, contiguous predecessor.
func validateStructuredRESTBodyCLIArrayMappings(operation string, staticBody map[string]any, mappings []structuredRESTBodyCLIFieldMapping) error {
	projections := make([]structuredRESTBodyCLIArrayProjection, 0)
	for _, mapping := range mappings {
		for position, step := range mapping.path.steps {
			if !step.array {
				continue
			}
			prefix := mapping.path.steps[:position]
			projectionIndex := -1
			for index := range projections {
				if structuredRESTBodyCLIPathStepsEqual(projections[index].prefix, prefix) {
					projectionIndex = index
					break
				}
			}
			if projectionIndex == -1 {
				projections = append(projections, structuredRESTBodyCLIArrayProjection{
					prefix:  append([]operationDirectWriteBodyPathStep(nil), prefix...),
					indices: make(map[int]structuredRESTBodyCLIArrayIndex),
				})
				projectionIndex = len(projections) - 1
			}
			previous, exists := projections[projectionIndex].indices[step.index]
			if !exists || mapping.path.raw < previous.field {
				previous.field = mapping.path.raw
			}
			previous.required = previous.required || mapping.flag.Required
			projections[projectionIndex].indices[step.index] = previous
		}
	}
	sort.Slice(projections, func(left, right int) bool {
		return operationDirectWriteBodyPathLess(
			operationDirectWriteBodyPath{steps: projections[left].prefix},
			operationDirectWriteBodyPath{steps: projections[right].prefix},
		)
	})
	for _, projection := range projections {
		indices := make([]int, 0, len(projection.indices))
		for index := range projection.indices {
			indices = append(indices, index)
		}
		sort.Ints(indices)
		expected := structuredRESTBodyCLIStaticArrayPrefixLength(staticBody, projection.prefix)
		for _, index := range indices {
			if !projection.indices[index].required || index < expected {
				continue
			}
			if index != expected {
				return fmt.Errorf("operation %q CLI body field %q uses sparse array index %d; rest.body or required CLI mappings must provide every preceding array item", operation, projection.indices[index].field, index)
			}
			expected++
		}
	}
	return nil
}

func structuredRESTBodyCLIPathStepsEqual(left, right []operationDirectWriteBodyPathStep) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func structuredRESTBodyCLIStaticArrayPrefixLength(staticBody map[string]any, prefix []operationDirectWriteBodyPathStep) int {
	value, found := operationDirectWriteBodyPathValue(staticBody, prefix)
	if !found {
		return 0
	}
	items, ok := value.([]any)
	if !ok {
		return 0
	}
	if _, ok := operationDirectWriteStaticBodyScaffold(items); !ok {
		return 0
	}
	return len(items)
}

type structuredRESTBodyCLICoverage struct {
	opaque bool
	object map[string]*structuredRESTBodyCLICoverage
	array  map[int]*structuredRESTBodyCLICoverage
}

func structuredRESTBodyCLICoverageFromValue(value any) *structuredRESTBodyCLICoverage {
	coverage := &structuredRESTBodyCLICoverage{}
	switch value := value.(type) {
	case map[string]any:
		coverage.object = make(map[string]*structuredRESTBodyCLICoverage, len(value))
		for _, name := range sortedMapKeys(value) {
			coverage.object[name] = structuredRESTBodyCLICoverageFromValue(value[name])
		}
	case []any:
		coverage.array = make(map[int]*structuredRESTBodyCLICoverage, len(value))
		for index, item := range value {
			coverage.array[index] = structuredRESTBodyCLICoverageFromValue(item)
		}
	default:
		coverage.opaque = true
	}
	return coverage
}

func structuredRESTBodyAddCLICoverage(coverage *structuredRESTBodyCLICoverage, steps []operationDirectWriteBodyPathStep) {
	if coverage == nil || coverage.opaque {
		return
	}
	if len(steps) == 0 {
		coverage.opaque = true
		coverage.object = nil
		coverage.array = nil
		return
	}
	step := steps[0]
	if step.array {
		if coverage.array == nil {
			coverage.array = make(map[int]*structuredRESTBodyCLICoverage)
		}
		child := coverage.array[step.index]
		if child == nil {
			child = &structuredRESTBodyCLICoverage{}
			coverage.array[step.index] = child
		}
		structuredRESTBodyAddCLICoverage(child, steps[1:])
		return
	}
	if coverage.object == nil {
		coverage.object = make(map[string]*structuredRESTBodyCLICoverage)
	}
	child := coverage.object[step.key]
	if child == nil {
		child = &structuredRESTBodyCLICoverage{}
		coverage.object[step.key] = child
	}
	structuredRESTBodyAddCLICoverage(child, steps[1:])
}

func validateStructuredRESTBodyCLILowerBounds(operation string, node map[string]any, coverage *structuredRESTBodyCLICoverage, path string) error {
	if coverage != nil && coverage.opaque {
		return nil
	}
	object, array := operationDirectWriteBodyNodeKinds(node)
	if object {
		properties, _ := node["properties"].(map[string]any)
		minimum, err := structuredRESTBodyObjectMinProperties(node)
		if err != nil {
			return fmt.Errorf("operation %q body_schema %s: %w", operation, structuredRESTBodyCoverageLabel(path), err)
		}
		present := 0
		if coverage != nil {
			present = len(coverage.object)
		}
		if present < minimum {
			return fmt.Errorf("operation %q requires %s to satisfy minProperties %d but only %d declared static or required CLI fields provide it", operation, structuredRESTBodyCoverageLabel(path), minimum, present)
		}
		required, _ := node["required"].([]any)
		requiredNames := make([]string, 0, len(required))
		for _, rawName := range required {
			if name, ok := rawName.(string); ok {
				requiredNames = append(requiredNames, name)
			}
		}
		sort.Strings(requiredNames)
		for _, name := range requiredNames {
			childCoverage := (*structuredRESTBodyCLICoverage)(nil)
			if coverage != nil {
				childCoverage = coverage.object[name]
			}
			childPath := structuredRESTBodyCoveragePath(path, operationDirectWriteBodyPathStep{key: name})
			child, _ := properties[name].(map[string]any)
			if childCoverage == nil {
				// A required array's actionable lower bound is its first
				// required element, not merely the container name. Preserve that
				// fact so a source declaration with minItems cannot be mistaken
				// for a generic missing object field.
				if _, childArray := operationDirectWriteBodyNodeKinds(child); childArray {
					return validateStructuredRESTBodyCLILowerBounds(operation, child, nil, childPath)
				}
				return fmt.Errorf("operation %q requires %s but no required command flag maps to it and rest.body does not provide it", operation, structuredRESTBodyCoverageLabel(childPath))
			}
			if err := validateStructuredRESTBodyCLILowerBounds(operation, child, childCoverage, childPath); err != nil {
				return err
			}
		}
		if coverage == nil {
			return nil
		}
		for _, name := range sortedMapKeys(coverage.object) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				continue
			}
			if err := validateStructuredRESTBodyCLILowerBounds(operation, child, coverage.object[name], structuredRESTBodyCoveragePath(path, operationDirectWriteBodyPathStep{key: name})); err != nil {
				return err
			}
		}
	}
	if array {
		minimum, err := structuredRESTBodyArrayMinItems(node)
		if err != nil {
			return fmt.Errorf("operation %q body_schema %s: %w", operation, structuredRESTBodyCoverageLabel(path), err)
		}
		for index := 0; index < minimum; index++ {
			var childCoverage *structuredRESTBodyCLICoverage
			if coverage != nil {
				childCoverage = coverage.array[index]
			}
			childPath := structuredRESTBodyCoveragePath(path, operationDirectWriteBodyPathStep{array: true, index: index})
			if childCoverage == nil {
				return fmt.Errorf("operation %q requires %s to satisfy minItems %d but no required command flag maps to it and rest.body does not provide it", operation, structuredRESTBodyCoverageLabel(childPath), minimum)
			}
			item, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				return fmt.Errorf("operation %q body_schema %s: %w", operation, structuredRESTBodyCoverageLabel(childPath), err)
			}
			if err := validateStructuredRESTBodyCLILowerBounds(operation, item, childCoverage, childPath); err != nil {
				return err
			}
		}
		if coverage == nil {
			return nil
		}
		indices := make([]int, 0, len(coverage.array))
		for index := range coverage.array {
			indices = append(indices, index)
		}
		sort.Ints(indices)
		for _, index := range indices {
			item, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				return fmt.Errorf("operation %q body_schema %s: %w", operation, structuredRESTBodyCoverageLabel(path), err)
			}
			if err := validateStructuredRESTBodyCLILowerBounds(operation, item, coverage.array[index], structuredRESTBodyCoveragePath(path, operationDirectWriteBodyPathStep{array: true, index: index})); err != nil {
				return err
			}
		}
	}
	return nil
}

func structuredRESTBodyCoveragePath(prefix string, step operationDirectWriteBodyPathStep) string {
	if step.array {
		if prefix == "" {
			return strconv.Itoa(step.index)
		}
		return prefix + "." + strconv.Itoa(step.index)
	}
	if prefix == "" {
		return step.key
	}
	return prefix + "." + step.key
}

func structuredRESTBodyCoverageLabel(path string) string {
	if path == "" {
		return "body"
	}
	return "body." + path
}

func validateOperationDirectWriteCLIFlagType(op OperationSpec, flag CLIFlag, path string, node map[string]any) error {
	valid := false
	switch flag.Type {
	case "json":
		valid = (isObjectType(node) || isArrayType(node)) && !structuredRESTBodyNodeHasEnum(node)
	case "", "string":
		valid = structuredRESTBodyNodeAllowsType(node, "string") && structuredRESTBodyNodeAcceptsAllCLIStrings(node)
	case "enum":
		valid = structuredRESTBodyNodeAllowsType(node, "string") && structuredRESTBodyCLIValuesFitNode(flag.Values, node)
	case "integer":
		valid = structuredRESTBodyNodeAllowsType(node, "integer") || structuredRESTBodyNodeAllowsType(node, "number")
		if valid && structuredRESTBodyNodeHasEnum(node) {
			valid = false
		}
	case "number":
		valid = structuredRESTBodyNodeAllowsType(node, "number")
		if valid && structuredRESTBodyNodeHasEnum(node) {
			valid = false
		}
	case "boolean":
		valid = structuredRESTBodyNodeAllowsType(node, "boolean")
		if valid && structuredRESTBodyNodeHasEnum(node) {
			valid = false
		}
	case "string_array":
		valid = structuredRESTBodyArrayAllowsStrings(node)
		if valid {
			maxItems, err := structuredRESTBodyArrayMaxItems(node)
			if err != nil {
				return fmt.Errorf("operation %q CLI flag --%s body field %q: %w", op.ID, flag.Name, path, err)
			}
			minItems, err := structuredRESTBodyArrayMinItems(node)
			if err != nil {
				return fmt.Errorf("operation %q CLI flag --%s body field %q: %w", op.ID, flag.Name, path, err)
			}
			if flag.MaxItems <= 0 || flag.MinItems < minItems || flag.MaxItems > maxItems || flag.MinItems > flag.MaxItems || !structuredRESTBodyArrayCLIValueDomainFits(node, flag.Values) {
				valid = false
			}
		}
	}
	if valid {
		return nil
	}
	return fmt.Errorf("operation %q CLI flag --%s type %q does not match declared body field %q", op.ID, flag.Name, flag.Type, path)
}

func structuredRESTBodyNodeAllowsType(node map[string]any, want string) bool {
	raw, exists := node["type"]
	if !exists {
		return !isObjectType(node) && !isArrayType(node)
	}
	switch types := raw.(type) {
	case string:
		return types == want
	case []any:
		for _, rawType := range types {
			if typeName, ok := rawType.(string); ok && typeName == want {
				return true
			}
		}
	}
	return false
}

func structuredRESTBodyArrayAllowsStrings(node map[string]any) bool {
	if !isArrayType(node) {
		return false
	}
	maxItems, err := structuredRESTBodyArrayMaxItems(node)
	if err != nil || maxItems == 0 {
		return false
	}
	for index := 0; index < maxItems; index++ {
		item, err := operationDirectWriteBodyArrayItemSchema(node, index)
		if err != nil || !structuredRESTBodyNodeAllowsType(item, "string") {
			return false
		}
	}
	return true
}

func structuredRESTBodyNodeHasEnum(node map[string]any) bool {
	values, ok := node["enum"].([]any)
	return ok && len(values) != 0
}

func structuredRESTBodyNodeAcceptsAllCLIStrings(node map[string]any) bool {
	if structuredRESTBodyNodeHasEnum(node) {
		return false
	}
	if _, constrained := node["pattern"]; constrained {
		return false
	}
	format, _ := node["format"].(string)
	return format != "uri"
}

func structuredRESTBodyCLIValuesFitNode(values []string, node map[string]any) bool {
	if len(values) == 0 {
		return false
	}
	raw, err := json.Marshal(node)
	if err != nil {
		return false
	}
	schema, err := compileStructuredRESTBodySchemaDocument(raw)
	if err != nil {
		return false
	}
	for _, value := range values {
		if err := schema.Validate(value); err != nil {
			return false
		}
	}
	return true
}

func structuredRESTBodyArrayCLIValueDomainFits(node map[string]any, values []string) bool {
	maxItems, err := structuredRESTBodyArrayMaxItems(node)
	if err != nil {
		return false
	}
	for index := 0; index < maxItems; index++ {
		item, err := operationDirectWriteBodyArrayItemSchema(node, index)
		if err != nil {
			return false
		}
		if len(values) == 0 {
			if !structuredRESTBodyNodeAcceptsAllCLIStrings(item) {
				return false
			}
			continue
		}
		if !structuredRESTBodyCLIValuesFitNode(values, item) {
			return false
		}
	}
	return true
}

func operationStructuredJSONContentType(op OperationSpec) (string, string, error) {
	if op.REST == nil {
		return "", "", fmt.Errorf("operation %q has no rest declaration", op.ID)
	}
	declared := strings.TrimSpace(op.REST.ContentType)
	if declared == "" {
		return "application/json", "json", nil
	}
	mediaType, _, err := mime.ParseMediaType(declared)
	if err != nil {
		return "", "", fmt.Errorf("operation %q has invalid rest content_type %q: %w", op.ID, declared, err)
	}
	switch strings.ToLower(mediaType) {
	case "application/json", "application/scim+json":
		return strings.ToLower(mediaType), "json", nil
	default:
		return "", "", fmt.Errorf("operation %q structured JSON body requires application/json or application/scim+json content_type", op.ID)
	}
}

func structuredRESTBodySchema(op OperationSpec) (map[string]any, error) {
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return nil, err
	}
	return compiled.root, nil
}

func operationHasStructuredRESTBodyField(op OperationSpec) bool {
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return false
	}
	var root map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &root); err != nil {
		return false
	}
	properties, ok := root["properties"].(map[string]any)
	if !ok {
		return false
	}
	if closed, ok := root["additionalProperties"].(bool); ok && !closed {
		return true
	}
	for _, raw := range properties {
		node, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if isObjectType(node) || isArrayType(node) {
			return true
		}
		if _, ok := node["properties"]; ok {
			return true
		}
		if _, ok := node["additionalProperties"]; ok {
			return true
		}
		if _, ok := node["items"]; ok {
			return true
		}
		if _, ok := node["prefixItems"]; ok {
			return true
		}
	}
	return false
}

func OperationDirectWriteHasStructuredRESTBody(op OperationSpec) bool {
	return operationDirectWriteUsesJSONBody(op) && operationHasStructuredRESTBodyField(op)
}

type structuredRESTBodySchemaCompilation struct {
	root           map[string]any
	schema         *Schema
	fragmentSchema *Schema
}

func operationDirectWriteStaticBodyScaffold(value any) (any, bool) {
	switch value := value.(type) {
	case map[string]any:
		shape := make(map[string]any, len(value))
		for _, name := range sortedMapKeys(value) {
			child, ok := operationDirectWriteStaticBodyScaffold(value[name])
			if ok {
				shape[name] = child
			}
		}
		return shape, true
	case []any:
		shape := make([]any, len(value))
		for index := range value {
			child, ok := operationDirectWriteStaticBodyScaffold(value[index])
			if !ok {
				return nil, false
			}
			shape[index] = child
		}
		return shape, true
	default:
		return nil, false
	}
}

func compileStructuredRESTBodySchema(op OperationSpec) (*structuredRESTBodySchemaCompilation, error) {
	if op.REST == nil || len(op.REST.BodySchema) == 0 {
		return nil, fmt.Errorf("operation %q structured JSON body requires body_schema", op.ID)
	}
	var root map[string]any
	if err := json.Unmarshal(op.REST.BodySchema, &root); err != nil {
		return nil, fmt.Errorf("operation %q body_schema is not an object: %w", op.ID, err)
	}
	if root == nil {
		return nil, fmt.Errorf("operation %q body_schema must be an object", op.ID)
	}
	if err := normalizeStructuredRESTBodySchemaNode(op.ID, root, "body_schema"); err != nil {
		return nil, err
	}
	if !isObjectType(root) {
		return nil, fmt.Errorf("operation %q body_schema must be an object", op.ID)
	}
	normalized, err := json.Marshal(root)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: normalize: %w", op.ID, err)
	}
	sch, err := compileStructuredRESTBodySchemaDocument(normalized)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: %w", op.ID, err)
	}
	state := structuredRESTBodySchemaState{}
	if err := requireClosedBoundedStructuredRESTBody(op.ID, root, "body_schema", 1, &state); err != nil {
		return nil, err
	}
	minimumNodes, _, err := structuredRESTBodyNodeCosts(root)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: %w", op.ID, err)
	}
	if minimumNodes > maxStructuredRESTBodyNodes {
		return nil, structuredRESTBodyFoundationGap(op.ID, "body_schema", fmt.Sprintf("minimum valid body requires %d nodes, exceeding limit %d", minimumNodes, maxStructuredRESTBodyNodes))
	}
	// A schema may describe several independently bounded arrays whose
	// theoretical maxima cannot coexist inside the operation's byte limit. The
	// runtime enforces both the node ceiling and max_bytes on the materialized
	// value, so rejecting that safe union here would make declared operations
	// unreachable without strengthening the actual boundary.
	minimumBytes, err := structuredRESTBodyMinimumJSONBytes(root)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: %w", op.ID, err)
	}
	maxBytes := clampOperationDirectWriteMaxBytes(op.REST.MaxBytes)
	if minimumBytes > maxBytes {
		return nil, structuredRESTBodyFoundationGap(op.ID, "body_schema", fmt.Sprintf("minimum valid body requires %d bytes, exceeding limit %d", minimumBytes, maxBytes))
	}
	fragmentRoot, err := structuredRESTBodyFragmentSchemaRoot(root)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: %w", op.ID, err)
	}
	fragmentRaw, err := json.Marshal(fragmentRoot)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: encode fragment schema: %w", op.ID, err)
	}
	fragmentSchema, err := compileStructuredRESTBodySchemaDocument(fragmentRaw)
	if err != nil {
		return nil, fmt.Errorf("operation %q body_schema: compile fragment schema: %w", op.ID, err)
	}
	compiled := &structuredRESTBodySchemaCompilation{root: root, schema: sch, fragmentSchema: fragmentSchema}
	staticBody := map[string]any{}
	if len(op.REST.Body) != 0 {
		staticBody, err = canonicalizeStructuredRESTBodyFragment(compiled, op, op.REST.Body, "rest.body")
		if err != nil {
			if strings.Contains(err.Error(), "request body too large") {
				return nil, structuredRESTBodyFoundationGap(op.ID, "rest.body", fmt.Sprintf("minimum fixed body exceeds limit %d", maxBytes))
			}
			return nil, err
		}
		encoded, err := json.Marshal(staticBody)
		if err != nil {
			return nil, fmt.Errorf("operation %q rest.body: encode: %w", op.ID, err)
		}
		if len(encoded) > maxBytes {
			return nil, structuredRESTBodyFoundationGap(op.ID, "rest.body", fmt.Sprintf("minimum fixed body requires %d bytes, exceeding limit %d", len(encoded), maxBytes))
		}
	}
	minimumCompletionBytes, err := structuredRESTBodyMinimumJSONBytesWithStatic(root, staticBody)
	if err != nil {
		return nil, fmt.Errorf("operation %q rest.body: %w", op.ID, err)
	}
	if minimumCompletionBytes > maxBytes {
		return nil, structuredRESTBodyFoundationGap(op.ID, "rest.body", fmt.Sprintf("minimum valid body completion requires %d bytes, exceeding limit %d", minimumCompletionBytes, maxBytes))
	}
	return compiled, nil
}

func materializeStructuredRESTBody(op OperationSpec, staticBody, overrides map[string]any) (map[string]any, error) {
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return nil, err
	}
	if len(staticBody) > maxStructuredRESTBodyFields || len(overrides) > maxStructuredRESTBodyFields {
		return nil, fmt.Errorf("operation %q: body exceeds structured body field limit %d", op.ID, maxStructuredRESTBodyFields)
	}
	staticCanonical, err := canonicalizeStructuredRESTBodyFragment(compiled, op, staticBody, "rest.body")
	if err != nil {
		return nil, err
	}
	overrideCanonical, err := canonicalizeStructuredRESTBodyFragment(compiled, op, overrides, "body")
	if err != nil {
		return nil, err
	}
	body, err := mergeStructuredRESTBodyObject(compiled.root, staticCanonical, overrideCanonical, "body")
	if err != nil {
		return nil, fmt.Errorf("operation %q: %w", op.ID, err)
	}
	state := structuredRESTBodyValueState{maxBytes: clampOperationDirectWriteMaxBytes(op.REST.MaxBytes)}
	canonicalValue, err := canonicalizeStructuredRESTBodyValue(compiled.root, reflect.ValueOf(body), "body", 0, &state)
	if err != nil {
		return nil, fmt.Errorf("operation %q: %w", op.ID, err)
	}
	canonical, ok := canonicalValue.(map[string]any)
	if canonical == nil {
		return nil, fmt.Errorf("operation %q: structured body must be an object", op.ID)
	}
	if !ok {
		return nil, fmt.Errorf("operation %q: structured body must be an object", op.ID)
	}
	if err := compiled.schema.Validate(canonical); err != nil {
		return nil, fmt.Errorf("operation %q: body_schema: %w", op.ID, err)
	}
	return canonical, nil
}

func canonicalizeStructuredRESTBodyFragment(compiled *structuredRESTBodySchemaCompilation, op OperationSpec, body map[string]any, path string) (map[string]any, error) {
	if body == nil {
		return map[string]any{}, nil
	}
	state := structuredRESTBodyValueState{maxBytes: clampOperationDirectWriteMaxBytes(op.REST.MaxBytes)}
	canonicalValue, err := canonicalizeStructuredRESTBodyValue(compiled.root, reflect.ValueOf(body), path, 0, &state)
	if err != nil {
		return nil, fmt.Errorf("operation %q: %w", op.ID, err)
	}
	canonical, ok := canonicalValue.(map[string]any)
	if !ok || canonical == nil {
		return nil, fmt.Errorf("operation %q: %s must be an object", op.ID, path)
	}
	if err := compiled.fragmentSchema.Validate(canonical); err != nil {
		return nil, fmt.Errorf("operation %q: body_schema fragment: %w", op.ID, err)
	}
	return canonical, nil
}

func mergeStructuredRESTBodyObject(node map[string]any, staticBody, overrideBody map[string]any, path string) (map[string]any, error) {
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		return nil, fmt.Errorf("%s has no declared object properties", path)
	}
	keys := make(map[string]struct{}, len(staticBody)+len(overrideBody))
	for key := range staticBody {
		keys[key] = struct{}{}
	}
	for key := range overrideBody {
		keys[key] = struct{}{}
	}
	names := make([]string, 0, len(keys))
	for name := range keys {
		names = append(names, name)
	}
	sort.Strings(names)
	merged := make(map[string]any, len(names))
	for _, name := range names {
		staticValue, hasStatic := staticBody[name]
		overrideValue, hasOverride := overrideBody[name]
		if !hasStatic {
			merged[name] = overrideValue
			continue
		}
		if !hasOverride {
			merged[name] = staticValue
			continue
		}
		child, ok := properties[name].(map[string]any)
		if !ok {
			return nil, fmt.Errorf("%s.%s has an invalid declared schema", path, name)
		}
		value, err := mergeStructuredRESTBodyValue(child, staticValue, overrideValue, path+"."+name)
		if err == nil {
			merged[name] = value
			continue
		}
		return nil, fmt.Errorf("%s.%s is fixed by rest.body and cannot be caller-overridden", path, name)
	}
	return merged, nil
}

func mergeStructuredRESTBodyValue(node map[string]any, staticValue, overrideValue any, path string) (any, error) {
	if isObjectType(node) {
		staticObject, staticObjectOK := staticValue.(map[string]any)
		overrideObject, overrideObjectOK := overrideValue.(map[string]any)
		if staticObjectOK && overrideObjectOK {
			return mergeStructuredRESTBodyObject(node, staticObject, overrideObject, path)
		}
	}
	if isArrayType(node) {
		staticArray, staticArrayOK := staticValue.([]any)
		overrideArray, overrideArrayOK := overrideValue.([]any)
		if staticArrayOK && overrideArrayOK {
			return mergeStructuredRESTBodyArray(node, staticArray, overrideArray, path)
		}
	}
	return nil, fmt.Errorf("%s is fixed by rest.body and cannot be caller-overridden", path)
}

func mergeStructuredRESTBodyArray(node map[string]any, staticBody, overrideBody []any, path string) ([]any, error) {
	length := len(staticBody)
	if len(overrideBody) > length {
		length = len(overrideBody)
	}
	merged := make([]any, length)
	for index := 0; index < length; index++ {
		hasStatic := index < len(staticBody)
		hasOverride := index < len(overrideBody)
		switch {
		case !hasStatic:
			merged[index] = overrideBody[index]
		case !hasOverride:
			merged[index] = staticBody[index]
		default:
			item, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				return nil, err
			}
			value, err := mergeStructuredRESTBodyValue(item, staticBody[index], overrideBody[index], fmt.Sprintf("%s.%d", path, index))
			if err != nil {
				return nil, err
			}
			merged[index] = value
		}
	}
	return merged, nil
}

func structuredRESTBodyFragmentSchemaRoot(root map[string]any) (map[string]any, error) {
	raw, err := json.Marshal(root)
	if err != nil {
		return nil, err
	}
	var fragment map[string]any
	if err := json.Unmarshal(raw, &fragment); err != nil {
		return nil, err
	}
	if err := stripStructuredRESTBodyFragmentRequirements(fragment); err != nil {
		return nil, err
	}
	return fragment, nil
}

func stripStructuredRESTBodyFragmentRequirements(node map[string]any) error {
	delete(node, "required")
	delete(node, "minProperties")
	delete(node, "minItems")
	if properties, ok := node["properties"].(map[string]any); ok {
		for _, name := range sortedMapKeys(properties) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				return fmt.Errorf("properties.%s must be a schema object", name)
			}
			if err := stripStructuredRESTBodyFragmentRequirements(child); err != nil {
				return err
			}
		}
	}
	if items, ok := node["items"].(map[string]any); ok {
		if err := stripStructuredRESTBodyFragmentRequirements(items); err != nil {
			return err
		}
	}
	if prefixItems, ok := node["prefixItems"].([]any); ok {
		for index, rawItem := range prefixItems {
			item, ok := rawItem.(map[string]any)
			if !ok {
				return fmt.Errorf("prefixItems.%d must be a schema object", index)
			}
			if err := stripStructuredRESTBodyFragmentRequirements(item); err != nil {
				return err
			}
		}
	}
	return nil
}

type structuredRESTBodyValueState struct {
	maxBytes int
	bytes    int
	nodes    int
	visiting map[structuredRESTBodyContainer]struct{}
}

type structuredRESTBodyContainer struct {
	kind    reflect.Kind
	pointer uintptr
}

func (s *structuredRESTBodyValueState) addNode(path string) error {
	if s.nodes >= maxStructuredRESTBodyNodes {
		return fmt.Errorf("%s exceeds structured body node limit %d", path, maxStructuredRESTBodyNodes)
	}
	s.nodes++
	return nil
}

func (s *structuredRESTBodyValueState) addBytes(path string, count int) error {
	if count < 0 || count > s.maxBytes-s.bytes {
		return fmt.Errorf("request body too large: %s exceeds limit %d", path, s.maxBytes)
	}
	s.bytes += count
	return nil
}

func (s *structuredRESTBodyValueState) enterContainer(value reflect.Value, path string) (func(), error) {
	if value.Kind() != reflect.Map && value.Kind() != reflect.Slice {
		return func() {}, nil
	}
	pointer := value.Pointer()
	if pointer == 0 {
		return func() {}, nil
	}
	container := structuredRESTBodyContainer{kind: value.Kind(), pointer: pointer}
	if s.visiting == nil {
		s.visiting = make(map[structuredRESTBodyContainer]struct{})
	}
	if _, exists := s.visiting[container]; exists {
		return nil, fmt.Errorf("%s exceeds structured body depth limit %d", path, maxStructuredRESTBodyDepth)
	}
	s.visiting[container] = struct{}{}
	return func() { delete(s.visiting, container) }, nil
}

func structuredRESTBodyMapReferencesSelf(value reflect.Value) bool {
	pointer := value.Pointer()
	iter := value.MapRange()
	for iter.Next() {
		candidate := iter.Value()
		for candidate.IsValid() && (candidate.Kind() == reflect.Interface || candidate.Kind() == reflect.Pointer) {
			if candidate.IsNil() {
				break
			}
			candidate = candidate.Elem()
		}
		if candidate.IsValid() && candidate.Kind() == reflect.Map && !candidate.IsNil() && candidate.Pointer() == pointer {
			return true
		}
	}
	return false
}

func (s *structuredRESTBodyValueState) addJSONString(path, value string) error {
	if err := s.addBytes(path, 2); err != nil {
		return err
	}
	for len(value) > 0 {
		r, width := utf8.DecodeRuneInString(value)
		if r == utf8.RuneError && width == 1 {
			if err := s.addBytes(path, 6); err != nil {
				return err
			}
			value = value[width:]
			continue
		}
		count := width
		switch r {
		case '"', '\\':
			count = 2
		case '\b', '\f', '\n', '\r', '\t':
			count = 2
		case '<', '>', '&', '\u2028', '\u2029':
			count = 6
		default:
			if r < 0x20 {
				count = 6
			}
		}
		if err := s.addBytes(path, count); err != nil {
			return err
		}
		value = value[width:]
	}
	return nil
}

func canonicalizeStructuredRESTBodyValue(node map[string]any, value reflect.Value, path string, depth int, state *structuredRESTBodyValueState) (any, error) {
	if depth > maxStructuredRESTBodyDepth {
		return nil, fmt.Errorf("%s exceeds structured body depth limit %d", path, maxStructuredRESTBodyDepth)
	}
	if !value.IsValid() {
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if err := state.addBytes(path, len("null")); err != nil {
			return nil, err
		}
		return nil, nil
	}
	if value.CanInterface() {
		if _, ok := value.Interface().(json.Marshaler); ok {
			return nil, fmt.Errorf("%s does not permit custom JSON marshalers", path)
		}
	}
	for value.Kind() == reflect.Interface || value.Kind() == reflect.Pointer {
		if value.IsNil() {
			if err := state.addNode(path); err != nil {
				return nil, err
			}
			if err := state.addBytes(path, len("null")); err != nil {
				return nil, err
			}
			return nil, nil
		}
		depth++
		if depth > maxStructuredRESTBodyDepth {
			return nil, fmt.Errorf("%s exceeds structured body depth limit %d", path, maxStructuredRESTBodyDepth)
		}
		value = value.Elem()
		if value.CanInterface() {
			if _, ok := value.Interface().(json.Marshaler); ok {
				return nil, fmt.Errorf("%s does not permit custom JSON marshalers", path)
			}
		}
	}
	if value.CanInterface() {
		if number, ok := value.Interface().(json.Number); ok {
			if err := state.addNode(path); err != nil {
				return nil, err
			}
			if err := state.addBytes(path, len(number)); err != nil {
				return nil, err
			}
			return number, nil
		}
	}

	switch value.Kind() {
	case reflect.Map:
		if !isObjectType(node) {
			return nil, fmt.Errorf("%s: value does not match type", path)
		}
		if value.IsNil() {
			if err := state.addNode(path); err != nil {
				return nil, err
			}
			if err := state.addBytes(path, len("null")); err != nil {
				return nil, err
			}
			return nil, nil
		}
		if value.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("%s does not permit non-string map keys", path)
		}
		leave, err := state.enterContainer(value, path)
		if err != nil {
			return nil, err
		}
		defer leave()
		properties, _ := node["properties"].(map[string]any)
		if value.Len() > maxStructuredRESTBodyFields {
			return nil, fmt.Errorf("%s exceeds structured body field limit %d", path, maxStructuredRESTBodyFields)
		}
		if structuredRESTBodyMapReferencesSelf(value) {
			return nil, fmt.Errorf("%s exceeds structured body depth limit %d", path, maxStructuredRESTBodyDepth)
		}
		if value.Len() > len(properties) {
			return nil, fmt.Errorf("%s: additional property not allowed", path)
		}
		keys := make([]string, 0, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			keys = append(keys, iter.Key().String())
		}
		sort.Strings(keys)
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if err := state.addBytes(path, 2+max(0, len(keys)-1)); err != nil {
			return nil, err
		}
		out := make(map[string]any, len(keys))
		for _, key := range keys {
			child, ok := properties[key]
			if !ok {
				return nil, fmt.Errorf("%s: additional property not allowed", path)
			}
			childNode, ok := child.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("%s/%s has an invalid declared schema", path, key)
			}
			if err := state.addJSONString(path, key); err != nil {
				return nil, err
			}
			if err := state.addBytes(path, 1); err != nil {
				return nil, err
			}
			mapKey := reflect.ValueOf(key).Convert(value.Type().Key())
			normalized, err := canonicalizeStructuredRESTBodyValue(childNode, value.MapIndex(mapKey), path+"."+key, depth+1, state)
			if err != nil {
				return nil, err
			}
			out[key] = normalized
		}
		return out, nil
	case reflect.Slice, reflect.Array:
		if value.Kind() == reflect.Slice && value.IsNil() {
			if err := state.addNode(path); err != nil {
				return nil, err
			}
			if err := state.addBytes(path, len("null")); err != nil {
				return nil, err
			}
			return nil, nil
		}
		if value.Type().Elem().Kind() == reflect.Uint8 {
			if err := state.addNode(path); err != nil {
				return nil, err
			}
			encodedLen := base64.StdEncoding.EncodedLen(value.Len())
			if err := state.addBytes(path, encodedLen+2); err != nil {
				return nil, err
			}
			bytes := make([]byte, value.Len())
			for index := range bytes {
				bytes[index] = byte(value.Index(index).Uint())
			}
			return base64.StdEncoding.EncodeToString(bytes), nil
		}
		if !isArrayType(node) {
			return nil, fmt.Errorf("%s: value does not match type", path)
		}
		leave, err := state.enterContainer(value, path)
		if err != nil {
			return nil, err
		}
		defer leave()
		maxItems, err := structuredRESTBodyArrayMaxItems(node)
		if err != nil {
			return nil, err
		}
		if value.Len() > maxItems {
			return nil, fmt.Errorf("%s: maxItems %d exceeded (got %d)", path, maxItems, value.Len())
		}
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if err := state.addBytes(path, 2+max(0, value.Len()-1)); err != nil {
			return nil, err
		}
		out := make([]any, value.Len())
		for index := 0; index < value.Len(); index++ {
			childNode, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				return nil, err
			}
			normalized, err := canonicalizeStructuredRESTBodyValue(childNode, value.Index(index), fmt.Sprintf("%s.%d", path, index), depth+1, state)
			if err != nil {
				return nil, err
			}
			out[index] = normalized
		}
		return out, nil
	case reflect.Bool:
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if value.Bool() {
			if err := state.addBytes(path, len("true")); err != nil {
				return nil, err
			}
		} else if err := state.addBytes(path, len("false")); err != nil {
			return nil, err
		}
		return value.Bool(), nil
	case reflect.String:
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if err := state.addJSONString(path, value.String()); err != nil {
			return nil, err
		}
		return value.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		number := json.Number(strconv.FormatInt(value.Int(), 10))
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if err := state.addBytes(path, len(number)); err != nil {
			return nil, err
		}
		return number, nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		number := json.Number(strconv.FormatUint(value.Uint(), 10))
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if err := state.addBytes(path, len(number)); err != nil {
			return nil, err
		}
		return number, nil
	case reflect.Float32, reflect.Float64:
		number, err := json.Marshal(value.Float())
		if err != nil {
			return nil, fmt.Errorf("%s does not permit non-finite numbers", path)
		}
		if err := state.addNode(path); err != nil {
			return nil, err
		}
		if err := state.addBytes(path, len(number)); err != nil {
			return nil, err
		}
		return value.Float(), nil
	case reflect.Struct:
		return nil, fmt.Errorf("%s does not permit struct values", path)
	default:
		return nil, fmt.Errorf("%s does not permit values of type %s", path, value.Type())
	}
}

func structuredRESTBodyArrayMaxItems(node map[string]any) (int, error) {
	raw, ok := node["maxItems"]
	if !ok {
		return 0, fmt.Errorf("array body field has no maxItems")
	}
	maxItems, ok := raw.(float64)
	if !ok || math.Trunc(maxItems) != maxItems || maxItems < 0 || maxItems > maxStructuredRESTBodyItems {
		return 0, fmt.Errorf("array body field has invalid maxItems")
	}
	return int(maxItems), nil
}

func structuredRESTBodyArrayMinItems(node map[string]any) (int, error) {
	raw, ok := node["minItems"]
	if !ok {
		return 0, nil
	}
	minItems, ok := raw.(float64)
	if !ok || math.Trunc(minItems) != minItems || minItems < 0 || minItems > maxStructuredRESTBodyItems {
		return 0, fmt.Errorf("array body field has invalid minItems")
	}
	return int(minItems), nil
}

func structuredRESTBodyObjectMinProperties(node map[string]any) (int, error) {
	raw, ok := node["minProperties"]
	if !ok {
		return 0, nil
	}
	minProperties, ok := raw.(float64)
	if !ok || math.Trunc(minProperties) != minProperties || minProperties < 0 || minProperties > maxStructuredRESTBodyFields {
		return 0, fmt.Errorf("object body field has invalid minProperties")
	}
	return int(minProperties), nil
}

func normalizeStructuredRESTBodyValue(value any, path string) (any, error) {
	if number, ok := value.(json.Number); ok {
		return number, nil
	}
	return normalizeStructuredRESTBodyReflectValue(reflect.ValueOf(value), path, 0)
}

func normalizeStructuredRESTBodyReflectValue(value reflect.Value, path string, depth int) (any, error) {
	if depth > maxStructuredRESTBodyDepth {
		return nil, fmt.Errorf("%s exceeds structured body depth limit %d", path, maxStructuredRESTBodyDepth)
	}
	if !value.IsValid() {
		return nil, nil
	}
	if value.CanInterface() {
		if _, ok := value.Interface().(json.Marshaler); ok {
			return nil, fmt.Errorf("%s does not permit custom JSON marshalers", path)
		}
	}
	switch value.Kind() {
	case reflect.Interface, reflect.Pointer:
		if value.IsNil() {
			return nil, nil
		}
		return normalizeStructuredRESTBodyReflectValue(value.Elem(), path, depth+1)
	case reflect.Map:
		if value.IsNil() {
			return nil, nil
		}
		if value.Type().Key().Kind() != reflect.String {
			return nil, fmt.Errorf("%s does not permit non-string map keys", path)
		}
		out := make(map[string]any, value.Len())
		iter := value.MapRange()
		for iter.Next() {
			key := iter.Key().String()
			normalized, err := normalizeStructuredRESTBodyReflectValue(iter.Value(), path+"."+key, depth+1)
			if err != nil {
				return nil, err
			}
			out[key] = normalized
		}
		return out, nil
	case reflect.Slice:
		if value.IsNil() {
			return nil, nil
		}
		if value.Type().Elem().Kind() == reflect.Uint8 {
			out := make([]byte, value.Len())
			for index := 0; index < value.Len(); index++ {
				out[index] = byte(value.Index(index).Uint())
			}
			return out, nil
		}
		fallthrough
	case reflect.Array:
		out := make([]any, value.Len())
		for index := 0; index < value.Len(); index++ {
			normalized, err := normalizeStructuredRESTBodyReflectValue(value.Index(index), fmt.Sprintf("%s.%d", path, index), depth+1)
			if err != nil {
				return nil, err
			}
			out[index] = normalized
		}
		return out, nil
	case reflect.Bool:
		return value.Bool(), nil
	case reflect.String:
		return value.String(), nil
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return json.Number(strconv.FormatInt(value.Int(), 10)), nil
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return json.Number(strconv.FormatUint(value.Uint(), 10)), nil
	case reflect.Float32, reflect.Float64:
		return value.Float(), nil
	case reflect.Struct:
		return nil, fmt.Errorf("%s does not permit struct values", path)
	default:
		return nil, fmt.Errorf("%s does not permit values of type %s", path, value.Type())
	}
}

type structuredRESTBodySchemaState struct {
	fields int
}

func requireClosedBoundedStructuredRESTBody(operation string, node map[string]any, path string, depth int, state *structuredRESTBodySchemaState) error {
	if depth > maxStructuredRESTBodyDepth {
		return fmt.Errorf("operation %q %s exceeds structured body depth limit %d", operation, path, maxStructuredRESTBodyDepth)
	}
	if structuredRESTBodyNodeMayAcceptObjectOrArray(node) && !structuredRESTBodyNodeHasExplicitType(node) {
		return structuredRESTBodyFoundationGap(operation, path, "may accept objects or arrays; declare an explicit type")
	}
	object, array := operationDirectWriteBodyNodeKinds(node)
	if object && array {
		return structuredRESTBodyFoundationGap(operation, path, "declares both object and array structure")
	}
	if object {
		if closed, ok := node["additionalProperties"].(bool); !ok || closed {
			return fmt.Errorf("operation %q %s is an object and must declare additionalProperties: false", operation, path)
		}
		properties, ok := node["properties"].(map[string]any)
		if !ok {
			return fmt.Errorf("operation %q %s is an object and must declare properties", operation, path)
		}
		if rawRequired, exists := node["required"]; exists {
			required, ok := rawRequired.([]any)
			if !ok {
				return fmt.Errorf("operation %q %s required must be an array", operation, path)
			}
			seen := make(map[string]struct{}, len(required))
			for _, rawName := range required {
				name, ok := rawName.(string)
				if !ok || name == "" {
					return fmt.Errorf("operation %q %s required must contain non-empty property names", operation, path)
				}
				if _, duplicate := seen[name]; duplicate {
					return fmt.Errorf("operation %q %s required duplicates property %q", operation, path, name)
				}
				seen[name] = struct{}{}
				if _, declared := properties[name]; !declared {
					return fmt.Errorf("operation %q %s required property %q is not declared in properties", operation, path, name)
				}
			}
		}
		minProperties, err := structuredRESTBodyObjectMinProperties(node)
		if err != nil {
			return fmt.Errorf("operation %q %s %w", operation, path, err)
		}
		if minProperties > len(properties) {
			return structuredRESTBodyFoundationGap(operation, path, fmt.Sprintf("minProperties %d exceeds declared properties %d", minProperties, len(properties)))
		}
	}
	if array {
		raw, ok := node["maxItems"]
		if !ok {
			return fmt.Errorf("operation %q %s declares an array without maxItems", operation, path)
		}
		maxItems, ok := raw.(float64)
		if !ok || math.Trunc(maxItems) != maxItems || maxItems < 0 {
			return fmt.Errorf("operation %q %s has invalid maxItems", operation, path)
		}
		if maxItems > maxStructuredRESTBodyItems {
			return fmt.Errorf("operation %q %s maxItems %.0f exceeds structured body limit %d", operation, path, maxItems, maxStructuredRESTBodyItems)
		}
		minItems, err := structuredRESTBodyArrayMinItems(node)
		if err != nil {
			return fmt.Errorf("operation %q %s %w", operation, path, err)
		}
		if minItems > int(maxItems) {
			return fmt.Errorf("operation %q %s minItems %d exceeds maxItems %.0f", operation, path, minItems, maxItems)
		}
		rawItems, hasItems := node["items"]
		rawPrefixItems, hasPrefixItems := node["prefixItems"]
		if !hasItems && !hasPrefixItems {
			return fmt.Errorf("operation %q %s declares an array without an items schema", operation, path)
		}
		if hasItems {
			items, ok := rawItems.(map[string]any)
			if !ok || len(items) == 0 {
				return fmt.Errorf("operation %q %s declares an array with an empty items schema", operation, path)
			}
			if err := requireClosedBoundedStructuredRESTBody(operation, items, path+"/items", depth+1, state); err != nil {
				return err
			}
		}
		if hasPrefixItems {
			prefixItems, ok := rawPrefixItems.([]any)
			if !ok || len(prefixItems) == 0 {
				return fmt.Errorf("operation %q %s declares an array with an empty prefixItems schema", operation, path)
			}
			for index, rawPrefixItem := range prefixItems {
				prefixItem, ok := rawPrefixItem.(map[string]any)
				if !ok || len(prefixItem) == 0 {
					return fmt.Errorf("operation %q %s/prefixItems/%d must be a non-empty schema object", operation, path, index)
				}
				if err := requireClosedBoundedStructuredRESTBody(operation, prefixItem, fmt.Sprintf("%s/prefixItems/%d", path, index), depth+1, state); err != nil {
					return err
				}
			}
			if !hasItems && maxItems > float64(len(prefixItems)) {
				return structuredRESTBodyFoundationGap(operation, path, "prefixItems does not constrain every allowed array item; declare an items schema for remaining items")
			}
		}
	}
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		return nil
	}
	state.fields += len(properties)
	if state.fields > maxStructuredRESTBodyFields {
		return fmt.Errorf("operation %q %s exceeds structured body field limit %d", operation, path, maxStructuredRESTBodyFields)
	}
	for _, name := range sortedMapKeys(properties) {
		child, ok := properties[name].(map[string]any)
		if !ok {
			return fmt.Errorf("operation %q %s/%s must be a schema object", operation, path, name)
		}
		if err := requireClosedBoundedStructuredRESTBody(operation, child, path+"/"+name, depth+1, state); err != nil {
			return err
		}
	}
	return nil
}

func structuredRESTBodyNodeCosts(node map[string]any) (int, int, error) {
	object, array := operationDirectWriteBodyNodeKinds(node)
	if object && array {
		return 0, 0, fmt.Errorf("declares both object and array structure")
	}
	minimum := 1
	maximum := 1
	if object {
		properties, ok := node["properties"].(map[string]any)
		if !ok {
			return 0, 0, fmt.Errorf("object has no properties")
		}
		required := make(map[string]struct{})
		if rawRequired, ok := node["required"].([]any); ok {
			for _, rawName := range rawRequired {
				name, ok := rawName.(string)
				if !ok {
					return 0, 0, fmt.Errorf("required contains a non-string property name")
				}
				required[name] = struct{}{}
			}
		}
		minProperties, err := structuredRESTBodyObjectMinProperties(node)
		if err != nil {
			return 0, 0, err
		}
		if minProperties > len(properties) {
			return 0, 0, fmt.Errorf("minProperties %d exceeds declared properties %d", minProperties, len(properties))
		}
		optionalMinimums := make([]int, 0, len(properties))
		for _, name := range sortedMapKeys(properties) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				return 0, 0, fmt.Errorf("properties.%s must be a schema object", name)
			}
			childMinimum, childMaximum, err := structuredRESTBodyNodeCosts(child)
			if err != nil {
				return 0, 0, err
			}
			maximum = structuredRESTBodyCostAdd(maximum, childMaximum)
			if _, isRequired := required[name]; isRequired {
				minimum = structuredRESTBodyCostAdd(minimum, childMinimum)
			} else {
				optionalMinimums = append(optionalMinimums, childMinimum)
			}
			if maximum > maxStructuredRESTBodyNodes {
				return minimum, maximum, nil
			}
		}
		if minProperties > len(required) {
			sort.Ints(optionalMinimums)
			for _, childMinimum := range optionalMinimums[:minProperties-len(required)] {
				minimum = structuredRESTBodyCostAdd(minimum, childMinimum)
				if minimum > maxStructuredRESTBodyNodes {
					return minimum, maximum, nil
				}
			}
		}
	}
	if array {
		minItems, err := structuredRESTBodyArrayMinItems(node)
		if err != nil {
			return 0, 0, err
		}
		maxItems, err := structuredRESTBodyArrayMaxItems(node)
		if err != nil {
			return 0, 0, err
		}
		for index := 0; index < maxItems; index++ {
			item, err := operationDirectWriteBodyArrayItemSchema(node, index)
			if err != nil {
				return 0, 0, err
			}
			itemMinimum, itemMaximum, err := structuredRESTBodyNodeCosts(item)
			if err != nil {
				return 0, 0, err
			}
			maximum = structuredRESTBodyCostAdd(maximum, itemMaximum)
			if index < minItems {
				minimum = structuredRESTBodyCostAdd(minimum, itemMinimum)
			}
			if maximum > maxStructuredRESTBodyNodes {
				return minimum, maximum, nil
			}
		}
	}
	return minimum, maximum, nil
}

func structuredRESTBodyCostAdd(left, right int) int {
	limit := maxStructuredRESTBodyNodes + 1
	if left >= limit || right >= limit || left > limit-right {
		return limit
	}
	return left + right
}

func structuredRESTBodyMinimumJSONBytes(node map[string]any) (int, error) {
	return structuredRESTBodyMinimumJSONBytesForStatic(node, nil, false)
}

func structuredRESTBodyMinimumJSONBytesWithStatic(node map[string]any, static any) (int, error) {
	return structuredRESTBodyMinimumJSONBytesForStatic(node, static, true)
}

func structuredRESTBodyMinimumJSONBytesForStatic(node map[string]any, static any, hasStatic bool) (int, error) {
	if hasStatic {
		switch value := static.(type) {
		case map[string]any:
			if !isObjectType(node) {
				return 0, fmt.Errorf("static value does not match an object schema")
			}
			return structuredRESTBodyMinimumJSONObjectBytesWithStatic(node, value)
		case []any:
			if !isArrayType(node) {
				return 0, fmt.Errorf("static value does not match an array schema")
			}
			return structuredRESTBodyMinimumJSONArrayBytesWithStatic(node, value)
		default:
			encoded, err := json.Marshal(static)
			if err != nil {
				return 0, err
			}
			return structuredRESTBodyByteCost(len(encoded)), nil
		}
	}
	if rawValues, exists := node["enum"]; exists {
		values, ok := rawValues.([]any)
		if !ok || len(values) == 0 {
			return 0, fmt.Errorf("enum must contain at least one value")
		}
		rawNode, err := json.Marshal(node)
		if err != nil {
			return 0, err
		}
		schema, err := compileStructuredRESTBodySchemaDocument(rawNode)
		if err != nil {
			return 0, err
		}
		minimum := maxOperationDirectWriteBytes + 1
		found := false
		for _, value := range values {
			if err := schema.Validate(value); err != nil {
				continue
			}
			found = true
			encoded, err := json.Marshal(value)
			if err != nil {
				return 0, err
			}
			cost := len(encoded)
			if cost > maxOperationDirectWriteBytes+1 {
				cost = maxOperationDirectWriteBytes + 1
			}
			if cost < minimum {
				minimum = cost
			}
		}
		if !found {
			return 0, fmt.Errorf("enum has no schema-valid value")
		}
		return minimum, nil
	}

	minimum := maxOperationDirectWriteBytes + 1
	found := false
	consider := func(value int) {
		if value < minimum {
			minimum = value
		}
		found = true
	}
	if structuredRESTBodyNodeAllowsType(node, "null") {
		consider(len("null"))
	}
	if structuredRESTBodyNodeAllowsType(node, "boolean") {
		consider(len("false"))
	}
	if structuredRESTBodyNodeAllowsType(node, "integer") || structuredRESTBodyNodeAllowsType(node, "number") {
		consider(1)
	}
	if structuredRESTBodyNodeAllowsType(node, "string") {
		value, err := structuredRESTBodyMinimumJSONStringBytes(node)
		if err == nil {
			consider(value)
		}
	}
	if isObjectType(node) {
		value, err := structuredRESTBodyMinimumJSONObjectBytesWithStatic(node, nil)
		if err != nil {
			return 0, err
		}
		consider(value)
	}
	if isArrayType(node) {
		value, err := structuredRESTBodyMinimumJSONArrayBytesWithStatic(node, nil)
		if err != nil {
			return 0, err
		}
		consider(value)
	}
	if !found || minimum > maxOperationDirectWriteBytes {
		if structuredRESTBodyNodeAllowsType(node, "string") {
			if _, err := structuredRESTBodyMinimumJSONStringBytes(node); err != nil {
				return 0, err
			}
		}
		return 0, fmt.Errorf("has no supported minimum JSON encoding")
	}
	return minimum, nil
}

func structuredRESTBodyMinimumJSONStringBytes(node map[string]any) (int, error) {
	candidates, err := structuredRESTBodyStringWitnessCandidates(node)
	if err != nil {
		return 0, err
	}
	rawNode, err := json.Marshal(node)
	if err != nil {
		return 0, err
	}
	schema, err := compileStructuredRESTBodySchemaDocument(rawNode)
	if err != nil {
		return 0, err
	}
	minimum := maxOperationDirectWriteBytes + 1
	for _, candidate := range candidates {
		if err := schema.Validate(candidate); err != nil {
			continue
		}
		encoded, err := json.Marshal(candidate)
		if err != nil {
			return 0, err
		}
		if len(encoded) < minimum {
			minimum = len(encoded)
		}
	}
	if minimum > maxOperationDirectWriteBytes {
		return 0, fmt.Errorf("cannot prove a schema-valid string witness")
	}
	return minimum, nil
}

func structuredRESTBodyStringWitnessCandidates(node map[string]any) ([]string, error) {
	pattern, hasPattern := node["pattern"].(string)
	format, _ := node["format"].(string)
	minLength, err := structuredRESTBodyStringMinLength(node)
	if err != nil {
		return nil, err
	}
	if !hasPattern {
		if _, exists := node["pattern"]; exists {
			return nil, fmt.Errorf("pattern must be a string")
		}
		if format == "uri" {
			return []string{structuredRESTBodyPadStringWitness("x:", minLength)}, nil
		}
		return []string{structuredRESTBodyPadStringWitness("", minLength)}, nil
	}
	witness, err := structuredRESTBodyPatternMinimumString(pattern)
	if err != nil {
		return nil, fmt.Errorf("cannot prove a schema-valid string witness: %w", err)
	}
	candidates := []string{witness}
	if format == "uri" {
		candidates = append(candidates, "x:"+witness, witness+"x:", "x:"+witness+"x:")
	}
	seen := make(map[string]struct{}, len(candidates))
	unique := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		if _, duplicate := seen[candidate]; duplicate {
			continue
		}
		seen[candidate] = struct{}{}
		unique = append(unique, candidate)
	}
	return unique, nil
}

// structuredRESTBodyStringMinLength mirrors the shared schema compiler's
// non-negative integral minLength rule. The structured-body normalizer holds
// this decoded node as ordinary JSON values, so accept only the exact number
// shapes json.Unmarshal creates and leave malformed forms to the compiler.
func structuredRESTBodyStringMinLength(node map[string]any) (int, error) {
	raw, exists := node["minLength"]
	if !exists {
		return 0, nil
	}
	value, ok := raw.(float64)
	if !ok || value < 0 || math.Trunc(value) != value || value > float64(maxOperationDirectWriteBytes) {
		return 0, fmt.Errorf("minLength must be a bounded non-negative integer")
	}
	return int(value), nil
}

func structuredRESTBodyPadStringWitness(prefix string, minLength int) string {
	if padding := minLength - utf8.RuneCountInString(prefix); padding > 0 {
		return prefix + strings.Repeat("x", padding)
	}
	return prefix
}

func structuredRESTBodyPatternMinimumString(pattern string) (string, error) {
	re, err := syntax.Parse(pattern, syntax.Perl)
	if err != nil {
		return "", err
	}
	return structuredRESTBodyRegexpMinimumString(re.Simplify())
}

func structuredRESTBodyRegexpMinimumString(re *syntax.Regexp) (string, error) {
	if re == nil {
		return "", fmt.Errorf("pattern is empty")
	}
	switch re.Op {
	case syntax.OpNoMatch:
		return "", fmt.Errorf("pattern has no matching value")
	case syntax.OpEmptyMatch, syntax.OpBeginLine, syntax.OpEndLine, syntax.OpBeginText, syntax.OpEndText:
		return "", nil
	case syntax.OpWordBoundary, syntax.OpNoWordBoundary:
		return "", fmt.Errorf("pattern has a word-boundary expression")
	case syntax.OpLiteral:
		var value strings.Builder
		for _, literal := range re.Rune {
			value.WriteString(structuredRESTBodyMinimumRuneJSONString(literal, re.Flags&syntax.FoldCase != 0))
		}
		return value.String(), nil
	case syntax.OpCharClass:
		if len(re.Rune) == 0 || len(re.Rune)%2 != 0 {
			return "", fmt.Errorf("pattern has an invalid character class")
		}
		minimum := ""
		found := false
		for index := 0; index < len(re.Rune); index += 2 {
			candidate, err := structuredRESTBodyMinimumRuneRangeJSONString(re.Rune[index], re.Rune[index+1], re.Flags&syntax.FoldCase != 0)
			if err != nil {
				continue
			}
			if !found || structuredRESTBodyJSONStringBytes(candidate) < structuredRESTBodyJSONStringBytes(minimum) {
				minimum = candidate
				found = true
			}
		}
		if !found {
			return "", fmt.Errorf("pattern has no matching character class")
		}
		return minimum, nil
	case syntax.OpAnyCharNotNL, syntax.OpAnyChar:
		return "a", nil
	case syntax.OpCapture, syntax.OpPlus:
		if len(re.Sub) != 1 {
			return "", fmt.Errorf("pattern has an invalid unary expression")
		}
		return structuredRESTBodyRegexpMinimumString(re.Sub[0])
	case syntax.OpStar, syntax.OpQuest:
		return "", nil
	case syntax.OpRepeat:
		if len(re.Sub) != 1 {
			return "", fmt.Errorf("pattern has an invalid repetition")
		}
		value, err := structuredRESTBodyRegexpMinimumString(re.Sub[0])
		if err != nil {
			return "", err
		}
		if value == "" || re.Min == 0 {
			return "", nil
		}
		if len(value) > maxOperationDirectWriteBytes/re.Min {
			return "", fmt.Errorf("pattern minimum exceeds supported body size")
		}
		return strings.Repeat(value, re.Min), nil
	case syntax.OpConcat:
		var value strings.Builder
		for _, child := range re.Sub {
			childValue, err := structuredRESTBodyRegexpMinimumString(child)
			if err != nil {
				return "", err
			}
			if value.Len() > maxOperationDirectWriteBytes-len(childValue) {
				return "", fmt.Errorf("pattern minimum exceeds supported body size")
			}
			value.WriteString(childValue)
		}
		return value.String(), nil
	case syntax.OpAlternate:
		if len(re.Sub) == 0 {
			return "", fmt.Errorf("pattern has no alternatives")
		}
		minimum := ""
		found := false
		for _, child := range re.Sub {
			value, err := structuredRESTBodyRegexpMinimumString(child)
			if err != nil {
				continue
			}
			if !found || structuredRESTBodyJSONStringBytes(value) < structuredRESTBodyJSONStringBytes(minimum) {
				minimum = value
				found = true
			}
		}
		if !found {
			return "", fmt.Errorf("pattern has no provable matching alternative")
		}
		return minimum, nil
	default:
		return "", fmt.Errorf("pattern has unsupported expression")
	}
}

func structuredRESTBodyMinimumRuneJSONString(value rune, foldCase bool) string {
	minimum := string(value)
	if !foldCase {
		return minimum
	}
	for candidate := unicode.SimpleFold(value); candidate != value; candidate = unicode.SimpleFold(candidate) {
		candidateValue := string(candidate)
		if structuredRESTBodyJSONStringBytes(candidateValue) < structuredRESTBodyJSONStringBytes(minimum) {
			minimum = candidateValue
		}
	}
	return minimum
}

func structuredRESTBodyMinimumRuneRangeJSONString(first, last rune, foldCase bool) (string, error) {
	minimum := ""
	found := false
	consider := func(candidate rune) {
		if candidate < first || candidate > last {
			return
		}
		candidateValue := structuredRESTBodyMinimumRuneJSONString(candidate, foldCase)
		if !found || structuredRESTBodyJSONStringBytes(candidateValue) < structuredRESTBodyJSONStringBytes(minimum) {
			minimum = candidateValue
			found = true
		}
	}
	for candidate := rune(0); candidate <= 0x7f; candidate++ {
		consider(candidate)
	}
	for _, candidate := range []rune{first, last, 0x80, 0x800, 0x2028, 0x2029, 0x202a, 0x10000} {
		consider(candidate)
	}
	if !found {
		return "", fmt.Errorf("pattern has no matching character")
	}
	return minimum, nil
}

func structuredRESTBodyJSONStringBytes(value string) int {
	encoded, err := json.Marshal(value)
	if err != nil {
		return maxOperationDirectWriteBytes + 1
	}
	return len(encoded)
}

func structuredRESTBodyMinimumJSONObjectBytesWithStatic(node map[string]any, staticBody map[string]any) (int, error) {
	properties, ok := node["properties"].(map[string]any)
	if !ok {
		return 0, fmt.Errorf("object has no properties")
	}
	required := make(map[string]struct{})
	if rawRequired, ok := node["required"].([]any); ok {
		for _, rawName := range rawRequired {
			name, ok := rawName.(string)
			if !ok {
				return 0, fmt.Errorf("required contains a non-string property name")
			}
			required[name] = struct{}{}
		}
	}
	minimumProperties, err := structuredRESTBodyObjectMinProperties(node)
	if err != nil {
		return 0, err
	}
	minimum := 2
	fields := 0
	optionalCosts := make([]int, 0, len(properties))
	for _, name := range sortedMapKeys(properties) {
		child, ok := properties[name].(map[string]any)
		if !ok {
			return 0, fmt.Errorf("properties.%s must be a schema object", name)
		}
		encodedName, err := json.Marshal(name)
		if err != nil {
			return 0, err
		}
		if staticValue, hasStatic := staticBody[name]; hasStatic {
			childCost, err := structuredRESTBodyMinimumJSONBytesForStatic(child, staticValue, true)
			if err != nil {
				return 0, err
			}
			minimum = structuredRESTBodyByteCostAdd(minimum, structuredRESTBodyByteCostAdd(len(encodedName)+1, childCost))
			fields++
			continue
		}
		childCost, err := structuredRESTBodyMinimumJSONBytes(child)
		if err != nil {
			return 0, err
		}
		cost := structuredRESTBodyByteCostAdd(len(encodedName)+1, childCost)
		if _, isRequired := required[name]; isRequired {
			minimum = structuredRESTBodyByteCostAdd(minimum, cost)
			fields++
		} else {
			optionalCosts = append(optionalCosts, cost)
		}
	}
	if minimumProperties > len(properties) {
		return 0, fmt.Errorf("minProperties %d exceeds declared properties %d", minimumProperties, len(properties))
	}
	if minimumProperties > fields {
		sort.Ints(optionalCosts)
		for _, cost := range optionalCosts[:minimumProperties-fields] {
			minimum = structuredRESTBodyByteCostAdd(minimum, cost)
		}
		fields = minimumProperties
	}
	if fields > 1 {
		minimum = structuredRESTBodyByteCostAdd(minimum, fields-1)
	}
	return minimum, nil
}

func structuredRESTBodyMinimumJSONArrayBytesWithStatic(node map[string]any, staticBody []any) (int, error) {
	minimumItems, err := structuredRESTBodyArrayMinItems(node)
	if err != nil {
		return 0, err
	}
	minimum := 2
	items := len(staticBody)
	if minimumItems > items {
		items = minimumItems
	}
	for index := 0; index < items; index++ {
		item, err := operationDirectWriteBodyArrayItemSchema(node, index)
		if err != nil {
			return 0, err
		}
		var cost int
		if index < len(staticBody) {
			cost, err = structuredRESTBodyMinimumJSONBytesForStatic(item, staticBody[index], true)
		} else {
			cost, err = structuredRESTBodyMinimumJSONBytes(item)
		}
		if err != nil {
			return 0, err
		}
		minimum = structuredRESTBodyByteCostAdd(minimum, cost)
		if index > 0 {
			minimum = structuredRESTBodyByteCostAdd(minimum, 1)
		}
	}
	return minimum, nil
}

func structuredRESTBodyByteCostAdd(left, right int) int {
	limit := maxOperationDirectWriteBytes + 1
	if left >= limit || right >= limit || left > limit-right {
		return limit
	}
	return left + right
}

func structuredRESTBodyByteCost(value int) int {
	if value > maxOperationDirectWriteBytes {
		return maxOperationDirectWriteBytes + 1
	}
	return value
}

func normalizeStructuredRESTBodySchemaNode(operation string, node map[string]any, path string) error {
	types, hasExplicitType, err := structuredRESTBodyExplicitTypes(operation, path, node)
	if err != nil {
		return err
	}
	_, hasProperties := node["properties"]
	_, hasAdditionalProperties := node["additionalProperties"]
	_, hasItems := node["items"]
	_, hasPrefixItems := node["prefixItems"]
	hasObjectStructure := hasProperties || hasAdditionalProperties
	hasArrayStructure := hasItems || hasPrefixItems

	if !hasExplicitType {
		switch {
		case hasObjectStructure && hasArrayStructure:
			return structuredRESTBodyFoundationGap(operation, path, "has both object and array structure without an explicit type")
		case hasObjectStructure:
			node["type"] = "object"
		case hasArrayStructure:
			node["type"] = "array"
		}
	} else {
		if hasObjectStructure && !structuredRESTBodyTypeAllows(types, "object") {
			return structuredRESTBodyFoundationGap(operation, path, "object structure conflicts with its explicit type")
		}
		if hasArrayStructure && !structuredRESTBodyTypeAllows(types, "array") {
			return structuredRESTBodyFoundationGap(operation, path, "array structure conflicts with its explicit type")
		}
	}

	if hasProperties {
		properties, ok := node["properties"].(map[string]any)
		if !ok {
			return structuredRESTBodyFoundationGap(operation, path, "properties must be an object of schema objects")
		}
		for _, name := range sortedMapKeys(properties) {
			child, ok := properties[name].(map[string]any)
			if !ok {
				return structuredRESTBodyFoundationGap(operation, path+"/"+name, "must be a schema object")
			}
			if err := normalizeStructuredRESTBodySchemaNode(operation, child, path+"/"+name); err != nil {
				return err
			}
		}
	}
	if hasItems {
		items, ok := node["items"].(map[string]any)
		if !ok {
			return structuredRESTBodyFoundationGap(operation, path+"/items", "must be a schema object")
		}
		if err := normalizeStructuredRESTBodySchemaNode(operation, items, path+"/items"); err != nil {
			return err
		}
	}
	if hasPrefixItems {
		prefixItems, ok := node["prefixItems"].([]any)
		if !ok {
			return structuredRESTBodyFoundationGap(operation, path+"/prefixItems", "must be an array of schema objects")
		}
		for index, rawPrefixItem := range prefixItems {
			prefixItem, ok := rawPrefixItem.(map[string]any)
			if !ok {
				return structuredRESTBodyFoundationGap(operation, fmt.Sprintf("%s/prefixItems/%d", path, index), "must be a schema object")
			}
			if err := normalizeStructuredRESTBodySchemaNode(operation, prefixItem, fmt.Sprintf("%s/prefixItems/%d", path, index)); err != nil {
				return err
			}
		}
	}
	return nil
}

func structuredRESTBodyExplicitTypes(operation string, path string, node map[string]any) ([]string, bool, error) {
	raw, ok := node["type"]
	if !ok {
		return nil, false, nil
	}
	switch types := raw.(type) {
	case string:
		if types == "" {
			return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must be a non-empty JSON type")
		}
		return []string{types}, true, nil
	case []any:
		if len(types) == 0 {
			return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must be a non-empty JSON type list")
		}
		declared := make([]string, 0, len(types))
		for _, rawType := range types {
			typeName, ok := rawType.(string)
			if !ok || typeName == "" {
				return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must contain only non-empty JSON type names")
			}
			declared = append(declared, typeName)
		}
		return declared, true, nil
	default:
		return nil, false, structuredRESTBodyFoundationGap(operation, path, "type must be a JSON type name or list")
	}
}

func structuredRESTBodyTypeAllows(types []string, want string) bool {
	for _, declared := range types {
		if declared == want {
			return true
		}
	}
	return false
}

func structuredRESTBodyFoundationGap(operation string, path string, detail string) error {
	return fmt.Errorf("operation %q %s has a structured-body foundation gap: %s", operation, path, detail)
}

func structuredRESTBodyNodeMayAcceptObjectOrArray(node map[string]any) bool {
	if enum, ok := node["enum"].([]any); ok && len(enum) > 0 {
		for _, value := range enum {
			if isStructuredJSONBodyValue(value) {
				return true
			}
		}
		return false
	}

	switch types := node["type"].(type) {
	case string:
		return types == "object" || types == "array"
	case []any:
		if len(types) == 0 {
			return true
		}
		for _, value := range types {
			typeName, ok := value.(string)
			if !ok || typeName == "object" || typeName == "array" {
				return true
			}
		}
		return false
	default:
		return true
	}
}

func structuredRESTBodyNodeHasExplicitType(node map[string]any) bool {
	switch types := node["type"].(type) {
	case string:
		return types != ""
	case []any:
		return len(types) > 0
	default:
		return false
	}
}

func isStructuredJSONBodyValue(value any) bool {
	if value == nil {
		return false
	}
	if reflect.ValueOf(value).Kind() == reflect.Map {
		return true
	}
	_, ok := arrayElements(value)
	return ok
}

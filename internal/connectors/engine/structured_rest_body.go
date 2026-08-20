package engine

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"mime"
	"reflect"
	"sort"
	"strconv"
	"strings"
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
// structured input names a closed, bounded top-level field of its fixed
// operation. It is deliberately a name-only preflight: callers never provide
// a body template, route, method, media type, or arbitrary nested path.
func PreflightOperationStructuredJSONBodyField(b Bundle, operation, field string) error {
	op, err := findOperation(b, operation)
	if err != nil {
		return err
	}
	return ValidateOperationStructuredJSONBodyField(op, field)
}

// ValidateOperationStructuredJSONBodyField is the common declaration gate for
// a json command flag on a fixed write operation. GraphQL retains its existing
// closed-variable contract; REST uses the equivalent closed body-schema
// contract. Commandrunner and connectorgen call this same function; typed
// execution compiles the same schema at its shared body boundary.
func ValidateOperationStructuredJSONBodyField(op OperationSpec, field string) error {
	field = strings.TrimSpace(field)
	if err := safety.ValidateIdentifier(field, "structured body field"); err != nil {
		return err
	}
	if strings.Contains(field, ".") {
		return fmt.Errorf("operation %q structured body field %q must be top-level", op.ID, field)
	}

	switch op.Kind {
	case "graphql_mutation":
		return ValidateGraphQLOperationStructuredJSONVariable(op, field)
	case "rest_write":
		return validateRESTStructuredJSONBodyField(op, field)
	default:
		return fmt.Errorf("operation %q structured JSON body requires rest_write or graphql_mutation, got %q", op.ID, op.Kind)
	}
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
	properties, ok := root["properties"].(map[string]any)
	if !ok {
		return fmt.Errorf("operation %q body_schema must declare properties", op.ID)
	}
	raw, ok := properties[field]
	if !ok {
		return fmt.Errorf("operation %q body_schema does not declare structured field %q", op.ID, field)
	}
	node, ok := raw.(map[string]any)
	if !ok || (!isObjectType(node) && !isArrayType(node)) {
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
	mappings := make([]CLIFlag, 0, len(flags))
	for _, flag := range flags {
		path, ok := strings.CutPrefix(strings.TrimSpace(flag.MapsTo), "body.")
		if !ok || path == "" {
			continue
		}
		node, err := operationDirectWriteBodySchemaPath(compiled.root, path)
		if err != nil {
			return fmt.Errorf("operation %q CLI flag --%s body field %q: %w", op.ID, flag.Name, path, err)
		}
		if err := validateOperationDirectWriteCLIFlagType(op, flag, path, node); err != nil {
			return err
		}
		mappings = append(mappings, flag)
	}
	for _, path := range structuredRESTBodyRequiredMappingPaths(compiled.root, "") {
		if structuredRESTBodyStaticValueAtPath(op.REST.Body, path) {
			continue
		}
		flag, found := structuredRESTBodyRequiredFlag(path, mappings)
		if !found {
			return fmt.Errorf("operation %q requires body.%s but no command flag maps to it and rest.body does not provide it", op.ID, path)
		}
		if !flag.Required {
			return fmt.Errorf("operation %q requires body.%s but CLI flag --%s is not required", op.ID, path, flag.Name)
		}
	}
	return nil
}

func validateOperationDirectWriteCLIFlagType(op OperationSpec, flag CLIFlag, path string, node map[string]any) error {
	valid := false
	switch flag.Type {
	case "json":
		valid = !strings.Contains(path, ".") && (isObjectType(node) || isArrayType(node))
	case "", "string", "enum":
		valid = structuredRESTBodyNodeAllowsType(node, "string")
	case "integer":
		valid = structuredRESTBodyNodeAllowsType(node, "integer") || structuredRESTBodyNodeAllowsType(node, "number")
	case "number":
		valid = structuredRESTBodyNodeAllowsType(node, "number")
	case "boolean":
		valid = structuredRESTBodyNodeAllowsType(node, "boolean")
	case "string_array":
		valid = structuredRESTBodyArrayAllowsStrings(node)
		if valid {
			maxItems, err := structuredRESTBodyArrayMaxItems(node)
			if err != nil {
				return fmt.Errorf("operation %q CLI flag --%s body field %q: %w", op.ID, flag.Name, path, err)
			}
			if flag.MaxItems > maxItems || flag.MinItems > maxItems {
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

func structuredRESTBodyRequiredMappingPaths(node map[string]any, prefix string) []string {
	properties, _ := node["properties"].(map[string]any)
	required, _ := node["required"].([]any)
	names := make([]string, 0, len(required))
	for _, rawName := range required {
		name, ok := rawName.(string)
		if !ok || name == "" {
			continue
		}
		names = append(names, name)
	}
	sort.Strings(names)
	paths := make([]string, 0, len(names))
	for _, name := range names {
		child, ok := properties[name].(map[string]any)
		if !ok {
			continue
		}
		path := name
		if prefix != "" {
			path = prefix + "." + name
		}
		children := []string(nil)
		if isObjectType(child) {
			children = structuredRESTBodyRequiredMappingPaths(child, path)
		}
		if len(children) == 0 {
			paths = append(paths, path)
			continue
		}
		paths = append(paths, children...)
	}
	return paths
}

func structuredRESTBodyStaticValueAtPath(body map[string]any, path string) bool {
	var current any = body
	for _, part := range strings.Split(path, ".") {
		switch value := current.(type) {
		case map[string]any:
			var ok bool
			current, ok = value[part]
			if !ok {
				return false
			}
		case []any:
			index, err := operationDirectWriteBodyArrayIndex(part)
			if err != nil || index >= len(value) {
				return false
			}
			current = value[index]
		default:
			return false
		}
	}
	return true
}

func structuredRESTBodyRequiredFlag(path string, flags []CLIFlag) (CLIFlag, bool) {
	var optional CLIFlag
	foundOptional := false
	for _, flag := range flags {
		target, ok := strings.CutPrefix(strings.TrimSpace(flag.MapsTo), "body.")
		if !ok {
			continue
		}
		covers := target == path || (flag.Type == "json" && strings.HasPrefix(path, target+"."))
		if !covers {
			continue
		}
		if flag.Required {
			return flag, true
		}
		if !foundOptional {
			optional = flag
			foundOptional = true
		}
	}
	return optional, foundOptional
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

func operationHasStructuredRESTBodyValue(body map[string]any) bool {
	for _, value := range body {
		if isStructuredJSONBodyValue(value) {
			return true
		}
	}
	return false
}

type structuredRESTBodySchemaCompilation struct {
	root   map[string]any
	schema *Schema
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
	return &structuredRESTBodySchemaCompilation{root: root, schema: sch}, nil
}

func materializeStructuredRESTBody(op OperationSpec, staticBody, overrides map[string]any) (map[string]any, error) {
	compiled, err := compileStructuredRESTBodySchema(op)
	if err != nil {
		return nil, err
	}
	if len(staticBody) > maxStructuredRESTBodyFields || len(overrides) > maxStructuredRESTBodyFields {
		return nil, fmt.Errorf("operation %q: body exceeds structured body field limit %d", op.ID, maxStructuredRESTBodyFields)
	}
	body := make(map[string]any, len(staticBody)+len(overrides))
	for key, value := range staticBody {
		body[key] = value
	}
	for key, value := range overrides {
		body[key] = value
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
		number := strconv.FormatFloat(value.Float(), 'g', -1, value.Type().Bits())
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
	if isObjectType(node) {
		if closed, ok := node["additionalProperties"].(bool); !ok || closed {
			return fmt.Errorf("operation %q %s is an object and must declare additionalProperties: false", operation, path)
		}
	}
	if isArrayType(node) {
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

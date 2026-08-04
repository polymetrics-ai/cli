package engine

import (
	"encoding/json"
	"fmt"
	"sort"
	"strconv"
	"strings"
)

type requestSchemaCollector struct {
	root   map[string]any
	active map[string]bool
}

func inlineRequestSchema(raw json.RawMessage) (json.RawMessage, error) {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return nil, fmt.Errorf("decode body_schema: %w", err)
	}
	if root == nil {
		return nil, fmt.Errorf("body_schema must be an object")
	}
	collector := requestSchemaCollector{root: root, active: map[string]bool{}}
	inlined, err := collector.inline(root)
	if err != nil {
		return nil, err
	}
	encoded, err := json.Marshal(inlined)
	if err != nil {
		return nil, fmt.Errorf("encode inlined body_schema: %w", err)
	}
	return encoded, nil
}

func (c requestSchemaCollector) inline(schema map[string]any) (map[string]any, error) {
	if rawRef, ok := schema["$ref"]; ok {
		ref, ok := rawRef.(string)
		if !ok || strings.TrimSpace(ref) == "" {
			return nil, fmt.Errorf("body_schema $ref must be a non-empty string")
		}
		resolved, err := c.resolve(ref)
		if err != nil {
			return nil, err
		}
		if c.active[ref] {
			return nil, fmt.Errorf("body_schema reference cycle at %q", ref)
		}
		resolvedSchema, ok := resolved.(map[string]any)
		if !ok {
			return nil, fmt.Errorf("body_schema local reference %q does not resolve to an object schema", ref)
		}
		c.active[ref] = true
		target, err := c.inline(resolvedSchema)
		delete(c.active, ref)
		if err != nil {
			return nil, err
		}

		siblings := make(map[string]any, len(schema)-1)
		for key, value := range schema {
			if key != "$ref" {
				siblings[key] = value
			}
		}
		if len(siblings) == 0 {
			return target, nil
		}
		inlinedSiblings, err := c.inline(siblings)
		if err != nil {
			return nil, err
		}
		return map[string]any{"allOf": []any{target, inlinedSiblings}}, nil
	}

	inlined := make(map[string]any, len(schema))
	for key, value := range schema {
		switch key {
		case "$defs", "definitions", "components":
			continue
		case "properties":
			properties, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("body_schema properties must be an object")
			}
			inlinedProperties := make(map[string]any, len(properties))
			for name, rawProperty := range properties {
				property, ok := rawProperty.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("body_schema property %q must be an object schema", name)
				}
				inlinedProperty, err := c.inline(property)
				if err != nil {
					return nil, err
				}
				inlinedProperties[name] = inlinedProperty
			}
			inlined[key] = inlinedProperties
		case "items":
			items, ok := value.(map[string]any)
			if !ok {
				return nil, fmt.Errorf("body_schema items must be an object schema")
			}
			inlinedItems, err := c.inline(items)
			if err != nil {
				return nil, err
			}
			inlined[key] = inlinedItems
		case "allOf", "anyOf", "oneOf":
			branches, ok := value.([]any)
			if !ok || len(branches) == 0 {
				return nil, fmt.Errorf("body_schema %s must be a non-empty array", key)
			}
			inlinedBranches := make([]any, 0, len(branches))
			for i, rawBranch := range branches {
				branch, ok := rawBranch.(map[string]any)
				if !ok {
					return nil, fmt.Errorf("body_schema %s[%d] must be an object schema", key, i)
				}
				inlinedBranch, err := c.inline(branch)
				if err != nil {
					return nil, err
				}
				inlinedBranches = append(inlinedBranches, inlinedBranch)
			}
			inlined[key] = inlinedBranches
		default:
			inlined[key] = value
		}
	}
	return inlined, nil
}

func collectRequestSchemaFields(raw json.RawMessage, prefix string, fields map[string]bool) error {
	var root map[string]any
	if err := json.Unmarshal(raw, &root); err != nil {
		return fmt.Errorf("decode body_schema: %w", err)
	}
	if root == nil {
		return fmt.Errorf("body_schema must be an object")
	}
	collector := requestSchemaCollector{root: root, active: map[string]bool{}}
	found, err := collector.collect(root, prefix, fields, true)
	if err != nil {
		return err
	}
	if !found {
		return fmt.Errorf("body_schema does not enumerate request fields")
	}
	return nil
}

func (c requestSchemaCollector) collect(schema map[string]any, prefix string, fields map[string]bool, root bool) (bool, error) {
	_ = root
	found := false
	if requestSchemaIsObject(schema) {
		closed, ok := schema["additionalProperties"].(bool)
		if !ok || closed {
			return false, fmt.Errorf("body_schema object at %q must declare additionalProperties false", prefix)
		}
	}
	if rawRef, ok := schema["$ref"]; ok {
		ref, ok := rawRef.(string)
		if !ok || strings.TrimSpace(ref) == "" {
			return false, fmt.Errorf("body_schema $ref must be a non-empty string")
		}
		resolved, err := c.resolve(ref)
		if err != nil {
			return false, err
		}
		if c.active[ref] {
			return false, fmt.Errorf("body_schema reference cycle at %q", ref)
		}
		resolvedSchema, ok := resolved.(map[string]any)
		if !ok {
			return false, fmt.Errorf("body_schema local reference %q does not resolve to an object schema", ref)
		}
		c.active[ref] = true
		refFound, err := c.collect(resolvedSchema, prefix, fields, root)
		delete(c.active, ref)
		if err != nil {
			return false, err
		}
		found = found || refFound
	}

	for _, keyword := range []string{"allOf", "anyOf", "oneOf"} {
		rawBranches, ok := schema[keyword]
		if !ok {
			continue
		}
		branches, ok := rawBranches.([]any)
		if !ok || len(branches) == 0 {
			return false, fmt.Errorf("body_schema %s must be a non-empty array", keyword)
		}
		for i, rawBranch := range branches {
			branch, ok := rawBranch.(map[string]any)
			if !ok {
				return false, fmt.Errorf("body_schema %s[%d] must be an object schema", keyword, i)
			}
			branchFound, err := c.collect(branch, prefix, fields, root)
			if err != nil {
				return false, err
			}
			found = found || branchFound
		}
	}

	if rawProperties, ok := schema["properties"]; ok {
		properties, ok := rawProperties.(map[string]any)
		if !ok {
			return false, fmt.Errorf("body_schema properties must be an object")
		}
		names := make([]string, 0, len(properties))
		for name := range properties {
			names = append(names, name)
		}
		sort.Strings(names)
		for _, name := range names {
			path := appendRequestFieldPointer(prefix, name)
			fields[path] = true
			found = true
			child, ok := properties[name].(map[string]any)
			if !ok {
				continue
			}
			if _, err := c.collect(child, path, fields, false); err != nil {
				return false, err
			}
		}
	}

	if rawItems, ok := schema["items"]; ok {
		items, ok := rawItems.(map[string]any)
		if !ok {
			return false, fmt.Errorf("body_schema items must be an object schema")
		}
		itemPrefix := appendRequestFieldPointer(prefix, "0")
		fields[itemPrefix] = true
		found = true
		itemFound, err := c.collect(items, itemPrefix, fields, false)
		if err != nil {
			return false, err
		}
		found = found || itemFound
	}
	return found, nil
}

func requestSchemaIsObject(schema map[string]any) bool {
	if _, ok := schema["properties"]; ok {
		return true
	}
	switch rawType := schema["type"].(type) {
	case string:
		return rawType == "object"
	case []any:
		for _, value := range rawType {
			if value == "object" {
				return true
			}
		}
	}
	return false
}

func (c requestSchemaCollector) resolve(ref string) (any, error) {
	if !strings.HasPrefix(ref, "#/") {
		if strings.HasPrefix(ref, "#") {
			return nil, fmt.Errorf("body_schema cannot resolve local reference %q", ref)
		}
		return nil, fmt.Errorf("body_schema external reference %q cannot be resolved locally", ref)
	}
	var current any = c.root
	for _, rawToken := range strings.Split(strings.TrimPrefix(ref, "#/"), "/") {
		token, err := decodeJSONPointerToken(rawToken)
		if err != nil {
			return nil, fmt.Errorf("body_schema cannot resolve local reference %q: %w", ref, err)
		}
		switch value := current.(type) {
		case map[string]any:
			next, exists := value[token]
			if !exists {
				return nil, fmt.Errorf("body_schema cannot resolve local reference %q", ref)
			}
			current = next
		case []any:
			index, err := strconv.Atoi(token)
			if err != nil || index < 0 || index >= len(value) {
				return nil, fmt.Errorf("body_schema cannot resolve local reference %q", ref)
			}
			current = value[index]
		default:
			return nil, fmt.Errorf("body_schema cannot resolve local reference %q", ref)
		}
	}
	return current, nil
}

func decodeJSONPointerToken(token string) (string, error) {
	var decoded strings.Builder
	for i := 0; i < len(token); i++ {
		if token[i] != '~' {
			decoded.WriteByte(token[i])
			continue
		}
		if i+1 >= len(token) || token[i+1] != '0' && token[i+1] != '1' {
			return "", fmt.Errorf("invalid JSON pointer escape in %q", token)
		}
		i++
		if token[i] == '0' {
			decoded.WriteByte('~')
		} else {
			decoded.WriteByte('/')
		}
	}
	return decoded.String(), nil
}

func declaredWriteRequestFields(action WriteAction) ([]string, error) {
	if action.DynamicFields != nil {
		return nil, fmt.Errorf("dynamic_fields request keys cannot be enumerated for request-contract citations")
	}
	if bodyTypeOf(action) == "base64_upload" && len(action.Query) > 0 {
		return nil, fmt.Errorf("base64_upload query parameters are not transmitted by the runtime")
	}
	fields := map[string]bool{}
	pathFields, err := declaredWritePathFields(action)
	if err != nil {
		return nil, err
	}
	for _, name := range pathFields {
		fields[requestFieldPointer("path", name)] = true
	}
	for name := range action.Query {
		if strings.TrimSpace(name) == "" {
			return nil, fmt.Errorf("query contains an empty parameter name")
		}
		if name != strings.TrimSpace(name) {
			return nil, fmt.Errorf("query parameter %q contains surrounding whitespace", name)
		}
		fields[requestFieldPointer("query", name)] = true
	}

	switch bodyTypeOf(action) {
	case "none":
		if err := addSelectedWriteBodyFields(fields, action, action.BodyFields); err != nil {
			return nil, err
		}
	case "graphql":
		if action.GraphQL != nil {
			collectGraphQLRequestFields(action.GraphQL.Variables, requestFieldPointer("body", "variables"), fields)
		}
	case "json_array":
		if err := validateJSONArrayRequestSchema(action.BodySchema); err != nil {
			return nil, err
		}
		if err := collectRequestSchemaFields(action.BodySchema, "/body", fields); err != nil {
			return nil, fmt.Errorf("body_schema: %w", err)
		}
	case "multipart":
		if action.Multipart != nil {
			for _, part := range action.Multipart.Parts {
				if part.Name != strings.TrimSpace(part.Name) {
					return nil, fmt.Errorf("multipart part name %q contains surrounding whitespace", part.Name)
				}
				fields[requestFieldPointer("body", part.Name)] = true
			}
		}
	case "base64_upload":
		if err := addWriteJSONBodyFields(fields, action); err != nil {
			return nil, err
		}
		if action.Base64Upload != nil {
			deleteRequestFieldPrefix(fields, requestFieldPointer("body", action.Base64Upload.SourceField))
			fields[requestFieldPointer("body", action.Base64Upload.ContentField)] = true
		}
	case "form":
		formAction := action
		formAction.BodyFields = nil
		if err := addWriteJSONBodyFields(fields, formAction); err != nil {
			return nil, err
		}
	default:
		if len(action.BodyFields) > 0 {
			if err := addSelectedWriteBodyFields(fields, action, action.BodyFields); err != nil {
				return nil, err
			}
		} else if err := addWriteJSONBodyFields(fields, action); err != nil {
			return nil, err
		}
	}

	paths := make([]string, 0, len(fields))
	for path := range fields {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
}

func validateJSONArrayRequestSchema(raw json.RawMessage) error {
	if len(raw) == 0 {
		return fmt.Errorf("json_array body_schema is required")
	}
	schema, err := CompileSchema(raw)
	if err != nil {
		return fmt.Errorf("body_schema: %w", err)
	}
	rootTypes := schema.node.mappingTypeSet()
	if rootTypes.unsatisfiable {
		return fmt.Errorf("json_array body_schema root has unsatisfiable allOf type constraints")
	}
	if !rootTypes.constrained || len(rootTypes.types) != 1 || rootTypes.types[0] != "array" {
		got := "unconstrained"
		if rootTypes.constrained {
			got = strings.Join(rootTypes.types, ",")
		}
		return fmt.Errorf("json_array body_schema root must be exclusively array, got %s", got)
	}
	return nil
}

func declaredWritePathFields(action WriteAction) ([]string, error) {
	declared := make(map[string]bool, len(action.PathFields))
	for _, rawName := range action.PathFields {
		name := strings.TrimSpace(rawName)
		if name == "" {
			return nil, fmt.Errorf("path_fields contains an empty field")
		}
		if declared[name] {
			return nil, fmt.Errorf("path_fields duplicates %q", name)
		}
		declared[name] = true
	}

	referenced := map[string]bool{}
	for _, match := range templatePattern.FindAllStringSubmatch(action.Path, -1) {
		expr := strings.TrimSpace(match[1])
		if paths, ok, err := coalesceRecordPathsExpression(expr); ok || err != nil {
			if err != nil {
				return nil, fmt.Errorf("path template: %w", err)
			}
			for _, path := range paths {
				referenced[path] = true
			}
			continue
		}
		ref := strings.TrimSpace(strings.SplitN(expr, "|", 2)[0])
		if strings.HasPrefix(ref, "record.") {
			name := strings.TrimSpace(strings.TrimPrefix(ref, "record."))
			if name == "" {
				return nil, fmt.Errorf("path template contains an empty record reference")
			}
			referenced[name] = true
		}
	}
	for name := range referenced {
		if !declared[name] {
			return nil, fmt.Errorf("path template record field %q is missing from path_fields", name)
		}
	}
	for name := range declared {
		if !referenced[name] {
			return nil, fmt.Errorf("path_fields entry %q is not used by the path template", name)
		}
	}

	paths := make([]string, 0, len(declared))
	for name := range declared {
		paths = append(paths, name)
	}
	sort.Strings(paths)
	return paths, nil
}

func addSelectedWriteBodyFields(fields map[string]bool, action WriteAction, names []string) error {
	if len(names) == 0 {
		return nil
	}
	declared := map[string]bool{}
	if err := collectRequestSchemaFields(action.RecordSchema, "/body", declared); err != nil {
		return fmt.Errorf("record_schema: %w", err)
	}
	for _, rawName := range names {
		name := strings.TrimSpace(rawName)
		if name == "" {
			continue
		}
		prefix := requestFieldPointer("body", name)
		matched := false
		for path := range declared {
			if requestFieldPointerWithin(path, prefix) {
				fields[path] = true
				matched = true
			}
		}
		if !matched {
			return fmt.Errorf("body_fields entry %q is not declared by record_schema", name)
		}
	}
	return nil
}

func addWriteJSONBodyFields(fields map[string]bool, action WriteAction) error {
	if len(action.BodyFields) > 0 {
		return addSelectedWriteBodyFields(fields, action, action.BodyFields)
	}
	bodyFields := map[string]bool{}
	if err := collectRequestSchemaFields(action.RecordSchema, "/body", bodyFields); err != nil {
		return fmt.Errorf("record_schema: %w", err)
	}
	excluded := map[string]bool{}
	for _, field := range action.PathFields {
		if !strings.Contains(field, ".") {
			excluded[requestFieldPointer("body", field)] = true
		}
	}
	for path := range bodyFields {
		excludedPath := false
		for prefix := range excluded {
			if requestFieldPointerWithin(path, prefix) {
				excludedPath = true
				break
			}
		}
		if !excludedPath {
			fields[path] = true
		}
	}
	return nil
}

func deleteRequestFieldPrefix(fields map[string]bool, prefix string) {
	for path := range fields {
		if requestFieldPointerWithin(path, prefix) {
			delete(fields, path)
		}
	}
}

func requestFieldPointerWithin(path, prefix string) bool {
	return path == prefix || strings.HasPrefix(path, prefix+"/")
}

func collectGraphQLRequestFields(values map[string]any, prefix string, fields map[string]bool) {
	names := make([]string, 0, len(values))
	for name := range values {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path := appendRequestFieldPointer(prefix, name)
		fields[path] = true
		collectGraphQLRequestField(values[name], path, fields)
	}
}

func collectGraphQLRequestField(value any, path string, fields map[string]bool) {
	switch nested := value.(type) {
	case map[string]any:
		if _, descriptor := nested["template"]; descriptor {
			return
		}
		collectGraphQLRequestFields(nested, path, fields)
	case []any:
		elementPath := appendRequestFieldPointer(path, "0")
		fields[elementPath] = true
		for _, item := range nested {
			collectGraphQLRequestField(item, elementPath, fields)
		}
	}
}

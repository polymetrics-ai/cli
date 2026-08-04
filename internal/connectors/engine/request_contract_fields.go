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
	found := false
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
			path := prefix + "." + name
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
		if root {
			return false, fmt.Errorf("body_schema root arrays cannot be represented as cited request fields")
		}
		items, ok := rawItems.(map[string]any)
		if !ok {
			return false, fmt.Errorf("body_schema items must be an object schema")
		}
		itemFound, err := c.collect(items, prefix+"[]", fields, false)
		if err != nil {
			return false, err
		}
		found = found || itemFound
	}
	return found, nil
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
	fields := map[string]bool{}
	pathFields, err := declaredWritePathFields(action)
	if err != nil {
		return nil, err
	}
	for _, name := range pathFields {
		fields["path."+name] = true
	}
	for name := range action.Query {
		name = strings.TrimSpace(name)
		if name == "" {
			return nil, fmt.Errorf("query contains an empty parameter name")
		}
		fields["query."+name] = true
	}

	switch bodyTypeOf(action) {
	case "none":
		addWriteBodyFields(fields, action.BodyFields)
	case "graphql":
		if action.GraphQL != nil {
			collectRequestValueFields(action.GraphQL.Variables, "body", fields)
		}
	case "json_array":
		if name := strings.TrimSpace(action.BodyField); name != "" {
			fields["body."+name] = true
		}
	case "multipart":
		if action.Multipart != nil {
			for _, part := range action.Multipart.Parts {
				fields["body."+strings.TrimSpace(part.Name)] = true
			}
		}
	case "base64_upload":
		if err := addWriteJSONBodyFields(fields, action); err != nil {
			return nil, err
		}
		if action.Base64Upload != nil {
			delete(fields, "body."+action.Base64Upload.SourceField)
			fields["body."+action.Base64Upload.ContentField] = true
		}
	default:
		if len(action.BodyFields) > 0 {
			addWriteBodyFields(fields, action.BodyFields)
		} else if err := addWriteJSONBodyFields(fields, action); err != nil {
			return nil, err
		}
	}
	if action.DynamicFields != nil {
		fields["body."+strings.TrimSpace(action.DynamicFields.Field)] = true
	}

	paths := make([]string, 0, len(fields))
	for path := range fields {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	return paths, nil
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

func addWriteBodyFields(fields map[string]bool, names []string) {
	for _, name := range names {
		name = strings.TrimSpace(name)
		if name != "" {
			fields["body."+name] = true
		}
	}
}

func addWriteJSONBodyFields(fields map[string]bool, action WriteAction) error {
	if len(action.BodyFields) > 0 {
		addWriteBodyFields(fields, action.BodyFields)
		return nil
	}
	bodyFields := map[string]bool{}
	if err := collectRequestSchemaFields(action.RecordSchema, "body", bodyFields); err != nil {
		return fmt.Errorf("record_schema: %w", err)
	}
	excluded := map[string]bool{}
	for _, field := range action.PathFields {
		if !strings.Contains(field, ".") {
			excluded["body."+field] = true
		}
	}
	for path := range bodyFields {
		excludedPath := false
		for prefix := range excluded {
			if path == prefix || strings.HasPrefix(path, prefix+".") || strings.HasPrefix(path, prefix+"[]") {
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

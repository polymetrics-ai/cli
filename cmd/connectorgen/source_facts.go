package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"golang.org/x/net/html"
	"gopkg.in/yaml.v3"
	"io"
	"sort"
	"strconv"
	"strings"
)

type sourceFactRef struct {
	Section     string `json:"section,omitempty"`
	Part        string `json:"part,omitempty"`
	DocumentID  string `json:"document_id"`
	Pointer     string `json:"pointer"`
	ValueSHA256 string `json:"value_sha256"`
}

type sourceParameterFact struct {
	In       string          `json:"in"`
	Name     string          `json:"name"`
	Required bool            `json:"required"`
	Node     json.RawMessage `json:"node"`
	Ref      sourceFactRef   `json:"ref"`
}

type sourceFacts struct {
	referenceRoot any                        `json:"-"`
	bindings      *sourceLaneBindingInputs   `json:"-"`
	analysis      *sourceShapeAnalysis       `json:"-"`
	Document      json.RawMessage            `json:"-"`
	RefPrefix     string                     `json:"-"`
	Parameters    []sourceParameterFact      `json:"effective_parameters"`
	Status        string                     `json:"status"`
	Method        string                     `json:"method"`
	Path          string                     `json:"path"`
	Protocol      string                     `json:"protocol"`
	OperationID   string                     `json:"operation_id"`
	Groups        map[string]json.RawMessage `json:"groups"`
	Refs          map[string]sourceFactRef   `json:"refs"`
	Diagnostics   []string                   `json:"diagnostics"`
}

// normalizeSourceFacts copies provider groups and records citations into the
// pinned document. Executable definitions never supply missing source facts.
func normalizeSourceFacts(row retainedSourceOperation, doc retainedSourceDocument, rawDoc *retainedSourceDocument) sourceFacts {
	facts := sourceFacts{Parameters: []sourceParameterFact{}, Status: "unavailable", Groups: map[string]json.RawMessage{}, Refs: map[string]sourceFactRef{}, Diagnostics: []string{}, Document: doc.Payload, RefPrefix: "/source_contract"}
	if !row.Observed {
		facts.Diagnostics = append(facts.Diagnostics, "source_unavailable")
		return facts
	}
	var node map[string]json.RawMessage
	if err := decodeSourceJSON(row.Node, &node); err != nil {
		facts.Diagnostics = append(facts.Diagnostics, "source_node_invalid")
		return facts
	}
	_ = json.Unmarshal(node["method"], &facts.Method)
	facts.Method = strings.ToUpper(facts.Method)
	_ = json.Unmarshal(node["path"], &facts.Path)
	_ = json.Unmarshal(node["protocol"], &facts.Protocol)
	_ = json.Unmarshal(node["operation_id"], &facts.OperationID)
	group := func(name string, value json.RawMessage, documentID, pointer string) {
		if len(value) == 0 {
			facts.Groups[name] = json.RawMessage("null")
			return
		}
		value = append(json.RawMessage(nil), value...)
		facts.Groups[name] = value
		canonical, err := canonicalSourceJSON(value)
		if err != nil {
			facts.Diagnostics = append(facts.Diagnostics, "source_fact_invalid:"+name)
			return
		}
		facts.Refs[name] = sourceFactRef{DocumentID: documentID, Pointer: pointer, ValueSHA256: sourceBytesHash(canonical)}
	}
	for _, name := range []string{"method", "path", "protocol", "operation_id"} {
		group(name, node[name], doc.ID, row.Pointer+"/"+name)
	}
	view, err := sourceDocumentViewFor(doc)
	if err != nil {
		facts.Diagnostics = append(facts.Diagnostics, "source_document_invalid")
		return facts
	}
	root := view.Root
	facts.referenceRoot = view.ReferenceRoot
	var operation map[string]json.RawMessage
	operationPointer := row.Pointer + "/source_operation"
	operationDocument := doc.ID
	if len(node["source_operation"]) > 0 {
		if err := decodeSourceJSON(node["source_operation"], &operation); err != nil {
			facts.Diagnostics = append(facts.Diagnostics, "source_operation_invalid")
			return facts
		}
	} else if rawDoc != nil && rawDoc.ContentType == "text/html" {
		return normalizeRenderedSourceFacts(row, *rawDoc, facts)
	} else if rawDoc != nil {
		facts.Document = rawDoc.Payload
		facts.RefPrefix = ""
		rawView, err := sourceDocumentViewFor(*rawDoc)
		if err != nil {
			facts.Diagnostics = append(facts.Diagnostics, "raw_document_invalid")
			return facts
		}
		root = rawView.Root
		facts.referenceRoot = rawView.ReferenceRoot
		operationPointer = "/paths/" + escapeSourcePointer(facts.Path) + "/" + strings.ToLower(facts.Method)
		raw, err := sourceDocumentPointer(*rawDoc, operationPointer)
		if err != nil {
			facts.Diagnostics = append(facts.Diagnostics, "raw_operation_missing")
			return facts
		}
		if err := decodeSourceJSON(raw, &operation); err != nil {
			facts.Diagnostics = append(facts.Diagnostics, "raw_operation_invalid")
			return facts
		}
		operationDocument = rawDoc.ID
		params, err := sourceDocumentPointer(*rawDoc, "/paths/"+escapeSourcePointer(facts.Path)+"/parameters")
		if err == nil {
			group("path_parameters", params, rawDoc.ID, "/paths/"+escapeSourcePointer(facts.Path)+"/parameters")
		}
	} else {
		facts.Diagnostics = append(facts.Diagnostics, "source_operation_not_retained")
		group("source_node", row.Node, doc.ID, row.Pointer)
		return facts
	}
	opBytes, err := json.Marshal(operation)
	if err != nil {
		facts.Diagnostics = append(facts.Diagnostics, "source_operation_invalid")
		return facts
	}
	group("source_operation", opBytes, operationDocument, operationPointer)
	names := map[string]string{"parameters": "parameters", "request_body": "requestBody", "responses": "responses", "summary": "summary", "description": "description", "callbacks": "callbacks", "deprecated": "deprecated", "external_docs": "externalDocs"}
	for name, field := range names {
		group(name, operation[field], operationDocument, operationPointer+"/"+escapeSourcePointer(field))
	}
	contract := view.Contract
	if rawDoc != nil {
		contract = root
	}
	facts.Diagnostics = append(facts.Diagnostics, view.Diagnostics...)
	contractPointer := "/source_contract"
	contractDocument := doc.ID
	if rawDoc != nil {
		contractPointer = ""
		contractDocument = rawDoc.ID
	}
	// Shared provider contracts live once in manifest documents, not in every row.
	if security, exists := operation["security"]; exists {
		group("security", security, operationDocument, operationPointer+"/security")
	} else {
		group("security", contract["security"], contractDocument, contractPointer+"/security")
	}
	var components map[string]json.RawMessage
	if len(contract["components"]) > 0 {
		if err := decodeSourceJSON(contract["components"], &components); err != nil {
			facts.Diagnostics = append(facts.Diagnostics, "source_components_invalid")
		}
	}
	group("security_schemes", components["securitySchemes"], contractDocument, contractPointer+"/components/securitySchemes")
	// Schema references resolve against the retained shared document.
	group("webhooks", contract["webhooks"], contractDocument, contractPointer+"/webhooks")
	for _, name := range []string{"path_bridge", "event_schema_inventory", "batch_action_inventory"} {
		group(name, view.Rest[name], doc.ID, "/rest/"+name)
	}
	facts.Status = "available"
	if facts.Method == "" || facts.Path == "" || facts.Protocol == "" {
		facts.Status = "partial"
		facts.Diagnostics = append(facts.Diagnostics, "source_identity_facts_missing")
	}
	if len(operation["responses"]) == 0 {
		facts.Status = "partial"
		facts.Diagnostics = append(facts.Diagnostics, "source_responses_missing")
	}
	if _, exists := facts.Groups["path_parameters"]; !exists {
		facts.Groups["path_parameters"] = json.RawMessage("null")
	}
	facts.Parameters, facts.Diagnostics = effectiveSourceParameters(facts, facts.Diagnostics)
	return facts
}

func canonicalSourceJSON(raw []byte) ([]byte, error) {
	var value any
	if err := decodeSourceJSON(raw, &value); err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func escapeSourcePointer(s string) string {
	return strings.ReplaceAll(strings.ReplaceAll(s, "~", "~0"), "/", "~1")
}

func sourceJSONPointer(document []byte, pointer string) (json.RawMessage, error) {
	if pointer == "" {
		return append(json.RawMessage(nil), document...), nil
	}
	if !strings.HasPrefix(pointer, "/") {
		return nil, fmt.Errorf("invalid local source pointer")
	}
	current := json.RawMessage(document)
	for _, encoded := range strings.Split(pointer[1:], "/") {
		part := strings.ReplaceAll(strings.ReplaceAll(encoded, "~1", "/"), "~0", "~")
		trimmed := bytes.TrimSpace(current)
		if len(trimmed) == 0 {
			return nil, fmt.Errorf("source pointer absent")
		}
		switch trimmed[0] {
		case '{':
			var obj map[string]json.RawMessage
			if err := json.Unmarshal(current, &obj); err != nil {
				return nil, err
			}
			value, ok := obj[part]
			if !ok {
				return nil, fmt.Errorf("source pointer absent")
			}
			current = value
		case '[':
			var array []json.RawMessage
			if err := json.Unmarshal(current, &array); err != nil {
				return nil, err
			}
			n, err := strconv.Atoi(part)
			if err != nil || n < 0 || n >= len(array) || strconv.Itoa(n) != part {
				return nil, fmt.Errorf("invalid source array pointer")
			}
			current = array[n]
		default:
			return nil, fmt.Errorf("source pointer traverses scalar")
		}
	}
	return current, nil
}

// sourceYAMLDocument rejects duplicate keys and additional YAML documents. A
// node walk preserves numeric lexemes and bounds aliases/depth before JSON.
func sourceYAMLDocument(raw []byte) (json.RawMessage, error) {
	decoder := yaml.NewDecoder(bytes.NewReader(raw))
	var node yaml.Node
	if err := decoder.Decode(&node); err != nil {
		return nil, err
	}
	var trailing yaml.Node
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return nil, fmt.Errorf("trailing source YAML document")
		}
		return nil, err
	}
	count := 0
	value, err := sourceYAMLValue(&node, 0, &count, map[*yaml.Node]bool{})
	if err != nil {
		return nil, err
	}
	return json.Marshal(value)
}

func sourceYAMLValue(node *yaml.Node, depth int, count *int, active map[*yaml.Node]bool) (any, error) {
	*count++
	if depth > 256 || *count > 1000000 {
		return nil, fmt.Errorf("source YAML limit exceeded")
	}
	if active[node] {
		return nil, fmt.Errorf("cyclic source YAML alias")
	}
	active[node] = true
	defer delete(active, node)
	switch node.Kind {
	case yaml.DocumentNode:
		if len(node.Content) != 1 {
			return nil, fmt.Errorf("invalid source YAML document")
		}
		return sourceYAMLValue(node.Content[0], depth+1, count, active)
	case yaml.AliasNode:
		return sourceYAMLValue(node.Alias, depth+1, count, active)
	case yaml.MappingNode:
		result := map[string]any{}
		for i := 0; i < len(node.Content); i += 2 {
			key := node.Content[i]
			if key.Kind != yaml.ScalarNode {
				return nil, fmt.Errorf("non-scalar source YAML key")
			}
			if _, ok := result[key.Value]; ok {
				return nil, fmt.Errorf("duplicate source YAML key %q", key.Value)
			}
			value, err := sourceYAMLValue(node.Content[i+1], depth+1, count, active)
			if err != nil {
				return nil, err
			}
			result[key.Value] = value
		}
		return result, nil
	case yaml.SequenceNode:
		result := []any{}
		for _, child := range node.Content {
			value, err := sourceYAMLValue(child, depth+1, count, active)
			if err != nil {
				return nil, err
			}
			result = append(result, value)
		}
		return result, nil
	case yaml.ScalarNode:
		switch node.Tag {
		case "!!null":
			return nil, nil
		case "!!bool":
			return strconv.ParseBool(node.Value)
		case "!!int", "!!float":
			number := json.Number(node.Value)
			if _, err := json.Marshal(number); err != nil {
				return nil, fmt.Errorf("unsupported non-JSON YAML number %q", node.Value)
			}
			return number, nil
		default:
			return node.Value, nil
		}
	}
	return nil, fmt.Errorf("unsupported source YAML node")
}

func effectiveSourceParameters(facts sourceFacts, diagnostics []string) ([]sourceParameterFact, []string) {
	facts.analysis = &sourceShapeAnalysis{Root: facts.referenceRoot, Objects: map[string]map[string]json.RawMessage{}, Shapes: map[string]sourceShape{}}
	byKey := map[string]sourceParameterFact{}
	for _, group := range []string{"path_parameters", "parameters"} {
		raw := facts.Groups[group]
		if len(raw) == 0 || string(raw) == "null" {
			continue
		}
		var parameters []json.RawMessage
		if err := json.Unmarshal(raw, &parameters); err != nil {
			diagnostics = append(diagnostics, "source_parameters_invalid")
			continue
		}
		for i, parameter := range parameters {
			node, ok := sourceResolveObject(facts, parameter, map[string]bool{}, 0)
			if !ok {
				diagnostics = append(diagnostics, "source_parameter_unresolved")
				continue
			}
			var name, in string
			var required bool
			if err := json.Unmarshal(node["name"], &name); err != nil {
				diagnostics = append(diagnostics, "source_parameter_name_invalid")
				continue
			}
			if err := json.Unmarshal(node["in"], &in); err != nil {
				diagnostics = append(diagnostics, "source_parameter_location_invalid")
				continue
			}
			if len(node["required"]) > 0 {
				if err := json.Unmarshal(node["required"], &required); err != nil {
					diagnostics = append(diagnostics, "source_parameter_required_invalid")
					continue
				}
			}
			ref := facts.Refs[group]
			ref.Pointer += fmt.Sprintf("/%d", i)
			canonical, err := canonicalSourceJSON(parameter)
			if err != nil {
				diagnostics = append(diagnostics, "source_parameter_invalid")
				continue
			}
			ref.ValueSHA256 = sourceBytesHash(canonical)
			encoded, err := json.Marshal(node)
			if err != nil {
				diagnostics = append(diagnostics, "source_parameter_invalid")
				continue
			}
			byKey[in+":"+name] = sourceParameterFact{In: in, Name: name, Required: required, Node: encoded, Ref: ref}
		}
	}
	keys := make([]string, 0, len(byKey))
	for key := range byKey {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	result := make([]sourceParameterFact, 0, len(keys))
	for _, key := range keys {
		result = append(result, byKey[key])
	}
	return result, diagnostics
}

func normalizeRenderedSourceFacts(row retainedSourceOperation, document retainedSourceDocument, facts sourceFacts) sourceFacts {
	markup, err := sourceRenderedMarkup(document)
	if err != nil {
		facts.Diagnostics = append(facts.Diagnostics, "rendered_document_invalid")
		return facts
	}
	heading, text, err := sourceRenderedSection(markup, row.SourceLocation)
	if err != nil {
		facts.Diagnostics = append(facts.Diagnostics, "rendered_section_unavailable")
		return facts
	}
	facts.Document = document.Payload
	facts.RefPrefix = ""
	for _, group := range []struct{ name, part, value string }{{"summary", "heading", heading}, {"rendered_reference", "text", text}} {
		raw, err := json.Marshal(group.value)
		if err != nil {
			facts.Diagnostics = append(facts.Diagnostics, "rendered_value_invalid")
			return facts
		}
		facts.Groups[group.name] = raw
		facts.Refs[group.name] = sourceFactRef{DocumentID: document.ID, Section: row.SourceLocation, Part: group.part, ValueSHA256: sourceBytesHash(raw)}
	}
	facts.Status = "partial"
	facts.Diagnostics = append(facts.Diagnostics, "rendered_parameter_contract_unmapped", "rendered_request_response_contract_unmapped")
	return facts
}

func sourceRenderedMarkup(document retainedSourceDocument) ([]byte, error) {
	if document.ContentType != "text/html" {
		return nil, fmt.Errorf("not a rendered HTML document")
	}
	var markup string
	if err := decodeSourceJSON(document.Payload, &markup); err != nil {
		return nil, err
	}
	raw := []byte(markup)
	if len(raw) > 64<<20 || int64(len(raw)) != document.Bytes || sourceBytesHash(raw) != document.RetainedFileSHA256 {
		return nil, fmt.Errorf("rendered bytes do not match retained identity")
	}
	return raw, nil
}

// sourceRenderedSection extracts displayed text from one retained heading
// section. It never follows links or executes HTML. Inline text remains
// contiguous; block elements delimit words, and scripts/styles are excluded.
func sourceRenderedSection(markup []byte, section string) (string, string, error) {
	if len(markup) > 64<<20 || !strings.HasPrefix(section, "#") || len(section) < 2 || !validSourceID(section) {
		return "", "", fmt.Errorf("invalid rendered section")
	}
	anchor := strings.TrimPrefix(section, "#")
	tokenizer := html.NewTokenizer(bytes.NewReader(markup))
	var heading, text strings.Builder
	found, active, inHeading := false, false, false
	level := 0
	skip := ""
	headingLevel := func(name string) int {
		if len(name) == 2 && name[0] == 'h' && name[1] >= '1' && name[1] <= '6' {
			return int(name[1] - '0')
		}
		return 0
	}
	block := func(name string) bool {
		switch name {
		case "p", "div", "pre", "li", "br", "table", "tr", "td", "th", "ul", "ol", "blockquote":
			return true
		}
		return headingLevel(name) > 0
	}
	for count := 0; ; count++ {
		if count >= 1000000 {
			return "", "", fmt.Errorf("rendered token budget exceeded")
		}
		kind := tokenizer.Next()
		if kind == html.ErrorToken {
			if err := tokenizer.Err(); err != io.EOF {
				return "", "", err
			}
			break
		}
		token := tokenizer.Token()
		if skip != "" {
			if kind == html.EndTagToken && token.Data == skip {
				skip = ""
			}
			continue
		}
		if (kind == html.StartTagToken || kind == html.SelfClosingTagToken) && (token.Data == "script" || token.Data == "style") {
			if kind != html.SelfClosingTagToken {
				skip = token.Data
			}
			continue
		}
		if kind == html.StartTagToken {
			currentLevel := headingLevel(token.Data)
			if currentLevel > 0 {
				id := ""
				for _, attribute := range token.Attr {
					if attribute.Key == "id" {
						id = attribute.Val
					}
				}
				if id == anchor {
					if found {
						return "", "", fmt.Errorf("duplicate rendered section")
					}
					found, active, inHeading = true, true, true
					level = currentLevel
				} else if active && currentLevel <= level {
					active = false
				}
			}
		}
		if !active {
			continue
		}
		switch kind {
		case html.TextToken:
			text.WriteString(token.Data)
			if inHeading {
				heading.WriteString(token.Data)
			}
		case html.StartTagToken, html.EndTagToken, html.SelfClosingTagToken:
			if block(token.Data) {
				text.WriteByte(' ')
			}
			if kind == html.EndTagToken && headingLevel(token.Data) == level {
				inHeading = false
			}
		}
	}
	title := strings.Join(strings.Fields(heading.String()), " ")
	if !found || inHeading || title == "" {
		return "", "", fmt.Errorf("rendered section absent or incomplete")
	}
	return title, strings.Join(strings.Fields(text.String()), " "), nil
}

// resolveSourceFactValue re-reads a citation from its immutable document.
// Rendered selectors are checked against the same bounded section parser;
// ordinary source facts continue to use exact local JSON pointers.
func resolveSourceFactValue(document retainedSourceDocument, ref sourceFactRef) (json.RawMessage, error) {
	if document.ID != ref.DocumentID {
		return nil, fmt.Errorf("source document identity mismatch")
	}
	if ref.Section == "" {
		if ref.Part != "" {
			return nil, fmt.Errorf("part without rendered section")
		}
		return sourceJSONPointer(document.Payload, ref.Pointer)
	}
	if ref.Pointer != "" {
		return nil, fmt.Errorf("rendered citation cannot also select JSON")
	}
	markup, err := sourceRenderedMarkup(document)
	if err != nil {
		return nil, err
	}
	heading, text, err := sourceRenderedSection(markup, ref.Section)
	if err != nil {
		return nil, err
	}
	switch ref.Part {
	case "heading":
		return json.Marshal(heading)
	case "text":
		return json.Marshal(text)
	}
	return nil, fmt.Errorf("unknown rendered citation part")
}

// sourceDocumentView belongs to one immutable retained document. It avoids
// reparsing the provider's complete root and shared references for each row.
type sourceDocumentView struct {
	PayloadSHA256 string
	Root          map[string]json.RawMessage
	Contract      map[string]json.RawMessage
	Rest          map[string]json.RawMessage
	ReferenceRoot any
	Diagnostics   []string
}

func prepareSourceDocument(document retainedSourceDocument) (retainedSourceDocument, error) {
	if document.ContentType == "text/html" {
		return document, nil
	}
	view, err := sourceDocumentViewFor(document)
	if err != nil {
		return document, err
	}
	document.view = view
	return document, nil
}

func sourceDocumentViewFor(document retainedSourceDocument) (*sourceDocumentView, error) {
	digest := sourceBytesHash(document.Payload)
	if document.view != nil {
		if document.view.PayloadSHA256 != digest {
			return nil, fmt.Errorf("retained document changed after preparation")
		}
		return document.view, nil
	}
	view := &sourceDocumentView{PayloadSHA256: digest, Diagnostics: []string{}}
	if err := decodeSourceJSON(document.Payload, &view.Root); err != nil {
		return nil, err
	}
	if view.Root == nil {
		return nil, fmt.Errorf("source document is not an object")
	}
	if err := decodeSourceJSON(document.Payload, &view.ReferenceRoot); err != nil {
		return nil, err
	}
	if raw := view.Root["source_contract"]; len(raw) > 0 {
		if err := decodeSourceJSON(raw, &view.Contract); err != nil {
			view.Diagnostics = append(view.Diagnostics, "source_contract_invalid")
		}
	}
	if raw := view.Root["rest"]; len(raw) > 0 {
		if err := decodeSourceJSON(raw, &view.Rest); err != nil {
			view.Diagnostics = append(view.Diagnostics, "source_inventory_metadata_invalid")
		}
	}
	return view, nil
}

func sourceDocumentPointer(document retainedSourceDocument, pointer string) (json.RawMessage, error) {
	view, err := sourceDocumentViewFor(document)
	if err != nil {
		return nil, err
	}
	facts := sourceFacts{Document: document.Payload, analysis: &sourceShapeAnalysis{Root: view.ReferenceRoot}}
	return sourceAnalysisPointer(facts, pointer)
}

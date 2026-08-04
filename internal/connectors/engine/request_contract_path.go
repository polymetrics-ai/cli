package engine

import (
	"fmt"
	"strings"
)

func requestFieldPointer(namespace string, tokens ...string) string {
	parts := make([]string, 0, len(tokens)+1)
	parts = append(parts, namespace)
	for _, token := range tokens {
		parts = append(parts, escapeRequestFieldPointerToken(token))
	}
	return "/" + strings.Join(parts, "/")
}

func appendRequestFieldPointer(pointer, token string) string {
	return pointer + "/" + escapeRequestFieldPointerToken(token)
}

func escapeRequestFieldPointerToken(token string) string {
	token = strings.ReplaceAll(token, "~", "~0")
	return strings.ReplaceAll(token, "/", "~1")
}

func parseRequestFieldPointer(pointer string) (string, []string, error) {
	if pointer == "" || pointer != strings.TrimSpace(pointer) || strings.ContainsAny(pointer, "\r\n\t") || !strings.HasPrefix(pointer, "/") {
		return "", nil, fmt.Errorf("must be an absolute escaped JSON Pointer")
	}
	rawTokens := strings.Split(strings.TrimPrefix(pointer, "/"), "/")
	if len(rawTokens) < 2 {
		return "", nil, fmt.Errorf("must contain a namespace and field token")
	}
	namespace := rawTokens[0]
	switch namespace {
	case "body", "path", "query":
	default:
		return "", nil, fmt.Errorf("unsupported namespace %q", namespace)
	}
	tokens := make([]string, 0, len(rawTokens)-1)
	for _, rawToken := range rawTokens[1:] {
		token, err := unescapeRequestFieldPointerToken(rawToken)
		if err != nil {
			return "", nil, err
		}
		if token == "" {
			return "", nil, fmt.Errorf("contains an empty field token")
		}
		tokens = append(tokens, token)
	}
	return namespace, tokens, nil
}

func unescapeRequestFieldPointerToken(token string) (string, error) {
	var decoded strings.Builder
	for i := 0; i < len(token); i++ {
		if token[i] != '~' {
			decoded.WriteByte(token[i])
			continue
		}
		if i+1 >= len(token) || token[i+1] != '0' && token[i+1] != '1' {
			return "", fmt.Errorf("contains invalid JSON Pointer escape in %q", token)
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

func validRequestContractFieldPath(path string) bool {
	_, _, err := parseRequestFieldPointer(path)
	return err == nil
}

func requestContractFieldNamespace(path string) string {
	namespace, _, err := parseRequestFieldPointer(path)
	if err != nil {
		return ""
	}
	return namespace
}

func CanonicalRequestFieldPointer(mapping string, bodySchema *Schema) (string, error) {
	mapping = strings.TrimSpace(mapping)
	for _, namespace := range []string{"path", "query"} {
		if name, ok := strings.CutPrefix(mapping, namespace+"."); ok {
			if name == "" {
				return "", fmt.Errorf("%s mapping has an empty field name", namespace)
			}
			return requestFieldPointer(namespace, name), nil
		}
	}
	bodyPath, ok := strings.CutPrefix(mapping, "body.")
	if !ok || bodyPath == "" {
		return "", fmt.Errorf("unsupported request mapping %q", mapping)
	}
	if bodySchema == nil || bodySchema.node == nil {
		return "", fmt.Errorf("body mapping requires body_schema")
	}
	tokens, err := bodySchema.canonicalMappingTokens(bodyPath)
	if err != nil {
		return "", err
	}
	return requestFieldPointer("body", tokens...), nil
}

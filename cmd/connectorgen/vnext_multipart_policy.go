package main

import "encoding/json"

// canonicalMultipartFilenamePolicy removes only the explicit file-part default.
// Provider body schemas and other string-valued properties are never rewritten.
func canonicalMultipartFilenamePolicy(raw json.RawMessage, operation bool) json.RawMessage {
	var root map[string]json.RawMessage
	if json.Unmarshal(raw, &root) != nil {
		return cloneRawJSON(raw)
	}
	owner := root
	if operation {
		owner = nil
		if json.Unmarshal(root["rest"], &owner) != nil {
			return cloneRawJSON(raw)
		}
	}
	var spec map[string]json.RawMessage
	if json.Unmarshal(owner["multipart"], &spec) != nil {
		return cloneRawJSON(raw)
	}
	var parts []map[string]json.RawMessage
	if json.Unmarshal(spec["parts"], &parts) != nil {
		return cloneRawJSON(raw)
	}
	changed := false
	for _, part := range parts {
		var kind, encoding string
		if json.Unmarshal(part["type"], &kind) == nil && kind == "file" && json.Unmarshal(part["filename_encoding"], &encoding) == nil && encoding == "identity" {
			delete(part, "filename_encoding")
			changed = true
		}
	}
	if !changed {
		return cloneRawJSON(raw)
	}
	spec["parts"], _ = json.Marshal(parts)
	owner["multipart"], _ = json.Marshal(spec)
	if operation {
		root["rest"], _ = json.Marshal(owner)
	}
	out, _ := json.Marshal(root)
	return out
}

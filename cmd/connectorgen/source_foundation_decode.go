package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
)

// Capacity comes from the caller's separately reviewed assertion allowlist,
// never fabricated provider keys. Refuse extra records before decoding them.
func decodeSourceFoundationProofDocument(ctx context.Context, raw []byte, maxRecords int) (sourceFoundationProofDocument, error) {
	var document sourceFoundationProofDocument
	d := json.NewDecoder(bytes.NewReader(raw))
	if tok, err := d.Token(); err != nil || tok != json.Delim('{') || maxRecords < 0 {
		return document, fmt.Errorf("foundation proof envelope invalid")
	}
	seen := map[string]bool{}
	tuples := 0
	for d.More() {
		if err := ctx.Err(); err != nil {
			return document, err
		}
		tok, err := d.Token()
		name, ok := tok.(string)
		if err != nil || !ok || seen[name] {
			return document, fmt.Errorf("foundation proof field invalid")
		}
		seen[name] = true
		switch name {
		case "schema_version":
			err = d.Decode(&document.SchemaVersion)
		case "kind":
			err = d.Decode(&document.Kind)
		case "atlas":
			var value json.RawMessage
			if err = d.Decode(&value); err == nil {
				if _, err = sourceFoundationRequiredObject(value, "path", "sha256", "bytes"); err == nil {
					err = decodeSourceJSON(value, &document.Atlas)
				}
				if err == nil {
					err = decodeStrictJSON(value, &document.Atlas)
				}
			}
		case "records":
			if tok, e := d.Token(); e != nil || tok != json.Delim('[') {
				return document, fmt.Errorf("foundation proof records invalid")
			}
			document.Records = []sourceFoundationProofRecord{}
			for d.More() {
				if err := ctx.Err(); err != nil {
					return document, err
				}
				if len(document.Records) >= maxRecords {
					return document, fmt.Errorf("foundation proof record capacity exceeded")
				}
				var value json.RawMessage
				if err := d.Decode(&value); err != nil {
					return document, fmt.Errorf("foundation proof record encoding: %w", err)
				}
				if err := sourceFoundationProofInputBudget(ctx, value, &tuples); err != nil {
					return document, err
				}
				if err := sourceFoundationProofRequiredMembers(value); err != nil {
					return document, err
				}
				var record sourceFoundationProofRecord
				if decodeSourceJSON(value, &record) != nil || decodeStrictJSON(value, &record) != nil {
					return document, fmt.Errorf("foundation proof record invalid")
				}
				document.Records = append(document.Records, record)
			}
			if tok, e := d.Token(); e != nil || tok != json.Delim(']') {
				return document, fmt.Errorf("foundation proof records incomplete")
			}
		default:
			return document, fmt.Errorf("foundation proof unknown field")
		}
		if err != nil {
			return document, fmt.Errorf("foundation proof field encoding: %w", err)
		}
	}
	if tok, err := d.Token(); err != nil || tok != json.Delim('}') || d.Decode(new(any)) != io.EOF || len(seen) != 4 || document.SchemaVersion != 1 || document.Kind != "foundation_contract_proofs" {
		return document, fmt.Errorf("foundation proof envelope incomplete")
	}
	return document, nil
}

// All members of this closed dialect are required and non-null. Typed Go
// decoding alone cannot distinguish an omitted/null integer from a real zero.
func sourceFoundationRequiredObject(raw []byte, names ...string) (map[string]json.RawMessage, error) {
	var object map[string]json.RawMessage
	if err := json.Unmarshal(raw, &object); err != nil || len(object) != len(names) {
		return nil, fmt.Errorf("foundation required object invalid")
	}
	for _, name := range names {
		value, exists := object[name]
		if !exists || bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
			return nil, fmt.Errorf("foundation required member absent or null: %s", name)
		}
	}
	return object, nil
}

func sourceFoundationProofRequiredMembers(raw []byte) error {
	object, err := sourceFoundationRequiredObject(raw, "id", "atlas_id", "contract", "owner_symbols", "test", "assertion", "inputs", "capture", "output", "execution_class", "scope", "limitations")
	if err != nil {
		return err
	}
	for _, shape := range []struct {
		name string
		keys []string
	}{
		{"contract", []string{"pointer", "value_sha256"}},
		{"test", []string{"file", "symbol", "selected", "package"}},
		{"assertion", []string{"statement", "start_line", "end_line", "source_sha256"}},
		{"capture", []string{"path", "sha256", "bytes"}},
		{"output", []string{"path", "sha256", "bytes"}},
	} {
		if _, err := sourceFoundationRequiredObject(object[shape.name], shape.keys...); err != nil {
			return err
		}
	}
	for _, shape := range []struct {
		name string
		keys []string
	}{
		{"owner_symbols", []string{"file", "name"}},
		{"inputs", []string{"path", "sha256", "role", "bytes"}},
	} {
		var values []json.RawMessage
		if err := json.Unmarshal(object[shape.name], &values); err != nil {
			return fmt.Errorf("foundation required array invalid")
		}
		for _, value := range values {
			if _, err := sourceFoundationRequiredObject(value, shape.keys...); err != nil {
				return err
			}
		}
	}
	return nil
}

func sourceFoundationProofInputBudget(ctx context.Context, raw []byte, tuples *int) error {
	if _, err := countSourceJSONNodes(ctx, raw, 1<<20); err != nil {
		return fmt.Errorf("foundation proof record node capacity: %w", err)
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	if tok, err := d.Token(); err != nil || tok != json.Delim('{') {
		return fmt.Errorf("foundation proof record object invalid")
	}
	for d.More() {
		if err := ctx.Err(); err != nil {
			return err
		}
		tok, err := d.Token()
		if err != nil {
			return err
		}
		if tok == "inputs" {
			if tok, err := d.Token(); err != nil || tok != json.Delim('[') {
				return fmt.Errorf("foundation proof inputs invalid")
			}
			count := 0
			for d.More() {
				if count >= 4096 || *tuples >= 131072 {
					return fmt.Errorf("foundation proof input capacity exceeded")
				}
				count++
				*tuples++
				var input json.RawMessage
				if err := d.Decode(&input); err != nil {
					return err
				}
			}
			if _, err := d.Token(); err != nil {
				return err
			}
		} else {
			var field json.RawMessage
			if err := d.Decode(&field); err != nil {
				return err
			}
		}
	}
	return nil
}

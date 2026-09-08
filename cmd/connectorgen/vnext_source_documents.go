package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"
)

func validateSourceProjectionDocuments(p *vNextSourceProjection) error {
	if len(p.Documents) > 1024 {
		return fmt.Errorf("source_projection documentary count limit exceeded")
	}
	ids, paths := map[string]bool{}, map[string]bool{}
	var total int64
	for _, pin := range p.Inventories {
		ids[pin.ID] = true
		paths[pin.Path] = true
		total += pin.Bytes
	}
	for i, pin := range p.Documents {
		if !sourceLaneIdentityPart(pin.ID) || ids[pin.ID] || paths[pin.Path] || !sourceLaneRelativePath(pin.Path) || !strings.HasPrefix(pin.Path, "sources/") || !sourceLaneDigest(pin.SHA256) || pin.Bytes <= 0 || pin.Bytes > 64<<20 {
			return fmt.Errorf("source_projection document %d has invalid or duplicate identity/pin", i)
		}
		ids[pin.ID], paths[pin.Path] = true, true
		switch pin.Format {
		case "json", "yaml", "html", "text":
		default:
			return fmt.Errorf("source_projection document %d has unsupported format", i)
		}
		origin, err := url.Parse(pin.SourceURL)
		if err != nil || origin.Host == "" || (origin.Scheme != "https" && origin.Scheme != "http") || origin.User != nil || origin.Fragment != "" || len(pin.SourceURL) > 8192 || strings.TrimSpace(pin.Revision) == "" || len(pin.Revision) > 4096 {
			return fmt.Errorf("source_projection document %d has invalid provenance", i)
		}
		if _, err = time.Parse(time.RFC3339, pin.RetrievedAt); err != nil {
			return fmt.Errorf("source_projection document %d has invalid retrieval time", i)
		}
		total += pin.Bytes
		if total > 512<<20 {
			return fmt.Errorf("source_projection retained byte budget exceeded")
		}
	}
	return nil
}

func (inputs *vNextSourceProjectionInputs) loadDocuments(lock vNextSourceLock, inventory *retainedSourceInventory) error {
	ids := map[string]bool{}
	for _, doc := range inventory.Documents {
		ids[doc.ID] = true
	}
	for _, pin := range lock.SourceProjection.Documents {
		if ids[pin.ID] {
			return fmt.Errorf("source_projection documentary identity collides with inventory document")
		}
		ids[pin.ID] = true
		raw, err := inputs.read(pin.Path, pin.Bytes)
		if err != nil {
			return err
		}
		if int64(len(raw)) != pin.Bytes || sourceBytesHash(raw) != pin.SHA256 {
			return fmt.Errorf("source_projection document %q pin mismatch", pin.ID)
		}
		contentType := "text/plain"
		payload := json.RawMessage(raw)
		switch pin.Format {
		case "json":
			contentType = "application/json"
			if _, err = countSourceJSONNodes(inputs.ctx, raw, 1_000_000); err != nil {
				return err
			}
		case "html":
			contentType = "text/html"
		case "yaml":
			contentType = "application/yaml"
		}
		if pin.Format != "json" {
			payload, err = json.Marshal(string(raw))
			if err != nil {
				return err
			}
		}
		inventory.Documents = append(inventory.Documents, retainedSourceDocument{ID: pin.ID, Path: pin.Path, Bytes: pin.Bytes, RetainedFileSHA256: pin.SHA256, ContentType: contentType, Payload: payload, documentaryBytes: raw})
	}
	return nil
}

func validateSourceProjectionCitation(doc retainedSourceDocument, ref sourceFactRef) error {
	if ref.Span != nil {
		span := ref.Span
		if ref.Pointer != "" || ref.Section != "" || ref.Part != "" || ref.ValueSHA256 != "" || span.Offset < 0 || span.Length <= 0 || span.Length > 64<<20 || !sourceLaneDigest(span.SHA256) {
			return fmt.Errorf("source_projection byte citation has conflicting selector or invalid bounds")
		}
		raw := doc.documentaryBytes
		if span.Offset > int64(len(raw)) || span.Length > int64(len(raw))-span.Offset {
			return fmt.Errorf("source_projection byte citation is outside retained document")
		}
		if sourceBytesHash(raw[span.Offset:span.Offset+span.Length]) != span.SHA256 {
			return fmt.Errorf("source_projection byte citation hash mismatch")
		}
		return nil
	}
	if !sourceLaneDigest(ref.ValueSHA256) {
		return fmt.Errorf("source_projection citation requires a value hash")
	}
	value, err := resolveSourceFactValue(doc, ref)
	if err != nil {
		return err
	}
	canonical, err := canonicalSourceJSON(value)
	if err != nil {
		return err
	}
	if sourceBytesHash(canonical) != ref.ValueSHA256 {
		return fmt.Errorf("source_projection citation value hash mismatch")
	}
	return nil
}

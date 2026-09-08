package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

// vNextSourceProjection is authoring input, never a runtime execution reader.
// Its derived output must enter the existing canonical graph.
type vNextSourceProjection struct {
	Version     int                              `json:"version"`
	Inventories []vNextSourceProjectionInventory `json:"inventories"`
	Semantics   []vNextSourceProjectionSemantic  `json:"semantics,omitempty"`
}

type vNextSourceProjectionInventory struct {
	ID     string `json:"id"`
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
	Bytes  int64  `json:"bytes"`
}

type vNextSourceProjectionKey struct {
	Inventory string `json:"inventory"`
	ID        string `json:"id"`
}

type vNextSourceProjectionSemantic struct {
	Source     vNextSourceProjectionKey         `json:"source"`
	Effect     string                           `json:"effect,omitempty"`
	Collection *vNextSourceProjectionCollection `json:"collection,omitempty"`
	Write      *vNextSourceProjectionWrite      `json:"write,omitempty"`
	Evidence   []sourceFactRef                  `json:"evidence,omitempty"`
}

type vNextSourceProjectionWrite struct {
	RowDelivery string `json:"row_delivery"`
	Batchable   *bool  `json:"batchable"`
	Risk        string `json:"risk"`
	Retry       string `json:"retry"`
}

type vNextSourceProjectionCollection struct {
	Records    vNextSourceCoordinate `json:"records"`
	PrimaryKey []string              `json:"primary_key"`
}

type vNextSourceCoordinate struct {
	Parameter      *vNextSourceParameterCoordinate      `json:"parameter,omitempty"`
	Request        *vNextSourceRequestCoordinate        `json:"request,omitempty"`
	Response       *vNextSourceResponseCoordinate       `json:"response,omitempty"`
	ResponseHeader *vNextSourceResponseHeaderCoordinate `json:"response_header,omitempty"`
}

type vNextSourceParameterCoordinate struct {
	In   string `json:"in"`
	Name string `json:"name"`
}

type vNextSourceRequestCoordinate struct {
	Media   string  `json:"media"`
	Pointer *string `json:"pointer"`
}

type vNextSourceResponseCoordinate struct {
	Status  string  `json:"status"`
	Media   string  `json:"media"`
	Pointer *string `json:"pointer"`
}

type vNextSourceResponseHeaderCoordinate struct {
	Status string `json:"status"`
	Name   string `json:"name"`
}

func validateVNextSourceProjection(lock vNextSourceLock, members map[string]json.RawMessage) error {
	p := lock.SourceProjection
	if lock.SchemaVersion != vNextSourceLockSchemaVersion || p == nil || p.Version != 1 {
		return fmt.Errorf("source_projection requires schema_version 4 and projection version 1")
	}
	for _, field := range []string{"operations", "schemas", "execution", "cli"} {
		if _, present := members[field]; present {
			return fmt.Errorf("source_projection cannot be combined with authored %s", field)
		}
	}
	if len(p.Inventories) == 0 || len(p.Inventories) > 1024 {
		return fmt.Errorf("source_projection requires 1..1024 pinned inventories")
	}
	ids, paths := map[string]bool{}, map[string]bool{}
	var total int64
	for index, pin := range p.Inventories {
		if !sourceLaneIdentityPart(pin.ID) || ids[pin.ID] || paths[pin.Path] ||
			!sourceLaneRelativePath(pin.Path) || !strings.HasPrefix(pin.Path, "sources/") ||
			!sourceLaneDigest(pin.SHA256) || pin.Bytes <= 0 || pin.Bytes > 64<<20 {
			return fmt.Errorf("source_projection inventory %d has invalid or duplicate identity/pin", index)
		}
		ids[pin.ID], paths[pin.Path] = true, true
		total += pin.Bytes
		if total > 512<<20 {
			return fmt.Errorf("source_projection retained byte budget exceeded")
		}
	}
	if len(p.Semantics) > 100000 {
		return fmt.Errorf("source_projection semantic budget exceeded")
	}
	seen := map[vNextSourceProjectionKey]bool{}
	for index, semantic := range p.Semantics {
		if !ids[semantic.Source.Inventory] || !validSourceID(semantic.Source.ID) || seen[semantic.Source] {
			return fmt.Errorf("source_projection semantic %d has invalid or duplicate source key", index)
		}
		seen[semantic.Source] = true
		if semantic.Effect != "" && semantic.Effect != "read" && semantic.Effect != "mutation" {
			return fmt.Errorf("source_projection semantic %d has invalid effect", index)
		}
		if write := semantic.Write; write != nil {
			if semantic.Effect != "mutation" || semantic.Collection != nil || write.RowDelivery != "one_request" || write.Batchable == nil || write.Retry != "single_attempt" {
				return fmt.Errorf("source_projection semantic %d has incomplete or unsupported write semantics", index)
			}
			switch write.Risk {
			case "low", "medium", "high", "critical":
			default:
				return fmt.Errorf("source_projection semantic %d has invalid write risk", index)
			}
		}
		if c := semantic.Collection; c != nil {
			r := c.Records.Response
			if r == nil || c.Records.Parameter != nil || c.Records.Request != nil || c.Records.ResponseHeader != nil {
				return fmt.Errorf("source_projection semantic %d collection requires one response coordinate", index)
			}
			_, knownStatus := sourceResponseStatus(r.Status)
			if !knownStatus || r.Media == "" || r.Pointer == nil || !sourceLaneProjectionPointer(*r.Pointer) {
				return fmt.Errorf("source_projection semantic %d collection has invalid response coordinate", index)
			}
			if len(c.PrimaryKey) == 0 || len(c.PrimaryKey) > 256 {
				return fmt.Errorf("source_projection semantic %d collection requires bounded primary key coordinates", index)
			}
			keys := map[string]bool{}
			for _, key := range c.PrimaryKey {
				if key == "" || !sourceLaneProjectionPointer(key) || keys[key] {
					return fmt.Errorf("source_projection semantic %d collection has invalid or duplicate primary key", index)
				}
				keys[key] = true
			}
		}
	}
	return nil
}

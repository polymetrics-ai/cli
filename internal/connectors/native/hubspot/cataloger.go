package hubspot

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/connectors/discovery"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	schemasPath    = "/crm-object-schemas/v3/schemas"
	propertiesPath = "/crm/v3/properties/{objectType}"
	objectsPath    = "/crm/v3/objects/{objectType}"
)

type hubspotProvider struct {
	runtime *engine.Runtime
}

type schemaListResponse struct {
	Results []hubspotObjectSchema `json:"results"`
}

type hubspotObjectSchema struct {
	ObjectTypeID       string `json:"objectTypeId"`
	FullyQualifiedName string `json:"fullyQualifiedName"`
	Labels             struct {
		Singular string `json:"singular"`
		Plural   string `json:"plural"`
	} `json:"labels"`
}

type propertiesResponse struct {
	Results []hubspotProperty `json:"results"`
}

// hubspotProperty is the documented response shape needed to make a JSON
// schema field. Every field in a discovered schema originates in one of these
// provider descriptions; no fixed field list is merged in.
type hubspotProperty struct {
	Name                 string `json:"name"`
	Label                string `json:"label"`
	Description          string `json:"description"`
	Type                 string `json:"type"`
	HasUniqueValue       bool   `json:"hasUniqueValue"`
	ReferencedObjectType string `json:"referencedObjectType"`
	Options              []struct {
		Value string `json:"value"`
	} `json:"options"`
}

func (p hubspotProvider) List(ctx context.Context) ([]discovery.Object, error) {
	requester, err := p.requester(httpMethodGet, schemasPath)
	if err != nil {
		return nil, err
	}
	var response schemaListResponse
	if err := requester.DoJSON(ctx, httpMethodGet, schemasPath, nil, nil, &response); err != nil {
		return nil, err
	}
	objects := standardObjects()
	for _, schema := range response.Results {
		name := strings.TrimSpace(schema.ObjectTypeID)
		if name == "" {
			name = strings.TrimSpace(schema.FullyQualifiedName)
		}
		description := strings.TrimSpace(schema.Labels.Plural)
		if description == "" {
			description = strings.TrimSpace(schema.Labels.Singular)
		}
		objects = append(objects, hubspotObject(name, description))
	}
	return objects, nil
}

func (p hubspotProvider) Describe(ctx context.Context, object discovery.Object) ([]discovery.Field, error) {
	if !safeObjectType(object.Name) {
		return nil, errors.New("HubSpot object type is not safe for a fixed collection path")
	}
	path := "/crm/v3/properties/" + object.Name
	requester, err := p.requester(httpMethodGet, propertiesPath)
	if err != nil {
		return nil, err
	}
	var response propertiesResponse
	if err := requester.DoJSON(ctx, httpMethodGet, path, nil, nil, &response); err != nil {
		return nil, err
	}
	fields := make([]discovery.Field, 0, len(response.Results))
	for _, property := range response.Results {
		fields = append(fields, discovery.Field{
			Name:        property.Name,
			Description: firstNonEmpty(property.Description, property.Label),
			// HubSpot property metadata does not declare a non-null guarantee.
			// Keep this true so the shared primary-key derivation never claims
			// one from hasUniqueValue alone; hs_object_id is used only when the
			// provider actually returned that property description.
			Nullable: true,
			Unique:   property.HasUniqueValue,
			Raw:      property,
		})
	}
	return fields, nil
}

func (p hubspotProvider) requester(method, declaredPath string) (*connsdk.Requester, error) {
	requester, err := p.runtime.RequesterFor(method, declaredPath)
	if err != nil {
		return nil, err
	}
	// Discovery, rather than each connector, owns 429 retry/backoff/progress.
	// RequesterFor can return the runtime's shared base requester when no
	// declaration-specific rate policy applies; copy the value before changing
	// retry policy because Describe runs concurrently.
	isolated := *requester
	isolated.DisableRetries = true
	return &isolated, nil
}

func hubspotFieldSchema(field discovery.Field) (json.RawMessage, error) {
	property, ok := field.Raw.(hubspotProperty)
	if !ok {
		return nil, errors.New("HubSpot field metadata is unavailable")
	}
	schema := map[string]any{}
	if description := firstNonEmpty(property.Description, property.Label); description != "" {
		schema["description"] = description
	}
	switch property.Type {
	case "bool":
		schema["type"] = "boolean"
	case "number":
		schema["type"] = "number"
	case "date":
		schema["type"] = "string"
		schema["format"] = "date"
	case "datetime":
		schema["type"] = "string"
		schema["format"] = "date-time"
	case "string", "enumeration":
		schema["type"] = "string"
	case "json", "object_coordinates":
		schema["type"] = "object"
	}
	if property.Type == "enumeration" {
		values := make([]string, 0, len(property.Options))
		for _, option := range property.Options {
			if option.Value != "" {
				values = append(values, option.Value)
			}
		}
		if len(values) > 0 {
			schema["enum"] = values
		}
	}
	if reference := strings.TrimSpace(property.ReferencedObjectType); reference != "" {
		schema["x-references"] = []string{reference}
	}
	raw, err := json.Marshal(schema)
	if err != nil {
		return nil, fmt.Errorf("marshal HubSpot property schema: %w", err)
	}
	return raw, nil
}

func standardObjects() []discovery.Object {
	names := []string{
		"calls", "companies", "communications", "contacts", "deals", "emails",
		"feedback_submissions", "line_items", "meetings", "notes", "products",
		"quotes", "tasks", "tickets",
	}
	objects := make([]discovery.Object, 0, len(names))
	for _, name := range names {
		objects = append(objects, hubspotObject(name, "HubSpot "+strings.ReplaceAll(name, "_", " ")+" records"))
	}
	return objects
}

func hubspotObject(name, description string) discovery.Object {
	return discovery.Object{
		Name:        name,
		Description: description,
		PrimaryKey:  []string{"hs_object_id"},
		CursorField: "hs_lastmodifieddate",
	}
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if value = strings.TrimSpace(value); value != "" {
			return value
		}
	}
	return ""
}

func safeObjectType(value string) bool {
	if value == "" || len(value) > 256 {
		return false
	}
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}

const httpMethodGet = "GET"

package asana

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const (
	asanaEventSourceContractPath       = "event_source_contract.json"
	asanaEventSourceContractSchemaPath = "event_source_contract.schema.json"
)

type asanaEventSourceContract struct {
	Schema        string `json:"$schema"`
	SchemaVersion int    `json:"schema_version"`
	Definition    struct {
		SyncTransport string                                  `json:"sync_transport"`
		Executor      connectors.TransportExecutorReference   `json:"executor"`
		Conformance   connectors.ConformanceEvidenceReference `json:"conformance"`
	} `json:"definition_binding"`
	SourceLock struct {
		Path          string `json:"path"`
		FileSHA256    string `json:"file_sha256"`
		SchemaVersion int    `json:"schema_version"`
		Connector     string `json:"connector"`
		RESTSHA256    string `json:"rest_sha256"`
	} `json:"source_lock"`
	Provider struct {
		EligibleStream string `json:"eligible_stream"`
		EventScope     struct {
			OperationID         string `json:"operation_id"`
			SourceLocation      string `json:"source_location"`
			Method              string `json:"method"`
			Path                string `json:"path"`
			ResourceParameter   string `json:"resource_parameter"`
			ResourceKind        string `json:"resource_kind"`
			EmittedResourceType string `json:"emitted_resource_type"`
			SyncParameter       string `json:"sync_parameter"`
			InitialSync         string `json:"initial_sync"`
			InitialStatus       int    `json:"initial_status"`
			ExpiredStatus       int    `json:"expired_status"`
			RecordsPointer      string `json:"records_pointer"`
			NextTokenPointer    string `json:"next_token_pointer"`
			HasMorePointer      string `json:"has_more_pointer"`
			EventTotalOrder     string `json:"event_total_order"`
		} `json:"event_scope"`
		EventRecord struct {
			Schema              string   `json:"schema"`
			ActionPointer       string   `json:"action_pointer"`
			ResourceGIDPointer  string   `json:"resource_gid_pointer"`
			ResourceTypePointer string   `json:"resource_type_pointer"`
			Actions             []string `json:"actions"`
			TombstoneAction     string   `json:"tombstone_action"`
			ScopeChangeAction   string   `json:"scope_change_action"`
		} `json:"event_record"`
		Hydration struct {
			OperationID     string `json:"operation_id"`
			SourceLocation  string `json:"source_location"`
			Method          string `json:"method"`
			Path            string `json:"path"`
			PathParameter   string `json:"path_parameter"`
			ResponsePointer string `json:"response_pointer"`
		} `json:"hydration"`
		Snapshot struct {
			OperationID         string `json:"operation_id"`
			SourceLocation      string `json:"source_location"`
			Method              string `json:"method"`
			Path                string `json:"path"`
			ScopeQueryParameter string `json:"scope_query_parameter"`
			RecordsPointer      string `json:"records_pointer"`
			ContinuationPointer string `json:"continuation_pointer"`
		} `json:"snapshot"`
		Security struct {
			EventOperation    []string `json:"event_operation"`
			TaskReadOperation []string `json:"task_read_operation"`
		} `json:"security"`
	} `json:"provider_contract"`
}

type asanaEventContractSourceLock struct {
	SchemaVersion int    `json:"schema_version"`
	Connector     string `json:"connector"`
	REST          struct {
		SHA256     string                              `json:"sha256"`
		Operations []asanaEventContractLockedOperation `json:"operations"`
	} `json:"rest"`
	SourceContract struct {
		Components struct {
			Parameters map[string]asanaEventContractParameter `json:"parameters"`
			Schemas    map[string]json.RawMessage             `json:"schemas"`
		} `json:"components"`
	} `json:"source_contract"`
}

type asanaEventContractLockedOperation struct {
	ID              string `json:"id"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	SourceLocation  string `json:"source_location"`
	SourceOperation struct {
		Description    string                        `json:"description"`
		Parameters     []asanaEventContractParameter `json:"parameters"`
		PathParameters []asanaEventContractParameter `json:"path_parameters"`
		Responses      map[string]json.RawMessage    `json:"responses"`
		Security       []map[string][]string         `json:"security"`
	} `json:"source_operation"`
}

type asanaEventContractParameter struct {
	Ref         string `json:"$ref"`
	Name        string `json:"name"`
	In          string `json:"in"`
	Required    *bool  `json:"required"`
	Description string `json:"description"`
}

func TestAsanaEventSourceContractIsSchemaValidAndSourceLocked(t *testing.T) {
	contractRaw, err := os.ReadFile(asanaEventSourceContractPath)
	if err != nil {
		t.Fatalf("read Asana event source contract: %v", err)
	}
	schemaRaw, err := os.ReadFile(asanaEventSourceContractSchemaPath)
	if err != nil {
		t.Fatalf("read Asana event source contract schema: %v", err)
	}
	schema, err := engine.CompileSchema(schemaRaw)
	if err != nil {
		t.Fatalf("compile Asana event source contract schema: %v", err)
	}
	var contractDocument any
	decoder := json.NewDecoder(bytes.NewReader(contractRaw))
	decoder.UseNumber()
	if err := decoder.Decode(&contractDocument); err != nil {
		t.Fatalf("decode Asana event source contract for schema validation: %v", err)
	}
	if err := schema.Validate(contractDocument); err != nil {
		t.Fatalf("validate Asana event source contract schema: %v", err)
	}

	var contract asanaEventSourceContract
	decoder = json.NewDecoder(bytes.NewReader(contractRaw))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&contract); err != nil {
		t.Fatalf("strict-decode Asana event source contract: %v", err)
	}
	if contract.Schema != asanaEventSourceContractSchemaPath || contract.SchemaVersion != 1 {
		t.Fatalf("event source contract identity = schema %q version %d", contract.Schema, contract.SchemaVersion)
	}

	bundle := loadBundle(t)
	if bundle.SyncTransport == nil || bundle.SyncTransport.Source == nil {
		t.Fatal("Asana bundle has no source transport to bind")
	}
	source := bundle.SyncTransport.Source
	if contract.Definition.SyncTransport != "sync_transport.json" || contract.Definition.Executor != source.Executor || contract.Definition.Conformance != source.Conformance {
		t.Fatalf("event source contract binding = %+v, transport executor/conformance = %+v/%+v", contract.Definition, source.Executor, source.Conformance)
	}
	if !reflect.DeepEqual(source.EligibleStreams, []string{contract.Provider.EligibleStream}) {
		t.Fatalf("event source contract stream %q does not match transport streams %v", contract.Provider.EligibleStream, source.EligibleStreams)
	}
	assertAsanaEventDefinitionProjection(t, bundle, contract)

	lockRaw, err := os.ReadFile(contract.SourceLock.Path)
	if err != nil {
		t.Fatalf("read event contract source lock %q: %v", contract.SourceLock.Path, err)
	}
	lockDigest := sha256.Sum256(lockRaw)
	if got := hex.EncodeToString(lockDigest[:]); got != contract.SourceLock.FileSHA256 {
		t.Fatalf("source-lock file SHA-256 = %s, want contract %s", got, contract.SourceLock.FileSHA256)
	}
	var lock asanaEventContractSourceLock
	if err := json.Unmarshal(lockRaw, &lock); err != nil {
		t.Fatalf("decode event contract source lock: %v", err)
	}
	if lock.SchemaVersion != contract.SourceLock.SchemaVersion || lock.Connector != contract.SourceLock.Connector || lock.REST.SHA256 != contract.SourceLock.RESTSHA256 {
		t.Fatalf("source-lock identity = schema %d connector %q REST %q, contract = %+v", lock.SchemaVersion, lock.Connector, lock.REST.SHA256, contract.SourceLock)
	}

	eventOperation := findAsanaEventContractOperation(t, lock.REST.Operations, contract.Provider.EventScope.OperationID)
	assertAsanaEventOperationContract(t, lock, eventOperation, contract)
	assertAsanaEventRecordContract(t, lock, contract)
	assertAsanaTaskHydrationContract(t, lock, contract)
	assertAsanaTaskSnapshotContract(t, lock, contract)
	assertAsanaEventSecurityContract(t, eventOperation, findAsanaEventContractOperation(t, lock.REST.Operations, contract.Provider.Hydration.OperationID), findAsanaEventContractOperation(t, lock.REST.Operations, contract.Provider.Snapshot.OperationID), contract)

	for _, forbidden := range []string{`"handler"`, `"hook"`, `"retry"`, `"page_cap"`, `"window_coalescing"`, `"checkpoint_commit"`} {
		if bytes.Contains(contractRaw, []byte(forbidden)) {
			t.Fatalf("provider evidence contract contains runtime policy/code-hook key %s", forbidden)
		}
	}
}

func TestAsanaEventSourceContractSchemaRejectsRuntimeHooksAndOpenExecutorIDs(t *testing.T) {
	schemaRaw, err := os.ReadFile(asanaEventSourceContractSchemaPath)
	if err != nil {
		t.Fatal(err)
	}
	schema, err := engine.CompileSchema(schemaRaw)
	if err != nil {
		t.Fatal(err)
	}
	contractRaw, err := os.ReadFile(asanaEventSourceContractPath)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(map[string]any)
	}{
		{
			name: "runtime lifecycle hook",
			mutate: func(document map[string]any) {
				provider := document["provider_contract"].(map[string]any)
				provider["retry"] = map[string]any{"handler": "arbitrary"}
			},
		},
		{
			name: "open executor id",
			mutate: func(document map[string]any) {
				binding := document["definition_binding"].(map[string]any)
				executor := binding["executor"].(map[string]any)
				executor["id"] = "caller_supplied_executor"
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var document map[string]any
			decoder := json.NewDecoder(bytes.NewReader(contractRaw))
			decoder.UseNumber()
			if err := decoder.Decode(&document); err != nil {
				t.Fatal(err)
			}
			test.mutate(document)
			if err := schema.Validate(document); err == nil {
				t.Fatal("closed event source contract schema accepted an executable/open extension")
			}
		})
	}
}

func assertAsanaEventDefinitionProjection(t *testing.T, bundle engine.Bundle, contract asanaEventSourceContract) {
	t.Helper()
	assertOperation := func(id, sourceID, method, path string) {
		t.Helper()
		for _, operation := range bundle.Operations {
			if operation.ID != id {
				continue
			}
			if operation.SourceOperation == nil || operation.SourceOperation.ID != sourceID || operation.SourceOperation.Method != method || operation.SourceOperation.Path != path {
				t.Fatalf("projected operation %q = %+v, want source %q %s %s", id, operation.SourceOperation, sourceID, method, path)
			}
			return
		}
		t.Fatalf("bundle has no projected operation %q", id)
	}
	assertOperation("get_events", contract.Provider.EventScope.OperationID, contract.Provider.EventScope.Method, contract.Provider.EventScope.Path)
	assertOperation("get_task", contract.Provider.Hydration.OperationID, contract.Provider.Hydration.Method, contract.Provider.Hydration.Path)
	assertOperation("get_tasks", contract.Provider.Snapshot.OperationID, contract.Provider.Snapshot.Method, contract.Provider.Snapshot.Path)

	for _, stream := range bundle.Streams {
		if stream.Name != contract.Provider.EligibleStream {
			continue
		}
		method := stream.Method
		if method == "" {
			method = "GET"
		}
		query, ok := stream.Query[contract.Provider.Snapshot.ScopeQueryParameter]
		if method != contract.Provider.Snapshot.Method || stream.Path != contract.Provider.Snapshot.Path || stream.Records.Path != strings.TrimPrefix(contract.Provider.Snapshot.RecordsPointer, "$.") || !ok || query.Template != "{{ config.project_id }}" || !query.OmitWhenAbsent {
			t.Fatalf("tasks snapshot stream = %+v, project query = %+v", stream, query)
		}
		if bundle.HTTP.Pagination == nil || bundle.HTTP.Pagination.NextURLPath != strings.TrimPrefix(contract.Provider.Snapshot.ContinuationPointer, "$.") {
			t.Fatalf("tasks snapshot pagination = %+v, want %s", bundle.HTTP.Pagination, contract.Provider.Snapshot.ContinuationPointer)
		}
		return
	}
	t.Fatalf("bundle has no projected stream %q", contract.Provider.EligibleStream)
}

func assertAsanaEventOperationContract(t *testing.T, lock asanaEventContractSourceLock, operation asanaEventContractLockedOperation, contract asanaEventSourceContract) {
	t.Helper()
	event := contract.Provider.EventScope
	if operation.Method != event.Method || operation.Path != event.Path || operation.SourceLocation != event.SourceLocation {
		t.Fatalf("event operation = %s %s at %s, contract = %s %s at %s", operation.Method, operation.Path, operation.SourceLocation, event.Method, event.Path, event.SourceLocation)
	}
	resource := findAsanaEventContractParameter(t, lock, operation.SourceOperation.PathParameters, event.ResourceParameter)
	if resource.In != "query" || resource.Required == nil || !*resource.Required || !strings.Contains(strings.ToLower(resource.Description), event.ResourceKind) {
		t.Fatalf("event resource parameter = %+v, want required %s query scope", resource, event.ResourceKind)
	}
	syncParameter := findAsanaEventContractParameter(t, lock, operation.SourceOperation.PathParameters, event.SyncParameter)
	if syncParameter.In != "query" || syncParameter.Required == nil || *syncParameter.Required {
		t.Fatalf("event sync parameter = %+v, want optional query", syncParameter)
	}
	syncDescription := strings.ToLower(syncParameter.Description)
	for _, phrase := range []string{"first request", "omit the sync token", "412 precondition failed", "fresh sync token"} {
		if !strings.Contains(syncDescription, phrase) {
			t.Fatalf("sync parameter does not document %q: %q", phrase, syncParameter.Description)
		}
	}
	if event.InitialSync != "omit" || event.InitialStatus != 412 || event.ExpiredStatus != 412 {
		t.Fatalf("event bootstrap/reset projection = %+v", event)
	}
	description := strings.ToLower(operation.SourceOperation.Description)
	if !strings.Contains(description, "has_more: true") || !strings.Contains(description, "subscription to a project will contain events for\ntasks contained within the project") {
		t.Fatalf("getEvents description does not prove continuation and project-task scope: %q", operation.SourceOperation.Description)
	}
	if strings.Contains(description, "ordered") || event.EventTotalOrder != "not_documented" {
		t.Fatalf("event total order = %q; provider description must not be promoted to source ordering", event.EventTotalOrder)
	}

	okResponse := decodeAsanaEventContractJSON(t, operation.SourceOperation.Responses["200"], "getEvents 200 response")
	okSchema := asanaEventContractObjectAt(t, okResponse, "content", "application/json", "schema")
	okProperties := asanaEventContractObjectAt(t, okSchema, "properties")
	assertAsanaEventContractResponseField(t, okProperties, strings.TrimPrefix(event.RecordsPointer, "$."), "array")
	assertAsanaEventContractResponseField(t, okProperties, strings.TrimPrefix(event.NextTokenPointer, "$."), "string")
	assertAsanaEventContractResponseField(t, okProperties, strings.TrimPrefix(event.HasMorePointer, "$."), "boolean")
	data := asanaEventContractObjectAt(t, okProperties, strings.TrimPrefix(event.RecordsPointer, "$."))
	items := asanaEventContractObjectAt(t, data, "items")
	if got := asanaEventContractStringAt(t, items, "$ref"); got != "#/components/schemas/"+contract.Provider.EventRecord.Schema {
		t.Fatalf("event data item ref = %q", got)
	}

	resetResponse := decodeAsanaEventContractJSON(t, operation.SourceOperation.Responses[fmt.Sprint(event.ExpiredStatus)], "getEvents 412 response")
	resetSchema := asanaEventContractObjectAt(t, resetResponse, "content", "application/json", "schema")
	resetProperties := asanaEventContractObjectAt(t, resetSchema, "properties")
	assertAsanaEventContractResponseField(t, resetProperties, strings.TrimPrefix(event.NextTokenPointer, "$."), "string")
	errorsSchema := asanaEventContractObjectAt(t, resetProperties, "errors", "items", "properties", "message")
	message := strings.ToLower(asanaEventContractStringAt(t, errorsSchema, "example"))
	if !strings.Contains(message, "fetch the full dataset") || !strings.Contains(message, "new sync token") {
		t.Fatalf("412 response does not document full-snapshot rebootstrap: %q", message)
	}
}

func assertAsanaEventRecordContract(t *testing.T, lock asanaEventContractSourceLock, contract asanaEventSourceContract) {
	t.Helper()
	record := contract.Provider.EventRecord
	raw, ok := lock.SourceContract.Components.Schemas[record.Schema]
	if !ok {
		t.Fatalf("source lock has no event schema %q", record.Schema)
	}
	schema := decodeAsanaEventContractJSON(t, raw, record.Schema)
	properties := asanaEventContractObjectAt(t, schema, "properties")
	action := asanaEventContractObjectAt(t, properties, strings.TrimPrefix(record.ActionPointer, "$."))
	actionDescription := strings.ToLower(asanaEventContractStringAt(t, action, "description"))
	wantActions := []string{"changed", "added", "removed", "deleted", "undeleted"}
	if !reflect.DeepEqual(record.Actions, wantActions) || record.TombstoneAction != "deleted" || record.ScopeChangeAction != "removed" {
		t.Fatalf("event action projection = %+v", record)
	}
	for _, actionName := range wantActions {
		if !strings.Contains(actionDescription, "`"+actionName+"`") {
			t.Fatalf("EventResponse.action does not document %q: %q", actionName, actionDescription)
		}
	}
	resource := asanaEventContractObjectAt(t, properties, "resource")
	allOf := asanaEventContractArrayAt(t, resource, "allOf")
	resourceRef := asanaEventContractObject(t, allOf[0], "EventResponse.resource allOf[0]")
	if got := asanaEventContractStringAt(t, resourceRef, "$ref"); got != "#/components/schemas/AsanaNamedResource" {
		t.Fatalf("EventResponse.resource ref = %q", got)
	}
	resourceSchema := decodeAsanaEventContractJSON(t, lock.SourceContract.Components.Schemas["AsanaNamedResource"], "AsanaNamedResource")
	resourceProperties := asanaEventContractObjectAt(t, resourceSchema, "properties")
	for _, pointer := range []string{record.ResourceGIDPointer, record.ResourceTypePointer} {
		field := strings.TrimPrefix(pointer, "$.resource.")
		assertAsanaEventContractResponseField(t, resourceProperties, field, "string")
	}
}

func assertAsanaTaskHydrationContract(t *testing.T, lock asanaEventContractSourceLock, contract asanaEventSourceContract) {
	t.Helper()
	hydration := contract.Provider.Hydration
	operation := findAsanaEventContractOperation(t, lock.REST.Operations, hydration.OperationID)
	if operation.Method != hydration.Method || operation.Path != hydration.Path || operation.SourceLocation != hydration.SourceLocation {
		t.Fatalf("hydration operation = %s %s at %s, contract = %+v", operation.Method, operation.Path, operation.SourceLocation, hydration)
	}
	parameter := findAsanaEventContractParameter(t, lock, operation.SourceOperation.PathParameters, hydration.PathParameter)
	if parameter.In != "path" || parameter.Required == nil || !*parameter.Required {
		t.Fatalf("hydration path parameter = %+v", parameter)
	}
	response := decodeAsanaEventContractJSON(t, operation.SourceOperation.Responses["200"], "getTask 200 response")
	schema := asanaEventContractObjectAt(t, response, "content", "application/json", "schema")
	properties := asanaEventContractObjectAt(t, schema, "properties")
	data := asanaEventContractObjectAt(t, properties, strings.TrimPrefix(hydration.ResponsePointer, "$."))
	if got := asanaEventContractStringAt(t, data, "$ref"); got != "#/components/schemas/TaskResponse" {
		t.Fatalf("hydration response ref = %q", got)
	}
}

func assertAsanaTaskSnapshotContract(t *testing.T, lock asanaEventContractSourceLock, contract asanaEventSourceContract) {
	t.Helper()
	snapshot := contract.Provider.Snapshot
	operation := findAsanaEventContractOperation(t, lock.REST.Operations, snapshot.OperationID)
	if operation.Method != snapshot.Method || operation.Path != snapshot.Path || operation.SourceLocation != snapshot.SourceLocation {
		t.Fatalf("snapshot operation = %s %s at %s, contract = %+v", operation.Method, operation.Path, operation.SourceLocation, snapshot)
	}
	parameter := findAsanaEventContractParameter(t, lock, operation.SourceOperation.Parameters, snapshot.ScopeQueryParameter)
	if parameter.In != "query" || !strings.Contains(strings.ToLower(parameter.Description), "project") {
		t.Fatalf("snapshot project filter = %+v", parameter)
	}
	response := decodeAsanaEventContractJSON(t, operation.SourceOperation.Responses["200"], "getTasks 200 response")
	schema := asanaEventContractObjectAt(t, response, "content", "application/json", "schema")
	properties := asanaEventContractObjectAt(t, schema, "properties")
	data := asanaEventContractObjectAt(t, properties, strings.TrimPrefix(snapshot.RecordsPointer, "$."))
	items := asanaEventContractObjectAt(t, data, "items")
	if got := asanaEventContractStringAt(t, items, "$ref"); got != "#/components/schemas/TaskCompact" {
		t.Fatalf("snapshot record ref = %q", got)
	}
	nextPage := asanaEventContractObjectAt(t, properties, strings.Split(strings.TrimPrefix(snapshot.ContinuationPointer, "$."), ".")[0])
	if got := asanaEventContractStringAt(t, nextPage, "$ref"); got != "#/components/schemas/NextPage" {
		t.Fatalf("snapshot continuation ref = %q", got)
	}
	nextPageSchema := decodeAsanaEventContractJSON(t, lock.SourceContract.Components.Schemas["NextPage"], "NextPage")
	nextPageProperties := asanaEventContractObjectAt(t, nextPageSchema, "properties")
	continuationField := strings.Split(strings.TrimPrefix(snapshot.ContinuationPointer, "$."), ".")[1]
	assertAsanaEventContractResponseField(t, nextPageProperties, continuationField, "string")
}

func assertAsanaEventSecurityContract(t *testing.T, event, hydration, snapshot asanaEventContractLockedOperation, contract asanaEventSourceContract) {
	t.Helper()
	wantEvent := []string{"personalAccessToken", "oauth2"}
	wantTaskRead := []string{"personalAccessToken", "oauth2:tasks:read"}
	if !reflect.DeepEqual(contract.Provider.Security.EventOperation, wantEvent) || !reflect.DeepEqual(contract.Provider.Security.TaskReadOperation, wantTaskRead) {
		t.Fatalf("event source security projection = %+v", contract.Provider.Security)
	}
	if !asanaEventContractHasSecurity(event, "personalAccessToken", "") || !asanaEventContractHasSecurity(event, "oauth2", "") {
		t.Fatalf("getEvents security = %+v", event.SourceOperation.Security)
	}
	for _, operation := range []asanaEventContractLockedOperation{hydration, snapshot} {
		if !asanaEventContractHasSecurity(operation, "personalAccessToken", "") || !asanaEventContractHasSecurity(operation, "oauth2", "tasks:read") {
			t.Fatalf("%s security = %+v, want PAT or OAuth tasks:read", operation.ID, operation.SourceOperation.Security)
		}
	}
}

func findAsanaEventContractOperation(t *testing.T, operations []asanaEventContractLockedOperation, id string) asanaEventContractLockedOperation {
	t.Helper()
	for _, operation := range operations {
		if operation.ID == id {
			return operation
		}
	}
	t.Fatalf("source lock has no operation %q", id)
	return asanaEventContractLockedOperation{}
}

func findAsanaEventContractParameter(t *testing.T, lock asanaEventContractSourceLock, parameters []asanaEventContractParameter, name string) asanaEventContractParameter {
	t.Helper()
	for _, parameter := range parameters {
		if parameter.Ref != "" {
			prefix := "#/components/parameters/"
			resolved, ok := lock.SourceContract.Components.Parameters[strings.TrimPrefix(parameter.Ref, prefix)]
			if !ok || !strings.HasPrefix(parameter.Ref, prefix) {
				t.Fatalf("source lock cannot resolve parameter ref %q", parameter.Ref)
			}
			parameter = resolved
		}
		if parameter.Name == name {
			return parameter
		}
	}
	t.Fatalf("source operation has no parameter %q", name)
	return asanaEventContractParameter{}
}

func asanaEventContractHasSecurity(operation asanaEventContractLockedOperation, scheme, scope string) bool {
	for _, alternative := range operation.SourceOperation.Security {
		scopes, ok := alternative[scheme]
		if !ok {
			continue
		}
		if scope == "" {
			return true
		}
		for _, candidate := range scopes {
			if candidate == scope {
				return true
			}
		}
	}
	return false
}

func decodeAsanaEventContractJSON(t *testing.T, raw json.RawMessage, label string) any {
	t.Helper()
	if len(raw) == 0 {
		t.Fatalf("%s is absent", label)
	}
	var value any
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.UseNumber()
	if err := decoder.Decode(&value); err != nil {
		t.Fatalf("decode %s: %v", label, err)
	}
	return value
}

func asanaEventContractObject(t *testing.T, value any, label string) map[string]any {
	t.Helper()
	object, ok := value.(map[string]any)
	if !ok {
		t.Fatalf("%s = %T, want object", label, value)
	}
	return object
}

func asanaEventContractObjectAt(t *testing.T, value any, path ...string) map[string]any {
	t.Helper()
	current := value
	for _, field := range path {
		object := asanaEventContractObject(t, current, strings.Join(path, "."))
		var ok bool
		current, ok = object[field]
		if !ok {
			t.Fatalf("JSON object has no field %q along %s", field, strings.Join(path, "."))
		}
	}
	return asanaEventContractObject(t, current, strings.Join(path, "."))
}

func asanaEventContractArrayAt(t *testing.T, value any, field string) []any {
	t.Helper()
	object := asanaEventContractObject(t, value, field)
	array, ok := object[field].([]any)
	if !ok {
		t.Fatalf("%s = %T, want array", field, object[field])
	}
	return array
}

func asanaEventContractStringAt(t *testing.T, value any, field string) string {
	t.Helper()
	object := asanaEventContractObject(t, value, field)
	result, ok := object[field].(string)
	if !ok {
		t.Fatalf("%s = %T, want string", field, object[field])
	}
	return result
}

func assertAsanaEventContractResponseField(t *testing.T, properties map[string]any, field, wantType string) {
	t.Helper()
	property := asanaEventContractObjectAt(t, properties, field)
	if got := asanaEventContractStringAt(t, property, "type"); got != wantType {
		t.Fatalf("response field %s type = %q, want %q", field, got, wantType)
	}
}

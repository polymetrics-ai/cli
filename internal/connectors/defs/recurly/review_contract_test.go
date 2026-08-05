package recurly

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	connectorDefs "polymetrics.ai/internal/connectors/defs"
	"polymetrics.ai/internal/connectors/engine"
)

type reviewWriteAction struct {
	Name                 string                     `json:"name"`
	Kind                 string                     `json:"kind"`
	Method               string                     `json:"method"`
	Path                 string                     `json:"path"`
	BodyType             string                     `json:"body_type"`
	BodyRequired         bool                       `json:"body_required"`
	BodyFields           []string                   `json:"body_fields"`
	PathFields           []string                   `json:"path_fields"`
	IdempotencyKeyHeader string                     `json:"idempotency_key_header"`
	RedactFields         []string                   `json:"redact_fields"`
	Query                map[string]json.RawMessage `json:"query"`
	Delete               *struct {
		Idempotent bool `json:"idempotent"`
	} `json:"delete"`
	RecordSchema json.RawMessage `json:"record_schema"`
}

type reviewCLICommand struct {
	Path      string          `json:"path"`
	Write     string          `json:"write"`
	Operation string          `json:"operation"`
	Flags     []reviewCLIFlag `json:"flags"`
}

type reviewCLIFlag struct {
	Name     string   `json:"name"`
	Type     string   `json:"type"`
	MapsTo   string   `json:"maps_to"`
	Required bool     `json:"required"`
	Values   []string `json:"values"`
}

type reviewStream struct {
	Name           string            `json:"name"`
	Records        map[string]any    `json:"records"`
	Projection     string            `json:"projection"`
	Schema         string            `json:"schema"`
	Pagination     map[string]any    `json:"pagination"`
	ComputedFields map[string]string `json:"computed_fields"`
}

func TestReviewWriteContractsMatchProviderShapes(t *testing.T) {
	var document struct {
		Actions []reviewWriteAction `json:"actions"`
	}
	readReviewJSON(t, "writes.json", &document)
	actions := make(map[string]reviewWriteAction, len(document.Actions))
	var jsonBodies, requiredBodies int
	for _, action := range document.Actions {
		actions[action.Name] = action
		assertReviewObjectBoundaries(t, reviewSchemaObject(t, action.RecordSchema), "write "+action.Name)
		if action.BodyType == "json" {
			jsonBodies++
		}
		if action.BodyRequired {
			requiredBodies++
		}
		if action.IdempotencyKeyHeader != "Idempotency-Key" {
			t.Errorf("write %q idempotency_key_header = %q", action.Name, action.IdempotencyKeyHeader)
		}
	}
	if jsonBodies != 59 || requiredBodies != 49 {
		t.Errorf("write body counts = json:%d required:%d, want json:59 required:49", jsonBodies, requiredBodies)
	}

	requiredBodyActions := []string{
		"update_account_acquisition", "update_billing_info", "update_a_billing_info",
		"update_shipping_address", "update_coupon", "generate_unique_coupon_codes",
		"restore_coupon", "update_general_ledger_account", "update_item",
		"update_measured_unit", "update_invoice", "record_external_transaction",
		"update_plan", "update_plan_add_on", "update_shipping_method",
		"create_subscription_change",
	}
	for _, name := range requiredBodyActions {
		action := requireReviewAction(t, actions, name)
		if action.BodyType != "json" || !action.BodyRequired || len(action.BodyFields) == 0 {
			t.Errorf("write %q body contract = type:%q required:%v fields:%v", name, action.BodyType, action.BodyRequired, action.BodyFields)
		}
	}

	giftCard := reviewSchemaObject(t, requireReviewAction(t, actions, "create_gift_card").RecordSchema)
	requireReviewFields(t, giftCard, "delivery", "gifter_account", "product_code", "unit_amount", "currency")
	for _, field := range []string{"delivery", "gifter_account"} {
		property := reviewProperty(t, giftCard, field)
		if !reviewSchemaHasType(property, "object") || property["additionalProperties"] != false {
			t.Errorf("create_gift_card.%s is not a closed object: %#v", field, property)
		}
	}
	plan := reviewSchemaObject(t, requireReviewAction(t, actions, "create_plan").RecordSchema)
	requireReviewFields(t, plan, "code", "name", "currencies")

	subscription := reviewSchemaObject(t, requireReviewAction(t, actions, "update_subscription").RecordSchema)
	policy := reviewProperty(t, subscription, "credit_application_policy")
	if !reviewSchemaHasType(policy, "object") || policy["additionalProperties"] != false {
		t.Errorf("credit_application_policy = %#v, want closed object", policy)
	}
	requireReviewFields(t, policy, "mode")

	billing := requireReviewAction(t, actions, "create_billing_info")
	for _, field := range []string{"iban", "sort_code", "tax_identifier"} {
		if !reviewContains(billing.RedactFields, field) {
			t.Errorf("create_billing_info redact_fields lacks %q", field)
		}
	}

	acquisition := reviewSchemaObject(t, requireReviewAction(t, actions, "update_account_acquisition").RecordSchema)
	costItems, _ := reviewProperty(t, acquisition, "cost")["items"].(map[string]any)
	if costItems["additionalProperties"] != true {
		t.Errorf("update_account_acquisition.cost items = %#v, want provider-declared free-form object", costItems)
	}
}

func TestReviewCLISurfaceCarriesTypedRequiredBodies(t *testing.T) {
	var writesDocument struct {
		Actions []reviewWriteAction `json:"actions"`
	}
	readReviewJSON(t, "writes.json", &writesDocument)
	actions := make(map[string]reviewWriteAction, len(writesDocument.Actions))
	for _, action := range writesDocument.Actions {
		actions[action.Name] = action
	}

	var cliDocument struct {
		Commands []reviewCLICommand `json:"commands"`
	}
	readReviewJSON(t, "cli_surface.json", &cliDocument)
	for _, command := range cliDocument.Commands {
		action, ok := actions[command.Write]
		if !ok || !action.BodyRequired {
			continue
		}
		pathFields := make(map[string]bool, len(action.PathFields))
		for _, field := range action.PathFields {
			pathFields[field] = true
		}
		hasBodyInput := false
		for _, flag := range command.Flags {
			field := strings.TrimPrefix(flag.MapsTo, "record.")
			if strings.HasPrefix(flag.MapsTo, "record.") && !pathFields[field] {
				hasBodyInput = true
				break
			}
		}
		if !hasBodyInput {
			t.Errorf("command %q for required-body write %q has no typed body input", command.Path, command.Write)
		}
	}

	for _, command := range cliDocument.Commands {
		if command.Operation != "preview_gift_card" {
			continue
		}
		for _, flag := range command.Flags {
			if flag.MapsTo == "body.unit_amount" {
				if flag.Type != "number" {
					t.Errorf("preview gift-card unit_amount flag type = %q, want number", flag.Type)
				}
				return
			}
		}
		t.Error("preview gift-card command lacks body.unit_amount flag")
		return
	}
	t.Error("preview_gift_card command is missing")
}

func TestReviewMutationQueryControls(t *testing.T) {
	var writesDocument struct {
		Actions []reviewWriteAction `json:"actions"`
	}
	readReviewJSON(t, "writes.json", &writesDocument)
	actions := make(map[string]reviewWriteAction, len(writesDocument.Actions))
	for _, action := range writesDocument.Actions {
		actions[action.Name] = action
	}

	deactivate := requireReviewAction(t, actions, "deactivate_account")
	assertReviewQuery(t, deactivate, "redact", "{{ record.redact }}", false)
	deactivateSchema := reviewSchemaObject(t, deactivate.RecordSchema)
	if !reviewSchemaHasType(reviewProperty(t, deactivateSchema, "redact"), "boolean") {
		t.Errorf("deactivate_account.redact is not boolean")
	}
	requireReviewFields(t, deactivateSchema, "redact")

	terminate := requireReviewAction(t, actions, "terminate_subscription")
	assertReviewQuery(t, terminate, "refund", "{{ record.refund }}", false)
	assertReviewQuery(t, terminate, "charge", "{{ record.charge }}", false)
	terminateSchema := reviewSchemaObject(t, terminate.RecordSchema)
	refund := reviewProperty(t, terminateSchema, "refund")
	if !reviewSchemaHasType(refund, "string") || !reflect.DeepEqual(reviewStringSlice(refund["enum"]), []string{"full", "none", "partial"}) {
		t.Errorf("terminate_subscription.refund = %#v", refund)
	}
	if !reviewSchemaHasType(reviewProperty(t, terminateSchema, "charge"), "boolean") {
		t.Errorf("terminate_subscription.charge is not boolean")
	}
	requireReviewFields(t, terminateSchema, "refund", "charge")

	var cliDocument struct {
		Commands []reviewCLICommand `json:"commands"`
	}
	readReviewJSON(t, "cli_surface.json", &cliDocument)
	deactivateCommand := requireReviewCommand(t, cliDocument.Commands, "deactivate_account")
	redactFlag := requireReviewFlag(t, deactivateCommand, "record.redact")
	if redactFlag.Type != "boolean" || !redactFlag.Required {
		t.Errorf("deactivate redact flag = %#v", redactFlag)
	}
	terminateCommand := requireReviewCommand(t, cliDocument.Commands, "terminate_subscription")
	refundFlag := requireReviewFlag(t, terminateCommand, "record.refund")
	if refundFlag.Type != "enum" || !refundFlag.Required || !reflect.DeepEqual(refundFlag.Values, []string{"full", "none", "partial"}) {
		t.Errorf("terminate refund flag = %#v", refundFlag)
	}
	chargeFlag := requireReviewFlag(t, terminateCommand, "record.charge")
	if chargeFlag.Type != "boolean" || !chargeFlag.Required {
		t.Errorf("terminate charge flag = %#v", chargeFlag)
	}

	assertReviewWriteFixtureQuery(t, "deactivate_account", map[string]string{"redact": "true"})
	assertReviewWriteFixtureQuery(t, "terminate_subscription", map[string]string{"refund": "partial", "charge": "false"})
}

func TestReviewTerminateSubscriptionRequiresChargeBeforePreview(t *testing.T) {
	bundle, err := engine.Load(connectorDefs.FS, "recurly")
	if err != nil {
		t.Fatalf("load Recurly bundle: %v", err)
	}
	req := connectors.WriteRequest{
		Action: "terminate_subscription",
		Config: connectors.RuntimeConfig{
			CredentialRevision:  "recurly-fixture-credential-revision",
			ConfigurationDigest: "recurly-fixture-configuration-digest",
		},
	}
	withoutCharge := connectors.Record{"subscription_id": "fixture_subscription_id", "refund": "none"}
	if err := engine.ValidateWrite(context.Background(), bundle, req, []connectors.Record{withoutCharge}); err == nil {
		t.Fatal("terminate_subscription validation accepted a record without charge")
	}
	if _, err := engine.DryRunWrite(context.Background(), bundle, req, []connectors.Record{withoutCharge}, nil); err == nil {
		t.Fatal("terminate_subscription preview accepted a record without charge")
	}

	withCharge := connectors.Record{"subscription_id": "fixture_subscription_id", "refund": "none", "charge": false}
	if err := engine.ValidateWrite(context.Background(), bundle, req, []connectors.Record{withCharge}); err != nil {
		t.Fatalf("terminate_subscription validation with charge: %v", err)
	}
	if _, err := engine.DryRunWrite(context.Background(), bundle, req, []connectors.Record{withCharge}, nil); err != nil {
		t.Fatalf("terminate_subscription preview with charge: %v", err)
	}
}

func TestReviewDeleteRetryEvidenceIsOperationScoped(t *testing.T) {
	var writesDocument struct {
		Actions []reviewWriteAction `json:"actions"`
	}
	readReviewJSON(t, "writes.json", &writesDocument)

	var evidence struct {
		OpenAPISHA256 string `json:"openapi_sha256"`
		Operations    []struct {
			Action                         string `json:"action"`
			OperationID                    string `json:"operation_id"`
			Method                         string `json:"method"`
			Path                           string `json:"path"`
			OperationSourceURL             string `json:"operation_source_url"`
			OperationEvidencePath          string `json:"operation_evidence_path"`
			IntrinsicIdempotencyDocumented bool   `json:"intrinsic_idempotency_documented"`
			RetryMode                      string `json:"retry_mode"`
			IdempotencyKeyHeader           string `json:"idempotency_key_header"`
			IdempotencySourceURL           string `json:"idempotency_source_url"`
			IdempotencyEvidencePath        string `json:"idempotency_evidence_path"`
		} `json:"operations"`
		MutationQueryControls []struct {
			Action           string         `json:"action"`
			OperationID      string         `json:"operation_id"`
			Field            string         `json:"field"`
			LocalSourcePath  string         `json:"local_source_path"`
			ProviderRequired bool           `json:"provider_required"`
			LocalRequired    bool           `json:"local_required"`
			SourceURL        string         `json:"source_url"`
			EvidenceType     string         `json:"evidence_type"`
			EvidencePath     string         `json:"evidence_path"`
			Schema           map[string]any `json:"schema"`
		} `json:"mutation_query_controls"`
	}
	readReviewJSON(t, filepath.Join("..", "..", "..", "..", ".planning", "phases", "recurly-parity-resume-r1", "RECURLY-WRITE-RETRY-RESEARCH.json"), &evidence)
	if evidence.OpenAPISHA256 != "b98a3f85d0a1190c2c8e11f57fa5ec13b841665e658596dcb5d7f3ddce70baca" {
		t.Errorf("retry evidence OAS digest = %q", evidence.OpenAPISHA256)
	}
	byOperation := make(map[string]bool, len(evidence.Operations))
	for _, operation := range evidence.Operations {
		byOperation[operation.Action] = operation.IntrinsicIdempotencyDocumented
		if operation.Method != "DELETE" || operation.Path == "" || operation.OperationSourceURL == "" || operation.OperationEvidencePath == "" {
			t.Errorf("incomplete operation evidence for %q: %#v", operation.Action, operation)
		}
		if operation.RetryMode != "provider_idempotency_key" || operation.IdempotencyKeyHeader != "Idempotency-Key" || operation.IdempotencySourceURL == "" || operation.IdempotencyEvidencePath == "" {
			t.Errorf("incomplete retry evidence for %q: %#v", operation.Action, operation)
		}
	}

	var deletes int
	for _, action := range writesDocument.Actions {
		if action.Kind != "delete" {
			continue
		}
		deletes++
		documented, ok := byOperation[action.Name]
		if !ok {
			t.Errorf("delete %q has no operation-scoped retry evidence", action.Name)
			continue
		}
		declared := action.Delete != nil && action.Delete.Idempotent
		if declared != documented {
			t.Errorf("delete %q idempotent declaration = %v, documented = %v", action.Name, declared, documented)
		}
	}
	if deletes != 23 || len(evidence.Operations) != deletes {
		t.Errorf("delete evidence count = %d for %d actions, want 23", len(evidence.Operations), deletes)
	}
	wantControls := map[string]bool{
		"deactivate_account.query.redact":     true,
		"terminate_subscription.query.refund": true,
		"terminate_subscription.query.charge": true,
	}
	if len(evidence.MutationQueryControls) != len(wantControls) {
		t.Errorf("mutation query evidence count = %d, want %d", len(evidence.MutationQueryControls), len(wantControls))
	}
	for _, control := range evidence.MutationQueryControls {
		key := control.Action + "." + control.Field
		wantRequired, ok := wantControls[key]
		if !ok {
			t.Errorf("unexpected mutation query evidence %q", key)
			continue
		}
		if control.LocalRequired != wantRequired || control.LocalSourcePath == "" || control.SourceURL == "" || control.EvidenceType != "openapi.parameter" || control.EvidencePath == "" || len(control.Schema) == 0 {
			t.Errorf("incomplete mutation query evidence %q: %#v", key, control)
		}
		if control.ProviderRequired {
			t.Errorf("mutation query evidence %q unexpectedly marks provider-required", key)
		}
		if key == "terminate_subscription.query.charge" {
			description, _ := control.Schema["description"].(string)
			if control.Schema["type"] != "boolean" || control.Schema["default"] != true || !strings.Contains(description, "If true") || !strings.Contains(description, "If false") {
				t.Errorf("charge evidence does not preserve Recurly's documented boolean values: %#v", control.Schema)
			}
		}
	}
}

func TestReviewStreamContractsAreProviderShaped(t *testing.T) {
	var streamsDocument struct {
		Base    map[string]any `json:"base"`
		Streams []reviewStream `json:"streams"`
	}
	readReviewJSON(t, "streams.json", &streamsDocument)
	if _, ok := streamsDocument.Base["rate_limit"]; ok {
		t.Error("streams base still enforces a connector-wide rate_limit")
	}
	pagination, _ := streamsDocument.Base["pagination"].(map[string]any)
	if pagination["page_size"] != float64(200) {
		t.Errorf("base pagination page_size = %#v, want 200", pagination["page_size"])
	}

	legacy := map[string]bool{"accounts": true, "invoices": true, "plans": true, "subscriptions": true, "transactions": true}
	for _, stream := range streamsDocument.Streams {
		if legacy[stream.Name] {
			continue
		}
		var schema map[string]any
		readReviewJSON(t, stream.Schema, &schema)
		assertReviewObjectBoundaries(t, schema, stream.Schema)
		if schema["additionalProperties"] != false {
			t.Errorf("schema %q is not closed", stream.Schema)
		}
		properties, _ := schema["properties"].(map[string]any)
		primaryKey := reviewStringSlice(schema["x-primary-key"])
		if len(primaryKey) == 0 {
			t.Errorf("schema %q has no durable primary key", stream.Schema)
		}
		for _, field := range primaryKey {
			if _, ok := properties[field]; !ok {
				t.Errorf("schema %q primary key %q is undeclared", stream.Schema, field)
			}
		}

		fixturePath := filepath.Join("fixtures", "streams", stream.Name, "page_1.json")
		var fixture struct {
			Response struct {
				Body any `json:"body"`
			} `json:"response"`
		}
		readReviewJSON(t, fixturePath, &fixture)
		record := reviewFixtureRecord(fixture.Response.Body, stream.Records)
		for field := range record {
			if _, ok := properties[field]; !ok {
				t.Errorf("fixture %q invents undeclared field %q", fixturePath, field)
			}
		}
		for _, field := range reviewStringSlice(schema["required"]) {
			if _, ok := record[field]; !ok && stream.ComputedFields[field] == "" {
				t.Errorf("fixture %q lacks required field %q", fixturePath, field)
			}
		}
	}

	performance := requireReviewStream(t, streamsDocument.Streams, "get_performance_obligations")
	if performance.Records["path"] != "data" || performance.Records["single_object"] == true || performance.Pagination["type"] == "none" {
		t.Errorf("get_performance_obligations extraction = records:%#v pagination:%#v", performance.Records, performance.Pagination)
	}
	entitlements := requireReviewStream(t, streamsDocument.Streams, "list_entitlements")
	if entitlements.ComputedFields["customer_permission_id"] != "{{ record.customer_permission.id }}" {
		t.Errorf("list_entitlements computed_fields = %#v", entitlements.ComputedFields)
	}

	if _, err := os.Stat(filepath.Join("fixtures", "streams", "list_sites", "page_2.json")); err != nil {
		t.Errorf("list_sites second page fixture: %v", err)
	}
	var sitePage struct {
		Response struct {
			Headers map[string][]string `json:"headers"`
			Body    struct {
				Data []map[string]any `json:"data"`
			} `json:"body"`
		} `json:"response"`
	}
	readReviewJSON(t, filepath.Join("fixtures", "streams", "list_sites", "page_1.json"), &sitePage)
	if len(sitePage.Response.Headers["Link"]) == 0 {
		t.Error("list_sites first page has no Link response header")
	}
	if len(sitePage.Response.Body.Data) == 0 {
		t.Fatal("list_sites first page has no record")
	}
	for _, field := range []string{"code", "state"} {
		if _, ok := sitePage.Response.Body.Data[0][field]; ok {
			t.Errorf("list_sites fixture invents Site.%s", field)
		}
	}

	var transactionSchema map[string]any
	readReviewJSON(t, filepath.Join("schemas", "get_transaction.json"), &transactionSchema)
	gatewayValues := reviewProperty(t, transactionSchema, "gateway_response_values")
	if gatewayValues["additionalProperties"] != true {
		t.Errorf("gateway_response_values = %#v, want provider-declared free-form object", gatewayValues)
	}
}

func readReviewJSON(t *testing.T, path string, target any) {
	t.Helper()
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if err := json.Unmarshal(raw, target); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}
}

func requireReviewAction(t *testing.T, actions map[string]reviewWriteAction, name string) reviewWriteAction {
	t.Helper()
	action, ok := actions[name]
	if !ok {
		t.Fatalf("write %q is missing", name)
	}
	return action
}

func assertReviewQuery(t *testing.T, action reviewWriteAction, name, wantTemplate string, wantOmit bool) {
	t.Helper()
	raw, ok := action.Query[name]
	if !ok {
		t.Fatalf("write %q query lacks %q", action.Name, name)
	}
	var template string
	if err := json.Unmarshal(raw, &template); err == nil {
		if wantOmit {
			t.Fatalf("write %q query %q is a plain template, want omit_when_absent", action.Name, name)
		}
	} else {
		var value struct {
			Template       string `json:"template"`
			OmitWhenAbsent bool   `json:"omit_when_absent"`
		}
		if err := json.Unmarshal(raw, &value); err != nil {
			t.Fatalf("parse write %q query %q: %v", action.Name, name, err)
		}
		template = value.Template
		if value.OmitWhenAbsent != wantOmit {
			t.Errorf("write %q query %q omit_when_absent = %v, want %v", action.Name, name, value.OmitWhenAbsent, wantOmit)
		}
	}
	if template != wantTemplate {
		t.Errorf("write %q query %q template = %q, want %q", action.Name, name, template, wantTemplate)
	}
}

func requireReviewCommand(t *testing.T, commands []reviewCLICommand, write string) reviewCLICommand {
	t.Helper()
	for _, command := range commands {
		if command.Write == write {
			return command
		}
	}
	t.Fatalf("CLI command for write %q is missing", write)
	return reviewCLICommand{}
}

func requireReviewFlag(t *testing.T, command reviewCLICommand, mapsTo string) reviewCLIFlag {
	t.Helper()
	for _, flag := range command.Flags {
		if flag.MapsTo == mapsTo {
			return flag
		}
	}
	t.Fatalf("CLI command %q lacks flag mapped to %q", command.Path, mapsTo)
	return reviewCLIFlag{}
}

func assertReviewWriteFixtureQuery(t *testing.T, name string, want map[string]string) {
	t.Helper()
	var fixture struct {
		Record map[string]any `json:"record"`
		Expect struct {
			Query map[string]string `json:"query"`
		} `json:"expect"`
	}
	readReviewJSON(t, filepath.Join("fixtures", "writes", name+".json"), &fixture)
	if !reflect.DeepEqual(fixture.Expect.Query, want) {
		t.Errorf("write fixture %q query = %#v, want %#v", name, fixture.Expect.Query, want)
	}
	for field, value := range want {
		if got, ok := fixture.Record[field]; !ok || fmt.Sprint(got) != value {
			t.Errorf("write fixture %q record.%s = %#v, want %q", name, field, got, value)
		}
	}
}

func reviewSchemaObject(t *testing.T, raw json.RawMessage) map[string]any {
	t.Helper()
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatalf("parse record schema: %v", err)
	}
	return schema
}

func reviewProperty(t *testing.T, schema map[string]any, name string) map[string]any {
	t.Helper()
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatalf("schema has no properties: %#v", schema)
	}
	property, ok := properties[name].(map[string]any)
	if !ok {
		t.Fatalf("schema has no property %q", name)
	}
	return property
}

func requireReviewFields(t *testing.T, schema map[string]any, fields ...string) {
	t.Helper()
	required := reviewStringSlice(schema["required"])
	sort.Strings(required)
	for _, field := range fields {
		if !reviewContains(required, field) {
			t.Errorf("required fields %v lack %q", required, field)
		}
	}
}

func reviewSchemaHasType(schema map[string]any, want string) bool {
	switch value := schema["type"].(type) {
	case string:
		return value == want
	case []any:
		for _, item := range value {
			if item == want {
				return true
			}
		}
	}
	return false
}

func assertReviewObjectBoundaries(t *testing.T, schema map[string]any, path string) {
	t.Helper()
	properties, hasProperties := schema["properties"].(map[string]any)
	if reviewSchemaHasType(schema, "object") {
		if hasProperties && len(properties) > 0 {
			if schema["additionalProperties"] != false {
				t.Errorf("%s is not closed: %#v", path, schema["additionalProperties"])
			}
		}
	}
	for name, property := range properties {
		child, _ := property.(map[string]any)
		assertReviewObjectBoundaries(t, child, path+"."+name)
	}
	if items, ok := schema["items"].(map[string]any); ok {
		assertReviewObjectBoundaries(t, items, path+"[]")
	}
}

func reviewContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

func reviewStringSlice(value any) []string {
	items, _ := value.([]any)
	result := make([]string, 0, len(items))
	for _, item := range items {
		if text, ok := item.(string); ok {
			result = append(result, text)
		}
	}
	return result
}

func reviewFixtureRecord(body any, records map[string]any) map[string]any {
	object, _ := body.(map[string]any)
	if records["single_object"] == true || records["path"] == "." {
		return object
	}
	data, _ := object["data"].([]any)
	if len(data) == 0 {
		return map[string]any{}
	}
	record, _ := data[0].(map[string]any)
	return record
}

func requireReviewStream(t *testing.T, streams []reviewStream, name string) reviewStream {
	t.Helper()
	for _, stream := range streams {
		if stream.Name == name {
			return stream
		}
	}
	t.Fatalf("stream %q is missing", name)
	return reviewStream{}
}

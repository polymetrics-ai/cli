package shopify

import (
	"encoding/json"
	"os"
	"strings"
	"testing"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/engine"
)

const shopifyReferenceRetrievedAt = "2026-08-06"

type shopifySourceInventory struct {
	ReviewedAt string `json:"reviewed_at"`
	Counts     struct {
		GraphQLQueries               int `json:"graphql_queries"`
		GraphQLMutations             int `json:"graphql_mutations"`
		RESTGetRows                  int `json:"rest_get_rows"`
		RESTPostRows                 int `json:"rest_post_rows"`
		RESTPutRows                  int `json:"rest_put_rows"`
		RESTDeleteRows               int `json:"rest_delete_rows"`
		RESTRows                     int `json:"rest_rows"`
		APISurfaceRows               int `json:"api_surface_rows"`
		TypedDestructiveRESTWriteOps int `json:"typed_destructive_rest_write_operations"`
		StaticCLICommands            int `json:"static_cli_commands"`
	} `json:"counts"`
	Rows []shopifySourceInventoryRow `json:"rows"`
}

type shopifySourceInventoryRow struct {
	Key            string `json:"key"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	CitationURL    string `json:"citation_url"`
	SourceArtifact string `json:"source_artifact"`
	RetrievedAt    string `json:"retrieved_at"`
}

func TestPublishedReferenceInventoryHasOneCitedStaticCommandPerOperation(t *testing.T) {
	bundle := loadBundle(t)
	if bundle.Surface == nil || bundle.CLISurface == nil {
		t.Fatal("Shopify bundle is missing api_surface or cli_surface")
	}

	raw, err := os.ReadFile("source_inventory.json")
	if err != nil {
		t.Fatalf("read source_inventory.json: %v", err)
	}
	var inventory shopifySourceInventory
	if err := json.Unmarshal(raw, &inventory); err != nil {
		t.Fatalf("parse source_inventory.json: %v", err)
	}

	if inventory.ReviewedAt != shopifyReferenceRetrievedAt {
		t.Fatalf("inventory reviewed_at = %q, want %q", inventory.ReviewedAt, shopifyReferenceRetrievedAt)
	}
	if got, want := len(bundle.Surface.Endpoints), 1098; got != want {
		t.Fatalf("api_surface endpoint count = %d, want %d", got, want)
	}
	if got, want := len(inventory.Rows), 1098; got != want {
		t.Fatalf("source inventory row count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.GraphQLQueries, 287; got != want {
		t.Fatalf("graphql query count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.GraphQLMutations, 518; got != want {
		t.Fatalf("graphql mutation count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.RESTGetRows, 152; got != want {
		t.Fatalf("REST GET count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.RESTPostRows, 73; got != want {
		t.Fatalf("REST POST count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.RESTPutRows, 35; got != want {
		t.Fatalf("REST PUT count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.RESTDeleteRows, 33; got != want {
		t.Fatalf("REST DELETE count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.RESTRows, 293; got != want {
		t.Fatalf("REST total = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.APISurfaceRows, 1098; got != want {
		t.Fatalf("inventory api_surface count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.StaticCLICommands, 1098; got != want {
		t.Fatalf("inventory static command count = %d, want %d", got, want)
	}
	if got, want := inventory.Counts.TypedDestructiveRESTWriteOps, 33; got != want {
		t.Fatalf("inventory typed destructive REST write count = %d, want %d", got, want)
	}

	citations := make(map[string]shopifySourceInventoryRow, len(inventory.Rows))
	for _, row := range inventory.Rows {
		key := shopifyEndpointKey(row.Method, row.Path)
		if row.Key != key {
			t.Fatalf("source row key = %q, want %q", row.Key, key)
		}
		if _, exists := citations[key]; exists {
			t.Fatalf("duplicate source inventory row %q", key)
		}
		if row.RetrievedAt != shopifyReferenceRetrievedAt {
			t.Fatalf("source row %q retrieval date = %q, want %q", key, row.RetrievedAt, shopifyReferenceRetrievedAt)
		}
		if !strings.HasPrefix(row.CitationURL, "https://shopify.dev/docs/") {
			t.Fatalf("source row %q citation = %q, want official Shopify documentation URL", key, row.CitationURL)
		}
		if !strings.HasPrefix(row.SourceArtifact, "https://shopify.dev/docs/") {
			t.Fatalf("source row %q artifact = %q, want official Shopify documentation URL", key, row.SourceArtifact)
		}
		citations[key] = row
	}

	commandPaths := map[string]bool{}
	commandReferences := map[string]int{}
	for _, command := range bundle.CLISurface.Commands {
		if strings.Contains(command.Path, "<") || strings.Contains(command.Path, ">") {
			t.Fatalf("command path %q is generic rather than static", command.Path)
		}
		if commandPaths[command.Path] {
			t.Fatalf("duplicate command path %q", command.Path)
		}
		commandPaths[command.Path] = true
		if command.Path == "shop read" {
			if command.Intent != "etl" || command.Availability != "implemented" || command.Stream != "shop" || command.SourceCLIPath != "GET /admin/api/latest/shop.json" {
				t.Fatalf("shop read command = %+v, want implemented shop ETL stream", command)
			}
			continue
		}
		if len(command.APISurface) != 0 {
			t.Fatalf("planned command %q claims api_surface execution binding", command.Path)
		}
		if !strings.Contains(command.SourceCLIPath, " ") {
			t.Fatalf("command %q source_cli_path = %q, want exact METHOD endpoint mapping", command.Path, command.SourceCLIPath)
		}
		commandReferences[command.SourceCLIPath]++
		if strings.TrimSpace(command.SourceURL) == "" {
			t.Fatalf("command %q has no source_url", command.Path)
		}
		if command.Intent == "raw_api" {
			t.Fatalf("command %q exposes forbidden raw_api intent", command.Path)
		}
	}
	if got, want := len(commandPaths), 1098; got != want {
		t.Fatalf("distinct command count = %d, want %d", got, want)
	}

	for _, endpoint := range bundle.Surface.Endpoints {
		key := shopifyEndpointKey(endpoint.Method, endpoint.Path)
		if _, ok := citations[key]; !ok {
			t.Fatalf("api_surface endpoint %q lacks a per-row source citation", key)
		}
		if endpoint.CoveredBy != nil && endpoint.CoveredBy.Stream == "shop" {
			continue
		}
		if got, want := commandReferences[key], 1; got != want {
			t.Fatalf("api_surface endpoint %q has %d static command references, want %d", key, got, want)
		}
	}
}

func TestTypedDestructiveDeleteDeclarationsAreCompleteAndFixtureBacked(t *testing.T) {
	bundle := loadBundle(t)
	if bundle.Surface == nil || bundle.CLISurface == nil {
		t.Fatal("Shopify bundle is missing api_surface or cli_surface")
	}

	connector := engine.New(bundle, nil)
	deleteEndpoints := map[string]engine.SurfaceEndpoint{}
	for _, endpoint := range bundle.Surface.Endpoints {
		if endpoint.Method == "DELETE" {
			deleteEndpoints[endpoint.Path] = endpoint
		}
	}
	if got, want := len(deleteEndpoints), 33; got != want {
		t.Fatalf("DELETE api_surface endpoints = %d, want %d", got, want)
	}
	if got, want := len(bundle.Operations), 33; got != want {
		t.Fatalf("typed destructive operations = %d, want %d", got, want)
	}

	commandByOperation := map[string]engine.CLICommand{}
	for _, command := range bundle.CLISurface.Commands {
		if command.Operation != "" {
			commandByOperation[command.Operation] = command
		}
	}

	fixtures, err := os.ReadDir("fixtures/writes")
	if err != nil {
		t.Fatalf("read typed-delete fixtures: %v", err)
	}
	fixtureNames := map[string]bool{}
	for _, fixture := range fixtures {
		if fixture.IsDir() || !strings.HasSuffix(fixture.Name(), ".json") {
			continue
		}
		fixtureNames[strings.TrimSuffix(fixture.Name(), ".json")] = true
	}
	if got, want := len(fixtureNames), 33; got != want {
		t.Fatalf("typed-delete fixture count = %d, want %d", got, want)
	}

	for _, operation := range bundle.Operations {
		if operation.Kind != "rest_write" || operation.REST == nil {
			t.Fatalf("operation %q is not a REST write: %+v", operation.ID, operation)
		}
		if operation.REST.Method != "DELETE" || operation.MutationClass != "destructive" || !operation.Destructive {
			t.Fatalf("operation %q is not a destructive DELETE: %+v", operation.ID, operation)
		}
		if operation.Confirmation == nil || operation.Confirmation.Kind != connectors.ConfirmationKindDestructive {
			t.Fatalf("operation %q confirmation = %+v, want typed destructive confirmation", operation.ID, operation.Confirmation)
		}
		if operation.IsBatchable() || operation.OutputPolicy != "json" || operation.REST.MaxBytes != 1048576 {
			t.Fatalf("operation %q policy = batchable:%t output:%q max_bytes:%d, want false/json/1048576", operation.ID, operation.IsBatchable(), operation.OutputPolicy, operation.REST.MaxBytes)
		}
		endpoint, ok := deleteEndpoints[operation.REST.Path]
		if !ok || endpoint.Operation == nil || endpoint.Operation.Model != "destructive_action" {
			t.Fatalf("operation %q path %q has no destructive api_surface operation row", operation.ID, operation.REST.Path)
		}
		metadata, err := connector.OperationDirectWriteMetadata(operation.ID)
		if err != nil {
			t.Fatalf("operation metadata for %q: %v", operation.ID, err)
		}
		if metadata.ConfirmationChallenge != string(connectors.ConfirmationKindDestructive) || metadata.OutputPolicy != "json" || metadata.Batchable {
			t.Fatalf("operation metadata for %q = %+v, want destructive/json/non-batchable", operation.ID, metadata)
		}

		command, ok := commandByOperation[operation.ID]
		if !ok {
			t.Fatalf("operation %q lacks a static command declaration", operation.ID)
		}
		if command.Intent != "direct_write" || command.Availability != "planned" || len(command.APISurface) != 0 {
			t.Fatalf("command for %q = %+v, want planned typed direct_write", operation.ID, command)
		}
		if got, want := command.SourceCLIPath, "DELETE "+operation.REST.Path; got != want {
			t.Fatalf("command for %q source_cli_path = %q, want %q", operation.ID, got, want)
		}
		for _, variable := range shopifyPathVariables(operation.REST.Path) {
			if !shopifyRequiredPathFlag(command, variable) {
				t.Fatalf("command %q has no required typed flag for operation path variable %q", command.Path, variable)
			}
		}

		fixtureName := strings.TrimPrefix(operation.ID, "shopify.")
		if !fixtureNames[fixtureName] {
			t.Fatalf("operation %q lacks fixture %s.json", operation.ID, fixtureName)
		}
	}
}

func shopifyEndpointKey(method, path string) string {
	return strings.ToUpper(method) + " " + path
}

func shopifyPathVariables(path string) []string {
	var names []string
	for rest := path; ; {
		start := strings.Index(rest, "{")
		if start < 0 {
			return names
		}
		rest = rest[start+1:]
		end := strings.Index(rest, "}")
		if end < 0 {
			return names
		}
		names = append(names, rest[:end])
		rest = rest[end+1:]
	}
}

func shopifyRequiredPathFlag(command engine.CLICommand, variable string) bool {
	for _, flag := range command.Flags {
		if flag.MapsTo == "path."+variable && flag.Required {
			return true
		}
	}
	return false
}

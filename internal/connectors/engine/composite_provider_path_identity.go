package engine

import (
	"fmt"
	"io/fs"
)

// CompositeProviderPathIdentity cites the sole admitted composite provider
// identity. Its complete shape is validated against the engine-owned CircleCI
// manifest; it is not a template-substitution mechanism.
type CompositeProviderPathIdentity struct {
	SourceURL    string                         `json:"source_url"`
	SourceSHA256 string                         `json:"source_sha256"`
	Placeholder  string                         `json:"placeholder"`
	ConfigKeys   []string                       `json:"config_keys"`
	Bindings     []CompositeProviderPathBinding `json:"bindings"`
}

// CompositeProviderPathBinding records one provider operation and the one
// command binding eligible to use the closed composite identity proof.
type CompositeProviderPathBinding struct {
	SourceID            string `json:"source_id"`
	ProviderOperationID string `json:"provider_operation_id"`
	SourceLocation      string `json:"source_location"`
	Intent              string `json:"intent"`
	BindingKind         string `json:"binding_kind"`
	BindingID           string `json:"binding_id"`
	Method              string `json:"method"`
	Path                string `json:"path"`
}

const compositeProviderPathIdentityFile = "composite_provider_path_identity.json"

const (
	circleCICompositeProviderPathSourceURL   = "https://circleci.com/api/v2/openapi.json"
	circleCICompositeProviderPathSourceSHA   = "61c6ce11e8de509948aa3d53dcd0169913f52de20920b130b6a85dea41d66d07"
	circleCICompositeProviderPathPlaceholder = "project-slug"
)

var circleCICompositeProviderPathConfigKeys = []string{"vcs_type", "org", "repo"}

var circleCICompositeProviderPathBindings = []CompositeProviderPathBinding{
	{SourceID: "circleci.rest.getProjectBySlug", ProviderOperationID: "getProjectBySlug", SourceLocation: `paths["/project/{project-slug}"].get`, Intent: "etl", BindingKind: "stream", BindingID: "projects", Method: "GET", Path: "/project/{project-slug}"},
	{SourceID: "circleci.rest.listPipelinesForProject", ProviderOperationID: "listPipelinesForProject", SourceLocation: `paths["/project/{project-slug}/pipeline"].get`, Intent: "etl", BindingKind: "stream", BindingID: "pipelines", Method: "GET", Path: "/project/{project-slug}/pipeline"},
	{SourceID: "circleci.rest.listSchedulesForProject", ProviderOperationID: "listSchedulesForProject", SourceLocation: `paths["/project/{project-slug}/schedule"].get`, Intent: "etl", BindingKind: "stream", BindingID: "schedules", Method: "GET", Path: "/project/{project-slug}/schedule"},
	{SourceID: "circleci.rest.listCheckoutKeys", ProviderOperationID: "listCheckoutKeys", SourceLocation: `paths["/project/{project-slug}/checkout-key"].get`, Intent: "etl", BindingKind: "stream", BindingID: "checkout_keys", Method: "GET", Path: "/project/{project-slug}/checkout-key"},
	{SourceID: "circleci.rest.listEnvVars", ProviderOperationID: "listEnvVars", SourceLocation: `paths["/project/{project-slug}/envvar"].get`, Intent: "etl", BindingKind: "stream", BindingID: "environment_variables", Method: "GET", Path: "/project/{project-slug}/envvar"},
	{SourceID: "circleci.rest.getProjectWorkflowMetrics", ProviderOperationID: "getProjectWorkflowMetrics", SourceLocation: `paths["/insights/{project-slug}/workflows"].get`, Intent: "etl", BindingKind: "stream", BindingID: "insights_workflow_summary", Method: "GET", Path: "/insights/{project-slug}/workflows"},
	{SourceID: "circleci.rest.createSchedule", ProviderOperationID: "createSchedule", SourceLocation: `paths["/project/{project-slug}/schedule"].post`, Intent: "reverse_etl", BindingKind: "write", BindingID: "create_schedule", Method: "POST", Path: "/project/{project-slug}/schedule"},
	{SourceID: "circleci.rest.createEnvVar", ProviderOperationID: "createEnvVar", SourceLocation: `paths["/project/{project-slug}/envvar"].post`, Intent: "reverse_etl", BindingKind: "write", BindingID: "create_environment_variable", Method: "POST", Path: "/project/{project-slug}/envvar"},
	{SourceID: "circleci.rest.createCheckoutKey", ProviderOperationID: "createCheckoutKey", SourceLocation: `paths["/project/{project-slug}/checkout-key"].post`, Intent: "reverse_etl", BindingKind: "write", BindingID: "create_checkout_key", Method: "POST", Path: "/project/{project-slug}/checkout-key"},
	{SourceID: "circleci.rest.deleteEnvVar", ProviderOperationID: "deleteEnvVar", SourceLocation: `paths["/project/{project-slug}/envvar/{name}"].delete`, Intent: "reverse_etl", BindingKind: "write", BindingID: "delete_environment_variable", Method: "DELETE", Path: "/project/{project-slug}/envvar/{name}"},
	{SourceID: "circleci.rest.deleteCheckoutKey", ProviderOperationID: "deleteCheckoutKey", SourceLocation: `paths["/project/{project-slug}/checkout-key/{fingerprint}"].delete`, Intent: "reverse_etl", BindingKind: "write", BindingID: "delete_checkout_key", Method: "DELETE", Path: "/project/{project-slug}/checkout-key/{fingerprint}"},
}

// validateCompositeProviderPathIdentity intentionally admits one fully named
// configuration, not a class of placeholder expansions. Keeping the complete
// source identity here means another connector cannot opt in by declaring
// fields that merely look similar.
func validateCompositeProviderPathIdentity(connector string, identity *CompositeProviderPathIdentity) error {
	if identity == nil {
		return nil
	}
	if connector != "circleci" {
		return fmt.Errorf("composite provider path identity is only defined for circleci, not %q", connector)
	}
	if identity.SourceURL != circleCICompositeProviderPathSourceURL {
		return fmt.Errorf("composite provider path identity source_url = %q, want the retained CircleCI OpenAPI source", identity.SourceURL)
	}
	if identity.SourceSHA256 != circleCICompositeProviderPathSourceSHA {
		return fmt.Errorf("composite provider path identity source_sha256 does not match the retained CircleCI OpenAPI source")
	}
	if identity.Placeholder != circleCICompositeProviderPathPlaceholder {
		return fmt.Errorf("composite provider path identity placeholder = %q, want %q", identity.Placeholder, circleCICompositeProviderPathPlaceholder)
	}
	if !sameCompositeProviderPathStrings(identity.ConfigKeys, circleCICompositeProviderPathConfigKeys) {
		return fmt.Errorf("composite provider path identity config_keys = %q, want %q in that order", identity.ConfigKeys, circleCICompositeProviderPathConfigKeys)
	}
	if len(identity.Bindings) != len(circleCICompositeProviderPathBindings) {
		return fmt.Errorf("composite provider path identity has %d bindings, want the closed CircleCI set of %d", len(identity.Bindings), len(circleCICompositeProviderPathBindings))
	}
	for index, want := range circleCICompositeProviderPathBindings {
		if got := identity.Bindings[index]; got != want {
			return fmt.Errorf("composite provider path identity binding %d = %+v, want the retained CircleCI source binding %+v", index, got, want)
		}
	}
	return nil
}

func sameCompositeProviderPathStrings(got, want []string) bool {
	if len(got) != len(want) {
		return false
	}
	for index := range want {
		if got[index] != want[index] {
			return false
		}
	}
	return true
}

func loadCompositeProviderPathIdentity(sub fs.FS, dirName string) (*CompositeProviderPathIdentity, error) {
	if !fileExists(sub, compositeProviderPathIdentityFile) {
		return nil, nil
	}
	raw, err := readFile(sub, compositeProviderPathIdentityFile)
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.compositeProviderPathIdentity.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: %s: %w", dirName, compositeProviderPathIdentityFile, err)
	}
	var identity CompositeProviderPathIdentity
	if err := strictDecode(raw, &identity); err != nil {
		return nil, fmt.Errorf("load bundle %s: %s: %w", dirName, compositeProviderPathIdentityFile, err)
	}
	if err := validateCompositeProviderPathIdentity(dirName, &identity); err != nil {
		return nil, fmt.Errorf("load bundle %s: %s: %w", dirName, compositeProviderPathIdentityFile, err)
	}
	return &identity, nil
}

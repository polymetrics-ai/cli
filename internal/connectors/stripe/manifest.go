package stripe

import "polymetrics/internal/connectors"

// allSyncModes mirrors connectors.allSyncModes (unexported there). Kept local so
// the stripe package does not depend on an unexported connectors helper.
func allSyncModes() []string {
	return []string{
		"full_refresh_append",
		"full_refresh_overwrite",
		"full_refresh_overwrite_deduped",
		"incremental_append",
		"incremental_append_deduped",
	}
}

// readSourceSyncModes mirrors connectors.readSourceSyncModes (unexported there).
func readSourceSyncModes() []string {
	return []string{"full_refresh", "incremental"}
}

// Manifest implements connectors.ManifestProvider for the Stripe connector. The
// connectors registry and guide renderer detect this method via the
// ManifestProvider interface. It documents the config, auth, streams, write
// actions, and the id-cursor pagination shape so the generated MANUAL/SKILL are
// complete for the reference declarative-HTTP template.
func (c Connector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		ConfigFields: []connectors.ConfigField{
			{Name: "base_url", Description: "Stripe API base URL override for tests or proxies.", Default: stripeDefaultBaseURL},
			{Name: "account_id", Description: "Optional Stripe account ID; sent as the Stripe-Account header for Connect."},
			{Name: "start_date", Description: "RFC3339 lower bound; only objects created at or after this time are read."},
			{Name: "page_size", Description: "Records per page (1-100).", Default: "100"},
			{Name: "max_pages", Description: "Maximum pages; use 0, all, or unlimited to exhaust the stream.", Default: "0"},
			{Name: "mode", Description: "Runtime mode: live (default) or fixture for credential-free conformance."},
		},
		SecretFields: []connectors.SecretField{
			{Name: "client_secret", Description: "Stripe secret API key (sk_...). Used only for Bearer auth; never logged.", Required: true},
		},
		AuthModes: []connectors.AuthModeSpec{
			{Name: "fixture", Description: "Fixture-backed conformance mode; no credentials required.", Read: true, Write: true},
			{Name: "api_key", Description: "Live Stripe secret API key via Bearer auth.", SecretFields: []string{"client_secret"}, Read: true, Write: true},
		},
		Streams:         stripeStreams(),
		WriteActions:    stripeWriteActionSpecs(),
		SyncModes:       allSyncModes(),
		SourceSyncModes: readSourceSyncModes(),
		Pagination: connectors.PaginationSpec{
			Type:           "id_cursor",
			PageSizeField:  "page_size",
			PageLimitField: "max_pages",
			DelayField:     "page_delay",
			DefaultLimit:   "0",
		},
		Risk: connectors.RiskSpec{
			Read:     "external Stripe API read of customer and billing data",
			Write:    "external Stripe API mutation",
			Mutation: "creates or updates Stripe customers through allow-listed reverse ETL actions",
			Approval: "reverse ETL plan approval required before writes",
		},
	}
}

// stripeWriteActionSpecs documents the allow-listed reverse-ETL actions. It is
// derived from stripeWriteActions so the manifest and the executable allow-list
// never drift.
func stripeWriteActionSpecs() []connectors.WriteActionSpec {
	return []connectors.WriteActionSpec{
		{
			Name:           "create_customer",
			Description:    "Create a Stripe customer.",
			RequiredFields: []string{"email"},
			OptionalFields: []string{"name", "description", "phone"},
			Method:         stripeWriteActions["create_customer"].method,
			Path:           "/" + stripeWriteActions["create_customer"].resource,
			Risk:           "external mutation; approval required",
		},
		{
			Name:           "update_customer",
			Description:    "Update an existing Stripe customer addressed by id.",
			RequiredFields: []string{"id"},
			OptionalFields: []string{"email", "name", "description", "phone"},
			Method:         stripeWriteActions["update_customer"].method,
			Path:           "/" + stripeWriteActions["update_customer"].resource + "/{id}",
			Risk:           "external mutation; approval required",
		},
	}
}

package freshdesk

import "polymetrics/internal/connectors"

// allSyncModes mirrors connectors.allSyncModes (unexported there). Kept local so
// the freshdesk package does not depend on an unexported connectors helper.
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

// Manifest implements connectors.ManifestProvider for the Freshdesk connector. It
// documents the config, auth, streams, and the Link-header pagination shape so
// the generated MANUAL/SKILL are complete. Freshdesk is exposed read-only, so no
// write actions are declared.
func (c Connector) Manifest() connectors.Manifest {
	return connectors.Manifest{
		Metadata: c.Metadata(),
		ConfigFields: []connectors.ConfigField{
			{Name: "domain", Description: "Freshdesk domain, e.g. acme.freshdesk.com. Used to build the https://<domain>/api/v2 base URL.", Required: true},
			{Name: "base_url", Description: "Freshdesk base URL override for tests or proxies; /api/v2 is appended automatically."},
			{Name: "start_date", Description: "RFC3339 lower bound; for tickets, only objects updated at or after this time are read (updated_since)."},
			{Name: "page_size", Description: "Records per page (1-100).", Default: "100"},
			{Name: "max_pages", Description: "Maximum pages; use 0, all, or unlimited to exhaust the stream.", Default: "0"},
			{Name: "mode", Description: "Runtime mode: live (default) or fixture for credential-free conformance."},
		},
		SecretFields: []connectors.SecretField{
			{Name: "api_key", Description: "Freshdesk API key. Sent as the HTTP Basic username (password is the literal X); never logged.", Required: true},
		},
		AuthModes: []connectors.AuthModeSpec{
			{Name: "fixture", Description: "Fixture-backed conformance mode; no credentials required.", Read: true},
			{Name: "api_key", Description: "Live Freshdesk API key via HTTP Basic auth (api_key:X).", SecretFields: []string{"api_key"}, Read: true},
		},
		Streams:         freshdeskStreams(),
		SyncModes:       allSyncModes(),
		SourceSyncModes: readSourceSyncModes(),
		Pagination: connectors.PaginationSpec{
			Type:           "link_header",
			PageSizeField:  "page_size",
			PageLimitField: "max_pages",
			DefaultLimit:   "0",
		},
		Risk: connectors.RiskSpec{
			Read: "external Freshdesk API read of support tickets, contacts, companies, agents, and groups",
		},
	}
}

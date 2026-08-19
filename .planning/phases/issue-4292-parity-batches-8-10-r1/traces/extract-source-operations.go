// Source-operation extractor for issue #4292.
//
// It only reads provider-published documentation artifacts. It emits the
// exact bytes/hash for every artifact, then derives REST operations from an
// OpenAPI/Swagger document or an official rendered-reference document API.
package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html"
	"io"
	"net/http"
	"os"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

type sourceConfig struct {
	Kind                 string
	URLs                 []string
	ReferenceIndexURLs   []string
	LandingURL           string
	BaseURL              string
	PathPrefix           string
	PathStripPrefix      string
	ExtractAbsolutePaths bool
	Confidence           string
	Basis                string
}

type document struct {
	SourceURL       string `json:"source_url"`
	SHA256          string `json:"sha256"`
	Bytes           int    `json:"bytes"`
	ContentType     string `json:"content_type"`
	RetrievalMethod string `json:"retrieval_method"`
	Representation  string `json:"representation"`
}

type operation struct {
	ID             string `json:"id"`
	Protocol       string `json:"protocol"`
	Method         string `json:"method"`
	Path           string `json:"path"`
	OperationID    string `json:"operation_id"`
	Deprecated     bool   `json:"deprecated"`
	SourceLocation string `json:"source_location"`
	SourceURL      string `json:"source_url"`
}

type result struct {
	SourceURL      string      `json:"source_url"`
	Documents      []document  `json:"documents"`
	Operations     []operation `json:"operations"`
	Representation string      `json:"representation"`
	Confidence     string      `json:"coverage_confidence"`
	Basis          string      `json:"coverage_basis"`
}

var configs = map[string]sourceConfig{
	"brex": {
		Kind: "openapi", URLs: []string{
			"https://developer.brex.com/_bundle/openapi/transactions_api.yaml",
			"https://developer.brex.com/_bundle/openapi/expenses_api.yaml",
			"https://developer.brex.com/_bundle/openapi/team_api.yaml",
			"https://developer.brex.com/_bundle/openapi/payments_api.yaml",
			"https://developer.brex.com/_bundle/openapi/webhooks_api.yaml",
			"https://developer.brex.com/_bundle/openapi/budgets_api.yaml",
			"https://developer.brex.com/_bundle/openapi/accounting_api.yaml",
			"https://developer.brex.com/_bundle/openapi/fields_api.yaml",
			"https://developer.brex.com/_bundle/openapi/onboarding_api.yaml",
			"https://developer.brex.com/_bundle/openapi/travel_api.yaml",
		}, Confidence: "machine-readable", Basis: "Brex publishes this complete per-product OpenAPI bundle at developer.brex.com; every HTTP path item in all ten current product documents is extracted."},
	"zoho-books": {Kind: "zip-openapi", URLs: []string{"https://www.zoho.com/books/api/v3/openapi-all.zip"}, Confidence: "machine-readable", Basis: "Zoho Books publishes its complete API v3 OpenAPI bundle as a ZIP; every HTTP path item in every bundled API description is extracted."},
	"amplitude": {Kind: "html-reference", ReferenceIndexURLs: []string{"https://amplitude.com/docs/llms-full.txt"}, URLs: []string{
		"https://amplitude.com/docs/api/content/apis/analytics/ai-feedback",
		"https://amplitude.com/docs/api/content/apis/analytics/attribution",
		"https://amplitude.com/docs/api/content/apis/analytics/batch-event-upload",
		"https://amplitude.com/docs/api/content/apis/analytics/behavioral-cohorts",
		"https://amplitude.com/docs/api/content/apis/analytics/ccpa-dsar",
		"https://amplitude.com/docs/api/content/apis/analytics/channel-classifiers",
		"https://amplitude.com/docs/api/content/apis/analytics/chart-annotations",
		"https://amplitude.com/docs/api/content/apis/analytics/dashboard-rest",
		"https://amplitude.com/docs/api/content/apis/analytics/event-streaming-metrics",
		"https://amplitude.com/docs/api/content/apis/analytics/export",
		"https://amplitude.com/docs/api/content/apis/analytics/group-identify",
		"https://amplitude.com/docs/api/content/apis/analytics/http-api-quickstart",
		"https://amplitude.com/docs/api/content/apis/analytics/http-v2",
		"https://amplitude.com/docs/api/content/apis/analytics/identify",
		"https://amplitude.com/docs/api/content/apis/analytics/lookup-table",
		"https://amplitude.com/docs/api/content/apis/analytics/lookup-table-2",
		"https://amplitude.com/docs/api/content/apis/analytics/releases",
		"https://amplitude.com/docs/api/content/apis/analytics/scim",
		"https://amplitude.com/docs/api/content/apis/analytics/session-replay",
		"https://amplitude.com/docs/api/content/apis/analytics/taxonomy",
		"https://amplitude.com/docs/api/content/apis/analytics/user-mapping",
		"https://amplitude.com/docs/api/content/apis/analytics/user-privacy",
		"https://amplitude.com/docs/api/content/apis/analytics/user-privacy-v2",
		"https://amplitude.com/docs/api/content/apis/analytics/user-profile",
		"https://amplitude.com/docs/api/content/apis/audit-logs",
		"https://amplitude.com/docs/api/content/apis/authentication",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-evaluation-api",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-management-api",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-management-api-deployments",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-management-api-experiments",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-management-api-flags",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-management-api-holdouts",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-management-api-mutexs",
		"https://amplitude.com/docs/api/content/apis/experiment/experiment-management-api-version-endpoints",
		"https://amplitude.com/docs/api/content/apis/guides-and-surveys/guides-and-surveys-api-localization",
		"https://amplitude.com/docs/api/content/apis/keys-and-tokens",
		"https://amplitude.com/docs/api/content/apis/user-management",
		"https://amplitude.com/docs/api/content/apis/user-management/effective-access/get-project-effective-access",
		"https://amplitude.com/docs/api/content/apis/user-management/effective-access/get-user-effective-access",
		"https://amplitude.com/docs/api/content/apis/user-management/groups/remove-group-member",
		"https://amplitude.com/docs/api/content/apis/user-management/project-role-assignments/create-project-role-assignment",
		"https://amplitude.com/docs/api/content/apis/user-management/project-role-assignments/delete-project-role-assignment",
		"https://amplitude.com/docs/api/content/apis/user-management/project-role-assignments/get-project-role-assignment",
		"https://amplitude.com/docs/api/content/apis/user-management/project-role-assignments/list-project-role-assignments",
		"https://amplitude.com/docs/api/content/apis/user-management/project-role-assignments/update-project-role-assignment",
		"https://amplitude.com/docs/api/content/apis/user-management/roles/get-role",
		"https://amplitude.com/docs/api/content/apis/user-management/roles/list-roles",
	}, ExtractAbsolutePaths: true, Confidence: "complete-rendered-reference", Basis: "Amplitude's provider-published llms-full corpus enumerates the public /docs/apis reference pages. The corpus and every enumerated API reference page are retrieved; formal method/path declarations from the complete rendered reference are deduplicated to provider operations."},
	"coda":     {Kind: "openapi", URLs: []string{"https://coda.io/apis/v1/openapi.json"}, Confidence: "machine-readable", Basis: "Coda publishes the complete v1 OpenAPI document at this URL; every HTTP path item is extracted."},
	"posthog":  {Kind: "openapi", URLs: []string{"https://app.posthog.com/api/schema/"}, Confidence: "machine-readable", Basis: "PostHog publicly serves its complete OpenAPI schema at /api/schema/; every HTTP path item is extracted."},
	"metabase": {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/metabase/metabase/master/resources/openapi/openapi.json"}, Confidence: "machine-readable", Basis: "Metabase publishes the generated server OpenAPI document in its official repository; every HTTP path item is extracted."},
	"dbt":      {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/dbt-labs/dbt-cloud-openapi-spec/master/openapi-v2.yaml"}, Confidence: "machine-readable", Basis: "dbt Labs publishes the dbt Cloud v2 OpenAPI document in its official repository; every HTTP path item is extracted."},
	"looker":   {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/looker-open-source/sdk-codegen/main/packages/sdk-codegen-utils/data/Looker.4.0.oas.json"}, Confidence: "machine-readable", Basis: "Looker publishes its API 4.0 OpenAPI document in its official SDK code-generation repository; every HTTP path item is extracted."},
	"mode":     {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/mode/api-swagger/master/swagger.json"}, Confidence: "machine-readable", Basis: "Mode publishes its Swagger document in its official repository; every HTTP path item is extracted."},
	"dremio": {Kind: "html-reference", URLs: []string{
		"https://docs.dremio.com/25.x/reference/api/",
		"https://docs.dremio.com/25.x/reference/api/catalog/",
		"https://docs.dremio.com/25.x/reference/api/catalog/container-folder",
		"https://docs.dremio.com/25.x/reference/api/catalog/container-space",
		"https://docs.dremio.com/25.x/reference/api/catalog/source/",
		"https://docs.dremio.com/25.x/reference/api/datasets/",
		"https://docs.dremio.com/25.x/reference/api/external-token-providers/",
		"https://docs.dremio.com/25.x/reference/api/job/",
		"https://docs.dremio.com/25.x/reference/api/ldap-authorization",
		"https://docs.dremio.com/25.x/reference/api/nodeCollections/",
		"https://docs.dremio.com/25.x/reference/api/reflections/",
		"https://docs.dremio.com/25.x/reference/api/reflections/reflection-summary/",
		"https://docs.dremio.com/25.x/reference/api/roles/",
		"https://docs.dremio.com/25.x/reference/api/scripts/",
		"https://docs.dremio.com/25.x/reference/api/source",
		"https://docs.dremio.com/25.x/reference/api/sql/",
		"https://docs.dremio.com/25.x/reference/api/token/",
		"https://docs.dremio.com/25.x/reference/api/user/",
		"https://docs.dremio.com/25.x/reference/api/wlm/",
	}, PathStripPrefix: "/api/v3", Confidence: "complete-rendered-reference", Basis: "Dremio's official API Reference index and every linked Reference API category page are retrieved. Their formal /api/v3 method/path declarations are extracted; the version prefix is removed only because this connector's declared base URL owns it."},
	"clickup-api":  {Kind: "readme-reference", LandingURL: "https://developer.clickup.com/reference", BaseURL: "https://developer.clickup.com/clickup/api-next/v2/branches/2.0", Confidence: "complete-rendered-reference", Basis: "ClickUp's official public ReadMe reference sidebar and every public REST endpoint document are retrieved; each endpoint document supplies its rendered method and path."},
	"calendly":     {Kind: "openapi", URLs: []string{"https://stoplight.io/api/v1/projects/calendly/api-docs/nodes/fern/apis/public-api/openapi.yaml?fromExportButton=true&snapshotType=http_operation"}, Confidence: "machine-readable", Basis: "Calendly's official developer site publishes this Scheduling API OpenAPI document through its Stoplight project; every HTTP path item is extracted."},
	"ashby":        {Kind: "readme-reference", LandingURL: "https://developers.ashbyhq.com/reference", BaseURL: "https://developers.ashbyhq.com/ashby/api-next/v2/branches/1.0", Confidence: "complete-rendered-reference", Basis: "Ashby's official public ReadMe reference sidebar and every public REST endpoint document are retrieved; webhook callbacks are counted separately and are not REST operations."},
	"workable":     {Kind: "readme-reference", LandingURL: "https://workable.readme.io/reference", BaseURL: "https://workable.readme.io/workable/api-next/v2/branches/3.21.0", Confidence: "complete-rendered-reference", Basis: "Workable's official public ReadMe reference sidebar and every public REST endpoint document are retrieved; each endpoint document supplies its rendered method and path."},
	"recruitee":    {Kind: "html-reference", URLs: []string{"https://apidocs.recruitee.com/"}, Confidence: "complete-rendered-reference", Basis: "Recruitee publishes the complete REST reference as one public rendered document; every formal METHOD /path declaration is extracted after stripping presentation markup."},
	"hibob":        {Kind: "readme-reference", LandingURL: "https://apidocs.hibob.com/", BaseURL: "https://apidocs.hibob.com/hibob/api-next/v2/branches/1", Confidence: "complete-rendered-reference", Basis: "HiBob's official public ReadMe reference sidebar and every public REST endpoint document are retrieved; each endpoint document supplies its rendered method and path."},
	"factorial":    {Kind: "readme-reference", LandingURL: "https://apidoc.factorialhr.com/", BaseURL: "https://apidoc.factorialhr.com/factorial/api-next/v2/branches/1.0", Confidence: "complete-rendered-reference", Basis: "Factorial's official public ReadMe reference sidebar and every public REST endpoint document are retrieved; each endpoint document supplies its rendered method and path."},
	"lever-hiring": {Kind: "html-reference", URLs: []string{"https://hire.lever.co/developer/documentation?output=1"}, Confidence: "complete-rendered-reference", Basis: "Lever publishes the complete Hiring API reference as one public rendered document; every formal METHOD /path declaration in that document is extracted and example query strings are not counted as distinct operations."},
	"datadog": {Kind: "openapi", URLs: []string{
		"https://raw.githubusercontent.com/DataDog/datadog-api-client-go/master/.generator/schemas/v1/openapi.yaml",
		"https://raw.githubusercontent.com/DataDog/datadog-api-client-go/master/.generator/schemas/v2/openapi.yaml",
	}, Confidence: "machine-readable", Basis: "Datadog publishes the complete current v1 and v2 OpenAPI documents in its official client repository; every HTTP path item in both documents is extracted."},
	"pagerduty":     {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/PagerDuty/api-schema/main/reference/REST/openapiv3.json"}, Confidence: "machine-readable", Basis: "PagerDuty publishes its REST OpenAPI document in the official api-schema repository; every HTTP path item is extracted."},
	"auth0":         {Kind: "openapi", URLs: []string{"https://auth0.com/docs/oas/management/v2/management-api-oas.json"}, PathPrefix: "/api/v2", Confidence: "machine-readable", Basis: "Auth0 links to this Management API OpenAPI v3.1 document from its official reference; its declared server base path /api/v2 is retained before every extracted HTTP path item."},
	"okta":          {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/okta/okta-management-openapi-spec/master/dist/current/management-minimal.yaml"}, Confidence: "machine-readable", Basis: "Okta publishes this current management OpenAPI document in its official repository; every HTTP path item is extracted."},
	"firehydrant":   {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/firehydrant/firehydrant-go-sdk/main/openapi.yaml"}, Confidence: "machine-readable", Basis: "FireHydrant publishes the current OpenAPI document in its official generated Go SDK repository; every HTTP path item is extracted."},
	"commercetools": {Kind: "openapi", URLs: []string{"https://raw.githubusercontent.com/commercetools/commercetools-api-reference/main/oas/api/openapi.yaml"}, Confidence: "machine-readable", Basis: "commercetools publishes its platform OpenAPI document in the official API-reference repository; every HTTP path item is extracted."},
	"recharge":      {Kind: "html-reference", URLs: []string{"https://developer.getrecharge.com/2021-11"}, Confidence: "complete-rendered-reference", Basis: "Recharge publishes the complete 2021-11 API reference as one public rendered document; every formal METHOD /path declaration is extracted after stripping presentation markup."},
	"docuseal":      {Kind: "openapi", URLs: []string{"https://console.docuseal.com/openapi.yml"}, Confidence: "machine-readable", Basis: "DocuSeal publishes its OpenAPI document at this official console URL; every HTTP path item is extracted."},
}

func main() {
	connector := flag.String("connector", "", "connector definition name")
	flag.Parse()
	config, ok := configs[*connector]
	if !ok {
		panic(fmt.Sprintf("no complete source configuration for %q", *connector))
	}
	var out result
	var err error
	switch config.Kind {
	case "openapi":
		out, err = extractOpenAPI(*connector, config)
	case "zip-openapi":
		out, err = extractZIPOpenAPI(*connector, config)
	case "readme-reference":
		out, err = extractReadMe(*connector, config)
	case "html-reference":
		out, err = extractHTMLReference(*connector, config)
	default:
		err = fmt.Errorf("unsupported source kind %q", config.Kind)
	}
	if err != nil {
		panic(err)
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		panic(err)
	}
}

func fetch(url string) ([]byte, document, error) {
	var lastErr error
	for attempt := 0; attempt < 3; attempt++ {
		client := &http.Client{Timeout: 30 * time.Second}
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return nil, document{}, err
		}
		req.Header.Set("User-Agent", "Polymetrics-source-lock/2.0 (+https://github.com/polymetrics-ai/cli)")
		resp, err := client.Do(req)
		if err != nil {
			lastErr = err
			continue
		}
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			lastErr = readErr
			continue
		}
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			lastErr = fmt.Errorf("%s: HTTP %d", url, resp.StatusCode)
			if resp.StatusCode < 500 && resp.StatusCode != http.StatusTooManyRequests {
				break
			}
			continue
		}
		sum := sha256.Sum256(data)
		return data, document{SourceURL: resp.Request.URL.String(), SHA256: hex.EncodeToString(sum[:]), Bytes: len(data), ContentType: resp.Header.Get("Content-Type"), RetrievalMethod: "http", Representation: "machine-readable-spec"}, nil
	}
	return nil, document{}, lastErr
}

func extractOpenAPI(connector string, config sourceConfig) (result, error) {
	out := result{Representation: "machine-readable-spec", Confidence: config.Confidence, Basis: config.Basis}
	for documentIndex, url := range config.URLs {
		data, doc, err := fetch(url)
		if err != nil {
			return result{}, err
		}
		out.Documents = append(out.Documents, doc)
		ops, err := operationsFromSpec(connector, data, doc.SourceURL, documentIndex, "", config.PathPrefix)
		if err != nil {
			return result{}, fmt.Errorf("%s: %w", url, err)
		}
		out.Operations = append(out.Operations, ops...)
	}
	if len(out.Documents) > 0 {
		out.SourceURL = out.Documents[0].SourceURL
	}
	return finalize(out)
}

func extractZIPOpenAPI(connector string, config sourceConfig) (result, error) {
	if len(config.URLs) != 1 {
		return result{}, errors.New("ZIP source requires one URL")
	}
	data, doc, err := fetch(config.URLs[0])
	if err != nil {
		return result{}, err
	}
	out := result{SourceURL: doc.SourceURL, Documents: []document{doc}, Representation: "machine-readable-spec", Confidence: config.Confidence, Basis: config.Basis}
	reader, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return result{}, err
	}
	for index, file := range reader.File {
		if !strings.HasSuffix(strings.ToLower(file.Name), ".json") && !strings.HasSuffix(strings.ToLower(file.Name), ".yaml") && !strings.HasSuffix(strings.ToLower(file.Name), ".yml") {
			continue
		}
		stream, err := file.Open()
		if err != nil {
			return result{}, err
		}
		contents, err := io.ReadAll(stream)
		stream.Close()
		if err != nil {
			return result{}, err
		}
		ops, err := operationsFromSpec(connector, contents, doc.SourceURL, index, file.Name, config.PathPrefix)
		if err != nil {
			continue
		}
		out.Operations = append(out.Operations, ops...)
	}
	return finalize(out)
}

func extractReadMe(connector string, config sourceConfig) (result, error) {
	sidebarBytes, sidebarDoc, err := fetch(config.BaseURL + "/sidebar?page_type=reference")
	if err != nil {
		return result{}, err
	}
	sidebarDoc.Representation = "complete-rendered-reference"
	var sidebar any
	if err := json.Unmarshal(sidebarBytes, &sidebar); err != nil {
		return result{}, err
	}
	pages := endpointPages(sidebar)
	out := result{SourceURL: config.LandingURL, Documents: []document{sidebarDoc}, Representation: "complete-rendered-reference", Confidence: config.Confidence, Basis: config.Basis}
	unique := map[string]map[string]any{}
	for _, page := range pages {
		slug := stringValue(page["slug"])
		if slug != "" {
			unique[slug] = page
		}
	}
	slugs := make([]string, 0, len(unique))
	for slug := range unique {
		slugs = append(slugs, slug)
	}
	sort.Strings(slugs)
	type fetchedPage struct {
		doc document
		op  *operation
		err error
	}
	jobs := make(chan string)
	results := make(chan fetchedPage, len(slugs))
	const workers = 4
	var group sync.WaitGroup
	for i := 0; i < workers; i++ {
		group.Add(1)
		go func() {
			defer group.Done()
			for slug := range jobs {
				page := unique[slug]
				data, doc, err := fetch(config.BaseURL + "/reference/" + slug + "?reduce=false")
				if err != nil {
					results <- fetchedPage{err: fmt.Errorf("%s: %w", slug, err)}
					continue
				}
				doc.Representation = "complete-rendered-reference"
				var pageJSON map[string]any
				if err := json.Unmarshal(data, &pageJSON); err != nil {
					results <- fetchedPage{err: fmt.Errorf("%s: %w", slug, err)}
					continue
				}
				api := mapValue(mapValue(pageJSON, "data"), "api")
				method, requestPath := strings.ToUpper(stringValue(api["method"])), stringValue(api["path"])
				var op *operation
				if isMethod(method) && requestPath != "" && !boolValue(page["api_webhook"]) {
					candidate := operation{ID: fmt.Sprintf("%s.rendered.%s", connector, slug), Protocol: "rest", Method: method, Path: requestPath, OperationID: slug, Deprecated: boolValue(page["deprecated"]), SourceLocation: "$data.api", SourceURL: doc.SourceURL}
					op = &candidate
				}
				results <- fetchedPage{doc: doc, op: op}
			}
		}()
	}
	go func() {
		for _, slug := range slugs {
			jobs <- slug
		}
		close(jobs)
		group.Wait()
		close(results)
	}()
	for page := range results {
		if page.err != nil {
			return result{}, page.err
		}
		out.Documents = append(out.Documents, page.doc)
		if page.op != nil {
			out.Operations = append(out.Operations, *page.op)
		}
	}
	sort.Slice(out.Documents[1:], func(i, j int) bool { return out.Documents[1+i].SourceURL < out.Documents[1+j].SourceURL })
	return finalize(out)
}

// renderedOperationRE recognizes both a formal Markdown declaration such as
// "**GET** `/v1/things`" and a rendered URL such as
// "POST https://api.example.test/v1/things". It intentionally requires a
// method immediately followed by a route, which avoids counting prose that
// merely mentions an HTTP verb.
var renderedOperationRE = regexp.MustCompile("(?i)\\b(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)(?:\\*\\*)?\\s+[`\\\"']?(/[^\\s<?`\\\"')]+)")
var renderedAbsoluteOperationRE = regexp.MustCompile("(?i)\\b(GET|POST|PUT|PATCH|DELETE|HEAD|OPTIONS)(?:\\*\\*)?\\s+[`\\\"']?https?://[^/\\s`\\\"']+(/[^\\s<?`\\\"')]+)")
var renderedColonParameterRE = regexp.MustCompile(`:([A-Za-z][A-Za-z0-9_-]*)`)
var renderedHTMLTagRE = regexp.MustCompile(`(?s)<[^>]+>`)

// extractHTMLReference reads one provider-owned rendered reference document.
// The regular expression is deliberately limited to its formal METHOD /path
// declarations; shell examples and query strings are not independent API
// operations. It is used only for providers that publish their entire
// reference on one public page rather than a machine-readable description.
func extractHTMLReference(connector string, config sourceConfig) (result, error) {
	if len(config.URLs) == 0 {
		return result{}, errors.New("HTML reference source requires at least one URL")
	}
	out := result{Representation: "complete-rendered-reference", Confidence: config.Confidence, Basis: config.Basis}
	for _, sourceURL := range config.ReferenceIndexURLs {
		_, doc, err := fetch(sourceURL)
		if err != nil {
			return result{}, err
		}
		doc.Representation = "complete-rendered-reference"
		out.Documents = append(out.Documents, doc)
	}
	for _, sourceURL := range config.URLs {
		data, doc, err := fetch(sourceURL)
		if err != nil {
			return result{}, err
		}
		doc.Representation = "complete-rendered-reference"
		if out.SourceURL == "" {
			out.SourceURL = doc.SourceURL
		}
		out.Documents = append(out.Documents, doc)
		content := renderedHTMLTagRE.ReplaceAllString(html.UnescapeString(string(data)), " ")
		appendRenderedOperations(&out, connector, config, content, doc.SourceURL, renderedOperationRE)
		if config.ExtractAbsolutePaths {
			appendRenderedOperations(&out, connector, config, content, doc.SourceURL, renderedAbsoluteOperationRE)
		}
	}
	sort.Slice(out.Documents, func(i, j int) bool { return out.Documents[i].SourceURL < out.Documents[j].SourceURL })
	return finalize(out)
}

func appendRenderedOperations(out *result, connector string, config sourceConfig, content, sourceURL string, matcher *regexp.Regexp) {
	matches := matcher.FindAllStringSubmatchIndex(content, -1)
	for _, match := range matches {
		// A cURL sample is evidence for the formal declaration on the page, not
		// a second operation. In particular, its example identifiers must not
		// become invented API paths in the source inventory.
		lineStart := strings.LastIndex(content[:match[0]], "\n") + 1
		if strings.Contains(strings.ToLower(content[lineStart:match[0]]), "curl") {
			continue
		}
		method := strings.ToUpper(content[match[2]:match[3]])
		requestPath := normalizeRenderedPath(config, content[match[4]:match[5]])
		// A webhook receiver URL is documentation for the consumer's callback,
		// not an operation at api.lever.co/v1.
		if requestPath == "/yourWebhookUrl" || requestPath == "" {
			continue
		}
		out.Operations = append(out.Operations, operation{
			ID:             fmt.Sprintf("%s.rendered.%s", connector, slug(method+" "+requestPath)),
			Protocol:       "rest",
			Method:         method,
			Path:           requestPath,
			OperationID:    slug(method + " " + requestPath),
			SourceLocation: fmt.Sprintf("rendered-document-byte-%d", match[0]),
			SourceURL:      sourceURL,
		})
	}
}

func normalizeRenderedPath(config sourceConfig, requestPath string) string {
	requestPath = strings.TrimRight(strings.TrimSpace(requestPath), ".,;:")
	requestPath = strings.SplitN(requestPath, "?", 2)[0]
	requestPath = renderedColonParameterRE.ReplaceAllString(requestPath, "{$1}")
	if config.PathStripPrefix != "" && strings.HasPrefix(requestPath, config.PathStripPrefix+"/") {
		requestPath = strings.TrimPrefix(requestPath, config.PathStripPrefix)
	}
	if len(requestPath) > 1 {
		requestPath = strings.TrimSuffix(requestPath, "/")
	}
	return requestPath
}

func endpointPages(value any) []map[string]any {
	var found []map[string]any
	var walk func(any)
	walk = func(current any) {
		switch typed := current.(type) {
		case []any:
			for _, child := range typed {
				walk(child)
			}
		case map[string]any:
			if stringValue(typed["type"]) == "endpoint" {
				found = append(found, typed)
			}
			if children, ok := typed["pages"]; ok {
				walk(children)
			}
		}
	}
	walk(value)
	return found
}

func operationsFromSpec(connector string, data []byte, sourceURL string, documentIndex int, entry, pathPrefix string) ([]operation, error) {
	var root map[string]any
	if err := json.Unmarshal(data, &root); err != nil {
		if err := yaml.Unmarshal(data, &root); err != nil {
			return nil, err
		}
	}
	paths := mapValue(root, "paths")
	if len(paths) == 0 {
		return nil, errors.New("document has no OpenAPI paths")
	}
	pathNames := sortedKeys(paths)
	var output []operation
	for _, sourcePath := range pathNames {
		requestPath := prefixedPath(pathPrefix, sourcePath)
		pathItem := mapValue(paths, sourcePath)
		for _, method := range []string{"get", "post", "put", "patch", "delete", "head", "options"} {
			item, ok := pathItem[method]
			if !ok {
				continue
			}
			spec := mapAny(item)
			if len(spec) == 0 {
				continue
			}
			opID := stringValue(spec["operationId"])
			if opID == "" {
				opID = slug(strings.ToUpper(method) + " " + requestPath)
			}
			location := fmt.Sprintf("paths[%q].%s", requestPath, method)
			if entry != "" {
				location = entry + ":" + location
			}
			output = append(output, operation{ID: fmt.Sprintf("%s.rest.%s-source-%d", connector, slug(opID), documentIndex+1), Protocol: "rest", Method: strings.ToUpper(method), Path: requestPath, OperationID: opID, Deprecated: boolValue(spec["deprecated"]), SourceLocation: location, SourceURL: sourceURL})
		}
	}
	return output, nil
}

func prefixedPath(prefix, requestPath string) string {
	prefix = strings.TrimSuffix(strings.TrimSpace(prefix), "/")
	if prefix == "" || strings.HasPrefix(requestPath, prefix+"/") || requestPath == prefix {
		return requestPath
	}
	return prefix + "/" + strings.TrimPrefix(requestPath, "/")
}

func finalize(out result) (result, error) {
	if len(out.Operations) == 0 {
		return result{}, errors.New("source extraction found zero REST operations")
	}
	// api_surface.json is a method+path ledger. A rendered reference can repeat
	// a declaration while showing parameter or response examples, but that is
	// still one provider operation and must not inflate the source count.
	// First remove concrete sample routes whenever the same documentation set
	// declares the parameterized route for that method. This is intentionally
	// structural rather than provider-specific: a formal /things/{id} route
	// owns /things/123 in a code sample, but it never owns a different static
	// endpoint with a different segment count or literal layout.
	parameterized := map[string][]operation{}
	for _, candidate := range out.Operations {
		if strings.Contains(candidate.Path, "{") {
			parameterized[candidate.Method] = append(parameterized[candidate.Method], candidate)
		}
	}
	filtered := make([]operation, 0, len(out.Operations))
	for _, candidate := range out.Operations {
		if strings.HasPrefix(candidate.SourceLocation, "rendered-document-byte-") && !strings.Contains(candidate.Path, "{") && matchesParameterizedRoute(candidate.Path, parameterized[candidate.Method]) {
			continue
		}
		filtered = append(filtered, candidate)
	}
	out.Operations = filtered

	byEndpoint := map[string]operation{}
	for _, candidate := range out.Operations {
		key := candidate.Method + " " + canonicalEndpointPath(candidate.Path)
		if _, seen := byEndpoint[key]; !seen {
			byEndpoint[key] = candidate
		}
	}
	out.Operations = make([]operation, 0, len(byEndpoint))
	for _, candidate := range byEndpoint {
		out.Operations = append(out.Operations, candidate)
	}
	seen := map[string]int{}
	for i := range out.Operations {
		base := out.Operations[i].ID
		seen[base]++
		if seen[base] > 1 {
			out.Operations[i].ID = fmt.Sprintf("%s-duplicate-%d", base, seen[base])
		}
	}
	sort.Slice(out.Operations, func(i, j int) bool { return out.Operations[i].ID < out.Operations[j].ID })
	return out, nil
}

func matchesParameterizedRoute(requestPath string, candidates []operation) bool {
	segments := strings.Split(strings.TrimPrefix(requestPath, "/"), "/")
	for _, candidate := range candidates {
		other := strings.Split(strings.TrimPrefix(candidate.Path, "/"), "/")
		if len(segments) != len(other) {
			continue
		}
		matched := true
		for index, segment := range other {
			if strings.HasPrefix(segment, "{") && strings.HasSuffix(segment, "}") {
				continue
			}
			if segments[index] != segment {
				matched = false
				break
			}
		}
		if matched {
			return true
		}
	}
	return false
}

var endpointParameterRE = regexp.MustCompile(`\{[^}]+\}`)

// canonicalEndpointPath is used only to recognize a rendered documentation
// page spelling the same route with different placeholder labels
// (`{posting}` versus `{postingId}`). The first provider spelling is retained
// as the actual api_surface path.
func canonicalEndpointPath(requestPath string) string {
	return endpointParameterRE.ReplaceAllString(requestPath, "{}")
}

func mapAny(value any) map[string]any                          { mapped, _ := value.(map[string]any); return mapped }
func mapValue(value map[string]any, key string) map[string]any { return mapAny(value[key]) }
func stringValue(value any) string                             { text, _ := value.(string); return text }
func boolValue(value any) bool                                 { truth, _ := value.(bool); return truth }
func isMethod(method string) bool {
	return map[string]bool{"GET": true, "POST": true, "PUT": true, "PATCH": true, "DELETE": true, "HEAD": true, "OPTIONS": true}[method]
}
func sortedKeys(values map[string]any) []string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}
func slug(value string) string {
	value = strings.ToLower(value)
	var b strings.Builder
	dash := false
	for _, r := range value {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
			dash = false
		} else if !dash {
			b.WriteByte('-')
			dash = true
		}
	}
	return strings.Trim(b.String(), "-")
}

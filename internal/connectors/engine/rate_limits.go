package engine

import (
	"fmt"
	"io/fs"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"polymetrics.ai/internal/connectors/connsdk"
)

var rateLimitScopeConfigKeyPattern = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// loadRateLimits loads the optional, provider-cited rate_limits.json contract.
// Absence deliberately remains valid during the fleet migration: an omitted
// file says no new declaration has been authored, rather than inventing an
// uncited provider limit from the legacy metadata field.
func loadRateLimits(sub fs.FS, dirName string, spec *Schema) (*connsdk.RateLimits, error) {
	exists, err := bundleFilePresent(sub, "rate_limits.json")
	if err != nil {
		return nil, err
	}
	if !exists {
		return nil, nil
	}
	raw, err := readFile(sub, "rate_limits.json")
	if err != nil {
		return nil, fmt.Errorf("load bundle %s: %w", dirName, err)
	}
	if err := metaSchemas.rateLimits.Validate(mustDecodeAny(raw)); err != nil {
		return nil, fmt.Errorf("load bundle %s: rate_limits.json: %w", dirName, err)
	}

	var rateLimits connsdk.RateLimits
	if err := strictDecode(raw, &rateLimits); err != nil {
		return nil, fmt.Errorf("load bundle %s: rate_limits.json: %w", dirName, err)
	}
	if err := validateRateLimits(rateLimits, spec); err != nil {
		return nil, fmt.Errorf("load bundle %s: rate_limits.json: %w", dirName, err)
	}
	return &rateLimits, nil
}

func validateRateLimits(declaration connsdk.RateLimits, spec *Schema) error {
	if declaration.SchemaVersion != 1 {
		return diagnosticAt("/schema_version", "schema_version_invalid", "schema_version must be 1", fmt.Errorf("schema_version must be 1"))
	}

	switch declaration.State {
	case connsdk.RateLimitStateDeclared:
		if len(declaration.Policies) == 0 {
			return diagnosticAt("/policies", "policies_required", "declared state requires at least one policy", fmt.Errorf("state declared requires at least one policy"))
		}
	case connsdk.RateLimitStateUnknown, connsdk.RateLimitStateNotApplicable:
		if strings.TrimSpace(declaration.Reason) == "" {
			return diagnosticAt("/reason", "reason_required", "state requires a nonblank reason", fmt.Errorf("state %s requires a nonblank reason", declaration.State))
		}
		if len(declaration.Policies) != 0 {
			return diagnosticAt("/policies", "policies_forbidden", "undeclared state must not declare policies", fmt.Errorf("state %s must not declare policies", declaration.State))
		}
	default:
		return diagnosticAt("/state", "state_invalid", "state must be declared, unknown, or not_applicable", fmt.Errorf("state must be declared, unknown, or not_applicable"))
	}

	seenPolicies := make(map[string]bool, len(declaration.Policies))
	costHeaders := make([]string, len(declaration.Policies))
	for i, policy := range declaration.Policies {
		if !namePattern.MatchString(policy.ID) {
			return diagnosticAt(fmt.Sprintf("/policies/%d/id", i), "policy_id_invalid", "policy id does not match the required format", fmt.Errorf("policies[%d].id %q does not match %s", i, policy.ID, namePattern.String()))
		}
		if seenPolicies[policy.ID] {
			return diagnosticAt(fmt.Sprintf("/policies/%d/id", i), "policy_id_duplicate", "policy id is duplicated", fmt.Errorf("policies[%d].id %q is duplicated", i, policy.ID))
		}
		seenPolicies[policy.ID] = true
		if err := validateRateLimitPolicy(policy, spec); err != nil {
			return diagnosticWithin(fmt.Sprintf("/policies/%d", i), fmt.Errorf("policies[%d]: %w", i, err))
		}
		header, err := rateLimitCostHeader(policy)
		if err != nil {
			return diagnosticWithin(fmt.Sprintf("/policies/%d", i), fmt.Errorf("policies[%d]: %w", i, err))
		}
		costHeaders[i] = header
	}
	return validateRateLimitCostHeaderConflicts(declaration.Policies, costHeaders)
}

func validateRateLimitPolicy(policy connsdk.RateLimitPolicy, spec *Schema) error {
	if err := validateRateLimitSource(policy.Source); err != nil {
		return diagnosticWithin("/source", fmt.Errorf("source: %w", err))
	}
	if err := validateRateLimitSelector(policy.Selector); err != nil {
		return diagnosticWithin("/selector", fmt.Errorf("selector: %w", err))
	}
	if err := validateRateLimitScope(policy.Scope, spec); err != nil {
		return err
	}
	if policy.Coordination != "" && policy.Coordination != connsdk.RateLimitCoordinationRequireShared {
		return diagnosticAt("/coordination", "coordination_invalid", "coordination must be require_shared when present", fmt.Errorf("coordination must be require_shared when present"))
	}
	if len(policy.Budgets) == 0 {
		return diagnosticAt("/budgets", "budgets_required", "budgets must contain at least one provider budget", fmt.Errorf("budgets must contain at least one provider budget"))
	}

	seenBudgets := make(map[string]bool, len(policy.Budgets))
	for i, budget := range policy.Budgets {
		key := strings.Join([]string{string(budget.Model), string(budget.Dimension), string(budget.Unit)}, "/")
		if seenBudgets[key] {
			return diagnosticAt(fmt.Sprintf("/budgets/%d", i), "budget_duplicate", "budget duplicates model, dimension, and unit", fmt.Errorf("budgets[%d] duplicates model/dimension/unit %q", i, key))
		}
		seenBudgets[key] = true
		if err := validateRateLimitBudget(budget); err != nil {
			return diagnosticWithin(fmt.Sprintf("/budgets/%d", i), fmt.Errorf("budgets[%d]: %w", i, err))
		}
	}
	return nil
}

func validateRateLimitSource(source connsdk.RateLimitSource) error {
	rawURL := strings.TrimSpace(source.URL)
	parsed, err := url.ParseRequestURI(rawURL)
	if err != nil || parsed.Scheme != "https" || parsed.Host == "" || parsed.User != nil || parsed.ForceQuery || parsed.RawQuery != "" {
		return diagnosticAt("/url", "source_url_invalid", "source URL must be absolute HTTPS without userinfo or query parameters", fmt.Errorf("url must be an absolute https provider artifact URL without userinfo or query parameters"))
	}
	_, rawFragment, hasFragment := strings.Cut(rawURL, "#")
	if hasFragment && hasCredentialLikeRateLimitFragment(rawFragment) {
		return diagnosticAt("/url", "source_url_fragment_invalid", "source URL must not carry credential-like fragment parameters", fmt.Errorf("url must not carry credential-like fragment parameters"))
	}
	if _, err := time.Parse(time.DateOnly, strings.TrimSpace(source.RetrievedAt)); err != nil {
		return diagnosticAt("/retrieved_at", "source_date_invalid", "retrieved_at must be an ISO date", fmt.Errorf("retrieved_at must be an ISO date: %w", err))
	}
	return nil
}

var credentialLikeRateLimitFragmentKeys = map[string]struct{}{
	"accesskey":     {},
	"accesstoken":   {},
	"apikey":        {},
	"apitoken":      {},
	"authorization": {},
	"authtoken":     {},
	"bearertoken":   {},
	"clientsecret":  {},
	"credential":    {},
	"credentials":   {},
	"idtoken":       {},
	"key":           {},
	"password":      {},
	"privatekey":    {},
	"refreshtoken":  {},
	"secret":        {},
	"secretkey":     {},
	"sig":           {},
	"signature":     {},
	"token":         {},
}

func hasCredentialLikeRateLimitFragment(rawFragment string) bool {
	fragment, err := url.PathUnescape(rawFragment)
	if err != nil {
		return true
	}
	for _, part := range strings.FieldsFunc(fragment, func(r rune) bool {
		return r == '&' || r == ';' || r == '?'
	}) {
		key, _, hasValue := strings.Cut(part, "=")
		if !hasValue {
			continue
		}
		key, err = url.PathUnescape(key)
		if err != nil {
			return true
		}
		key = strings.NewReplacer("_", "", "-", "", ".", "").Replace(strings.ToLower(strings.TrimSpace(key)))
		if _, ok := credentialLikeRateLimitFragmentKeys[key]; ok {
			return true
		}
	}
	return false
}

func validateRateLimitSelector(selector connsdk.RateLimitSelector) error {
	if selector.All {
		if len(selector.Endpoints) != 0 || len(selector.ExcludeEndpoints) != 0 || len(selector.Tiers) != 0 || len(selector.AuthTypes) != 0 {
			return diagnosticAt("", "selector_conflict", "all cannot be combined with endpoint, tier, or auth selectors", fmt.Errorf("all cannot be combined with endpoints, exclude_endpoints, tiers, or auth_types"))
		}
		return nil
	}
	if len(selector.Endpoints) == 0 && len(selector.Tiers) == 0 && len(selector.AuthTypes) == 0 {
		return diagnosticAt("", "selector_empty", "selector must select all or an endpoint, tier, or auth type", fmt.Errorf("must select all or at least one endpoint, tier, or auth type"))
	}
	if err := validateRateLimitEndpoints("endpoints", selector.Endpoints); err != nil {
		return err
	}
	if err := validateRateLimitEndpoints("exclude_endpoints", selector.ExcludeEndpoints); err != nil {
		return err
	}
	if err := validateRateLimitNames("tiers", selector.Tiers); err != nil {
		return err
	}
	return validateRateLimitNames("auth_types", selector.AuthTypes)
}

func validateRateLimitEndpoints(field string, endpoints []connsdk.RateLimitEndpointSelector) error {
	for i, endpoint := range endpoints {
		if !validRateLimitMethod(endpoint.Method) {
			return diagnosticAt(fmt.Sprintf("/%s/%d/method", field, i), "selector_method_invalid", "endpoint method must be an HTTP method", fmt.Errorf("%s[%d].method %q is not an HTTP method", field, i, endpoint.Method))
		}
		path := strings.TrimSpace(endpoint.Path)
		if path != endpoint.Path || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") || strings.ContainsAny(path, "\r\n?#") {
			return diagnosticAt(fmt.Sprintf("/%s/%d/path", field, i), "selector_path_invalid", "endpoint path must be rooted and connector-relative", fmt.Errorf("%s[%d].path must be a rooted connector-relative path", field, i))
		}
	}
	return nil
}

func validateRateLimitNames(field string, values []string) error {
	seen := make(map[string]bool, len(values))
	for i, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			return diagnosticAt(fmt.Sprintf("/%s/%d", field, i), "selector_value_blank", "selector value must not be blank", fmt.Errorf("%s[%d] must not be blank", field, i))
		}
		if seen[trimmed] {
			return diagnosticAt(fmt.Sprintf("/%s/%d", field, i), "selector_value_duplicate", "selector value is duplicated", fmt.Errorf("%s[%d] %q is duplicated", field, i, trimmed))
		}
		seen[trimmed] = true
	}
	return nil
}

func validRateLimitMethod(method string) bool {
	switch method {
	case http.MethodConnect, http.MethodDelete, http.MethodGet, http.MethodHead, http.MethodOptions,
		http.MethodPatch, http.MethodPost, http.MethodPut, http.MethodTrace:
		return true
	default:
		return false
	}
}

func validRateLimitScopeSubject(subject connsdk.RateLimitScopeSubjectKind) bool {
	switch subject {
	case connsdk.RateLimitScopeAccount,
		connsdk.RateLimitScopeInstallation,
		connsdk.RateLimitScopeApplication,
		connsdk.RateLimitScopeEndpoint,
		connsdk.RateLimitScopeIP:
		return true
	default:
		return false
	}
}

func validateRateLimitScope(scope connsdk.RateLimitScope, spec *Schema) error {
	if !validRateLimitScopeSubject(scope.SubjectKind) {
		return diagnosticAt("/scope/subject_kind", "scope_kind_invalid", "subject_kind must identify a supported non-secret subject", fmt.Errorf("scope.subject_kind %q is not a supported non-secret subject", scope.SubjectKind))
	}
	if !rateLimitScopeConfigKeyPattern.MatchString(scope.SubjectConfig) {
		return diagnosticAt("/scope/subject_config", "scope_property_invalid", "subject_config must name a non-secret config property", fmt.Errorf("scope.subject_config %q must name a non-secret config property", scope.SubjectConfig))
	}
	if spec == nil {
		return diagnosticAt("/scope/subject_config", "scope_spec_missing", "subject_config cannot be verified without spec.json", fmt.Errorf("scope.subject_config %q cannot be verified without spec.json", scope.SubjectConfig))
	}
	propertyFound := false
	for _, property := range spec.Properties() {
		if property == scope.SubjectConfig {
			propertyFound = true
			break
		}
	}
	if !propertyFound {
		return diagnosticAt("/scope/subject_config", "scope_property_absent", "subject_config must name a spec.json property", fmt.Errorf("scope.subject_config %q must name a spec.json property", scope.SubjectConfig))
	}
	for _, secret := range spec.SecretKeys() {
		if secret == scope.SubjectConfig {
			return diagnosticAt("/scope/subject_config", "scope_property_secret", "subject_config must name a non-secret spec.json property", fmt.Errorf("scope.subject_config %q must name a non-secret spec.json property", scope.SubjectConfig))
		}
	}
	return nil
}

func validateRateLimitBudget(budget connsdk.RateLimitBudget) error {
	if budget.Dimension != connsdk.RateLimitBudgetBurst && budget.Dimension != connsdk.RateLimitBudgetSustained {
		return diagnosticAt("/dimension", "budget_dimension_invalid", "dimension must be burst or sustained", fmt.Errorf("dimension must be burst or sustained"))
	}
	if budget.Unit != connsdk.RateLimitBudgetRequests && budget.Unit != connsdk.RateLimitBudgetPoints {
		return diagnosticAt("/unit", "budget_unit_invalid", "unit must be requests or points", fmt.Errorf("unit must be requests or points"))
	}

	switch budget.Model {
	case connsdk.RateLimitBudgetFixedWindow, connsdk.RateLimitBudgetSlidingWindow:
		if err := requirePositiveRateLimitInt("limit", budget.Limit); err != nil {
			return err
		}
		if err := requirePositiveRateLimitInt("window_seconds", budget.WindowSeconds); err != nil {
			return err
		}
		if budget.Capacity != nil || budget.RestorePerSecond != nil {
			return diagnosticAt("", "budget_fields_conflict", "window budget must not declare capacity or restore_per_second", fmt.Errorf("%s must not declare capacity or restore_per_second", budget.Model))
		}
	case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
		if err := requirePositiveRateLimitInt("capacity", budget.Capacity); err != nil {
			return err
		}
		if err := requirePositiveRateLimitFloat("restore_per_second", budget.RestorePerSecond); err != nil {
			return err
		}
		if budget.Limit != nil || budget.WindowSeconds != nil {
			return diagnosticAt("", "budget_fields_conflict", "bucket budget must not declare limit or window_seconds", fmt.Errorf("%s must not declare limit or window_seconds", budget.Model))
		}
	default:
		return diagnosticAt("/model", "budget_model_invalid", "model must be a supported window or bucket model", fmt.Errorf("model must be fixed_window, sliding_window, token_bucket, or leaky_bucket"))
	}
	if err := validateRateLimitCost(budget.Cost); err != nil {
		return err
	}
	if budget.Unit != connsdk.RateLimitBudgetPoints || budget.Cost == nil || budget.Cost.DefaultCost == nil {
		return nil
	}
	capacity := 0
	switch budget.Model {
	case connsdk.RateLimitBudgetFixedWindow, connsdk.RateLimitBudgetSlidingWindow:
		capacity = *budget.Limit
	case connsdk.RateLimitBudgetTokenBucket, connsdk.RateLimitBudgetLeakyBucket:
		capacity = *budget.Capacity
	}
	if *budget.Cost.DefaultCost > float64(capacity) {
		return diagnosticAt("/cost/default_cost", "cost_exceeds_capacity", "default_cost must not exceed the declared budget capacity", fmt.Errorf("cost.default_cost must not exceed the declared budget capacity"))
	}
	return nil
}

func requirePositiveRateLimitInt(field string, value *int) error {
	if value == nil || *value <= 0 {
		return diagnosticAt("/"+field, "positive_integer_required", "value must be a positive integer", fmt.Errorf("%s must be a positive integer", field))
	}
	return nil
}

func requirePositiveRateLimitFloat(field string, value *float64) error {
	if value == nil || *value <= 0 {
		return diagnosticAt("/"+strings.ReplaceAll(field, ".", "/"), "positive_number_required", "value must be a positive number", fmt.Errorf("%s must be a positive number", field))
	}
	return nil
}

func validateRateLimitCost(cost *connsdk.RateLimitCost) error {
	if cost == nil {
		return nil
	}
	if cost.DefaultCost == nil && strings.TrimSpace(cost.ResponseHeader) == "" && strings.TrimSpace(cost.ResponseBody) == "" {
		return diagnosticAt("/cost", "cost_source_required", "cost must declare default_cost, response_header, or response_body", fmt.Errorf("cost must declare default_cost, response_header, or response_body"))
	}
	if err := requirePositiveRateLimitFloat("cost.default_cost", cost.DefaultCost); err != nil && cost.DefaultCost != nil {
		return err
	}
	if cost.ResponseHeader != "" {
		header := strings.TrimSpace(cost.ResponseHeader)
		if header == "" || header != cost.ResponseHeader || !httpHeaderNamePattern.MatchString(header) {
			return diagnosticAt("/cost/response_header", "cost_header_invalid", "cost response_header must be an HTTP field name", fmt.Errorf("cost.response_header must be an HTTP field name"))
		}
	}
	if cost.ResponseBody != "" && cost.ResponseBody != string(connsdk.RateLimitCostSourceGraphQLRateLimit) {
		return diagnosticAt("/cost/response_body", "cost_body_invalid", "cost response_body must be graphql_rate_limit", fmt.Errorf("cost.response_body must be graphql_rate_limit"))
	}
	if cost.ResponseHeader != "" && cost.ResponseBody != "" {
		return diagnosticAt("/cost", "cost_sources_conflict", "cost must not combine response_header and response_body", fmt.Errorf("cost must not combine response_header and response_body"))
	}
	return nil
}

func rateLimitCostHeader(policy connsdk.RateLimitPolicy) (string, error) {
	var header string
	for i, budget := range policy.Budgets {
		if budget.Cost == nil || budget.Cost.ResponseHeader == "" {
			continue
		}
		if header != "" && !strings.EqualFold(header, budget.Cost.ResponseHeader) {
			return "", diagnosticAt(fmt.Sprintf("/budgets/%d/cost/response_header", i), "cost_headers_conflict", "cost response_header must name at most one header per policy", fmt.Errorf("cost.response_header must name at most one header per policy"))
		}
		header = budget.Cost.ResponseHeader
	}
	return header, nil
}

func validateRateLimitCostHeaderConflicts(policies []connsdk.RateLimitPolicy, headers []string) error {
	for i, policy := range policies {
		if headers[i] == "" {
			continue
		}
		for j := i + 1; j < len(policies); j++ {
			if headers[j] == "" || strings.EqualFold(headers[i], headers[j]) || !rateLimitSelectorsOverlap(policy.Selector, policies[j].Selector) {
				continue
			}
			return diagnosticAt(fmt.Sprintf("/policies/%d/budgets", j), "policy_cost_headers_conflict", "overlapping policies must not declare different cost response headers", fmt.Errorf("policies %q and %q can both match a request but declare different cost.response_header values", policy.ID, policies[j].ID))
		}
	}
	return nil
}

func rateLimitSelectorsOverlap(left, right connsdk.RateLimitSelector) bool {
	return rateLimitEndpointSelectorsOverlap(left, right) &&
		rateLimitSelectorValuesOverlap(left.Tiers, right.Tiers) &&
		rateLimitSelectorValuesOverlap(left.AuthTypes, right.AuthTypes)
}

func rateLimitEndpointSelectorsOverlap(left, right connsdk.RateLimitSelector) bool {
	if len(left.Endpoints) == 0 && len(right.Endpoints) == 0 {
		return true
	}
	if len(left.Endpoints) == 0 {
		for _, endpoint := range right.Endpoints {
			if !rateLimitEndpointMatches(left.ExcludeEndpoints, endpoint.Method, endpoint.Path) && !rateLimitEndpointMatches(right.ExcludeEndpoints, endpoint.Method, endpoint.Path) {
				return true
			}
		}
		return false
	}
	if len(right.Endpoints) == 0 {
		for _, endpoint := range left.Endpoints {
			if !rateLimitEndpointMatches(left.ExcludeEndpoints, endpoint.Method, endpoint.Path) && !rateLimitEndpointMatches(right.ExcludeEndpoints, endpoint.Method, endpoint.Path) {
				return true
			}
		}
		return false
	}
	for _, leftEndpoint := range left.Endpoints {
		if rateLimitEndpointMatches(left.ExcludeEndpoints, leftEndpoint.Method, leftEndpoint.Path) {
			continue
		}
		for _, rightEndpoint := range right.Endpoints {
			if leftEndpoint.Method == rightEndpoint.Method && leftEndpoint.Path == rightEndpoint.Path && !rateLimitEndpointMatches(right.ExcludeEndpoints, rightEndpoint.Method, rightEndpoint.Path) {
				return true
			}
		}
	}
	return false
}

func rateLimitSelectorValuesOverlap(left, right []string) bool {
	if len(left) == 0 || len(right) == 0 {
		return true
	}
	for _, leftValue := range left {
		for _, rightValue := range right {
			if leftValue == rightValue {
				return true
			}
		}
	}
	return false
}

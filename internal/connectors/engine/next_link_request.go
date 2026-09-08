package engine

import (
	"fmt"
	"net/url"
	"strings"

	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/connectors/connsdk"
	"polymetrics.ai/internal/safety"
)

// NextURLQuerySpec is a closed source-selected query contract. A nil contract
// uses the complete returned URL without carrying initial query parameters.
// Allowed extends direct cursor admission; Retain may fill only missing keys.
// Neither list grants a new origin, path, credential, or raw request capability.
type NextURLQuerySpec struct {
	Allowed []string `json:"allowed,omitempty"`
	Retain  []string `json:"retain,omitempty"`
}

func validateNextURLQuery(spec PaginationSpec) error {
	if spec.NextURLQuery == nil {
		return nil
	}
	if spec.Type != "next_url" && spec.Type != "link_header" {
		return fmt.Errorf("next_url_query requires URL pagination")
	}
	allowed := map[string]bool{}
	for _, name := range spec.NextURLQuery.Allowed {
		if name == "" || name != strings.TrimSpace(name) {
			return fmt.Errorf("next_url_query has invalid key")
		}
		if err := safety.ValidateIdentifier(name, "next-link query key"); err != nil {
			return fmt.Errorf("next_url_query key: %w", err)
		}
		if allowed[name] {
			return fmt.Errorf("next_url_query repeats allowed key %q", name)
		}
		allowed[name] = true
	}
	positions, _ := pagingParamsForStrategy(spec, spec.Type)
	positions = append(positions, spec.PageParam, spec.CursorParam, spec.OffsetParam, valueOrDefault(spec.CursorParam, "page"))
	retained := map[string]bool{}
	for _, name := range spec.NextURLQuery.Retain {
		if !allowed[name] || retained[name] {
			return fmt.Errorf("next_url_query retained key %q must occur once in allowed", name)
		}
		for _, position := range positions {
			if name == position && name != "" {
				return fmt.Errorf("next_url_query cannot retain navigation position %q", name)
			}
		}
		retained[name] = true
	}
	return nil
}

func validateNextLinkPagination(base *PaginationSpec, streams []StreamSpec) error {
	if base != nil {
		if err := validateNextURLQuery(*base); err != nil {
			return err
		}
	}
	for _, s := range streams {
		if s.Pagination != nil {
			if err := validateNextURLQuery(*s.Pagination); err != nil {
				return fmt.Errorf("stream %q: %w", s.Name, err)
			}
		}
	}
	return nil
}

// nextLinkRequests compares the effective request identity while leaving the
// returned URL's physical query bytes untouched. Ordered duplicate values stay
// ordered; only key order/transport-equivalent escaping normalize for loops.
type nextLinkRequests struct {
	spec    PaginationSpec
	base    string
	initial url.Values
	seen    map[string]bool
}

func newNextLinkRequests(spec PaginationSpec, base string, initial url.Values) (*nextLinkRequests, error) {
	if err := validateNextURLQuery(spec); err != nil {
		return nil, err
	}
	return &nextLinkRequests{spec: spec, base: base, initial: initial, seen: map[string]bool{}}, nil
}
func (p *nextLinkRequests) active() bool {
	return p.spec.Type == "next_url" || p.spec.Type == "link_header"
}
func (p *nextLinkRequests) compose(raw string) (string, error) {
	if err := connectors.ValidateDirectReadPageCursor(raw); err != nil {
		return "", err
	}
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || (u.Scheme != "http" && u.Scheme != "https") || u.User != nil || u.Fragment != "" {
		return "", fmt.Errorf("invalid next-link URL")
	}
	scheme, host := requesterOrigin(p.base)
	if err := checkOrigin(raw, baseOrigin{scheme: scheme, host: host}, p.spec.AllowCrossHost); err != nil {
		return "", err
	}
	query, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", fmt.Errorf("invalid next-link query: %w", err)
	}
	extra := url.Values{}
	if p.spec.NextURLQuery != nil {
		for _, name := range p.spec.NextURLQuery.Retain {
			if _, exists := query[name]; !exists {
				if values, ok := p.initial[name]; ok {
					extra[name] = append([]string(nil), values...)
				}
			}
		}
	}
	if encoded := extra.Encode(); encoded != "" {
		separator := "?"
		if strings.Contains(raw, "?") {
			separator = "&"
			if strings.HasSuffix(raw, "?") || strings.HasSuffix(raw, "&") {
				separator = ""
			}
		}
		raw += separator + encoded
	}
	if err := connectors.ValidateDirectReadPageCursor(raw); err != nil {
		return "", err
	}
	return raw, nil
}
func nextLinkIdentity(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", err
	}
	q, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		return "", err
	}
	u.Scheme = strings.ToLower(u.Scheme)
	u.Host = strings.ToLower(u.Host)
	u.RawQuery = q.Encode()
	u.ForceQuery = false
	return u.String(), nil
}
func (p *nextLinkRequests) request(path string, query url.Values, continuation bool) (string, url.Values, error) {
	if !p.active() {
		return path, query, nil
	}
	effective := ""
	if continuation {
		var err error
		effective, err = p.compose(path)
		if err != nil {
			return "", nil, err
		}
		path, query = effective, nil
	} else {
		u, err := directReadRequestTarget(p.base, path)
		if err != nil {
			return "", nil, err
		}
		if len(query) > 0 {
			q := u.Query()
			for name, values := range query {
				q[name] = values
			}
			u.RawQuery = q.Encode()
		}
		effective = u.String()
	}
	key, err := nextLinkIdentity(effective)
	if err != nil {
		return "", nil, err
	}
	if p.seen[key] {
		return "", nil, fmt.Errorf("next-link loop detected for effective request")
	}
	p.seen[key] = true
	return path, query, nil
}
func (p *nextLinkRequests) continuation(next *connsdk.NextPage) (*connsdk.NextPage, error) {
	if !p.active() || next == nil || next.URL == "" {
		return next, nil
	}
	raw, err := p.compose(next.URL)
	if err != nil {
		return nil, err
	}
	key, err := nextLinkIdentity(raw)
	if err != nil {
		return nil, err
	}
	if p.seen[key] {
		return nil, fmt.Errorf("next-link loop detected for effective request")
	}
	return &connsdk.NextPage{URL: raw}, nil
}

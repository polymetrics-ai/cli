package boundary

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type ledger struct {
	Exceptions []Exception `json:"exceptions"`
}

// Exception is one exact exception ledger row. It is intentionally narrow:
// a row binds one rule, connector, path, and match string, and bounds the
// number of live matches allowed for that tuple.
type Exception struct {
	ID                string `json:"id"`
	Rule              string `json:"rule"`
	Connector         string `json:"connector"`
	Path              string `json:"path"`
	Match             string `json:"match"`
	Reason            string `json:"reason"`
	MigrationIssueURL string `json:"migration_issue_url"`
	Owner             string `json:"owner"`
	ExpiresOn         string `json:"expires_on"`
	MaxMatches        int    `json:"max_matches"`
}

func loadLedger(root, relPath string) (ledger, bool, error) {
	path := filepath.Join(root, filepath.FromSlash(relPath))
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ledger{}, false, nil
		}
		return ledger{}, false, &ConfigError{Err: fmt.Errorf("read boundary exceptions: %w", err)}
	}
	var l ledger
	if err := json.Unmarshal(b, &l); err != nil {
		return ledger{}, false, &ConfigError{Err: fmt.Errorf("parse boundary exceptions: %w", err)}
	}
	for i := range l.Exceptions {
		l.Exceptions[i].Path = normalizeRelPath(l.Exceptions[i].Path)
		l.Exceptions[i].Connector = strings.ToLower(strings.TrimSpace(l.Exceptions[i].Connector))
	}
	if err := validateLedger(l); err != nil {
		return ledger{}, false, &ConfigError{Err: err}
	}
	return l, true, nil
}

func validateLedger(l ledger) error {
	ids := map[string]bool{}
	for i, ex := range l.Exceptions {
		prefix := fmt.Sprintf("boundary exception %d", i+1)
		switch {
		case strings.TrimSpace(ex.ID) == "":
			return fmt.Errorf("%s: id is required", prefix)
		case ids[ex.ID]:
			return fmt.Errorf("%s: duplicate id %q", prefix, ex.ID)
		case strings.TrimSpace(ex.Rule) == "":
			return fmt.Errorf("%s: rule is required", prefix)
		case strings.TrimSpace(ex.Connector) == "":
			return fmt.Errorf("%s: connector is required", prefix)
		case strings.TrimSpace(ex.Path) == "":
			return fmt.Errorf("%s: path is required", prefix)
		case strings.TrimSpace(ex.Match) == "":
			return fmt.Errorf("%s: match is required", prefix)
		case strings.TrimSpace(ex.Reason) == "":
			return fmt.Errorf("%s: reason is required", prefix)
		case strings.TrimSpace(ex.MigrationIssueURL) == "":
			return fmt.Errorf("%s: migration_issue_url is required", prefix)
		case strings.TrimSpace(ex.Owner) == "":
			return fmt.Errorf("%s: owner is required", prefix)
		case ex.MaxMatches <= 0:
			return fmt.Errorf("%s: max_matches must be positive", prefix)
		}
		if _, err := parseExpiry(ex.ExpiresOn); err != nil {
			return fmt.Errorf("%s: expires_on: %w", prefix, err)
		}
		ids[ex.ID] = true
	}
	return nil
}

func applyExceptions(findings []Finding, l ledger, now time.Time, ledgerPath string) ([]Finding, []AppliedException) {
	if len(l.Exceptions) == 0 {
		return findings, []AppliedException{}
	}

	exceptions := append([]Exception(nil), l.Exceptions...)
	sort.Slice(exceptions, func(i, j int) bool { return exceptions[i].ID < exceptions[j].ID })

	suppressed := make([]bool, len(findings))
	var exceptionFindings []Finding
	applied := []AppliedException{}
	for _, ex := range exceptions {
		indices := matchingFindingIndices(findings, ex)
		expiry, _ := parseExpiry(ex.ExpiresOn)
		expired := isExpired(expiry, now)
		if len(indices) == 0 {
			exceptionFindings = append(exceptionFindings, exceptionFinding(RuleExceptionStale, ex, ledgerPath, 0, "exception no longer matches any boundary finding"))
			continue
		}
		if expired {
			exceptionFindings = append(exceptionFindings, exceptionFinding(RuleExceptionExpired, ex, ledgerPath, len(indices), "exception has expired"))
			continue
		}
		if len(indices) > ex.MaxMatches {
			exceptionFindings = append(exceptionFindings, exceptionFinding(RuleExceptionBroadened, ex, ledgerPath, len(indices), fmt.Sprintf("exception matches %d finding(s), max_matches is %d", len(indices), ex.MaxMatches)))
			continue
		}
		for _, idx := range indices {
			suppressed[idx] = true
		}
		applied = append(applied, AppliedException{
			ID:        ex.ID,
			Rule:      ex.Rule,
			Connector: ex.Connector,
			Path:      ex.Path,
			Match:     ex.Match,
			Matches:   len(indices),
			IssueURL:  ex.MigrationIssueURL,
			Owner:     ex.Owner,
			ExpiresOn: ex.ExpiresOn,
		})
	}

	filtered := make([]Finding, 0, len(findings)+len(exceptionFindings))
	for i, finding := range findings {
		if !suppressed[i] {
			filtered = append(filtered, finding)
		}
	}
	filtered = append(filtered, exceptionFindings...)
	sortFindings(filtered)
	return filtered, applied
}

func matchingFindingIndices(findings []Finding, ex Exception) []int {
	var indices []int
	for i, finding := range findings {
		if finding.Rule == ex.Rule && finding.Connector == ex.Connector && finding.Path == ex.Path && finding.Match == ex.Match {
			indices = append(indices, i)
		}
	}
	return indices
}

func suppressAppliedExceptions(findings []Finding, applied []AppliedException) []Finding {
	if len(findings) == 0 || len(applied) == 0 {
		return findings
	}
	suppressed := map[string]bool{}
	for _, ex := range applied {
		suppressed[exceptionMatchKey(ex.Rule, ex.Connector, ex.Path, ex.Match)] = true
	}
	filtered := make([]Finding, 0, len(findings))
	for _, finding := range findings {
		if suppressed[exceptionMatchKey(finding.Rule, finding.Connector, finding.Path, finding.Match)] {
			continue
		}
		filtered = append(filtered, finding)
	}
	sortFindings(filtered)
	return filtered
}

func exceptionMatchKey(rule, connector, path, match string) string {
	return rule + "\x00" + connector + "\x00" + path + "\x00" + match
}

func exceptionFinding(rule string, ex Exception, ledgerPath string, matches int, message string) Finding {
	if matches > 0 {
		message = fmt.Sprintf("%s: %s", message, ex.ID)
	} else {
		message = fmt.Sprintf("%s: %s", message, ex.ID)
	}
	return Finding{
		Rule:        rule,
		Severity:    SeverityError,
		Connector:   ex.Connector,
		Path:        ledgerPath,
		Match:       ex.Match,
		Message:     message,
		Remediation: "remove the exception, migrate the provider-specific behavior to the connector definition, or update the bounded exception with a fresh migration issue",
	}
}

func parseExpiry(value string) (time.Time, error) {
	if strings.TrimSpace(value) == "" {
		return time.Time{}, fmt.Errorf("date is required")
	}
	return time.Parse("2006-01-02", value)
}

func isExpired(expiry, now time.Time) bool {
	if now.IsZero() {
		now = time.Now().UTC()
	}
	y, m, d := now.UTC().Date()
	today := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	return expiry.Before(today)
}

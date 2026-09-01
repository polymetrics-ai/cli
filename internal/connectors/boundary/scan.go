package boundary

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

// Scan scans root for connector-specific shared production Go policy.
func Scan(root string, opts Options) (Report, error) {
	absRoot, err := validateRoot(root)
	if err != nil {
		return Report{}, err
	}
	if opts.ExceptionsPath == "" {
		opts.ExceptionsPath = DefaultExceptionsPath
	}
	if opts.Now.IsZero() {
		opts.Now = time.Now().UTC()
	}

	lx, err := loadLexicon(absRoot)
	if err != nil {
		return Report{}, &ConfigError{Err: err}
	}
	mode := ModeWholeTree
	var limit map[string]bool
	if opts.BaseRef != "" {
		mode = ModeBaseDiff
		limit, err = diffLimit(absRoot, opts.BaseRef)
		if err != nil {
			return Report{}, &ConfigError{Err: err}
		}
	}

	files, err := scanFileList(absRoot, limit, lx)
	if err != nil {
		return Report{}, err
	}
	allFiles, err := scanFileList(absRoot, nil, lx)
	if err != nil {
		return Report{}, err
	}

	findings, checked, err := scanFiles(absRoot, files, lx)
	if err != nil {
		return Report{}, err
	}
	allFindings, _, err := scanFiles(absRoot, allFiles, lx)
	if err != nil {
		return Report{}, err
	}

	ledger, _, err := loadLedger(absRoot, opts.ExceptionsPath)
	if err != nil {
		return Report{}, err
	}
	// Exceptions are validated against the whole tree so stale/expired/broadened
	// rows fail even during a base-diff scan. Suppression is then applied to the
	// active scan findings plus exception contract findings.
	allAfterExceptions, applied := applyExceptions(allFindings, ledger, opts.Now, opts.ExceptionsPath)
	exceptionContractFindings := filterExceptionContractFindings(allAfterExceptions)
	findings = suppressAppliedExceptions(findings, applied)
	findings = mergeExceptionContractFindings(findings, exceptionContractFindings)
	sortFindings(findings)

	outcome := OutcomeClean
	if len(findings) > 0 {
		outcome = OutcomePolicyViolations
	}
	return Report{
		APIVersion:       APIVersion,
		Kind:             Kind,
		Outcome:          outcome,
		RepoRoot:         absRoot,
		Mode:             mode,
		BaseRef:          opts.BaseRef,
		CheckedFiles:     checked,
		ConnectorsLoaded: len(lx.connectors),
		Findings:         nonNilFindings(findings),
		Warnings:         []Finding{},
		Exceptions:       nonNilApplied(applied),
	}, nil
}

func validateRoot(root string) (string, error) {
	if strings.TrimSpace(root) == "" {
		return "", &ConfigError{Err: fmt.Errorf("repo root is required")}
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return "", &ConfigError{Err: fmt.Errorf("resolve repo root: %w", err)}
	}
	info, err := os.Stat(abs)
	if err != nil {
		return "", &ConfigError{Err: fmt.Errorf("stat repo root: %w", err)}
	}
	if !info.IsDir() {
		return "", &ConfigError{Err: fmt.Errorf("repo root is not a directory: %s", root)}
	}
	return abs, nil
}

func diffLimit(root, baseRef string) (map[string]bool, error) {
	if err := validateBaseRef(baseRef); err != nil {
		return nil, err
	}
	changed := map[string]bool{}
	cmd := exec.Command("git", "-C", root, "diff", "--name-only", "--diff-filter=ACMRT", "--end-of-options", baseRef, "--")
	out, err := cmd.Output()
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			return nil, fmt.Errorf("git diff --name-only %s: %s", baseRef, strings.TrimSpace(string(ee.Stderr)))
		}
		return nil, fmt.Errorf("git diff --name-only %s: %w", baseRef, err)
	}
	for _, line := range strings.Split(string(out), "\n") {
		line = normalizeRelPath(strings.TrimSpace(line))
		if line != "" {
			changed[line] = true
		}
	}
	untrackedCmd := exec.Command("git", "-C", root, "ls-files", "--others", "--exclude-standard")
	untrackedOut, err := untrackedCmd.Output()
	if err == nil {
		for _, line := range strings.Split(string(untrackedOut), "\n") {
			line = normalizeRelPath(strings.TrimSpace(line))
			if line != "" {
				changed[line] = true
			}
		}
	}
	return changed, nil
}

func validateBaseRef(baseRef string) error {
	if strings.TrimSpace(baseRef) == "" {
		return fmt.Errorf("base ref is required")
	}
	if strings.HasPrefix(baseRef, "-") {
		return fmt.Errorf("invalid base ref %q: must not start with '-'", baseRef)
	}
	for _, r := range baseRef {
		if r < 0x20 || r == 0x7f {
			return fmt.Errorf("invalid base ref %q: must not contain control characters", baseRef)
		}
	}
	return nil
}

func scanFileList(root string, limit map[string]bool, lx lexicon) ([]string, error) {
	var files []string
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			name := d.Name()
			if name == ".git" || name == "node_modules" || name == "vendor" || name == "dist" || name == "build" || name == ".next" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return err
		}
		rel = normalizeRelPath(rel)
		if limit != nil && !limit[rel] {
			return nil
		}
		pc := classifyPath(rel, lx)
		if pc.ScanGo {
			files = append(files, rel)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("walk repo: %w", err)
	}
	sort.Strings(files)
	return files, nil
}

func scanFiles(root string, files []string, lx lexicon) ([]Finding, int, error) {
	var findings []Finding
	checked := 0
	for _, rel := range files {
		pc := classifyPath(rel, lx)
		if !pc.ScanGo {
			continue
		}
		checked++
		fileFindings, err := scanGoFile(root, rel, pc, lx)
		if err != nil {
			return nil, checked, err
		}
		findings = append(findings, fileFindings...)
	}
	sortFindings(findings)
	return findings, checked, nil
}

func scanGoFile(root, rel string, pc pathClass, lx lexicon) ([]Finding, error) {
	fset := token.NewFileSet()
	abs := filepath.Join(root, filepath.FromSlash(rel))
	file, err := parser.ParseFile(fset, abs, nil, parser.SkipObjectResolution)
	if err != nil {
		return nil, &ConfigError{Err: fmt.Errorf("parse %s: %w", rel, err)}
	}
	literalContexts := map[*ast.BasicLit]literalContext{}
	skipLiterals := map[*ast.BasicLit]bool{}
	for _, spec := range file.Imports {
		skipLiterals[spec.Path] = true
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.SwitchStmt:
			context := literalContextSwitch
			if exprLooksConnectorDiscriminator(n.Tag) {
				context = literalContextConnectorSwitch
			}
			if n.Body != nil {
				for _, stmt := range n.Body.List {
					if clause, ok := stmt.(*ast.CaseClause); ok {
						for _, expr := range clause.List {
							if lit, ok := expr.(*ast.BasicLit); ok && lit.Kind == token.STRING {
								markLiteralContext(literalContexts, lit, context)
							}
						}
					}
				}
			}
		case *ast.BinaryExpr:
			if n.Op == token.EQL || n.Op == token.NEQ {
				if lit, ok := n.X.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					context := literalContextComparison
					if exprLooksConnectorDiscriminator(n.Y) {
						context = literalContextConnectorComparison
					}
					markLiteralContext(literalContexts, lit, context)
				}
				if lit, ok := n.Y.(*ast.BasicLit); ok && lit.Kind == token.STRING {
					context := literalContextComparison
					if exprLooksConnectorDiscriminator(n.X) {
						context = literalContextConnectorComparison
					}
					markLiteralContext(literalContexts, lit, context)
				}
			}
		}
		return true
	})

	scanner := goScanner{fset: fset, path: rel, pathClass: pc, lexicon: lx, literalContexts: literalContexts, skipLiterals: skipLiterals}
	for _, spec := range file.Imports {
		scanner.scanImport(spec)
	}
	ast.Inspect(file, func(node ast.Node) bool {
		switch n := node.(type) {
		case *ast.BasicLit:
			if n.Kind == token.STRING {
				scanner.scanStringLiteral(n)
			}
		case *ast.Ident:
			scanner.scanIdentifier(n)
		}
		return true
	})
	return scanner.findings, nil
}

type goScanner struct {
	fset            *token.FileSet
	path            string
	pathClass       pathClass
	lexicon         lexicon
	findings        []Finding
	literalContexts map[*ast.BasicLit]literalContext
	skipLiterals    map[*ast.BasicLit]bool
}

func (s *goScanner) scanImport(spec *ast.ImportSpec) {
	value, err := strconv.Unquote(spec.Path.Value)
	if err != nil {
		return
	}
	connector, ok := connectorFromImportPath(value, s.lexicon)
	if !ok {
		return
	}
	s.addFinding(RuleConnectorImport, connector, spec.Path.Pos(), value, "connector-specific import in shared production Go", "move connector-specific implementation into internal/connectors/defs/<connector>, hooks/<connector>, native/<connector>, or generated hook/native wiring")
}

func connectorFromImportPath(importPath string, lx lexicon) (string, bool) {
	for _, marker := range []string{"/internal/connectors/hooks/", "/internal/connectors/native/", "/internal/connectors/"} {
		idx := strings.Index(importPath, marker)
		if idx < 0 {
			continue
		}
		rest := importPath[idx+len(marker):]
		name := rest
		if slash := strings.Index(name, "/"); slash >= 0 {
			name = name[:slash]
		}
		name = strings.ToLower(strings.TrimSpace(name))
		if _, ok := lx.byName[name]; ok {
			return name, true
		}
	}
	return "", false
}

func (s *goScanner) scanStringLiteral(lit *ast.BasicLit) {
	if s.skipLiterals[lit] {
		return
	}
	context := s.literalContexts[lit]
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return
	}
	matches := s.lexicon.literalMatches(value, context.isConnectorIdentity(), s.pathClass.DocsOutput)
	if s.pathClass.DocsOutput {
		seenConnectors := map[string]bool{}
		for _, match := range matches {
			if seenConnectors[match.Connector] {
				continue
			}
			seenConnectors[match.Connector] = true
			canonical := literalMatch{Connector: match.Connector, Match: match.Connector, Exact: true}
			rule := s.ruleFor(canonical, context)
			s.addFinding(rule, canonical.Connector, lit.Pos(), canonical.Match, messageForRule(rule, canonical), remediationForRule(rule, canonical))
		}
		return
	}
	for _, match := range matches {
		rule := s.ruleFor(match, context)
		s.addFinding(rule, match.Connector, lit.Pos(), match.Match, messageForRule(rule, match), remediationForRule(rule, match))
	}
}

func (s *goScanner) scanIdentifier(ident *ast.Ident) {
	for _, match := range s.lexicon.identifierMatches(ident.Name) {
		rule := s.ruleFor(match, literalContextIdentifier)
		s.addFinding(rule, match.Connector, ident.Pos(), match.Match, messageForRule(rule, match), remediationForRule(rule, match))
	}
}

func (s *goScanner) ruleFor(match literalMatch, context literalContext) string {
	if s.pathClass.DocsOutput {
		return RuleDocsExample
	}
	if match.Alias {
		return RuleLegacyAlias
	}
	if match.Policy {
		return RuleProviderPolicy
	}
	if context == literalContextSwitch || context == literalContextComparison || context == literalContextConnectorSwitch || context == literalContextConnectorComparison {
		return RuleConnectorSwitch
	}
	return RuleConnectorLiteral
}

func (s *goScanner) addFinding(rule, connector string, pos token.Pos, match, message, remediation string) {
	line := 0
	if pos.IsValid() {
		line = s.fset.Position(pos).Line
	}
	s.findings = append(s.findings, Finding{
		Rule:        rule,
		Severity:    SeverityError,
		Connector:   connector,
		Path:        s.path,
		Line:        line,
		Match:       match,
		Message:     message,
		Remediation: remediation,
	})
}

type literalContext int

const (
	literalContextOther literalContext = iota
	literalContextSwitch
	literalContextComparison
	literalContextConnectorSwitch
	literalContextConnectorComparison
	literalContextIdentifier
)

func (c literalContext) isConnectorIdentity() bool {
	return c == literalContextConnectorSwitch || c == literalContextConnectorComparison
}

func markLiteralContext(contexts map[*ast.BasicLit]literalContext, lit *ast.BasicLit, context literalContext) {
	if contextPriority(context) >= contextPriority(contexts[lit]) {
		contexts[lit] = context
	}
}

func contextPriority(context literalContext) int {
	switch context {
	case literalContextConnectorSwitch, literalContextConnectorComparison:
		return 3
	case literalContextSwitch, literalContextComparison:
		return 2
	case literalContextIdentifier:
		return 1
	default:
		return 0
	}
}

func exprLooksConnectorDiscriminator(expr ast.Expr) bool {
	switch n := expr.(type) {
	case nil:
		return false
	case *ast.Ident:
		return nameLooksConnectorDiscriminator(n.Name)
	case *ast.SelectorExpr:
		return nameLooksConnectorDiscriminator(n.Sel.Name)
	case *ast.ParenExpr:
		return exprLooksConnectorDiscriminator(n.X)
	case *ast.IndexExpr:
		return exprLooksConnectorDiscriminator(n.X) || literalLooksConnectorKey(n.Index)
	case *ast.CallExpr:
		if exprLooksConnectorDiscriminator(n.Fun) {
			return true
		}
		for _, arg := range n.Args {
			if exprLooksConnectorDiscriminator(arg) {
				return true
			}
		}
		return false
	default:
		return false
	}
}

func literalLooksConnectorKey(expr ast.Expr) bool {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return false
	}
	value, err := strconv.Unquote(lit.Value)
	if err != nil {
		return false
	}
	return nameLooksConnectorDiscriminator(value)
}

func nameLooksConnectorDiscriminator(name string) bool {
	lower := strings.ToLower(strings.TrimSpace(name))
	return lower == "connector" || lower == "provider" || strings.HasSuffix(lower, "connector") || strings.HasSuffix(lower, "provider")
}

func messageForRule(rule string, match literalMatch) string {
	switch rule {
	case RuleProviderPolicy:
		return fmt.Sprintf("connector-specific provider policy %q in shared production Go", match.Match)
	case RuleDocsExample:
		return fmt.Sprintf("connector-specific example or resource %q embedded in shared Go docs/help output", match.Match)
	case RuleLegacyAlias:
		return fmt.Sprintf("legacy connector alias %q in shared production Go", match.Match)
	case RuleConnectorImport:
		return "connector-specific import in shared production Go"
	case RuleConnectorSwitch:
		return fmt.Sprintf("connector-specific branch %q in shared production Go", match.Match)
	default:
		return fmt.Sprintf("connector-specific literal %q in shared production Go", match.Match)
	}
}

func remediationForRule(rule string, _ literalMatch) string {
	switch rule {
	case RuleDocsExample:
		return "move provider-specific examples/resources into connector docs or generated definition-owned output"
	case RuleProviderPolicy:
		return "replace provider-prefixed shared policy with a provider-neutral mechanism configured from connector definitions"
	case RuleConnectorImport:
		return "use generated hook/native wiring or keep code under hooks/<connector> or native/<connector>"
	default:
		return "move connector-specific behavior into internal/connectors/defs/<connector> or an approved hook/native escape hatch"
	}
}

func filterExceptionContractFindings(findings []Finding) []Finding {
	var out []Finding
	for _, finding := range findings {
		if strings.HasPrefix(finding.Rule, "exception_") {
			out = append(out, finding)
		}
	}
	return out
}

func mergeExceptionContractFindings(findings, contract []Finding) []Finding {
	seen := map[string]bool{}
	for _, finding := range findings {
		seen[findingKey(finding)] = true
	}
	for _, finding := range contract {
		if !seen[findingKey(finding)] {
			findings = append(findings, finding)
		}
	}
	return findings
}

func findingKey(f Finding) string {
	return f.Rule + "\x00" + f.Connector + "\x00" + f.Path + "\x00" + f.Match + "\x00" + strconv.Itoa(f.Line)
}

func sortFindings(findings []Finding) {
	sort.SliceStable(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		if findings[i].Line != findings[j].Line {
			return findings[i].Line < findings[j].Line
		}
		if findings[i].Rule != findings[j].Rule {
			return findings[i].Rule < findings[j].Rule
		}
		if findings[i].Connector != findings[j].Connector {
			return findings[i].Connector < findings[j].Connector
		}
		return findings[i].Match < findings[j].Match
	})
}

func nonNilFindings(findings []Finding) []Finding {
	if findings == nil {
		return []Finding{}
	}
	return findings
}

func nonNilApplied(exceptions []AppliedException) []AppliedException {
	if exceptions == nil {
		return []AppliedException{}
	}
	return exceptions
}

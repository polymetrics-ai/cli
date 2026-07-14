package gsd

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/polymetrics-ai/cli/agent-runtime/shepherd/internal/contract"
)

type IssueProjectIdentity struct {
	DeliveryID  string `json:"delivery_id"`
	Issue       int    `json:"issue"`
	ParentIssue int    `json:"parent_issue"`
	Branch      string `json:"branch"`
	BaseBranch  string `json:"base_branch"`
	ProjectRoot string `json:"project_root"`
	InitialHead string `json:"initial_head"`
	ContextHash string `json:"context_hash"`
	GSDVersion  string `json:"gsd_version"`
}

func BootstrapIssueProject(root string, identity IssueProjectIdentity, issueContext contract.IssueContext) error {
	if err := validateIssueProjectIdentity(root, identity, issueContext); err != nil {
		return err
	}
	gsdDir := filepath.Join(root, ".gsd")
	if info, err := os.Lstat(gsdDir); err == nil {
		if !info.IsDir() || info.Mode()&os.ModeSymlink != 0 {
			return errors.New("issue GSD project root must be a real directory")
		}
		return adoptIssueProject(gsdDir, identity, issueContext)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	temporary, err := os.MkdirTemp(root, ".gsd-bootstrap-*")
	if err != nil {
		return fmt.Errorf("create issue GSD bootstrap: %w", err)
	}
	defer func() { _ = os.RemoveAll(temporary) }()
	if err := writeIssueProjectFiles(temporary, identity, issueContext); err != nil {
		return err
	}
	if err := os.Rename(temporary, gsdDir); err != nil {
		return fmt.Errorf("adopt issue GSD project: %w", err)
	}
	return syncDirectory(root)
}

func adoptIssueProject(gsdDir string, identity IssueProjectIdentity, issueContext contract.IssueContext) error {
	marker := filepath.Join(gsdDir, "ISSUE.json")
	if raw, err := os.ReadFile(marker); err == nil {
		var existing IssueProjectIdentity
		if json.Unmarshal(raw, &existing) != nil || existing != identity {
			return errors.New("GSD project is already bound to a different issue identity")
		}
	} else if errors.Is(err, os.ErrNotExist) {
		if err := writeIdentity(marker, identity); err != nil {
			return err
		}
	} else {
		return err
	}
	for name, raw := range issueProjectDocuments(identity, issueContext) {
		path := filepath.Join(gsdDir, name)
		if _, err := os.Lstat(path); errors.Is(err, os.ErrNotExist) {
			if err := writeNewFile(path, raw, 0o600); err != nil {
				return err
			}
		} else if err != nil {
			return err
		}
	}
	return syncDirectory(gsdDir)
}

func writeIssueProjectFiles(directory string, identity IssueProjectIdentity, issueContext contract.IssueContext) error {
	if err := writeIdentity(filepath.Join(directory, "ISSUE.json"), identity); err != nil {
		return err
	}
	for name, raw := range issueProjectDocuments(identity, issueContext) {
		if err := writeNewFile(filepath.Join(directory, name), raw, 0o600); err != nil {
			return err
		}
	}
	return syncDirectory(directory)
}

func writeIdentity(path string, identity IssueProjectIdentity) error {
	raw, err := json.MarshalIndent(identity, "", "  ")
	if err != nil {
		return err
	}
	return writeNewFile(path, append(raw, '\n'), 0o600)
}

func writeNewFile(path string, raw []byte, mode os.FileMode) error {
	file, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, mode)
	if err != nil {
		return err
	}
	if _, err := file.Write(raw); err != nil {
		_ = file.Close()
		return err
	}
	return errors.Join(file.Sync(), file.Close())
}

func issueProjectDocuments(identity IssueProjectIdentity, issueContext contract.IssueContext) map[string][]byte {
	var requirements strings.Builder
	requirements.WriteString("# Requirements\n\n")
	for index, criterion := range issueContext.AcceptanceCriteria {
		fmt.Fprintf(&requirements, "- R%03d: %s\n", index+1, criterion)
	}
	project := fmt.Sprintf("# Project\n\nGitHub issue #%d: %s\n\nParent issue: #%d\nBranch: `%s`\nBase: `%s`\n",
		identity.Issue, issueContext.Objective, identity.ParentIssue, identity.Branch, identity.BaseBranch)
	preferences := `---
version: 1
models:
  research: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  planning: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  discuss: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  execution: { provider: openai-codex, model: gpt-5.5, thinking: high }
  execution_simple: { provider: openai-codex, model: gpt-5.5, thinking: high }
  completion: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  validation: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
  subagent: { provider: openai-codex, model: gpt-5.5, thinking: high }
  uat: { provider: openai-codex, model: gpt-5.6-sol, thinking: high }
git:
  auto_push: false
  auto_pr: false
  merge_strategy: squash
---
`
	return map[string][]byte{
		"PROJECT.md": []byte(project), "REQUIREMENTS.md": []byte(requirements.String()),
		"PREFERENCES.md": []byte(preferences),
	}
}

func validateIssueProjectIdentity(root string, identity IssueProjectIdentity, issueContext contract.IssueContext) error {
	if !filepath.IsAbs(root) || filepath.Clean(root) != filepath.Clean(identity.ProjectRoot) ||
		identity.DeliveryID != fmt.Sprintf("issue-%d", identity.Issue) || identity.Issue <= 0 ||
		identity.ParentIssue <= 0 || issueContext.Issue != identity.Issue ||
		issueContext.ParentIssue != identity.ParentIssue || issueContext.Branch != identity.Branch ||
		issueContext.PRBase != identity.BaseBranch || len(identity.InitialHead) != 40 ||
		!strings.HasPrefix(identity.ContextHash, "sha256:") || identity.GSDVersion == "" {
		return errors.New("complete matching issue GSD project identity is required")
	}
	return nil
}

func syncDirectory(path string) error {
	directory, err := os.Open(path)
	if err != nil {
		return err
	}
	return errors.Join(directory.Sync(), directory.Close())
}

func sameBytes(left, right []byte) bool { return bytes.Equal(left, right) }

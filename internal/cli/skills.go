package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors"
)

type skillDoc struct {
	Name        string
	Description string
	Body        string
}

func runSkills(dir string, stdout io.Writer, jsonOut bool) error {
	if dir == "" {
		return validationErrorf("missing --dir")
	}
	registry := appRegistry()
	generated, err := generateSkills(dir, registry.ListManifests())
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "SkillGeneration", "dir": dir, "skills": generated})
	}
	fmt.Fprintf(stdout, "Generated skills in %s\n", dir)
	return nil
}

func generateSkills(dir string, manifests []connectors.Manifest) ([]string, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create skills dir: %w", err)
	}
	docs := baseSkillDocs(manifests)
	names := make([]string, 0, len(docs))
	for _, doc := range docs {
		path := filepath.Join(dir, doc.Name)
		if err := os.MkdirAll(path, 0o755); err != nil {
			return nil, fmt.Errorf("create skill dir %s: %w", doc.Name, err)
		}
		if err := os.WriteFile(filepath.Join(path, "SKILL.md"), []byte(doc.Body), 0o644); err != nil {
			return nil, fmt.Errorf("write skill %s: %w", doc.Name, err)
		}
		names = append(names, doc.Name)
	}
	sort.Strings(names)
	if err := writeSkillIndex(filepath.Join(dir, "skills.md"), docs); err != nil {
		return nil, err
	}
	return names, nil
}

func baseSkillDocs(manifests []connectors.Manifest) []skillDoc {
	docs := []skillDoc{
		{
			Name:        "pm-shared",
			Description: "Shared Polymetrics CLI safety rules and output contracts.",
			Body: skillBody("pm-shared", "Shared Polymetrics CLI safety rules and output contracts.", []string{
				"Use --json for machine-readable output.",
				"Never ask for or print secret values.",
				"Use plan, preview, approval, and execute boundaries for mutations.",
				"Prefer dependency-free commands unless runtime-backed mode is explicitly requested.",
			}),
		},
		{
			Name:        "pm-connectors",
			Description: "Inspect connector manifests and safe capabilities.",
			Body: skillBody("pm-connectors", "Inspect connector manifests and safe capabilities.", []string{
				"Run `pm connectors list --json` before choosing a connector.",
				"Run `pm connectors inspect <name> --json` to read manifest fields.",
				"Connector inspection never reads encrypted credentials.",
			}),
		},
		{
			Name:        "pm-etl",
			Description: "Run bounded ETL syncs from configured connections.",
			Body: skillBody("pm-etl", "Run bounded ETL syncs from configured connections.", []string{
				"Use `pm etl run --connection <name> --stream <stream> --json`.",
				"Use `--batch-size` for large streams when the caller requests bounded memory behavior.",
				"Supported sync modes are `full_refresh_append`, `full_refresh_overwrite`, `full_refresh_overwrite_deduped`, `incremental_append`, and `incremental_append_deduped`.",
				"Incremental modes require a cursor. Deduped modes require a primary key.",
				"Inspect `batch_count` and `checkpoint` in JSON output after runs.",
			}),
		},
		{
			Name:        "pm-reverse-etl",
			Description: "Plan, preview, approve, and execute reverse ETL.",
			Body: skillBody("pm-reverse-etl", "Plan, preview, approve, and execute reverse ETL.", []string{
				"Run `pm reverse plan` before any write.",
				"Run `pm reverse preview <plan-id> --json` before approval.",
				"Run `pm reverse run <plan-id> --approve <token>` only after explicit approval.",
			}),
		},
		{
			Name:        "pm-runtime",
			Description: "Check optional PostgreSQL, DragonflyDB, and Temporal runtime services.",
			Body: skillBody("pm-runtime", "Check optional PostgreSQL, DragonflyDB, and Temporal runtime services.", []string{
				"Run `pm runtime doctor --json` before runtime-backed operations.",
				"Runtime services are optional; dependency-free mode is default.",
			}),
		},
		{
			Name:        "recipe-github-prs-to-warehouse",
			Description: "Sync GitHub pull requests into the local warehouse.",
			Body: skillBody("recipe-github-prs-to-warehouse", "Sync GitHub pull requests into the local warehouse.", []string{
				"Create a GitHub credential with config `owner`, `repo`, and `auth_type`, plus optional token from environment.",
				"Create a warehouse credential with a local path.",
				"Create a connection with stream `pull_requests` and table `github_pull_requests`.",
				"Run `pm etl run --connection github_to_warehouse --stream pull_requests --batch-size 100 --json`.",
			}),
		},
		{
			Name:        "recipe-preview-approve-reverse-etl",
			Description: "Preview and approve a reverse ETL plan safely.",
			Body: skillBody("recipe-preview-approve-reverse-etl", "Preview and approve a reverse ETL plan safely.", []string{
				"Create the reverse plan and inspect the preview.",
				"Do not execute without an explicit approval token from the user.",
				"Record the reverse run receipt after execution.",
			}),
		},
	}
	for _, manifest := range manifests {
		if manifest.Metadata.Name == "" {
			continue
		}
		name := "pm-" + manifest.Metadata.Name
		if name == "pm-warehouse" || name == "pm-outbox" || name == "pm-file" || name == "pm-sample" || name == "pm-github" {
			docs = append(docs, connectorSkill(manifest.Metadata.Name))
		}
	}
	return docs
}

func connectorSkill(name string) skillDoc {
	registry := appRegistry()
	connector, ok := registry.Get(name)
	if !ok {
		return skillDoc{}
	}
	guide := connectors.GuideOf(connector)
	return skillDoc{
		Name:        "pm-" + guide.Name,
		Description: guide.DisplayName + " connector knowledge and safe action guide.",
		Body:        connectors.RenderGuideSkill(guide),
	}
}

func skillBody(name, description string, bullets []string) string {
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("name: " + name + "\n")
	b.WriteString("description: " + description + "\n")
	b.WriteString("---\n\n")
	b.WriteString("# " + name + "\n\n")
	for _, bullet := range bullets {
		b.WriteString("- " + bullet + "\n")
	}
	return b.String()
}

func writeSkillIndex(path string, docs []skillDoc) error {
	sort.Slice(docs, func(i, j int) bool { return docs[i].Name < docs[j].Name })
	var b strings.Builder
	b.WriteString("# Skills Index\n\n")
	b.WriteString("> Auto-generated by `pm skills generate`.\n\n")
	for _, doc := range docs {
		b.WriteString("- [" + doc.Name + "](" + doc.Name + "/SKILL.md): " + doc.Description + "\n")
	}
	if err := os.WriteFile(path, []byte(b.String()), 0o644); err != nil {
		return fmt.Errorf("write skills index: %w", err)
	}
	return nil
}

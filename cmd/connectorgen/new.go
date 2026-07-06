package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// scaffoldNew creates internal/connectors/defs/<name>/ (rooted at defsRoot)
// from embedded templates: a bundle that passes `connectorgen validate`
// as-is (one stream, no write actions, a fully covered api_surface, and a
// fixed-heading docs.md). It rejects invalid names and pre-existing
// directories rather than silently overwriting.
func scaffoldNew(defsRoot, name string) error {
	if !namePattern.MatchString(name) {
		return fmt.Errorf("new: connector name %q does not match %s", name, namePatternDescription)
	}

	dir := filepath.Join(defsRoot, name)
	if _, err := os.Stat(dir); err == nil {
		return fmt.Errorf("new: %s already exists", dir)
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("new: stat %s: %w", dir, err)
	}

	if err := os.MkdirAll(filepath.Join(dir, "schemas"), 0o755); err != nil {
		return fmt.Errorf("new: %w", err)
	}

	files := scaffoldFiles(name)
	for relPath, contents := range files {
		full := filepath.Join(dir, relPath)
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			return fmt.Errorf("new: mkdir %s: %w", filepath.Dir(full), err)
		}
		if err := os.WriteFile(full, []byte(contents), 0o644); err != nil {
			return fmt.Errorf("new: write %s: %w", full, err)
		}
	}
	return nil
}

// scaffoldFiles returns the template file set for a new bundle named name.
// The templates are deliberately minimal (one read-only stream, no writes)
// but form a bundle that validate accepts outright: every stream is present
// in api_surface.json, the schema declares both x-primary-key and
// x-cursor-field, and docs.md carries every required heading.
func scaffoldFiles(name string) map[string]string {
	return map[string]string{
		"metadata.json": fmt.Sprintf(`{
  "name": %q,
  "display_name": %q,
  "description": "TODO: describe this connector",
  "integration_type": "api",
  "release_stage": "alpha",
  "capabilities": { "check": true, "read": true, "write": false, "query": false, "cdc": false, "dynamic_schema": false }
}
`, name, displayNameFor(name)),

		"spec.json": `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "required": ["base_url"],
  "properties": {
    "base_url": { "type": "string" },
    "token": { "type": "string", "x-secret": true }
  }
}
`,

		"streams.json": `{
  "base": {
    "url": "{{ config.base_url }}",
    "auth": [ { "mode": "bearer", "token": "{{ secrets.token }}", "when": "{{ config.base_url }}" } ],
    "pagination": { "type": "none" },
    "check": { "method": "GET", "path": "/ping" },
    "error_map": []
  },
  "streams": [
    {
      "name": "items",
      "path": "/items",
      "records": { "path": "data" },
      "schema": "schemas/items.json"
    }
  ]
}
`,

		"schemas/items.json": `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "object",
  "x-primary-key": ["id"],
  "x-cursor-field": "updated_at",
  "properties": {
    "id": { "type": "integer" },
    "updated_at": { "type": "string" }
  }
}
`,

		"api_surface.json": fmt.Sprintf(`{
  "api": %q,
  "reviewed_at": "TODO: YYYY-MM-DD",
  "scope": "TODO: describe API scope covered by this bundle",
  "endpoints": [
    { "method": "GET", "path": "/items", "covered_by": { "stream": "items" } }
  ]
}
`, displayNameFor(name)+" API"),

		"docs.md": fmt.Sprintf(`# Overview

TODO: describe %s.

## Auth setup

TODO: describe how to obtain the %s bearer token.

## Streams notes

`+"`items`"+` is a read-only stream.

## Write actions & risks

None yet.

## Known limits

TODO.
`, displayNameFor(name), displayNameFor(name)),
	}
}

// displayNameFor derives a human-readable title from a bare connector name
// (hyphen-separated -> Title Case), used only for scaffold placeholder text.
func displayNameFor(name string) string {
	title := true
	var b []rune
	for _, r := range name {
		if r == '-' {
			b = append(b, ' ')
			title = true
			continue
		}
		if title && r >= 'a' && r <= 'z' {
			r -= 'a' - 'A'
			title = false
		} else {
			title = false
		}
		b = append(b, r)
	}
	return string(b)
}

// runNew implements `connectorgen new <name>` against the real repo's
// internal/connectors/defs/ directory.
func runNew(args []string, stdout, stderr io.Writer) int {
	root, err := repoRoot()
	if err != nil {
		logln(stderr, "connectorgen new:", err)
		return 1
	}
	return runNewAt(args, stdout, stderr, filepath.Join(root, "internal/connectors/defs"))
}

// runNewAt is runNew with an explicit defsRoot, used directly by tests
// against scratch trees.
func runNewAt(args []string, stdout, stderr io.Writer, defsRoot string) int {
	if len(args) < 2 || args[1] == "" {
		logln(stderr, "connectorgen new: usage: connectorgen new <name>")
		return 2
	}
	name := args[1]
	if err := scaffoldNew(defsRoot, name); err != nil {
		logln(stderr, "connectorgen new:", err)
		return 1
	}
	logf(stdout, "connectorgen new: scaffolded %s\n", filepath.Join(defsRoot, name))
	return 0
}

package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"polymetrics.ai/internal/connectors/engine"
)

const certificationCandidatesFile = "certification.json"

// readCandidateGeneration is the narrow source-owned input used to derive
// direct-read candidates. Endpoint and command data always comes from the CLI
// surface; this shape contains only fixture values and auditable membership.
type readCandidateGeneration struct {
	RequiredFlagDefaults map[string]string
	Cohorts              []engine.CertificationReadCandidateCohort
}

// buildGeneratedReadCandidates derives serially executable direct-read
// candidates from declared command paths and required flags. It never promotes
// a candidate to a pass and deliberately excludes write, ETL, and unavailable
// commands: those require their own lifecycle contracts.
func buildGeneratedReadCandidates(connector string, commands []engine.CLICommand, generation readCandidateGeneration) ([]engine.CertificationCommandCandidate, error) {
	cohortByCommand := make(map[string]string)
	for _, cohort := range generation.Cohorts {
		if strings.TrimSpace(cohort.Name) == "" {
			return nil, fmt.Errorf("read candidate cohort has an empty name")
		}
		for _, path := range cohort.Commands {
			if prior, exists := cohortByCommand[path]; exists {
				return nil, fmt.Errorf("read candidate command %q belongs to cohorts %q and %q", path, prior, cohort.Name)
			}
			cohortByCommand[path] = cohort.Name
		}
	}

	byPath := make(map[string]engine.CLICommand, len(commands))
	for _, command := range commands {
		if strings.TrimSpace(command.Path) == "" {
			return nil, fmt.Errorf("cli surface contains a command with an empty path")
		}
		if _, exists := byPath[command.Path]; exists {
			return nil, fmt.Errorf("cli surface duplicates command path %q", command.Path)
		}
		byPath[command.Path] = command
	}
	for path := range cohortByCommand {
		if _, exists := byPath[path]; !exists {
			return nil, fmt.Errorf("read candidate cohort command %q is absent from cli_surface.json", path)
		}
	}

	paths := make([]string, 0, len(cohortByCommand))
	for path := range cohortByCommand {
		paths = append(paths, path)
	}
	sort.Strings(paths)
	generated := make([]engine.CertificationCommandCandidate, 0, len(paths))
	for _, path := range paths {
		command := byPath[path]
		if command.Intent != "direct_read" || command.Availability != "implemented" {
			continue
		}
		args := []engine.CertificationCommandArg{{Connector: true}}
		for _, token := range strings.Fields(command.Path) {
			args = append(args, engine.CertificationCommandArg{Literal: token})
		}
		args = append(args,
			engine.CertificationCommandArg{Literal: "--credential"},
			engine.CertificationCommandArg{SourceCredential: true},
		)
		for _, flag := range command.Flags {
			if !flag.Required {
				continue
			}
			defaultValue, found := generation.RequiredFlagDefaults[flag.Name]
			if !found || strings.TrimSpace(defaultValue) == "" {
				return nil, fmt.Errorf("read candidate %q requires --%s without a connector-owned default", command.Path, flag.Name)
			}
			args = append(args,
				engine.CertificationCommandArg{Literal: "--" + flag.Name},
				engine.CertificationCommandArg{ConfigKey: flag.Name, Default: defaultValue},
			)
		}
		args = append(args, engine.CertificationCommandArg{Literal: "--json"})
		generated = append(generated, engine.CertificationCommandCandidate{
			StageName: "generated_direct_read_" + strings.NewReplacer(" ", "_", "-", "_").Replace(command.Path),
			Command:   command.Path,
			Args:      args,
			OutputAssertions: []engine.CertificationOutputAssertion{{
				JSONPointer: "/response",
				ValueType:   "object_or_array",
			}},
			Cohort:    cohortByCommand[path],
			Generated: true,
		})
	}
	return generated, nil
}

func runCertificationCandidates(args []string, stdout, stderr io.Writer) int {
	root := "."
	connector := ""
	check := false
	for index := 1; index < len(args); index++ {
		arg := args[index]
		switch {
		case arg == "--check":
			check = true
		case strings.HasPrefix(arg, "--connector="):
			connector = strings.TrimPrefix(arg, "--connector=")
		case arg == "--connector":
			if index+1 >= len(args) || strings.HasPrefix(args[index+1], "-") {
				logln(stderr, "connectorgen certification-candidates: --connector requires a name")
				return 2
			}
			index++
			connector = args[index]
		case strings.HasPrefix(arg, "-"):
			logf(stderr, "connectorgen certification-candidates: unknown flag %q\n", arg)
			return 2
		case root == ".":
			root = arg
		default:
			logf(stderr, "connectorgen certification-candidates: unexpected extra argument %q\n", arg)
			return 2
		}
	}
	if strings.TrimSpace(connector) == "" {
		logln(stderr, "connectorgen certification-candidates: --connector requires a name")
		return 2
	}
	if code := generateCertificationCandidates(root, connector, check); code != nil {
		logf(stderr, "connectorgen certification-candidates: %v\n", code)
		return 1
	}
	if check {
		logf(stdout, "certification candidates are current: connector=%s\n", connector)
	} else {
		logf(stdout, "generated certification candidates: connector=%s\n", connector)
	}
	return 0
}

func generateCertificationCandidates(root, connector string, check bool) error {
	absRoot, err := filepath.Abs(root)
	if err != nil {
		return fmt.Errorf("resolve repository root: %w", err)
	}
	definitionsRoot := filepath.Join(absRoot, "internal", "connectors", "defs")
	bundle, err := engine.Load(os.DirFS(definitionsRoot), connector)
	if err != nil {
		return fmt.Errorf("load connector %q: %w", connector, err)
	}
	if bundle.Certification == nil || bundle.Certification.DirectReadGeneration == nil {
		return fmt.Errorf("connector %q has no direct_read_generation declaration", connector)
	}
	if bundle.CLISurface == nil {
		return fmt.Errorf("connector %q has no cli_surface.json", connector)
	}
	generation := bundle.Certification.DirectReadGeneration
	generated, err := buildGeneratedReadCandidates(connector, bundle.CLISurface.Commands, readCandidateGeneration{
		RequiredFlagDefaults: generation.RequiredFlagDefaults,
		Cohorts:              generation.Cohorts,
	})
	if err != nil {
		return err
	}
	bundle.Certification.DirectReadCandidates, err = mergeGeneratedReadCandidates(bundle.Certification.DirectReadCandidates, generated)
	if err != nil {
		return err
	}
	raw, err := json.MarshalIndent(bundle.Certification, "", "  ")
	if err != nil {
		return fmt.Errorf("render certification candidates: %w", err)
	}
	raw = append(raw, '\n')
	path := filepath.Join(definitionsRoot, connector, certificationCandidatesFile)
	if check {
		committed, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read certification candidates: %w", err)
		}
		if !bytes.Equal(committed, raw) {
			return fmt.Errorf("certification candidates are stale; run `connectorgen certification-candidates --connector %s`", connector)
		}
		return nil
	}
	if err := writeGeneratedArtifact(path, raw); err != nil {
		return fmt.Errorf("write certification candidates: %w", err)
	}
	return nil
}

// mergeGeneratedReadCandidates leaves explicitly authored candidates intact.
// A manual command shadows a generated command because only its author can
// supply a more specific produced-value assertion or argument shape.
func mergeGeneratedReadCandidates(existing, generated []engine.CertificationCommandCandidate) ([]engine.CertificationCommandCandidate, error) {
	manual := make([]engine.CertificationCommandCandidate, 0, len(existing))
	manualCommands := make(map[string]struct{})
	for _, candidate := range existing {
		if candidate.Generated {
			continue
		}
		if _, duplicate := manualCommands[candidate.Command]; duplicate {
			return nil, fmt.Errorf("manual direct-read candidate command %q is duplicated", candidate.Command)
		}
		manualCommands[candidate.Command] = struct{}{}
		manual = append(manual, candidate)
	}
	filtered := generated[:0]
	for _, candidate := range generated {
		if _, manual := manualCommands[candidate.Command]; manual {
			continue
		}
		filtered = append(filtered, candidate)
	}
	return append(manual, filtered...), nil
}

package cli

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sort"
	"strings"

	"polymetrics.ai/internal/config"
	"polymetrics.ai/internal/pmbroker"
)

func runPMBrokerContext(cfg config.Config, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	store, err := pmBrokerStore()
	if err != nil {
		return err
	}
	switch args[0] {
	case "create":
		if len(args) < 2 {
			return errUsage
		}
		ctx, err := pmBrokerContextFromFlags(args[1], cfg, parseFlags(args[2:]))
		if err != nil {
			return pmBrokerValidationError(err)
		}
		state, err := store.UpsertContext(ctx)
		if err != nil {
			return pmBrokerValidationError(err)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "PMBrokerContext", "context": ctx, "active_context": state.ActiveContext})
		}
		fmt.Fprintf(stdout, "Saved PM Broker context %s (%s / %s / %s)\n", ctx.Name, ctx.Organization.DisplayName, ctx.Workspace.DisplayName, ctx.Environment.DisplayName)
		return nil
	case "use":
		if len(args) < 2 {
			return errUsage
		}
		state, err := store.UseContext(args[1])
		if err != nil {
			return pmBrokerValidationError(err)
		}
		ctx, _ := state.ContextByName(args[1])
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "PMBrokerContext", "context": ctx, "active_context": state.ActiveContext})
		}
		fmt.Fprintf(stdout, "Active PM Broker context: %s\n", args[1])
		return nil
	case "show":
		flags := parseFlags(args[1:])
		state, err := store.Load()
		if err != nil {
			return pmBrokerValidationError(err)
		}
		resolved, err := pmbroker.ResolveContext(pmbroker.ResolveRequest{
			State:                  state,
			ExplicitContext:        strings.TrimSpace(flags.first("context")),
			ProjectRequiredContext: cfg.Broker.RequiredContext,
			ProjectDefaultContext:  cfg.Broker.DefaultContext,
			AllowLegacyLocal:       true,
		})
		if err != nil {
			return pmBrokerValidationError(err)
		}
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "PMBrokerResolvedContext", "source": string(resolved.Source), "context": resolved.Context})
		}
		writePMBrokerContextHuman(stdout, resolved.Context, resolved.Source)
		return nil
	case "list":
		state, err := store.Load()
		if err != nil {
			return pmBrokerValidationError(err)
		}
		contexts := append([]pmbroker.Context(nil), state.Contexts...)
		sort.Slice(contexts, func(i, j int) bool { return contexts[i].Name < contexts[j].Name })
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "PMBrokerContextList", "active_context": state.ActiveContext, "contexts": contexts})
		}
		for _, ctx := range contexts {
			active := ""
			if ctx.Name == state.ActiveContext {
				active = "\tactive"
			}
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\t%s%s\n", ctx.Name, ctx.Organization.DisplayName, ctx.Workspace.DisplayName, ctx.Environment.DisplayName, ctx.Runtime.Mode, active)
		}
		return nil
	default:
		return errUsage
	}
}

func runPMBrokerMetadata(kind string, _ config.Config, args []string, stdout io.Writer, jsonOut bool) error {
	if len(args) == 0 {
		return errUsage
	}
	if !validPMBrokerMetadataAction(kind, args[0]) {
		return errUsage
	}
	store, err := pmBrokerStore()
	if err != nil {
		return err
	}
	state, err := store.Load()
	if err != nil {
		return pmBrokerValidationError(err)
	}
	flags := parseFlags(args[1:])
	switch kind {
	case "organizations":
		return runPMBrokerOrganizations(state, args, stdout, jsonOut)
	case "workspaces":
		return runPMBrokerWorkspaces(state, args, flags, stdout, jsonOut)
	case "environments":
		return runPMBrokerEnvironments(state, args, flags, stdout, jsonOut)
	default:
		return errUsage
	}
}

func validPMBrokerMetadataAction(kind, action string) bool {
	switch kind {
	case "organizations", "workspaces", "environments":
		return action == "list" || action == "show"
	default:
		return false
	}
}

func runPMBrokerOrganizations(state pmbroker.UserState, args []string, stdout io.Writer, jsonOut bool) error {
	switch args[0] {
	case "list":
		organizations := state.Organizations()
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "OrganizationList", "organizations": organizations})
		}
		for _, org := range organizations {
			fmt.Fprintf(stdout, "%s\t%s\n", org.ID, org.DisplayName)
		}
		return nil
	case "show":
		if len(args) < 2 {
			return errUsage
		}
		for _, org := range state.Organizations() {
			if string(org.ID) == args[1] {
				if jsonOut {
					return writeJSON(stdout, envelope{"kind": "Organization", "organization": org})
				}
				b, _ := json.MarshalIndent(org, "", "  ")
				fmt.Fprintln(stdout, string(b))
				return nil
			}
		}
		return pmBrokerValidationError(pmbroker.ErrContextNotFound)
	default:
		return errUsage
	}
}

func runPMBrokerWorkspaces(state pmbroker.UserState, args []string, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	switch args[0] {
	case "list":
		workspaces := state.Workspaces(pmbroker.OrganizationID(flags.first("organization")))
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "WorkspaceList", "workspaces": workspaces})
		}
		for _, workspace := range workspaces {
			fmt.Fprintf(stdout, "%s\t%s\t%s\n", workspace.ID, workspace.OrganizationID, workspace.DisplayName)
		}
		return nil
	case "show":
		if len(args) < 2 {
			return errUsage
		}
		for _, workspace := range state.Workspaces("") {
			if string(workspace.ID) == args[1] {
				if jsonOut {
					return writeJSON(stdout, envelope{"kind": "Workspace", "workspace": workspace})
				}
				b, _ := json.MarshalIndent(workspace, "", "  ")
				fmt.Fprintln(stdout, string(b))
				return nil
			}
		}
		return pmBrokerValidationError(pmbroker.ErrContextNotFound)
	default:
		return errUsage
	}
}

func runPMBrokerEnvironments(state pmbroker.UserState, args []string, flags parsedFlags, stdout io.Writer, jsonOut bool) error {
	switch args[0] {
	case "list":
		environments := state.Environments(pmbroker.WorkspaceID(flags.first("workspace")))
		if jsonOut {
			return writeJSON(stdout, envelope{"kind": "EnvironmentList", "environments": environments})
		}
		for _, env := range environments {
			fmt.Fprintf(stdout, "%s\t%s\t%s\t%s\n", env.ID, env.WorkspaceID, env.Type, env.DisplayName)
		}
		return nil
	case "show":
		if len(args) < 2 {
			return errUsage
		}
		for _, env := range state.Environments("") {
			if string(env.ID) == args[1] {
				if jsonOut {
					return writeJSON(stdout, envelope{"kind": "Environment", "environment": env})
				}
				b, _ := json.MarshalIndent(env, "", "  ")
				fmt.Fprintln(stdout, string(b))
				return nil
			}
		}
		return pmBrokerValidationError(pmbroker.ErrContextNotFound)
	default:
		return errUsage
	}
}

func pmBrokerContextFromFlags(name string, cfg config.Config, flags parsedFlags) (pmbroker.Context, error) {
	envType, err := pmbroker.ParseEnvironmentType(flags.first("environment-type"))
	if err != nil {
		return pmbroker.Context{}, fmt.Errorf("--environment-type: %w", err)
	}
	runtimeModeRaw := valueOr(flags.first("runtime-mode"), cfg.Broker.RuntimeMode)
	mode, err := pmbroker.ParseRuntimeMode(runtimeModeRaw)
	if err != nil {
		return pmbroker.Context{}, fmt.Errorf("--runtime-mode: %w", err)
	}
	policy := valueOr(flags.first("hybrid-policy"), cfg.Broker.HybridPolicy)
	ctx := pmbroker.Context{
		Name: strings.TrimSpace(name),
		Organization: pmbroker.Organization{
			ID:          pmbroker.OrganizationID(flags.first("organization")),
			DisplayName: flags.first("organization-name"),
		},
		Workspace: pmbroker.Workspace{
			ID:             pmbroker.WorkspaceID(flags.first("workspace")),
			OrganizationID: pmbroker.OrganizationID(flags.first("organization")),
			DisplayName:    flags.first("workspace-name"),
		},
		Environment: pmbroker.Environment{
			ID:             pmbroker.EnvironmentID(flags.first("environment")),
			WorkspaceID:    pmbroker.WorkspaceID(flags.first("workspace")),
			OrganizationID: pmbroker.OrganizationID(flags.first("organization")),
			DisplayName:    flags.first("environment-name"),
			Type:           envType,
		},
		BrokerProfile: pmbroker.BrokerProfile{
			ID:             pmbroker.BrokerProfileID(flags.first("broker-profile")),
			OrganizationID: pmbroker.OrganizationID(flags.first("organization")),
			WorkspaceID:    pmbroker.WorkspaceID(flags.first("workspace")),
			EnvironmentID:  pmbroker.EnvironmentID(flags.first("environment")),
			DisplayName:    flags.first("broker-profile-name"),
		},
		Runtime: pmbroker.RuntimeModeSelection{Mode: mode, PolicyBindingID: policy},
	}
	if err := ctx.Validate(); err != nil {
		return pmbroker.Context{}, err
	}
	return ctx, nil
}

func writePMBrokerContextHuman(stdout io.Writer, ctx pmbroker.Context, source pmbroker.ResolveSource) {
	fmt.Fprintf(stdout, "context=%s source=%s\n", ctx.Name, source)
	fmt.Fprintf(stdout, "Organization\t%s\t%s\n", ctx.Organization.ID, ctx.Organization.DisplayName)
	fmt.Fprintf(stdout, "Workspace\t%s\t%s\n", ctx.Workspace.ID, ctx.Workspace.DisplayName)
	fmt.Fprintf(stdout, "Environment\t%s\t%s\t%s\n", ctx.Environment.ID, ctx.Environment.Type, ctx.Environment.DisplayName)
	fmt.Fprintf(stdout, "BrokerProfile\t%s\t%s\n", ctx.BrokerProfile.ID, ctx.BrokerProfile.DisplayName)
	fmt.Fprintf(stdout, "RuntimeMode\t%s\n", ctx.Runtime.Mode)
}

func pmBrokerStore() (pmbroker.Store, error) {
	path, err := pmbroker.DefaultUserStatePath()
	if err != nil {
		return pmbroker.Store{}, err
	}
	return pmbroker.Store{Path: path}, nil
}

func pmBrokerValidationError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pmbroker.ErrInvalidIdentityBoundary) ||
		errors.Is(err, pmbroker.ErrInvalidContextName) ||
		errors.Is(err, pmbroker.ErrContextNotFound) ||
		errors.Is(err, pmbroker.ErrContextMismatch) ||
		errors.Is(err, pmbroker.ErrDuplicateContext) ||
		errors.Is(err, pmbroker.ErrInvalidRuntimeMode) ||
		errors.Is(err, pmbroker.ErrInvalidRuntimeOperation) ||
		errors.Is(err, pmbroker.ErrHybridPolicyRequired) ||
		errors.Is(err, pmbroker.ErrInvalidPolicyBindingID) ||
		errors.Is(err, pmbroker.ErrProductionLocalFallbackForbidden) ||
		errors.Is(err, pmbroker.ErrUnsafeState) {
		return validationErrorf("%v", err)
	}
	return err
}

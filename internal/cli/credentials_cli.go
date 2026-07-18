package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"polymetrics.ai/internal/app"
	"polymetrics.ai/internal/connectors"
	"polymetrics.ai/internal/safety"
)

type credentialsAddFlags struct {
	Connectors []string
	FromEnv    []string
	ValueStdin []string
	Configs    []string
}

type credentialEnvSource struct {
	field    string
	variable string
}

func newCredentialsCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := &cobra.Command{
		Use:               "credentials",
		Args:              cobra.ArbitraryArgs,
		SilenceErrors:     true,
		SilenceUsage:      true,
		ValidArgsFunction: completeNoFile,
		RunE: func(_ *cobra.Command, args []string) error {
			if len(args) > 0 {
				return errUsage
			}
			return markCobraLegacyError(writeManual("credentials", stdout, jsonOut))
		},
	}
	setManualHelp(cmd, "credentials", stdout, jsonOut)
	cmd.AddCommand(newCredentialsAddCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newCredentialsListCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newCredentialsInspectCobraCommand(root, stdout, jsonOut))
	cmd.AddCommand(newCredentialsTestCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newCredentialsRemoveCobraCommand(ctx, root, stdout, jsonOut))
	cmd.AddCommand(newCredentialsHelpCobraCommand(stdout, jsonOut))
	return cmd
}

func newCredentialsAddCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	var flags credentialsAddFlags
	cmd := newCredentialsActionCobraCommand("add <name>", func(cmd *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runCredentialsAdd(ctx, a, args[0], flags, cmd.InOrStdin(), stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "credentials", stdout, jsonOut)
	addCredentialsStringArrayFlag(cmd, &flags.Connectors, "connector", "connector that owns the credential")
	addCredentialsStringArrayFlag(cmd, &flags.FromEnv, "from-env", "secret source field=ENV")
	addCredentialsStringArrayFlag(cmd, &flags.ValueStdin, "value-stdin", "secret field read from standard input")
	addCredentialsStringArrayFlag(cmd, &flags.Configs, "config", "non-secret connector config key=value")
	return cmd
}

func newCredentialsListCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newCredentialsActionCobraCommand("list", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runCredentialsList(a, stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "credentials", stdout, jsonOut)
	return cmd
}

func newCredentialsInspectCobraCommand(root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newCredentialsActionCobraCommand("inspect <name>", func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runCredentialsInspect(a, args[0], stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "credentials", stdout, jsonOut)
	return cmd
}

func newCredentialsTestCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newCredentialsActionCobraCommand("test <name>", func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runCredentialsTest(ctx, a, args[0], stdout, jsonOut)
		}))
	})
	setManualHelp(cmd, "credentials", stdout, jsonOut)
	return cmd
}

func newCredentialsRemoveCobraCommand(ctx context.Context, root string, stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newCredentialsActionCobraCommand("remove <name>", func(_ *cobra.Command, args []string) error {
		if len(args) < 1 {
			return errUsage
		}
		return markCobraLegacyError(withApp(root, func(a *app.App) error {
			return runCredentialsRemove(ctx, a, args[0], stdout)
		}))
	})
	setManualHelp(cmd, "credentials", stdout, jsonOut)
	return cmd
}

func newCredentialsHelpCobraCommand(stdout io.Writer, jsonOut bool) *cobra.Command {
	cmd := newCredentialsActionCobraCommand("help", func(_ *cobra.Command, _ []string) error {
		return markCobraLegacyError(writeManual("credentials", stdout, jsonOut))
	})
	cmd.Hidden = true
	setManualHelp(cmd, "credentials", stdout, jsonOut)
	return cmd
}

func newCredentialsActionCobraCommand(use string, run func(*cobra.Command, []string) error) *cobra.Command {
	return &cobra.Command{
		Use:           use,
		Args:          cobra.ArbitraryArgs,
		SilenceErrors: true,
		SilenceUsage:  true,
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
		ValidArgsFunction: completeNoFile,
		RunE:              run,
	}
}

func addCredentialsStringArrayFlag(cmd *cobra.Command, target *[]string, name, usage string) {
	cmd.Flags().StringArrayVar(target, name, nil, usage)
	if flag := cmd.Flags().Lookup(name); flag != nil {
		flag.NoOptDefVal = "true"
	}
}

func runCredentialsAdd(ctx context.Context, a *app.App, name string, flags credentialsAddFlags, stdin io.Reader, stdout io.Writer, jsonOut bool) error {
	if err := validateCredentialIdentifier(name, "credential"); err != nil {
		return err
	}
	connector := lastString(flags.Connectors)
	if connector == "" {
		return errors.New("missing --connector")
	}
	if err := validateCredentialIdentifier(connector, "connector"); err != nil {
		return err
	}
	if err := connectors.RejectLegacyConnectorName(connector); err != nil {
		return err
	}

	envSources, err := parseCredentialEnvSources(flags.FromEnv)
	if err != nil {
		return err
	}
	stdinField := lastString(flags.ValueStdin)
	if stdinField != "" {
		if err := validateCredentialIdentifier(stdinField, "secret field"); err != nil {
			return err
		}
	}
	credentialConfig, err := parseCredentialConfig(flags.Configs)
	if err != nil {
		return err
	}
	if err := validateCredentialConfig(a, connector, credentialConfig); err != nil {
		return err
	}

	secrets := make(map[string]string, len(envSources)+1)
	for _, source := range envSources {
		secrets[source.field] = os.Getenv(source.variable)
		if secrets[source.field] == "" {
			return fmt.Errorf("environment variable %s is empty", source.variable)
		}
	}
	if stdinField != "" {
		value, err := io.ReadAll(stdin)
		if err != nil {
			return fmt.Errorf("read stdin secret: %w", err)
		}
		secrets[stdinField] = strings.TrimRight(string(value), "\r\n")
	}

	cred, err := a.AddCredential(ctx, app.AddCredentialRequest{
		Name:      name,
		Connector: connector,
		Config:    credentialConfig,
		Secrets:   secrets,
	})
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "Credential", "credential": cred})
	}
	fmt.Fprintf(stdout, "Saved credential %s for connector %s\n", cred.Name, cred.Connector)
	return nil
}

func runCredentialsList(a *app.App, stdout io.Writer, jsonOut bool) error {
	creds := a.ListCredentials()
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CredentialList", "credentials": creds})
	}
	for _, cred := range creds {
		fmt.Fprintf(stdout, "%s\t%s\t%s\n", cred.Name, cred.ID, cred.Connector)
	}
	return nil
}

func runCredentialsInspect(a *app.App, name string, stdout io.Writer, jsonOut bool) error {
	if err := validateCredentialIdentifier(name, "credential"); err != nil {
		return err
	}
	cred, err := a.InspectCredential(name)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "Credential", "credential": cred})
	}
	value, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return fmt.Errorf("encode credential metadata: %w", err)
	}
	fmt.Fprintln(stdout, string(value))
	return nil
}

func runCredentialsTest(ctx context.Context, a *app.App, name string, stdout io.Writer, jsonOut bool) error {
	if err := validateCredentialIdentifier(name, "credential"); err != nil {
		return err
	}
	cred, err := a.TestCredential(ctx, name)
	if err != nil {
		return err
	}
	if jsonOut {
		return writeJSON(stdout, envelope{"kind": "CredentialTest", "status": "ok", "credential": cred})
	}
	fmt.Fprintf(stdout, "Credential %s validated\n", cred.Name)
	return nil
}

func runCredentialsRemove(ctx context.Context, a *app.App, name string, stdout io.Writer) error {
	if err := validateCredentialIdentifier(name, "credential"); err != nil {
		return err
	}
	if err := a.RemoveCredential(ctx, name); err != nil {
		return err
	}
	fmt.Fprintf(stdout, "Removed credential %s\n", name)
	return nil
}

func parseCredentialEnvSources(values []string) ([]credentialEnvSource, error) {
	out := make([]credentialEnvSource, 0, len(values))
	for _, spec := range values {
		field, variable, ok := strings.Cut(spec, "=")
		if !ok || field == "" || variable == "" {
			return nil, fmt.Errorf("invalid --from-env %q, want field=ENV", spec)
		}
		if err := validateCredentialIdentifier(field, "secret field"); err != nil {
			return nil, err
		}
		if err := validateEnvironmentVariableName(variable); err != nil {
			return nil, err
		}
		out = append(out, credentialEnvSource{field: field, variable: variable})
	}
	return out, nil
}

func parseCredentialConfig(values []string) (map[string]string, error) {
	out, err := keyValues(values)
	if err != nil {
		return nil, err
	}
	for key := range out {
		if err := validateCredentialIdentifier(key, "config key"); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func validateCredentialIdentifier(value, field string) error {
	if err := safety.ValidateIdentifier(value, field); err != nil {
		return validationErrorf("%v", err)
	}
	if field == "credential" || field == "connector" {
		first := value[0]
		if !((first >= 'a' && first <= 'z') || (first >= 'A' && first <= 'Z') || (first >= '0' && first <= '9')) {
			return validationErrorf("%s must start with an ASCII letter or digit", field)
		}
	}
	return nil
}

func validateEnvironmentVariableName(value string) error {
	if value == "" {
		return validationErrorf("environment variable name is required")
	}
	for i, r := range value {
		if (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || r == '_' || (i > 0 && r >= '0' && r <= '9') {
			continue
		}
		return validationErrorf("environment variable name contains invalid characters")
	}
	return nil
}

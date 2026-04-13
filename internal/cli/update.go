package cli

import (
	"fmt"
	"io"

	"boottree/internal/render"
	"boottree/internal/update"

	"github.com/spf13/cobra"
)

type updatePrompter interface {
	ConfirmUpdate(in io.Reader, out io.Writer, errOut io.Writer, plan update.Plan) (bool, error)
}

type updateCommandFlags struct {
	CheckOnly   bool
	Yes         bool
	Version     string
	Channel     string
	ManifestURL string
	InstallPath string
}

func newUpdateCommand() *cobra.Command {
	return newUpdateCommandWithDependencies(update.NewService(), surveyInitPrompter{})
}

func newUpdateCommandWithDependencies(service update.Service, prompter updatePrompter) *cobra.Command {
	flags := &updateCommandFlags{}
	cmd := &cobra.Command{
		Use:   "update",
		Short: "Check for and install a newer BootTree binary",
		Long:  "Resolve an update plan from a release manifest, then optionally download, verify, and install the matching BootTree binary for the current platform.",
		Example: "  boottree update --check\n" +
			"  boottree update --yes\n" +
			"  boottree update --version 0.2.0 --manifest-url https://example.com/boottree/manifest.json",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runUpdate(cmd, service, prompter, flags)
		},
	}

	cmd.Flags().BoolVar(&flags.CheckOnly, "check", false, "Only check whether an update is available")
	cmd.Flags().BoolVar(&flags.Yes, "yes", false, "Apply the update without confirmation")
	cmd.Flags().StringVar(&flags.Version, "version", "", "Install a specific version from the manifest instead of the latest")
	cmd.Flags().StringVar(&flags.Channel, "channel", "stable", "Release channel to resolve from the manifest")
	cmd.Flags().StringVar(&flags.ManifestURL, "manifest-url", "", "HTTPS manifest URL that describes available releases")
	cmd.Flags().StringVar(&flags.InstallPath, "install-path", "", "Override the target install path for the updated binary")
	return cmd
}

func runUpdate(cmd *cobra.Command, service update.Service, prompter updatePrompter, flags *updateCommandFlags) error {
	options := update.Options{
		CheckOnly:   flags.CheckOnly,
		Yes:         flags.Yes,
		Version:     flags.Version,
		Channel:     flags.Channel,
		ManifestURL: flags.ManifestURL,
		InstallPath: flags.InstallPath,
	}

	plan, err := service.BuildPlan(cmd.Context(), options)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintln(cmd.OutOrStdout(), render.UpdatePlan(plan)); err != nil {
		return err
	}
	if plan.IsNoop || flags.CheckOnly {
		return nil
	}

	if !flags.Yes && isInteractiveTerminal(cmd.InOrStdin(), cmd.OutOrStdout()) {
		confirmed, err := prompter.ConfirmUpdate(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), plan)
		if err != nil {
			return err
		}
		if !confirmed {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Update canceled.")
			return nil
		}
	}

	result, err := service.Apply(cmd.Context(), plan)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintln(cmd.OutOrStdout(), render.UpdateResult(result))
	return err
}

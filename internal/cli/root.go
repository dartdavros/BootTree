package cli

import (
	"fmt"

	"boottree/internal/buildinfo"
	"boottree/internal/platform"

	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	return newRootCommandWithDependencies(platform.SelfInstaller{}, surveyInitPrompter{})
}

func newRootCommandWithDependencies(installer selfInstaller, prompter installPrompter) *cobra.Command {
	cmd := &cobra.Command{
		Use:           "boottree",
		Short:         "Standardize and inspect local project structure",
		Long:          "BootTree is a cross-platform CLI for initializing, previewing, and inspecting standardized project structures.",
		SilenceUsage:  true,
		SilenceErrors: true,
		Version:       buildinfo.Version,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			return maybeOfferInstall(cmd, installer, prompter)
		},
		Example: "  boottree init\n" +
			"  boottree init --preset software-product --mode folders-only --dry-run\n" +
			"  boottree tree --depth 2\n" +
			"  boottree stats\n" +
			"  boottree install\n" +
			"  boottree update --check\n" +
			"  boottree version",
	}
	cmd.SetVersionTemplate("boottree {{.Version}}\n")
	cmd.CompletionOptions.DisableDefaultCmd = false
	cmd.InitDefaultCompletionCmd()

	cmd.AddCommand(newVersionCommand())
	cmd.AddCommand(newInitCommand())
	cmd.AddCommand(newTreeCommand())
	cmd.AddCommand(newStatsCommand())
	cmd.AddCommand(newInstallCommandWithDependencies(installer, prompter))
	cmd.AddCommand(newUpdateCommand())
	return cmd
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:     "version",
		Short:   "Print CLI version information",
		Long:    "Print BootTree version information, including version, commit, and build date when available.",
		Example: "  boottree version\n  boottree --version",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := fmt.Fprintln(cmd.OutOrStdout(), buildinfo.Detailed())
			return err
		},
	}
}

package cli

import (
	"fmt"
	"runtime"
	"strings"

	"boottree/internal/platform"

	"github.com/spf13/cobra"
)

type selfInstaller interface {
	Detect() (platform.InstallState, error)
	InstallForCurrentUser() (platform.InstallResult, error)
}

type installCommandFlags struct {
	Yes bool
}

func newInstallCommand() *cobra.Command {
	return newInstallCommandWithDependencies(platform.SelfInstaller{}, surveyInitPrompter{})
}

func newInstallCommandWithDependencies(installer selfInstaller, prompter installPrompter) *cobra.Command {
	flags := &installCommandFlags{}
	cmd := &cobra.Command{
		Use:   "install",
		Short: "Install BootTree for the current user",
		Long:  "Copy the current BootTree executable to a standard user-local bin directory and update PATH when the platform supports it.",
		Example: "  boottree install\n" +
			"  boottree install --yes",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInstall(cmd, installer, prompter, flags)
		},
	}
	cmd.Flags().BoolVar(&flags.Yes, "yes", false, "Install without confirmation")
	return cmd
}

func runInstall(cmd *cobra.Command, installer selfInstaller, prompter installPrompter, flags *installCommandFlags) error {
	state, err := installer.Detect()
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), renderInstallPlan(state))

	if !flags.Yes && isInteractiveTerminal(cmd.InOrStdin(), cmd.OutOrStdout()) {
		confirmed, err := prompter.ConfirmInstall(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), state)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Installation canceled.")
			return nil
		}
	}

	result, err := installer.InstallForCurrentUser()
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), renderInstallResult(result))
	return nil
}

func shouldOfferInstall(commandPath string, goos string, interactive bool) bool {
	if goos != "windows" || !interactive {
		return false
	}
	if commandPath == "boottree" {
		return false
	}
	if strings.HasSuffix(commandPath, " help") {
		return false
	}
	forbidden := map[string]struct{}{
		"boottree completion": {},
		"boottree install":    {},
		"boottree update":     {},
		"boottree version":    {},
	}
	_, blocked := forbidden[commandPath]
	return !blocked
}

func maybeOfferInstall(cmd *cobra.Command, installer selfInstaller, prompter installPrompter) error {
	if helpRequested(cmd) {
		return nil
	}
	if !shouldOfferInstall(cmd.CommandPath(), runtime.GOOS, isInteractiveTerminal(cmd.InOrStdin(), cmd.OutOrStdout())) {
		return nil
	}

	state, err := installer.Detect()
	if err != nil {
		return nil
	}
	if state.AvailableInPath {
		return nil
	}

	confirmed, err := prompter.ConfirmInstall(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), state)
	if err != nil {
		return err
	}
	if !confirmed {
		fmt.Fprintln(cmd.ErrOrStderr(), "Tip: run 'boottree install' later to add BootTree to PATH for the current user.")
		return nil
	}

	result, err := installer.InstallForCurrentUser()
	if err != nil {
		return err
	}
	fmt.Fprintln(cmd.ErrOrStderr(), renderInstallResult(result))
	return nil
}

func renderInstallPlan(state platform.InstallState) string {
	lines := []string{
		"Install plan",
		fmt.Sprintf("  Command name: %s", state.CommandName),
		fmt.Sprintf("  Current executable: %s", state.CurrentExecutable),
		fmt.Sprintf("  Target directory: %s", state.SuggestedInstallDir),
	}
	if state.AvailableInPath {
		lines = append(lines, "  PATH status: BootTree is already available in PATH")
	} else {
		lines = append(lines, "  PATH status: BootTree is not currently available in PATH")
	}
	return strings.Join(lines, "\n")
}

func renderInstallResult(result platform.InstallResult) string {
	lines := []string{
		"BootTree installation completed.",
		fmt.Sprintf("  Installed executable: %s", result.InstalledExecutable),
	}
	if result.PathUpdated {
		lines = append(lines, "  PATH update: user PATH was updated")
	} else if result.ManualPathUpdateRequired {
		lines = append(lines, fmt.Sprintf("  PATH update: add %s to PATH manually", result.InstallDir))
	} else {
		lines = append(lines, "  PATH update: install directory was already present in PATH")
	}
	if result.ShellRestartRecommended {
		lines = append(lines, "  Next step: close this terminal window and open a new one before using the installed command.")
	}
	return strings.Join(lines, "\n")
}

func helpRequested(cmd *cobra.Command) bool {
	flag := cmd.Flags().Lookup("help")
	return flag != nil && flag.Changed
}

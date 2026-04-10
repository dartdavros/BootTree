package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"

	"boottree/internal/app"
	"boottree/internal/core/model"
	"boottree/internal/core/planner"
	"boottree/internal/core/scanner"
	"boottree/internal/fs"
	"boottree/internal/render"

	"github.com/spf13/cobra"
)

type initCommandFlags struct {
	Preset  string
	Mode    string
	Include []string
	DryRun  bool
	Yes     bool
	Force   bool
}

type parsedInitArgs struct {
	Preset      string
	Mode        string
	Include     []string
	DryRun      bool
	Yes         bool
	Force       bool
	Interactive bool
}

func newInitCommand() *cobra.Command {
	return newInitCommandWithPrompter(surveyInitPrompter{})
}

func newInitCommandWithPrompter(prompter initPrompter) *cobra.Command {
	flags := &initCommandFlags{}
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the current project from a preset",
		Long:  "Analyze the current directory, build an execution plan from a preset, render a preview, and optionally apply the changes.",
		Example: "  boottree init\n" +
			"  boottree init --preset software-product --dry-run\n" +
			"  boottree init --mode folders-only --include 01_business,06_engineering --yes\n" +
			"  boottree init --force --yes",
		Args: cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runInit(cmd, prompter, flags)
		},
	}
	bindInitFlags(cmd, flags)
	return cmd
}

func bindInitFlags(cmd *cobra.Command, flags *initCommandFlags) {
	cmd.Flags().StringVar(&flags.Preset, "preset", "", "Preset to use for initialization")
	cmd.Flags().StringVar(&flags.Mode, "mode", "", "Initialization mode: folders-only or folders-and-templates")
	cmd.Flags().StringSliceVar(&flags.Include, "include", nil, "Comma-separated top-level sections to include")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Render the execution plan without writing changes")
	cmd.Flags().BoolVar(&flags.Yes, "yes", false, "Apply changes without confirmation")
	cmd.Flags().BoolVar(&flags.Force, "force", false, "Overwrite template files that BootTree is about to create")
}

func runInit(cmd *cobra.Command, prompter initPrompter, flags *initCommandFlags) error {
	cwd, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("resolve current directory: %w", err)
	}

	bootstrap := app.NewBootstrap()
	presetRepo := bootstrap.Presets
	fsys := fs.OSFileSystem{}
	initPlanner := app.InitPlanner{Scanner: scanner.Service{FS: fsys}, Planner: planner.Service{}}
	initApplier := app.InitApplier{FS: fsys, Templates: bootstrap.Templates, Renderer: bootstrap.Renderer}

	ctx := context.Background()
	parsed := readInitFlags(cmd, flags)
	options, presetName, err := completeInitOptions(ctx, presetRepo, cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), prompter, parsed)
	if err != nil {
		return err
	}

	preset, err := presetRepo.Get(ctx, presetName)
	if err != nil {
		return err
	}

	_, plan, err := initPlanner.BuildExecutionPlan(ctx, cwd, preset, options)
	if err != nil {
		return err
	}

	fmt.Fprintln(cmd.OutOrStdout(), render.RenderExecutionPlan(plan))

	if len(plan.Conflicts) > 0 {
		return fmt.Errorf("execution plan contains conflicts; resolve them or change options")
	}
	if options.DryRun {
		fmt.Fprintln(cmd.OutOrStdout(), render.RenderApplySummary(plan, true))
		return nil
	}

	if !parsed.Yes {
		confirmed, err := prompter.ConfirmApply(cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr())
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(cmd.OutOrStdout(), "Apply canceled.")
			return nil
		}
	}

	if err := initApplier.Apply(ctx, cwd, preset, options, plan); err != nil {
		return err
	}
	fmt.Fprintln(cmd.OutOrStdout(), render.RenderApplySummary(plan, false))
	return nil
}

func readInitFlags(cmd *cobra.Command, flags *initCommandFlags) parsedInitArgs {
	return parsedInitArgs{
		Preset:      strings.TrimSpace(flags.Preset),
		Mode:        strings.TrimSpace(flags.Mode),
		Include:     compactStrings(flags.Include),
		DryRun:      flags.DryRun,
		Yes:         flags.Yes,
		Force:       flags.Force,
		Interactive: !anyInitFlagsChanged(cmd),
	}
}

func anyInitFlagsChanged(cmd *cobra.Command) bool {
	for _, name := range []string{"preset", "mode", "include", "dry-run", "yes", "force"} {
		if cmd.Flags().Changed(name) {
			return true
		}
	}
	return false
}

func completeInitOptions(ctx context.Context, repo model.PresetRepository, in io.Reader, out io.Writer, errOut io.Writer, prompter initPrompter, parsed parsedInitArgs) (model.InitOptions, string, error) {
	presets, err := repo.List(ctx)
	if err != nil {
		return model.InitOptions{}, "", fmt.Errorf("list presets: %w", err)
	}
	if len(presets) == 0 {
		return model.InitOptions{}, "", fmt.Errorf("no presets available")
	}
	sort.Slice(presets, func(i, j int) bool { return presets[i].Name < presets[j].Name })

	presetName := parsed.Preset
	if presetName == "" {
		if parsed.Interactive {
			presetName, err = prompter.SelectPreset(in, out, errOut, presets)
			if err != nil {
				return model.InitOptions{}, "", err
			}
		} else {
			presetName = presets[0].Name
		}
	}

	preset, err := repo.Get(ctx, presetName)
	if err != nil {
		return model.InitOptions{}, "", fmt.Errorf("resolve preset %q: %w", presetName, err)
	}

	mode := model.InitMode(parsed.Mode)
	if mode == "" {
		if parsed.Interactive {
			mode, err = prompter.SelectMode(in, out, errOut)
			if err != nil {
				return model.InitOptions{}, "", err
			}
		} else {
			mode = model.InitModeFoldersAndTemplates
		}
	}
	if mode != model.InitModeFoldersOnly && mode != model.InitModeFoldersAndTemplates {
		return model.InitOptions{}, "", fmt.Errorf("unsupported init mode %q", mode)
	}

	selectedSections := parsed.Include
	if len(selectedSections) == 0 {
		if parsed.Interactive {
			selectedSections, err = prompter.SelectSections(in, out, errOut, preset.Sections)
			if err != nil {
				return model.InitOptions{}, "", err
			}
		} else {
			selectedSections = allSectionIDs(preset.Sections)
		}
	} else if err := validateSections(selectedSections, preset.Sections); err != nil {
		return model.InitOptions{}, "", err
	}

	return model.InitOptions{
		Preset:           presetName,
		Mode:             mode,
		SelectedSections: selectedSections,
		DryRun:           parsed.DryRun,
		Force:            parsed.Force,
		Interactive:      parsed.Interactive,
	}, presetName, nil
}

func allSectionIDs(sections []model.Section) []string {
	result := make([]string, 0, len(sections))
	for _, section := range sections {
		result = append(result, section.ID)
	}
	return result
}

func validateSections(selected []string, sections []model.Section) error {
	allowed := make(map[string]struct{}, len(sections))
	for _, section := range sections {
		allowed[section.ID] = struct{}{}
	}
	for _, item := range selected {
		if _, ok := allowed[item]; !ok {
			return fmt.Errorf("unknown section %q", item)
		}
	}
	return nil
}

func compactStrings(values []string) []string {
	result := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		trimmed := strings.TrimSpace(value)
		if trimmed == "" {
			continue
		}
		if _, exists := seen[trimmed]; exists {
			continue
		}
		seen[trimmed] = struct{}{}
		result = append(result, trimmed)
	}
	return result
}

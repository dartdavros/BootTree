package cli

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"boottree/internal/app"
	"boottree/internal/core/model"
	"boottree/internal/core/planner"
	"boottree/internal/core/scanner"
	"boottree/internal/fs"
	"boottree/internal/render"

	"github.com/spf13/cobra"
)

type initFlagValues struct {
	Preset  string
	Mode    string
	Include string
	DryRun  bool
	Yes     bool
	Force   bool
}

func newInitCommand() *cobra.Command {
	var flags initFlagValues
	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize the current project from a preset",
		Long:  "Analyze the current directory, build an execution plan from a preset, render a preview, and optionally apply the changes.",
		Example: "  boottree init\n" +
			"  boottree init --preset software-product --dry-run\n" +
			"  boottree init --mode folders-only --include 01_business,06_engineering --yes\n" +
			"  boottree init --force --yes",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runInit(cmd, flags, args); err != nil {
				fmt.Fprintln(cmd.ErrOrStderr(), "Error:", err)
				os.Exit(1)
			}
		},
	}
	cmd.Flags().StringVar(&flags.Preset, "preset", "", "Preset to use for initialization")
	cmd.Flags().StringVar(&flags.Mode, "mode", "", "Initialization mode: folders-only or folders-and-templates")
	cmd.Flags().StringVar(&flags.Include, "include", "", "Comma-separated top-level sections to include")
	cmd.Flags().BoolVar(&flags.DryRun, "dry-run", false, "Render the execution plan without writing files")
	cmd.Flags().BoolVar(&flags.Yes, "yes", false, "Apply without confirmation")
	cmd.Flags().BoolVar(&flags.Force, "force", false, "Overwrite template files managed by BootTree")
	return cmd
}

func runInit(cmd *cobra.Command, flagValues initFlagValues, args []string) error {
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
	parsed, err := parseInitFlags(cmd, flagValues, args)
	if err != nil {
		return err
	}

	input := bufio.NewReader(cmd.InOrStdin())
	output := cmd.OutOrStdout()
	options, presetName, err := completeInitOptions(ctx, presetRepo, input, output, parsed)
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

	fmt.Fprintln(output, render.RenderExecutionPlan(plan))

	if len(plan.Conflicts) > 0 {
		return fmt.Errorf("execution plan contains conflicts; resolve them or change options")
	}
	if options.DryRun {
		fmt.Fprintln(output, render.RenderApplySummary(plan, true))
		return nil
	}

	if !parsed.Yes {
		confirmed, err := confirmApply(input, output)
		if err != nil {
			return err
		}
		if !confirmed {
			fmt.Fprintln(output, "Apply canceled.")
			return nil
		}
	}

	if err := initApplier.Apply(ctx, cwd, preset, options, plan); err != nil {
		return err
	}
	fmt.Fprintln(output, render.RenderApplySummary(plan, false))
	return nil
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

func parseInitFlags(cmd *cobra.Command, flagValues initFlagValues, args []string) (parsedInitArgs, error) {
	if len(args) > 0 {
		return parsedInitArgs{}, fmt.Errorf("unknown argument: %s", args[0])
	}
	flags := cmd.Flags()
	interactive := !(flags.Changed("preset") || flags.Changed("mode") || flags.Changed("include") || flags.Changed("dry-run") || flags.Changed("yes") || flags.Changed("force"))
	return parsedInitArgs{
		Preset:      strings.TrimSpace(flagValues.Preset),
		Mode:        strings.TrimSpace(flagValues.Mode),
		Include:     splitCommaList(flagValues.Include),
		DryRun:      flagValues.DryRun,
		Yes:         flagValues.Yes,
		Force:       flagValues.Force,
		Interactive: interactive,
	}, nil
}

func completeInitOptions(ctx context.Context, repo model.PresetRepository, in *bufio.Reader, out io.Writer, parsed parsedInitArgs) (model.InitOptions, string, error) {
	presets, err := repo.List(ctx)
	if err != nil {
		return model.InitOptions{}, "", fmt.Errorf("list presets: %w", err)
	}
	if len(presets) == 0 {
		return model.InitOptions{}, "", fmt.Errorf("no presets available")
	}
	sort.Slice(presets, func(i, j int) bool { return presets[i].Name < presets[j].Name })

	presetName := parsed.Preset
	if strings.TrimSpace(presetName) == "" {
		if parsed.Interactive {
			presetName, err = promptPresetSelection(in, out, presets)
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
			mode, err = promptModeSelection(in, out)
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
			selectedSections, err = promptSectionSelection(in, out, preset.Sections)
			if err != nil {
				return model.InitOptions{}, "", err
			}
		} else {
			selectedSections = allSectionIDs(preset.Sections)
		}
	} else if err := validateSections(selectedSections, preset.Sections); err != nil {
		return model.InitOptions{}, "", err
	}

	options := model.InitOptions{
		Preset:           presetName,
		Mode:             mode,
		SelectedSections: selectedSections,
		DryRun:           parsed.DryRun,
		Force:            parsed.Force,
		Interactive:      parsed.Interactive,
	}
	return options, presetName, nil
}

func allSectionIDs(sections []model.Section) []string {
	result := make([]string, 0, len(sections))
	for _, section := range sections {
		result = append(result, section.ID)
	}
	return result
}

func validateSections(selected []string, sections []model.Section) error {
	allowed := map[string]struct{}{}
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

func splitCommaList(value string) []string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	parts := strings.Split(value, ",")
	result := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

func confirmApply(in *bufio.Reader, out io.Writer) (bool, error) {
	for {
		fmt.Fprint(out, "Apply changes? [y/N]: ")
		line, err := readPromptLine(in)
		if err != nil {
			return false, fmt.Errorf("read confirmation: %w", err)
		}
		answer := strings.TrimSpace(strings.ToLower(line))
		switch answer {
		case "", "n", "no":
			return false, nil
		case "y", "yes":
			return true, nil
		default:
			fmt.Fprintln(out, "Enter y or n.")
		}
	}
}

func promptPresetSelection(in *bufio.Reader, out io.Writer, presets []model.Preset) (string, error) {
	options := make([]promptOption, 0, len(presets))
	for _, preset := range presets {
		options = append(options, promptOption{
			Value:       preset.Name,
			Label:       preset.Name,
			Description: preset.Description,
		})
	}
	return promptSingleChoice(in, out, "Select preset", options, 0)
}

func promptModeSelection(in *bufio.Reader, out io.Writer) (model.InitMode, error) {
	choice, err := promptSingleChoice(in, out, "Select initialization mode", []promptOption{
		{
			Value:       string(model.InitModeFoldersAndTemplates),
			Label:       string(model.InitModeFoldersAndTemplates),
			Description: "create directories and template files",
		},
		{
			Value:       string(model.InitModeFoldersOnly),
			Label:       string(model.InitModeFoldersOnly),
			Description: "create only directories from the preset",
		},
	}, 0)
	if err != nil {
		return "", err
	}
	return model.InitMode(choice), nil
}

func promptSectionSelection(in *bufio.Reader, out io.Writer, sections []model.Section) ([]string, error) {
	fmt.Fprintln(out, "Select sections to include:")
	for index, section := range sections {
		fmt.Fprintf(out, "  %d) %s\n", index+1, formatSectionLabel(section))
	}
	fmt.Fprintln(out, "Press Enter to include all sections, or enter comma-separated numbers.")

	for {
		fmt.Fprint(out, "Sections [all]: ")
		line, err := readPromptLine(in)
		if err != nil {
			return nil, fmt.Errorf("read section selection: %w", err)
		}
		if strings.TrimSpace(line) == "" {
			return allSectionIDs(sections), nil
		}

		indexes, err := parseNumericSelection(line, len(sections))
		if err != nil {
			fmt.Fprintf(out, "Invalid selection: %v\n", err)
			continue
		}

		result := make([]string, 0, len(indexes))
		for _, index := range indexes {
			result = append(result, sections[index].ID)
		}
		return result, nil
	}
}

type promptOption struct {
	Value       string
	Label       string
	Description string
}

func promptSingleChoice(in *bufio.Reader, out io.Writer, title string, options []promptOption, defaultIndex int) (string, error) {
	if len(options) == 0 {
		return "", fmt.Errorf("no options available for %s", strings.ToLower(title))
	}

	fmt.Fprintf(out, "%s:\n", title)
	for index, option := range options {
		fmt.Fprintf(out, "  %d) %s\n", index+1, formatPromptOption(option))
	}

	for {
		fmt.Fprintf(out, "Choice [%d]: ", defaultIndex+1)
		line, err := readPromptLine(in)
		if err != nil {
			return "", fmt.Errorf("read selection: %w", err)
		}
		if strings.TrimSpace(line) == "" {
			return options[defaultIndex].Value, nil
		}

		choice, err := strconv.Atoi(strings.TrimSpace(line))
		if err != nil || choice < 1 || choice > len(options) {
			fmt.Fprintf(out, "Enter a number between 1 and %d.\n", len(options))
			continue
		}
		return options[choice-1].Value, nil
	}
}

func formatPromptOption(option promptOption) string {
	if strings.TrimSpace(option.Description) == "" {
		return option.Label
	}
	return fmt.Sprintf("%s — %s", option.Label, option.Description)
}

func formatSectionLabel(section model.Section) string {
	label := section.ID
	if strings.TrimSpace(section.Label) != "" {
		label = fmt.Sprintf("%s — %s", section.ID, section.Label)
	}
	if strings.TrimSpace(section.Description) != "" {
		label = fmt.Sprintf("%s (%s)", label, section.Description)
	}
	return label
}

func parseNumericSelection(raw string, max int) ([]int, error) {
	parts := strings.Split(raw, ",")
	result := make([]int, 0, len(parts))
	seen := make(map[int]struct{}, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed == "" {
			continue
		}
		value, err := strconv.Atoi(trimmed)
		if err != nil {
			return nil, fmt.Errorf("%q is not a number", trimmed)
		}
		if value < 1 || value > max {
			return nil, fmt.Errorf("%d is outside the allowed range 1..%d", value, max)
		}
		index := value - 1
		if _, exists := seen[index]; exists {
			continue
		}
		seen[index] = struct{}{}
		result = append(result, index)
	}
	if len(result) == 0 {
		return nil, fmt.Errorf("at least one selection is required")
	}
	return result, nil
}

func readPromptLine(in *bufio.Reader) (string, error) {
	line, err := in.ReadString('\n')
	if err != nil {
		if err == io.EOF {
			return strings.TrimRight(line, "\r\n"), nil
		}
		return "", err
	}
	return strings.TrimRight(line, "\r\n"), nil
}

func projectNameFromDir(path string) string {
	return filepath.Base(filepath.Clean(path))
}

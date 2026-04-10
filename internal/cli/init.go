package cli

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"boottree/internal/app"
	"boottree/internal/core/model"
	"boottree/internal/core/planner"
	"boottree/internal/core/scanner"
	"boottree/internal/fs"
	"boottree/internal/render"

	"github.com/spf13/cobra"
)

func newInitCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "init",
		Short: "Initialize the current project from a preset",
		Run: func(cmd *cobra.Command, args []string) {
			if err := runInit(cmd, args); err != nil {
				fmt.Fprintln(cmd.OutOrStdout(), "Error:", err)
				os.Exit(1)
			}
		},
	}
}

func runInit(cmd *cobra.Command, args []string) error {
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
	parsed, err := parseInitArgs(args)
	if err != nil {
		return err
	}
	options, presetName, err := completeInitOptions(ctx, presetRepo, cmd.OutOrStdout(), parsed)
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
		confirmed, err := confirmApply(cmd.OutOrStdout())
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

type parsedInitArgs struct {
	Preset      string
	Mode        string
	Include     []string
	DryRun      bool
	Yes         bool
	Force       bool
	Interactive bool
}

func parseInitArgs(args []string) (parsedInitArgs, error) {
	parsed := parsedInitArgs{Interactive: len(args) == 0}
	for i := 0; i < len(args); i++ {
		arg := args[i]
		switch {
		case arg == "--dry-run":
			parsed.DryRun = true
			parsed.Interactive = false
		case arg == "--yes":
			parsed.Yes = true
			parsed.Interactive = false
		case arg == "--force":
			parsed.Force = true
			parsed.Interactive = false
		case strings.HasPrefix(arg, "--preset="):
			parsed.Preset = strings.TrimPrefix(arg, "--preset=")
			parsed.Interactive = false
		case arg == "--preset":
			i++
			if i >= len(args) {
				return parsedInitArgs{}, fmt.Errorf("flag --preset requires a value")
			}
			parsed.Preset = args[i]
			parsed.Interactive = false
		case strings.HasPrefix(arg, "--mode="):
			parsed.Mode = strings.TrimPrefix(arg, "--mode=")
			parsed.Interactive = false
		case arg == "--mode":
			i++
			if i >= len(args) {
				return parsedInitArgs{}, fmt.Errorf("flag --mode requires a value")
			}
			parsed.Mode = args[i]
			parsed.Interactive = false
		case strings.HasPrefix(arg, "--include="):
			parsed.Include = splitCommaList(strings.TrimPrefix(arg, "--include="))
			parsed.Interactive = false
		case arg == "--include":
			i++
			if i >= len(args) {
				return parsedInitArgs{}, fmt.Errorf("flag --include requires a value")
			}
			parsed.Include = splitCommaList(args[i])
			parsed.Interactive = false
		default:
			return parsedInitArgs{}, fmt.Errorf("unknown argument: %s", arg)
		}
	}
	return parsed, nil
}

func completeInitOptions(ctx context.Context, repo model.PresetRepository, out interface{ Write([]byte) (int, error) }, parsed parsedInitArgs) (model.InitOptions, string, error) {
	presets, err := repo.List(ctx)
	if err != nil {
		return model.InitOptions{}, "", fmt.Errorf("list presets: %w", err)
	}
	if len(presets) == 0 {
		return model.InitOptions{}, "", fmt.Errorf("no presets available")
	}

	presetName := parsed.Preset
	if strings.TrimSpace(presetName) == "" {
		if parsed.Interactive {
			presetName = presets[0].Name
			fmt.Fprintf(out, "Preset: %s\n", presetName)
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
			mode = model.InitModeFoldersAndTemplates
			fmt.Fprintf(out, "Mode: %s\n", mode)
		} else {
			mode = model.InitModeFoldersAndTemplates
		}
	}
	if mode != model.InitModeFoldersOnly && mode != model.InitModeFoldersAndTemplates {
		return model.InitOptions{}, "", fmt.Errorf("unsupported init mode %q", mode)
	}

	selectedSections := parsed.Include
	if len(selectedSections) == 0 {
		selectedSections = allSectionIDs(preset.Sections)
		if parsed.Interactive {
			fmt.Fprintf(out, "Sections: %s\n", strings.Join(selectedSections, ", "))
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

func confirmApply(out interface{ Write([]byte) (int, error) }) (bool, error) {
	fmt.Fprint(out, "Apply this plan? [y/N]: ")
	reader := bufio.NewReader(os.Stdin)
	line, err := reader.ReadString('\n')
	if err != nil && !strings.Contains(err.Error(), "EOF") {
		return false, fmt.Errorf("read confirmation: %w", err)
	}
	answer := strings.ToLower(strings.TrimSpace(line))
	return answer == "y" || answer == "yes", nil
}

func targetProjectName(root string) string { return filepath.Base(filepath.Clean(root)) }

package cli

import (
	"errors"
	"fmt"
	"io"
	"strings"

	"boottree/internal/core/model"
	"boottree/internal/platform"
	"boottree/internal/update"

	survey "github.com/AlecAivazis/survey/v2"
	surveyterminal "github.com/AlecAivazis/survey/v2/terminal"
)

type initPrompter interface {
	SelectPreset(in io.Reader, out io.Writer, errOut io.Writer, presets []model.Preset) (string, error)
	SelectMode(in io.Reader, out io.Writer, errOut io.Writer) (model.InitMode, error)
	SelectSections(in io.Reader, out io.Writer, errOut io.Writer, sections []model.Section) ([]string, error)
	ConfirmApply(in io.Reader, out io.Writer, errOut io.Writer) (bool, error)
}

type installPrompter interface {
	ConfirmInstall(in io.Reader, out io.Writer, errOut io.Writer, state platform.InstallState) (bool, error)
}

type surveyInitPrompter struct{}

func (surveyInitPrompter) SelectPreset(in io.Reader, out io.Writer, errOut io.Writer, presets []model.Preset) (string, error) {
	options := make([]string, 0, len(presets))
	for _, preset := range presets {
		options = append(options, preset.Name)
	}

	prompt := &survey.Select{
		Message: "Select preset:",
		Options: options,
		Default: options[0],
		Description: func(value string, index int) string {
			return presets[index].Description
		},
	}

	var answer string
	if err := survey.AskOne(prompt, &answer, askOpts(in, out, errOut)...); err != nil {
		return "", wrapPromptError("select preset", err)
	}
	return answer, nil
}

func (surveyInitPrompter) SelectMode(in io.Reader, out io.Writer, errOut io.Writer) (model.InitMode, error) {
	options := []string{string(model.InitModeFoldersAndTemplates), string(model.InitModeFoldersOnly)}
	descriptions := map[string]string{
		string(model.InitModeFoldersAndTemplates): "create directories and template files",
		string(model.InitModeFoldersOnly):         "create only directories from the preset",
	}

	prompt := &survey.Select{
		Message: "Select initialization mode:",
		Options: options,
		Default: options[0],
		Description: func(value string, index int) string {
			return descriptions[value]
		},
	}

	var answer string
	if err := survey.AskOne(prompt, &answer, askOpts(in, out, errOut)...); err != nil {
		return "", wrapPromptError("select mode", err)
	}
	return model.InitMode(answer), nil
}

func (surveyInitPrompter) SelectSections(in io.Reader, out io.Writer, errOut io.Writer, sections []model.Section) ([]string, error) {
	options := make([]string, 0, len(sections))
	defaults := make([]string, 0, len(sections))
	optionToID := make(map[string]string, len(sections))
	for _, section := range sections {
		option := sectionOptionLabel(section)
		options = append(options, option)
		defaults = append(defaults, option)
		optionToID[option] = section.ID
	}

	prompt := &survey.MultiSelect{
		Message:  "Select sections to include:",
		Options:  options,
		Default:  defaults,
		PageSize: len(options),
		Description: func(value string, index int) string {
			return describeSection(sections[index])
		},
	}

	var answer []string
	if err := survey.AskOne(prompt, &answer, askOpts(in, out, errOut)...); err != nil {
		return nil, wrapPromptError("select sections", err)
	}
	if len(answer) == 0 {
		return nil, fmt.Errorf("at least one section must be selected")
	}

	selected := make([]string, 0, len(answer))
	for _, option := range answer {
		id, ok := optionToID[option]
		if !ok {
			return nil, fmt.Errorf("select sections: unknown section option %q", option)
		}
		selected = append(selected, id)
	}
	return selected, nil
}

func (p surveyInitPrompter) ConfirmApply(in io.Reader, out io.Writer, errOut io.Writer) (bool, error) {
	return p.confirmBoolean(in, out, errOut, "Apply changes?", false, "confirm apply")
}

func (p surveyInitPrompter) ConfirmInstall(in io.Reader, out io.Writer, errOut io.Writer, state platform.InstallState) (bool, error) {
	message := fmt.Sprintf("Install %s into %s and add it to PATH for the current user?", state.CommandName, state.SuggestedInstallDir)
	return p.confirmBoolean(in, out, errOut, message, true, "confirm install")
}

func (p surveyInitPrompter) ConfirmUpdate(in io.Reader, out io.Writer, errOut io.Writer, plan update.Plan) (bool, error) {
	message := fmt.Sprintf("Install BootTree %s over %s?", plan.TargetVersion, plan.CurrentVersion)
	if plan.RequiresDeferredSwap {
		message += " The current Windows executable will be replaced after this process exits."
	}
	return p.confirmBoolean(in, out, errOut, message, true, "confirm update")
}

func (surveyInitPrompter) confirmBoolean(in io.Reader, out io.Writer, errOut io.Writer, message string, defaultValue bool, action string) (bool, error) {
	prompt := &survey.Confirm{Message: message, Default: defaultValue}
	var confirmed bool
	if err := survey.AskOne(prompt, &confirmed, askOpts(in, out, errOut)...); err != nil {
		return false, wrapPromptError(action, err)
	}
	return confirmed, nil
}

func askOpts(in io.Reader, out io.Writer, errOut io.Writer) []survey.AskOpt {
	inFile, ok := in.(surveyterminal.FileReader)
	if !ok { return nil }
	outFile, ok := out.(surveyterminal.FileWriter)
	if !ok { return nil }
	return []survey.AskOpt{survey.WithStdio(inFile, outFile, errOut)}
}

func wrapPromptError(action string, err error) error {
	if err == nil { return nil }
	if errors.Is(err, surveyterminal.InterruptErr) {
		return fmt.Errorf("%s: interrupted", action)
	}
	return fmt.Errorf("%s: %w", action, err)
}

func describeSection(section model.Section) string {
	parts := make([]string, 0, 2)
	if label := strings.TrimSpace(section.Label); label != "" { parts = append(parts, label) }
	if description := strings.TrimSpace(section.Description); description != "" { parts = append(parts, description) }
	return strings.Join(parts, " — ")
}

func sectionOptionLabel(section model.Section) string {
	if label := strings.TrimSpace(section.Label); label != "" { return label }
	return section.ID
}

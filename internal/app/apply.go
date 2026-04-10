package app

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"boottree/internal/core/model"
)

type InitApplier struct {
	FS        model.FileSystem
	Templates model.TemplateRepository
	Renderer  model.TemplateRenderer
}

func (a InitApplier) Apply(ctx context.Context, root string, preset model.Preset, options model.InitOptions, plan model.ExecutionPlan) error {
	if a.FS == nil {
		return fmt.Errorf("apply execution plan: file system is required")
	}
	if a.Templates == nil {
		return fmt.Errorf("apply execution plan: template repository is required")
	}
	if a.Renderer == nil {
		return fmt.Errorf("apply execution plan: template renderer is required")
	}

	cleanRoot := filepath.Clean(root)
	for _, action := range plan.DirectoriesToCreate {
		target := filepath.Join(cleanRoot, action.Path)
		if err := a.FS.MkdirAll(target); err != nil {
			return fmt.Errorf("create directory %q: %w", action.Path, err)
		}
	}

	if options.Mode != model.InitModeFoldersAndTemplates {
		return nil
	}

	selectedSections := selectedSectionSet(preset, options.SelectedSections)
	data := model.TemplateData{
		ProjectName: filepath.Base(cleanRoot),
		Year:        time.Now().Year(),
		CreatedAt:   time.Now().UTC(),
	}

	fileTargets := make(map[string]struct{}, len(plan.FilesToCreate))
	for _, action := range plan.FilesToCreate {
		fileTargets[filepath.Clean(action.Path)] = struct{}{}
	}

	for _, tpl := range preset.Templates {
		if !includesSection(tpl.Sections, selectedSections) {
			continue
		}
		targetPath := filepath.Clean(tpl.TargetPath)
		if _, ok := fileTargets[targetPath]; !ok {
			continue
		}

		templateText, err := a.Templates.Get(ctx, tpl.SourceTemplate)
		if err != nil {
			return fmt.Errorf("load template %q: %w", tpl.SourceTemplate, err)
		}

		data.Variables = cloneVariables(tpl.Variables)
		rendered, err := a.Renderer.Render(templateText, data)
		if err != nil {
			return fmt.Errorf("render template %q to %q: %w", tpl.SourceTemplate, targetPath, err)
		}

		fullPath := filepath.Join(cleanRoot, targetPath)
		parent := filepath.Dir(fullPath)
		if parent != "." && parent != "" {
			if err := a.FS.MkdirAll(parent); err != nil {
				return fmt.Errorf("prepare parent directory for %q: %w", targetPath, err)
			}
		}
		if err := a.FS.WriteFile(fullPath, []byte(rendered)); err != nil {
			return fmt.Errorf("write file %q: %w", targetPath, err)
		}
	}

	return nil
}

func selectedSectionSet(preset model.Preset, requested []string) map[string]struct{} {
	if len(requested) > 0 {
		result := make(map[string]struct{}, len(requested))
		for _, item := range requested {
			result[strings.TrimSpace(item)] = struct{}{}
		}
		return result
	}

	result := make(map[string]struct{}, len(preset.Sections))
	for _, section := range preset.Sections {
		result[section.ID] = struct{}{}
	}
	return result
}

func includesSection(itemSections []string, selected map[string]struct{}) bool {
	if len(itemSections) == 0 {
		return true
	}
	for _, item := range itemSections {
		if _, ok := selected[item]; ok {
			return true
		}
	}
	return false
}

func cloneVariables(input map[string]string) map[string]string {
	if len(input) == 0 {
		return nil
	}
	cloned := make(map[string]string, len(input))
	for key, value := range input {
		cloned[key] = value
	}
	return cloned
}

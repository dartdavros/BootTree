package planner

import (
	"fmt"
	"path/filepath"
	"sort"

	"boottree/internal/core/model"
)

type Service struct{}

func (Service) Build(snapshot model.TreeSnapshot, preset model.Preset, options model.InitOptions) (model.ExecutionPlan, error) {
	if preset.Name == "" {
		return model.ExecutionPlan{}, fmt.Errorf("build execution plan: preset is required")
	}

	selectedSections := resolveSelectedSections(preset, options.SelectedSections)
	existing := make(map[string]bool, len(snapshot.Entries))
	for _, entry := range snapshot.Entries {
		existing[filepath.Clean(entry.Path)] = entry.IsDir
	}

	plan := model.ExecutionPlan{
		DirectoriesToCreate: make([]model.PlanAction, 0),
		FilesToCreate:       make([]model.PlanAction, 0),
		SkippedExisting:     make([]model.PlanAction, 0),
		Conflicts:           make([]model.PlanAction, 0),
		Warnings:            make([]string, 0),
	}
	plannedDirs := map[string]struct{}{}
	plannedFiles := map[string]struct{}{}

	for _, node := range preset.Directories {
		flattenDirectoryNode(node, "", nil, func(path string, sections []string) {
			if !isIncludedBySections(sections, selectedSections) {
				return
			}
			cleanPath := filepath.Clean(path)
			if _, exists := plannedDirs[cleanPath]; exists {
				return
			}
			if isDir, exists := existing[cleanPath]; exists {
				if isDir {
					plan.SkippedExisting = append(plan.SkippedExisting, model.PlanAction{Path: cleanPath, Reason: "directory already exists"})
				} else {
					plan.Conflicts = append(plan.Conflicts, model.PlanAction{Path: cleanPath, Reason: "target directory path already occupied by file"})
				}
				plannedDirs[cleanPath] = struct{}{}
				return
			}
			plannedDirs[cleanPath] = struct{}{}
			plan.DirectoriesToCreate = append(plan.DirectoriesToCreate, model.PlanAction{Path: cleanPath, Reason: "missing directory from preset"})
		})
	}

	if options.Mode == model.InitModeFoldersAndTemplates {
		for _, tpl := range preset.Templates {
			if !isIncludedBySections(tpl.Sections, selectedSections) {
				continue
			}
			cleanPath := filepath.Clean(tpl.TargetPath)
			if _, duplicate := plannedFiles[cleanPath]; duplicate {
				plan.Warnings = append(plan.Warnings, fmt.Sprintf("template target %q is duplicated in preset", cleanPath))
				continue
			}
			plannedFiles[cleanPath] = struct{}{}

			if parent := filepath.Dir(cleanPath); parent != "." {
				if isDir, exists := existing[parent]; exists && !isDir {
					plan.Conflicts = append(plan.Conflicts, model.PlanAction{Path: cleanPath, Reason: "parent path is occupied by file"})
					continue
				}
			}

			if isDir, exists := existing[cleanPath]; exists {
				if isDir {
					plan.Conflicts = append(plan.Conflicts, model.PlanAction{Path: cleanPath, Reason: "target file path already occupied by directory"})
				} else if options.Force {
					plan.FilesToCreate = append(plan.FilesToCreate, model.PlanAction{Path: cleanPath, Reason: "overwrite existing template file"})
				} else {
					plan.SkippedExisting = append(plan.SkippedExisting, model.PlanAction{Path: cleanPath, Reason: "file already exists"})
				}
				continue
			}

			plan.FilesToCreate = append(plan.FilesToCreate, model.PlanAction{Path: cleanPath, Reason: "missing template target"})
		}
	}

	sortPlan(&plan)
	return plan, nil
}

func resolveSelectedSections(preset model.Preset, requested []string) map[string]struct{} {
	if len(requested) > 0 {
		selected := make(map[string]struct{}, len(requested))
		for _, item := range requested {
			selected[item] = struct{}{}
		}
		return selected
	}

	selected := make(map[string]struct{}, len(preset.Sections))
	for _, section := range preset.Sections {
		selected[section.ID] = struct{}{}
	}
	return selected
}

func isIncludedBySections(itemSections []string, selected map[string]struct{}) bool {
	if len(itemSections) == 0 {
		return true
	}
	for _, section := range itemSections {
		if _, ok := selected[section]; ok {
			return true
		}
	}
	return false
}

func flattenDirectoryNode(node model.DirectoryNode, parent string, inherited []string, yield func(path string, sections []string)) {
	currentPath := filepath.Clean(filepath.Join(parent, node.Path))
	sections := node.Sections
	if len(sections) == 0 {
		sections = inherited
	}
	yield(currentPath, sections)
	for _, child := range node.Children {
		flattenDirectoryNode(child, currentPath, sections, yield)
	}
}

func sortPlan(plan *model.ExecutionPlan) {
	sort.Slice(plan.DirectoriesToCreate, func(i, j int) bool { return plan.DirectoriesToCreate[i].Path < plan.DirectoriesToCreate[j].Path })
	sort.Slice(plan.FilesToCreate, func(i, j int) bool { return plan.FilesToCreate[i].Path < plan.FilesToCreate[j].Path })
	sort.Slice(plan.SkippedExisting, func(i, j int) bool { return plan.SkippedExisting[i].Path < plan.SkippedExisting[j].Path })
	sort.Slice(plan.Conflicts, func(i, j int) bool { return plan.Conflicts[i].Path < plan.Conflicts[j].Path })
	sort.Strings(plan.Warnings)
}

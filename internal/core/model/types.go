package model

import "time"

type Preset struct {
	Name        string          `json:"name"`
	Description string          `json:"description"`
	Sections    []Section       `json:"sections"`
	Directories []DirectoryNode `json:"directories"`
	Templates   []TemplateFile  `json:"templates"`
}

type Section struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	Description string `json:"description,omitempty"`
}

type DirectoryNode struct {
	Path     string          `json:"path"`
	Sections []string        `json:"sections,omitempty"`
	Children []DirectoryNode `json:"children,omitempty"`
}

type TemplateFile struct {
	SourceTemplate string            `json:"sourceTemplate"`
	TargetPath     string            `json:"targetPath"`
	Sections       []string          `json:"sections,omitempty"`
	Variables      map[string]string `json:"variables,omitempty"`
}

type InitMode string

const (
	InitModeFoldersOnly         InitMode = "folders-only"
	InitModeFoldersAndTemplates InitMode = "folders-and-templates"
)

type InitOptions struct {
	Preset            string
	Mode              InitMode
	SelectedSections  []string
	DryRun            bool
	Force             bool
	Interactive       bool
	OverwriteExisting bool
}

type PlanAction struct {
	Path   string
	Reason string
}

type ExecutionPlan struct {
	DirectoriesToCreate []PlanAction
	FilesToCreate       []PlanAction
	SkippedExisting     []PlanAction
	Conflicts           []PlanAction
	Warnings            []string
}

type TreeEntry struct {
	Path  string
	IsDir bool
}

type TreeSnapshot struct {
	Root    string
	Entries []TreeEntry
}

type ExtensionStat struct {
	Extension string
	Count     int
}

type ProjectStats struct {
	Files               int
	Directories         int
	EmptyDirectories    int
	ByExtension         []ExtensionStat
	SecretLikeFilePaths []string
}

type TemplateData struct {
	ProjectName string
	Year        int
	CreatedAt   time.Time
	Variables   map[string]string
}

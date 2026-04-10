package model

import "context"

type FileSystem interface {
	Exists(path string) (bool, error)
	MkdirAll(path string) error
	WriteFile(path string, data []byte) error
	ReadFile(path string) ([]byte, error)
	ReadDir(path string) ([]DirEntry, error)
	WalkDir(root string, fn WalkDirFunc) error
}

type DirEntry interface {
	Name() string
	IsDir() bool
}

type WalkDirFunc func(path string, entry DirEntry, err error) error

type PresetRepository interface {
	List(ctx context.Context) ([]Preset, error)
	Get(ctx context.Context, name string) (Preset, error)
}

type TemplateRepository interface {
	Get(ctx context.Context, path string) (string, error)
}

type TemplateRenderer interface {
	Render(templateText string, data TemplateData) (string, error)
}

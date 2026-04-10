package preset

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"boottree/internal/assets"
	"boottree/internal/core/model"
)

type EmbeddedRepository struct{ root fs.FS }

func NewEmbeddedRepository() EmbeddedRepository {
	root, err := fs.Sub(assets.FS, "presets")
	if err != nil {
		panic(err)
	}
	return EmbeddedRepository{root: root}
}

func (r EmbeddedRepository) List(ctx context.Context) ([]model.Preset, error) {
	_ = ctx
	entries, err := fs.ReadDir(r.root, ".")
	if err != nil {
		return nil, fmt.Errorf("list presets: %w", err)
	}
	presets := make([]model.Preset, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		preset, err := r.Get(ctx, entry.Name())
		if err != nil {
			return nil, err
		}
		presets = append(presets, preset)
	}
	sort.Slice(presets, func(i, j int) bool { return presets[i].Name < presets[j].Name })
	return presets, nil
}

func (r EmbeddedRepository) Get(ctx context.Context, name string) (model.Preset, error) {
	_ = ctx
	data, err := fs.ReadFile(r.root, path.Join(name, "preset.json"))
	if err != nil {
		return model.Preset{}, fmt.Errorf("read preset %q: %w", name, err)
	}
	preset, err := Load(data)
	if err != nil {
		return model.Preset{}, fmt.Errorf("load preset %q: %w", name, err)
	}
	return preset, nil
}

func Load(data []byte) (model.Preset, error) {
	var preset model.Preset
	if err := json.Unmarshal(data, &preset); err != nil {
		return model.Preset{}, fmt.Errorf("decode preset json: %w", err)
	}
	if err := Validate(preset); err != nil {
		return model.Preset{}, err
	}
	return preset, nil
}

func Validate(p model.Preset) error {
	if strings.TrimSpace(p.Name) == "" {
		return ValidationError{Field: "name", Message: "must not be empty"}
	}
	if len(p.Sections) == 0 {
		return ValidationError{Field: "sections", Message: "must contain at least one section"}
	}
	if len(p.Directories) == 0 {
		return ValidationError{Field: "directories", Message: "must contain at least one directory"}
	}
	sectionIDs := map[string]struct{}{}
	for i, s := range p.Sections {
		if strings.TrimSpace(s.ID) == "" {
			return ValidationError{Field: fmt.Sprintf("sections[%d].id", i), Message: "must not be empty"}
		}
		if _, ok := sectionIDs[s.ID]; ok {
			return ValidationError{Field: fmt.Sprintf("sections[%d].id", i), Message: "must be unique"}
		}
		sectionIDs[s.ID] = struct{}{}
	}
	for i, node := range p.Directories {
		if err := validateDirectoryNode(node, fmt.Sprintf("directories[%d]", i), sectionIDs); err != nil {
			return err
		}
	}
	for i, tpl := range p.Templates {
		if strings.TrimSpace(tpl.SourceTemplate) == "" {
			return ValidationError{Field: fmt.Sprintf("templates[%d].sourceTemplate", i), Message: "must not be empty"}
		}
		if strings.TrimSpace(tpl.TargetPath) == "" {
			return ValidationError{Field: fmt.Sprintf("templates[%d].targetPath", i), Message: "must not be empty"}
		}
		for _, id := range tpl.Sections {
			if _, ok := sectionIDs[id]; !ok {
				return ValidationError{Field: fmt.Sprintf("templates[%d].sections", i), Message: fmt.Sprintf("unknown section %q", id)}
			}
		}
	}
	return nil
}

func validateDirectoryNode(node model.DirectoryNode, field string, sectionIDs map[string]struct{}) error {
	if strings.TrimSpace(node.Path) == "" {
		return ValidationError{Field: field + ".path", Message: "must not be empty"}
	}
	for _, id := range node.Sections {
		if _, ok := sectionIDs[id]; !ok {
			return ValidationError{Field: field + ".sections", Message: fmt.Sprintf("unknown section %q", id)}
		}
	}
	for i, child := range node.Children {
		if err := validateDirectoryNode(child, fmt.Sprintf("%s.children[%d]", field, i), sectionIDs); err != nil {
			return err
		}
	}
	return nil
}
